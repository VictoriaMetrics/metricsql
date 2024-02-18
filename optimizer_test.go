package metricsql

import (
	"testing"
)

func TestPushdownBinaryOpFilters(t *testing.T) {
	f := func(q, filters, resultExpected string) {
		t.Helper()
		e, err := Parse(q)
		if err != nil {
			t.Fatalf("unexpected error in Parse(%s): %s", q, err)
		}
		sOrig := string(e.AppendString(nil))
		filtersExpr, err := Parse(filters)
		if err != nil {
			t.Fatalf("cannot parse filters %s: %s", filters, err)
		}
		me, ok := filtersExpr.(*MetricExpr)
		if !ok {
			t.Fatalf("filters=%s must be a metrics expression; got %T", filters, filtersExpr)
		}
		if len(me.LabelFilterss) > 1 {
			t.Fatalf("filters=%s mustn't contain 'or'", filters)
		}
		var lfs []LabelFilter
		if len(me.LabelFilterss) == 1 {
			lfs = me.LabelFilterss[0]
		}
		resultExpr := PushdownBinaryOpFilters(e, lfs)
		result := resultExpr.AppendString(nil)
		if string(result) != resultExpected {
			t.Fatalf("unexpected result for PushdownBinaryOpFilters(%s, %s);\ngot\n%s\nwant\n%s", q, filters, result, resultExpected)
		}
		// Verify that the original e didn't change after PushdownBinaryOpFilters() call
		s := string(e.AppendString(nil))
		if s != sOrig {
			t.Fatalf("the original expression has been changed;\ngot\n%s\nwant\n%s", s, sOrig)
		}
	}
	f(`foo`, `{}`, `foo`)
	f(`foo`, `{a="b"}`, `foo{a="b"}`)
	f(`foo + bar{x="y"}`, `{c="d",a="b"}`, `foo{a="b",c="d"} + bar{a="b",c="d",x="y"}`)
	f(`sum(x)`, `{a="b"}`, `sum(x)`)
	f(`foo or bar`, `{a="b"}`, `foo{a="b"} or bar{a="b"}`)
	f(`foo or on(x) bar`, `{a="b"}`, `foo or on(x) bar`)
	f(`foo == on(x) group_LEft bar`, `{a="b"}`, `foo == on(x) group_left() bar`)
	f(`foo{x="y"} > ignoRIng(x) group_left(abc) bar`, `{a="b"}`, `foo{a="b",x="y"} > ignoring(x) group_left(abc) bar{a="b"}`)
	f(`foo{x="y"} >bool ignoring(x) group_right(abc,def) bar`, `{a="b"}`, `foo{a="b",x="y"} >bool ignoring(x) group_right(abc,def) bar{a="b"}`)
	f(`foo * ignoring(x) bar`, `{a="b"}`, `foo{a="b"} * ignoring(x) bar{a="b"}`)
	f(`foo{f1!~"x"} UNLEss bar{f2=~"y.+"}`, `{a="b",x=~"y"}`, `foo{a="b",f1!~"x",x=~"y"} unless bar{a="b",f2=~"y.+",x=~"y"}`)
	f(`a / sum(x)`, `{a="b",c=~"foo|bar"}`, `a{a="b",c=~"foo|bar"} / sum(x)`)
	f(`round(rate(x[5m] offset -1h)) + 123 / {a="b"}`, `{x!="y"}`, `round(rate(x{x!="y"}[5m] offset -1h)) + (123 / {a="b",x!="y"})`)
	f(`scalar(foo)+bar`, `{a="b"}`, `scalar(foo) + bar{a="b"}`)
	f(`vector(foo)`, `{a="b"}`, `vector(foo)`)
	f(`{a="b"} + on() group_left() {c="d"}`, `{a="b"}`, `{a="b"} + on() group_left() {c="d"}`)

	// pushdown for 'or' filters
	f(`foo{a="b" or c="d" or x="y",q="w"}`, `{x="y"}`, `foo{a="b",x="y" or c="d",x="y" or q="w",x="y"}`)
	f(`{a="b" or x="y",q="w"} + bar`, `{x="y"}`, `{a="b",x="y" or q="w",x="y"} + bar{x="y"}`)

	// pushdown for label_set
	f(`label_set(foo, "a", "b") + bar{baz="a"}`, `{x="y"}`, `label_set(foo{x="y"}, "a", "b") + bar{baz="a",x="y"}`)
	f(`label_set(foo, "a", "b", "x", "aa") + bar{baz="a"}`, `{x="y"}`, `label_set(foo, "a", "b", "x", "aa") + bar{baz="a",x="y"}`)
	f(`label_set(label_set(foo, "a", "b"), "c", "d") + bar`, `{x="y"}`, `label_set(label_set(foo{x="y"}, "a", "b"), "c", "d") + bar{x="y"}`)
}

