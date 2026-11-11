# Scalability Architecture: Hybrid Feed & Distributed Systems Design

**System:** ChatApp Social Platform  
**Target Scale:** 10M+ Users, 100k DAU, Peak Feed Query Load 15,000 QPS  
**Architecture Paradigm:** Modular Monolith with Event-Driven Hybrid Fanout  

---

## 1. The Social Feed Scalability Problem

In social networks like Instagram, the naive query approach:
```sql
SELECT p.* FROM posts p
JOIN follows f ON p.personnel_id = f.following_id
WHERE f.follower_id = $1
ORDER BY p.created_at DESC LIMIT 50;
```
degrades drastically as follower/following counts grow:
1. **Index Scan Exhaustion:** Following hundreds of users means scanning across large B-Trees and merging random post rows.
2. **Offset Pagination Breakdown:** As users scroll deep into their feed (`OFFSET 10000`), Postgres must scan and discard 10,000 rows.
3. **Celebrity Fanout Bottleneck:** If a user with 5 million followers posts, fanout-on-write writes 5 million entries into follower inboxes simultaneously, swamping database writes and memory buffers.

---

## 2. The Hybrid Fanout Architecture

To solve this trade-off between write amplification and read latency, ChatApp implements an **Instagram-style Hybrid Fanout Strategy**:

```
                       [ User Publishes Post ]
                                  |
                                  v
                        [ Store in Posts DB ]
                                  |
                                  v
                       [ Outbox Event Created ]
                                  |
                                  v
                       [ Background Dispatcher ]
                                  |
               +------------------+------------------+
               |                                     |
    (Followers <= Threshold)               (Followers > Threshold)
       Standard User                          Celebrity User
               |                                     |
               v                                     v
       [ Fanout-on-Write ]                   [ Fanout-on-Read ]
               |                                     |
     Push post ID into each              Post stored only in author's
      follower's Redis Feed               timeline; skipped from mass
      Sorted Set (ZSET)                   write fanout
```

### Threshold Configuration
- **Celebrity Threshold:** Default 25,000 followers (dynamically configurable via environment variable `FEED_CELEBRITY_THRESHOLD`).
- When `FollowerCount(author) <= threshold`: Post undergoes **Fanout-on-Write**.
- When `FollowerCount(author) > threshold`: Post undergoes **Fanout-on-Read**.

---

## 3. Data Structures & Feed Inboxes

### Redis Feed Cache (ZSET)
Each active user maintains a feed timeline cached in Redis as a Sorted Set (`feed:{user_id}`):
- **Key:** `feed:{user_id}`
- **Member:** Post UUID (`post_id`)
- **Score:** Unix timestamp in milliseconds (`created_at`)

#### Operations:
- **Write:** `ZADD feed:{follower_id} <timestamp> <post_id>`
- **Inbox Size Cap:** `ZREMRANGEBYRANK feed:{follower_id} 0 -801` (keeps the most recent 800 posts).
- **TTL:** 14 days idle expiration (`EXPIRE feed:{follower_id} 1209600`).
- **Read (Cursor Pagination):**
  ```
  ZREVRANGEBYSCORE feed:{user_id} (<cursor_score> -inf WITHSCORES LIMIT 0 50
  ```

### Relational Fallback (`feed_cache` Table)
For cold/inactive users whose Redis cache has expired:
- Rows are materialized in PostgreSQL `feed_cache(personnel_id, post_id, created_at)`.
- Rehydrated into Redis on demand when the user opens the application.

---

## 4. Feed Retrieval & Aggregation Flow

When a user requests their feed (`GET /api/v1/posts/feed?cursor=...&limit=50`):

```
                   [ Incoming GET Feed Request ]
                                |
                                v
               +----------------+----------------+
               |                                 |
               v                                 v
    [ Query User's Redis ZSET ]       [ Find Followed Celebrities ]
      (Fanout-on-Write items)                    |
               |                                 v
               |                      [ Fetch Recent Posts from ]
               |                      [ Followed Celebrities    ]
               |                                 |
               +----------------+----------------+
                                |
                                v
                     [ Merge & Deduplicate ]
                                |
                                v
                     [ Apply Bloom Filter ]
                     (Filter already seen)
                                |
                                v
                     [ Multi-Get Post Hydration ]
                     (Redis MGET / Postgres SQL)
                                |
                                v
                     [ Return Ranked JSON Page ]
                     (With Next-Cursor)
```

