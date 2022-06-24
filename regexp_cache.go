package metricsql

import (
	"regexp"
	"sync"
	"sync/atomic"

	"github.com/VictoriaMetrics/metrics"
)

// CompileRegexpAnchored returns compiled regexp `^re$`.
func CompileRegexpAnchored(re string) (*regexp.Regexp, error) {
	reAnchored := "^(?:" + re + ")$"
	return CompileRegexp(reAnchored)
}

// CompileRegexp returns compile regexp re.
func CompileRegexp(re string) (*regexp.Regexp, error) {
	rcv := regexpCacheV.Get(re)
	if rcv != nil {
		return rcv.r, rcv.err
	}
	r, err := regexp.Compile(re)
	rcv = &regexpCacheValue{
		r:   r,
		err: err,
	}
	regexpCacheV.Put(re, rcv)
	return rcv.r, rcv.err
}

const regexpCacheCharsMax = 1e6

var regexpCacheV = func() *regexpCache {
	rc := &regexpCache{
		m:          make(map[string]*regexpCacheValue),
		charsLimit: regexpCacheCharsMax,
	}
	metrics.NewGauge(`vm_cache_requests_total{type="promql/regexp"}`, func() float64 {
		return float64(rc.Requests())
	})
	metrics.NewGauge(`vm_cache_misses_total{type="promql/regexp"}`, func() float64 {
		return float64(rc.Misses())
	})
	metrics.NewGauge(`vm_cache_entries{type="promql/regexp"}`, func() float64 {
		return float64(rc.Len())
	})
	metrics.NewGauge(`vm_cache_chars{type="promql/regexp"}`, func() float64 {
		return float64(rc.chars)
	})
	return rc
}()

type regexpCacheValue struct {
	r   *regexp.Regexp
	err error
}

type regexpCache struct {
	// Move atomic counters to the top of struct for 8-byte alignment on 32-bit arch.
	// See https://github.com/VictoriaMetrics/VictoriaMetrics/issues/212

	requests uint64
	misses   uint64

	// chars stores the total number of characters used in stored regexps.
	// is used for memory usage estimation
	chars int
	// charsLimit limits the max number of chars stored in cache across all entries.
	// we limit by number of chars since calculating the exact size of each regexp is problematic,
	// while using chars seems like universal approach for simple and complex expressions.
	charsLimit int

	m  map[string]*regexpCacheValue
	mu sync.RWMutex
}

func (rc *regexpCache) Requests() uint64 {
	return atomic.LoadUint64(&rc.requests)
}

func (rc *regexpCache) Misses() uint64 {
	return atomic.LoadUint64(&rc.misses)
}

func (rc *regexpCache) Len() uint64 {
	rc.mu.RLock()
	n := len(rc.m)
	rc.mu.RUnlock()
	return uint64(n)
}

func (rc *regexpCache) Get(regexp string) *regexpCacheValue {
	atomic.AddUint64(&rc.requests, 1)

	rc.mu.RLock()
	rcv := rc.m[regexp]
	rc.mu.RUnlock()

	if rcv == nil {
		atomic.AddUint64(&rc.misses, 1)
	}
	return rcv
}

func (rc *regexpCache) Put(regexp string, rcv *regexpCacheValue) {
	rc.mu.Lock()
	if rc.chars-rc.charsLimit > 0 {
		// Remove items accounting for 10% chars from the cache.
		overflow := int(float64(rc.charsLimit) * 0.1)
		for k := range rc.m {
			delete(rc.m, k)

			size := len(regexp)
			overflow -= size
			rc.chars -= size

			if overflow <= 0 {
				break
			}
		}
	}
	rc.m[regexp] = rcv
	rc.chars += len(regexp)
	rc.mu.Unlock()
}
