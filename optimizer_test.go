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

	// supported binary expression
	f("a + b", "a + b")
	f(`foo{label1="value1"} == bar`, `foo{label1="value1"} == bar{label1="value1"}`)
	f(`foo{label1="value1"} == bar{label2="value2"}`, `foo{label1="value1", label2="value2"} == bar{label1="value1", label2="value2"}`)
	f(`foo + bar{b=~"a.*", a!="ss"}`, `foo{a!="ss", b=~"a.*"} + bar{a!="ss", b=~"a.*"}`)
	f(`foo{bar="1"} / 234`, `foo{bar="1"} / 234`)
	f(`foo{bar="1"} / foo{bar="1"}`, `foo{bar="1"} / foo{bar="1"}`)
	f(`123 + foo{bar!~"xx"}`, `123 + foo{bar!~"xx"}`)
	f(`foo or bar{x="y"}`, `foo or bar{x="y"}`)
	f(`foo{x="y"} * on() baz{a="b"}`, `foo{x="y"} * on () baz{a="b"}`)
	f(`foo{x="y"} * on(a) baz{a="b"}`, `foo{a="b", x="y"} * on (a) baz{a="b"}`)
	f(`foo{x="y"} * on(bar) baz{a="b"}`, `foo{x="y"} * on (bar) baz{a="b"}`)
	f(`foo{x="y"} * on(x,a,bar) baz{a="b"}`, `foo{a="b", x="y"} * on (x, a, bar) baz{a="b", x="y"}`)
	f(`foo{x="y"} * ignoring() baz{a="b"}`, `foo{a="b", x="y"} * ignoring () baz{a="b", x="y"}`)
	f(`foo{x="y"} * ignoring(a) baz{a="b"}`, `foo{x="y"} * ignoring (a) baz{a="b", x="y"}`)
	f(`foo{x="y"} * ignoring(bar) baz{a="b"}`, `foo{a="b", x="y"} * ignoring (bar) baz{a="b", x="y"}`)
	f(`foo{x="y"} * ignoring(x,a,bar) baz{a="b"}`, `foo{x="y"} * ignoring (x, a, bar) baz{a="b"}`)
	f(`foo{x="y"} * on(a) group_left baz{a="b"}`, `foo{a="b", x="y"} * on (a) group_left () baz{a="b"}`)
	f(`foo{x="y"} * on(a) group_right(x, y) baz{a="b"}`, `foo{a="b", x="y"} * on (a) group_right (x, y) baz{a="b"}`)
	f(`f(foo, bar{baz=~"sdf"} + aa{baz=~"axx", aa="b"})`, `f(foo, bar{aa="b", baz=~"axx", baz=~"sdf"} + aa{aa="b", baz=~"axx", baz=~"sdf"})`)
	f(`sum(foo, bar{baz=~"sdf"} + aa{baz=~"axx", aa="b"})`, `sum(foo, bar{aa="b", baz=~"axx", baz=~"sdf"} + aa{aa="b", baz=~"axx", baz=~"sdf"})`)
	f(`foo AND bar{baz="aa"}`, `foo{baz="aa"} and bar{baz="aa"}`)
	f(`{x="y",__name__="a"} + {a="b"}`, `a{a="b", x="y"} + {a="b", x="y"}`)
	f(`{x="y",__name__=~"a|b"} + {a="b"}`, `{__name__=~"a|b", a="b", x="y"} + {a="b", x="y"}`)
	f(`a{x="y",__name__=~"a|b"} + {a="b"}`, `a{__name__=~"a|b", a="b", x="y"} + {a="b", x="y"}`)

	// unsupported binary expression
	f(`foo{a="b"} or bar{x="y"}`, `foo{a="b"} or bar{x="y"}`)

	// aggregate funcs
	f(`sum(foo{bar="baz"}) / a{b="c"}`, `sum(foo{bar="baz"}) / a{b="c"}`)
	f(`sum(foo{bar="baz"}) by () / a{b="c"}`, `sum(foo{bar="baz"}) by () / a{b="c"}`)
	f(`sum(foo{bar="baz"}) by (bar) / a{b="c"}`, `sum(foo{bar="baz"}) by (bar) / a{b="c", bar="baz"}`)
	f(`sum(foo{bar="baz"}) by (b) / a{b="c"}`, `sum(foo{b="c", bar="baz"}) by (b) / a{b="c"}`)
	f(`sum(foo{bar="baz"}) by (x) / a{b="c"}`, `sum(foo{bar="baz"}) by (x) / a{b="c"}`)
	f(`sum(foo{bar="baz"}) by (bar,b) / a{b="c"}`, `sum(foo{b="c", bar="baz"}) by (bar, b) / a{b="c", bar="baz"}`)
	f(`sum(foo{bar="baz"}) without () / a{b="c"}`, `sum(foo{b="c", bar="baz"}) without () / a{b="c", bar="baz"}`)
	f(`sum(foo{bar="baz"}) without (bar) / a{b="c"}`, `sum(foo{b="c", bar="baz"}) without (bar) / a{b="c"}`)
	f(`sum(foo{bar="baz"}) without (b) / a{b="c"}`, `sum(foo{bar="baz"}) without (b) / a{b="c", bar="baz"}`)
	f(`sum(foo{bar="baz"}) without (x) / a{b="c"}`, `sum(foo{b="c", bar="baz"}) without (x) / a{b="c", bar="baz"}`)
	f(`sum(foo{bar="baz"}) without (bar,b) / a{b="c"}`, `sum(foo{bar="baz"}) without (bar, b) / a{b="c"}`)
	f(`sum(foo, bar) by (a) + baz{a="b"}`, `sum(foo{a="b"}, bar) by (a) + baz{a="b"}`)
	f(`topk(3, foo) by (baz,x) + bar{baz="a"}`, `topk(3, foo{baz="a"}) by (baz, x) + bar{baz="a"}`)
	f(`topk(a, foo) without (x,y) + bar{baz="a"}`, `topk(a, foo{baz="a"}) without (x, y) + bar{baz="a"}`)
	f(`a{b="c"} + quantiles("foo", 0.1, 0.2, bar{x="y"}) by (b, x, y)`, `a{b="c", x="y"} + quantiles("foo", 0.1, 0.2, bar{b="c", x="y"}) by (b, x, y)`)
	f(`count_values("foo", bar{baz="a"}) by (bar,b) + a{b="c"}`, `count_values("foo", bar{baz="a"}) by (bar, b) + a{b="c"}`)

	// unknown func
	f(`f(foo) + bar{baz="a"}`, `f(foo) + bar{baz="a"}`)

	// transform funcs
	f(`round(foo{bar="baz"}) + sqrt(a{z=~"c"})`, `round(foo{bar="baz", z=~"c"}) + sqrt(a{bar="baz", z=~"c"})`)
	f(`foo{bar="baz"} + SQRT(a{z=~"c"})`, `foo{bar="baz", z=~"c"} + SQRT(a{bar="baz", z=~"c"})`)
	f(`round({__name__="foo"}) + bar`, `round(foo) + bar`)
	f(`round({__name__=~"foo|bar"}) + baz`, `round({__name__=~"foo|bar"}) + baz`)
	f(`round({__name__=~"foo|bar",a="b"}) + baz`, `round({__name__=~"foo|bar", a="b"}) + baz{a="b"}`)
	f(`round({__name__=~"foo|bar",a="b"}) + sqrt(baz)`, `round({__name__=~"foo|bar", a="b"}) + sqrt(baz{a="b"})`)
	f(`round(foo) + {__name__="bar",x="y"}`, `round(foo{x="y"}) + bar{x="y"}`)
	f(`absent(foo{bar="baz"}) + sqrt(a{z=~"c"})`, `absent(foo{bar="baz"}) + sqrt(a{z=~"c"})`)
	f(`ABSENT(foo{bar="baz"}) + sqrt(a{z=~"c"})`, `ABSENT(foo{bar="baz"}) + sqrt(a{z=~"c"})`)
	f(`label_set(foo{bar="baz"}, "xx", "y") + a{x="y"}`, `label_set(foo{bar="baz"}, "xx", "y") + a{x="y"}`)
	f(`now() + foo{bar="baz"} + x{y="x"}`, `(now() + foo{bar="baz", y="x"}) + x{bar="baz", y="x"}`)
	f(`limit_offset(5, 10, {x="y"}) if {a="b"}`, `limit_offset(5, 10, {a="b", x="y"}) if {a="b", x="y"}`)
	f(`buckets_limit(aa, {x="y"}) if {a="b"}`, `buckets_limit(aa, {a="b", x="y"}) if {a="b", x="y"}`)
	f(`histogram_quantiles("q", 0.1, 0.9, {x="y"}) - {a="b"}`, `histogram_quantiles("q", 0.1, 0.9, {a="b", x="y"}) - {a="b", x="y"}`)
	f(`histogram_quantiles("q", 0.1, 0.9, sum(rate({x="y"}[5m])) by (le)) - {a="b"}`, `histogram_quantiles("q", 0.1, 0.9, sum(rate({x="y"}[5m])) by (le)) - {a="b"}`)
	f(`histogram_quantiles("q", 0.1, 0.9, sum(rate({x="y"}[5m])) by (le,x)) - {a="b"}`, `histogram_quantiles("q", 0.1, 0.9, sum(rate({x="y"}[5m])) by (le, x)) - {a="b", x="y"}`)
	f(`histogram_quantiles("q", 0.1, 0.9, sum(rate({x="y"}[5m])) by (le,x,a)) - {a="b"}`, `histogram_quantiles("q", 0.1, 0.9, sum(rate({a="b", x="y"}[5m])) by (le, x, a)) - {a="b", x="y"}`)

	// multilevel transform funcs
	f(`round(sqrt(foo)) + bar`, `round(sqrt(foo)) + bar`)
	f(`round(sqrt(foo)) + bar{b="a"}`, `round(sqrt(foo{b="a"})) + bar{b="a"}`)
	f(`round(sqrt(foo{a="b"})) + bar{x="y"}`, `round(sqrt(foo{a="b", x="y"})) + bar{a="b", x="y"}`)

	// rollup funcs
	f(`RATE(foo[5m]) / rate(baz{a="b"}) + increase(x{y="z"} offset 5i)`, `(RATE(foo{a="b", y="z"}[5m]) / rate(baz{a="b", y="z"})) + increase(x{a="b", y="z"} offset 5i)`)
	f(`sum(rate(foo[5m])) / rate(baz{a="b"})`, `sum(rate(foo[5m])) / rate(baz{a="b"})`)
	f(`sum(rate(foo[5m])) by (a) / rate(baz{a="b"})`, `sum(rate(foo{a="b"}[5m])) by (a) / rate(baz{a="b"})`)
	f(`rate({__name__="foo"}) + rate({__name__="bar",x="y"}) - rate({__name__=~"baz"})`, `(rate(foo{x="y"}) + rate(bar{x="y"})) - rate({__name__=~"baz", x="y"})`)
	f(`rate({__name__=~"foo|bar", x="y"}) + rate(baz)`, `rate({__name__=~"foo|bar", x="y"}) + rate(baz{x="y"})`)
	f(`absent_over_time(foo{x="y"}[5m]) + bar{a="b"}`, `absent_over_time(foo{x="y"}[5m]) + bar{a="b"}`)
	f(`{x="y"} + quantile_over_time(0.5, {a="b"})`, `{a="b", x="y"} + quantile_over_time(0.5, {a="b", x="y"})`)
	f(`quantiles_over_time("quantile", 0.1, 0.9, foo{x="y"}[5m] offset 4h) + bar{a!="b"}`, `quantiles_over_time("quantile", 0.1, 0.9, foo{a!="b", x="y"}[5m] offset 4h) + bar{a!="b", x="y"}`)

	// @ modifier
	f(`foo @ end() + bar{baz="a"}`, `foo{baz="a"} @ end() + bar{baz="a"}`)
	f(`sum(foo @ end()) + bar{baz="a"}`, `sum(foo @ end()) + bar{baz="a"}`)
	f(`foo @ (bar{a="b"} + baz{x="y"})`, `foo @ (bar{a="b", x="y"} + baz{a="b", x="y"})`)

	// subqueries
	f(`rate(avg_over_time(foo[5m:])) + bar{baz="a"}`, `rate(avg_over_time(foo{baz="a"}[5m:])) + bar{baz="a"}`)
	f(`rate(sum(foo[5m:])) + bar{baz="a"}`, `rate(sum(foo[5m:])) + bar{baz="a"}`)
	f(`rate(sum(foo[5m:]) by (baz)) + bar{baz="a"}`, `rate(sum(foo{baz="a"}[5m:]) by (baz)) + bar{baz="a"}`)

	// binary ops with constants or scalars
	f(`100 * foo / bar{baz="a"}`, `(100 * foo{baz="a"}) / bar{baz="a"}`)
	f(`foo * 100 / bar{baz="a"}`, `(foo{baz="a"} * 100) / bar{baz="a"}`)
	f(`foo / bar{baz="a"} * 100`, `(foo{baz="a"} / bar{baz="a"}) * 100`)
	f(`scalar(x) * foo / bar{baz="a"}`, `(scalar(x) * foo{baz="a"}) / bar{baz="a"}`)
	f(`SCALAR(x) * foo / bar{baz="a"}`, `(SCALAR(x) * foo{baz="a"}) / bar{baz="a"}`)
	f(`100 * on(foo) bar{baz="z"} + a`, `(100 * on (foo) bar{baz="z"}) + a`)
}
