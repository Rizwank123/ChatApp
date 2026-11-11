package cache

import (
	"context"
	"fmt"
	"hash/fnv"
	"math"
	"sync"
	"sync/atomic"

	"github.com/Rizwank123/chatApp/internal/pkg/util"
)

// BloomStats tracks performance metrics and queries avoided by the Bloom filter.
type BloomStats struct {
	Capacity          uint64  `json:"capacity"`
	ItemsCount        uint64  `json:"items_count"`
	TargetFPRate      float64 `json:"target_fp_rate"`
	CurrentFPRate     float64 `json:"current_fp_rate"`
	MemoryBytes       uint64  `json:"memory_bytes"`
	DBQueriesAvoided  uint64  `json:"db_queries_avoided"`
	DBQueriesVerified uint64  `json:"db_queries_verified"`
	FalsePositives    uint64  `json:"false_positives"`
}

// UsernameBloomFilter defines the interface for username existence pre-filtering.
type UsernameBloomFilter interface {
	Add(ctx context.Context, username string) error
	MightContain(ctx context.Context, username string) (bool, error)
	Capacity() uint64
	FalsePositiveRate() float64
	EstimatedCount() uint64
	Clear(ctx context.Context) error
}

// MemoryBloomFilter is a concurrent, zero-dependency Bloom filter using double hashing.
type MemoryBloomFilter struct {
	mu           sync.RWMutex
	bits         []uint64
	m            uint64  // total number of bits
	k            uint    // number of hash functions
	n            uint64  // target capacity
	p            float64 // target false positive rate
	count        uint64  // approximate number of items added
}

// NewMemoryBloomFilter creates an in-memory Bloom filter optimized for capacity n and false positive rate p.
// Typical: n = 100,000, p = 0.01 (1%).
func NewMemoryBloomFilter(capacity uint64, falsePositiveRate float64) *MemoryBloomFilter {
	if capacity == 0 {
		capacity = 100000
	}
	if falsePositiveRate <= 0 || falsePositiveRate >= 1 {
		falsePositiveRate = 0.01
	}

	// Calculate optimal m (bits) and k (hash functions)
	// m = - (n * ln(p)) / (ln(2)^2)
	// k = (m / n) * ln(2)
	ln2 := math.Ln2
	mFloat := -float64(capacity) * math.Log(falsePositiveRate) / (ln2 * ln2)
	m := uint64(math.Ceil(mFloat))
	if m < 64 {
		m = 64
	}

	kFloat := (float64(m) / float64(capacity)) * ln2
	k := uint(math.Round(kFloat))
	if k < 1 {
		k = 1
	}

	words := (m + 63) / 64
	return &MemoryBloomFilter{
		bits:  make([]uint64, words),
		m:     m,
		k:     k,
		n:     capacity,
		p:     falsePositiveRate,
		count: 0,
	}
}

// getHashes produces k hash locations using the Kirsch-Mitzenmacher double-hashing technique:
// gi(x) = (h1(x) + i * h2(x)) % m
func (b *MemoryBloomFilter) getHashes(data []byte) []uint64 {
	h1 := fnv64a(data)
	h2 := fnv64(data)
	if h2 == 0 {
		h2 = 1
	}

	indices := make([]uint64, b.k)
	for i := uint(0); i < b.k; i++ {
		idx := (h1 + uint64(i)*h2) % b.m
		indices[i] = idx
	}
	return indices
}

func fnv64a(data []byte) uint64 {
	h := fnv.New64a()
	h.Write(data)
	return h.Sum64()
}

func fnv64(data []byte) uint64 {
	h := fnv.New64()
	h.Write(data)
	return h.Sum64()
}

// Add inserts a normalized username into the filter.
func (b *MemoryBloomFilter) Add(ctx context.Context, username string) error {
	norm := util.NormalizeUsername(username)
	if norm == "" {
		return nil
	}

	hashes := b.getHashes([]byte(norm))

	b.mu.Lock()
	defer b.mu.Unlock()

	for _, bitIdx := range hashes {
		wordIdx := bitIdx / 64
		bitOffset := bitIdx % 64
		b.bits[wordIdx] |= (1 << bitOffset)
	}
	b.count++
	return nil
}

// MightContain checks if the username might be in the filter.
func (b *MemoryBloomFilter) MightContain(ctx context.Context, username string) (bool, error) {
	norm := util.NormalizeUsername(username)
	if norm == "" {
		return false, nil
	}

	hashes := b.getHashes([]byte(norm))

	b.mu.RLock()
	defer b.mu.RUnlock()

	for _, bitIdx := range hashes {
		wordIdx := bitIdx / 64
		bitOffset := bitIdx % 64
		if (b.bits[wordIdx] & (1 << bitOffset)) == 0 {
			return false, nil // Definitely does NOT exist
		}
	}
	return true, nil // MIGHT exist
}

func (b *MemoryBloomFilter) Capacity() uint64 {
	return b.n
}

func (b *MemoryBloomFilter) FalsePositiveRate() float64 {
	return b.p
}

