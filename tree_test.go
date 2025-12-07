package prefix_tree

import (
	"context"
	"fmt"
	"runtime"
	"testing"
)

func TestTree_Insert_Search_Delete(t *testing.T) {
	ctx := context.Background()
	tr := NewTree[string]()

	// basic insert
	key := []byte{0xC0, 0xA8, 0x01, 0x01} // 192.168.1.1
	mask := []byte{0xFF, 0xFF, 0xFF, 0xFF}

	val := "value1"
	res, err := tr.Insert(ctx, key, mask, &val)
	if err != nil || res != Ok {
		t.Fatalf("Insert failed: res=%v err=%v", res, err)
	}

	if tr.numNodes != 1 {
		t.Fatalf("expected NumNodes 1 got %d", tr.numNodes)
	}

	// duplicate insert returns Dup
	val1dup := "value1-dup"
	res, err = tr.Insert(ctx, key, mask, &val1dup)
	if err != nil || res != Dup {
		t.Fatalf("expected Dup on duplicate insert, got res=%v err=%v", res, err)
	}

	// exact search must find value
	res, v, err := tr.SearchExact(ctx, key, mask)
	if err != nil || res != Match {
		t.Fatalf("SearchExact failed: res=%v err=%v", res, err)
	}
	if *v != val {
		t.Fatalf("unexpected value: %v", v)
	}

	// partial search for same key should also find the value
	res, v, err = tr.SearchPartial(ctx, key, mask)
	if err != nil || res != Match {
		t.Fatalf("SearchPartial failed: res=%v err=%v", res, err)
	}
	if *v != val {
		t.Fatalf("unexpected value: %v", v)
	}

	// delete should return the stored value and decrement NumNodes
	res, pval, err := tr.Delete(ctx, key, mask)
	if err != nil || res != Match {
		t.Fatalf("Delete failed: res=%v err=%v", res, err)
	}
	if *pval != val {
		t.Fatalf("Delete returned wrong value: %v", val)
	}
	if tr.numNodes != 0 {
		t.Fatalf("expected NumNodes 0 got %d", tr.numNodes)
	}

	// delete non-existent should return error
	res, _, err = tr.Delete(ctx, key, mask)
	if err == nil || res != Error {
		t.Fatalf("expected error deleting non-existent key")
	}
}

func TestTree_Partial_Prefixs(t *testing.T) {
	ctx := context.Background()
	tr := NewTree[string]()

	// insert a /24 prefix: 10.0.1.0/24 -> mask first three octets
	key := []byte{10, 0, 1, 0}
	mask := []byte{0xFF, 0xFF, 0xFF, 0x00} // /24

	vnet := "net-10-0-1"
	res, err := tr.Insert(ctx, key, mask, &vnet)
	if err != nil || res != Ok {
		t.Fatalf("Insert prefix failed: %v %v", res, err)
	}

	// search for an address inside prefix should match partial search
	addr := []byte{10, 0, 1, 42}
	res, pval, err := tr.SearchPartial(ctx, addr, mask)
	if err != nil || res != Match {
		t.Fatalf("SearchPartial for address inside prefix failed: %v %v", res, err)
	}
	if *pval != vnet {
		t.Fatalf("unexpected partial match value: %v", *pval)
	}

	// exact search for the specific network key should also match
	res, pval, err = tr.SearchExact(ctx, key, mask)
	if err != nil || res != Match {
		t.Fatalf("SearchExact for network key failed: %v %v", res, err)
	}
	if *pval != vnet {
		t.Fatalf("unexpected exact match value: %v", *pval)
	}

	// delete should return the stored value and decrement NumNodes
	res, pval, err = tr.Delete(ctx, key, mask)
	if err != nil || res != Match {
		t.Fatalf("Delete prefix failed: %v %v", res, err)
	}
	if *pval != vnet {
		t.Fatalf("Delete returned wrong value for prefix: %v", *pval)
	}
	if tr.numNodes != 0 {
		t.Fatalf("expected NumNodes 0 got %d", tr.numNodes)
	}

	// search after delete should fail
	res, _, err = tr.SearchExact(ctx, key, mask)
	if err == nil || res != Error {
		t.Fatalf("expected error searching deleted prefix")
	}

	// partial search for address inside deleted prefix should also fail
	res, _, err = tr.SearchPartial(ctx, addr, mask)
	if err == nil || res != Error {
		t.Fatalf("expected error searching address inside deleted prefix")
	}

	// Insert again to ensure tree is still functional
	valagain := "net-10-0-1-again"
	res, err = tr.Insert(ctx, key, mask, &valagain)
	if err != nil || res != Ok {
		t.Fatalf("Re-insert prefix failed: %v %v", res, err)
	}

	// Delete with partial key should work
	res, pval, err = tr.Delete(ctx, addr, mask)
	if err != nil {
		t.Fatalf("Delete with partial key failed: %v %v", res, err)
	}
	if *pval != valagain {
		t.Fatalf("Delete with partial key returned wrong value: %v", *pval)
	}
	if tr.numNodes != 0 {
		t.Fatalf("expected NumNodes 0 got %d", tr.numNodes)
	}
}

