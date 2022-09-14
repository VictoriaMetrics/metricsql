package metricsql

import (
	"fmt"
	"regexp"
	"testing"
	"time"
)

func TestRegexpCacheConcurrent(t *testing.T) {
	goroutines := 5
	maxChars := 1000
	rc := newRegexpCache(maxChars)
	resultCh := make(chan error, goroutines)
	for i := 0; i < goroutines; i++ {
		go func() {
			resultCh <- testRegexpCache(rc)
		}()
	}
	timer := time.NewTimer(time.Second * 5)
	for i := 0; i < goroutines; i++ {
		select {
		case <-timer.C:
			t.Fatalf("timeout")
		case err := <-resultCh:
			if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}
		}
	}
	maxChars += int(float64(maxChars) * 0.1)
	if chars := rc.CharsCurrent(); chars > maxChars {
		t.Fatalf("too many chars in the regexpCache; got %d; expected no more than %d", chars, maxChars)
	}
}

func testRegexpCache(rc *regexpCache) error {
	for i := 0; i < 10000; i++ {
		key := fmt.Sprintf("foo|regexp-%d", i)
		rcv := rc.Get(key)
		if rcv != nil {
			if rcv.err != nil {
				return fmt.Errorf("unexpected error obtained for key %q: %w", key, rcv.err)
			}
			if re := rcv.r.String(); re != key {
				return fmt.Errorf("unexpected regexp obtained for key %q: %q; want %q", key, re, key)
			}
		} else {
			r, err := regexp.Compile(key)
			rcv := &regexpCacheValue{
				r:   r,
				err: err,
			}
			rc.Put(key, rcv)
		}
	}
	return nil
}

func TestRegexpCache(t *testing.T) {
	fn := func(maxChars int, regexps []string, expectedEntries, expectedChars int) {
		t.Helper()
		rc := newRegexpCache(maxChars)
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
		if entries := rc.Len(); entries != expectedEntries {
			t.Fatalf("unexpected number of entries; got %d; want %d", entries, expectedEntries)
		}
		if chars := rc.CharsCurrent(); chars != expectedChars {
			t.Fatalf("unexpected charsCurrent; got %d; want %d", chars, expectedChars)
		}
	}

	fn(10, []string{"a", "b", "c"}, 3, 3)
	fn(2, []string{"a", "b", "c"}, 3, 3) // overflow by 1 entry is allowed
	fn(2, []string{"a", "b", "c", "d"}, 3, 3)
	fn(1, []string{"a", "b", "c"}, 2, 2)           // overflow by 1 tnery is allowed
	fn(2, []string{"abcd", "efgh", "ijkl"}, 1, 4)  // overflow by 1 entry is allowed
	fn(2, []string{"123", "fd{456", "789"}, 1, 3)  // overflow by 1 entry is allowed
	fn(9, []string{"123", "fd{456", "789"}, 3, 12) // overflow by 1 entry is allowed
	fn(12, []string{"123", "fd{456", "789"}, 3, 12)
	fn(15, []string{"123", "fd{456", "789"}, 3, 12)
}
