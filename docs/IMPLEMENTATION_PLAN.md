# Implementation Plan: ChatApp Social Media Architecture

**Role:** Principal Backend Engineer & Distributed Systems Architect  
**Project:** ChatApp (Instagram-inspired scalable backend)  
**Date:** November 2026  
**Status:** Approved Roadmap

---

## 1. Executive Summary

ChatApp is an existing Go application leveraging Echo v4, PostgreSQL with pgx/v5 and Squirrel, Redis, and Gorilla WebSocket. While basic models and endpoints exist for users, personnel profiles, direct messaging, posts, comments, likes, and follows, the current implementation has significant architectural and scalability bottlenecks:
- Naive feed query in `internal/repository/post_repository.go` joining follows and posts on the fly.
- In-memory message broker (`internal/broker/memory_broker.go`) lacking persistence, message guarantees, and distributed fanout.
- Missing context propagation and transaction handling in repository and service methods.
- Lack of database indexes, foreign key constraints, and denormalized counters on critical relations.
- Zero test coverage.

This plan details the phased transformation into a production-grade, modular monolith capable of scaling to millions of users with a hybrid feed fanout architecture.

---

## 2. Phase-by-Phase Roadmap

### Phase 1: Analysis & Technical Debt Identification (COMPLETED)
- Repository analysis across all layers (`cmd/`, `internal/domain/`, `internal/repository/`, `internal/service/`, `internal/http/`, `internal/websockets/`, `internal/broker/`).
- Identified missing models: blocks, saved posts, notification system, sessions, user counters, feed inboxes.
- Created `docs/ARCHITECTURE.md` and `docs/DATABASE_DESIGN.md`.

### Phase 2: Architecture Restructuring & Foundations
- Clean modular package structure:
  ```
  internal/
    domain/          # Pure entities, domain errors, value objects
    repository/      # Database interfaces & implementations (PostgreSQL, pgx, SQLC/raw SQL)
    service/         # Business logic orchestration, validation
    http/
      controller/    # HTTP handlers, request/response DTOs
      middleware/    # Auth, logging, rate limiting, metrics, tracing
    events/          # Event definitions, Outbox entity
    broker/          # Redis Streams / Kafka producer and consumer abstractions
    feed/            # Hybrid feed engine (fanout-on-write & fanout-on-read)
    cache/           # Redis client, key management, bloom filters
    observability/   # Prometheus metrics, OpenTelemetry tracing, structured logger
  ```
- Strict Context Propagation: Update all service and repository interfaces to accept `context.Context` as their first parameter.
- Unified error handling with custom domain error types mapping cleanly to HTTP status codes.

### Phase 3: Database Schema Migrations & Storage
- Create Goose migrations:
  - Add missing indexes on `follows(follower_id, following_id)`, `posts(personnel_id, created_at DESC)`, `likes(post_id, personnel_id)`, and `comments(post_id, created_at DESC)`.
  - Add columns to `posts` (`view_count`, `share_count`, `media_metadata JSONB`) and `personnel` (`bio`, `is_private`, `website`).
  - Create `notifications`, `blocks`, `saved_posts`, `follow_requests`, `user_sessions`, `user_counters`, and `feed_cache` tables.
  - Implement Transactional Outbox table (`outbox_events`).

### Phase 4: Core Social Media Features Implementation
- **Auth & Sessions:** JWT access/refresh token rotation, token revocation in Redis, user sessions.
- **Profiles & Privacy:** Private account toggle, follow requests lifecycle (request, accept, reject), blocking and filtering.
- **Posts & Media:** Presigned URL generation for S3 uploads, multi-media support, post deletion cascade, hashtags extraction and search.
- **Engagement:** Concurrency-safe post/comment likes with atomic counter updates, nested comment replies, saved posts.
- **Notifications:** Notification events triggered asynchronously via outbox.

### Phase 5: Instagram-Inspired Scalable Feed System & Probabilistic Filters
- **Hybrid Fanout Engine:**
  - Fanout-on-write for standard users (threshold configurable, default <= 25,000 followers) pushing post IDs to follower feed cache (Redis ZSET + fallback database `feed_cache`).
  - Fanout-on-read for high-follower/celebrity users (> 25,000 followers) dynamically querying recent posts at feed request time.
  - Feed generation aggregator merging ZSET feed cache with active celebrity posts, ranking by recency and engagement score, and supporting cursor-based pagination.
