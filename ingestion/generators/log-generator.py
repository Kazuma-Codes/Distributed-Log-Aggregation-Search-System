#!/usr/bin/env python3
"""
Distributed Log Aggregation & Search System — Log Generator
============================================================
Produces realistic, structured log entries in JSON for testing and
load-generation against Vector / Kafka pipelines.

Usage examples:
    python log-generator.py --rate 100 --format microservice --output stdout
    python log-generator.py --rate 500 --format nginx --output file --output-path /tmp/logs
    python log-generator.py --rate 1000 --format app --output kafka --kafka-brokers kafka:9092
    python log-generator.py --rate 200 --duration 60       # run for 60 seconds
"""

from __future__ import annotations

import argparse
import datetime
import json
import os
import random
import signal
import sys
import threading
import time
import uuid
from collections import deque
from typing import Any, Deque, Dict, List, Optional

# ---------------------------------------------------------------------------
# Optional heavy imports — graceful degradation
# ---------------------------------------------------------------------------
try:
    from kafka import KafkaProducer  # type: ignore
except ImportError:
    KafkaProducer = None  # type: ignore

try:
    from faker import Faker  # type: ignore
    fake = Faker()
except ImportError:
    fake = None  # type: ignore

# ---------------------------------------------------------------------------
# Constants
# ---------------------------------------------------------------------------
SERVICES: List[str] = [
    "api-gateway",
    "user-service",
    "payment-service",
    "order-service",
    "inventory-service",
    "notification-service",
    "auth-service",
]

LEVEL_DISTRIBUTION: List[tuple[str, float]] = [
    ("info", 0.70),
    ("debug", 0.15),
    ("warn", 0.08),
    ("error", 0.05),
    ("critical", 0.02),
]

STATUS_DISTRIBUTION: List[tuple[int, float]] = [
    (200, 0.60),
    (201, 0.15),
    (301, 0.025),
    (302, 0.025),
    (400, 0.08),
    (401, 0.04),
    (403, 0.03),
    (404, 0.03),
    (500, 0.02),
]

HTTP_METHODS = ["GET", "GET", "GET", "POST", "POST", "PUT", "PATCH", "DELETE"]

REQUEST_PATHS: List[str] = [
    "/api/v1/users",
    "/api/v1/users/{id}",
    "/api/v1/users/{id}/profile",
    "/api/v1/orders",
    "/api/v1/orders/{id}",
    "/api/v1/orders/{id}/status",
    "/api/v1/payments",
    "/api/v1/payments/{id}",
    "/api/v1/payments/{id}/refund",
    "/api/v1/inventory",
    "/api/v1/inventory/{id}",
    "/api/v1/inventory/{id}/stock",
    "/api/v1/auth/login",
    "/api/v1/auth/logout",
    "/api/v1/auth/refresh",
    "/api/v1/notifications",
    "/api/v1/notifications/{id}",
    "/api/v1/health",
    "/api/v1/ready",
    "/api/v2/search",
    "/api/v2/search/suggestions",
]

MESSAGES_BY_LEVEL: Dict[str, List[str]] = {
    "info": [
        "Request processed successfully",
        "User authenticated via OAuth2",
        "Order {order_id} created successfully",
        "Cache hit for key {cache_key}",
        "Database query completed in {duration}ms",
        "Health check passed",
        "Connection pool stats: active={active}, idle={idle}",
        "Rate limit check passed for user {user_id}",
        "Message published to queue {queue}",
        "Scheduled job {job} completed",
    ],
    "debug": [
        "Entering handler {handler}",
        "SQL: SELECT * FROM {table} WHERE id = {id}",
        "Cache lookup for key {cache_key}",
        "Request headers: content-type={ct}",
        "Response serialization took {duration}ms",
        "JWT token validated for subject {sub}",
        "Middleware chain: [{middlewares}]",
    ],
    "warn": [
        "Slow query detected: {duration}ms on {table}",
        "Connection pool nearing capacity: {pct}% used",
        "Retry attempt {attempt}/3 for upstream {upstream}",
        "Deprecated API version called: v{version}",
        "Rate limit threshold approaching for IP {ip}",
        "Certificate expires in {days} days",
    ],
    "error": [
        "Failed to connect to database: {error}",
        "Upstream service {upstream} returned 503",
        "Payment processing failed: {reason}",
        "Unhandled exception in {handler}: {error}",
        "Message delivery failed after 3 retries",
        "Circuit breaker OPEN for {upstream}",
    ],
    "critical": [
        "Out of memory: heap usage at {pct}%",
        "Data corruption detected in {table}",
        "All replicas unavailable for partition {partition}",
        "Security: brute-force detected from {ip}",
    ],
}

