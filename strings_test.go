package prefix_tree

import (
	"context"
	"testing"
)

func TestStrings(t *testing.T) {
	st := NewStringsTree[int]()

	urls := []string{
		"/api/v1/resource",
		"/api/v1/resource/123",
		"/api/v1/resource/456",
		"/api/v2/resource",
		"/api/v2/resource/789",
		"/home",
		"/about",
		"/contact",
	}

	ctx := context.Background()
	for i, url := range urls {
		ival := i
		res, err := st.Insert(ctx, url, &ival)
		if err != nil || res != Ok {
			t.Fatalf("Failed to insert %s", url)
		}

		res, pival, err := st.Search(ctx, url)
		if err != nil || res != Match {
			t.Fatalf("Failed to find %s", url)
		}
		if pival == nil || *pival != ival {
			t.Fatalf("Invalid value for %s: expected %d, got %v", url, ival, pival)
		}

		res, pival, err = st.SearchExact(ctx, url)
		if err != nil || res != Match {
			t.Fatalf("Failed to find (exact) %s", url)
		}
		if pival == nil || *pival != ival {
			t.Fatalf("Invalid value for (exact) %s: expected %d, got %v", url, ival, pival)
		}

		res, err = st.Insert(ctx, url, &ival)
		if err != nil || res != Dup {
			t.Fatalf("Failed to recognize %s as duplicate", url)
		}

		res, pival, err = st.Delete(ctx, url)
		if err != nil || res != Match {
			t.Fatalf("Failed to delete %s", url)
		}
		if pival == nil || *pival != ival {
			t.Fatalf("Invalid value for deleted %s: expected %d, got %v", url, ival, pival)
		}

		res, _, err = st.Search(ctx, url)
		if nil == err || res != Error {
			t.Fatalf("Found non existent key %s", url)
		}
	}
}

func generateStringKeys(n int) []string {
	keys := make([]string, n)
	for i := 0; i < n; i++ {
		keys[i] = "string_key_" + string(rune(i))
	}
	return keys
}

// BenchmarkStringsInsert benchmarks StringsTree.Insert
func BenchmarkStringsTreeInsert(b *testing.B) {
	ctx := context.Background()
	stree := NewStringsTree[int]()
	stringKeys := generateStringKeys(b.N)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ival := i
		stree.Insert(ctx, stringKeys[i], &ival)
	}
}

// BenchmarkStringsSearch benchmarks StringsTree.Search
func BenchmarkStringsTreeSearch(b *testing.B) {
	ctx := context.Background()
	stree := NewStringsTree[int]()
	stringKeys := generateStringKeys(b.N)

	// Pre-populate
	for i := 0; i < b.N; i++ {
		ival := i
		stree.Insert(ctx, stringKeys[i], &ival)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		stree.Search(ctx, stringKeys[i])
	}
}