func (b *MemoryBloomFilter) EstimatedCount() uint64 {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.count
}

func (b *MemoryBloomFilter) Clear(ctx context.Context) error {
	b.mu.Lock()
	defer b.mu.Unlock()
	for i := range b.bits {
		b.bits[i] = 0
	}
	b.count = 0
	return nil
}

// MemoryBytes returns approximate heap usage in bytes.
func (b *MemoryBloomFilter) MemoryBytes() uint64 {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return uint64(len(b.bits)) * 8
}

// CurrentFalsePositiveRate computes (1 - e^(-k*n/m))^k based on current item count.
func (b *MemoryBloomFilter) CurrentFalsePositiveRate() float64 {
	b.mu.RLock()
	defer b.mu.RUnlock()
	if b.count == 0 {
		return 0
	}
	exponent := -float64(b.k) * float64(b.count) / float64(b.m)
	return math.Pow(1.0-math.Exp(exponent), float64(b.k))
}

// UserExistenceChecker orchestrates Bloom-filter assisted checks with DB authoritative verification.
type UserExistenceChecker struct {
	filter            UsernameBloomFilter
	dbLookupFunc      func(ctx context.Context, username string) (bool, error)
	dbAvoidedCount    uint64
	dbVerifiedCount   uint64
	falsePositives    uint64
}

// NewUserExistenceChecker builds a checker wrapping a filter and a DB lookup function.
func NewUserExistenceChecker(
	filter UsernameBloomFilter,
	dbLookupFunc func(ctx context.Context, username string) (bool, error),
) *UserExistenceChecker {
	if filter == nil {
		filter = NewMemoryBloomFilter(100000, 0.01)
	}
	return &UserExistenceChecker{
		filter:       filter,
		dbLookupFunc: dbLookupFunc,
	}
}

// Exists executes the 5-step lookup:
// 1. Normalize username.
// 2. Check Bloom filter.
// 3. If false, avoid DB query and return false.
// 4. If true, verify existence in PostgreSQL.
// 5. Return authoritative DB result.
func (c *UserExistenceChecker) Exists(ctx context.Context, rawUsername string) (bool, error) {
	norm := util.NormalizeUsername(rawUsername)
	if norm == "" {
		return false, nil
	}

	// 2. Check Bloom filter
	mightExist, err := c.filter.MightContain(ctx, norm)
	if err != nil {
		// Filter infrastructure failure fallback: directly check PostgreSQL
		atomic.AddUint64(&c.dbVerifiedCount, 1)
		return c.dbLookupFunc(ctx, norm)
	}

	// 3. Negative lookup: definitely does not exist
	if !mightExist {
		atomic.AddUint64(&c.dbAvoidedCount, 1)
		return false, nil
	}

	// 4. Might exist: Authoritative database verification
	atomic.AddUint64(&c.dbVerifiedCount, 1)
	existsInDB, err := c.dbLookupFunc(ctx, norm)
	if err != nil {
		return false, err
	}

	// 5. Track false positive if filter claimed true but DB has no record
	if !existsInDB {
		atomic.AddUint64(&c.falsePositives, 1)
	}

	return existsInDB, nil
}

// RecordUser adds the newly committed username to the Bloom filter.
func (c *UserExistenceChecker) RecordUser(ctx context.Context, rawUsername string) error {
	norm := util.NormalizeUsername(rawUsername)
	if norm == "" {
		return nil
	}
	return c.filter.Add(ctx, norm)
}

// Backfill populates the Bloom filter from existing usernames in PostgreSQL on startup.
func (c *UserExistenceChecker) Backfill(ctx context.Context, usernames []string) error {
	for _, u := range usernames {
		if err := c.RecordUser(ctx, u); err != nil {
			return fmt.Errorf("failed to backfill user %s: %w", u, err)
		}
	}
	return nil
}

// PopulateUsernames is a convenience helper for backfilling usernames.
func (c *UserExistenceChecker) PopulateUsernames(usernames []string) error {
	return c.Backfill(context.Background(), usernames)
}

// Stats returns the performance and query avoidance metrics.
func (c *UserExistenceChecker) Stats() BloomStats {
	memBytes := uint64(0)
	currentFPRate := float64(0)
	if memFilter, ok := c.filter.(*MemoryBloomFilter); ok {
		memBytes = memFilter.MemoryBytes()
		currentFPRate = memFilter.CurrentFalsePositiveRate()
	}

	return BloomStats{
		Capacity:          c.filter.Capacity(),
		ItemsCount:        c.filter.EstimatedCount(),
		TargetFPRate:      c.filter.FalsePositiveRate(),
		CurrentFPRate:     currentFPRate,
		MemoryBytes:       memBytes,
		DBQueriesAvoided:  atomic.LoadUint64(&c.dbAvoidedCount),
		DBQueriesVerified: atomic.LoadUint64(&c.dbVerifiedCount),
		FalsePositives:    atomic.LoadUint64(&c.falsePositives),
	}
}

// Filter returns the underlying filter.
func (c *UserExistenceChecker) Filter() UsernameBloomFilter {
	return c.filter
}
