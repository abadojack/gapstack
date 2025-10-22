# Backend architecture design

┌─────────────────┐    ┌──────────────────┐    ┌─────────────────┐
│   Client Apps   │───▶│   API Gateway    │───▶│   Load Balancer │
└─────────────────┘    └──────────────────┘    └─────────────────┘
                            │                          │
┌─────────────────────────────────────────────────────────────┐
│                    Application Layer                        │
└─────────────────────────────────────────────────────────────┘
    │              │              │              │
┌───────┐      ┌───────┐      ┌───────┐      ┌─────────────┐
│Auth & │      │Rate   │      │Idempo-│      │Transaction  │
│AuthZ  │      │Limiter│      │tency  │      │Router       │
└───────┘      └───────┘      └───────┘      └─────────────┘
                                        │
                    ┌─────────────────────────────────┐
                    │        Message Queue            │
                    │    (Apache Kafka/RabbitMQ)      │
                    └─────────────────────────────────┘
                                        │
    ┌─────────────────────────────────────────────────────┐
    │              Transaction Processing                 │
    └─────────────────────────────────────────────────────┘
        │              │              │              │
┌─────────────┐  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐
│   Fraud     │  │   AML       │  │   Business  │  │   Audit     │
│ Detection   │  │   Service   │  │   Rules     │  │   Logger    │
└─────────────┘  └─────────────┘  └─────────────┘  └─────────────┘
        │              │              │              │
┌─────────────────────────────────────────────────────────────┐
│                    Data Layer                               │
└─────────────────────────────────────────────────────────────┘
    │              │              │              │
┌───────┐      ┌───────┐      ┌───────┐      ┌───────┐
│Redis  │      │Postgre│      │Times- │      │S3/    │
│Cache  │      │SQL    │      │caleDB │      │Blob   │
└───────┘      └───────┘      └───────┘      └───────┘

## Technology Choices

### Databases:

* PostgreSQL: Primary transactional store with ACID compliance

* TimescaleDB: For time-series data (audit logs, metrics)

* Redis: For idempotency keys, caching, and rate limiting

### Queues:

* Apache Kafka: Preferred for high throughput, durability, and replay capability

* Alternative: RabbitMQ with mirrored queues

### Caching:

* Redis Cluster: For distributed caching and session storage

* Application-level caching: For frequently accessed reference data

### Communication Protocols:

* REST/gRPC: For synchronous external API calls (AML service)

* WebSocket: For real-time fraud alerts to clients

* HTTP/2: For internal microservice communication

## Idempotency Guarantee
* Client-provided idempotency keys or server-generated deterministic hashes

* Redis distributed locks to prevent race conditions

* Request deduplication at API gateway level

* Idempotent database operations using unique constraints

## AML and Fraud Detection Integration
* AML Check: Synchronous, before transaction acceptance

* Fraud Detection: Both synchronous (basic rules) and asynchronous (complex ML)

* Business Rules: During transaction processing pipeline


## Scaling Approach:

* Horizontal Scaling: Stateless services behind load balancers

* Database Sharding: By customer ID or transaction date

* Caching Strategy: Multi-level caching (Redis + application cache)

* Async Processing: Non-critical path operations deferred

* Circuit Breakers: For external service dependencies

## Monitoring and Observability:

* Distributed tracing with Jaeger/Zipkin

* Real-time metrics with Prometheus

* Structured logging with correlation IDs

* Alerting on SLA violations
