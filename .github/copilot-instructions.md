# SkillSwap — Copilot Instructions

A Skill Swap Platform: users list skills they offer/want, request swaps, chat (with optional client-side E2EE), and place LiveKit video calls.

> **Always read `architecture.md` (repo root) first.** It is the authoritative map of services, layering, routes, runtime flows, and a "Where To Change What" playbook. These instructions only highlight what isn't obvious from a single file or from `architecture.md`.

## Repository layout

Monorepo with two deployable apps:

- `backend/skillswap/` — Go 1.25 / Gin / GORM / PostgreSQL API + WebSocket server (port 8080)
- `frontend/` — Next.js 15 App Router, React 19, TypeScript, TanStack Query (port 3000)
- Root: `docker-compose.yml`, `Dockerfile` (backend), `heroku.yml`, `architecture.md`

## Build / test / lint

### Backend (`backend/skillswap/`)

```bash
go build ./...
go test ./...
go test ./internal/app/service -run TestChatServiceE2EE        # single test
go vet ./...
go run ./cmd/server                                            # local run (needs DATABASE_URL, JWT_SECRET)
```

There is **no separate lint config**; rely on `go vet` and `gofmt`. Tests are sparse — mainly E2EE units in `internal/app/service/`.

### Frontend (`frontend/`)

Uses npm scripts (a `bun.lock` is present but `package-lock.json` is the source of truth in `package.json` scripts):

```bash
npm run dev            # next dev --turbopack
npm run build          # next build
npm run lint           # next lint (ESLint 9 + eslint-config-next)
npm run test           # vitest run (happy-dom env, see vitest.config.ts)
npx vitest run src/lib/e2ee/crypto.test.ts   # single test file
npm run format         # prettier . --write
```

### Full stack

```bash
docker compose up --build                          # db + backend + frontend
docker compose --profile seed up db-seed           # load backend/skillswap/seeds/mock_data.sql
```

## Architecture essentials (read `architecture.md` for full detail)

- **Backend layering:** `router/*_routes.go` → `internal/<domain>/handler.go` → `internal/app/service/*.go` → `internal/app/repository/*.go` → `internal/model/*.go`. Keep this separation; do not call repos from handlers.
- **API base prefix:** `/api/v1`. WebSocket lives at `/api/v1/ws?token=...`.
- **Realtime:** Hub (`internal/chat/hub.go`) maps `user_id → set of connections` (multi-tab supported). `ws_handler` validates JWT + participant membership, then persists via `ChatService` and broadcasts via the hub. Slow clients are dropped when buffers fill.
- **Frontend API client:** All REST calls go through `frontend/src/lib/api.ts` (`apiRequest` wrapper). It auto-refreshes on 401 once, then clears auth and redirects to signin. **Do not call `fetch` directly** for backend endpoints — extend the grouped clients (`auth`/`users`/`skills`/`swaps`/`chat`/`e2ee`/`video`/...).
- **Frontend WebSocket:** Single tab-wide singleton in `frontend/src/lib/websocket.ts` with exponential reconnect (cap 30s) and pub/sub event dispatch. Rebuilds URL with fresh token on reconnect.
- **Auth identity (frontend):** `useAuth` decodes the JWT payload for trusted `user_id`/`email`. The `user` object in localStorage is **display cache only** — never trust it for authorization decisions.

## Key conventions / gotchas

- **Migrations have two sources, and runtime wins.** SQL files live in `backend/skillswap/migrations/` (with duplicate `005_*`/`006_*` prefixes), but the active path on startup is `internal/database/database.go` (`database.Migrate`, table/column existence checks + ad-hoc SQL). When changing schema, update **both** and treat `database.go` as the source of truth for what auto-applies.
- **Swap completion requires both flags.** A swap only becomes `completed` when both `requester_completed` and `responder_completed` are set (logic in `chat_service.go`, not `swap_service.go`).
- **One conversation per swap:** `conversations.swap_id` is unique. Conversations are auto-created on swap acceptance (best-effort) or lazily on first chat open.
- **Unread counts** are computed via `messages.created_at > message_read_status.last_read_at`, not a counter column.
- **Chat images** can exist before being linked to a message; they are attached later by `message_id`. The public chat-image endpoint is intentionally unauthenticated and relies on UUID entropy.
- **Rate limiting** is in-memory and per-process (auth endpoints strict per-IP; chat send/image per-user). It resets on restart and does not coordinate across instances.
- **WebSocket `CheckOrigin` returns true** by design — auth is token-based. Don't "fix" this without also adding origin validation.
- **E2EE is client-side only.** Keys live in IndexedDB; the server stores public keys + a password-encrypted private-key backup and treats encrypted message bodies as opaque ciphertext. Never add server-side decryption. See `frontend/src/lib/e2ee/{keys,crypto,backup,init}.ts`.
- **Chat HTML is rendered with DOMPurify sanitization** on the client. Preserve sanitization when touching `ChatWindow`/`ChatInput`.
- **Notifications:** Always create the in-app notification first; email fan-out via Resend is best-effort and async (goroutine). Don't block request handlers on email.
- **Video tokens** are issued by `video_service.go` using a deterministic room name from conversation ID. Participant authorization at the handler boundary is currently weak — if you touch this code, add the participant check.

## Design system (frontend)

For any UI generation, refactoring, or visual polish work, **invoke the `frontend-design` skill** (and related design skills like `polish`, `arrange`, `typeset`, `colorize`, `audit`) rather than free-styling. The skill carries the project's visual language and anti-patterns and produces output consistent with the rest of the app.

## Environment variables (minimum)

Backend: `DATABASE_URL`, `JWT_SECRET`, `PORT`, `FRONTEND_URL`, `BASE_URL`, `UPLOAD_DIR`. Optional: `GOOGLE_CLIENT_ID`/`SECRET`/`REDIRECT_URL`, `RESEND_API_KEY`/`FROM_EMAIL`, `LIVEKIT_API_KEY`/`API_SECRET`/`URL`.
Frontend: `NEXT_PUBLIC_API_BASE_URL` (e.g. `http://localhost:8080/api/v1`).
