package prefix_tree

import (
	"context"
	"fmt"
	"runtime"
	"testing"
)

// BenchmarkInsertExact benchmarks inserting exact IPv4/IPv6 addresses into the tree
func BenchmarkInsertExact(b *testing.B) {
	ctx := context.Background()
	tree := NewTree()
	keys := generateTestKeys(b.N)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := keys[i].key
		mask := keys[i].mask
		tree.Insert(ctx, key, mask, i)
	}
}

// BenchmarkSearchExact benchmarks exact searches in a pre-populated tree
func BenchmarkSearchExact(b *testing.B) {
	ctx := context.Background()
	tree := NewTree()
	keys := generateTestKeys(b.N)

	// Pre-populate the tree
	for i := 0; i < b.N; i++ {
		key := keys[i].key
		mask := keys[i].mask
		tree.Insert(ctx, key, mask, i)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := keys[i].key
		mask := keys[i].mask
		tree.SearchExact(ctx, key, mask)
	}
}

// BenchmarkSearchPartial benchmarks partial searches in a pre-populated tree
func BenchmarkSearchPartial(b *testing.B) {
	ctx := context.Background()
	tree := NewTree()
	keys := generateTestKeys(b.N)

	// Pre-populate the tree
	for i := 0; i < b.N; i++ {
		key := keys[i].key
		mask := keys[i].mask
		tree.Insert(ctx, key, mask, i)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := keys[i].key
		mask := keys[i].mask
		tree.SearchPartial(ctx, key, mask)
	}
}

// BenchmarkDelete benchmarks deletion from the tree
func BenchmarkDelete(b *testing.B) {
	ctx := context.Background()
	keys := generateTestKeys(1000)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tree := NewTree()
		// Pre-populate with keys
		for j := 0; j < len(keys); j++ {
			key := keys[j].key
			mask := keys[j].mask
			tree.Insert(ctx, key, mask, j)
		}

		b.StopTimer()
		// Benchmark deletion of one key per iteration
		if i < len(keys) {
			key := keys[i].key
			mask := keys[i].mask
			b.StartTimer()
			tree.Delete(ctx, key, mask)
		}
	}
}

// BenchmarkGCPressure measures heap allocations and GC counts during insertions
func BenchmarkGCPressure(b *testing.B) {
	ctx := context.Background()
	tests := []struct {
		name   string
		count  int
		keyLen int
	}{
		{"Small_IPv4", 1000, 4},
		{"Large_IPv4", 10000, 4},
		{"Small_IPv6", 1000, 16},
		{"Large_IPv6", 10000, 16},
	}

	for _, tt := range tests {
		b.Run(tt.name, func(b *testing.B) {
			keys := generateTestKeys(tt.count)

			var m runtime.MemStats

			// Baseline: measure before insertions
			runtime.GC()
			runtime.ReadMemStats(&m)
			baseHeap := m.HeapAlloc
			baseGC := m.NumGC

			b.ResetTimer()

			tree := NewTree()
			for i := 0; i < tt.count; i++ {
				key := keys[i].key
				mask := keys[i].mask
				tree.Insert(ctx, key, mask, i)
			}

			b.StopTimer()

			// Measure after insertions
			runtime.ReadMemStats(&m)
			heapUsed := m.HeapAlloc - baseHeap
			gcCount := m.NumGC - baseGC

			// Report metrics
			b.ReportMetric(float64(heapUsed), "bytes_alloc")
			b.ReportMetric(float64(gcCount), "gc_runs")
			b.ReportMetric(float64(heapUsed)/float64(tt.count), "bytes_per_insert")
			b.Logf("%s: Heap used: %d bytes, GC runs: %d, Avg per insert: %.2f bytes",
				tt.name, heapUsed, gcCount, float64(heapUsed)/float64(tt.count))
		})
	}
}

// BenchmarkMixedOperations benchmarks a realistic mix of insert/search/delete
func BenchmarkMixedOperations(b *testing.B) {
	ctx := context.Background()
	tree := NewTree()
	keys := generateTestKeys(b.N)

	// Pre-populate 50% of keys
	for i := 0; i < b.N/2; i++ {
		key := keys[i].key
		mask := keys[i].mask
		tree.Insert(ctx, key, mask, i)
	}

	b.ResetTimer()

	// Mix of operations: 40% insert, 50% search, 10% delete
	for i := 0; i < b.N; i++ {
		op := i % 100
		idx := i % len(keys)
		key := keys[idx].key
		mask := keys[idx].mask

		if op < 40 {
			tree.Insert(ctx, key, mask, i)
		} else if op < 90 {
			tree.SearchExact(ctx, key, mask)
		} else {
			tree.Delete(ctx, key, mask)
		}
	}
}

