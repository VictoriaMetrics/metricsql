package metricsql

import "testing"

func TestPrettier(t *testing.T) {
	check := func(s, expected string) {
		t.Helper()

		got, err := Prettier(s, false)
		if err != nil {
			t.Fatalf("unexpected error when parsing %q: %s", s, err)
		}

		if expected != got {
			t.Fatalf("string not prettified;\ngot\n%q\nwant\n%q", got, expected)
		}
	}

	// check(`with () x`, `x`)
	// check(`with (x=1,) x`, `1`)
	// another(`with (x = m offset 5h) x + x`, `m offset 5h + m offset 5h`)
	// another(`with (x = m offset 5i) x + x`, `m offset 5i + m offset 5i`)
	// another(`with (foo = bar{x="x"}) 1`, `1`)
	// another(`with (foo = bar{x="x"}) "x"`, `"x"`)
	// another(`with (f="x") f`, `"x"`)
	// another(`with (foo = bar{x="x"}) x{x="y"}`, `x{x="y"}`)
	// another(`with (foo = bar{x="x"}) 1+1`, `2`)
	// another(`with (foo = bar{x="x"}) f()`, `f()`)
	// another(`with (foo = bar{x="x"}) sum(x)`, `sum(x)`)
	// another(`with (foo = bar{x="x"}) baz{foo="bar"}`, `baz{foo="bar"}`)
	// another(`with (foo = bar) baz`, `baz`)
	// another(`with (foo = bar) foo + foo{a="b"}`, `bar + bar{a="b"}`)
	// another(`with (foo = bar, bar=baz + f()) test`, `test`)
	// another(`with (ct={job="test"}) a{ct} + ct() + f({ct="x"})`, `(a{job="test"} + {job="test"}) + f({ct="x"})`)
	// another(`with (ct={job="test", i="bar"}) ct + {ct, x="d"} + foo{ct, ct} + ctx(1)`,
	// 	`(({job="test", i="bar"} + {job="test", i="bar", x="d"}) + foo{job="test", i="bar"}) + ctx(1)`)
	// another(`with (foo = bar) {__name__=~"foo"}`, `{__name__=~"foo"}`)
	// another(`with (foo = bar) foo{__name__="foo"}`, `bar`)
	// another(`with (foo = bar) {__name__="foo", x="y"}`, `bar{x="y"}`)
	// another(`with (foo(bar) = {__name__!="bar"}) foo(x)`, `{__name__!="bar"}`)
	// another(`with (foo(bar) = bar{__name__="bar"}) foo(x)`, `x`)
	// another(`with (foo\-bar(baz) = baz + baz) foo\-bar((x,y))`, `(x, y) + (x, y)`)
	// another(`with (foo\-bar(baz) = baz + baz) foo\-bar(x*y)`, `(x * y) + (x * y)`)
	// another(`with (foo\-bar(baz) = baz + baz) foo\-bar(x\*y)`, `x\*y + x\*y`)
	// another(`with (foo\-bar(b\ az) = b\ az + b\ az) foo\-bar(x\*y)`, `x\*y + x\*y`)
	// // override ttf to something new.
	// another(`with (ttf = a) ttf + b`, `a + b`)
	// override ttf to ru

	// 	check(`histogram_quantile(0.9, rate(instance_cpu_time_seconds{app="lion", proc="web",job="cluster-manager"}[5m]))`, `histogram_quantile (
	//   0.9,
	//   rate (
	//     instance_cpu_time_seconds{app="lion", proc="web", job="cluster-manager"}[5m:]
	//   )
	// )`)
	// 	check(`with () x`, `x`)
	// 	check(`with (x=1,) x`, `1`)
	check(
		`(node_timex_offset_seconds > 0.05 and deriv(node_timex_offset_seconds[5m]) >= 0) or (node_timex_offset_seconds < -0.05 and deriv(node_timex_offset_seconds[5m]) <= 0)`,
		`  (
        node_timex_offset_seconds
      >
        0.05
    and
        deriv(
          node_timex_offset_seconds[5m]
        )
      >=
        0
  )
or
  (
        node_timex_offset_seconds
      <
        -0.05
    and
        deriv(
          node_timex_offset_seconds[5m]
        )
      <=
        0
  )`)
}
