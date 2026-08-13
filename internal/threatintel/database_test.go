package threatintel

import (
	"archive/zip"
	"bytes"
	"compress/flate"
	"container/list"
	"context"
	"crypto/aes"
	"crypto/cipher"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"hash/crc32"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"sync"
	"testing"
	"time"
)

func TestDatabaseLookupFastAndSlowIPv4IPv6(t *testing.T) {
	tests := []struct {
		name    string
		header  databaseHeader
		address string
		key     string
	}{
		{name: "fast IPv6", header: databaseHeader{StreamMode: true}, address: "2001:db8::8", key: md5Hex("2001:db8::8")},
		{name: "slow IPv4", header: databaseHeader{StreamMode: true, SlowMode: true, SlowModeKeySize: 3}, address: "203.0.113.8", key: md5Hex("203.0.113.8_a9z")},
	}
	want := Result{Labels: []string{"Hosting"}, Behaviors: []string{"Port Scanning"}, Level: 2}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "intelligence.db")
			writeTestDatabase(t, path, test.header, map[string]Result{test.key: want})
			database, err := Load(path)
			if err != nil {
				t.Fatalf("Load() error = %v", err)
			}
			got, found := database.Lookup(test.address)
			if !found || !reflect.DeepEqual(got, want) {
				t.Fatalf("Lookup() = %#v, %v; want %#v, true", got, found, want)
			}
			if _, found := database.Lookup("198.51.100.200"); found {
				t.Fatal("unexpected match for absent address")
			}
		})
	}
}

func TestDatabaseRejectsTamperedRecord(t *testing.T) {
	path := filepath.Join(t.TempDir(), "intelligence.db")
	writeTestDatabase(t, path, databaseHeader{StreamMode: true}, map[string]Result{md5Hex("203.0.113.9"): {Level: 2, Labels: []string{"Hosting"}}})
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	data[len(data)-1] ^= 0xff
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := Load(path); err == nil {
		t.Fatal("Load() accepted a record with a broken AES-GCM tag")
	}
}

func TestManagerDownloadsEncryptedArchiveAndKeepsOldDatabaseOnFailure(t *testing.T) {
	want := Result{Labels: []string{"Company"}, Behaviors: []string{"Port Scanning"}, Level: 3}
	databaseBytes := testDatabaseBytes(t, databaseHeader{StreamMode: true}, map[string]Result{md5Hex("2001:db8::44"): want})
	archive := encryptedTestZIP(t, "test-password", databaseBytes)
	server := httptest.NewTLSServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Set("Content-Type", "application/zip")
		_, _ = writer.Write(archive)
	}))
	defer server.Close()
	databasePath := filepath.Join(t.TempDir(), "threat-intelligence.db")
	manager, err := NewManager(Config{Enabled: true, DatabasePath: databasePath, DownloadURL: server.URL, ArchivePassword: "test-password", UpdateInterval: time.Hour, HTTPClient: server.Client()})
	if err != nil {
		t.Fatal(err)
	}
	if err := manager.Update(context.Background()); err != nil {
		t.Fatalf("Update() error = %v", err)
	}
	got, found := manager.Lookup("2001:db8::44")
	if !found || !reflect.DeepEqual(got, want) {
		t.Fatalf("Lookup() = %#v, %v; want %#v, true", got, found, want)
	}
	status := manager.Status()
	if !status.Loaded || status.RecordCount != 1 || status.DatabaseUpdated == nil {
		t.Fatalf("Status() = %#v", status)
	}
	manager.cfg.ArchivePassword = "wrong-password"
	if err := manager.Update(context.Background()); err == nil {
		t.Fatal("Update() succeeded with a wrong archive password")
	}
	got, found = manager.Lookup("2001:db8::44")
	if !found || !reflect.DeepEqual(got, want) {
		t.Fatalf("failed update replaced last-known-good data: %#v, %v", got, found)
	}
}

func TestEventTagsAreBoundedAndLocalized(t *testing.T) {
	result := Result{
		Level:     3,
		Labels:    []string{"Hosting", "Hosting", ""},
		Behaviors: []string{"Port Scanning", "Bruteforce", "Brute Force", "Vulnerability Scanning"},
	}
	want := []string{"威胁情报命中", "情报等级：高危", "情报标签：托管主机", "情报行为：端口扫描", "情报行为：暴力破解", "情报行为：漏洞扫描"}
	if got := EventTags(result); !reflect.DeepEqual(got, want) {
		t.Fatalf("EventTags() = %#v; want %#v", got, want)
	}
}