func TestTree_LockingHandlers(t *testing.T) {
	ctx := context.Background()
	// Very small test to ensure lock handlers are called
	called := struct{ r, ru, w, u int }{}

	rlock := func(_ context.Context) { called.r++ }
	runlock := func(_ context.Context) { called.ru++ }
	wlock := func(_ context.Context) { called.w++ }
	unlock := func(_ context.Context) { called.u++ }

	tr := NewTreeWithLockHandlers[string](rlock, runlock, wlock, unlock)

	key := []byte{1, 2, 3, 4}
	mask := []byte{0xFF, 0xFF, 0xFF, 0xFF}

	val := "x"

	// Insert should call write lock/unlock
	_, _ = tr.Insert(ctx, key, mask, &val)
	if called.w == 0 || called.u == 0 {
		t.Fatalf("expected write lock/unlock called, got w=%d u=%d", called.w, called.u)
	}

	// Search should call read lock/unlock
	_, _, _ = tr.SearchExact(ctx, key, mask)
	if called.r == 0 || called.ru == 0 {
		t.Fatalf("expected read lock/unlock called, got r=%d ru=%d", called.r, called.ru)
	}

	// SearchPartial should call read lock/unlock
	_, _, _ = tr.SearchPartial(ctx, key, mask)
	if called.r < 2 || called.ru < 2 {
		t.Fatalf("expected read lock/unlock called twice, got r=%d ru=%d", called.r, called.ru)
	}

	// Delete should call write lock/unlock
	_, _, _ = tr.Delete(ctx, key, mask)
	if called.w < 2 || called.u < 2 {
		t.Fatalf("expected write lock/unlock called twice, got w=%d u=%d", called.w, called.u)
	}
}
func TestWalk_DepthFirstTraversal(t *testing.T) {
	ctx := context.Background()
	tr := NewTree[string]()

	// Build a small tree with controlled structure
	// Insert keys to create a specific tree shape:
	// 10000000 (bit 0) = left, binary 128
	// 01000000 (bit 1) = right, binary 64
	// etc.

	// Key1: 10000000 (128) with full mask
	key1 := []byte{128}
	mask1 := []byte{0xFF}
	val1 := "node1"
	tr.Insert(ctx, key1, mask1, &val1)

	// Key2: 01000000 (64) with full mask
	key2 := []byte{64}
	mask2 := []byte{0xFF}
	val2 := "node2"
	tr.Insert(ctx, key2, mask2, &val2)

	// Key3: 11000000 (192) with full mask
	key3 := []byte{192}
	mask3 := []byte{0xFF}
	val3 := "node3"
	tr.Insert(ctx, key3, mask3, &val3)

	// Key4: 00100000 (32) with full mask
	key4 := []byte{32}
	mask4 := []byte{0xFF}
	val4 := "node4"
	tr.Insert(ctx, key4, mask4, &val4)

	// Walk the tree and collect all values
	visited := []string{}
	err := tr.Walk(ctx, func(c context.Context, v *string) error {
		visited = append(visited, *v)
		return nil
	})

	if err != nil {
		t.Fatalf("Walk failed: %v", err)
	}

	// Verify all nodes were visited
	if len(visited) != 4 {
		t.Fatalf("expected 4 visited nodes, got %d: %v", len(visited), visited)
	}

	// Check that all values are present (order may vary for DFS)
	expected := map[string]bool{"node1": true, "node2": true, "node3": true, "node4": true}
	for _, v := range visited {
		if !expected[v] {
			t.Fatalf("unexpected node visited: %s", v)
		}
		delete(expected, v)
	}

	if len(expected) > 0 {
		t.Fatalf("some nodes were not visited: %v", expected)
	}
}

