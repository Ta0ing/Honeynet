package threatintel

import (
	"container/list"
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5" // #nosec G501 -- required by the documented database lookup format, not used for security.
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/netip"
	"os"
	"strings"
	"sync"
	"time"
)

const (
	maxDatabaseBytes = 256 << 20
	maxRecordBytes   = 1 << 20
	maxItems         = 2_000_000
	cacheEntries     = 65536
)

// Result is the bounded, presentation-neutral subset of an IP intelligence
// record that Honeynet consumes. The source database remains an offline file;
// no record is copied into the business database.
type Result struct {
	Labels    []string `json:"labels"`
	Behaviors []string `json:"behaviors"`
	Level     int      `json:"level"`
}

type databaseHeader struct {
	SlowMode        bool `json:"slow_mode"`
	SlowModeKeySize int  `json:"slow_mode_key_size"`
	StreamMode      bool `json:"stream_mode"`
}

type cacheValue struct {
	address string
	result  Result
	found   bool
}

// Database is immutable after loading. Lookups are safe for concurrent event
// ingestion; only the small bounded result cache is mutable.
type Database struct {
	items       map[string]Result
	header      databaseHeader
	path        string
	loadedAt    time.Time
	cacheMu     sync.Mutex
	cacheList   *list.List
	cacheByAddr map[string]*list.Element
	inflight    map[string]*lookupCall
}

type lookupCall struct {
	done   chan struct{}
	result Result
	found  bool
}

func Load(path string) (*Database, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open intelligence database: %w", err)
	}
	defer file.Close()
	info, err := file.Stat()
	if err != nil {
		return nil, fmt.Errorf("stat intelligence database: %w", err)
	}
	if info.Size() <= 4 || info.Size() > maxDatabaseBytes {
		return nil, fmt.Errorf("intelligence database size %d is outside the accepted range", info.Size())
	}
	key := databaseEncryptionKey()
	block, err := aes.NewCipher(key[:])
	if err != nil {
		return nil, fmt.Errorf("initialize intelligence database cipher: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("initialize intelligence database GCM: %w", err)
	}
	readRecord := func() ([]byte, error) {
		var lengthBytes [4]byte
		if _, err := io.ReadFull(file, lengthBytes[:]); err != nil {
			return nil, err
		}
		length := binary.BigEndian.Uint32(lengthBytes[:])
		if length < uint32(gcm.NonceSize()+gcm.Overhead()) || length > maxRecordBytes {
			return nil, fmt.Errorf("invalid encrypted record length %d", length)
		}
		encrypted := make([]byte, int(length))
		if _, err := io.ReadFull(file, encrypted); err != nil {
			return nil, fmt.Errorf("read encrypted record: %w", err)
		}
		plain, err := gcm.Open(nil, encrypted[:gcm.NonceSize()], encrypted[gcm.NonceSize():], nil)
		if err != nil {
			return nil, errors.New("decrypt intelligence database record: authentication failed")
		}
		return plain, nil
	}
	headerData, err := readRecord()
	if err != nil {
		return nil, fmt.Errorf("read intelligence database header: %w", err)
	}
	var header databaseHeader
	if err := json.Unmarshal(headerData, &header); err != nil {
		return nil, fmt.Errorf("decode intelligence database header: %w", err)
	}
	if !header.StreamMode {
		return nil, errors.New("intelligence database must use the streaming format")
	}
	if header.SlowMode && (header.SlowModeKeySize < 1 || header.SlowModeKeySize > 3) {
		return nil, fmt.Errorf("unsupported slow lookup key size %d", header.SlowModeKeySize)
	}
	items := make(map[string]Result, 150_000)
	for {
		record, err := readRecord()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("read intelligence database item %d: %w", len(items)+1, err)
		}
		var values map[string]Result
		if err := json.Unmarshal(record, &values); err != nil || len(values) == 0 {
			return nil, fmt.Errorf("decode intelligence database item %d: invalid JSON record", len(items)+1)
		}
		for lookupKey, value := range values {
			if !validLookupKey(lookupKey) {
				return nil, fmt.Errorf("decode intelligence database item %d: invalid lookup key", len(items)+1)
			}
			if err := validateResult(value); err != nil {
				return nil, fmt.Errorf("decode intelligence database item %d: %w", len(items)+1, err)
			}
			items[lookupKey] = cloneResult(value)
			if len(items) > maxItems {
				return nil, errors.New("intelligence database contains too many items")
			}
		}
	}
	if len(items) == 0 {
		return nil, errors.New("intelligence database contains no items")
	}
	return &Database{items: items, header: header, path: path, loadedAt: time.Now(), cacheList: list.New(), cacheByAddr: make(map[string]*list.Element), inflight: make(map[string]*lookupCall)}, nil
}