func TestGetCommonLabelFilters(t *testing.T) {
	f := func(q, resultExpected string) {
		t.Helper()
		e, err := Parse(q)
		if err != nil {
			t.Fatalf("unexpected error in Parse(%s): %s", q, err)
		}
		lfs := getCommonLabelFilters(e)
		var me MetricExpr
		if len(lfs) > 0 {
			me.LabelFilterss = [][]LabelFilter{lfs}
		}
		result := me.AppendString(nil)
		if string(result) != resultExpected {
			t.Fatalf("unexpected result for getCommonLabelFilters(%s);\ngot\n%s\nwant\n%s", q, result, resultExpected)
		}
	}
	f(`{}`, `{}`)
	f(`foo`, `{}`)
	f(`{__name__="foo"}`, `{}`)
	f(`{__name__=~"bar"}`, `{}`)
	f(`{__name__=~"a|b",x="y"}`, `{x="y"}`)
	f(`foo{c!="d",a="b"}`, `{c!="d",a="b"}`)
	f(`1+foo`, `{}`)
	f(`foo + bar{a="b"}`, `{a="b"}`)
	f(`foo + bar / baz{a="b"}`, `{a="b"}`)
	f(`foo{x!="y"} + bar / baz{a="b"}`, `{x!="y",a="b"}`)
	f(`foo{x!="y"} + bar{x=~"a|b",q!~"we|rt"} / baz{a="b"}`, `{x!="y",x=~"a|b",q!~"we|rt",a="b"}`)
	f(`{a="b"} + on() {c="d"}`, `{}`)
	f(`{a="b"} + on() group_left() {c="d"}`, `{a="b"}`)
	f(`{a="b"} + on(a) group_left() {c="d"}`, `{a="b"}`)
	f(`{a="b"} + on(c) group_left() {c="d"}`, `{a="b",c="d"}`)
	f(`{a="b"} + on(a,c) group_left() {c="d"}`, `{a="b",c="d"}`)
	f(`{a="b"} + on(d) group_left() {c="d"}`, `{a="b"}`)
	f(`{a="b"} + on() group_right(s) {c="d"}`, `{c="d"}`)
	f(`{a="b"} + On(a) groUp_right() {c="d"}`, `{a="b",c="d"}`)
	f(`{a="b"} + on(c) group_right() {c="d"}`, `{c="d"}`)
	f(`{a="b"} + on(a,c) group_right() {c="d"}`, `{a="b",c="d"}`)
	f(`{a="b"} + on(d) group_right() {c="d"}`, `{c="d"}`)
	f(`{a="b"} or {c="d"}`, `{}`)
	f(`{a="b",x="y"} or {x="y",c="d"}`, `{x="y"}`)
	f(`{a="b",x="y"} Or on() {x="y",c="d"}`, `{}`)
	f(`{a="b",x="y"} Or on(a) {x="y",c="d"}`, `{}`)
	f(`{a="b",x="y"} Or on(x) {x="y",c="d"}`, `{x="y"}`)
	f(`{a="b",x="y"} Or oN(x,y) {x="y",c="d"}`, `{x="y"}`)
	f(`{a="b",x="y"} Or on(y) {x="y",c="d"}`, `{}`)
	f(`(foo{a="b"} + bar{c="d"}) or (baz{x="y"} <= x{a="b"})`, `{a="b"}`)
	f(`{a="b"} unless {c="d"}`, `{a="b"}`)
	f(`{a="b"} unless on() {c="d"}`, `{}`)
	f(`{a="b"} unLess on(a) {c="d"}`, `{a="b"}`)
	f(`{a="b"} unLEss on(c) {c="d"}`, `{}`)
	f(`{a="b"} unless on(a,c) {c="d"}`, `{a="b"}`)
	f(`{a="b"} Unless on(x) {c="d"}`, `{}`)

	// common filters for 'or' filters
	f(`{a="b" or c="d",a="b"}`, `{a="b"}`)
	f(`{a="b",c="d" or c="d",a="b"}`, `{c="d",a="b"}`)
	f(`foo{x="y",a="b",c="d" or c="d",a="b"}`, `{c="d",a="b"}`)
}