func TestWalk_SinglePath(t *testing.T) {
	ctx := context.Background()
	tr := NewTree[int]()

	// Create a simple linear path: root -> node1 -> node2
	// Insert two keys that share a prefix but diverge later
	key1 := []byte{0xC0, 0x00} // 11000000 00000000
	mask1 := []byte{0xFF, 0xFF}
	val1 := 100
	tr.Insert(ctx, key1, mask1, &val1)

	key2 := []byte{0xC0, 0x80} // 11000000 10000000
	mask2 := []byte{0xFF, 0xFF}
	val2 := 200
	tr.Insert(ctx, key2, mask2, &val2)

	visited := []int{}
	err := tr.Walk(ctx, func(c context.Context, v *int) error {
		visited = append(visited, *v)
		return nil
	})

	if err != nil {
		t.Fatalf("Walk failed: %v", err)
	}

	if len(visited) != 2 {
		t.Fatalf("expected 2 nodes, got %d", len(visited))
	}

	// Both values should be present
	found := make(map[int]bool)
	for _, v := range visited {
		found[v] = true
	}
	if !found[100] || !found[200] {
		t.Fatalf("not all values found in walk: %v", visited)
	}
}

func TestWalk_OnlyTerminalNodes(t *testing.T) {
	// Verify that Walk only calls walkerFn for terminal nodes,
	// not intermediate nodes
	ctx := context.Background()
	tr := NewTree[string]()

	// Insert a long key that creates many intermediate nodes
	key := []byte{0xFF, 0xFF, 0xFF}
	mask := []byte{0xFF, 0xFF, 0xFF}
	val := "deep_value"
	tr.Insert(ctx, key, mask, &val)

	// Insert a shorter prefix that shares bits with the long key
	key2 := []byte{0xFF, 0x00, 0x00}
	mask2 := []byte{0xFF, 0x00, 0x00}
	val2 := "short_value"
	tr.Insert(ctx, key2, mask2, &val2)

	visited := []string{}
	err := tr.Walk(ctx, func(c context.Context, v *string) error {
		visited = append(visited, *v)
		return nil
	})

	if err != nil {
		t.Fatalf("Walk failed: %v", err)
	}

	// Should only visit the 2 terminal nodes, not intermediate nodes
	if len(visited) != 2 {
		t.Fatalf("expected 2 terminal nodes visited, got %d", len(visited))
	}

	found := make(map[string]bool)
	for _, v := range visited {
		found[v] = true
	}
	if !found["deep_value"] || !found["short_value"] {
		t.Fatalf("not all terminal values found: %v", visited)
	}
}

func TestWalk_EmptyTree(t *testing.T) {
	ctx := context.Background()
	tr := NewTree[string]()

	// Walk empty tree should not call walkerFn
	callCount := 0
	err := tr.Walk(ctx, func(c context.Context, v *string) error {
		callCount++
		return nil
	})

	if err != nil {
		t.Fatalf("Walk on empty tree failed: %v", err)
	}

	if callCount != 0 {
		t.Fatalf("expected 0 calls to walkerFn on empty tree, got %d", callCount)
	}
}