- **Probabilistic Filtering (Cuckoo vs. Bloom Filter):**
  - **Cuckoo Filter:** Selected for dynamic feed seen/hidden filtering and relationship caching due to support for **native item deletion (`CF.DEL`)**, lower false positive rate under low error budgets (<1%), and superior cache locality.
  - **Bloom Filter:** Used for append-only static existence checks (e.g., username availability checks `BF.ADD usernames <name>`).
- **Feed Invalidation & Cache Trimming:** Capping user feed inboxes to recent 800 post IDs; TTL expiration on inactive feeds.

### Phase 6: Event-Driven Processing with Apache Kafka / Redpanda & Transactional Outbox
- Outbox pattern for publishing events atomically with database operations (avoiding dual-write discrepancies).
- Background outbox publisher polling/listening to `outbox_events` and dispatching to **Apache Kafka** / Redpanda topics:
  - `social.posts.created` (12 partitions, key: `author_id`)
  - `social.posts.deleted` (12 partitions, key: `post_id`)
  - `social.engagement.likes` (8 partitions, key: `post_id`)
  - `social.engagement.comments` (8 partitions, key: `post_id`)
  - `social.graph.follows` (8 partitions, key: `following_id`)
  - `social.realtime.notifications` (16 partitions, key: `recipient_id`)
- Worker pool processing:
  - `FanoutWorker`: Distributing post IDs to followers across Kafka consumer groups.
  - `NotificationWorker`: Generating notification records and pushing to WebSocket gateway.
  - `CounterWorker`: Reconciling and backfilling denormalized counters.
- Dead-letter queue (DLQ) and exponential backoff retry strategy.

### Phase 7: Real-Time Communication & Distributed WebSocket Gateway
- Upgrade Gorilla WebSocket hub:
  - Authentication via JWT during connection handshake query param or header.
  - Redis Pub/Sub coordination across multiple server replicas (`ws:user:{recipient_id}`).
  - Heartbeat/ping-pong monitoring (30s intervals), 256-msg write backpressure buffer, per-user connection tracking.
  - Real-time events: chat messages, read receipts, typing indicators, live social notifications.

### Phase 8: Caching & Bloom Filter Optimization
- Redis caching layer:
  - User profile cache with TTL and cache-aside invalidation.
  - Counter caching using Redis hashes with periodic PostgreSQL sync.
- Redis Bloom Filter:
  - Probabilistic "post seen" checks to suppress repetitive feed content for active sessions.
  - Username availability rapid check before hitting PostgreSQL.

### Phase 9: Observability, Metrics & Tracing
- Prometheus metrics: HTTP latency, error rates, feed generation time, fanout queue depth, cache hit/miss ratio, active WebSocket connections.
- OpenTelemetry tracing across HTTP handlers, SQL queries, Redis calls, and asynchronous workers.
- Health check endpoints (`/health/live`, `/health/ready`).
- Provisioned Grafana dashboard configs.

### Phase 10: Testing, Validation & Load Testing
- Comprehensive unit tests for domain and service logic with mock repositories.
- Integration tests using Testcontainers or local test DB for SQL queries and Redis operations.
- Race detector validation (`go test -race ./...`).
- k6 load testing scripts covering user feed retrieval, post creation, and fanout workloads.

---

## 3. Milestones & Acceptance Criteria

| Milestone | Deliverables | Verification Criteria |
|-----------|--------------|-----------------------|
| M1: Architecture Docs | 5 architecture specifications in `docs/` | Complete and aligned with requirements |
| M2: Migrations & Schema | Goose migration files for all new tables and indexes | Migrations run clean up and down on PostgreSQL 15+ |
| M3: Service Refactoring | Context propagation, transaction management, DI | Unit test suite passing with `go test` |
| M4: Hybrid Feed System | Fanout worker, Redis ZSET feed, aggregator | Latency < 50ms for 50-post feed retrieval |
| M5: Real-time & Events | Outbox worker, Redis Pub/Sub WebSocket hub | Messages & notifications delivered real-time across instances |
| M6: Observability & Tests | Prometheus metrics, OpenTelemetry, k6 scripts | Zero race conditions (`-race`), reproducible load tests |
