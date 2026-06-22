# ADR 003: Vector over Logstash/Fluentd

**Status:** Accepted

## Context
A critical part of the log pipeline is the agent responsible for collecting, parsing, enriching, and routing logs from application nodes to the Kafka buffer. Traditional choices include Logstash and Fluentd. We evaluated Vector as a modern alternative.

## Decision
We decided to use **Vector** as the log collector and processor.

## Reasons
1. **Performance**: Vector is written in Rust, offering up to 10x higher throughput than Logstash (JVM) and Fluentd (Ruby/C), with significantly lower CPU usage.
2. **Memory Footprint**: Vector has a tiny and predictable memory footprint, making it ideal for deployment as a DaemonSet on every Kubernetes node without starving applications of resources.
3. **Reliability**: Vector has robust, built-in backpressure handling and disk buffers, reducing the risk of log loss during downstream outages.
4. **VRL (Vector Remap Language)**: VRL is a custom, highly performant language built into Vector for transforming data. It is safer and faster than writing custom Ruby scripts or complex grok patterns.

## Consequences
- **Ecosystem**: Vector is a newer project compared to Logstash and Fluentd. It has a smaller, albeit growing, plugin ecosystem. If a specific niche integration is required, we may need to write a custom component.
- **Learning Curve**: Teams familiar with Logstash configurations will need to learn Vector's TOML configuration and VRL syntax.
