# ADR 001: Kafka 4.x (KRaft) over Redpanda

**Status:** Accepted

## Context
We need a highly available, high-throughput streaming platform to buffer logs before they are ingested into ClickHouse. This buffer is critical to prevent data loss when ClickHouse is undergoing maintenance or experiencing a spike in traffic. The system targets 100K+ logs/second. We evaluated Apache Kafka 4.x (KRaft mode) and Redpanda.

## Decision
We decided to use **Apache Kafka 4.x in KRaft mode**.

## Reasons
1. **License**: Apache Kafka is licensed under the true Apache 2.0 license. Redpanda uses the Business Source License (BSL), which restricts certain use cases and requires transition to open source only after several years. Our project aims to be completely open-source without enterprise restrictions.
2. **KRaft Mode**: Kafka 4.x eliminates the ZooKeeper dependency, which historically made Kafka hard to manage. KRaft provides an integrated consensus mechanism, significantly simplifying the operational burden.
3. **Ecosystem**: Kafka has a massive, mature ecosystem of connectors, clients, and operational tools (e.g., Strimzi operator for Kubernetes).
4. **Proven at Scale**: Kafka has been battle-tested for over a decade at massive scale (LinkedIn, Netflix, etc.).

## Consequences
- **Resource Usage**: Kafka is JVM-based, which generally requires higher memory allocation and careful tuning compared to Redpanda's C++ Seastar architecture.
- **Operations**: We will rely on the Strimzi Kubernetes Operator to manage the Kafka clusters, which introduces a dependency on a third-party operator.
