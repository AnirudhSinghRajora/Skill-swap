# SkillSwap Architecture Reference

Last updated: 2026-04-09

## 1. Purpose

This document is a practical architecture map of the current SkillSwap codebase. It is meant to speed up future implementation tasks by answering:

- What runs where?
- Which modules own which behavior?
- Where to edit for common feature requests?
- What constraints and caveats already exist?

This is based on actual code structure and behavior (not only planning docs).

## 2. Repository Topology

Monorepo layout:

- `backend/skillswap`: Go API server (Gin + GORM + PostgreSQL)
- `frontend`: Next.js App Router client (React 19 + TypeScript + React Query)
- Root deployment/dev files: Docker Compose, Dockerfiles, Heroku descriptors

At runtime, primary services are:

- Frontend (port 3000)
- Backend API + WebSocket (port 8080)
- PostgreSQL (port internal to Docker network, external if mapped)

## 3. System Architecture (High Level)

```text
Browser (Next.js app)
  |-- REST (JSON + multipart)
  |-- WebSocket (/api/v1/ws?token=...)
  v
Go Backend (Gin)
  |-- middleware (JWT, CORS, rate limits, security headers)
  |-- handlers -> services -> repositories
  v
PostgreSQL

External integrations:
  - Google OAuth (login)
  - Resend (transactional email)
  - LiveKit (video token generation + media plane)
```

Key product domains:

- Auth and user profiles
- Skills and matching
- Swap lifecycle
- Chat with WebSocket realtime and image sharing
- Optional end-to-end message encryption (client side)
- Notifications (in-app + optional email)
- Ratings and availability
- Admin operations

## 4. Backend Architecture

### 4.1 Boot and Composition

Entrypoint: `backend/skillswap/cmd/server/main.go`

Startup flow:

1. Load env config (`internal/config/config.go`)
2. Initialize DB connection (`internal/database/database.go`)
3. Run migrations/seeding (`database.Migrate`)
4. Build Gin engine and global middleware
5. Register health routes and API routes
6. Start HTTP server on configured port

Global middleware in use:

- Request logging
- Panic recovery
- CORS
- Security headers

### 4.2 Layering Pattern

Primary backend layering:

- Router layer (`internal/router/*_routes.go`): route groups and middleware wiring
- Handler layer (`internal/<domain>/handler.go`): request parsing + response codes
- Service layer (`internal/app/service/*.go`): business rules/orchestration
- Repository layer (`internal/app/repository/*.go`): DB access patterns
- Model layer (`internal/model/*.go`): schema/domain structs and DTOs

This is mostly cleanly applied. Chat and auth have clear service boundaries.

### 4.3 Route Surface (by Domain)

Base prefix: `/api/v1`

- Auth: `/auth/*`
  - register/login/refresh/logout/me
  - google oauth redirect/callback
  - forgot/reset password
  - verify email
- Users: `/users/*`, `/public/users/*`
  - profile and public profile/search
  - E2EE key endpoints (`/me/e2ee-keys`, `/me/key-backup`, `/:id/public-key`)
- Skills: `/skills`, `/users/skills/*`
- Swaps: `/swaps`, `/swaps/matches`, status transitions
- Ratings: `/ratings`, `/users/:id/ratings*`
- Availability: `/availability*`
- Notifications: `/notifications*` (plus admin create)
- Search: `/search/*` public + protected
- File upload: `/files/users/*`
- Chat REST: `/chat/*`
- Video token: `/video/token`
- WebSocket endpoint: `/ws`

### 4.4 Domain Ownership

Auth (`auth_service.go`):

- Local email/password auth and JWT issuance
- Refresh token flow
- OAuth with Google
- Password strength checks
- Password reset token generation/validation
- Email verification token generation/validation

Swaps (`swap_service.go`):

- Validates skill compatibility before creating request
- Controls status transitions (accept/reject/cancel)
- Auto-creates conversation on acceptance (best-effort)
- Emits notifications for request/status changes

Chat (`chat_service.go`, `chat_repository.go`, `internal/chat/*`):

- Conversation fetch/create by swap
- Message send/history/edit/delete
- Unread tracking via `message_read_status.last_read_at`
- Chat image upload/storage/linking
- Swap completion confirmation (`requester_completed`, `responder_completed`)
- WebSocket realtime events: new_message, typing, stop_typing, mark_read, call signaling

Notifications (`notification_service.go`):

- In-app notifications table write/read/update/delete
- Optional email fan-out via Resend service

Video (`video_service.go`, `video/handler.go`):

- Generates LiveKit JWT access token + deterministic room name
- Returns room token/url to frontend

### 4.5 WebSocket Runtime Model

Core files:

- `internal/chat/hub.go`
- `internal/chat/client.go`
- `internal/chat/ws_handler.go`

Model:

