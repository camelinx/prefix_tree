package prefix_tree

import (
	"context"
	"testing"
)

func TestReversedStrings(t *testing.T) {
	st := NewStringsTree[int]()

	domains := []string{
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
	for i, domain := range domains {
		ival := i
		res, err := st.Insert(ctx, domain, &ival)
		if err != nil || res != Ok {
			t.Fatalf("Failed to insert %s", domain)
		}

		res, pival, err := st.Search(ctx, domain)
		if err != nil || res != Match {
			t.Fatalf("Failed to find %s", domain)
		}
		if pival == nil || *pival != ival {
			t.Fatalf("Invalid value for %s: expected %d, got %v", domain, ival, pival)
		}

		res, pival, err = st.SearchExact(ctx, domain)
		if err != nil || res != Match {
			t.Fatalf("Failed to find (exact) %s", domain)
		}
		if pival == nil || *pival != ival {
			t.Fatalf("Invalid value for (exact) %s: expected %d, got %v", domain, ival, pival)
		}

		res, err = st.Insert(ctx, domain, &ival)
		if err != nil || res != Dup {
			t.Fatalf("Failed to recognize %s as duplicate", domain)
		}

		res, pival, err = st.Delete(ctx, domain)
		if err != nil || res != Match {
			t.Fatalf("Failed to delete %s", domain)
		}
		if pival == nil || *pival != ival {
			t.Fatalf("Invalid value for deleted %s: expected %d, got %v", domain, ival, pival)
		}

		res, _, err = st.Search(ctx, domain)
		if nil == err || res != Error {
			t.Fatalf("Found non existent key %s", domain)
		}
	}

	for i, domain := range domains {
		ival := i
		res, err := st.Insert(ctx, domain, &ival)
		if err != nil || res != Ok {
			t.Fatalf("Failed to insert %s", domain)
		}
	}

	expectedValuesCount := len(domains)
	walkedValuesCount := 0
	st.Walk(ctx, func(ctx context.Context, ival *int) error {
		if *ival >= expectedValuesCount {
			t.Fatalf("Unexpected value %d returned in walk. Expected a value less than %d", *ival, expectedValuesCount)
		}

		walkedValuesCount++
		return nil
	})

	if walkedValuesCount != expectedValuesCount {
		t.Fatalf("Expected %d value in walk. Actual walked values count is %d", expectedValuesCount, walkedValuesCount)
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
