package cache_test

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"

	"github.com/Rizwank123/chatApp/internal/cache"
)

func TestMemoryBloomFilter_Basic(t *testing.T) {
	ctx := context.Background()
	filter := cache.NewMemoryBloomFilter(10000, 0.01)

	// Add users
	usernames := []string{"rizwan", "alice", "bob_123", "charlie.brown"}
	for _, u := range usernames {
		if err := filter.Add(ctx, u); err != nil {
			t.Fatalf("failed to add %s: %v", u, err)
		}
	}

	// Verify all existing users return true (zero false negatives)
	for _, u := range usernames {
		contains, err := filter.MightContain(ctx, u)
		if err != nil {
			t.Fatalf("unexpected error checking %s: %v", u, err)
		}
		if !contains {
			t.Fatalf("expected filter to contain '%s', got false", u)
		}
	}

	// Case insensitivity normalization check
	contains, err := filter.MightContain(ctx, "  RIZWAN ")
	if err != nil || !contains {
		t.Fatalf("expected case-insensitive match for RIZWAN")
	}

	// Non-existing user
	contains, err = filter.MightContain(ctx, "non_existing_user_xyz")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if contains {
		t.Fatalf("expected non_existing_user_xyz to be false in fresh small filter")
	}
}

func TestUserExistenceChecker_Flow(t *testing.T) {
	ctx := context.Background()

	// Mock PostgreSQL database
	dbStore := map[string]bool{
		"rizwan": true,
		"john":   true,
	}

	dbLookup := func(ctx context.Context, username string) (bool, error) {
		return dbStore[username], nil
	}

	filter := cache.NewMemoryBloomFilter(1000, 0.01)
	checker := cache.NewUserExistenceChecker(filter, dbLookup)

	// 1. Check non-existing user without backfill: avoided DB query
	exists, err := checker.Exists(ctx, "nobody")
	if err != nil || exists {
		t.Fatalf("expected nobody not to exist")
	}
	stats := checker.Stats()
	if stats.DBQueriesAvoided != 1 {
		t.Fatalf("expected 1 avoided query, got %d", stats.DBQueriesAvoided)
	}

	// 2. Backfill existing users
	err = checker.Backfill(ctx, []string{"rizwan", "john"})
	if err != nil {
		t.Fatalf("failed to backfill: %v", err)
	}

	// 3. Check existing user: filter says true -> DB verifies true
	exists, err = checker.Exists(ctx, "Rizwan")
	if err != nil || !exists {
		t.Fatalf("expected Rizwan to exist")
	}
	stats = checker.Stats()
	if stats.DBQueriesVerified != 1 {
		t.Fatalf("expected 1 verified query, got %d", stats.DBQueriesVerified)
	}

	// 4. Test False Positive simulation: item is in filter but deleted or not in DB
	_ = filter.Add(ctx, "ghost_user")
	exists, err = checker.Exists(ctx, "ghost_user")
	if err != nil || exists {
		t.Fatalf("expected ghost_user not to exist in DB")
	}
	stats = checker.Stats()
	if stats.FalsePositives != 1 {
		t.Fatalf("expected 1 tracked false positive, got %d", stats.FalsePositives)
	}
}

func TestUserExistenceChecker_Fallbacks(t *testing.T) {
	ctx := context.Background()

	// A failing filter
	failingFilter := &mockErrorFilter{}
	dbStore := map[string]bool{"alice": true}
	dbLookup := func(ctx context.Context, username string) (bool, error) {
		return dbStore[username], nil
	}

	checker := cache.NewUserExistenceChecker(failingFilter, dbLookup)

	// When filter returns an error, checker safely falls back to DB query
	exists, err := checker.Exists(ctx, "alice")
	if err != nil || !exists {
		t.Fatalf("expected DB fallback to return true for alice, got %v, err=%v", exists, err)
	}
	stats := checker.Stats()
	if stats.DBQueriesVerified != 1 {
		t.Fatalf("expected 1 DB query verified on fallback, got %d", stats.DBQueriesVerified)
	}
}

func TestUserExistenceChecker_Concurrency(t *testing.T) {
	ctx := context.Background()
	filter := cache.NewMemoryBloomFilter(50000, 0.01)

	var mu sync.Mutex
	dbUsers := make(map[string]bool)

	checker := cache.NewUserExistenceChecker(filter, func(ctx context.Context, username string) (bool, error) {
		mu.Lock()
		defer mu.Unlock()
		return dbUsers[username], nil
	})

	var wg sync.WaitGroup
	// Concurrently register 500 users
	for i := 0; i < 500; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			uname := fmt.Sprintf("user_%04d", id)

			mu.Lock()
			dbUsers[uname] = true
			mu.Unlock()

			_ = checker.RecordUser(ctx, uname)
		}(i)
	}

	// Concurrently query existing and non-existing users
	for i := 0; i < 500; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			uname := fmt.Sprintf("user_%04d", id)
			nonUname := fmt.Sprintf("fake_%04d", id)

			_, _ = checker.Exists(ctx, uname)
			_, _ = checker.Exists(ctx, nonUname)
		}(i)
	}

	wg.Wait()
}

type mockErrorFilter struct{}

func (m *mockErrorFilter) Add(ctx context.Context, username string) error {
	return errors.New("redis connection refused")
}
func (m *mockErrorFilter) MightContain(ctx context.Context, username string) (bool, error) {
	return false, errors.New("redis connection refused")
}
func (m *mockErrorFilter) Capacity() uint64           { return 0 }
func (m *mockErrorFilter) FalsePositiveRate() float64 { return 0 }
func (m *mockErrorFilter) EstimatedCount() uint64     { return 0 }
func (m *mockErrorFilter) Clear(ctx context.Context) error {
	return errors.New("redis connection refused")
}