func (d *Database) Lookup(address string) (Result, bool) {
	parsed, err := netip.ParseAddr(strings.TrimSpace(address))
	if err != nil {
		return Result{}, false
	}
	if parsed.Is4In6() {
		parsed = parsed.Unmap()
	}
	normalized := parsed.String()
	if value, ok := d.cached(normalized); ok {
		return cloneResult(value.result), value.found
	}
	call, leader := d.beginLookup(normalized)
	if !leader {
		<-call.done
		return cloneResult(call.result), call.found
	}
	result, found := d.lookupUncached(normalized)
	d.remember(normalized, result, found)
	d.finishLookup(normalized, call, result, found)
	return cloneResult(result), found
}

func (d *Database) lookupUncached(address string) (Result, bool) {
	if !d.header.SlowMode {
		value, ok := d.items[md5Hex(address)]
		return value, ok
	}
	alphabet := "abcdefghijklmnopqrstuvwxyz0123456789"
	keySize := d.header.SlowModeKeySize
	indices := make([]int, keySize)
	prefix := []byte(address + "_")
	lookupInput := make([]byte, len(prefix)+keySize)
	copy(lookupInput, prefix)
	suffix := lookupInput[len(prefix):]
	for {
		for i := range indices {
			suffix[i] = alphabet[indices[i]]
		}
		digest := md5.Sum(lookupInput) // #nosec G401 -- publisher-defined lookup key.
		var lookupKey [32]byte
		hex.Encode(lookupKey[:], digest[:])
		if value, ok := d.items[string(lookupKey[:])]; ok {
			return value, true
		}
		position := len(indices) - 1
		for position >= 0 {
			indices[position]++
			if indices[position] < len(alphabet) {
				break
			}
			indices[position] = 0
			position--
		}
		if position < 0 {
			return Result{}, false
		}
	}
}

func (d *Database) Count() int          { return len(d.items) }
func (d *Database) Path() string        { return d.path }
func (d *Database) LoadedAt() time.Time { return d.loadedAt }
func (d *Database) SlowMode() bool      { return d.header.SlowMode }
func (d *Database) SlowKeySize() int    { return d.header.SlowModeKeySize }

func (d *Database) cached(address string) (cacheValue, bool) {
	d.cacheMu.Lock()
	defer d.cacheMu.Unlock()
	element, ok := d.cacheByAddr[address]
	if !ok {
		return cacheValue{}, false
	}
	d.cacheList.MoveToFront(element)
	return element.Value.(cacheValue), true
}

func (d *Database) remember(address string, result Result, found bool) {
	d.cacheMu.Lock()
	defer d.cacheMu.Unlock()
	if existing, ok := d.cacheByAddr[address]; ok {
		existing.Value = cacheValue{address: address, result: cloneResult(result), found: found}
		d.cacheList.MoveToFront(existing)
		return
	}
	element := d.cacheList.PushFront(cacheValue{address: address, result: cloneResult(result), found: found})
	d.cacheByAddr[address] = element
	if d.cacheList.Len() > cacheEntries {
		last := d.cacheList.Back()
		delete(d.cacheByAddr, last.Value.(cacheValue).address)
		d.cacheList.Remove(last)
	}
}

func (d *Database) beginLookup(address string) (*lookupCall, bool) {
	d.cacheMu.Lock()
	defer d.cacheMu.Unlock()
	if existing, ok := d.inflight[address]; ok {
		return existing, false
	}
	call := &lookupCall{done: make(chan struct{})}
	d.inflight[address] = call
	return call, true
}

func (d *Database) finishLookup(address string, call *lookupCall, result Result, found bool) {
	d.cacheMu.Lock()
	call.result, call.found = cloneResult(result), found
	delete(d.inflight, address)
	close(call.done)
	d.cacheMu.Unlock()
}

func databaseEncryptionKey() [32]byte {
	// This is format compatibility with the publisher's MIT-licensed 1.0.2
	// SDK. Authentication is provided by AES-GCM; this embedded format key is
	// not an operator secret and must never be confused with the ZIP password.
	first := md5.Sum([]byte("chaitin-intelligence-db-secret-key")) // #nosec G401 -- database format compatibility.
	second := md5.Sum(first[:])                                    // #nosec G401 -- database format compatibility.
	var key [32]byte
	copy(key[:16], first[:])
	copy(key[16:], second[:])
	return key
}

func md5Hex(value string) string {
	sum := md5.Sum([]byte(value)) // #nosec G401 -- publisher-defined lookup key, not a security digest.
	return hex.EncodeToString(sum[:])
}

func validLookupKey(value string) bool {
	if len(value) != 32 {
		return false
	}
	_, err := hex.DecodeString(value)
	return err == nil && value == strings.ToLower(value)
}

func validateResult(value Result) error {
	if value.Level < 0 || value.Level > 10 {
		return errors.New("threat level is outside the accepted range")
	}
	if len(value.Labels) > 64 || len(value.Behaviors) > 64 {
		return errors.New("too many threat labels or behaviors")
	}
	for _, text := range append(append([]string{}, value.Labels...), value.Behaviors...) {
		if strings.TrimSpace(text) == "" || len(text) > 128 {
			return errors.New("invalid threat label or behavior")
		}
	}
	return nil
}

func cloneResult(value Result) Result {
	return Result{Labels: append([]string(nil), value.Labels...), Behaviors: append([]string(nil), value.Behaviors...), Level: value.Level}
}
