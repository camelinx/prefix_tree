# Benchmark Quick Reference

## Quick Start Commands

### See all available benchmarks:
```bash
go test -bench=. -run=^$ --list 2>&1 | grep Benchmark
```

### Run specific benchmarks:
```bash
# Just insert benchmarks
go test -bench=Insert -benchmem -run=^$

# Just GC pressure tests
go test -bench=GCPressure -benchmem -run=^$ -v

# IPv4 tests only
go test -bench=V4Tree -benchmem -run=^$
```

---

## GC Pressure Analysis Commands

### Option 1: Simple Memory Report (Recommended for Quick Checks)
```bash
go test -bench=BenchmarkGCPressure -benchmem -run=^$ -v
```

**Output shows:**
- Total bytes allocated
- GC cycles triggered  
- Bytes per operation

### Option 2: Detailed Profiling with pprof
```bash
# Generate memory profile
go test -bench=BenchmarkGCPressure -memprofile=mem.prof -run=^$ -timeout=30s

# Analyze
go tool pprof mem.prof
# Type: top10  (see top memory allocations)
# Type: alloc_space (cumulative allocations)
```

### Option 3: Real-time GC Statistics
```bash
GODEBUG=gctrace=1 go test -bench=BenchmarkGCPressure -run=^$ 2>&1 | grep "gc"
```

**Output shows:**
```
gc 1 @0.234s 0%: 0.045+0.38+0.001 ms clock, 0.36+0.10/0.26/0.26+0.008 ms cpu
```
- `0.045 ms` = mark phase time
- `0.38 ms` = sweep phase time
- `0%` = total GC time percentage

### Option 4: Memory Allocation Breakdown
```bash
go test -bench=BenchmarkAllocations -benchmem -run=^$ -v
```

**Shows:** `mallocs` metric - raw count of malloc calls

### Option 5: Full Benchmark Run with All Metrics
```bash
go test -bench=. -benchmem -run=^$ -cpuprofile=cpu.prof -memprofile=mem.prof -timeout=60s
```

Then analyze:
```bash
go tool pprof -http=:8080 mem.prof   # Opens interactive UI in browser
go tool pprof -top mem.prof           # Shows top allocators in terminal
```

---

## Understanding Benchmark Output

```
BenchmarkInsertExact-12  3487960  306.2 ns/op  16 B/op  1 allocs/op
```

| Part | Meaning | Good Range |
|------|---------|-----------|
| `3487960` | Operations run | N/A |
| `306.2 ns/op` | Time per operation | Lower is better |
| `16 B/op` | Bytes allocated per op | < 100 B |
| `1 allocs/op` | Allocations per op | < 5 |

### GC Pressure Benchmark Output:

```
BenchmarkGCPressure/Large_IPv4-12
  ... 5177824 bytes_alloc  517.8 bytes_per_insert  1 gc_runs
```

| Metric | Interpretation |
|--------|-----------------|
| `5177824 bytes_alloc` | Total heap used in test |
| `517.8 bytes_per_insert` | Average per operation - **watch this** |
| `1 gc_runs` | GC cycles triggered - if > 3 for 10K ops, pressure is high |

---

## Comparing Before/After Optimization

```bash
# Before change
go test -bench=BenchmarkGCPressure -benchmem -run=^$ > before.txt

# Make your optimization changes...

# After change
go test -bench=BenchmarkGCPressure -benchmem -run=^$ > after.txt

# Install benchstat if needed
go install golang.org/x/perf/cmd/benchstat@latest

# Compare
benchstat before.txt after.txt
```

**Example output:**
```
name                  old time/op    new time/op    delta
BenchmarkInsertExact       306.2 ns/op    280.1 ns/op   -8.51%
BenchmarkSearchExact      1105 ns/op    950 ns/op     -14.0%
```

---

## Key Metrics to Watch for GC Pressure

