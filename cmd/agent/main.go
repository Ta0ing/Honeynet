package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"time"

	agentclient "github.com/honeynet/honeynet/internal/agent/client"
	agentconfig "github.com/honeynet/honeynet/internal/agent/config"
	"github.com/honeynet/honeynet/internal/agentupdate"
	"github.com/honeynet/honeynet/internal/platformservice"
)

var version = "0.24.0-dev"

func main() {
	configPath := flag.String("config", agentconfig.DefaultPath(), "path to agent configuration")
	server := flag.String("server", "", "Honeynet Server URL")
	agentURL := flag.String("agent-url", "", "mTLS Agent gateway URL")
	nodeID := flag.String("node-id", "", "node ID")
	registrationToken := flag.String("registration-token", "", "one-time registration token")
	forceEnroll := flag.Bool("force-enroll", false, "discard an existing node certificate and use the supplied registration token")
	agentToken := flag.String("agent-token", "", "existing agent credential")
	caSHA256 := flag.String("ca-sha256", "", "Agent gateway CA SHA-256 fingerprint")
	caCert := flag.String("ca-cert", "", "path to trusted Agent gateway CA certificate")
	tlsServerName := flag.String("tls-server-name", "", "override Agent gateway TLS server name")
	stateDir := flag.String("state-dir", "", "state directory")
	templateRoot := flag.String("template-root", "", "honeypot-templates-server services directory")
	insecureTLS := flag.Bool("insecure-tls", false, "allow an untrusted Server TLS certificate (development only)")
	initOnly := flag.Bool("init-only", false, "write configuration and exit")
	enrollOnly := flag.Bool("enroll-only", false, "enroll the node certificate and exit")
	renewCertificate := flag.Bool("renew-certificate", false, "renew the current node certificate and exit")
	showVersion := flag.Bool("version", false, "print version")
	updateHelper := flag.Bool("update-helper", false, "internal Windows update helper")
	updateSource := flag.String("update-source", "", "internal update source")
	updateTarget := flag.String("update-target", "", "internal update target")
	updateState := flag.String("update-state", "", "internal update state")
	updateService := flag.String("update-service", "HoneynetAgent", "internal update service")
	updateParent := flag.Int("update-parent", 0, "internal update parent process")
	flag.Parse()
	if *showVersion {
		fmt.Println(version)
		return
	}
	if *updateHelper {
		if err := agentupdate.RunUpdateHelper(*updateSource, *updateTarget, *updateState, *updateService, *updateParent); err != nil {
			log.Fatalf("apply Agent update: %v", err)
		}
		return
	}
	cfg, err := agentconfig.Load(*configPath)
	if err != nil {
		log.Fatalf("load config: %v", err)
	}
	if *server != "" {
		cfg.ServerURL = *server
	}
	if *agentURL != "" {
		cfg.AgentURL = *agentURL
	}
	if *nodeID != "" {
		cfg.NodeID = *nodeID
	}
	if *registrationToken != "" {
		if *forceEnroll {
			cfg.ClientCertPath = ""
			cfg.ClientKeyPath = ""
			cfg.CertificateExpiry = time.Time{}
		}
		if !cfg.HasCertificate() {
			cfg.RegistrationToken = *registrationToken
		}
	}
	if cfg.HasCertificate() {
		cfg.RegistrationToken = ""
	}
	if *agentToken != "" {
		cfg.AgentToken = *agentToken
	}
	if *caSHA256 != "" {
		cfg.CAFingerprint = *caSHA256
	}
	if *caCert != "" {
		cfg.CACertPath = *caCert
	}
	if *tlsServerName != "" {
		cfg.TLSServerName = *tlsServerName
	}
	if *stateDir != "" {
		cfg.StateDir = *stateDir
	}
	if *templateRoot != "" {
		cfg.TemplateRoot = *templateRoot
	}
	if *insecureTLS {
		cfg.InsecureTLS = true
	}
	cfg.ConfigPath = *configPath
	if err := cfg.Normalize(); err != nil {
		log.Fatalf("config: %v", err)
	}
	if err := cfg.Save(); err != nil {
		log.Fatalf("save config: %v", err)
	}
	if *initOnly {
		log.Printf("Honeynet Agent configuration initialized at %s", cfg.ConfigPath)
		return
	}
	client, err := agentclient.New(&cfg, version)
	if err != nil {
		log.Fatalf("initialize agent: %v", err)
	}
	if *enrollOnly {
		if err := client.Enroll(context.Background()); err != nil {
			log.Fatalf("enroll agent: %v", err)
		}
		log.Printf("Honeynet Agent certificate enrolled at %s", cfg.ClientCertPath)
		return
	}
	if *renewCertificate {
		if err := client.RenewCertificate(context.Background()); err != nil {
			log.Fatalf("renew certificate: %v", err)
		}
		return
	}
	log.Printf("Honeynet Agent %s starting for node %s", version, cfg.NodeID)
	if err := platformservice.Run("HoneynetAgent", client.Run); err != nil {
		log.Fatalf("agent stopped: %v", err)
	}
}