ERROR_CODES: List[str] = [
    "ERR_DB_CONN_TIMEOUT",
    "ERR_DB_QUERY_FAILED",
    "ERR_UPSTREAM_503",
    "ERR_PAYMENT_GATEWAY",
    "ERR_AUTH_TOKEN_EXPIRED",
    "ERR_RATE_LIMIT",
    "ERR_SERIALIZATION",
    "ERR_OOM",
    "ERR_DISK_FULL",
    "ERR_DATA_INTEGRITY",
]

STACK_TRACES: List[str] = [
    (
        "Traceback (most recent call last):\n"
        '  File "app.py", line 142, in handle_request\n'
        '    result = await db.execute(query)\n'
        '  File "db.py", line 87, in execute\n'
        "    raise ConnectionError('Connection refused')\n"
        "ConnectionError: Connection refused"
    ),
    (
        "Traceback (most recent call last):\n"
        '  File "payment.py", line 63, in process_payment\n'
        '    resp = gateway.charge(amount)\n'
        '  File "gateway.py", line 29, in charge\n'
        "    raise TimeoutError('Gateway timeout after 30s')\n"
        "TimeoutError: Gateway timeout after 30s"
    ),
    (
        "java.lang.NullPointerException\n"
        "    at com.app.service.OrderService.create(OrderService.java:95)\n"
        "    at com.app.controller.OrderController.post(OrderController.java:42)\n"
        "    at sun.reflect.NativeMethodAccessorImpl.invoke0(Native Method)"
    ),
]

HOSTNAMES: List[str] = [
    "node-01.dc1.internal",
    "node-02.dc1.internal",
    "node-03.dc1.internal",
    "node-04.dc2.internal",
    "node-05.dc2.internal",
]

USER_AGENTS: List[str] = [
    "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/125.0.0.0 Safari/537.36",
    "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.5 Safari/605.1.15",
    "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:128.0) Gecko/20100101 Firefox/128.0",
    "curl/8.7.1",
    "python-requests/2.32.3",
    "grpc-go/1.64.0",
]


# ---------------------------------------------------------------------------
# Weighted random helpers
# ---------------------------------------------------------------------------
_level_values, _level_weights = zip(*LEVEL_DISTRIBUTION)
_status_values, _status_weights = zip(*STATUS_DISTRIBUTION)


def _pick_level() -> str:
    return random.choices(_level_values, weights=_level_weights, k=1)[0]


def _pick_status() -> int:
    return random.choices(_status_values, weights=_status_weights, k=1)[0]


def _pick_duration() -> float:
    """Normal distribution, mean 150 ms, std 80 ms, clamped > 0."""
    return round(max(0.1, random.gauss(150, 80)), 2)


def _random_ip() -> str:
    if fake:
        return fake.ipv4_private()
    return f"10.{random.randint(0,255)}.{random.randint(0,255)}.{random.randint(1,254)}"


def _random_user_id() -> str:
    return f"usr_{uuid.uuid4().hex[:8]}"


