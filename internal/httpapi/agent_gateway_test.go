package httpapi

import (
	"testing"

	"github.com/honeynet/honeynet/internal/store"
	"gorm.io/datatypes"
)

func TestPotProtocolTargetIncludesTemplateVersionAndYAML(t *testing.T) {
	template := &store.PotTemplate{
		Base: store.Base{ID: "template-1"}, Name: "Fake OA", Version: 3,
		YAML: "name: fake-oa\nlisten: {port: 8080}\npages: [{path: /}]\n",
	}
	pot := store.PotInstance{
		Base: store.Base{ID: "pot-1"}, ServiceCode: "web-template", Name: "OA portal", Port: 8080,
		Config: datatypes.JSON(`{"bind":"127.0.0.1"}`), DesiredStatus: "running", Template: template,
	}

	target := potProtocolTarget(pot)
	if target.Template == nil || target.Template.ID != template.ID || target.Template.Version != 3 || target.Template.YAML != template.YAML {
		t.Fatalf("potProtocolTarget() template = %#v", target.Template)
	}
	if target.Config["bind"] != "127.0.0.1" {
		t.Fatalf("potProtocolTarget() config = %#v", target.Config)
	}
}

func TestDecoyProtocolTargetIncludesConfig(t *testing.T) {
	item := store.Decoy{Base: store.Base{ID: "decoy-1"}, Name: "Payroll", Type: "file", Status: "enabled", Config: datatypes.JSON(`{"path":"/tmp/payroll.xlsx"}`)}
	target := decoyProtocolTarget(item)
	if target.Config["path"] != "/tmp/payroll.xlsx" || target.Status != "enabled" {
		t.Fatalf("unexpected decoy target: %#v", target)
	}
}