func TestSecureRedirectAllowsOnlyPublisherStorageTransition(t *testing.T) {
	original, _ := http.NewRequest(http.MethodGet, "https://intelligence-0.rivers.chaitin.cn/download", nil)
	storage, _ := http.NewRequest(http.MethodGet, "https://safepoint.oss-rg-china-mainland.aliyuncs.com/file.zip", nil)
	if err := secureRedirect(storage, []*http.Request{original}); err != nil {
		t.Fatalf("publisher storage redirect rejected: %v", err)
	}
	attacker, _ := http.NewRequest(http.MethodGet, "https://attacker.example/file.zip", nil)
	if err := secureRedirect(attacker, []*http.Request{original}); err == nil {
		t.Fatal("arbitrary cross-host redirect was accepted")
	}
	secondStorage, _ := http.NewRequest(http.MethodGet, "https://another.aliyuncs.com/file.zip", nil)
	if err := secureRedirect(secondStorage, []*http.Request{original, storage}); err == nil {
		t.Fatal("second cross-host redirect was accepted")
	}
}

func TestConcurrentMissesShareLookupAndPopulateNegativeCache(t *testing.T) {
	database := &Database{items: map[string]Result{}, header: databaseHeader{StreamMode: true, SlowMode: true, SlowModeKeySize: 3}, cacheList: list.New(), cacheByAddr: make(map[string]*list.Element), inflight: make(map[string]*lookupCall)}
	const workers = 64
	start := make(chan struct{})
	var wait sync.WaitGroup
	wait.Add(workers)
	for index := 0; index < workers; index++ {
		go func() {
			defer wait.Done()
			<-start
			_, _ = database.Lookup("203.0.113.200")
		}()
	}
	close(start)
	wait.Wait()
	if database.cacheList.Len() != 1 || len(database.inflight) != 0 {
		t.Fatalf("cache/inflight state = %d/%d, want 1/0", database.cacheList.Len(), len(database.inflight))
	}
	begin := time.Now()
	if _, found := database.Lookup("203.0.113.200"); found {
		t.Fatal("unexpected threat intelligence match from negative cache")
	}
	if elapsed := time.Since(begin); elapsed > 10*time.Millisecond {
		t.Fatalf("negative cache lookup took %s", elapsed)
	}
}

func BenchmarkSlowModeMiss(b *testing.B) {
	database := &Database{items: map[string]Result{}, header: databaseHeader{StreamMode: true, SlowMode: true, SlowModeKeySize: 3}, cacheList: list.New(), cacheByAddr: make(map[string]*list.Element), inflight: make(map[string]*lookupCall)}
	for index := 0; index < b.N; index++ {
		database.Lookup(fmt.Sprintf("198.51.%d.%d", (index/250)%250, index%250))
	}
}

func writeTestDatabase(t *testing.T, path string, header databaseHeader, items map[string]Result) {
	t.Helper()
	if err := os.WriteFile(path, testDatabaseBytes(t, header, items), 0o600); err != nil {
		t.Fatal(err)
	}
}

func testDatabaseBytes(t *testing.T, header databaseHeader, items map[string]Result) []byte {
	t.Helper()
	key := databaseEncryptionKey()
	block, err := aes.NewCipher(key[:])
	if err != nil {
		t.Fatal(err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		t.Fatal(err)
	}
	var output bytes.Buffer
	recordNumber := byte(1)
	writeRecord := func(value any) {
		plain, err := json.Marshal(value)
		if err != nil {
			t.Fatal(err)
		}
		nonce := make([]byte, gcm.NonceSize())
		nonce[len(nonce)-1] = recordNumber
		recordNumber++
		encrypted := append(nonce, gcm.Seal(nil, nonce, plain, nil)...)
		if err := binary.Write(&output, binary.BigEndian, uint32(len(encrypted))); err != nil {
			t.Fatal(err)
		}
		_, _ = output.Write(encrypted)
	}
	writeRecord(header)
	for key, value := range items {
		writeRecord(map[string]Result{key: value})
	}
	return output.Bytes()
}

func encryptedTestZIP(t *testing.T, password string, database []byte) []byte {
	t.Helper()
	var output bytes.Buffer
	writer := zip.NewWriter(&output)
	write := func(name string, plain []byte) {
		var compressed bytes.Buffer
		deflater, err := flate.NewWriter(&compressed, flate.DefaultCompression)
		if err != nil {
			t.Fatal(err)
		}
		_, _ = deflater.Write(plain)
		_ = deflater.Close()
		encrypted := zipCryptoEncrypt([]byte(password), append(make([]byte, 12), compressed.Bytes()...))
		header := &zip.FileHeader{Name: name, Method: zip.Deflate, Flags: 1 | 8, CRC32: crc32.ChecksumIEEE(plain), CompressedSize64: uint64(len(encrypted)), UncompressedSize64: uint64(len(plain))}
		raw, err := writer.CreateRaw(header)
		if err != nil {
			t.Fatal(err)
		}
		_, _ = raw.Write(encrypted)
	}
	write("ipv4_ipv6_slow.db", database)
	write("readme.txt", []byte("exported at 2026-08-12 13:32"))
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	return output.Bytes()
}

func zipCryptoEncrypt(password, plain []byte) []byte {
	state := newZipCryptoReader(bytes.NewReader(nil), password)
	output := make([]byte, len(plain))
	for index, value := range plain {
		temporary := state.keys[2] | 2
		output[index] = value ^ byte((temporary*(temporary^1))>>8)
		state.update(value)
	}
	return output
}
