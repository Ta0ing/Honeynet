package main

import (
	"context"
	"crypto/tls"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/honeynet/honeynet/internal/analytics"
	analyticsch "github.com/honeynet/honeynet/internal/analytics/clickhouse"
	"github.com/honeynet/honeynet/internal/config"
	"github.com/honeynet/honeynet/internal/httpapi"
	"github.com/honeynet/honeynet/internal/nodepki"
	"github.com/honeynet/honeynet/internal/platformservice"
	"github.com/honeynet/honeynet/internal/store"
)

var version = "0.24.0-dev"

func main() {
	configPath := flag.String("config", os.Getenv("HONEYPOT_CONFIG"), "path to server YAML configuration")
	showVersion := flag.Bool("version", false, "print version")
	flag.Parse()
	if *showVersion {
		fmt.Println(version)
		return
	}
	if err := platformservice.Run("HoneynetServer", func(ctx context.Context) error {
		return run(ctx, *configPath)
	}); err != nil {
		log.Fatalf("server stopped: %v", err)
	}
}

func run(ctx context.Context, configPath string) error {
	cfg, err := config.Load(configPath)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}
	cfg.Version = version
	db, err := store.Open(cfg)
	if err != nil {
		return fmt.Errorf("open database: %w", err)
	}
	if err := store.MigrateAndSeed(db, cfg); err != nil {
		return fmt.Errorf("initialize database: %w", err)
	}
	analyticsCfg, err := config.LoadAnalytics(configPath)
	if err != nil {
		return fmt.Errorf("load analytics config: %w", err)
	}
	// Keep the disabled value as a genuinely nil interface. Passing a typed
	// nil *Repository through analytics.Store would make `store != nil` true
	// and panic when compatibility-mode handlers call it.
	var eventStore analytics.Store
	if analyticsCfg.Enabled {
		repository, openErr := analyticsch.Open(analyticsch.Config{
			DSN: analyticsCfg.DSN, Database: analyticsCfg.Database, Table: analyticsCfg.Table,
			MaxOpenConns: analyticsCfg.MaxOpenConns, MaxIdleConns: analyticsCfg.MaxIdleConns,
			ConnMaxLifetime: analyticsCfg.ConnMaxLifetime, DialTimeout: analyticsCfg.DialTimeout, ReadTimeout: analyticsCfg.ReadTimeout,
		})
		if openErr != nil {
			return fmt.Errorf("open ClickHouse analytics: %w", openErr)
		}
		eventStore = repository
		defer repository.Close()
		pingCtx, cancel := context.WithTimeout(ctx, analyticsCfg.DialTimeout)
		err = repository.Ping(pingCtx)
		cancel()
		if err != nil {
			return fmt.Errorf("connect ClickHouse analytics: %w", err)
		}
		schemaCtx, cancelSchema := context.WithTimeout(ctx, analyticsCfg.ReadTimeout)
		err = repository.ValidateSchema(schemaCtx)
		cancelSchema()
		if err != nil {
			return fmt.Errorf("validate ClickHouse analytics schema: %w", err)
		}
		migration, migrationErr := store.MigrateLegacyEvents(ctx, db, repository, store.LegacyEventMigrationOptions{})
		if migrationErr != nil {
			return fmt.Errorf("migrate legacy MySQL events to ClickHouse: %w", migrationErr)
		}
		log.Printf("Honeynet legacy event migration ready (%d events, status %s)", migration.MigratedEvents, migration.Status)
		log.Printf("Honeynet ClickHouse analytics enabled (%s.%s)", analyticsCfg.Database, analyticsCfg.Table)
	} else {
		log.Printf("Honeynet ClickHouse analytics disabled; MySQL legacy event compatibility mode is active")
	}
	authority, err := nodepki.LoadOrCreate(cfg.PKIDir, cfg.AgentTLSNames, cfg.AgentCertValidity)
	if err != nil {
		return fmt.Errorf("initialize node PKI: %w", err)
	}
	runtime := httpapi.NewRuntimeWithAnalytics(ctx, cfg, db, authority, eventStore, analyticsCfg)

	serverLimits := func(addr string, handler http.Handler) *http.Server {
		return &http.Server{
			Addr: addr, Handler: handler,
			ReadHeaderTimeout: 10 * time.Second, ReadTimeout: 30 * time.Second,
			WriteTimeout: 60 * time.Second, IdleTimeout: 90 * time.Second,
			MaxHeaderBytes: 1 << 20,
		}
	}
	consoleServer := serverLimits(cfg.Addr, runtime.Console)
	agentServer := serverLimits(cfg.AgentAddr, runtime.Agent)
	consoleCertFile, consoleKeyFile := "", ""
	if cfg.TLSEnabled {
		if cfg.UsesExternalConsoleCertificate() {
			consoleServer.TLSConfig = externalConsoleTLSConfig()
			consoleCertFile, consoleKeyFile = cfg.TLSCertFile, cfg.TLSKeyFile
		} else {
			consoleServer.TLSConfig = consoleTLSConfig(authority)
		}
	}
	agentServer.TLSConfig = authority.ServerTLSConfig()
	done := make(chan error, 2)
	go func() {
		if cfg.TLSEnabled {
			done <- consoleServer.ListenAndServeTLS(consoleCertFile, consoleKeyFile)
			return
		}
		done <- consoleServer.ListenAndServe()
	}()
	go func() { done <- agentServer.ListenAndServeTLS("", "") }()
	consoleScheme := "http"
	if cfg.TLSEnabled {
		consoleScheme = "https"
	}
	log.Printf("Honeynet Server listening on %s with %s", cfg.Addr, consoleScheme)
	log.Printf("Honeynet Agent mTLS gateway listening on %s (%s)", cfg.AgentAddr, cfg.AgentPublicURL)
	shutdown := func() error {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()
		consoleErr := consoleServer.Shutdown(shutdownCtx)
		agentErr := agentServer.Shutdown(shutdownCtx)
		if consoleErr != nil {
			return consoleErr
		}
		return agentErr
	}
	select {
	case err := <-done:
		_ = shutdown()
		if err == http.ErrServerClosed {
			return nil
		}
		return err
	case <-ctx.Done():
		return shutdown()
	}
}

// consoleTLSConfig deliberately reuses the Server certificate issued by the
// Honeynet CA, but it does not request or accept node client identities. Agent
// authentication remains isolated on the dedicated mTLS gateway.
func consoleTLSConfig(authority *nodepki.Authority) *tls.Config {
	tlsConfig := authority.ServerTLSConfig().Clone()
	tlsConfig.MinVersion = tls.VersionTLS13
	tlsConfig.ClientAuth = tls.NoClientCert
	tlsConfig.ClientCAs = nil
	return tlsConfig
}

// externalConsoleTLSConfig deliberately contains no certificate. The standard
// library loads the operator-provided certificate and key passed to
// ListenAndServeTLS, while the Agent gateway continues to use nodepki on 8443.
func externalConsoleTLSConfig() *tls.Config {
	return &tls.Config{
		MinVersion: tls.VersionTLS13,
		ClientAuth: tls.NoClientCert,
	}
}
