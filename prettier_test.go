package metricsql

import "testing"

func TestPrettier(t *testing.T) {
	another := func(s, expected string) {
		t.Helper()

		const maxLineLength = 130
		got, err := Prettier(s, maxLineLength)
		if err != nil {
			t.Fatalf("unexpected error when parsing %q: %s", s, err)
		}

		if expected != got {
			t.Fatalf("string not prettified;\ngot:%s\nwant\n%q", got, expected)
		}
	}
	same := func(s string) {
		t.Helper()
		another(s, s)
	}

	same(`{}`)
	same(`{}[5m]`)
	same(`{}[5m:]`)
	same(`{}[5M:]`)
	same(`{}[:]`)
	another(`{}[: ]`, `{}[:]`)
	same(`{}[:3s]`)
	another(`{}[: 3s ]`, `{}[:3s]`)
	same(`{}[5m:3s]`)
	another(`{}[ 5m : 3s ]`, `{}[5m:3s]`)
	same(`{} offset 5m`)
	same(`{} offset -5m`)
	same(`{} offset 5M`)
	same(`{}[5m] offset 10y`)
	same(`{}[5.3m:3.4s] offset 10y`)
	same(`{}[:3.4s] offset 10y`)
	same(`{}[:3.4s] offset -10y`)
	same(`{Foo="bAR"}`)
	same(`{foo="bar"}`)
	same(`{foo="bar"}[5m]`)
	same(`{foo="bar"}[5m:]`)
	same(`{foo="bar"}[5m:3s]`)
	same(`{foo="bar"} offset 13.4ms`)
	same(`{foo="bar"}[5w4h-3.4m13.4ms]`)
	same(`{foo="bar"} offset 10y`)
	same(`{foo="bar"} offset -10y`)
	same(`{foo="bar"}[5m] offset 10y`)
	same(`{foo="bar"}[5m:3s] offset 10y`)
	another(`{foo="bar"}[5m] oFFSEt 10y`, `{foo="bar"}[5m] offset 10y`)
	same("METRIC")
	same("metric")
	same("m_e:tri44:_c123")
	another("-metric", "0 - metric")
	same(`metric offset 10h`)
	same("metric[5m]")
	same("metric[5m:3s]")
	same("metric[5m] offset 10h")
	same("metric[5m:3s] offset 10h")
	same("metric[5i:3i] offset 10i")
	same(`metric{foo="bar"}`)
	same(`metric{foo="bar"} offset 10h`)
	same(`metric{foo!="bar"}[2d]`)
	same(`metric{foo="bar"}[2d] offset 10h`)
	same(`metric{foo="bar", b="sdfsdf"}[2d:3h] offset 10h`)
	same(`metric{foo="bar", b="sdfsdf"}[2d:3h] offset 10`)
	same(`metric{foo="bar", b="sdfsdf"}[2d:3] offset 10h`)
	same(`metric{foo="bar", b="sdfsdf"}[2:3h] offset 10h`)
	same(`metric{foo="bar", b="sdfsdf"}[2.34:5.6] offset 3600.5`)
	same(`metric{foo="bar", b="sdfsdf"}[234:56] offset -3600`)
	another(`  metric  {  foo  = "bar"  }  [  2d ]   offset   10h  `, `metric{foo="bar"}[2d] offset 10h`)
	// @ modifier
	// See https://prometheus.io/docs/prometheus/latest/querying/basics/#modifier
	same(`foo @ 123.45`)
	same(`foo\@ @ 123.45`)
	same(`{foo=~"bar"} @ end()`)
	same(`foo{bar="baz"} @ start()`)
	same(`foo{bar="baz"}[5m] @ 12345`)
	same(`foo{bar="baz"}[5m:4s] offset 5m @ (end() - 3.5m)`)
	another(`foo{bar="baz"}[5m:4s] @ (end() - 3.5m) offset 2.4h`, `foo{bar="baz"}[5m:4s] offset 2.4h @ (end() - 3.5m)`)
	another(`foo @ start() + (bar offset 3m @ end()) / baz OFFSET -5m`, `foo @ start() + (bar offset 3m @ end() / baz offset -5m)`)
	same(`sum(foo) @ start() + rate(bar @ (end() - 5m))`)
	another(`time() @ (start())`, `time() @ start()`)
	another(`time() @ (start()+(1+1))`, `time() @ (start() + 2)`)
	same(`time() @ (end() - 10m)`)
	// metric name matching keywords
	same("rate")
	same("RATE")
	same("by")
	same("BY")
	same("bool")
	same("BOOL")
	same("unless")
	same("UNLESS")
	same("Ignoring")
	same("with")
	same("WITH")
	same("With")
	same("offset")
	same("keep_metric_names")
	same("alias")
	same(`alias{foo="bar"}`)
	same(`aLIas{alias="aa"}`)
	another(`al\ias`, `alias`)
	// identifiers with with escape chars
	same(`foo\ bar`)
	same(`foo\-bar\{{baz\+bar="aa"}`)
	another(`\x2E\x2ef\oo{b\xEF\ar="aa"}`, `\..foo{bïar="aa"}`)
	same(`温度{房间="水电费"}[5m] offset 10m`)
	another(`\温\度{\房\间="水电费"}[5m] offset 10m`, `温度{房间="水电费"}[5m] offset 10m`)
	same(`sum(fo\|o) by (b\|a, x)`)
	another(`sum(x) by (b\x7Ca)`, `sum(x) by (b\|a)`)
	// Duplicate filters
	same(`foo{__name__="bar"}`)
	same(`foo{a="b", a="c", __name__="aaa", b="d"}`)
	// Metric filters ending with comma
	another(`m{foo="bar",}`, `m{foo="bar"}`)
	// String concat in tag value
	another(`m{foo="bar" + "baz"}`, `m{foo="barbaz"}`)

	// Valid regexp
	same(`foo{bar=~"x"}`)
	same(`foo{bar=~"^x"}`)
	same(`foo{bar=~"^x$"}`)
	same(`foo{bar=~"^(a[bc]|d)$"}`)
	same(`foo{bar!~"x"}`)
	same(`foo{bar!~"^x"}`)
	same(`foo{bar!~"^x$"}`)
	same(`foo{bar!~"^(a[bc]|d)$"}`)

	// stringExpr
	same(`""`)
	same(`"\n\t\r 12:{}[]()44"`)
	another(`''`, `""`)
	another("``", `""`)
	another("   `foo\"b'ar`  ", "\"foo\\\"b'ar\"")
	another(`  'foo\'bar"BAZ'  `, `"foo'bar\"BAZ"`)
	// string concat
	another(`"foo"+'bar'`, `"foobar"`)

	// numberExpr
	same(`1`)
	same(`123.`)
	another(`-123.`, `-123`)
	same(`foo - 123.`)
	same(`12.e+4`)
	same(`12Ki`)
	same(`1.23Gb`)
	same(`foo - 23M`)
	another(`-1.23Gb`, `-1.23e+09`)
	same(`1.23`)
	same(`0.23`)
	same(`1.2e+45`)
	same(`1.2e-45`)
	same(`-1`)
	same(`-1.23`)
	same(`-0.23`)
	same(`-1.2e+45`)
	same(`-1.2e-45`)
	same(`-1.2e-45`)
	same(`12.5E34`)
	another(`-.2`, `-0.2`)
	another(`-.2E-2`, `-0.002`)
	same(`NaN`)
	same(`nan`)
	same(`NAN`)
	same(`nAN`)
	same(`Inf`)
	same(`INF`)
	same(`inf`)
	another(`+Inf`, `Inf`)
	same(`-Inf`)
	another(`-inF`, `-Inf`)
	same(`0x12`)
	same(`0x3b`)
	another(`-0x3b`, `-59`)
	another(`+0X3B`, `0X3B`)
	same(`0b1011`)
	same(`073`)
	another(`-0o12`, `-10`)

	// durationExpr
	same(`1h`)
	another(`-1h`, `0 - 1h`)
	same(`0.34h4m5s`)
	same(`0.34H4m5S`)
	another(`-0.34h4m5s`, `0 - 0.34h4m5s`)
	same(`sum_over_tme(m[1h]) / 1h`)
	same(`sum_over_time(m[3600]) / 3600`)

	// binaryOpExpr
	another(`nan == nan`, `NaN`)
	another(`nan ==bool nan`, `1`)
	another(`nan !=bool nan`, `0`)
	another(`nan !=bool 2`, `1`)
	another(`2 !=bool nan`, `1`)
	another(`nan >bool nan`, `0`)
	another(`nan <bool nan`, `0`)
	another(`1 ==bool nan`, `0`)
	another(`NaN !=bool 1`, `1`)
	another(`inf >=bool 2`, `1`)
	another(`-1 >bool -inf`, `1`)
	another(`-1 <bool -inf`, `0`)
	another(`nan + 2 *3 * inf`, `NaN`)
	another(`INF - Inf`, `NaN`)
	another(`Inf + inf`, `+Inf`)
	another(`1/0`, `+Inf`)
	another(`0/0`, `NaN`)
	another(`-m`, `0 - m`)
	same(`m + ignoring () n[5m]`)
	another(`M + IGNORING () N[5m]`, `M + ignoring () N[5m]`)
	same(`m + on (foo) n[5m]`)
	another(`m + ON (Foo) n[5m]`, `m + on (Foo) n[5m]`)
	same(`m + ignoring (a, b) n[5m]`)
	another(`1 or 2`, `1`)
	another(`1 and 2`, `1`)
	another(`1 unless 2`, `NaN`)
	another(`1 default 2`, `1`)
	another(`1 default NaN`, `1`)
	another(`NaN default 2`, `2`)
	another(`1 > 2`, `NaN`)
	another(`1 > bool 2`, `0`)
	another(`3 >= 2`, `3`)
	another(`3 <= bool 2`, `0`)
	another(`1 + -2 - 3`, `-4`)
	another(`1 / 0 + 2`, `+Inf`)
	another(`2 + -1 / 0`, `-Inf`)
	another(`(-1) ^ 0.5`, `NaN`)
	another(`-1 ^ 0.5`, `-1`)
	another(`512.5 - (1 + 3) * (2 ^ 2) ^ 3`, `256.5`)
	another(`1 == bool 1 != bool 24 < bool 4 > bool -1`, `1`)
	another(`1 == bOOl 1 != BOOL 24 < Bool 4 > booL -1`, `1`)
	another(`m1+on(foo)group_left m2`, `m1 + on (foo) group_left () m2`)
	another(`M1+ON(FOO)GROUP_left M2`, `M1 + on (FOO) group_left () M2`)
	same(`m1 + on (foo) group_right () m2`)
	same(`m1 + on (foo, bar) group_right (x, y) m2`)
	another(`m1 + on (foo, bar,) group_right (x, y,) m2`, `m1 + on (foo, bar) group_right (x, y) m2`)
	same(`m1 == bool on (foo, bar) group_right (x, y) m2`)
	another(`5 - 1 + 3 * 2 ^ 2 ^ 3 - 2  OR Metric {Bar= "Baz", aaa!="bb",cc=~"dd" ,zz !~"ff" } `,
		`770 or Metric{Bar="Baz", aaa!="bb", cc=~"dd", zz!~"ff"}`)
	same(`"foo" + bar()`)
	same(`"foo" + bar{x="y"}`)
	same(`("foo"[3s] + bar{x="y"})[5m:3s] offset 10s`)
	same(`("foo"[3s] + bar{x="y"})[5i:3i] offset 10i`)
	same(`bar + "foo" offset 3s`)
	same(`bar + "foo" offset 3i`)
	another(`1+2 if 2>3`, `NaN`)
	another(`1+4 if 2<3`, `5`)
	another(`2+6 default 3 if 2>3`, `8`)
	another(`2+6 if 2>3 default NaN`, `NaN`)
	another(`42 if 3>2 if 2+2<5`, `42`)
	another(`42 if 3>2 if 2+2>=5`, `NaN`)
	another(`1+2 ifnot 2>3`, `3`)
	another(`1+4 ifnot 2<3`, `NaN`)
	another(`2+6 default 3 ifnot 2>3`, `8`)
	another(`2+6 ifnot 2>3 default NaN`, `8`)
	another(`42 if 3>2 ifnot 2+2<5`, `NaN`)
	another(`42 if 3>2 ifnot 2+2>=5`, `42`)
	another(`"foo" + "bar"`, `"foobar"`)
	another(`"foo"=="bar"`, `NaN`)
	another(`"foo"=="foo"`, `1`)
	another(`"foo"!="bar"`, `1`)
	another(`"foo"+"bar"+"baz"`, `"foobarbaz"`)
	another(`"a">"b"`, `NaN`)
	another(`"a">bool"b"`, `0`)
	another(`"a"<"b"`, `1`)
	another(`"a">="b"`, `NaN`)
	another(`"a">=bool"b"`, `0`)
	another(`"a"<="b"`, `1`)
	same(`"a" - "b"`)

	// parensExpr
	another(`(-foo + ((bar) / (baz))) + ((23))`, `((0 - foo) + (bar / baz)) + 23`)
	another(`(FOO + ((Bar) / (baZ))) + ((23))`, `(FOO + (Bar / baZ)) + 23`)
	same(`(foo, bar)`)
	another(`((foo, bar),(baz))`, `((foo, bar), baz)`)
	same(`(foo, (bar, baz), ((x, y), (z, y), xx))`)
	another(`1+(foo, bar,)`, `1 + (foo, bar)`)
	another(`((foo(bar,baz)), (1+(2)+(3,4)+()))`, `(foo(bar, baz), (3 + (3, 4)) + ())`)
	same(`()`)

	// funcExpr
	same(`f()`)
	another(`f(x,)`, `f(x)`)
	another(`-f()-Ff()`, `(0 - f()) - Ff()`)
	same(`F()`)
	another(`+F()`, `F()`)
	another(`++F()`, `F()`)
	another(`--F()`, `0 - (0 - F())`)
	same(`f(http_server_request)`)
	same(`f(http_server_request)[4s:5m] offset 10m`)
	same(`f(http_server_request)[4i:5i] offset 10i`)
	same(`F(HttpServerRequest)`)
	same(`f(job, foo)`)
	same(`F(Job, Foo)`)
	another(` FOO (bar) + f  (  m  (  ),ff(1 + (  2.5)) ,M[5m ]  , "ff"  )`, `FOO(bar) + f(m(), ff(3.5), M[5m], "ff")`)
	same(`rate(foo[5m]) keep_metric_names`)
	another(`log2(foo) KEEP_metric_names + 1 / increase(bar[5m]) keep_metric_names offset 1h @ 435`,
		`log2(foo) keep_metric_names + (1 / increase(bar[5m]) keep_metric_names offset 1h @ 435)`)
	// funcName matching keywords
	same(`by(2)`)
	same(`BY(2)`)
	same(`or(2)`)
	same(`OR(2)`)
	same(`bool(2)`)
	same(`BOOL(2)`)
	same(`rate(rate(m))`)
	same(`rate(rate(m[5m]))`)
	same(`rate(rate(m[5m])[1h:])`)
	same(`rate(rate(m[5m])[1h:3s])`)
	// funcName with escape chars
	same(`foo\(ba\-r()`)

	// aggrFuncExpr
	same(`sum(http_server_request) by ()`)
	same(`sum(http_server_request) by (job)`)
	same(`sum(http_server_request) without (job, foo)`)
	another(`sum(x,y,) without (a,b,)`, `sum(x, y) without (a, b)`)
	another(`sum by () (xx)`, `sum(xx) by ()`)
	another(`sum by (s) (xx)[5s]`, `(sum(xx) by (s))[5s]`)
	another(`SUM BY (ZZ, aa) (XX)`, `sum(XX) by (ZZ, aa)`)
	another(`sum without (a, b) (xx,2+2)`, `sum(xx, 4) without (a, b)`)
	another(`Sum WIthout (a, B) (XX,2+2)`, `sum(XX, 4) without (a, B)`)
	same(`sum(a) or sum(b)`)
	same(`sum(a) by () or sum(b) without (x, y)`)
	same(`sum(a) + sum(b)`)
	same(`sum(x) * (1 + sum(a))`)
	same(`avg(x) limit 10`)
	same(`avg(x) without (z, b) limit 1`)
	another(`avg by(x) (z) limit 20`, `avg(z) by (x) limit 20`)

	// All the above
	another(`Sum(Ff(M) * M{X=""}[5m] Offset 7m - 123, 35) BY (X, y) * F2("Test")`,
		`sum((Ff(M) * M{X=""}[5m] offset 7m) - 123, 35) by (X, y) * F2("Test")`)
	another(`# comment
		Sum(Ff(M) * M{X=""}[5m] Offset 7m - 123, 35) BY (X, y) # yet another comment
		* F2("Test")`,
		`sum((Ff(M) * M{X=""}[5m] offset 7m) - 123, 35) by (X, y) * F2("Test")`)

	// withExpr
	another(`with () x`, `x`)
	another(`with (x=1,) x`, `1`)
	another(`with (x = m offset 5h) x + x`, `m offset 5h + m offset 5h`)
	another(`with (x = m offset 5i) x + x`, `m offset 5i + m offset 5i`)
	another(`with (foo = bar{x="x"}) 1`, `1`)
	another(`with (foo = bar{x="x"}) "x"`, `"x"`)
	another(`with (f="x") f`, `"x"`)
	another(`with (foo = bar{x="x"}) x{x="y"}`, `x{x="y"}`)
	another(`with (foo = bar{x="x"}) 1+1`, `2`)
	another(`with (foo = bar{x="x"}) f()`, `f()`)
	another(`with (foo = bar{x="x"}) sum(x)`, `sum(x)`)
	another(`with (foo = bar{x="x"}) baz{foo="bar"}`, `baz{foo="bar"}`)
	another(`with (foo = bar) baz`, `baz`)
	another(`with (foo = bar) foo + foo{a="b"}`, `bar + bar{a="b"}`)
	another(`with (foo = bar, bar=baz + f()) test`, `test`)
	another(`with (ct={job="test"}) a{ct} + ct() + f({ct="x"})`, `(a{job="test"} + {job="test"}) + f({ct="x"})`)
	another(`with (ct={job="test", i="bar"}) ct + {ct, x="d"} + foo{ct, ct} + ctx(1)`,
		`(({job="test", i="bar"} + {job="test", i="bar", x="d"}) + foo{job="test", i="bar"}) + ctx(1)`)
	another(`with (foo = bar) {__name__=~"foo"}`, `{__name__=~"foo"}`)
	another(`with (foo = bar) foo{__name__="foo"}`, `bar`)
	another(`with (foo = bar) {__name__="foo", x="y"}`, `bar{x="y"}`)
	another(`with (foo(bar) = {__name__!="bar"}) foo(x)`, `{__name__!="bar"}`)
	another(`with (foo(bar) = bar{__name__="bar"}) foo(x)`, `x`)
	another(`with (foo\-bar(baz) = baz + baz) foo\-bar((x,y))`, `(x, y) + (x, y)`)
	another(`with (foo\-bar(baz) = baz + baz) foo\-bar(x*y)`, `(x * y) + (x * y)`)
	another(`with (foo\-bar(baz) = baz + baz) foo\-bar(x\*y)`, `x\*y + x\*y`)
	another(`with (foo\-bar(b\ az) = b\ az + b\ az) foo\-bar(x\*y)`, `x\*y + x\*y`)
	// override ttf to something new.
	another(`with (ttf = a) ttf + b`, `a + b`)
	// override ttf to ru
	another(`with (ttf = ru(m, n)) ttf`, `(clamp_min(n - clamp_min(m, 0), 0) / clamp_min(n, 0)) * 100`)

	// Verify withExpr recursion and forward reference
	another(`with (x = x+y, y = x+x) y ^ 2`, `((x + y) + (x + y)) ^ 2`)
	another(`with (f1(x)=f2(x), f2(x)=f1(x)^2) f1(foobar)`, `f2(foobar)`)
	another(`with (f1(x)=f2(x), f2(x)=f1(x)^2) f2(foobar)`, `f2(foobar) ^ 2`)

	// Verify withExpr funcs
	another(`with (x() = y+1) x`, `y + 1`)
	another(`with (x(foo) = foo+1) x(a)`, `a + 1`)
	another(`with (x(a, b) = a + b) x(foo, bar)`, `foo + bar`)
	another(`with (x(a, b) = a + b) x(foo, x(1, 2))`, `foo + 3`)
	another(`with (x(a) = sum(a) by (b)) x(xx) / x(y)`, `sum(xx) by (b) / sum(y) by (b)`)
	another(`with (f(a,f,x)=ff(x,f,a)) f(f(x,y,z),1,2)`, `ff(2, 1, ff(z, y, x))`)
	another(`with (f(x)=1+f(x)) f(foo{bar="baz"})`, `1 + f(foo{bar="baz"})`)
	another(`with (a=foo, y=bar, f(a)= a+a+y) f(x)`, `(x + x) + bar`)
	another(`with (f(a, b) = m{a, b}) f({a="x", b="y"}, {c="d"})`, `m{a="x", b="y", c="d"}`)
	another(`with (xx={a="x"}, f(a, b) = m{a, b}) f({xx, b="y"}, {c="d"})`, `m{a="x", b="y", c="d"}`)
	another(`with (x() = {b="c"}) foo{x}`, `foo{b="c"}`)
	another(`with (f(x)=x{foo="bar"} offset 5m) f(m offset 10m)`, `(m{foo="bar"} offset 10m) offset 5m`)
	another(`with (f(x)=x{foo="bar",bas="a"}[5m]) f(m[10m] offset 3s)`, `(m{foo="bar", bas="a"}[10m] offset 3s)[5m]`)
	another(`with (f(x)=x{foo="bar"}[5m] offset 10m) f(m{x="y"})`, `m{x="y", foo="bar"}[5m] offset 10m`)
	another(`with (f(x)=x{foo="bar"}[5m] offset 10m) f({x="y", foo="bar", foo="bar"})`, `{x="y", foo="bar"}[5m] offset 10m`)
	another(`with (f(m, x)=m{x}[5m] offset 10m) f(foo, {})`, `foo[5m] offset 10m`)
	another(`with (f(m, x)=m{x, bar="baz"}[5m] offset 10m) f(foo, {})`, `foo{bar="baz"}[5m] offset 10m`)
	another(`with (f(x)=x[5m] offset 3s) f(foo[3m]+bar)`, `(foo[3m] + bar)[5m] offset 3s`)
	another(`with (f(x)=x[5m:3s] oFFsEt 1.5m) f(sum(s) by (a,b))`, `(sum(s) by (a, b))[5m:3s] offset 1.5m`)
	another(`with (x="a", y=x) y+"bc"`, `"abc"`)
	another(`with (x="a", y="b"+x) "we"+y+"z"+f()`, `"webaz" + f()`)
	another(`with (f(x) = m{foo=x+"y", bar="y"+x, baz=x} + x) f("qwe")`, `m{foo="qwey", bar="yqwe", baz="qwe"} + "qwe"`)
	another(`with (f(a)=a) f`, `f`)
	another(`with (f\q(a)=a) f\q`, `fq`)

	// Verify withExpr for aggr func modifiers
	another(`with (f(x) = x, y = sum(m) by (f)) y`, `sum(m) by (f)`)
	another(`with (f(x) = x, y = sum(m) by (f) limit 20) y`, `sum(m) by (f) limit 20`)
	another(`with (f(x) = sum(m) by (x)) f(foo)`, `sum(m) by (foo)`)
	another(`with (f(x) = sum(m) by (x) limit 42) f(foo)`, `sum(m) by (foo) limit 42`)
	another(`with (f(x) = sum(m) by (x)) f((foo, bar, foo))`, `sum(m) by (foo, bar)`)
	another(`with (f(x) = sum(m) without (x,y)) f((a, b))`, `sum(m) without (a, b, y)`)
	another(`with (f(x) = sum(m) without (y,x)) f((a, y))`, `sum(m) without (y, a)`)
	another(`with (f(x,y) = a + on (x,y) group_left (y,bar) b) f(foo,())`, `a + on (foo) group_left (bar) b`)
	another(`with (f(x,y) = a + on (x,y) group_left (y,bar) b) f((foo),())`, `a + on (foo) group_left (bar) b`)
	another(`with (f(x,y) = a + on (x,y) group_left (y,bar) b) f((foo,xx),())`, `a + on (foo, xx) group_left (bar) b`)

	// Verify nested with exprs
	another(`with (f(x) = (with(x=y) x) + x) f(z)`, `y + z`)
	another(`with (x=foo) f(a, with (y=x) y)`, `f(a, foo)`)
	another(`with (x=foo) a * x + (with (y=x) y) / y`, `(a * foo) + (foo / y)`)
	another(`with (x = with (y = foo) y + x) x/x`, `(foo + x) / (foo + x)`)
	another(`with (
		x = {foo="bar"},
		q = m{x, y="1"},
		f(x) =
			with (
				z(y) = x + y * q
			)
			z(foo) / f(x)
	)
	f(a)`, `(a + (foo * m{foo="bar", y="1"})) / f(a)`)

	// complex withExpr
	another(`WITH (
		treshold = (0.9),
		commonFilters = {job="cacher", instance=~"1.2.3.4"},
		hits = rate(cache{type="hit", commonFilters}[5m]),
		miss = rate(cache{type="miss", commonFilters}[5m]),
		sumByInstance(arg) = sum(arg) by (instance),
		hitRatio = sumByInstance(hits) / sumByInstance(hits + miss)
	)
	hitRatio < treshold`,
		`  (
      (
        sum(rate(cache{type="hit", job="cacher", instance=~"1.2.3.4"}[5m])) by (instance)
      )
    /
      (
        sum by (instance) (
          rate(cache{type="hit", job="cacher", instance=~"1.2.3.4"}[5m]) + rate(cache{type="miss", job="cacher", instance=~"1.2.3.4"}[5m])
        )
      )
  )
<
  0.9`)
	another(`WITH (
		x2(x) = x^2,
		f(x, y) = x2(x) + x*y + x2(y)
	)
	f(a, 3)
	`, `((a ^ 2) + (a * 3)) + 9`)
	another(`WITH (
		x2(x) = x^2,
		f(x, y) = x2(x) + x*y + x2(y)
	)
	f(2, 3)
	`, `19`)
	another(`WITH (
		commonFilters = {instance="foo"},
		timeToFuckup(currv, maxv) = (maxv - currv) / rate(currv)
	)
	timeToFuckup(diskUsage{commonFilters}, maxDiskSize{commonFilters})`,
		`(maxDiskSize{instance="foo"} - diskUsage{instance="foo"}) / rate(diskUsage{instance="foo"})`)
	another(`WITH (
	       commonFilters = {job="foo", instance="bar"},
	       sumRate(m, cf) = sum(rate(m{cf})) by (job, instance),
	       hitRate(hits, misses) = sumRate(hits, commonFilters) / (sumRate(hits, commonFilters) + sumRate(misses, commonFilters))
	   )
	   hitRate(cacheHits, cacheMisses)`,
		`  (
    sum(rate(cacheHits{job="foo", instance="bar"})) by (job, instance)
  )
/
  (
      (
        sum(rate(cacheHits{job="foo", instance="bar"})) by (job, instance)
      )
    +
      (
        sum(rate(cacheMisses{job="foo", instance="bar"})) by (job, instance)
      )
  )`)
	another(`with(y=123,z=5) union(with(y=3,f(x)=x*y) f(2) + f(3), with(x=5,y=2) x*y*z)`, `union(15, 50)`)

	another(`with(sum=123,now=5) union(with(sum=3,f(x)=x*sum) f(2) + f(3), with(x=5,sum=2) x*sum*now)`, `union(15, 50)`)
	another(`WITH(now = sum(rate(my_metric_total)), before = sum(rate(my_metric_total) offset 1h)) now/before*100`, `(sum(rate(my_metric_total)) / sum(rate(my_metric_total) offset 1h)) * 100`)
	another(`with (sum = x) sum`, `x`)
	another(`with (clamp_min=x) clamp_min`, `x`)
	another(`with (now=now(), sum=sum()) now`, `now()`)
	another(`with (now=now(), sum=sum()) now()`, `now()`)
	another(`with (now(a)=now()+a) now(1)`, `now() + 1`)
	another(`with (rate(a,b)=a+b) rate(1,2)`, `3`)
	another(`with (now=now(), sum=sum()) x`, `x`)
	another(`with (rate(a) = b) c`, `c`)
	another(`rate(x) + with (rate(a,b)=a*b) rate(2,b)`, `rate(x) + (2 * b)`)
	another(`with (sum(a,b)=a+b) sum(c,d)`, `c + d`)
}

