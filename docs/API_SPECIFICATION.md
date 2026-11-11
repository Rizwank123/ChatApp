# API Specification: ChatApp Social Media Platform

**Version:** 2.0  
**Base URL:** `/api/v1`  
**Protocol:** HTTPS / WSS  
**Authentication:** Bearer JWT in `Authorization` header

---

## 1. Authentication & Sessions

### POST /auth/register
Registers a new user account and associated personnel profile.
- **Request Body:**
  ```json
  {
    "user_name": "jane_doe",
    "password": "StrongPassword123!",
    "first_name": "Jane",
    "last_name": "Doe",
    "email": "jane@example.com",
    "mobile": "+1234567890",
    "gender": "FEMALE"
  }
  ```
- **Response 201 Created:**
  ```json
  {
    "status": true,
    "message": "User registered successfully",
    "data": {
      "user_id": "9b1deb4d-3b7d-4bad-9bdd-2b0d7b3dcb6d",
      "access_token": "eyJhbGciOi...",
      "refresh_token": "d7b3dcb6d9b1deb4d3b7d4bad9bdd2b0...",
      "expires_in": 3600
    }
  }
  ```

### POST /auth/login
Authenticates credentials and returns session tokens.
- **Request Body:**
  ```json
  {
    "user_name": "jane_doe",
    "password": "StrongPassword123!"
  }
  ```
- **Response 200 OK:**
  ```json
  {
    "status": true,
    "data": {
      "access_token": "eyJhbGciOi...",
      "refresh_token": "...",
      "expires_in": 3600
    }
  }
  ```

### POST /auth/refresh
Rotates refresh tokens and issues a new access token.
- **Request Body:** `{"refresh_token": "..."}`
- **Response 200 OK:** `{"access_token": "...", "refresh_token": "...", "expires_in": 3600}`

### POST /auth/logout
Invalidates the current session and revokes the JWT token.
- **Headers:** `Authorization: Bearer <token>`
- **Response 200 OK:** `{"status": true, "message": "Logged out"}`

---

## 2. Personnel & Profiles

### GET /personnel/profile/me
Retrieves current authenticated user's profile with denormalized counters.
- **Response 200 OK:**
  ```json
  {
    "id": "11111111-2222-3333-4444-555555555555",
    "user_name": "jane_doe",
    "first_name": "Jane",
    "last_name": "Doe",
    "avatar": "https://storage.example.com/avatars/jane.jpg",
    "bio": "Software Architect & Photographer",
    "is_private": false,
    "follower_count": 1420,
    "following_count": 312,
    "post_count": 84
  }
  ```

### PATCH /personnel/profile
Updates bio, avatar URL, or privacy settings.
- **Request Body:**
  ```json
  {
    "bio": "Updated bio",
    "is_private": true
  }
  ```

---

## 3. Social Graph (Follows & Blocks)

### POST /personnel/:id/follow
Follows a user, or creates a follow request if the user is private.
- **Response 200 OK:**
  ```json
  {
    "status": true,
    "relationship_status": "FOLLOWING" // or "REQUESTED"
  }
  ```

### POST /personnel/:id/unfollow
Unfollows a user and removes their posts from feed cache.
- **Response 200 OK:** `{"status": true, "relationship_status": "NONE"}`

### GET /personnel/:id/followers
Retrieves followers with cursor-based pagination.
- **Query Params:** `cursor=<uuid>&limit=20`

### GET /personnel/:id/following
Retrieves following list with cursor pagination.

### POST /personnel/:id/block
Blocks a user; immediately severs any mutual follow relationships.

---

## 4. Posts & Engagement

### POST /posts
Creates a new post with uploaded media URLs.
- **Request Body:**
  ```json
  {
    "caption": "Exploring Yosemite National Park! #nature #travel",
    "media_url": [
      "https://storage.example.com/posts/img1.jpg",
      "https://storage.example.com/posts/img2.jpg"
    ],
    "location": "Yosemite Valley",
    "visibility": "PUBLIC"
  }
  ```
- **Response 201 Created:** Post object with generated UUID.

### GET /posts/:id
Fetches a single post by ID including likes and comments count.

### DELETE /posts/:id
Soft-deletes a post; asynchronously invalidates followers' feed caches.

### POST /posts/:id/like
Likes a post. Concurrency-safe, idempotent.
- **Response 200 OK:** `{"liked": true, "like_count": 42}`

### DELETE /posts/:id/like
Removes like from a post.
- **Response 200 OK:** `{"liked": false, "like_count": 41}`

### POST /posts/:id/comments
Adds a comment or nested reply.
- **Request Body:**
  ```json
  {
    "content": "Stunning composition!",
    "parent_id": null
  }
  ```

### GET /posts/:id/comments
Retrieves post comments tree with cursor pagination.

---

## 5. Feed System

### GET /posts/feed
Retrieves the user's home timeline generated via hybrid fanout.
- **Query Params:**
  - `cursor`: Unix timestamp millisecond cursor of last seen item.
  - `limit`: Number of posts (default: 20, max: 50).
- **Response 200 OK:**
  ```json
  {
    "posts": [
      {
        "id": "e3b0c442-98fc-1c14-9afb-4c8996fb9242",
        "author": {
          "id": "11111111-2222-3333-4444-555555555555",
          "user_name": "jane_doe",
          "avatar": "https://storage.example.com/avatars/jane.jpg"
        },
        "caption": "Exploring Yosemite!",
        "media_url": ["https://storage.example.com/posts/img1.jpg"],
        "like_count": 142,
        "comment_count": 18,
        "has_liked": true,
        "created_at": "2026-11-07T18:00:00Z"
      }
    ],
    "next_cursor": "1794074400000",
    "has_more": true
  }
  ```

---

## 6. Notifications

### GET /notifications
Retrieves recent notifications.
- **Query Params:** `limit=30&unread_only=false`

### PATCH /notifications/:id/read
Marks an individual notification as read.

### PATCH /notifications/read-all
Marks all unread notifications as read.

---

## 7. Real-Time WebSockets

### WS /ws
Establishes WebSocket connection.
- **Query / Header Auth:** `token=<jwt_access_token>`
- **Client -> Server Messages:**
  - `{"action": "ping"}`
  - `{"action": "typing", "conversation_id": "..."}`
  - `{"action": "send_message", "conversation_id": "...", "content": "Hello!"}`
- **Server -> Client Events:**
  - `{"event": "pong"}`
  - `{"event": "new_message", "data": {...}}`
  - `{"event": "notification", "data": {"type": "LIKE", "actor": "...", "post_id": "..."}}`
  - `{"event": "presence_update", "user_id": "...", "status": "ONLINE"}`
