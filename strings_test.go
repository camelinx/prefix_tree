package prefix_tree

import (
	"context"
	"testing"
)

func TestStrings(t *testing.T) {
	st := NewStringsTree[*int]()

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

	for i, url := range urls {
		ival := i
		res, err := st.Insert(ctx, url, &ival)
		if err != nil || res != Ok {
			t.Fatalf("Failed to insert %s", url)
		}
	}

	expectedValuesCount := len(urls)
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

func TestPrefixStrings(t *testing.T) {
	st := NewStringsTree[int]()

	prefixUrls := []string{
		"/api/v1",
		"/api/v2",
		"/home",
		"/about",
		"/contact",
	}

	searchUrls := map[string]int{
		"/api/v1/resource":              0,
		"/api/v1/resource/123":          0,
		"/api/v1/resource/456":          0,
		"/api/v2/resource":              1,
		"/api/v2/resource/789":          1,
		"/home/user":                    2,
		"/home/user/downloads/file.jpg": 2,
		"/about/company.htm":            3,
		"/contact/address.html":         4,
	}

	failUrls := []string{
		"/api/v3/resource",
		"/hom/user",
		"/aboot/company.htm",
		"/conact/address.html",
	}

	ctx := context.Background()
	for i, prefixUrl := range prefixUrls {
		ival := i
		res, err := st.Insert(ctx, prefixUrl, ival)
		if err != nil || res != Ok {
			t.Fatalf("Failed to insert %s", prefixUrl)
		}
	}

	for searchUrl, i := range searchUrls {
		ival := i
		res, pival, err := st.Search(ctx, searchUrl)
		if err != nil || res != PartialMatch {
			t.Fatalf("Failed to prefix find api %s", searchUrl)
		}
		if pival != ival {
			t.Fatalf("Expected value %d for prefix api %s, got %d", ival, searchUrl, ival)
		}

		res, _, err = st.SearchExact(ctx, searchUrl)
		if err == nil || res != Error {
			t.Fatalf("Found prefix api %s in exact match", searchUrl)
		}
	}

	for _, failUrl := range failUrls {
		res, _, err := st.Search(ctx, failUrl)
		if err == nil || res != Error {
			t.Fatalf("Found prefix api %s", failUrl)
		}

		res, _, err = st.SearchExact(ctx, failUrl)
		if err == nil || res != Error {
			t.Fatalf("Found prefix api %s in exact match", failUrl)
		}
	}

	for i, prefixUrl := range prefixUrls {
		ival := i
		res, pival, err := st.Delete(ctx, prefixUrl)
		if err != nil || res != Match {
			t.Fatalf("Failed to delete %s", prefixUrl)
		}
		if pival != ival {
			t.Fatalf("Invalid value for deleted %s: expected %d, got %v", prefixUrl, ival, pival)
		}

		res, _, err = st.Search(ctx, prefixUrl)
		if nil == err || res != Error {
			t.Fatalf("Found non existent key %s", prefixUrl)
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
		stree.Insert(ctx, stringKeys[i], ival)
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
		stree.Insert(ctx, stringKeys[i], ival)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		stree.Search(ctx, stringKeys[i])
	}
}
