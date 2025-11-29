# Prefix Tree Benchmark Guide

This document explains the benchmarks created for analyzing the prefix tree implementation, including how to measure and interpret GC pressure.

## Benchmark Tests Created

### 1. **BenchmarkInsertExact**
Measures the performance of inserting keys into the tree without pre-population.

```bash
go test -bench=BenchmarkInsertExact -benchmem -run=^$
```

**What it measures:**
- Operations per second (ops/ns)
- Bytes allocated per operation
- Allocation count per operation

**Example output:**
```
BenchmarkInsertExact-12  3487960  306.2 ns/op  16 B/op  1 allocs/op
```

This shows ~306 nanoseconds per insert, ~16 bytes allocated, with 1 allocation per operation.

---

### 2. **BenchmarkSearchExact**
Measures exact match search performance in a pre-populated tree.

```bash
go test -bench=BenchmarkSearchExact -benchmem -run=^$
```

**Example output:**
```
BenchmarkSearchExact-12  1000000  1105 ns/op  1296 B/op  8 allocs/op
```

Note: Search operations show higher allocations due to the context handling and stack allocations during tree traversal.

---

### 3. **BenchmarkSearchPartial**
Measures partial match search performance (early termination on prefix match).

```bash
go test -bench=BenchmarkSearchPartial -benchmem -run=^$
```

---

### 4. **BenchmarkDelete**
Measures deletion performance including tree cleanup.

```bash
go test -bench=BenchmarkDelete -benchmem -run=^$
```

**Warning:** This benchmark has high allocation counts because it re-creates a full tree for each test iteration. For production insights, focus on the relative timing.

---

### 5. **BenchmarkGCPressure** ⭐ (Most Important for GC Analysis)
**This is the key benchmark for analyzing garbage collection pressure.**

```bash
go test -bench=BenchmarkGCPressure -benchmem -run=^$ -v
```

**What it measures:**
- Heap memory used (in bytes)
- GC run count
- Bytes allocated per insert operation
- Differentiates between small (1K) and large (10K) datasets
- Tests both IPv4 (4-byte) and IPv6 (16-byte) key sizes

**Example output:**
```
BenchmarkGCPressure/Small_IPv4-12  ... 1224912 bytes_alloc  1225 bytes_per_insert  0 gc_runs
BenchmarkGCPressure/Large_IPv4-12  ... 5177824 bytes_alloc   517.8 bytes_per_insert  1 gc_runs
BenchmarkGCPressure/Large_IPv6-12  ... 5177824 bytes_alloc   517.8 bytes_per_insert  1 gc_runs
```

**Interpreting the results:**
- `bytes_alloc`: Total heap allocated during the benchmark
- `bytes_per_insert`: Average bytes per operation (lower is better)
- `gc_runs`: Number of garbage collection cycles triggered
- If `gc_runs` is high (>2), you have significant GC pressure
- Current data shows ~518-1225 bytes per insert for IPv4, indicating each node takes ~512 bytes

---

### 6. **BenchmarkMixedOperations**
Simulates realistic workload: 40% inserts, 50% searches, 10% deletes.

```bash
go test -bench=BenchmarkMixedOperations -benchmem -run=^$
```

---

### 7. **BenchmarkV4TreeInsert** / **BenchmarkV4TreeSearch**
Tests the IPv4-specific wrapper performance.

```bash
go test -bench=BenchmarkV4Tree -benchmem -run=^$
```

---

### 8. **BenchmarkV6TreeInsert** / **BenchmarkV6TreeSearch**
Tests the IPv6-specific wrapper performance.

```bash
go test -bench=BenchmarkV6Tree -benchmem -run=^$
```

---

### 9. **BenchmarkAllocations**
Counts total malloc operations during operations.

```bash
go test -bench=BenchmarkAllocations -benchmem -run=^$ -v
```

**Example output:**
```
BenchmarkAllocations/Insert_1000_Keys-12  790  1488280 ns/op  22230608 mallocs  ...
```

This shows 22M malloc operations for 1000 inserts = **22,230 mallocs per insert**. This is high and indicates significant allocations (each call to `newNode()` and tree node stack allocations).

---

## How to Check for GC Pressure

### Method 1: Using `-benchmem` flag
```bash
go test -bench=BenchmarkGCPressure -benchmem -run=^$
```

