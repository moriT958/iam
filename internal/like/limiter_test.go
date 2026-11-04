package like

import (
	"testing"
	"time"

	"golang.org/x/time/rate"
)

func TestLimiterRefillsOverTime(t *testing.T) {
	now := time.Unix(1_000_000, 0)
	l := NewLimiter(rate.Every(time.Second), 2)
	l.now = func() time.Time { return now }

	if !l.Allow("a") || !l.Allow("a") {
		t.Fatal("burst of 2 should be allowed")
	}
	if l.Allow("a") {
		t.Fatal("third request should be limited")
	}
	now = now.Add(1100 * time.Millisecond)
	if !l.Allow("a") {
		t.Fatal("one token should have been refilled after 1s")
	}
	if l.Allow("a") {
		t.Fatal("only one token should have been refilled")
	}
}

func TestLimiterSweepsIdleKeys(t *testing.T) {
	now := time.Unix(1_000_000, 0)
	l := NewLimiter(rate.Every(time.Second), 1)
	l.now = func() time.Time { return now }

	for _, key := range []string{"a", "b", "c"} {
		l.Allow(key)
	}
	if len(l.buckets) != 3 {
		t.Fatalf("buckets = %d, want 3", len(l.buckets))
	}
	now = now.Add(l.idle + time.Minute)
	l.Allow("d")
	if len(l.buckets) != 1 {
		t.Fatalf("buckets after sweep = %d, want 1 (only d)", len(l.buckets))
	}
}
