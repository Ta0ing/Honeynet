package potcert

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestManagerPersistsAndRotatesCertificate(t *testing.T) {
	directory := t.TempDir()
	manager, err := New(directory)
	if err != nil {
		t.Fatal(err)
	}
	first, err := manager.certificate()
	if err != nil {
		t.Fatal(err)
	}
	secondManager, err := New(directory)
	if err != nil {
		t.Fatal(err)
	}
	second, err := secondManager.certificate()
	if err != nil {
		t.Fatal(err)
	}
	if first.Leaf.SerialNumber.Cmp(second.Leaf.SerialNumber) != 0 {
		t.Fatal("certificate was not reused from the node state directory")
	}
	manager.now = func() time.Time { return first.Leaf.NotAfter.Add(-renewBefore + time.Minute) }
	rotated, err := manager.certificate()
	if err != nil {
		t.Fatal(err)
	}
	if first.Leaf.SerialNumber.Cmp(rotated.Leaf.SerialNumber) == 0 {
		t.Fatal("certificate did not rotate inside the renewal window")
	}
	keyInfo, err := os.Stat(filepath.Join(directory, "pot-tls", "server.key"))
	if err != nil {
		t.Fatal(err)
	}
	if keyInfo.Mode().Perm() != 0600 {
		t.Fatalf("private key mode = %o, want 600", keyInfo.Mode().Perm())
	}
}