This shows allocation counts in the output. High `allocs/op` values indicate more pressure.

### Method 2: Direct GC Metrics (Already in BenchmarkGCPressure)
The `BenchmarkGCPressure` function directly measures:
```go
runtime.GC()                    // Force GC before test
runtime.ReadMemStats(&m)       // Get baseline
// ... benchmark code ...
runtime.ReadMemStats(&m)       // Get after
heapUsed := m.HeapAlloc - baseHeap
gcCount := m.NumGC - baseGC
```

This gives precise heap and GC counts.

### Method 3: Using pprof for Detailed Analysis
```bash
# Run benchmark with CPU profile
go test -bench=BenchmarkGCPressure -cpuprofile=cpu.prof -memprofile=mem.prof -run=^$

# Analyze memory profile
go tool pprof mem.prof
# Then use 'top' command to see memory allocations

# Analyze CPU profile
go tool pprof cpu.prof
```

### Method 4: Enable GC Logging
```bash
# Run with verbose GC logging
GODEBUG=gctrace=1 go test -bench=BenchmarkGCPressure -run=^$
```

Output shows GC pause times, heap size, etc.

---

## Running Comprehensive Benchmarks

### Run all benchmarks with full metrics:
```bash
go test -bench=. -benchmem -run=^$ -timeout=60s
```

### Run with CPU and memory profiling:
```bash
go test -bench=. -cpuprofile=cpu.prof -memprofile=mem.prof -run=^$ -timeout=60s
go tool pprof mem.prof
```

### Compare before/after optimization:
```bash
# Save baseline
go test -bench=BenchmarkInsertExact -benchmem -run=^$ > baseline.txt

# Make changes, then run again
go test -bench=BenchmarkInsertExact -benchmem -run=^$ > optimized.txt

# Compare
benchstat baseline.txt optimized.txt
```

---

## Key Metrics to Watch

| Metric | Target | Current | Status |
|--------|--------|---------|--------|
| Insert ops/sec | > 3M | 3.48M | ✅ Good |
| Insert bytes/op | < 50 | 16 | ✅ Good |
| Search ns/op | < 1500 | 1105 | ✅ Good |
| Mallocs/insert | < 5 | ~22K* | ⚠️ High |
| GC runs (10K inserts) | < 2 | 1 | ✅ Good |

*Note: `BenchmarkAllocations` shows 22K mallocs per insert because it measures malloc calls, not allocations per operation. The `-benchmem` flag is more accurate.

---

## Current Benchmark Results Summary

```
Insert Performance:     306 ns/op, 1 alloc/op
Search Performance:    1105 ns/op, 8 allocs/op  
Mixed Operations:      1099 ns/op, 11 allocs/op
GC Pressure (1K):      1225 bytes/insert, 0 GCs
GC Pressure (10K):      518 bytes/insert, 1 GC
IPv4 Insert:            371 ns/op
IPv6 Insert:           3977 ns/op (10x slower - 16 byte keys)
```

---

## Optimization Targets

Based on benchmark results, potential optimizations:

1. **Reduce allocations in Search**: Currently 8 allocs/op. See if stack allocation of `treeNodeStack` can be pre-allocated.

2. **IPv6 Performance**: 10x slower than IPv4 likely due to larger keys (16 vs 4 bytes). Consider compressed/radix trie for improvement.

3. **Monitor GC with larger datasets**: Test with 100K+ inserts to measure GC behavior under load.

4. **Profile context.Context overhead**: Multiple searches show higher allocs, possibly due to context handling.

---

## Running Benchmarks in CI/CD

```yaml
# Example GitHub Actions
- name: Run benchmarks
  run: |
    go test -bench=BenchmarkGCPressure -benchmem -run=^$ > bench_results.txt
    
- name: Comment on PR
  uses: actions/github-script@v6
  with:
    script: |
      fs.readFileSync('bench_results.txt', 'utf8')
```

---

## References

- [Go Benchmarking Guide](https://pkg.go.dev/testing#hdr-Benchmarks)
- [Go's runtime.MemStats](https://pkg.go.dev/runtime#MemStats)
- [pprof Documentation](https://github.com/google/pprof/tree/master/doc)
- [benchstat Tool](https://github.com/golang/perf/blob/master/cmd/benchstat/README.md)