1. **Step 1:** Fetch candidate post IDs from the user's Redis ZSET inbox using cursor score.
2. **Step 2:** Query followed accounts that exceed the celebrity threshold:
   ```sql
   SELECT following_id FROM follows f
   JOIN user_counters uc ON f.following_id = uc.personnel_id
   WHERE f.follower_id = $1 AND uc.follower_count > $2;
   ```
3. **Step 3:** For those celebrities, retrieve their recent post IDs within the cursor time window.
4. **Step 4:** In-memory k-way merge sort of the standard inbox and celebrity posts, ordered by score (`created_at`).
5. **Step 5:** Slice the top `limit` results and extract the score of the last item as `next_cursor`.
6. **Step 6:** Hydrate post details (caption, media URLs, author info, like counts) via Redis post hash (`post:{id}`) with fallback to Postgres `MGET`/`IN (...)`.

---

## 5. Cuckoo Filter vs. Bloom Filter Analysis & Recommendation

### Comparison Matrix

| Criteria | Bloom Filter | Cuckoo Filter (RedisBloom `CF.*`) | Winner & Recommendation |
| :--- | :--- | :--- | :--- |
| **Deletion Support** | ❌ Impossible (standard Bloom); requires counting Bloom filter (4x memory) | ✅ **Native item deletion (`CF.DEL`)** via fingerprint swapping | **Cuckoo Filter** (Critical for social unlikes, deletes, post hiding) |
| **Lookup Performance** | Checks $k$ hash positions across bit array | Checks at most 2 buckets (hash + alternate bucket) | **Cuckoo Filter** (Higher cache locality, faster lookups) |
| **Space Efficiency at Target FPR** | Better when target FPR is large (>3%) | **More space efficient when FPR is low (<3% or <1%)** | **Cuckoo Filter** (For 1% FPR, Cuckoo consumes ~12% less memory) |
| **Insertion When Nearly Full** | Never rejects, but FPR increases smoothly | Can fail when buckets overflow (requires kick-out/rehashing) | **Bloom Filter** for unbounded streams; **Cuckoo Filter** when sized with capacity |
| **Social Media Fit** | Static datasets (e.g. "ever signed up") | Dynamic feeds ("hide post", "unseen", "mark read") | **Cuckoo Filter** |

### Decision & Usage in ChatApp
We choose **Cuckoo Filter** (`redis-stack` / RedisBloom) as the primary probabilistic filter for dynamic features, with Bloom Filter for append-only static checks:
1. **Feed Deduplication & Seen Posts (`CF.ADD seen:{user_id} {post_id}`):** When a user scrolls past a post or dismisses it, it is inserted into their Cuckoo Filter. If a post is deleted by author or hidden by user, `CF.DEL` can cleanly remove it without rebuilding the filter.
2. **Follow Relationship Fast Rejection:** Rapidly filter out non-followers before querying PostgreSQL or Redis Sets.
3. **Username Availability Bloom Filter (`BF.ADD usernames {username}`):** Append-only check to fast-fail username registrations if already taken.

---

## 6. Apache Kafka & Event-Driven Streaming Pipeline

To decouple write requests from heavy fanout and notification distribution, we introduce an **Event-Driven Architecture powered by Kafka / Redpanda**:

```
[ HTTP Handlers ] 
       │
       ▼ (Atomic DB Transaction)
[ PostgreSQL: Core Tables + outbox_events ]
       │
       ▼ (Outbox Poller / CDC Debezium / PgListen)
[ Kafka Producer ] ───▶ [ Kafka Broker Topics ]
                               │
       ┌───────────────────────┼────────────────────────┐
       ▼                       ▼                        ▼
[ Fanout Consumer ]   [ Notification Consumer ]  [ Search/Analytics Consumer ]
(Redis Feed ZSETs)     (Push to WebSocket Hub)    (Postgres FTS / Elastic)
```

### Kafka Topic Definitions & Partitions

