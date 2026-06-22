-- Kafka Engine table to consume logs from various Kafka topics directly into ClickHouse.
-- Design decisions:
-- 1. Kafka Engine: Connects to the log-kafka broker and consumes messages in JSON format.
-- 2. Consumer Group: Uses 'clickhouse-consumer' for consumer offsets tracking.
-- 3. Block Size & Consumers: 4 consumers and 64k block size optimize throughput.
-- 4. Fault Tolerance: Skips up to 100 broken messages per block to prevent blocking the pipeline on malformed JSON.
-- 5. No Indexes/TTL: The Kafka engine doesn't store data persistently; it acts as a stream. Hence, no indexes or TTLs are applied.

CREATE TABLE IF NOT EXISTS logs_queue (
    timestamp DateTime64(3),
    service LowCardinality(String),
    level LowCardinality(String),
    message String,
    http_method LowCardinality(String),
    http_path String,
    http_status UInt16,
    duration_ms Float64,
    trace_id String,
    span_id String,
    parent_span_id String,
    host LowCardinality(String),
    environment LowCardinality(String),
    datacenter LowCardinality(String),
    user_id String,
    correlation_id String,
    request_size UInt32,
    response_size UInt32,
    error_code String,
    stack_trace String
) ENGINE = Kafka()
SETTINGS
    kafka_broker_list = 'log-kafka-kafka-bootstrap:9092',
    kafka_topic_list = 'logs-api-gateway,logs-user-service,logs-payment-service,logs-order-service,logs-inventory-service,logs-notification-service,logs-auth-service,logs-priority',
    kafka_group_name = 'clickhouse-consumer',
    kafka_format = 'JSONEachRow',
    kafka_num_consumers = 4,
    kafka_max_block_size = 65536,
    kafka_skip_broken_messages = 100;
