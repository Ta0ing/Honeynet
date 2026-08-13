package client

import (
	"runtime"
	"testing"
)

func TestAgentCapabilitiesIncludeImplementedPots(t *testing.T) {
	set := map[string]bool{}
	for _, capability := range agentCapabilities() {
		set[capability] = true
	}
	for _, capability := range []string{
		"pot.bacnet", "pot.coap", "pot.dns", "pot.elasticsearch", "pot.ftp", "pot.http", "pot.https", "pot.imap", "pot.imaps", "pot.kafka",
		"pot.ldap", "pot.ldaps", "pot.memcached", "pot.modbus", "pot.mongodb", "pot.mqtt", "pot.mssql", "pot.mysql",
		"pot.oracle", "pot.pop3", "pot.pop3s", "pot.postgresql", "pot.rdp", "pot.redis", "pot.rtsp-camera", "pot.smb",
		"pot.s7comm", "pot.smtp", "pot.smtps", "pot.snmp", "pot.ssh", "pot.telnet", "pot.tftp", "pot.vnc", "pot.web-template",
		"pot.zookeeper",
		"decoy.file",
		"network.ipv6",
	} {
		if !set[capability] {
			t.Fatalf("agentCapabilities() is missing %s", capability)
		}
	}
	if set["sense.passive"] != (runtime.GOOS == "linux") {
		t.Fatalf("sense.passive availability does not match %s", runtime.GOOS)
	}
}