func TestOptimize(t *testing.T) {
	f := func(q, qOptimizedExpected string) {
		t.Helper()
		e, err := Parse(q)
		if err != nil {
			t.Fatalf("unexpected error in Parse(%s): %s", q, err)
		}
		sOrig := string(e.AppendString(nil))
		eOptimized := Optimize(e)
		qOptimized := eOptimized.AppendString(nil)
		if string(qOptimized) != qOptimizedExpected {
			t.Fatalf("unexpected qOptimized;\ngot\n%s\nwant\n%s", qOptimized, qOptimizedExpected)
		}
		// Make sure the original e didn't change after Optimize() call
		s := string(e.AppendString(nil))
		if s != sOrig {
			t.Fatalf("the original expression has been changed;\ngot\n%s\nwant\n%s", s, sOrig)
		}
	}
	f("foo", "foo")

	// reserved words. See https://github.com/VictoriaMetrics/VictoriaMetrics/issues/4422
	f(`1 + (on)`, `1 + (on)`)
	f(`{a="b"} + (group_left)`, `{a="b"} + (group_left{a="b"})`)
	f(`bool{a="b"} + (ignoring{c="d"})`, `bool{a="b",c="d"} + (ignoring{a="b",c="d"})`)

	// common binary expressions
	f("a + b", "a + b")
	f(`foo{label1="value1"} == bar`, `foo{label1="value1"} == bar{label1="value1"}`)
	f(`foo{label1="value1"} == bar{label2="value2"}`, `foo{label1="value1",label2="value2"} == bar{label1="value1",label2="value2"}`)
	f(`foo + bar{b=~"a.*", a!="ss"}`, `foo{a!="ss",b=~"a.*"} + bar{a!="ss",b=~"a.*"}`)
	f(`foo{bar="1"} / 234`, `foo{bar="1"} / 234`)
	f(`foo{bar="1"} / foo{bar="1"}`, `foo{bar="1"} / foo{bar="1"}`)
	f(`123 + foo{bar!~"xx"}`, `123 + foo{bar!~"xx"}`)
	f(`foo or bar{x="y"}`, `foo or bar{x="y"}`)
	f(`foo{x="y"} * on() baz{a="b"}`, `foo{x="y"} * on() baz{a="b"}`)
	f(`foo{x="y"} * on(a) baz{a="b"}`, `foo{a="b",x="y"} * on(a) baz{a="b"}`)
	f(`foo{x="y"} * on(bar) baz{a="b"}`, `foo{x="y"} * on(bar) baz{a="b"}`)
	f(`foo{x="y"} * on(x,a,bar) baz{a="b"}`, `foo{a="b",x="y"} * on(x,a,bar) baz{a="b",x="y"}`)
	f(`foo{x="y"} * ignoring() baz{a="b"}`, `foo{a="b",x="y"} * ignoring() baz{a="b",x="y"}`)
	f(`foo{x="y"} * ignoring(a) baz{a="b"}`, `foo{x="y"} * ignoring(a) baz{a="b",x="y"}`)
	f(`foo{x="y"} * ignoring(bar) baz{a="b"}`, `foo{a="b",x="y"} * ignoring(bar) baz{a="b",x="y"}`)
	f(`foo{x="y"} * ignoring(x,a,bar) baz{a="b"}`, `foo{x="y"} * ignoring(x,a,bar) baz{a="b"}`)
	f(`foo{x="y"} * ignoring() group_left(foo,bar) baz{a="b"}`, `foo{a="b",x="y"} * ignoring() group_left(foo,bar) baz{a="b",x="y"}`)
	f(`foo{x="y"} * on(a) group_left baz{a="b"}`, `foo{a="b",x="y"} * on(a) group_left() baz{a="b"}`)
	f(`foo{x="y"} * on(a) group_right(x, y) baz{a="b"}`, `foo{a="b",x="y"} * on(a) group_right(x,y) baz{a="b"}`)
	f(`histogram_quantile(foo, bar{baz=~"sdf"} + aa{baz=~"axx", aa="b"})`, `histogram_quantile(foo, bar{aa="b",baz=~"axx",baz=~"sdf"} + aa{aa="b",baz=~"axx",baz=~"sdf"})`)
	f(`sum(foo, bar{baz=~"sdf"} + aa{baz=~"axx", aa="b"})`, `sum(foo, bar{aa="b",baz=~"axx",baz=~"sdf"} + aa{aa="b",baz=~"axx",baz=~"sdf"})`)
	f(`foo AND bar{baz="aa"}`, `foo{baz="aa"} and bar{baz="aa"}`)
	f(`{x="y",__name__="a"} + {a="b"}`, `a{a="b",x="y"} + {a="b",x="y"}`)
	f(`{x="y",__name__=~"a|b"} + {a="b"}`, `{__name__=~"a|b",a="b",x="y"} + {a="b",x="y"}`)
	f(`a{x="y",__name__=~"a|b"} + {a="b"}`, `a{__name__=~"a|b",a="b",x="y"} + {a="b",x="y"}`)
	f(`{a="b"} + ({c="d"} * on() group_left() {e="f"})`, `{a="b",c="d"} + ({c="d"} * on() group_left() {e="f"})`)
	f(`{a="b"} + ({c="d"} * on(a) group_left() {e="f"})`, `{a="b",c="d"} + ({a="b",c="d"} * on(a) group_left() {a="b",e="f"})`)
	f(`{a="b"} + ({c="d"} * on(c) group_left() {e="f"})`, `{a="b",c="d"} + ({c="d"} * on(c) group_left() {c="d",e="f"})`)
	f(`{a="b"} + ({c="d"} * on(e) group_left() {e="f"})`, `{a="b",c="d",e="f"} + ({c="d",e="f"} * on(e) group_left() {e="f"})`)
	f(`{a="b"} + ({c="d"} * on(x) group_left() {e="f"})`, `{a="b",c="d"} + ({c="d"} * on(x) group_left() {e="f"})`)
	f(`{a="b"} + ({c="d"} * on() group_right() {e="f"})`, `{a="b",e="f"} + ({c="d"} * on() group_right() {e="f"})`)
	f(`{a="b"} + ({c="d"} * on(a) group_right() {e="f"})`, `{a="b",e="f"} + ({a="b",c="d"} * on(a) group_right() {a="b",e="f"})`)
	f(`{a="b"} + ({c="d"} * on(c) group_right() {e="f"})`, `{a="b",c="d",e="f"} + ({c="d"} * on(c) group_right() {c="d",e="f"})`)
	f(`{a="b"} + ({c="d"} * on(e) group_right() {e="f"})`, `{a="b",e="f"} + ({c="d",e="f"} * on(e) group_right() {e="f"})`)
	f(`{a="b"} + ({c="d"} * on(x) group_right() {e="f"})`, `{a="b",e="f"} + ({c="d"} * on(x) group_right() {e="f"})`)
	f(`{a="b" or c="d"} + ({c="d"} * on(x) group_right() {e="f"})`, `{a="b",e="f" or c="d",e="f"} + ({c="d"} * on(x) group_right() {e="f"})`)
	f(`a + on(x) group_left(*) (prefix{x="a"})`, `a{x="a"} + on(x) group_left(*) (prefix{x="a"})`)
	f(`a + on(x) group_right(*) prefix "foo_" b{x="a"}`, `a{x="a"} + on(x) group_right(*) prefix "foo_" b{x="a"}`)
	f(`a{x="a"} + on(x) group_right(*) prefix "foo_" prefix`, `a{x="a"} + on(x) group_right(*) prefix "foo_" (prefix{x="a"})`)

	// specially handled binary expressions
	f(`foo{a="b"} or bar{x="y"}`, `foo{a="b"} or bar{x="y"}`)
	f(`(foo{a="b"} + bar{c="d"}) or (baz{x="y"} <= x{a="b"})`, `(foo{a="b",c="d"} + bar{a="b",c="d"}) or (baz{a="b",x="y"} <= x{a="b",x="y"})`)
	f(`(foo{a="b"} + bar{c="d"}) or on(x) (baz{x="y"} <= x{a="b"})`, `(foo{a="b",c="d"} + bar{a="b",c="d"}) or on(x) (baz{a="b",x="y"} <= x{a="b",x="y"})`)
	f(`foo + (bar or baz{a="b"})`, `foo + (bar or baz{a="b"})`)
	f(`foo + (bar{a="b"} or baz{a="b"})`, `foo{a="b"} + (bar{a="b"} or baz{a="b"})`)
	f(`foo + (bar{a="b",c="d"} or baz{a="b"})`, `foo{a="b"} + (bar{a="b",c="d"} or baz{a="b"})`)
	f(`foo{a="b"} + (bar OR baz{x="y"})`, `foo{a="b"} + (bar{a="b"} or baz{a="b",x="y"})`)
	f(`foo{a="b"} + (bar{x="y",z="456"} OR baz{x="y",z="123"})`, `foo{a="b",x="y"} + (bar{a="b",x="y",z="456"} or baz{a="b",x="y",z="123"})`)
	f(`foo{a="b"} unless bar{c="d"}`, `foo{a="b"} unless bar{a="b",c="d"}`)
	f(`foo{a="b"} unless on() bar{c="d"}`, `foo{a="b"} unless on() bar{c="d"}`)
	f(`foo + (bar{x="y"} unless baz{a="b"})`, `foo{x="y"} + (bar{x="y"} unless baz{a="b",x="y"})`)
	f(`foo + (bar{x="y"} unless on() baz{a="b"})`, `foo + (bar{x="y"} unless on() baz{a="b"})`)
	f(`foo{a="b"} + (bar UNLESS baz{x="y"})`, `foo{a="b"} + (bar{a="b"} unless baz{a="b",x="y"})`)
	f(`foo{a="b"} + (bar{x="y"} unLESS baz)`, `foo{a="b",x="y"} + (bar{a="b",x="y"} unless baz{a="b",x="y"})`)

	// aggregate funcs
	f(`sum(foo{bar="baz"}) / a{b="c"}`, `sum(foo{bar="baz"}) / a{b="c"}`)
	f(`sum(foo{bar="baz"}) by () / a{b="c"}`, `sum(foo{bar="baz"}) by() / a{b="c"}`)
	f(`sum(foo{bar="baz"}) by (bar) / a{b="c"}`, `sum(foo{bar="baz"}) by(bar) / a{b="c",bar="baz"}`)
	f(`sum(foo{bar="baz"}) by (b) / a{b="c"}`, `sum(foo{b="c",bar="baz"}) by(b) / a{b="c"}`)
	f(`sum(foo{bar="baz"}) by (x) / a{b="c"}`, `sum(foo{bar="baz"}) by(x) / a{b="c"}`)
	f(`sum(foo{bar="baz"}) by (bar,b) / a{b="c"}`, `sum(foo{b="c",bar="baz"}) by(bar,b) / a{b="c",bar="baz"}`)
	f(`sum(foo{bar="baz"}) without () / a{b="c"}`, `sum(foo{b="c",bar="baz"}) without() / a{b="c",bar="baz"}`)
	f(`sum(foo{bar="baz"}) without (bar) / a{b="c"}`, `sum(foo{b="c",bar="baz"}) without(bar) / a{b="c"}`)
	f(`sum(foo{bar="baz"}) without (b) / a{b="c"}`, `sum(foo{bar="baz"}) without(b) / a{b="c",bar="baz"}`)
	f(`sum(foo{bar="baz"}) without (x) / a{b="c"}`, `sum(foo{b="c",bar="baz"}) without(x) / a{b="c",bar="baz"}`)
	f(`sum(foo{bar="baz"}) without (bar,b) / a{b="c"}`, `sum(foo{bar="baz"}) without(bar,b) / a{b="c"}`)
	f(`sum(foo, bar) by (a) + baz{a="b"}`, `sum(foo{a="b"}, bar{a="b"}) by(a) + baz{a="b"}`)
	f(`topk(3, foo) by (baz,x) + bar{baz="a"}`, `topk(3, foo{baz="a"}) by(baz,x) + bar{baz="a"}`)
	f(`topk(a, foo) without (x,y) + bar{baz="a"}`, `topk(a, foo{baz="a"}) without(x,y) + bar{baz="a"}`)
	f(`a{b="c"} + quantiles("foo", 0.1, 0.2, bar{x="y"}) by (b, x, y)`, `a{b="c",x="y"} + quantiles("foo", 0.1, 0.2, bar{b="c",x="y"}) by(b,x,y)`)
	f(`count_values("foo", bar{baz="a"}) by (bar,b) + a{b="c"}`, `count_values("foo", bar{baz="a"}) by(bar,b) + a{b="c"}`)
	f(
		`sum(
				avg(foo{bar="one"}) by (bar),
				avg(foo{bar="two"}[1i]) by (bar)
			) by(bar)
			+ avg(foo{bar="three"}) by(bar)`,
		`sum(avg(foo{bar="one",bar="three"}) by(bar), avg(foo{bar="three",bar="two"}[1i]) by(bar)) by(bar) + avg(foo{bar="three"}) by(bar)`,
	)
	f(
		`sum(
				foo{bar="one"},
				avg(foo{bar="two"}[1i]) by (bar)
			) by(bar)
			+ avg(foo{bar="three"}) by(bar)`,
		`sum(foo{bar="one",bar="three"}, avg(foo{bar="three",bar="two"}[1i]) by(bar)) by(bar) + avg(foo{bar="three"}) by(bar)`,
	)
	f(`any(a{bar="x"}, b{bar="x",z="a"}) by (bar) + q{w="a"}`, `any(a{bar="x"}, b{bar="x",z="a"}) by(bar) + q{bar="x",w="a"}`)

	// transform funcs
	f(`round(foo{bar="baz"}) + sqrt(a{z=~"c"})`, `round(foo{bar="baz",z=~"c"}) + sqrt(a{bar="baz",z=~"c"})`)
	f(`foo{bar="baz"} + SQRT(a{z=~"c"})`, `foo{bar="baz",z=~"c"} + SQRT(a{bar="baz",z=~"c"})`)
	f(`round({__name__="foo"}) + bar`, `round(foo) + bar`)
	f(`round({__name__=~"foo|bar"}) + baz`, `round({__name__=~"foo|bar"}) + baz`)
	f(`round({__name__=~"foo|bar",a="b"}) + baz`, `round({__name__=~"foo|bar",a="b"}) + baz{a="b"}`)
	f(`round({__name__=~"foo|bar",a="b"}) + sqrt(baz)`, `round({__name__=~"foo|bar",a="b"}) + sqrt(baz{a="b"})`)
	f(`round(foo) + {__name__="bar",x="y"}`, `round(foo{x="y"}) + bar{x="y"}`)
	f(`absent(foo{bar="baz"}) + sqrt(a{z=~"c"})`, `absent(foo{bar="baz"}) + sqrt(a{z=~"c"})`)
	f(`ABSENT(foo{bar="baz"}) + sqrt(a{z=~"c"})`, `ABSENT(foo{bar="baz"}) + sqrt(a{z=~"c"})`)
	f(`now() + foo{bar="baz"} + x{y="x"}`, `(now() + foo{bar="baz",y="x"}) + x{bar="baz",y="x"}`)
	f(`limit_offset(5, 10, {x="y"}) if {a="b"}`, `limit_offset(5, 10, {a="b",x="y"}) if {a="b",x="y"}`)
	f(`buckets_limit(aa, {x="y"}) if {a="b"}`, `buckets_limit(aa, {a="b",x="y"}) if {a="b",x="y"}`)
	f(`histogram_quantiles("q", 0.1, 0.9, {x="y"}) - {a="b"}`, `histogram_quantiles("q", 0.1, 0.9, {a="b",x="y"}) - {a="b",x="y"}`)
	f(`histogram_quantiles("q", 0.1, 0.9, sum(rate({x="y"}[5m])) by (le)) - {a="b"}`, `histogram_quantiles("q", 0.1, 0.9, sum(rate({x="y"}[5m])) by(le)) - {a="b"}`)
	f(`histogram_quantiles("q", 0.1, 0.9, sum(rate({x="y"}[5m])) by (le,x)) - {a="b"}`, `histogram_quantiles("q", 0.1, 0.9, sum(rate({x="y"}[5m])) by(le,x)) - {a="b",x="y"}`)
	f(`histogram_quantiles("q", 0.1, 0.9, sum(rate({x="y"}[5m])) by (le,x,a)) - {a="b"}`, `histogram_quantiles("q", 0.1, 0.9, sum(rate({a="b",x="y"}[5m])) by(le,x,a)) - {a="b",x="y"}`)
	f(`vector(foo) + bar{a="b"}`, `vector(foo) + bar{a="b"}`)
	f(`vector(foo{x="y"} + a) + bar{a="b"}`, `vector(foo{x="y"} + a{x="y"}) + bar{a="b"}`)

	// Label manipulation functions, which are in reality do not change labels for the input series
	f(`labels_equal(foo{x="y"}, "a", "b") + label_match(bar{q="w"}, "foo", "bar")`, `labels_equal(foo{q="w",x="y"}, "a", "b") + label_match(bar{q="w",x="y"}, "foo", "bar")`)

	// label_set
	f(`label_set(foo, "__name__", "bar") + x`, `label_set(foo, "__name__", "bar") + x`)
	f(`label_set(foo, "a", "bar") + x{__name__="y"}`, `label_set(foo, "a", "bar") + x{__name__="y",a="bar"}`)
	f(`label_set(foo{bar="baz"}, "xx", "y") + a{x="y"}`, `label_set(foo{bar="baz",x="y"}, "xx", "y") + a{bar="baz",x="y",xx="y"}`)
	f(`label_set(foo{x="y"}, "q", "b", "x", "qwe") + label_set(bar{q="w"}, "x", "a", "q", "w")`, `label_set(foo{x="y"}, "q", "b", "x", "qwe") + label_set(bar{q="w"}, "x", "a", "q", "w")`)
	f(`label_set(foo{a="b"}, "a", "qwe") + bar{a="x"}`, `label_set(foo{a="b"}, "a", "qwe") + bar{a="qwe",a="x"}`)

	// alias
	f(`alias(foo, "bar") + abc`, `label_set(foo, "__name__", "bar") + abc`)
	f(`alias(foo, "bar") + abc{d="e"}`, `label_set(foo{d="e"}, "__name__", "bar") + abc{d="e"}`)
	f(`alias(foo{x="y"}, "bar") + abc{d="e"}`, `label_set(foo{d="e",x="y"}, "__name__", "bar") + abc{d="e",x="y"}`)

	// label_replace
	f(`label_replace(foo, "a", "b", "c", "d") + bar{x="y"}`, `label_replace(foo{x="y"}, "a", "b", "c", "d") + bar{x="y"}`)
	f(`label_replace(foo, "a", "b", "c", "d") + bar{a="y"}`, `label_replace(foo, "a", "b", "c", "d") + bar{a="y"}`)
	f(`label_replace(foo{x="qwe"}, "a", "b", "c", "d") + bar{a="y"}`, `label_replace(foo{x="qwe"}, "a", "b", "c", "d") + bar{a="y",x="qwe"}`)
	f(`label_replace(foo{x="qwe"}, "a", "b", "c", "d") + bar{x="y"}`, `label_replace(foo{x="qwe",x="y"}, "a", "b", "c", "d") + bar{x="qwe",x="y"}`)
	f(`label_replace(foo{aa!="qwe"}, "a", "b", "c", "d") + bar{x="y"}`, `label_replace(foo{aa!="qwe",x="y"}, "a", "b", "c", "d") + bar{aa!="qwe",x="y"}`)

	// label_join
	f(`label_join(foo, "a", "b", "c") + bar{x="y"}`, `label_join(foo{x="y"}, "a", "b", "c") + bar{x="y"}`)
	f(`label_join(foo, "a", "b", "c") + bar{a="y"}`, `label_join(foo, "a", "b", "c") + bar{a="y"}`)
	f(`label_join(foo{a="qwe"}, "a", "b", "c") + bar{x="y"}`, `label_join(foo{a="qwe",x="y"}, "a", "b", "c") + bar{x="y"}`)
	f(`label_join(foo{q="z"}, "a", "b", "c") + bar{a="y"}`, `label_join(foo{q="z"}, "a", "b", "c") + bar{a="y",q="z"}`)
	f(`label_join(foo{q="z"}, "a", "b", "c") + bar{w="y"}`, `label_join(foo{q="z",w="y"}, "a", "b", "c") + bar{q="z",w="y"}`)

	// label_copy
	f(`label_copy(foo, "a", "b") + bar{x="y"}`, `label_copy(foo{x="y"}, "a", "b") + bar{x="y"}`)
	f(`label_copy(foo, "a", "b", "c", "d") + bar{a="y",b="z"}`, `label_copy(foo{a="y"}, "a", "b", "c", "d") + bar{a="y",b="z"}`)
	f(`label_copy(foo{q="w"}, "a", "b") + bar{a="y",b="z"}`, `label_copy(foo{a="y",q="w"}, "a", "b") + bar{a="y",b="z",q="w"}`)
	f(`label_copy(foo{b="w"}, "a", "b") + bar{a="y",b="z"}`, `label_copy(foo{a="y",b="w"}, "a", "b") + bar{a="y",b="z"}`)

	// label_del
	f(`label_del(foo, "a", "b") + bar{x="y"}`, `label_del(foo{x="y"}, "a", "b") + bar{x="y"}`)
	f(`label_del(foo{a="q",b="w",z="d"}, "a", "b") + bar{a="y",b="z",x="y"}`, `label_del(foo{a="q",b="w",x="y",z="d"}, "a", "b") + bar{a="y",b="z",x="y",z="d"}`)

	// multilevel transform funcs
	f(`round(sqrt(foo)) + bar`, `round(sqrt(foo)) + bar`)
	f(`round(sqrt(foo)) + bar{b="a"}`, `round(sqrt(foo{b="a"})) + bar{b="a"}`)
	f(`round(sqrt(foo{a="b"})) + bar{x="y"}`, `round(sqrt(foo{a="b",x="y"})) + bar{a="b",x="y"}`)

	// rollup funcs
	f(`RATE(foo[5m]) / rate(baz{a="b"}) + increase(x{y="z"} offset 5i)`, `(RATE(foo{a="b",y="z"}[5m]) / rate(baz{a="b",y="z"})) + increase(x{a="b",y="z"} offset 5i)`)
	f(`sum(rate(foo[5m])) / rate(baz{a="b"})`, `sum(rate(foo[5m])) / rate(baz{a="b"})`)
	f(`sum(rate(foo[5m])) by (a) / rate(baz{a="b"})`, `sum(rate(foo{a="b"}[5m])) by(a) / rate(baz{a="b"})`)
	f(`rate({__name__="foo"}) + rate({__name__="bar",x="y"}) - rate({__name__=~"baz"})`, `(rate(foo{x="y"}) + rate(bar{x="y"})) - rate({__name__=~"baz",x="y"})`)
	f(`rate({__name__=~"foo|bar", x="y"}) + rate(baz)`, `rate({__name__=~"foo|bar",x="y"}) + rate(baz{x="y"})`)
	f(`absent_over_time(foo{x="y"}[5m]) + bar{a="b"}`, `absent_over_time(foo{x="y"}[5m]) + bar{a="b"}`)
	f(`{x="y"} + quantile_over_time(0.5, {a="b"})`, `{a="b",x="y"} + quantile_over_time(0.5, {a="b",x="y"})`)
	f(`quantiles_over_time("quantile", 0.1, 0.9, foo{x="y"}[5m] offset 4h) + bar{a!="b"}`, `quantiles_over_time("quantile", 0.1, 0.9, foo{a!="b",x="y"}[5m] offset 4h) + bar{a!="b",x="y"}`)

	// @ modifier
	f(`foo @ end() + bar{baz="a"}`, `(foo{baz="a"} @ end()) + bar{baz="a"}`)
	f(`sum(foo @ end()) + bar{baz="a"}`, `sum(foo @ end()) + bar{baz="a"}`)
	f(`foo @ (bar{a="b"} + baz{x="y"})`, `foo @ (bar{a="b",x="y"} + baz{a="b",x="y"})`)

	// subqueries
	f(`rate(avg_over_time(foo[5m:])) + bar{baz="a"}`, `rate(avg_over_time(foo{baz="a"}[5m:])) + bar{baz="a"}`)
	f(`rate(sum(foo[5m:])) + bar{baz="a"}`, `rate(sum(foo[5m:])) + bar{baz="a"}`)
	f(`rate(sum(foo[5m:]) by (baz)) + bar{baz="a"}`, `rate(sum(foo{baz="a"}[5m:]) by(baz)) + bar{baz="a"}`)

	// binary ops with constants or scalars
	f(`100 * foo / bar{baz="a"}`, `(100 * foo{baz="a"}) / bar{baz="a"}`)
	f(`foo * 100 / bar{baz="a"}`, `(foo{baz="a"} * 100) / bar{baz="a"}`)
	f(`foo / bar{baz="a"} * 100`, `(foo{baz="a"} / bar{baz="a"}) * 100`)
	f(`scalar(x) * foo / bar{baz="a"}`, `(scalar(x) * foo{baz="a"}) / bar{baz="a"}`)
	f(`SCALAR(x) * foo / bar{baz="a"}`, `(SCALAR(x) * foo{baz="a"}) / bar{baz="a"}`)
	f(`100 * on(foo) bar{baz="z"} + a`, `(100 * on(foo) bar{baz="z"}) + a`)
}
