package like

import (
	"sync"
	"time"

	"golang.org/x/time/rate"
)

// Limiter is a per-key token bucket. Idle keys are swept lazily so the map
// cannot grow without bound.
type Limiter struct {
	limit rate.Limit
	burst int
	idle  time.Duration
	now   func() time.Time

	mu        sync.Mutex
	buckets   map[string]*bucket
	lastSweep time.Time
}

type bucket struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

// NewLimiter allows burst requests at once per key, refilling at limit per second.
func NewLimiter(limit rate.Limit, burst int) *Limiter {
	return &Limiter{
		limit:   limit,
		burst:   burst,
		idle:    10 * time.Minute,
		now:     time.Now,
		buckets: make(map[string]*bucket),
	}
}

// Allow reports whether key may perform one more request now.
func (l *Limiter) Allow(key string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := l.now()
	l.sweep(now)

	b, ok := l.buckets[key]
	if !ok {
		b = &bucket{limiter: rate.NewLimiter(l.limit, l.burst)}
		l.buckets[key] = b
	}
	b.lastSeen = now
	return b.limiter.AllowN(now, 1)
}

func (l *Limiter) sweep(now time.Time) {
	if now.Sub(l.lastSweep) < l.idle {
		return
	}
	l.lastSweep = now
	for key, b := range l.buckets {
		if now.Sub(b.lastSeen) > l.idle {
			delete(l.buckets, key)
		}
	}
}
