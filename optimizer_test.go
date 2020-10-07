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
}