- Hub holds `map[user_id]set[connections]`
- One user can have multiple active sockets (multi-tab)
- Client read pump parses inbound frames, write pump sends outbound frames
- `ws_handler` validates token, authorizes participant actions, persists via ChatService, broadcasts through hub

Resilience behavior:

- Slow clients are dropped when buffers fill
- Frontend performs reconnect/backoff and refetches history on reconnect

## 5. Frontend Architecture

### 5.1 App Shell

Entrypoint layout: `frontend/src/app/layout.tsx`

Global wrappers:

- ThemeProvider
- QueryClientProvider (TanStack Query)
- Navigation and top-level toaster
- SmoothScroll helper

### 5.2 Frontend Data Access Pattern

Central API client: `frontend/src/lib/api.ts`

- Single `apiRequest` wrapper for all REST calls
- Injects access token from localStorage
- On 401: tries refresh token once, retries request, otherwise clears auth and redirects to signin
- Exposes grouped domain clients: auth/users/skills/swaps/ratings/notifications/availability/files/chat/e2ee/video

Realtime socket singleton: `frontend/src/lib/websocket.ts`

- One browser-tab singleton
- Exponential reconnect up to 30s
- Event pub/sub dispatch model
- Rebuilds URL with fresh token on reconnect

### 5.3 Frontend Auth State

Hook: `frontend/src/hooks/useAuth.ts`

- Derives trusted identity claims (`user_id`, `email`) by decoding signed JWT payload
- Treats localStorage `user` object as display cache only
- Emits/listens to `auth-change` window event for cross-component updates

### 5.4 Chat UI and Realtime State

Main components/hooks:

- `src/app/messages/page.tsx` split-pane message shell
- `src/components/ConversationList.tsx` conversation sidebar
- `src/components/ChatWindow.tsx` active conversation view
- `src/hooks/useChat.ts` message state machine
- `src/hooks/useWebSocket.ts` connection lifecycle
- `src/hooks/useUnreadCount.ts` unread aggregate

`useChat` responsibilities:

- Initial history load via REST
- Optimistic sends with temp IDs
- WS ack reconciliation and dedup
- Typing indicator management
- REST fallback sends if WS disconnected
- Cursor pagination for older messages
- Read marking + invalidation of unread counters

### 5.5 Client-side E2EE Model

Primary files:

- `src/lib/e2ee/keys.ts`, `crypto.ts`, `backup.ts`, `init.ts`
- `src/hooks/useE2EEKeys.ts`
- `src/hooks/useConversationKeys.ts`

Behavior:

- User keypair stored locally (IndexedDB)
- Public key + encrypted private-key backup stored server-side
- Backup encryption uses user password
- Shared key derived per conversation partner from keypairs
- If shared key exists, message content is encrypted client-side before transport

Important boundary:

- Server persists opaque ciphertext when encrypted mode is used
- Decryption occurs only on client

### 5.6 Video Calling UI

Component: `src/components/VideoCall.tsx`

- Requests video token from backend
- Connects to LiveKit room in fullscreen overlay
- Uses LiveKit React components for conference UI/audio rendering

## 6. Data Architecture

### 6.1 Core Entities

- Users
- Skills
- User offered/wanted skills (junction tables)
- Availability slots
- Swap requests
- Ratings
- Notifications
- Conversations
- Messages
- Message read status
- Chat images
- Password reset tokens
- Email verification tokens

### 6.2 Key Relational Patterns

- One conversation per swap: `conversations.swap_id` unique
- Messages belong to conversations; sender references users
- Unread count is computed using `created_at > last_read_at`
- Chat images can exist before message link, then are attached by `message_id`
- Swap completion now requires both participant flags before `completed` status

### 6.3 Migration and Schema Management Reality

Schema evolution is managed in two overlapping ways:

1. SQL files in `backend/skillswap/migrations/`
2. Runtime migration logic in `internal/database/database.go` (table/column existence checks + SQL)

Notable caveat:

- Migration filenames include duplicate prefixes (`005_*`, `006_*`), and runtime migration code is effectively the active path during app start. Treat `database.go` as source of truth for what auto-applies.

## 7. Security and Access Control

### 7.1 AuthN/AuthZ

- JWT access/refresh tokens (HMAC)
- Route-level JWT middleware for protected APIs
- Admin middleware checks `is_admin` claim
- Chat service/repo enforce participant checks for conversation/message actions

### 7.2 Rate Limiting

- Auth endpoints: strict per-IP middleware
- Chat send/image routes: additional per-user limits
- Implementation is in-memory (single process), not distributed

### 7.3 Input and Content Safety

- Chat HTML is rendered client-side with sanitization (DOMPurify)
- Image size/type validation in chat service (except encrypted blobs treated as opaque)

### 7.4 Current Security Tradeoffs

- WebSocket `CheckOrigin` currently allows all origins (`true`) because token-based auth is used.
- Public chat image endpoint is unauthenticated and relies on unguessable UUIDs.
- Video token endpoint currently generates room token from conversation ID, but handler does not appear to validate caller participation in that conversation.

