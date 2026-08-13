-- Small, replaceable rollups for the first stable dashboard queries. They can
-- be rebuilt from security_events; no mutable workflow state is stored here.
CREATE MATERIALIZED VIEW IF NOT EXISTS security_events_daily_mv
ENGINE = AggregatingMergeTree
PARTITION BY toYYYYMM(day)
ORDER BY (day, node_id, service, event_type)
AS
SELECT
    toStartOfDay(event_time) AS day,
    node_id,
    service,
    event_type,
    uniqExactState(event_id) AS event_ids
FROM security_events
GROUP BY day, node_id, service, event_type;
