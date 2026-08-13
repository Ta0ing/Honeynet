package store

import (
	"encoding/json"
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestMigrateNodeAddressCandidatesBackfillsLegacyPublicIP(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file:"+t.Name()+"?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&Node{}); err != nil {
		t.Fatal(err)
	}
	node := Node{Base: NewBase(), Name: "legacy-ipv6", Status: "offline", AddressMode: "auto", PublicIP: "2001:4860:4860::8888"}
	if err := db.Create(&node).Error; err != nil {
		t.Fatal(err)
	}
	if err := migrateNodeAddressCandidates(db); err != nil {
		t.Fatal(err)
	}
	var loaded Node
	if err := db.First(&loaded, "id = ?", node.ID).Error; err != nil {
		t.Fatal(err)
	}
	var publicIPs, privateIPs []string
	if err := json.Unmarshal(loaded.PublicIPs, &publicIPs); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(loaded.PrivateIPs, &privateIPs); err != nil {
		t.Fatal(err)
	}
	if len(publicIPs) != 1 || publicIPs[0] != node.PublicIP || privateIPs == nil {
		t.Fatalf("backfilled candidates = public %#v, private %#v", publicIPs, privateIPs)
	}
}
