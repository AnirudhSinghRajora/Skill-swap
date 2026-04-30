package middleware

import (
	"log"
	"os"
	"sync"
	"time"
)

// Limiter is the abstract interface used by the rate-limit middleware.
// Allow returns the remaining count and the time at which the window
// resets. ok=false means the caller has exceeded the budget for the
// current window.
type Limiter interface {
	Allow(key string, max int, window time.Duration) (ok bool, resetAt time.Time)
}

// MemoryLimiter is a per-process limiter backed by an in-memory map.
// Counters reset when the window expires; entries are GC'd lazily on
// access plus a background sweep every minute.
type MemoryLimiter struct {
	mu     sync.Mutex
	bucket map[string]*memoryEntry
}

type memoryEntry struct {
	count   int
	resetAt time.Time
}

func NewMemoryLimiter() *MemoryLimiter {
	m := &MemoryLimiter{bucket: make(map[string]*memoryEntry)}
	go m.sweep()
	return m
}

func (m *MemoryLimiter) sweep() {
	for {
		time.Sleep(time.Minute)
		m.mu.Lock()
		now := time.Now()
		for k, e := range m.bucket {
			if now.After(e.resetAt) {
				delete(m.bucket, k)
			}
		}
		m.mu.Unlock()
	}
}

func (m *MemoryLimiter) Allow(key string, max int, window time.Duration) (bool, time.Time) {
	m.mu.Lock()
	defer m.mu.Unlock()
	now := time.Now()
	e, ok := m.bucket[key]
	if !ok || now.After(e.resetAt) {
		m.bucket[key] = &memoryEntry{count: 1, resetAt: now.Add(window)}
		return true, now.Add(window)
	}
	if e.count >= max {
		return false, e.resetAt
	}
	e.count++
	return true, e.resetAt
}

// defaultLimiter is selected at process start. RATE_LIMIT_BACKEND=redis
// would route here once a Redis adapter is wired; for now we log and fall
// back to memory so behaviour is unchanged.
var defaultLimiter Limiter

func DefaultLimiter() Limiter {
	if defaultLimiter != nil {
		return defaultLimiter
	}
	switch os.Getenv("RATE_LIMIT_BACKEND") {
	case "redis":
		log.Println("⚠ RATE_LIMIT_BACKEND=redis requested but redis adapter not yet wired; falling back to in-memory limiter")
		defaultLimiter = NewMemoryLimiter()
	default:
		defaultLimiter = NewMemoryLimiter()
	}
	return defaultLimiter
}

// SetDefaultLimiter overrides the process-wide limiter. Tests use this to
// inject deterministic implementations.
func SetDefaultLimiter(l Limiter) { defaultLimiter = l }
