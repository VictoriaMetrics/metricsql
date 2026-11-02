package metricsql

import (
	"testing"
)

func TestIsAggrFuncModifierSuccess(t *testing.T) {
	f := func(s string) {
		t.Helper()
		if !isAggrFuncModifier(s) {
			t.Fatalf("expecting valid funcModifier: %q", s)
		}
	}
	f("by")
	f("BY")
	f("without")
	f("Without")
}

func TestIsAggrFuncModifierError(t *testing.T) {
	f := func(s string) {
		t.Helper()
		if isAggrFuncModifier(s) {
			t.Fatalf("unexpected valid funcModifier: %q", s)
		}
	}
	f("byfix")
	f("on")
	f("ignoring")
}

func TestIsAggrFuncSuccess(t *testing.T) {
	f := func(s string) {
		t.Helper()
		if !IsAggrFunc(s) {
			t.Fatalf("expecting valid aggrFunc: %q", s)
		}
	}
	f("sum")
	f("SUM")
	f("avg")
	f("share")
	f("Share")
	f("utilization")
	f("UTILIZATION")
}

func TestIsAggrFuncError(t *testing.T) {
	f := func(s string) {
		t.Helper()
		if IsAggrFunc(s) {
			t.Fatalf("unexpected valid aggrFunc: %q", s)
		}
	}
	f("unknown")
	f("rate")
	f("round")
}
