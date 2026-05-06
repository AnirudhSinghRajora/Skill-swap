# Frontend Rework Plan

## Problem Summary

The frontend is essentially a **UI mockup** — every page past auth uses dummy data, all actions log to console, there's no route protection, no API service layer, and the types/features don't align with the backend. The brand is also inconsistently called "SkillShare" (a real company's name).

---

## Phase 1: Foundation (must happen first)

| # | Task | Why |
|---|------|-----|
| **1.1** | **Create an API service layer** (`src/lib/api.ts`) | There's zero centralized API logic. Every page will need this. Build a client with base URL config, automatic token attachment, refresh token handling, and typed response wrappers. Use `NEXT_PUBLIC_API_BASE_URL` consistently (fix the `GO_BACKEND_URL` mismatch in `auth.ts`). |
| **1.2** | **Fix auth flow & route protection** | Right now protected pages have no guards — anyone can visit `/dashboard`. Add a `useAuth` hook that reads from the real token in localStorage + validates with `/auth/me`. Add a `ProtectedRoute` wrapper or Next.js middleware that redirects unauthenticated users to `/auth/signin`. |
| **1.3** | **Fix type definitions** to match backend responses | Frontend types use `userId`, `skillId` etc. but backend returns `user_id`, `skill_id` (snake_case). Types need a `Notification` type with all backend fields (`related_id`, `title`, `message`, `type` enum matching backend). `SwapRequest` needs `offered_skill_id`/`wanted_skill_id` references. Add response wrapper types (`PaginatedResponse<T>`, `ApiError`). |
| **1.4** | **Remove duplicate registration flow** | There are THREE auth entry points: `/auth/signin`, `/auth/signup`, and `/register`. The `/register` multi-step flow uses dummy skills and logs to console. Consolidate into: `/auth/signup` for account creation, then a **post-signup onboarding** step for skills/profile (which calls the real profile + skills APIs). Remove or redirect `/register`. |
| **1.5** | **Fix branding** | Components say "SkillShare" and "SkillShare Connect" in various places — this is a trademarked name. Rename consistently to **"SkillSwap"** across Hero, LoadingAnimation, Navigation, CTA. |

---

## Phase 2: Core Pages — Wire Up to Backend

| # | Task | Current State → Target State |
|---|------|------------------------------|
| **2.1** | **Dashboard** | Uses `getUserById("1")` + dummy helpers → Call `GET /api/v1/users/profile`, `GET /api/v1/swaps?limit=5`, `GET /api/v1/notifications?limit=3` via React Query. Show real stats (compute from swap list). Loading/error states. |
| **2.2** | **Browse** | Uses `dummyUsers` + `dummySkills` → Call `GET /api/v1/public/users/search` with query params for search, location, skills. Call `GET /api/v1/skills` for filter options. Add pagination (backend supports `page` + `limit`). |
| **2.3** | **Profile** | Hardcoded user "1" → Call `GET /api/v1/users/profile` for own profile. Ratings tab calls `GET /api/v1/users/{id}/ratings` + `GET /api/v1/users/{id}/ratings/stats`. Swaps tab calls `GET /api/v1/swaps`. Add edit button that opens an inline edit form calling `PUT /api/v1/users/profile`. |
| **2.4** | **Swaps** | Dummy swap data, console.log actions → Call `GET /api/v1/swaps?sent=true&received=true`. Accept/reject/cancel call `PUT /api/v1/swaps/{id}/status`. Add a "New Swap Request" flow using `GET /api/v1/swaps/matches` + `POST /api/v1/swaps`. Invalidate queries on mutation. |
| **2.5** | **Skills management** | Only exists as dummy badges in profile → Add a dedicated skills section (on profile or onboarding) calling `GET /api/v1/skills`, `POST /api/v1/users/skills/offered`, `POST /api/v1/users/skills/wanted`, `DELETE` endpoints. Add a skill search/autocomplete from master list. |
| **2.6** | **Notifications** | Only shown as a badge count in Navigation from dummy data → Call `GET /api/v1/notifications` in nav for count, create a notification dropdown/page. Wire mark-read (`PUT /api/v1/notifications/mark-read`), mark-all-read, delete. |
| **2.7** | **Settings** | All console.log → Profile tab calls `PUT /api/v1/users/profile`. Password tab calls appropriate auth endpoint. Photo upload calls `POST /api/v1/files/users/photo`. Remove notification/privacy toggle sections that the backend doesn't support (no endpoints for those). |

---

## Phase 3: Remove / Rethink Non-functional Features

| # | Task | Reasoning |
|---|------|-----------|
| **3.1** | **Remove Messages page entirely** | The backend has **zero messaging endpoints**. No message model, no message table, no API. This page is purely dummy data with fake conversations. Remove it from nav and routing. If messaging is wanted later, build the backend first. |
| **3.2** | **Remove `dummy-data.ts`** | Once all pages use real API calls, this file should be deleted. It's the root cause of the "looks like it works but doesn't" problem. |
| **3.3** | **Simplify landing page animations** | The 3-second forced loading screen (`LoadingAnimation`) with fake progress bar is poor UX — users stare at a spinner before seeing content. Remove the artificial delay. Keep GSAP scroll animations but remove the blocking loader. |
| **3.4** | **Rethink About page stats** | Currently shows hardcoded "1000+ users, 500+ skills" etc. Either remove fake stats or source them from a future public stats endpoint. For now, replace with descriptive content about how the platform works (no fake numbers). |

---

## Phase 4: Missing Features to Add

| # | Task | Why |
|---|------|-----|
| **4.1** | **Availability management UI** | Backend has full availability CRUD + overlap detection (`GET /api/v1/availability/common/{user_id}`). No frontend UI exists. Add an availability section to Profile/Settings where users set their time slots, and show common availability when viewing another user's profile. |
| **4.2** | **Rating flow after swap completion** | Backend supports `POST /api/v1/ratings`. Frontend shows ratings read-only. Add a "Rate this swap" prompt when a swap is marked completed. |
| **4.3** | **Public user profile page** (`/users/[id]`) | Currently only `/profile` exists (own profile). Need a route to view OTHER users' profiles (reached from Browse/UserCard). Would call public user search or a user detail endpoint. Show their skills, ratings, availability. Add "Request Swap" button. |
| **4.4** | **Photo upload** | Backend supports `POST /api/v1/files/users/photo` (multipart, 5MB max). Settings has a "Change Photo" button but it does nothing. Wire it up with a file picker and preview. |
| **4.5** | **Swap request creation flow** | There's no UI to actually create a swap request. When viewing another user's profile, add a "Request Swap" button that shows a modal: pick which skill you're offering, which of theirs you want, and submit via `POST /api/v1/swaps`. |

---

## Phase 5: Polish

| # | Task |
|---|------|
| **5.1** | Add proper loading skeletons for all data-fetching pages |
| **5.2** | Add error boundaries and toast notifications for API failures |
| **5.3** | Add empty states (e.g. "No swaps yet — browse users to get started") |
| **5.4** | Make Navigation highlight active route |
| **5.5** | Mobile responsiveness audit (some pages look off on small screens) |

---

## Execution Order

```
Phase 1 (Foundation)  →  Phase 2 (Wire up pages)  →  Phase 3 (Remove dead code)  →  Phase 4 (New features)  →  Phase 5 (Polish)
```

Phase 1 must be done first since everything else depends on the API layer and auth. Phases 2 and 3 can overlap. Phase 4 adds net-new functionality. Phase 5 is final pass.