def _fill_template(template: str) -> str:
    """Replace simple {placeholder} tokens with realistic values."""
    replacements: Dict[str, str] = {
        "order_id": f"ORD-{random.randint(100000,999999)}",
        "cache_key": f"cache:{random.choice(SERVICES)}:{random.randint(1,9999)}",
        "duration": str(round(random.gauss(150, 80), 1)),
        "user_id": _random_user_id(),
        "queue": random.choice(["orders", "payments", "notifications", "events"]),
        "job": random.choice(["cleanup", "report", "sync", "reindex"]),
        "handler": random.choice(["get_user", "create_order", "process_payment"]),
        "table": random.choice(["users", "orders", "payments", "inventory"]),
        "id": str(random.randint(1, 999999)),
        "ct": "application/json",
        "sub": _random_user_id(),
        "middlewares": "auth, ratelimit, logging",
        "pct": str(random.randint(85, 99)),
        "attempt": str(random.randint(1, 3)),
        "upstream": random.choice(SERVICES),
        "version": str(random.randint(1, 2)),
        "ip": _random_ip(),
        "days": str(random.randint(1, 30)),
        "error": random.choice(["connection refused", "timeout", "ECONNRESET"]),
        "reason": random.choice(["insufficient funds", "card declined", "gateway error"]),
        "partition": str(random.randint(0, 11)),
        "active": str(random.randint(5, 50)),
        "idle": str(random.randint(0, 10)),
    }
    result = template
    for key, val in replacements.items():
        result = result.replace("{" + key + "}", val)
    return result


# ---------------------------------------------------------------------------
# Shared trace-ID pool — simulates distributed traces across services
# ---------------------------------------------------------------------------
_trace_pool: Deque[str] = deque(maxlen=50)
_trace_lock = threading.Lock()


def _get_trace_id() -> str:
    """70% reuse an existing trace, 30% start new."""
    with _trace_lock:
        if _trace_pool and random.random() < 0.7:
            return random.choice(list(_trace_pool))
        new_id = str(uuid.uuid4())
        _trace_pool.append(new_id)
        return new_id


def _make_span_id() -> str:
    return uuid.uuid4().hex[:16]


# ---------------------------------------------------------------------------
# Log entry generators
# ---------------------------------------------------------------------------
def _base_entry(service: str, level: str) -> Dict[str, Any]:
    return {
        "timestamp": datetime.datetime.now(datetime.timezone.utc).isoformat(),
        "service": service,
        "level": level,
        "host": random.choice(HOSTNAMES),
    }


def generate_nginx_log() -> Dict[str, Any]:
    method = random.choice(HTTP_METHODS)
    path = random.choice(REQUEST_PATHS).replace("{id}", str(random.randint(1, 99999)))
    status = _pick_status()
    return {
        "remote_addr": _random_ip(),
        "remote_user": "-",
        "time_local": datetime.datetime.now().strftime("%d/%b/%Y:%H:%M:%S %z") or datetime.datetime.now().strftime("%d/%b/%Y:%H:%M:%S +0000"),
        "request": f"{method} {path} HTTP/1.1",
        "status": status,
        "body_bytes_sent": random.randint(0, 65536),
        "http_referer": random.choice(["https://example.com", "-", "https://app.example.com/dashboard"]),
        "http_user_agent": random.choice(USER_AGENTS),
        "request_time": round(_pick_duration() / 1000, 3),
    }


def generate_app_log() -> Dict[str, Any]:
    level = _pick_level()
    service = random.choice(SERVICES)
    entry = _base_entry(service, level)
    entry["message"] = _fill_template(random.choice(MESSAGES_BY_LEVEL[level]))
    entry["trace_id"] = _get_trace_id()
    entry["span_id"] = _make_span_id()
    entry["duration_ms"] = _pick_duration()
    if level in ("error", "critical"):
        entry["error_code"] = random.choice(ERROR_CODES)
        entry["stack_trace"] = random.choice(STACK_TRACES)
    return entry


