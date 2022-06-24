package metricsql

import (
	"regexp"
	"testing"
)

func TestRegexpCache(t *testing.T) {
	fn := func(limit int, regexps []string, expectedEntries, expectedChars int) {
		t.Helper()
		rc := newRegexpCache(limit)
		for _, re := range regexps {
			r, err := regexp.Compile(re)
			rcv := &regexpCacheValue{
				r:   r,
				err: err,
			}
			rc.Put(re, rcv)
			rcv1 := rc.Get(re)
			if rcv1 != rcv {
				t.Fatalf("unexpected result for regexp %q; got\n%v\nwant\n%v", re, rcv1, rcv)
			}
		}
		if requests := rc.Requests(); requests != uint64(len(regexps)) {
			t.Fatalf("unexpected number of requests; got %d; want %d", requests, len(regexps))
		}
		if misses := rc.Misses(); misses != 0 {
			t.Fatalf("unexpected number of misses; got %d; want 0", misses)
		}
		rcv := rc.Get("non-existing-regexp")
		if rcv != nil {
			t.Fatalf("expecting nil entry; got %v", rcv)
		}
		if misses := rc.Misses(); misses != 1 {
			t.Fatalf("unexpected number of misses; got %d; want 1", misses)
		}
		if entries := rc.Len(); entries != uint64(expectedEntries) {
			t.Fatalf("unexpected number of entries; got %d; want %d", entries, expectedEntries)
		}
		if chars := rc.CharsCurrent(); chars != expectedChars {
			t.Fatalf("unexpected charsCurrent; got %d; want %d", chars, expectedChars)
		}
	}

	fn(10, []string{"a", "b", "c"}, 3, 6)
	fn(2, []string{"a", "b", "c"}, 2, 4) // overflow by 1 entry is allowed
	fn(2, []string{"a", "b", "c", "d"}, 2, 4)
	fn(1, []string{"a", "b", "c"}, 1, 2)
	fn(2, []string{"abcd", "efgh", "ijkl"}, 1, 8)   // overflow by 1 entry is allowed
	fn(5, []string{"123", "fd{456", "789"}, 1, 6)   // overflow by 1 entry is allowed
	fn(18, []string{"123", "fd{456", "789"}, 3, 24) // overflow by 1 entry is allowed
	fn(24, []string{"123", "fd{456", "789"}, 3, 24)
	fn(30, []string{"123", "fd{456", "789"}, 3, 24)
}
