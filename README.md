# SkillSwap

A skill-exchange platform where users list skills they offer, find what they want, request swaps, chat (with optional client-side end-to-end encryption), and place LiveKit video calls.

> **Team:** Skill Issue · **Contact:** namankhandelwal.dev@gmail.com

---

## Stack

- **Backend** — Go 1.25, Gin, GORM, PostgreSQL, WebSockets (port `8080`)
- **Frontend** — Next.js 15 App Router, React 19, TypeScript, Tailwind v4, TanStack Query (port `3000`)
- **Optional** — LiveKit (video), Resend (transactional email), Google OAuth

---

## Run the whole app with Docker

The repo ships with a `docker-compose.yml` that builds Postgres + Go API + Next.js frontend and wires them together. The backend runs DB migrations automatically on startup, so a fresh `up` is enough to get a working stack.

### Prerequisites

- Docker Desktop (or Docker Engine + Compose plugin) — `docker compose version` should print 2.x+
- A `.env` file in the repo root (see template below)

### 1. Create `.env`

```env
# ── Required ────────────────────────────────────────────────
JWT_SECRET=replace-with-a-long-random-string-at-least-32-chars

# ── Postgres (defaults shown — fine for local dev) ──────────
POSTGRES_USER=skillswap
POSTGRES_PASSWORD=skillswap
POSTGRES_DB=skillswap

# ── Optional ────────────────────────────────────────────────
LIVEKIT_API_KEY=
LIVEKIT_API_SECRET=
LIVEKIT_URL=
```

`JWT_SECRET` is the only mandatory variable. The LiveKit triplet is only needed if you want to test video calls; the app boots without it.

### 2. Clean reset + fresh start

If you previously ran the stack and want to throw everything away (database volume included) and rebuild from scratch:

```bash
docker compose down -v --remove-orphans
docker compose build --no-cache
docker compose up -d
```

`-v` wipes the `postgres_data` and `uploads_data` volumes. Skip it if you want to preserve users between rebuilds.

### 3. Day-to-day commands

```bash
# Start (rebuild only when source/Dockerfile changes)
docker compose up -d --build

# Tail logs (Ctrl-C to detach — containers keep running)
docker compose logs -f backend frontend

# Stop without losing data
docker compose down

# Stop AND wipe the database
docker compose down -v
```

### 4. Seed mock data (optional)

A SQL fixture under `backend/skillswap/seeds/mock_data.sql` populates the DB with demo users, skills, and swaps. It runs as a one-shot container behind the `seed` profile, after the backend has created tables:

```bash
docker compose --profile seed up db-seed
```

### 5. Open the app

- Frontend → http://localhost:3000
- API → http://localhost:8080/api/v1
- Health → http://localhost:8080/api/v1/health (if exposed)

### Troubleshooting

| Symptom | Fix |
| --- | --- |
| `JWT_SECRET` missing → backend exits | Add it to `.env` and `docker compose up -d` again |
| `database "skillswap" does not exist` after volume wipe | Wait — Postgres init creates it on first boot; `logs db` to watch |
| Frontend can't reach API | The frontend image bakes `NEXT_PUBLIC_API_BASE_URL` at build time. Rebuild with `docker compose build --no-cache frontend` after changing it |
| Stale DB schema after model changes | `docker compose down -v && docker compose up -d --build` — auto-migrations run on every backend boot |
| Port 3000 / 8080 already in use | Stop the offender, or edit the `ports:` mapping in `docker-compose.yml` |

---

## Run without Docker (native dev)

### Backend

```bash
cd backend/skillswap
export DATABASE_URL=postgres://skillswap:skillswap@localhost:5432/skillswap?sslmode=disable
export JWT_SECRET=replace-me
export PORT=8080
export FRONTEND_URL=http://localhost:3000
export BASE_URL=http://localhost:8080
export UPLOAD_DIR=./uploads
go run ./cmd/server
```

### Frontend

```bash
cd frontend
echo 'NEXT_PUBLIC_API_BASE_URL=http://localhost:8080/api/v1' > .env.local
npm install
npm run dev
```

---

## Repository layout

```
backend/skillswap/   Go API (router → handler → service → repository → model)
frontend/            Next.js App Router app
architecture.md      Authoritative service map + "where to change what" guide
docker-compose.yml   Full-stack orchestration
Dockerfile           Backend image
heroku.yml           Heroku deployment manifest
```

For details on architecture, conventions, and how runtime flows (chat, swaps, E2EE, video) are wired, read **`architecture.md`** first.
