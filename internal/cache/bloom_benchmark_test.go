package cache_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/Rizwank123/chatApp/internal/cache"
)

// Simulated DB lookup with typical 1ms latency
func simulatedDBLookup(ctx context.Context, username string) (bool, error) {
	time.Sleep(100 * time.Microsecond) // simulate fast local DB round-trip
	return username == "existing_user", nil
}

func BenchmarkDatabaseOnly_ExistenceCheck(b *testing.B) {
	ctx := context.Background()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = simulatedDBLookup(ctx, "non_existing_random_user")
	}
}

func BenchmarkBloomFilter_NegativeCheck(b *testing.B) {
	ctx := context.Background()
	filter := cache.NewMemoryBloomFilter(100000, 0.01)
	_ = filter.Add(ctx, "existing_user")
	checker := cache.NewUserExistenceChecker(filter, simulatedDBLookup)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Filter returns false immediately, completely skipping DB
		_, _ = checker.Exists(ctx, "non_existing_random_user")
	}
}

func BenchmarkBloomFilter_PositiveCheck(b *testing.B) {
	ctx := context.Background()
	filter := cache.NewMemoryBloomFilter(100000, 0.01)
	_ = filter.Add(ctx, "existing_user")
	checker := cache.NewUserExistenceChecker(filter, simulatedDBLookup)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Filter returns true -> DB verifies
		_, _ = checker.Exists(ctx, "existing_user")
	}
}

func BenchmarkBloomFilter_ConcurrentMixed(b *testing.B) {
	ctx := context.Background()
	filter := cache.NewMemoryBloomFilter(100000, 0.01)
	for i := 0; i < 1000; i++ {
		_ = filter.Add(ctx, fmt.Sprintf("user_%d", i))
	}
	checker := cache.NewUserExistenceChecker(filter, func(ctx context.Context, username string) (bool, error) {
		return true, nil
	})

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			i++
			if i%5 == 0 {
				_, _ = checker.Exists(ctx, fmt.Sprintf("user_%d", i%1000))
			} else {
				_, _ = checker.Exists(ctx, fmt.Sprintf("unknown_%d", i))
			}
		}
	})
}
