package prefix_tree

import (
	"context"
	"testing"
)

func TestReversedStrings(t *testing.T) {
	st := NewStringsTree[int]()

	urls := []string{
		"google.com",
		"mail.google.com",
		"drive.google.com",
		"yahoo.com",
		"news.yahoo.com",
		"about.yahoo.com",
		"example.org",
		"sub.example.org",
		"test.sub.example.org",
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

func generateReversedStringKeys(n int) []string {
	keys := make([]string, n)
	for i := 0; i < n; i++ {
		keys[i] = string(rune(i)) + "_string_key"
	}
	return keys
}

// BenchmarkReversedStringsInsert benchmarks ReversedStringsTree.Insert
func BenchmarkReversedStringsTreeInsert(b *testing.B) {
	ctx := context.Background()
	rstree := NewReversedStringsTree[int]()
	stringKeys := generateReversedStringKeys(b.N)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ival := i
		rstree.Insert(ctx, stringKeys[i], &ival)
	}
}

// BenchmarkReversedStringsSearch benchmarks ReversedStringsTree.Search
func BenchmarkReversedStringsTreeSearch(b *testing.B) {
	ctx := context.Background()
	rstree := NewReversedStringsTree[int]()
	stringKeys := generateReversedStringKeys(b.N)

	// Pre-populate
	for i := 0; i < b.N; i++ {
		ival := i
		rstree.Insert(ctx, stringKeys[i], &ival)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rstree.Search(ctx, stringKeys[i])
	}
}