## 8. Runtime Flows

### 8.1 Authentication Flow

Local login/register:

1. Frontend submits credentials
2. Backend validates + returns access/refresh tokens
3. Frontend stores tokens in localStorage
4. Subsequent API calls use Bearer token; refresh happens automatically on 401

Google OAuth:

1. Browser opens backend `/auth/google`
2. Callback exchanges code and returns tokens via frontend redirect params
3. Callback page stores tokens and redirects to dashboard

### 8.2 Swap to Chat Flow

1. User creates swap request
2. Responder accepts
3. Conversation is auto-created (or lazily created on first chat open)
4. Users exchange messages via WS + REST history endpoints
5. Users mark completion independently; status becomes completed when both confirm

### 8.3 Message Delivery Flow

1. Client sends WS `send_message` with temp ID
2. Backend validates participant + persists message
3. Hub broadcasts to both users
4. Sender reconciles optimistic temp message with persisted message
5. Receiver appends, marks read if open, unread counters update

### 8.4 E2EE Message Flow

1. Client ensures local keypair exists (restore/generate)
2. Client fetches other participant public key
3. Shared key derived locally
4. Message encrypted before send; backend stores ciphertext
5. Receiver decrypts with derived shared key

## 9. Deployment and Environments

### 9.1 Local Docker Compose

Compose services:

- `db` (Postgres 16)
- `backend` (Go service)
- `frontend` (Next.js)
- optional `db-seed` profile for mock data import

Environment highlights:

- Backend requires `DATABASE_URL` (or fallback), `JWT_SECRET`
- Frontend uses `NEXT_PUBLIC_API_BASE_URL`
- Optional integrations via Google/Resend/LiveKit env vars

### 9.2 Heroku

Backend deployment docs and script are provided under `backend/skillswap/DEPLOYMENT.md` and `deploy-heroku.sh`.

## 10. Where To Change What (Task Playbook)

Use this as a quick edit map for future tasks.

Auth behavior:

- Backend logic: `internal/app/service/auth_service.go`
- Routes: `internal/router/auth_routes.go`
- Frontend auth forms/pages: `frontend/src/components/login-form.tsx`, `frontend/src/components/signup.tsx`, `frontend/src/app/auth/*`

User/profile/E2EE endpoints:

- Backend user handler/service/routes: `internal/user/*`, `internal/app/service/user_service.go`, `internal/router/user_routes.go`
- Frontend profile/settings + hooks: `frontend/src/app/profile*`, `frontend/src/app/settings/page.tsx`, `frontend/src/hooks/useE2EEKeys.ts`

Swap lifecycle:

- Backend rules: `internal/app/service/swap_service.go` and chat completion logic in `chat_service.go`
- Routes: `internal/router/swap_routes.go`, `internal/router/chat_routes.go`
- Frontend screens: `frontend/src/app/swaps/page.tsx`

Chat and realtime:

- Backend WS stack: `internal/chat/ws_handler.go`, `internal/chat/hub.go`, `internal/chat/client.go`
- Backend chat service/repo: `internal/app/service/chat_service.go`, `internal/app/repository/chat_repository.go`
- Frontend chat state/ui: `frontend/src/hooks/useChat.ts`, `frontend/src/lib/websocket.ts`, `frontend/src/components/ChatWindow.tsx`, `frontend/src/components/ChatInput.tsx`

Notifications:

- Backend: `internal/app/service/notification_service.go`, `internal/router/notification_routes.go`
- Frontend: `frontend/src/app/notifications/page.tsx`, `frontend/src/components/Navigation.tsx`

Video calls:

- Backend token generation: `internal/app/service/video_service.go`, `internal/video/handler.go`, `internal/router/video_routes.go`
- Frontend call UI: `frontend/src/components/VideoCall.tsx`

Database schema/migrations:

- Runtime migration logic: `backend/skillswap/internal/database/database.go`
- SQL migration artifacts: `backend/skillswap/migrations/*`

## 11. Known Gaps / Risks to Keep in Mind

1. Schema drift risk due to dual migration systems (runtime SQL + migration files).
2. In-memory rate limiting is process-local and resets on restart.
3. WebSocket origin policy is permissive (`CheckOrigin: true`).
4. Chat image endpoint is intentionally public; privacy depends on UUID entropy.
5. Video token issuance appears to miss conversation-participant authorization in handler/service boundary.
6. Email verification link construction should be rechecked against deployed frontend/backend domains.

## 12. Recommended Next Maintenance Steps

1. Consolidate migration strategy to one source of truth.
2. Add explicit participant validation before issuing video tokens.
3. Consider signed/expiring media URLs or auth proxy for chat images.
4. Add integration tests for swap completion + chat + E2EE + video token authorization.
5. Add architecture drift checklist in PR template to keep this file current.

---

If you ask for any future implementation task, this document should be treated as the first context source, then validated against changed code.
