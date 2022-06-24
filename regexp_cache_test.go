package metricsql

import "testing"

func TestRegexpCache_Put(t *testing.T) {
	fn := func(limit int, put []string, expChars int) {
		t.Helper()
		rc := &regexpCache{
			m:          make(map[string]*regexpCacheValue),
			charsLimit: limit,
		}
		for _, p := range put {
			rc.Put(p, nil)
		}
		if rc.chars != expChars {
			t.Errorf("expected cache to contain %d chars; got %d", expChars, rc.chars)
		}
	}

	fn(10, []string{"a", "b", "c"}, 3)
	fn(2, []string{"a", "b", "c"}, 3) // overflow by 1 entry is allowed
	fn(2, []string{"a", "b", "c", "d"}, 3)
	fn(1, []string{"a", "b", "c"}, 2)
	fn(2, []string{"abcd", "efgh", "ijkl"}, 4) // overflow by 1 entry is allowed
	fn(4, []string{"123", "456", "789"}, 6)    // overflow by 1 entry is allowed
}