def generate_microservice_log() -> Dict[str, Any]:
    level = _pick_level()
    service = random.choice(SERVICES)
    trace_id = _get_trace_id()
    span_id = _make_span_id()
    method = random.choice(HTTP_METHODS)
    path = random.choice(REQUEST_PATHS).replace("{id}", str(random.randint(1, 99999)))
    status = _pick_status()

    entry = _base_entry(service, level)
    entry.update(
        {
            "instance_id": f"{service}-{uuid.uuid4().hex[:6]}-{uuid.uuid4().hex[:5]}",
            "message": _fill_template(random.choice(MESSAGES_BY_LEVEL[level])),
            "trace_id": trace_id,
            "span_id": span_id,
            "parent_span_id": _make_span_id() if random.random() > 0.3 else None,
            "http_method": method,
            "http_path": path,
            "http_status": status,
            "duration_ms": _pick_duration(),
            "request_size": random.randint(0, 8192),
            "response_size": random.randint(64, 65536),
            "user_id": _random_user_id() if random.random() > 0.1 else None,
            "correlation_id": f"corr-{uuid.uuid4().hex[:12]}",
        }
    )
    if level in ("error", "critical"):
        entry["error_code"] = random.choice(ERROR_CODES)
        entry["stack_trace"] = random.choice(STACK_TRACES)
    return entry


FORMAT_MAP = {
    "nginx": generate_nginx_log,
    "app": generate_app_log,
    "microservice": generate_microservice_log,
}


# ---------------------------------------------------------------------------
# Output sinks
# ---------------------------------------------------------------------------
class StdoutSink:
    def write(self, record: Dict[str, Any]) -> None:
        sys.stdout.write(json.dumps(record, default=str) + "\n")
        sys.stdout.flush()

    def close(self) -> None:
        pass


class FileSink:
    def __init__(self, path: str) -> None:
        os.makedirs(path, exist_ok=True)
        ts = datetime.datetime.now().strftime("%Y%m%d_%H%M%S")
        self._fp = open(os.path.join(path, f"logs_{ts}.jsonl"), "a", buffering=8192)

    def write(self, record: Dict[str, Any]) -> None:
        self._fp.write(json.dumps(record, default=str) + "\n")

    def close(self) -> None:
        self._fp.flush()
        self._fp.close()


class KafkaSink:
    def __init__(self, brokers: str, topic_prefix: str = "logs") -> None:
        if KafkaProducer is None:
            raise RuntimeError(
                "kafka-python is not installed. "
                "Install it with: pip install kafka-python"
            )
        self._producer = KafkaProducer(
            bootstrap_servers=brokers.split(","),
            value_serializer=lambda v: json.dumps(v, default=str).encode("utf-8"),
            acks="all",
            retries=3,
            linger_ms=10,
            batch_size=16384,
            compression_type="gzip",
        )
        self._topic_prefix = topic_prefix

    def write(self, record: Dict[str, Any]) -> None:
        service = record.get("service", "unknown")
        topic = f"{self._topic_prefix}-{service}"
        self._producer.send(topic, value=record)

    def close(self) -> None:
        self._producer.flush(timeout=10)
        self._producer.close(timeout=10)


# ---------------------------------------------------------------------------
# Generator engine (threaded for high-rate)
# ---------------------------------------------------------------------------
_shutdown_event = threading.Event()


def _signal_handler(signum: int, frame: Any) -> None:  # noqa: ANN401
    print("\n[log-generator] Graceful shutdown requested …", file=sys.stderr)
    _shutdown_event.set()


def _worker(
    sink: Any,
    generator_fn: Any,
    rate: float,
    stats: Dict[str, int],
    lock: threading.Lock,
) -> None:
    """Worker thread that generates logs at the specified rate."""
    interval = 1.0 / rate if rate > 0 else 0
    while not _shutdown_event.is_set():
        t0 = time.monotonic()
        try:
            record = generator_fn()
            sink.write(record)
            with lock:
                stats["total"] += 1
                stats[record.get("level", "info")] = (
                    stats.get(record.get("level", "info"), 0) + 1
                )
        except Exception as exc:
            print(f"[log-generator] Error: {exc}", file=sys.stderr)
            with lock:
                stats["errors"] = stats.get("errors", 0) + 1
        elapsed = time.monotonic() - t0
        sleep_for = max(0, interval - elapsed)
        if sleep_for > 0:
            _shutdown_event.wait(sleep_for)