| Topic Name | Partitions | Key | Value Schema | Consumers |
| :--- | :--- | :--- | :--- | :--- |
| `social.posts.created` | 12 | `author_id` (guarantees per-author ordering) | `PostCreatedEvent` (Post ID, Author ID, Media, Timestamp) | `fanout-workers`, `notification-workers`, `search-indexer` |
| `social.posts.deleted` | 12 | `post_id` | `PostDeletedEvent` (Post ID, Author ID) | `feed-cleanup-workers`, `cache-invalidator` |
| `social.engagement.likes` | 8 | `post_id` | `PostLikedEvent` (Post ID, Actor ID, Target Personnel ID) | `counter-workers`, `notification-workers` |
| `social.engagement.comments`| 8 | `post_id` | `CommentCreatedEvent` (Comment ID, Post ID, Actor ID, Parent ID) | `notification-workers` |
| `social.graph.follows` | 8 | `following_id` | `UserFollowedEvent` (Follower ID, Following ID, Status) | `counter-workers`, `feed-backfill-workers`, `notification-workers` |
| `social.realtime.notifications`| 16 | `recipient_id` | `NotificationDispatchedEvent` | `websocket-gateway-workers` |

### Idempotency & Consumer Guarantees
- **At-Least-Once Delivery + Idempotency:** Each consumer uses Redis `SET {event_id}:processed EX 86400 NX` to eliminate duplicate processing.
- **Dead-Letter Queue (DLQ):** Messages failing 3 consecutive retries are sent to `*.dlq` topic with diagnostic headers.

---

## 7. Distributed WebSocket & Real-Time Gateway Architecture

WebSocket connections terminate on a horizontally scalable gateway and coordinate state across multiple instances using Redis Pub/Sub:

```
[ Mobile / Web Client ] ◄──(WSS)──► [ Echo/Gorilla WS Hub Node 1 ]
[ Mobile / Web Client ] ◄──(WSS)──► [ Echo/Gorilla WS Hub Node 2 ]
                                                │
                                                ▼ (Pub/Sub)
                                    [ Redis Pub/Sub Channel ]
                                       `ws:user:{recipient_id}`
```

1. **Authentication:** JWT validated on handshake query `wss://domain.com/ws?token=...`.
2. **Connection Registry:** Thread-safe per-user map (`map[uuid.UUID]map[*Client]bool` with `sync.RWMutex`).
3. **Multi-Node Broadcasting:** When a notification or direct message occurs, the event is published to Redis channel `ws:user:{recipient_id}`. All application instances listen; only the instance holding the active socket forwards the payload.
4. **Heartbeats & Backpressure:** Ping/Pong ticker every 30s. Write buffer queue per client (buffered channel of 256 messages) with connection dropping on buffer overflow to prevent slow-consumer memory bloat.

To eliminate database-broker dual write race conditions, all state changes write to PostgreSQL transactionally before publishing:

```sql
BEGIN;
  INSERT INTO posts (id, personnel_id, caption, media_url) VALUES (...);
  INSERT INTO outbox_events (id, aggregate_type, aggregate_id, event_type, payload)
  VALUES (gen_random_uuid(), 'POST', post_id, 'PostCreated', '{"post_id":"...", "author_id":"..."}');
COMMIT;
```

### Outbox Dispatcher
- Dedicated background worker tailing `outbox_events` via cursor or `LISTEN/NOTIFY`.
- Publishes to Redis Streams / Kafka topic `social.posts.events`.
- Marks outbox row as `published_at = NOW()`.

### Fanout Consumer Worker Pool
- Parallel worker goroutines consume `PostCreated` events.
- Checks author's follower count:
  - If `<= 25,000`: Batches follower IDs in chunks of 1,000 and executes Redis pipeline `ZADD` commands.
  - If `> 25,000`: Skips follower fanout, author marked as celebrity.
- Idempotency guaranteed via Redis `SETNX` on event processing key.

---

## 7. Performance Benchmarks & SLOs

| Operation | Target Latency (p95) | Target Latency (p99) | Max Throughput |
|-----------|----------------------|----------------------|----------------|
| GET Feed (first 50 posts) | < 30ms | < 60ms | 15,000 QPS |
| Post Creation (API response) | < 45ms | < 80ms | 2,500 QPS |
| Fanout 20k Followers | < 500ms (async) | < 1,500ms (async) | 500 posts/sec |
| Feed Memory per Active User | ~16 KB (800 post IDs in ZSET) | < 25 KB | 1M users = 16GB Redis |
