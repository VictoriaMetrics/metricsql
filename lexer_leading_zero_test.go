package metricsql

import "testing"

func TestParseLeadingZeroNumbers(t *testing.T) {
	f := func(expr string, expectedValue float64) {
		t.Helper()
		e, err := Parse(expr)
		if err != nil {
			t.Fatalf("cannot parse %q: %s", expr, err)
		}
		n, ok := e.(*NumberExpr)
		if !ok {
			return
		}
		if n.N != expectedValue {
			t.Fatalf("unexpected value for %q: got %v, want %v", expr, n.N, expectedValue)
		}
	}
	// Numbers with leading zeros must parse as decimal, like in Prometheus/Mimir.
	// See https://github.com/VictoriaMetrics/VictoriaMetrics/issues/11621
	f("hour() > 09", 9)
	f("hour() >= 08", 8)
	f("09 + 007", 16)
	f("09.5", 9.5)
	f("foo{bar='baz'} > 0099", 99)
	f("00.8", 0.8)
	f("07e8", 7e8)
	f("07_8", 78)

	// Valid octal, hex, binary and 0o forms must keep working.
	f("0123", 83)
	f("0x1f", 31)
	f("0b1010", 10)
	f("0o17", 15)
}
