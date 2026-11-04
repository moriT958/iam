package like

import (
	"context"
	"path/filepath"
	"sync"
	"testing"
)

func openTestStore(t *testing.T) *SQLiteStore {
	t.Helper()
	s, err := OpenSQLite(filepath.Join(t.TempDir(), "data", "iam.db"))
	if err != nil {
		t.Fatalf("OpenSQLite: %v", err)
	}
	t.Cleanup(func() { _ = s.Close() })
	return s
}

func TestStoreCounting(t *testing.T) {
	ctx := context.Background()
	s := openTestStore(t)

	if n, err := s.Count(ctx, "a"); err != nil || n != 0 {
		t.Fatalf("Count(unliked) = %d, %v; want 0, nil", n, err)
	}
	for want := int64(1); want <= 3; want++ {
		if n, err := s.Increment(ctx, "a"); err != nil || n != want {
			t.Fatalf("Increment = %d, %v; want %d", n, err, want)
		}
	}
	if n, err := s.Decrement(ctx, "a"); err != nil || n != 2 {
		t.Fatalf("Decrement = %d, %v; want 2", n, err)
	}
	if n, err := s.Count(ctx, "a"); err != nil || n != 2 {
		t.Fatalf("Count = %d, %v; want 2", n, err)
	}
}

func TestStoreDecrementNeverGoesBelowZero(t *testing.T) {
	ctx := context.Background()
	s := openTestStore(t)

	// Never liked.
	if n, err := s.Decrement(ctx, "a"); err != nil || n != 0 {
		t.Fatalf("Decrement(unliked) = %d, %v; want 0", n, err)
	}
	// Liked once, decremented twice.
	if _, err := s.Increment(ctx, "a"); err != nil {
		t.Fatal(err)
	}
	for range 2 {
		if n, err := s.Decrement(ctx, "a"); err != nil || n != 0 {
			t.Fatalf("Decrement = %d, %v; want 0", n, err)
		}
	}
}

func TestStoreCounts(t *testing.T) {
	ctx := context.Background()
	s := openTestStore(t)
	_, _ = s.Increment(ctx, "a")
	_, _ = s.Increment(ctx, "a")
	_, _ = s.Increment(ctx, "b")

	got, err := s.Counts(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 || got["a"] != 2 || got["b"] != 1 {
		t.Fatalf("Counts = %v; want map[a:2 b:1]", got)
	}
}

func TestStorePersistsAcrossReopen(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "iam.db")

	s1, err := OpenSQLite(path)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := s1.Increment(ctx, "a"); err != nil {
		t.Fatal(err)
	}
	if err := s1.Close(); err != nil {
		t.Fatal(err)
	}

	s2, err := OpenSQLite(path)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = s2.Close() }()
	if n, err := s2.Count(ctx, "a"); err != nil || n != 1 {
		t.Fatalf("Count after reopen = %d, %v; want 1", n, err)
	}
}

func TestStoreConcurrentIncrements(t *testing.T) {
	ctx := context.Background()
	s := openTestStore(t)

	const n = 50
	var wg sync.WaitGroup
	for range n {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if _, err := s.Increment(ctx, "a"); err != nil {
				t.Errorf("Increment: %v", err)
			}
		}()
	}
	wg.Wait()

	if got, err := s.Count(ctx, "a"); err != nil || got != n {
		t.Fatalf("Count = %d, %v; want %d (no lost updates)", got, err, n)
	}
}
