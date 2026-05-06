# Backend Error Handling Audit & Fix Plan

## Summary

Full audit of all Go files in `backend/skillswap/internal/`. Found **6 categories** of issues across handlers, services, middleware, and database layers.

---

## P0 — Panic-Prone Type Assertions (Runtime Crashes)

The JWT middleware stores `user_id` as a `string` via `c.Set("user_id", claims["user_id"])`. Several handlers assert it directly as `uuid.UUID`, causing panics.

### Already Fixed
- `availability/handler.go` — all 7 occurrences
- `rating/handler.go` — `CreateRating`, `UpdateRating`, `DeleteRating`

### Still Broken — `admin/handler.go` (6 occurrences)
| Line | Function | Broken Code |
|------|----------|-------------|
| 117 | `BanUser` | `adminID.(uuid.UUID)` |
| 148 | `UnbanUser` | `adminID.(uuid.UUID)` |
| 179 | `DeleteUser` | `adminID.(uuid.UUID)` |
| 215 | `MakeUserAdmin` | `adminID.(uuid.UUID)` |
| 246 | `RemoveUserAdmin` | `adminID.(uuid.UUID)` |
| 332 | `CancelSwap` | `adminID.(uuid.UUID)` |

**Fix:** Replace each with `uuid.Parse(adminID.(string))` + error check, same pattern as the already-fixed handlers.

### `middleware/auth.go` — `AdminAuth()` (line 80)
```go
// Current — panics if is_admin is not a bool
if !exists || !isAdmin.(bool) {
```
**Fix:** Use comma-ok pattern: `isAdminBool, ok := isAdmin.(bool); if !exists || !ok || !isAdminBool`.

---

## P1 — Unchecked Database Errors (Silent Failures)

GORM calls whose `.Error` is never checked. These silently swallow DB failures and proceed with zero-value results.

### `skill_service.go`
| Line | Operation | Impact |
|------|-----------|--------|
| 100–101 | `Count` on `UserSkillOffered` / `UserSkillWanted` | Deletion may succeed even if skill is in use |
| 120 | `Count` on `UserSkillOffered` | Duplicate skill check skipped on DB error |
| 152 | `Count` on `UserSkillWanted` | Duplicate skill check skipped on DB error |

### `swap_service.go`
| Lines | Operation | Impact |
|-------|-----------|--------|
| 73–76 | `Count` — requester has offered skill | Swap created without valid skill ownership |
| 82–85 | `Count` — responder wants offered skill | Swap created without match validation |
| 91–94 | `Count` — responder offers wanted skill | Same |
| 100–104 | `Count` — duplicate pending swap check | Duplicate swaps can be created |

### `database.go`
| Line | Operation | Impact |
|------|-----------|--------|
| 447 | `db.Find(&existingSkills)` | Seed proceeds with empty slice if query fails |

**Fix:** Check `.Error` on every GORM `Count`/`Find`/`First`/`Save` and return early with a wrapped error.

---

## P1 — Brittle Error String Comparisons (~43 instances)

Handlers compare `err.Error() == "some string"` to determine HTTP status codes. This is fragile — a typo or rewording in the service layer silently changes behavior to a 500.

### Distribution
| File | Count | Examples |
|------|-------|----------|
| `skill/handler.go` | 12 | `"skill not found"`, `"skill already in offered skills"` |
| `swap/handler.go` | 6 | `"swap request not found"`, `"only responder can accept..."` |
| `file/handler.go` | 6 | `"file too large"`, `"invalid file type"` |
| `availability/handler.go` | 3 | `"availability slot not found"` |
| `rating/handler.go` | 4 | `"rating not found"`, `"unauthorized"` |
| `admin/handler.go` | 5 | `"user not found"`, `"cannot ban an admin user"` |
| `auth/handler.go` | 3 | `"user with this email already exists"` |
| `notification/handler.go` | 2 | `"notification not found"` |

**Fix:** Define sentinel errors in a shared `apperrors` package:
```go
package apperrors

import "errors"

var (
    ErrNotFound     = errors.New("not found")
    ErrUnauthorized = errors.New("unauthorized")
    ErrConflict     = errors.New("already exists")
    ErrForbidden    = errors.New("forbidden")
    ErrValidation   = errors.New("validation error")
    ErrFileTooLarge = errors.New("file too large")
    ErrInvalidType  = errors.New("invalid file type")
)
```
Services wrap these: `fmt.Errorf("skill %w", apperrors.ErrNotFound)`.
Handlers check with: `errors.Is(err, apperrors.ErrNotFound)`.

---

## P2 — Missing Input Validation

### `swap/handler.go` — `GetUserSwapRequests`
- `limit` and `offset` query params are parsed with `strconv.Atoi` but missing bounds checks (negative values, absurdly large limits).

### `availability/handler.go` — `GetAvailabilityByDayAndTime`
- `day` parameter validated 1–7 ✓ — but `start_time` / `end_time` aren't checked for `start < end`.

### `rating/handler.go` — `CreateRating`
- Score bound validation happens in the service layer but not in the handler. If the DTO binding doesn't enforce `min=1,max=5`, invalid scores reach the DB.

### General
- No `max length` checks on text fields (`name`, `location`, `comment`, `label`). Long strings go straight to DB.

**Fix:** Add binding tags (`binding:"min=1,max=5"`, `binding:"max=255"`) to DTOs and validate query params for range/bounds in handlers.

---

## P2 — Missing Transaction Handling

### `swap_service.go` — `CreateSwapRequest`
Multiple validation queries + insert + preload without a transaction. If the insert succeeds but preload fails, the client gets an error but the swap exists in DB.

### `swap_service.go` — `UpdateSwapRequestStatus`
Status update + potential related record changes aren't atomic.

### `rating_service.go` — `CreateRating`
Reads swap to verify permissions, then creates rating. Swap could be deleted between the two operations.

**Fix:** Wrap multi-step DB operations in `db.Transaction(func(tx *gorm.DB) error { ... })`.

---

## P3 — Inconsistent Error Logging & Response Format

### Missing server-side logging
When handlers return 500, the underlying error is sent to the client but **not logged**. There's no `log.Error()` call before the response.

Example pattern that loses context:
```go
c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
```
Internal errors should be **logged with context** and the client should get a generic message.

### Inconsistent response shape
Most handlers use `gin.H{"error": "..."}` but some service functions return raw format strings. No standard envelope.

**Fix:**
1. Add `log.Printf("handler=%s err=%v", handlerName, err)` before every 500 response.
2. Never send `err.Error()` directly to client for 500s — use `"Internal server error"`.
3. Optionally adopt a response helper: `respondError(c, status, userMsg, err)` that logs + responds.

---

## Implementation Order

| Phase | Category | Effort | Files |
|-------|----------|--------|-------|
| **1** | P0 — Fix remaining type assertions | Small | `admin/handler.go`, `middleware/auth.go` |
| **2** | P1 — Check all GORM `.Error` returns | Medium | `skill_service.go`, `swap_service.go`, `database.go` |
| **3** | P1 — Sentinel errors + `errors.Is` | Large | New `apperrors/` pkg, all handlers, all services |
| **4** | P2 — Input validation | Medium | DTOs, handlers |
| **5** | P2 — Transaction wrapping | Medium | `swap_service.go`, `rating_service.go` |
| **6** | P3 — Logging + response standardization | Medium | All handlers |
