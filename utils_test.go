package metricsql

import (
	"testing"
)

func TestExpandWithExprsSuccess(t *testing.T) {
	f := func(q, qExpected string) {
		t.Helper()
		for i := 0; i < 3; i++ {
			qExpanded, err := ExpandWithExprs(q)
			if err != nil {
				t.Fatalf("unexpected error when expanding %q: %s", q, err)
			}
			if qExpanded != qExpected {
				t.Fatalf("unexpected expanded expression for %q;\ngot\n%q\nwant\n%q", q, qExpanded, qExpected)
			}
		}
	}

	f(`1`, `1`)
	f(`foobar`, `foobar`)
	f(`with (x = 1) x+x`, `2`)
	f(`with (f(x) = x*x) 3+f(2)+2`, `9`)
}

func TestExpandWithExprsError(t *testing.T) {
	f := func(q string) {
		t.Helper()
		for i := 0; i < 3; i++ {
			qExpanded, err := ExpandWithExprs(q)
			if err == nil {
				t.Fatalf("expecting non-nil error when expanding %q", q)
			}
			if qExpanded != "" {
				t.Fatalf("unexpected non-empty qExpanded=%q", qExpanded)
			}
		}
	}

	f(``)
	f(`  with (`)
}

func TestVisitAll(t *testing.T) {
	f := func(q, sExpected string) {
		t.Helper()
		expr, err := Parse(q)
		if err != nil {
			t.Fatalf("unexpected error: %s", err)
		}
		var buf []byte
		VisitAll(expr, func(e Expr) {
			buf = e.AppendString(buf)
			buf = append(buf, ',')
		})
		if string(buf) != sExpected {
			t.Fatalf("unexpected result; got\n%q\nwant\n%q", buf, sExpected)
		}
	}
	f("123", "123,")
	f("1+2", "3,")
	f("1+a", "1,a,(),(),1 + a,")
	f("avg(a<b+1, sum(x) by (y))", "a,b,1,(),(),b + 1,(),(),a < (b + 1),x,by(y),sum(x) by(y),(),avg(a < (b + 1), sum(x) by(y)),")
	f("x[1s]", "x,x[1s],")
}

func TestIsSupportedFunction(t *testing.T) {
	f := func(s string, expectedResult bool) {
		t.Helper()
		result := IsSupportedFunction(s)
		if result != expectedResult {
			t.Fatalf("unexpected result for IsSupportedFunction(%q); got %v; want %v", s, result, expectedResult)
		}
	}

	// empty function name is a synonim to union()
	f("", true)
	f("union", true)

	// rollup function
	f("rate", true)
	f("RATE", true)
	f("Increase", true)

	// transform function
	f("ceil", true)
	f("histogram_QUANTILe", true)

	// aggregate function
	f("sum", true)
	f("aVG", true)

	// Unknown function
	f("foo", false)
	f("BAR", false)
}