func TestPrettierShot(t *testing.T) {
	another := func(s, expected string) {
		t.Helper()

		const maxLineLength = 80
		got, err := Prettier(s, maxLineLength)
		if err != nil {
			t.Fatalf("unexpected error when parsing %q: %s", s, err)
		}

		if expected != got {
			t.Fatalf("string not prettified;\ngot:%s\nwant\n%q", got, expected)
		}
	}
	same := func(s string) {
		t.Helper()
		another(s, s)
	}

	same(`{}`)
	same(`{}[5m]`)
	same(`{}[5m:]`)
	same(`{}[5M:]`)
	same(`{}[:]`)
	another(`{}[: ]`, `{}[:]`)
	same(`{}[:3s]`)
	another(`{}[: 3s ]`, `{}[:3s]`)
	same(`{}[5m:3s]`)
	another(`{}[ 5m : 3s ]`, `{}[5m:3s]`)
	same(`{} offset 5m`)
	same(`{} offset -5m`)
	same(`{} offset 5M`)
	same(`{}[5m] offset 10y`)
	same(`{}[5.3m:3.4s] offset 10y`)
	same(`{}[:3.4s] offset 10y`)
	same(`{}[:3.4s] offset -10y`)
	same(`{Foo="bAR"}`)
	same(`{foo="bar"}`)
	same(`{foo="bar"}[5m]`)
	same(`{foo="bar"}[5m:]`)
	same(`{foo="bar"}[5m:3s]`)
	same(`{foo="bar"} offset 13.4ms`)
	same(`{foo="bar"}[5w4h-3.4m13.4ms]`)
	same(`{foo="bar"} offset 10y`)
	same(`{foo="bar"} offset -10y`)
	same(`{foo="bar"}[5m] offset 10y`)
	same(`{foo="bar"}[5m:3s] offset 10y`)
	another(`{foo="bar"}[5m] oFFSEt 10y`, `{foo="bar"}[5m] offset 10y`)
	same("METRIC")
	same("metric")
	same("m_e:tri44:_c123")
	another("-metric", "0 - metric")
	same(`metric offset 10h`)
	same("metric[5m]")
	same("metric[5m:3s]")
	same("metric[5m] offset 10h")
	same("metric[5m:3s] offset 10h")
	same("metric[5i:3i] offset 10i")
	same(`metric{foo="bar"}`)
	same(`metric{foo="bar"} offset 10h`)
	same(`metric{foo!="bar"}[2d]`)
	same(`metric{foo="bar"}[2d] offset 10h`)
	same(`metric{foo="bar", b="sdfsdf"}[2d:3h] offset 10h`)
	same(`metric{foo="bar", b="sdfsdf"}[2d:3h] offset 10`)
	same(`metric{foo="bar", b="sdfsdf"}[2d:3] offset 10h`)
	same(`metric{foo="bar", b="sdfsdf"}[2:3h] offset 10h`)
	same(`metric{foo="bar", b="sdfsdf"}[2.34:5.6] offset 3600.5`)
	same(`metric{foo="bar", b="sdfsdf"}[234:56] offset -3600`)
	another(`  metric  {  foo  = "bar"  }  [  2d ]   offset   10h  `, `metric{foo="bar"}[2d] offset 10h`)
	// @ modifier
	// See https://prometheus.io/docs/prometheus/latest/querying/basics/#modifier
	same(`foo @ 123.45`)
	same(`foo\@ @ 123.45`)
	same(`{foo=~"bar"} @ end()`)
	same(`foo{bar="baz"} @ start()`)
	same(`foo{bar="baz"}[5m] @ 12345`)
	same(`foo{bar="baz"}[5m:4s] offset 5m @ (end() - 3.5m)`)
	another(`foo{bar="baz"}[5m:4s] @ (end() - 3.5m) offset 2.4h`, `foo{bar="baz"}[5m:4s] offset 2.4h @ (end() - 3.5m)`)
	another(`foo @ start() + (bar offset 3m @ end()) / baz OFFSET -5m`, `foo @ start() + (bar offset 3m @ end() / baz offset -5m)`)
	same(`sum(foo) @ start() + rate(bar @ (end() - 5m))`)
	another(`time() @ (start())`, `time() @ start()`)
	another(`time() @ (start()+(1+1))`, `time() @ (start() + 2)`)
	same(`time() @ (end() - 10m)`)
	// metric name matching keywords
	same("rate")
	same("RATE")
	same("by")
	same("BY")
	same("bool")
	same("BOOL")
	same("unless")
	same("UNLESS")
	same("Ignoring")
	same("with")
	same("WITH")
	same("With")
	same("offset")
	same("keep_metric_names")
	same("alias")
	same(`alias{foo="bar"}`)
	same(`aLIas{alias="aa"}`)
	another(`al\ias`, `alias`)
	// identifiers with with escape chars
	same(`foo\ bar`)
	same(`foo\-bar\{{baz\+bar="aa"}`)
	another(`\x2E\x2ef\oo{b\xEF\ar="aa"}`, `\..foo{bïar="aa"}`)
	same(`温度{房间="水电费"}[5m] offset 10m`)
	another(`\温\度{\房\间="水电费"}[5m] offset 10m`, `温度{房间="水电费"}[5m] offset 10m`)
	same(`sum(fo\|o) by (b\|a, x)`)
	another(`sum(x) by (b\x7Ca)`, `sum(x) by (b\|a)`)
	// Duplicate filters
	same(`foo{__name__="bar"}`)
	same(`foo{a="b", a="c", __name__="aaa", b="d"}`)
	// Metric filters ending with comma
	another(`m{foo="bar",}`, `m{foo="bar"}`)
	// String concat in tag value
	another(`m{foo="bar" + "baz"}`, `m{foo="barbaz"}`)

	// Valid regexp
	same(`foo{bar=~"x"}`)
	same(`foo{bar=~"^x"}`)
	same(`foo{bar=~"^x$"}`)
	same(`foo{bar=~"^(a[bc]|d)$"}`)
	same(`foo{bar!~"x"}`)
	same(`foo{bar!~"^x"}`)
	same(`foo{bar!~"^x$"}`)
	same(`foo{bar!~"^(a[bc]|d)$"}`)

	// stringExpr
	same(`""`)
	same(`"\n\t\r 12:{}[]()44"`)
	another(`''`, `""`)
	another("``", `""`)
	another("   `foo\"b'ar`  ", "\"foo\\\"b'ar\"")
	another(`  'foo\'bar"BAZ'  `, `"foo'bar\"BAZ"`)
	// string concat
	another(`"foo"+'bar'`, `"foobar"`)

	// numberExpr
	same(`1`)
	same(`123.`)
	another(`-123.`, `-123`)
	same(`foo - 123.`)
	same(`12.e+4`)
	same(`12Ki`)
	same(`1.23Gb`)
	same(`foo - 23M`)
	another(`-1.23Gb`, `-1.23e+09`)
	same(`1.23`)
	same(`0.23`)
	same(`1.2e+45`)
	same(`1.2e-45`)
	same(`-1`)
	same(`-1.23`)
	same(`-0.23`)
	same(`-1.2e+45`)
	same(`-1.2e-45`)
	same(`-1.2e-45`)
	same(`12.5E34`)
	another(`-.2`, `-0.2`)
	another(`-.2E-2`, `-0.002`)
	same(`NaN`)
	same(`nan`)
	same(`NAN`)
	same(`nAN`)
	same(`Inf`)
	same(`INF`)
	same(`inf`)
	another(`+Inf`, `Inf`)
	same(`-Inf`)
	another(`-inF`, `-Inf`)
	same(`0x12`)
	same(`0x3b`)
	another(`-0x3b`, `-59`)
	another(`+0X3B`, `0X3B`)
	same(`0b1011`)
	same(`073`)
	another(`-0o12`, `-10`)

	// durationExpr
	same(`1h`)
	another(`-1h`, `0 - 1h`)
	same(`0.34h4m5s`)
	same(`0.34H4m5S`)
	another(`-0.34h4m5s`, `0 - 0.34h4m5s`)
	same(`sum_over_tme(m[1h]) / 1h`)
	same(`sum_over_time(m[3600]) / 3600`)

	// binaryOpExpr
	another(`nan == nan`, `NaN`)
	another(`nan ==bool nan`, `1`)
	another(`nan !=bool nan`, `0`)
	another(`nan !=bool 2`, `1`)
	another(`2 !=bool nan`, `1`)
	another(`nan >bool nan`, `0`)
	another(`nan <bool nan`, `0`)
	another(`1 ==bool nan`, `0`)
	another(`NaN !=bool 1`, `1`)
	another(`inf >=bool 2`, `1`)
	another(`-1 >bool -inf`, `1`)
	another(`-1 <bool -inf`, `0`)
	another(`nan + 2 *3 * inf`, `NaN`)
	another(`INF - Inf`, `NaN`)
	another(`Inf + inf`, `+Inf`)
	another(`1/0`, `+Inf`)
	another(`0/0`, `NaN`)
	another(`-m`, `0 - m`)
	same(`m + ignoring () n[5m]`)
	another(`M + IGNORING () N[5m]`, `M + ignoring () N[5m]`)
	same(`m + on (foo) n[5m]`)
	another(`m + ON (Foo) n[5m]`, `m + on (Foo) n[5m]`)
	same(`m + ignoring (a, b) n[5m]`)
	another(`1 or 2`, `1`)
	another(`1 and 2`, `1`)
	another(`1 unless 2`, `NaN`)
	another(`1 default 2`, `1`)
	another(`1 default NaN`, `1`)
	another(`NaN default 2`, `2`)
	another(`1 > 2`, `NaN`)
	another(`1 > bool 2`, `0`)
	another(`3 >= 2`, `3`)
	another(`3 <= bool 2`, `0`)
	another(`1 + -2 - 3`, `-4`)
	another(`1 / 0 + 2`, `+Inf`)
	another(`2 + -1 / 0`, `-Inf`)
	another(`(-1) ^ 0.5`, `NaN`)
	another(`-1 ^ 0.5`, `-1`)
	another(`512.5 - (1 + 3) * (2 ^ 2) ^ 3`, `256.5`)
	another(`1 == bool 1 != bool 24 < bool 4 > bool -1`, `1`)
	another(`1 == bOOl 1 != BOOL 24 < Bool 4 > booL -1`, `1`)
	another(`m1+on(foo)group_left m2`, `m1 + on (foo) group_left () m2`)
	another(`M1+ON(FOO)GROUP_left M2`, `M1 + on (FOO) group_left () M2`)
	same(`m1 + on (foo) group_right () m2`)
	same(`m1 + on (foo, bar) group_right (x, y) m2`)
	another(`m1 + on (foo, bar,) group_right (x, y,) m2`, `m1 + on (foo, bar) group_right (x, y) m2`)
	same(`m1 == bool on (foo, bar) group_right (x, y) m2`)
	another(`5 - 1 + 3 * 2 ^ 2 ^ 3 - 2  OR Metric {Bar= "Baz", aaa!="bb",cc=~"dd" ,zz !~"ff" } `,
		`770 or Metric{Bar="Baz", aaa!="bb", cc=~"dd", zz!~"ff"}`)
	same(`"foo" + bar()`)
	same(`"foo" + bar{x="y"}`)
	same(`("foo"[3s] + bar{x="y"})[5m:3s] offset 10s`)
	same(`("foo"[3s] + bar{x="y"})[5i:3i] offset 10i`)
	same(`bar + "foo" offset 3s`)
	same(`bar + "foo" offset 3i`)
	another(`1+2 if 2>3`, `NaN`)
	another(`1+4 if 2<3`, `5`)
	another(`2+6 default 3 if 2>3`, `8`)
	another(`2+6 if 2>3 default NaN`, `NaN`)
	another(`42 if 3>2 if 2+2<5`, `42`)
	another(`42 if 3>2 if 2+2>=5`, `NaN`)
	another(`1+2 ifnot 2>3`, `3`)
	another(`1+4 ifnot 2<3`, `NaN`)
	another(`2+6 default 3 ifnot 2>3`, `8`)
	another(`2+6 ifnot 2>3 default NaN`, `8`)
	another(`42 if 3>2 ifnot 2+2<5`, `NaN`)
	another(`42 if 3>2 ifnot 2+2>=5`, `42`)
	another(`"foo" + "bar"`, `"foobar"`)
	another(`"foo"=="bar"`, `NaN`)
	another(`"foo"=="foo"`, `1`)
	another(`"foo"!="bar"`, `1`)
	another(`"foo"+"bar"+"baz"`, `"foobarbaz"`)
	another(`"a">"b"`, `NaN`)
	another(`"a">bool"b"`, `0`)
	another(`"a"<"b"`, `1`)
	another(`"a">="b"`, `NaN`)
	another(`"a">=bool"b"`, `0`)
	another(`"a"<="b"`, `1`)
	same(`"a" - "b"`)

	// parensExpr
	another(`(-foo + ((bar) / (baz))) + ((23))`, `((0 - foo) + (bar / baz)) + 23`)
	another(`(FOO + ((Bar) / (baZ))) + ((23))`, `(FOO + (Bar / baZ)) + 23`)
	same(`(foo, bar)`)
	another(`((foo, bar),(baz))`, `((foo, bar), baz)`)
	same(`(foo, (bar, baz), ((x, y), (z, y), xx))`)
	another(`1+(foo, bar,)`, `1 + (foo, bar)`)
	another(`((foo(bar,baz)), (1+(2)+(3,4)+()))`, `(foo(bar, baz), (3 + (3, 4)) + ())`)
	same(`()`)

	// funcExpr
	same(`f()`)
	another(`f(x,)`, `f(x)`)
	another(`-f()-Ff()`, `(0 - f()) - Ff()`)
	same(`F()`)
	another(`+F()`, `F()`)
	another(`++F()`, `F()`)
	another(`--F()`, `0 - (0 - F())`)
	same(`f(http_server_request)`)
	same(`f(http_server_request)[4s:5m] offset 10m`)
	same(`f(http_server_request)[4i:5i] offset 10i`)
	same(`F(HttpServerRequest)`)
	same(`f(job, foo)`)
	same(`F(Job, Foo)`)
	another(` FOO (bar) + f  (  m  (  ),ff(1 + (  2.5)) ,M[5m ]  , "ff"  )`, `FOO(bar) + f(m(), ff(3.5), M[5m], "ff")`)
	same(`rate(foo[5m]) keep_metric_names`)
	another(`log2(foo) KEEP_metric_names + 1 / increase(bar[5m]) keep_metric_names offset 1h @ 435`,
		`  log2(foo) keep_metric_names
+
  (
    1 / increase(bar[5m]) keep_metric_names offset 1h @ 435
  )`)
	// funcName matching keywords
	same(`by(2)`)
	same(`BY(2)`)
	same(`or(2)`)
	same(`OR(2)`)
	same(`bool(2)`)
	same(`BOOL(2)`)
	same(`rate(rate(m))`)
	same(`rate(rate(m[5m]))`)
	same(`rate(rate(m[5m])[1h:])`)
	same(`rate(rate(m[5m])[1h:3s])`)
	// funcName with escape chars
	same(`foo\(ba\-r()`)

	// aggrFuncExpr
	same(`sum(http_server_request) by ()`)
	same(`sum(http_server_request) by (job)`)
	same(`sum(http_server_request) without (job, foo)`)
	another(`sum(x,y,) without (a,b,)`, `sum(x, y) without (a, b)`)
	another(`sum by () (xx)`, `sum(xx) by ()`)
	another(`sum by (s) (xx)[5s]`, `(sum(xx) by (s))[5s]`)
	another(`SUM BY (ZZ, aa) (XX)`, `sum(XX) by (ZZ, aa)`)
	another(`sum without (a, b) (xx,2+2)`, `sum(xx, 4) without (a, b)`)
	another(`Sum WIthout (a, B) (XX,2+2)`, `sum(XX, 4) without (a, B)`)
	same(`sum(a) or sum(b)`)
	same(`sum(a) by () or sum(b) without (x, y)`)
	same(`sum(a) + sum(b)`)
	same(`sum(x) * (1 + sum(a))`)
	same(`avg(x) limit 10`)
	same(`avg(x) without (z, b) limit 1`)
	another(`avg by(x) (z) limit 20`, `avg(z) by (x) limit 20`)

	// All the above
	another(`Sum(Ff(M) * M{X=""}[5m] Offset 7m - 123, 35) BY (X, y) * F2("Test")`,
		`sum((Ff(M) * M{X=""}[5m] offset 7m) - 123, 35) by (X, y) * F2("Test")`)
	another(`# comment
		Sum(Ff(M) * M{X=""}[5m] Offset 7m - 123, 35) BY (X, y) # yet another comment
		* F2("Test")`,
		`sum((Ff(M) * M{X=""}[5m] offset 7m) - 123, 35) by (X, y) * F2("Test")`)

	// withExpr
	another(`with () x`, `x`)
	another(`with (x=1,) x`, `1`)
	another(`with (x = m offset 5h) x + x`, `m offset 5h + m offset 5h`)
	another(`with (x = m offset 5i) x + x`, `m offset 5i + m offset 5i`)
	another(`with (foo = bar{x="x"}) 1`, `1`)
	another(`with (foo = bar{x="x"}) "x"`, `"x"`)
	another(`with (f="x") f`, `"x"`)
	another(`with (foo = bar{x="x"}) x{x="y"}`, `x{x="y"}`)
	another(`with (foo = bar{x="x"}) 1+1`, `2`)
	another(`with (foo = bar{x="x"}) f()`, `f()`)
	another(`with (foo = bar{x="x"}) sum(x)`, `sum(x)`)
	another(`with (foo = bar{x="x"}) baz{foo="bar"}`, `baz{foo="bar"}`)
	another(`with (foo = bar) baz`, `baz`)
	another(`with (foo = bar) foo + foo{a="b"}`, `bar + bar{a="b"}`)
	another(`with (foo = bar, bar=baz + f()) test`, `test`)
	another(`with (ct={job="test"}) a{ct} + ct() + f({ct="x"})`, `(a{job="test"} + {job="test"}) + f({ct="x"})`)
	another(`with (ct={job="test", i="bar"}) ct + {ct, x="d"} + foo{ct, ct} + ctx(1)`,
		`  (
      (
        {job="test", i="bar"} + {job="test", i="bar", x="d"}
      )
    +
      foo{job="test", i="bar"}
  )
+
  ctx(1)`)
	another(`with (foo = bar) {__name__=~"foo"}`, `{__name__=~"foo"}`)
	another(`with (foo = bar) foo{__name__="foo"}`, `bar`)
	another(`with (foo = bar) {__name__="foo", x="y"}`, `bar{x="y"}`)
	another(`with (foo(bar) = {__name__!="bar"}) foo(x)`, `{__name__!="bar"}`)
	another(`with (foo(bar) = bar{__name__="bar"}) foo(x)`, `x`)
	another(`with (foo\-bar(baz) = baz + baz) foo\-bar((x,y))`, `(x, y) + (x, y)`)
	another(`with (foo\-bar(baz) = baz + baz) foo\-bar(x*y)`, `(x * y) + (x * y)`)
	another(`with (foo\-bar(baz) = baz + baz) foo\-bar(x\*y)`, `x\*y + x\*y`)
	another(`with (foo\-bar(b\ az) = b\ az + b\ az) foo\-bar(x\*y)`, `x\*y + x\*y`)
	// override ttf to something new.
	another(`with (ttf = a) ttf + b`, `a + b`)
	// override ttf to ru
	another(`with (ttf = ru(m, n)) ttf`, `(clamp_min(n - clamp_min(m, 0), 0) / clamp_min(n, 0)) * 100`)

	// Verify withExpr recursion and forward reference
	another(`with (x = x+y, y = x+x) y ^ 2`, `((x + y) + (x + y)) ^ 2`)
	another(`with (f1(x)=f2(x), f2(x)=f1(x)^2) f1(foobar)`, `f2(foobar)`)
	another(`with (f1(x)=f2(x), f2(x)=f1(x)^2) f2(foobar)`, `f2(foobar) ^ 2`)

	// Verify withExpr funcs
	another(`with (x() = y+1) x`, `y + 1`)
	another(`with (x(foo) = foo+1) x(a)`, `a + 1`)
	another(`with (x(a, b) = a + b) x(foo, bar)`, `foo + bar`)
	another(`with (x(a, b) = a + b) x(foo, x(1, 2))`, `foo + 3`)
	another(`with (x(a) = sum(a) by (b)) x(xx) / x(y)`, `sum(xx) by (b) / sum(y) by (b)`)
	another(`with (f(a,f,x)=ff(x,f,a)) f(f(x,y,z),1,2)`, `ff(2, 1, ff(z, y, x))`)
	another(`with (f(x)=1+f(x)) f(foo{bar="baz"})`, `1 + f(foo{bar="baz"})`)
	another(`with (a=foo, y=bar, f(a)= a+a+y) f(x)`, `(x + x) + bar`)
	another(`with (f(a, b) = m{a, b}) f({a="x", b="y"}, {c="d"})`, `m{a="x", b="y", c="d"}`)
	another(`with (xx={a="x"}, f(a, b) = m{a, b}) f({xx, b="y"}, {c="d"})`, `m{a="x", b="y", c="d"}`)
	another(`with (x() = {b="c"}) foo{x}`, `foo{b="c"}`)
	another(`with (f(x)=x{foo="bar"} offset 5m) f(m offset 10m)`, `(m{foo="bar"} offset 10m) offset 5m`)
	another(`with (f(x)=x{foo="bar",bas="a"}[5m]) f(m[10m] offset 3s)`, `(m{foo="bar", bas="a"}[10m] offset 3s)[5m]`)
	another(`with (f(x)=x{foo="bar"}[5m] offset 10m) f(m{x="y"})`, `m{x="y", foo="bar"}[5m] offset 10m`)
	another(`with (f(x)=x{foo="bar"}[5m] offset 10m) f({x="y", foo="bar", foo="bar"})`, `{x="y", foo="bar"}[5m] offset 10m`)
	another(`with (f(m, x)=m{x}[5m] offset 10m) f(foo, {})`, `foo[5m] offset 10m`)
	another(`with (f(m, x)=m{x, bar="baz"}[5m] offset 10m) f(foo, {})`, `foo{bar="baz"}[5m] offset 10m`)
	another(`with (f(x)=x[5m] offset 3s) f(foo[3m]+bar)`, `(foo[3m] + bar)[5m] offset 3s`)
	another(`with (f(x)=x[5m:3s] oFFsEt 1.5m) f(sum(s) by (a,b))`, `(sum(s) by (a, b))[5m:3s] offset 1.5m`)
	another(`with (x="a", y=x) y+"bc"`, `"abc"`)
	another(`with (x="a", y="b"+x) "we"+y+"z"+f()`, `"webaz" + f()`)
	another(`with (f(x) = m{foo=x+"y", bar="y"+x, baz=x} + x) f("qwe")`, `m{foo="qwey", bar="yqwe", baz="qwe"} + "qwe"`)
	another(`with (f(a)=a) f`, `f`)
	another(`with (f\q(a)=a) f\q`, `fq`)

	// Verify withExpr for aggr func modifiers
	another(`with (f(x) = x, y = sum(m) by (f)) y`, `sum(m) by (f)`)
	another(`with (f(x) = x, y = sum(m) by (f) limit 20) y`, `sum(m) by (f) limit 20`)
	another(`with (f(x) = sum(m) by (x)) f(foo)`, `sum(m) by (foo)`)
	another(`with (f(x) = sum(m) by (x) limit 42) f(foo)`, `sum(m) by (foo) limit 42`)
	another(`with (f(x) = sum(m) by (x)) f((foo, bar, foo))`, `sum(m) by (foo, bar)`)
	another(`with (f(x) = sum(m) without (x,y)) f((a, b))`, `sum(m) without (a, b, y)`)
	another(`with (f(x) = sum(m) without (y,x)) f((a, y))`, `sum(m) without (y, a)`)
	another(`with (f(x,y) = a + on (x,y) group_left (y,bar) b) f(foo,())`, `a + on (foo) group_left (bar) b`)
	another(`with (f(x,y) = a + on (x,y) group_left (y,bar) b) f((foo),())`, `a + on (foo) group_left (bar) b`)
	another(`with (f(x,y) = a + on (x,y) group_left (y,bar) b) f((foo,xx),())`, `a + on (foo, xx) group_left (bar) b`)

	// Verify nested with exprs
	another(`with (f(x) = (with(x=y) x) + x) f(z)`, `y + z`)
	another(`with (x=foo) f(a, with (y=x) y)`, `f(a, foo)`)
	another(`with (x=foo) a * x + (with (y=x) y) / y`, `(a * foo) + (foo / y)`)
	another(`with (x = with (y = foo) y + x) x/x`, `(foo + x) / (foo + x)`)
	another(`with (
		x = {foo="bar"},
		q = m{x, y="1"},
		f(x) =
			with (
				z(y) = x + y * q
			)
			z(foo) / f(x)
	)
	f(a)`, `(a + (foo * m{foo="bar", y="1"})) / f(a)`)

	// complex withExpr
	another(`WITH (
		treshold = (0.9),
		commonFilters = {job="cacher", instance=~"1.2.3.4"},
		hits = rate(cache{type="hit", commonFilters}[5m]),
		miss = rate(cache{type="miss", commonFilters}[5m]),
		sumByInstance(arg) = sum(arg) by (instance),
		hitRatio = sumByInstance(hits) / sumByInstance(hits + miss)
	)
	hitRatio < treshold`,
		`  (
      (
        sum by (instance) (
          rate(cache{type="hit", job="cacher", instance=~"1.2.3.4"}[5m])
        )
      )
    /
      (
        sum by (instance) (
            rate(cache{type="hit", job="cacher", instance=~"1.2.3.4"}[5m])
          +
            rate(cache{type="miss", job="cacher", instance=~"1.2.3.4"}[5m])
        )
      )
  )
<
  0.9`)
	another(`WITH (
		x2(x) = x^2,
		f(x, y) = x2(x) + x*y + x2(y)
	)
	f(a, 3)
	`, `((a ^ 2) + (a * 3)) + 9`)
	another(`WITH (
		x2(x) = x^2,
		f(x, y) = x2(x) + x*y + x2(y)
	)
	f(2, 3)
	`, `19`)
	another(`WITH (
		commonFilters = {instance="foo"},
		timeToFuckup(currv, maxv) = (maxv - currv) / rate(currv)
	)
	timeToFuckup(diskUsage{commonFilters}, maxDiskSize{commonFilters})`,
		`  (
    maxDiskSize{instance="foo"} - diskUsage{instance="foo"}
  )
/
  rate(diskUsage{instance="foo"})`)
	another(`WITH (
	       commonFilters = {job="foo", instance="bar"},
	       sumRate(m, cf) = sum(rate(m{cf})) by (job, instance),
	       hitRate(hits, misses) = sumRate(hits, commonFilters) / (sumRate(hits, commonFilters) + sumRate(misses, commonFilters))
	   )
	   hitRate(cacheHits, cacheMisses)`,
		`  (
    sum(rate(cacheHits{job="foo", instance="bar"})) by (job, instance)
  )
/
  (
      (
        sum(rate(cacheHits{job="foo", instance="bar"})) by (job, instance)
      )
    +
      (
        sum(rate(cacheMisses{job="foo", instance="bar"})) by (job, instance)
      )
  )`)
	another(`with(y=123,z=5) union(with(y=3,f(x)=x*y) f(2) + f(3), with(x=5,y=2) x*y*z)`, `union(15, 50)`)

	another(`with(sum=123,now=5) union(with(sum=3,f(x)=x*sum) f(2) + f(3), with(x=5,sum=2) x*sum*now)`, `union(15, 50)`)
	another(`WITH(now = sum(rate(my_metric_total)), before = sum(rate(my_metric_total) offset 1h)) now/before*100`, `(sum(rate(my_metric_total)) / sum(rate(my_metric_total) offset 1h)) * 100`)
	another(`with (sum = x) sum`, `x`)
	another(`with (clamp_min=x) clamp_min`, `x`)
	another(`with (now=now(), sum=sum()) now`, `now()`)
	another(`with (now=now(), sum=sum()) now()`, `now()`)
	another(`with (now(a)=now()+a) now(1)`, `now() + 1`)
	another(`with (rate(a,b)=a+b) rate(1,2)`, `3`)
	another(`with (now=now(), sum=sum()) x`, `x`)
	another(`with (rate(a) = b) c`, `c`)
	another(`rate(x) + with (rate(a,b)=a*b) rate(2,b)`, `rate(x) + (2 * b)`)
	another(`with (sum(a,b)=a+b) sum(c,d)`, `c + d`)

}

