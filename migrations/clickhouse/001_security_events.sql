-- Honeynet analytical security-event fact table.
-- MySQL remains the control-plane/business database.
CREATE TABLE IF NOT EXISTS security_events
(
    event_id String,
    node_id String,
    pot_id String,
    decoy_id String,
    service LowCardinality(String),
    event_type LowCardinality(String),
    event_time DateTime64(3, 'UTC'),
    ingested_at DateTime64(3, 'UTC'),
    src_ip String,
    src_port UInt16,
    dst_ip String,
    dst_port UInt16,
    geo LowCardinality(String),
    asn LowCardinality(String),
    raw_packet String CODEC(ZSTD(3)),
    payload String CODEC(ZSTD(3)),
    tags String CODEC(ZSTD(3)),
    -- Exact original Agent + Server hit objects for forensic/audit display.
    detections String CODEC(ZSTD(3)),
    agent_rule_revision Int64,
    server_rule_revision Int64,
    session_id String,
    has_credential UInt8,
    credential_username String CODEC(ZSTD(3)),
    credential_password String CODEC(ZSTD(3)),
    credential_auth_response String CODEC(ZSTD(3)),
    credential_mechanism LowCardinality(String),
    record_version UInt64,
    INDEX idx_event_id event_id TYPE bloom_filter(0.001) GRANULARITY 1,
    INDEX idx_src_ip src_ip TYPE bloom_filter(0.001) GRANULARITY 4,
    INDEX idx_session session_id TYPE bloom_filter(0.001) GRANULARITY 4
)
ENGINE = ReplacingMergeTree(record_version)
PARTITION BY toYYYYMM(event_time)
ORDER BY (event_time, node_id, service, event_type, event_id)
TTL event_time + INTERVAL 365 DAY DELETE
SETTINGS index_granularity = 8192;

