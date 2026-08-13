package httpapi

import (
	"reflect"
	"testing"

	"github.com/honeynet/honeynet/internal/store"
	"gorm.io/datatypes"
)

func TestNodeSenseProtocolConfig(t *testing.T) {
	item := store.NewNodeSenseConfig("node-1")
	item.Enabled = true
	item.Interface = " eth0 "
	item.DistinctPorts = 6
	item.ExcludedPorts = datatypes.JSON(`[443,22,443]`)
	item.IgnoredCIDRs = datatypes.JSON(`["10.0.0.0/8"]`)

	config, err := nodeSenseProtocolConfig(item)
	if err != nil {
		t.Fatalf("nodeSenseProtocolConfig() error = %v", err)
	}
	if config.Interface != "eth0" || config.DistinctPorts != 6 {
		t.Fatalf("config = %#v", config)
	}
	if !reflect.DeepEqual(config.ExcludedPorts, []int{22, 443}) {
		t.Fatalf("ExcludedPorts = %#v", config.ExcludedPorts)
	}
}

func TestNodeSenseProtocolConfigRejectsCorruptJSON(t *testing.T) {
	item := store.NewNodeSenseConfig("node-1")
	item.ExcludedPorts = datatypes.JSON(`{"invalid":true}`)
	if _, err := nodeSenseProtocolConfig(item); err == nil {
		t.Fatal("nodeSenseProtocolConfig accepted invalid excluded_ports JSON")
	}
}