def run(args: argparse.Namespace) -> None:
    # --- Pick generator format ------------------------------------------------
    generator_fn = FORMAT_MAP.get(args.format)
    if generator_fn is None:
        print(f"Unknown format: {args.format}. Choose from {list(FORMAT_MAP)}", file=sys.stderr)
        sys.exit(1)

    # --- Pick sink ------------------------------------------------------------
    if args.output == "stdout":
        sink = StdoutSink()
    elif args.output == "file":
        sink = FileSink(args.output_path)
    elif args.output == "kafka":
        sink = KafkaSink(args.kafka_brokers, args.kafka_topic_prefix)
    else:
        print(f"Unknown output: {args.output}", file=sys.stderr)
        sys.exit(1)

    # --- Threading strategy ---------------------------------------------------
    num_threads = max(1, min(args.threads, 16))
    rate_per_thread = args.rate / num_threads

    stats: Dict[str, int] = {"total": 0}
    lock = threading.Lock()

    threads: List[threading.Thread] = []
    for i in range(num_threads):
        t = threading.Thread(
            target=_worker,
            args=(sink, generator_fn, rate_per_thread, stats, lock),
            daemon=True,
            name=f"gen-worker-{i}",
        )
        t.start()
        threads.append(t)

    print(
        f"[log-generator] Started {num_threads} worker(s) "
        f"| rate={args.rate}/s | format={args.format} | output={args.output}",
        file=sys.stderr,
    )

    # --- Duration or infinite -------------------------------------------------
    start = time.monotonic()
    try:
        while not _shutdown_event.is_set():
            _shutdown_event.wait(5)
            elapsed = time.monotonic() - start
            with lock:
                total = stats["total"]
            actual_rate = total / elapsed if elapsed > 0 else 0
            print(
                f"[log-generator] {total:,} logs generated "
                f"| {actual_rate:,.1f} logs/s "
                f"| elapsed {elapsed:,.0f}s",
                file=sys.stderr,
            )
            if args.duration > 0 and elapsed >= args.duration:
                print("[log-generator] Duration reached — shutting down.", file=sys.stderr)
                _shutdown_event.set()
    except KeyboardInterrupt:
        _shutdown_event.set()

    # --- Cleanup --------------------------------------------------------------
    for t in threads:
        t.join(timeout=5)

    sink.close()

    with lock:
        print(f"\n[log-generator] Final stats: {json.dumps(stats)}", file=sys.stderr)


# ---------------------------------------------------------------------------
# CLI
# ---------------------------------------------------------------------------
def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(
        description="Realistic log generator for the Distributed Log Aggregation & Search System"
    )
    parser.add_argument(
        "--rate",
        type=float,
        default=100,
        help="Target log generation rate in logs/second (default: 100)",
    )
    parser.add_argument(
        "--format",
        choices=list(FORMAT_MAP),
        default="microservice",
        help="Log format template (default: microservice)",
    )
    parser.add_argument(
        "--output",
        choices=["stdout", "file", "kafka"],
        default="stdout",
        help="Output destination (default: stdout)",
    )
    parser.add_argument(
        "--output-path",
        default="/tmp/generated-logs",
        help="Directory for file output (default: /tmp/generated-logs)",
    )
    parser.add_argument(
        "--kafka-brokers",
        default="log-kafka-kafka-bootstrap:9092",
        help="Kafka bootstrap servers (default: log-kafka-kafka-bootstrap:9092)",
    )
    parser.add_argument(
        "--kafka-topic-prefix",
        default="logs",
        help="Kafka topic prefix; topic = {prefix}-{service} (default: logs)",
    )
    parser.add_argument(
        "--duration",
        type=float,
        default=0,
        help="Run duration in seconds; 0 = infinite (default: 0)",
    )
    parser.add_argument(
        "--threads",
        type=int,
        default=4,
        help="Number of worker threads (default: 4, max: 16)",
    )
    return parser.parse_args()


# ---------------------------------------------------------------------------
# Entry point
# ---------------------------------------------------------------------------
if __name__ == "__main__":
    signal.signal(signal.SIGINT, _signal_handler)
    signal.signal(signal.SIGTERM, _signal_handler)
    run(parse_args())