### Red Flags (High GC Pressure):
- ❌ `gc_runs > 3` for 10K inserts
- ❌ `bytes_per_insert > 1000`
- ❌ `mallocs/op > 20` in BenchmarkAllocations
- ❌ GC time > 5% in GODEBUG output

### Healthy Benchmarks:
- ✅ `gc_runs <= 1` for 10K inserts
- ✅ `bytes_per_insert < 600` for IPv4
- ✅ `mallocs/op < 10` in normal benchmarks
- ✅ GC time < 2%

---

## Profiling Deep Dive

### CPU Profiling:
```bash
go test -bench=BenchmarkInsertExact -cpuprofile=cpu.prof -run=^$
go tool pprof -http=:8080 cpu.prof
# Click "Flame Graph" to visualize hot paths
```

### Memory Allocation:
```bash
go test -bench=BenchmarkInsertExact -memprofile=mem.prof -run=^$ -count=3
go tool pprof -alloc_space mem.prof  # Cumulative allocations
go tool pprof -alloc_objects mem.prof  # Count of objects
```

### Memory In-Use:
```bash
go tool pprof -inuse_space mem.prof  # Current heap usage
go tool pprof -inuse_objects mem.prof  # Current object count
```

---

## Test Scenarios

### Scenario 1: Check GC for small dataset (1K insertions)
```bash
go test -bench=BenchmarkGCPressure/Small -run=^$
# GC should be 0 - if not, high allocation pressure
```

### Scenario 2: Check GC for large dataset (10K insertions)
```bash
go test -bench=BenchmarkGCPressure/Large -run=^$
# Expect 1 GC cycle; if more, investigate allocations
```

### Scenario 3: Mixed realistic workload
```bash
go test -bench=BenchmarkMixedOperations -benchmem -run=^$ -count=5
# Runs 5 times, good for variance detection
```

### Scenario 4: IPv6 vs IPv4 performance
```bash
go test -bench=V[46]Tree -benchmem -run=^$
# Should show IPv4 ~10x faster than IPv6
```

### Scenario 5: Full profile for bottleneck identification
```bash
go test -bench=. -cpuprofile=cpu.prof -memprofile=mem.prof -run=^$
go tool pprof -http=:8080 cpu.prof
# Navigate to "Flame Graph" for visual analysis
```

---

## Common Issues & Solutions

### Issue: Too much GC activity
```bash
# Diagnose
go test -bench=BenchmarkGCPressure -run=^$ -v | grep "gc_runs"

# If gc_runs high for 1K inserts:
# - Likely: Many small allocations in hot path
# - Fix: Use sync.Pool for frequently allocated objects
# - Fix: Reduce allocations in find() or Insert()
```

### Issue: High bytes_per_insert
```bash
# If > 1000 bytes per insert:
# - Each node allocates ~1KB (parent, left, right pointers, interface{})
# - Fix: Consider radix trie to reduce node count
# - Fix: Use pointer-less compact representation
```

### Issue: Variable benchmark results
```bash
# Run with more iterations for stability
go test -bench=BenchmarkInsertExact -count=5 -benchmem -run=^$

# Use benchstat to analyze variance
benchstat results1.txt results2.txt
```

---

## Performance Targets

Based on current results, these are good targets:

- **Insert speed:** 300 ns/op (currently 306)
- **Search speed:** 1000 ns/op (currently 1105)
- **Bytes/insert:** < 600 B (currently 518-1225)
- **GC cycles:** 0-1 for 10K ops (currently 1)
- **Memory/op:** < 2 B/op (currently 16 B/op for insert due to test overhead)

---

## Further Reading

- [Go testing package](https://golang.org/pkg/testing/)
- [Profiling Go programs](https://github.com/golang/go/wiki/profiling)
- [pprof visualization guide](https://github.com/google/pprof/tree/master/doc)
- [Garbage Collection Tuning](https://golang.org/doc/gc-guide)
