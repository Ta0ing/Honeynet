VERSION ?= 0.24.0-dev
export GOTOOLCHAIN ?= go1.26.5
TARGET_OS ?= linux
TARGET_ARCH ?= amd64

.PHONY: dev-server dev-agent dev-web build build-server build-agent agents release release-server release-server-only release-agent release-agents release-all test smoke-pot-control clickhouse-install clickhouse-migrate clickhouse-smoke clickhouse-uninstall

dev-server:
	go run ./cmd/server

dev-web:
	cd web && npm run dev

dev-agent:
	go run ./cmd/agent --config ./tmp/agent.json

build: build-server build-agent

build-server:
	cd web && npm run build
	go build -ldflags "-X main.version=$(VERSION)" -o bin/honeynet-server ./cmd/server

build-agent:
	go build -ldflags "-X main.version=$(VERSION)" -o bin/honeynet-agent ./cmd/agent

agents:
	./scripts/build-agent-downloads.sh $(VERSION)

release: release-server

release-server:
	./scripts/build-release.sh $(VERSION) $(TARGET_OS) $(TARGET_ARCH)

release-server-only:
	./scripts/build-release.sh $(VERSION) $(TARGET_OS) $(TARGET_ARCH) server-only

release-agent:
	./scripts/build-agent-release.sh $(VERSION) $(TARGET_OS) $(TARGET_ARCH)

release-agents:
	./scripts/build-all-agent-releases.sh $(VERSION)

release-all: release-server release-agent

test:
	go test ./...
	./scripts/test-network-addresses.sh
	cd web && npm run lint

smoke-pot-control:
	node scripts/smoke-pot-control.mjs

clickhouse-install:
	sudo ./scripts/install-clickhouse.sh

clickhouse-migrate:
	sudo ./scripts/migrate-clickhouse.sh

clickhouse-smoke:
	sudo ./scripts/smoke-clickhouse.sh

clickhouse-uninstall:
	sudo ./scripts/uninstall-clickhouse.sh
