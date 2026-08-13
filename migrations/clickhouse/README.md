# ClickHouse migrations

These migrations create Honeynet's analytical security-event store. They do
not replace or mutate the MySQL business database.

Apply files in lexical order with a least-privilege migration account, for
example:

```sh
clickhouse-client --database honeynet_analytics --multiquery < migrations/clickhouse/001_security_events.sql
clickhouse-client --database honeynet_analytics --multiquery < migrations/clickhouse/002_event_rollups.sql
```

The application writer needs `INSERT` on `security_events`; query users need
`SELECT`. The table uses both deterministic insert deduplication tokens and
`ReplacingMergeTree(record_version)`. Initial query methods use `FINAL`, so
retries are invisible before background merges complete. The HTTP ingestion layer must not acknowledge an Agent
event until `InsertEvents` succeeds or a durable Server-side WAL has accepted
the event.