// BenchmarkV4TreeInsert benchmarks V4Tree.Insert
func BenchmarkV4TreeInsert(b *testing.B) {
	ctx := context.Background()
	v4tree := NewV4Tree()
	addresses := generateIPv4Addresses(b.N)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		v4tree.Insert(ctx, addresses[i], i)
	}
}

// BenchmarkV4TreeSearch benchmarks V4Tree.Search
func BenchmarkV4TreeSearch(b *testing.B) {
	ctx := context.Background()
	v4tree := NewV4Tree()
	addresses := generateIPv4Addresses(b.N)

	// Pre-populate
	for i := 0; i < b.N; i++ {
		v4tree.Insert(ctx, addresses[i], i)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		v4tree.Search(ctx, addresses[i])
	}
}

// BenchmarkV6TreeInsert benchmarks V6Tree.Insert
func BenchmarkV6TreeInsert(b *testing.B) {
	ctx := context.Background()
	v6tree := NewV6Tree()
	addresses := generateIPv6Addresses(b.N)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		v6tree.Insert(ctx, addresses[i], i)
	}
}

// BenchmarkV6TreeSearch benchmarks V6Tree.Search
func BenchmarkV6TreeSearch(b *testing.B) {
	ctx := context.Background()
	v6tree := NewV6Tree()
	addresses := generateIPv6Addresses(b.N)

	// Pre-populate
	for i := 0; i < b.N; i++ {
		v6tree.Insert(ctx, addresses[i], i)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		v6tree.Search(ctx, addresses[i])
	}
}

// BenchmarkAllocations measures allocation counts during operations
func BenchmarkAllocations(b *testing.B) {
	ctx := context.Background()
	tests := []struct {
		name string
		fn   func()
	}{
		{
			"Insert_1000_Keys",
			func() {
				tree := NewTree()
				keys := generateTestKeys(1000)
				for i := 0; i < 1000; i++ {
					key := keys[i].key
					mask := keys[i].mask
					tree.Insert(ctx, key, mask, i)
				}
			},
		},
		{
			"Search_1000_Keys",
			func() {
				tree := NewTree()
				keys := generateTestKeys(1000)
				for i := 0; i < 1000; i++ {
					key := keys[i].key
					mask := keys[i].mask
					tree.Insert(ctx, key, mask, i)
				}
				for i := 0; i < 1000; i++ {
					key := keys[i].key
					mask := keys[i].mask
					tree.SearchExact(ctx, key, mask)
				}
			},
		},
	}

	for _, tt := range tests {
		b.Run(tt.name, func(b *testing.B) {
			var m runtime.MemStats

			runtime.GC()
			runtime.ReadMemStats(&m)
			baseAllocs := m.Mallocs

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				tt.fn()
			}
			b.StopTimer()

			runtime.ReadMemStats(&m)
			allocCount := m.Mallocs - baseAllocs
			b.ReportMetric(float64(allocCount), "mallocs")
		})
	}
}

// Helper to generate test keys
type testKey struct {
	key  []byte
	mask []byte
}

func generateTestKeys(count int) []testKey {
	keys := make([]testKey, count)
	for i := 0; i < count; i++ {
		// Alternate between IPv4 (4 bytes) and IPv6 (16 bytes) for diversity
		if i%2 == 0 {
			keys[i] = testKey{
				key:    []byte{byte(i % 256), byte((i / 256) % 256), byte(i % 128), byte(i % 64)},
				mask:   []byte{0xFF, 0xFF, 0xFF, 0xFF},
			}
		} else {
			key := make([]byte, 16)
			mask := make([]byte, 16)
			for j := 0; j < 16; j++ {
				key[j] = byte((i + j) % 256)
				mask[j] = 0xFF
			}
			keys[i] = testKey{
				key:  key,
				mask: mask,
			}
		}
	}
	return keys
}

// Helper to generate test IPv4 addresses
func generateIPv4Addresses(count int) []string {
	addresses := make([]string, count)
	for i := 0; i < count; i++ {
		a := (i >> 24) & 0xFF
		b := (i >> 16) & 0xFF
		c := (i >> 8) & 0xFF
		d := i & 0xFF
		addresses[i] = fmt.Sprintf("%d.%d.%d.%d/32", a, b, c, d)
	}
	return addresses
}

// Helper to generate test IPv6 addresses
func generateIPv6Addresses(count int) []string {
	addresses := make([]string, count)
	for i := 0; i < count; i++ {
		// Simple IPv6 generation: 2001:db8:xxxx:xxxx::1/128
		h1 := (i >> 16) & 0xFFFF
		h2 := i & 0xFFFF
		addresses[i] = fmt.Sprintf("2001:db8:%04x:%04x::1/128", h1, h2)
	}
	return addresses
}
