package metricsql

import (
	"testing"
)

func TestOptimize(t *testing.T) {
	f := func(q, qOptimizedExpected string) {
		t.Helper()
		e, err := Parse(q)
		if err != nil {
			t.Fatalf("unexpected error in Parse(%q): %s", q, err)
		}
		e = Optimize(e)
		qOptimized := e.AppendString(nil)
		if string(qOptimized) != qOptimizedExpected {
			t.Fatalf("unexpected qOptimized;\ngot\n%s\nwant\n%s", qOptimized, qOptimizedExpected)
		}
	}
	f("foo", "foo")
	f("a + b", "a + b")
	f(`foo{label1="value1"} == bar`, `foo{label1="value1"} == bar{label1="value1"}`)
	f(`foo{label1="value1"} == bar{label2="value2"}`, `foo{label1="value1", label2="value2"} == bar{label1="value1", label2="value2"}`)
	f(`foo + bar{b=~"a.*", a!="ss"}`, `foo{a!="ss", b=~"a.*"} + bar{a!="ss", b=~"a.*"}`)
	f(`foo{bar="1"} / 234`, `foo{bar="1"} / 234`)
	f(`foo{bar="1"} / foo{bar="1"}`, `foo{bar="1"} / foo{bar="1"}`)
	f(`123 + foo{bar!~"xx"}`, `123 + foo{bar!~"xx"}`)
	f(`foo or bar{x="y"}`, `foo or bar{x="y"}`)
	f(`foo * on(bar) baz{a="b"}`, `foo * on (bar) baz{a="b"}`)
	f(`foo * on() baz{a="b"}`, `foo * on () baz{a="b"}`)
	f(`foo * on (x, y) group_left() baz{a="b"}`, `foo * on (x, y) group_left () baz{a="b"}`)
	f(`f(foo, bar{baz=~"sdf"} + aa{baz=~"axx", aa="b"})`, `f(foo, bar{aa="b", baz=~"axx", baz=~"sdf"} + aa{aa="b", baz=~"axx", baz=~"sdf"})`)
	f(`sum(foo, bar{baz=~"sdf"} + aa{baz=~"axx", aa="b"})`, `sum(foo, bar{aa="b", baz=~"axx", baz=~"sdf"} + aa{aa="b", baz=~"axx", baz=~"sdf"})`)
	f(`foo AND bar{baz="aa"}`, `foo{baz="aa"} and bar{baz="aa"}`)

	// aggregate funcs
	f(`sum(foo{bar="baz"}) / a{b="c"}`, `sum(foo{bar="baz"}) / a{b="c"}`)

	// unknown func
	f(`f(foo) + bar{baz="a"}`, `f(foo) + bar{baz="a"}`)

	// transform funcs
	f(`round(foo{bar="baz"}) + sqrt(a{z=~"c"})`, `round(foo{bar="baz", z=~"c"}) + sqrt(a{bar="baz", z=~"c"})`)
	f(`foo{bar="baz"} + SQRT(a{z=~"c"})`, `foo{bar="baz", z=~"c"} + SQRT(a{bar="baz", z=~"c"})`)

	// multilevel transform funcs
	f(`round(sqrt(foo)) + bar`, `round(sqrt(foo)) + bar`)

	// unsupported transform funcs
	f(`absent(foo{bar="baz"}) + sqrt(a{z=~"c"})`, `absent(foo{bar="baz"}) + sqrt(a{z=~"c"})`)
	f(`ABSENT(foo{bar="baz"}) + sqrt(a{z=~"c"})`, `ABSENT(foo{bar="baz"}) + sqrt(a{z=~"c"})`)

	// rollup funcs
	f(`RATE(foo[5m]) / rate(baz{a="b"}) + increase(x{y="z"} offset 5i)`, `(RATE(foo[5m]) / rate(baz{a="b"})) + increase(x{y="z"} offset 5i)`)

	// subqueries
	f(`rate(avg_over_time(foo[5m:])) + bar{baz="a"}`, `rate(avg_over_time(foo[5m:])) + bar{baz="a"}`)

	// binary ops with constants or scalars
	f(`100 * foo / bar{baz="a"}`, `(100 * foo{baz="a"}) / bar{baz="a"}`)
	f(`foo * 100 / bar{baz="a"}`, `(foo{baz="a"} * 100) / bar{baz="a"}`)
	f(`foo / bar{baz="a"} * 100`, `(foo{baz="a"} / bar{baz="a"}) * 100`)
	f(`scalar(x) * foo / bar{baz="a"}`, `(scalar(x) * foo{baz="a"}) / bar{baz="a"}`)
	f(`SCALAR(x) * foo / bar{baz="a"}`, `(SCALAR(x) * foo{baz="a"}) / bar{baz="a"}`)
	f(`100 * on(foo) bar{baz="z"} + a`, `(100 * on (foo) bar{baz="z"}) + a`)
}