func TestLongExpressions(t *testing.T) {
	another := func(s, expected string) {
		t.Helper()

		const maxLineLength = 10
		got, err := Prettier(s, maxLineLength)
		if err != nil {
			t.Fatalf("unexpected error when parsing %q: %s", s, err)
		}

		if expected != got {
			t.Fatalf("string not prettified;\ngot:%s\nwant\n%q", got, expected)
		}
	}

	another(`((node_memory_MemTotal_bytes{instance=~"node:port", job=~"job"}-node_memory_MemFree_bytes{instance=~"node:port", job=~"job"})/node_memory_MemTotal_bytes{instance=~"node:port", job=~"job"})*100`, `  (
      (
          node_memory_MemTotal_bytes{instance=~"node:port", job=~"job"}
        -
          node_memory_MemFree_bytes{instance=~"node:port", job=~"job"}
      )
    /
      node_memory_MemTotal_bytes{instance=~"node:port", job=~"job"}
  )
*
  100`)
	another(`(((count(count(node_cpu_seconds_total{instance=~"node:port",job=~"job"}) by (cpu)))-avg(sum by (mode) (rate(node_cpu_seconds_total{mode='idle',instance=~"node:port",job=~"job"}[5m]))))*100)/count(count(node_cpu_seconds_total{instance=~"node:port",job=~"job"}) by (cpu))`, `  (
      (
          (
            count (
              count by (cpu) (
                node_cpu_seconds_total{instance=~"node:port", job=~"job"}
              )
            )
          )
        -
          (
            avg (
              sum by (mode) (
                rate (
                  node_cpu_seconds_total{mode="idle", instance=~"node:port", job=~"job"}[5m]
                )
              )
            )
          )
      )
    *
      100
  )
/
  (
    count (
      count by (cpu) (
        node_cpu_seconds_total{instance=~"node:port", job=~"job"}
      )
    )
  )`)
	another(`sum by(job,foo) (sum by(job,foo) (sum by(job,foo) (task:errors:rate10s{job="s"})))`, `sum by (job, foo) (
  sum by (job, foo) (
    sum by (job, foo) (
      task:errors:rate10s{job="s"}
    )
  )
)`)
	another(`foo_1 + ignoring(foo) foo_2 + ignoring(job) group_left foo_3 + on(instance) group_right foo_4`, `  (
      (
          foo_1
        + ignoring (foo)
          foo_2
      )
    + ignoring (job) group_left ()
      foo_3
  )
+ on (instance) group_right ()
  foo_4`)
	another(`irate(very_long_vector_selector[10m:1m] @ start() offset 1m)`, `irate (
  very_long_vector_selector[10m:1m] offset 1m @ (start())
)`)
	another(`histogram_quantile(0.9, rate(instance_cpu_time_seconds{app="webapp", job="agent", instance="cluster-admin"}[5m]))`, `histogram_quantile (
  0.9,
  rate (
    instance_cpu_time_seconds{app="webapp", job="agent", instance="cluster-admin"}[5m]
  )
)`)
	another(`topk(10, (sum without(env) (instance_cpu_time_ns{app="app", job="admin", revision="14d2e34", env="dev", job="cluster-admin"})))`, `topk (
  10,
  sum without (env) (
    instance_cpu_time_ns{app="app", job="admin", revision="14d2e34", env="dev", job="cluster-admin"}
  )
)`)
	another(`max_over_time(rate(new_http_request_duration_seconds_count[1m])[1m:] @ start() offset 1m)`, `max_over_time (
  rate (
    new_http_request_duration_seconds_count[1m]
  )[1m:] offset 1m @ (start())
)`)
	another(`label_replace(label_replace(up{job="vmadmin",service="host:port"}, "foo", "$1", "service", "(.*):.*"), "foo", "$1", "service", "(.*):.*")`, `label_replace (
  label_replace (
    up{job="vmadmin", service="host:port"},
    "foo",
    "$1",
    "service",
    "(.*):.*"
  ),
  "foo",
  "$1",
  "service",
  "(.*):.*"
)`)
	another(`min(vm_free_disk_space_bytes{job=~"job_storage", instance=~"instance"}/ignoring(path) ((rate(vm_rows_added_to_storage_total{job=~"job_storage", instance=~"instance"}[1d])-ignoring(type) rate(vm_deduplicated_samples_total{job=~"job_storage", instance=~"instance", type="merge"}[1d])) * scalar(sum(vm_data_size_bytes{job=~"job_storage", instance=~"instance", type!~"indexdb.*"})/sum(vm_rows{job=~"job_storage", instance=~"instance", type!~"indexdb.*"}))))`, `min (
    vm_free_disk_space_bytes{job=~"job_storage", instance=~"instance"}
  / ignoring (path)
    (
        (
            rate (
              vm_rows_added_to_storage_total{job=~"job_storage", instance=~"instance"}[1d]
            )
          - ignoring (type)
            rate (
              vm_deduplicated_samples_total{job=~"job_storage", instance=~"instance", type="merge"}[1d]
            )
        )
      *
        scalar (
            (
              sum (
                vm_data_size_bytes{job=~"job_storage", instance=~"instance", type!~"indexdb.*"}
              )
            )
          /
            (
              sum (
                vm_rows{job=~"job_storage", instance=~"instance", type!~"indexdb.*"}
              )
            )
        )
    )
)`)
	another(`max(rate(process_cpu_seconds_total{job=~"job_storage", instance=~"instance"}[10m])/process_cpu_cores_available{job=~"job_storage", instance=~"instance"})`, `max (
    rate (
      process_cpu_seconds_total{job=~"job_storage", instance=~"instance"}[10m]
    )
  /
    process_cpu_cores_available{job=~"job_storage", instance=~"instance"}
)`)
	another(`max(sum(vm_data_size_bytes{job=~"$job", instance=~"$instance"}) by(job, instance) /(sum(vm_free_disk_space_bytes{job=~"$job", instance=~"$instance"}) by(job, instance) +sum(vm_data_size_bytes{job=~"$job", instance=~"$instance"}) by(job, instance)))`, `max (
    (
      sum by (job, instance) (
        vm_data_size_bytes{job=~"$job", instance=~"$instance"}
      )
    )
  /
    (
        (
          sum by (job, instance) (
            vm_free_disk_space_bytes{job=~"$job", instance=~"$instance"}
          )
        )
      +
        (
          sum by (job, instance) (
            vm_data_size_bytes{job=~"$job", instance=~"$instance"}
          )
        )
    )
)`)
	another(`max without (endpoint) (sum without (instance) (up{job=~".*etcd.*"} == bool 0) or count without (To) (sum without (instance) (rate(etcd_network_peer_sent_failures_total{job=~".*etcd.*"}[120s])) > 0.01))> 0`, `  (
    max without (endpoint) (
        (
          sum without (instance) (
              up{job=~".*etcd.*"}
            == bool
              0
          )
        )
      or
        (
          count without (To) (
              (
                sum without (instance) (
                  rate (
                    etcd_network_peer_sent_failures_total{job=~".*etcd.*"}[120s]
                  )
                )
              )
            >
              0.01
          )
        )
    )
  )
>
  0`)
	another(`histogram_quantile(0.99, sum(rate(grpc_server_handling_seconds_bucket{job=~".*etcd.*", grpc_method!="Defragment", grpc_type="unary"}[5m])) without(grpc_type))> 0.15`, `  histogram_quantile (
    0.99,
    sum without (grpc_type) (
      rate (
        grpc_server_handling_seconds_bucket{job=~".*etcd.*", grpc_method!="Defragment", grpc_type="unary"}[5m]
      )
    )
  )
>
  0.15`)
	another(`(rate(prometheus_tsdb_head_samples_appended_total{job="prometheus"}[5m]) <= 0 and (sum without(scrape_job) (prometheus_target_metadata_cache_entries{job="prometheus"}) > 0 or sum without(rule_group) (prometheus_rule_group_rules{job="prometheus"}) > 0))`, `  (
      rate (
        prometheus_tsdb_head_samples_appended_total{job="prometheus"}[5m]
      )
    <=
      0
  )
and
  (
      (
          (
            sum without (scrape_job) (
              prometheus_target_metadata_cache_entries{job="prometheus"}
            )
          )
        >
          0
      )
    or
      (
          (
            sum without (rule_group) (
              prometheus_rule_group_rules{job="prometheus"}
            )
          )
        >
          0
      )
  )`)
	another(`(histogram_quantile(0.90, sum by (job, instance, namespace, type, le) (rate(promscale_ingest_duration_seconds_bucket[5m]))) > 10 and sum by (job, instance, namespace, type) (rate(promscale_ingest_duration_seconds_bucket[5m]))) > 0`, `  (
      (
          histogram_quantile (
            0.90,
            sum by (job, instance, namespace, type, le) (
              rate (
                promscale_ingest_duration_seconds_bucket[5m]
              )
            )
          )
        >
          10
      )
    and
      (
        sum by (job, instance, namespace, type) (
          rate (
            promscale_ingest_duration_seconds_bucket[5m]
          )
        )
      )
  )
>
  0`)
	another(`1 - ((node_memory_MemAvailable_bytes{job="node"} or (node_memory_Buffers_bytes{job="node"} + node_memory_Cached_bytes{job="node"} + node_memory_MemFree_bytes{job="node"} + node_memory_Slab_bytes{job="node"}) ) / node_memory_MemTotal_bytes{job="node"})`, `  1
-
  (
      (
          node_memory_MemAvailable_bytes{job="node"}
        or
          (
              (
                  (
                      node_memory_Buffers_bytes{job="node"}
                    +
                      node_memory_Cached_bytes{job="node"}
                  )
                +
                  node_memory_MemFree_bytes{job="node"}
              )
            +
              node_memory_Slab_bytes{job="node"}
          )
      )
    /
      node_memory_MemTotal_bytes{job="node"}
  )`)
	another(`(node_timex_offset_seconds{job="node"} > 0.05 and deriv(node_timex_offset_seconds{job="node"}[5m]) >= 0) or (node_timex_offset_seconds{job="node"} < -0.05 and deriv(node_timex_offset_seconds{job="node"}[5m]) <= 0)`, `  (
      (
          node_timex_offset_seconds{job="node"}
        >
          0.05
      )
    and
      (
          deriv (
            node_timex_offset_seconds{job="node"}[5m]
          )
        >=
          0
      )
  )
or
  (
      (
          node_timex_offset_seconds{job="node"}
        <
          -0.05
      )
    and
      (
          deriv (
            node_timex_offset_seconds{job="node"}[5m]
          )
        <=
          0
      )
  )`)
	another(`(count by (job) (changes(process_start_time_seconds{job="alertmanager"}[10m]) > 4) / count by (job) (up{job="alertmanager"})) >= 0.5`, `  (
      (
        count by (job) (
            changes (
              process_start_time_seconds{job="alertmanager"}[10m]
            )
          >
            4
        )
      )
    /
      (
        count by (job) (
          up{job="alertmanager"}
        )
      )
  )
>=
  0.5`)
	another(`WITH (commonFilters = {instance=~"$node:$port",job=~"$job"}) (((count(count(node_cpu_seconds_total{commonFilters}) by (cpu)))-avg(sum by (mode) (rate(node_cpu_seconds_total{mode='idle',commonFilters}[5m]))))*100)/count(count(node_cpu_seconds_total{commonFilters}) by (cpu))`, `  (
      (
          (
            count (
              count by (cpu) (
                node_cpu_seconds_total{instance=~"$node:$port", job=~"$job"}
              )
            )
          )
        -
          (
            avg (
              sum by (mode) (
                rate (
                  node_cpu_seconds_total{mode="idle", instance=~"$node:$port", job=~"$job"}[5m]
                )
              )
            )
          )
      )
    *
      100
  )
/
  (
    count (
      count by (cpu) (
        node_cpu_seconds_total{instance=~"$node:$port", job=~"$job"}
      )
    )
  )`)
	another(`WITH (commonFilters = {instance=~"$node:$port",job=~"$job"}, cpuCount = count(count(node_cpu_seconds_total{commonFilters}) by (cpu)))((cpuCount-avg(sum by (mode) (rate(node_cpu_seconds_total{mode='idle',commonFilters}[5m]))))*100) / cpuCount`, `  (
      (
          (
            count (
              count by (cpu) (
                node_cpu_seconds_total{instance=~"$node:$port", job=~"$job"}
              )
            )
          )
        -
          (
            avg (
              sum by (mode) (
                rate (
                  node_cpu_seconds_total{mode="idle", instance=~"$node:$port", job=~"$job"}[5m]
                )
              )
            )
          )
      )
    *
      100
  )
/
  (
    count (
      count by (cpu) (
        node_cpu_seconds_total{instance=~"$node:$port", job=~"$job"}
      )
    )
  )`)
	another(`rate(remote_storage_samples_in_total{cluster=~"$cluster", instance=~"$instance"}[5m])- ignoring(remote_name, url) group_right(instance) (rate(remote_storage_succeeded_samples_total{cluster=~"$cluster", instance=~"$instance"}[5m]) or rate(remote_storage_samples_total{cluster=~"$cluster", instance=~"$instance"}[5m]))- (rate(remote_storage_dropped_samples_total{cluster=~"$cluster", instance=~"$instance"}[5m]) or rate(remote_storage_samples_dropped_total{cluster=~"$cluster", instance=~"$instance"}[5m]))`, `  (
      rate (
        remote_storage_samples_in_total{cluster=~"$cluster", instance=~"$instance"}[5m]
      )
    - ignoring (remote_name, url) group_right (instance)
      (
          rate (
            remote_storage_succeeded_samples_total{cluster=~"$cluster", instance=~"$instance"}[5m]
          )
        or
          rate (
            remote_storage_samples_total{cluster=~"$cluster", instance=~"$instance"}[5m]
          )
      )
  )
-
  (
      rate (
        remote_storage_dropped_samples_total{cluster=~"$cluster", instance=~"$instance"}[5m]
      )
    or
      rate (
        remote_storage_samples_dropped_total{cluster=~"$cluster", instance=~"$instance"}[5m]
      )
  )`)
}