func TestWalk_StopOnError(t *testing.T) {
	ctx := context.Background()
	tr := NewTree[string]()

	// Insert multiple nodes
	val1 := "node1"
	tr.Insert(ctx, []byte{0x80}, []byte{0xFF}, &val1)
	val2 := "node2"
	tr.Insert(ctx, []byte{0x40}, []byte{0xFF}, &val2)
	val3 := "node3"
	tr.Insert(ctx, []byte{0x20}, []byte{0xFF}, &val3)

	// Walk with error return
	visited := []string{}
	testErr := &testError{"intentional error"}
	err := tr.Walk(ctx, func(c context.Context, v *string) error {
		visited = append(visited, *v)
		if len(visited) == 2 {
			return testErr
		}
		return nil
	})

	if err != testErr {
		t.Fatalf("expected testErr to be returned, got %v", err)
	}

	if len(visited) != 2 {
		t.Fatalf("expected walk to stop after 2 nodes, got %d", len(visited))
	}
}

type testError struct {
	msg string
}

func (e *testError) Error() string {
	return e.msg
}

// Benchmark Tests

// BenchmarkInsertExact benchmarks inserting exact IPv4/IPv6 addresses into the tree
func BenchmarkInsertExact(b *testing.B) {
	ctx := context.Background()
	tree := NewTree[int]()
	keys := generateTestKeys(b.N)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := keys[i].key
		mask := keys[i].mask
		ival := i
		tree.Insert(ctx, key, mask, &ival)
	}
}

// BenchmarkSearchExact benchmarks exact searches in a pre-populated tree
func BenchmarkSearchExact(b *testing.B) {
	ctx := context.Background()
	tree := NewTree[int]()
	keys := generateTestKeys(b.N)

	// Pre-populate the tree
	for i := 0; i < b.N; i++ {
		key := keys[i].key
		mask := keys[i].mask
		ival := i
		tree.Insert(ctx, key, mask, &ival)
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
	tree := NewTree[int]()
	keys := generateTestKeys(b.N)

	// Pre-populate the tree
	for i := 0; i < b.N; i++ {
		key := keys[i].key
		mask := keys[i].mask
		ival := i
		tree.Insert(ctx, key, mask, &ival)
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
	tree := NewTree[int]()
	keys := generateTestKeys(b.N)

	// Pre-populate the tree
	for i := 0; i < b.N; i++ {
		key := keys[i].key
		mask := keys[i].mask
		ival := i
		tree.Insert(ctx, key, mask, &ival)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := keys[i].key
		mask := keys[i].mask
		tree.Delete(ctx, key, mask)
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

			tree := NewTree[int]()
			for i := 0; i < tt.count; i++ {
				key := keys[i].key
				mask := keys[i].mask

				ival := i
				tree.Insert(ctx, key, mask, &ival)
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
	tree := NewTree[int]()
	keys := generateTestKeys(b.N)

	// Pre-populate 50% of keys
	for i := 0; i < b.N/2; i++ {
		key := keys[i].key
		mask := keys[i].mask
		ival := i
		tree.Insert(ctx, key, mask, &ival)
	}

	b.ResetTimer()

	// Mix of operations: 40% insert, 50% search, 10% delete
	for i := 0; i < b.N; i++ {
		op := i % 100
		idx := i % len(keys)
		key := keys[idx].key
		mask := keys[idx].mask

		if op < 40 {
			ival := i
			tree.Insert(ctx, key, mask, &ival)
		} else if op < 90 {
			tree.SearchExact(ctx, key, mask)
		} else {
			tree.Delete(ctx, key, mask)
		}
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
				tree := NewTree[int]()
				keys := generateTestKeys(1000)
				for i := 0; i < 1000; i++ {
					key := keys[i].key
					mask := keys[i].mask
					ival := i
					tree.Insert(ctx, key, mask, &ival)
				}
			},
		},
		{
			"Search_1000_Keys",
			func() {
				tree := NewTree[int]()
				keys := generateTestKeys(1000)
				for i := 0; i < 1000; i++ {
					key := keys[i].key
					mask := keys[i].mask
					ival := i
					tree.Insert(ctx, key, mask, &ival)
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
				key:  []byte{byte(i % 256), byte((i / 256) % 256), byte(i % 128), byte(i % 64)},
				mask: []byte{0xFF, 0xFF, 0xFF, 0xFF},
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
