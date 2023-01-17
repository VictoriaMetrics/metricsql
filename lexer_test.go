package metricsql

import (
	"math"
	"reflect"
	"testing"
)

func TestScanNumMultiplier(t *testing.T) {
	f := func(s string, lenExpected int) {
		t.Helper()
		sLen := scanNumMultiplier(s)
		if sLen != lenExpected {
			t.Fatalf("unexpected len returned from scanNumMultiplier(%q); got %d; want %d", s, sLen, lenExpected)
		}
	}
	f("", 0)
	f("foo", 0)
	f("k", 1)
	f("KB", 2)
	f("Ki", 2)
	f("kiB", 3)
	f("M", 1)
	f("Mb", 2)
	f("mi", 2)
	f("MiB", 3)
	f("g", 1)
	f("GB", 2)
	f("GI", 2)
	f("GIB", 3)
	f("t", 1)
	f("tB", 2)
	f("tI", 2)
	f("tIb", 3)

	f("Gb   ", 2)
	f("tIb + 5", 3)
}

func TestScanPositiveNumberSuccess(t *testing.T) {
	f := func(s, nsExpected string) {
		t.Helper()
		ns, err := scanPositiveNumber(s)
		if err != nil {
			t.Fatalf("unexpected error in scanPositiveNumber(%q): %s", s, err)
		}
		if ns != nsExpected {
			t.Fatalf("unexpected number scanned from %q; got %q; want %q", s, ns, nsExpected)
		}
	}
	f("123", "123")
	f("123+5", "123")
	f("1.23 ", "1.23")
	f("12e5", "12e5")
	f("1.3E-3/5", "1.3E-3")
	f("234.", "234.")
	f("234. + foo", "234.")
	f("0xfe", "0xfe")
	f("0b0110", "0b0110")
	f("0O765", "0O765")
	f("0765", "0765")
	f("2k*34", "2k")
	f("2.3Kb / 43", "2.3Kb")
	f("3ki", "3ki")
	f("4.5Kib", "4.5Kib")
	f("2m", "2m")
	f("2.3Mb", "2.3Mb")
	f("3Mi", "3Mi")
	f("4.5mib", "4.5mib")
	f("2G", "2G")
	f("2.3gB", "2.3gB")
	f("3gI", "3gI")
	f("4.5GiB / foo", "4.5GiB")
	f("2T", "2T")
	f("2.3tb", "2.3tb")
	f("3tI", "3tI")
	f("4.5TIB   ", "4.5TIB")
}

func TestScanPositiveNumberFailure(t *testing.T) {
	f := func(s string) {
		t.Helper()
		ns, err := scanPositiveNumber(s)
		if err == nil {
			t.Fatalf("expecting non-nil error in scanPositiveNumber(%q); got result %q", s, ns)
		}
	}
	f("")
	f("foobar")
	f("123e")
	f("1233Ebc")
	f("12.34E+abc")
	f("12.34e-")
}

func TestParsePositiveNumberSuccess(t *testing.T) {
	f := func(s string, vExpected float64) {
		t.Helper()
		v, err := parsePositiveNumber(s)
		if err != nil {
			t.Fatalf("unexpected error in parsePositiveNumber(%q): %s", s, err)
		}
		if math.IsNaN(v) {
			if !math.IsNaN(vExpected) {
				t.Fatalf("unexpected value returned from parsePositiveNumber(%q); got %v; want %v", s, v, vExpected)
			}
		} else if v != vExpected {
			t.Fatalf("unexpected value returned from parsePositiveNumber(%q); got %v; want %v", s, v, vExpected)
		}
	}
	f("123", 123)
	f("1.23", 1.23)
	f("12e5", 12e5)
	f("1.3E-3", 1.3e-3)
	f("234.", 234)
	f("Inf", math.Inf(1))
	f("NaN", math.NaN())
	f("0xfe", 0xfe)
	f("0b0110", 0b0110)
	f("0O765", 0o765)
	f("0765", 0765)
	f("2k", 2*1000)
	f("2.3Kb", 2.3*1000)
	f("3ki", 3*1024)
	f("4.5Kib", 4.5*1024)
	f("2m", 2*1000*1000)
	f("2.3Mb", 2.3*1000*1000)
	f("3Mi", 3*1024*1024)
	f("4.5mib", 4.5*1024*1024)
	f("2G", 2*1000*1000*1000)
	f("2.3gB", 2.3*1000*1000*1000)
	f("3gI", 3*1024*1024*1024)
	f("4.5GiB", 4.5*1024*1024*1024)
	f("2T", 2*1000*1000*1000*1000)
	f("2.3tb", 2.3*1000*1000*1000*1000)
	f("3tI", 3*1024*1024*1024*1024)
	f("4.5TIB", 4.5*1024*1024*1024*1024)
}

func TestParsePositiveNumberFailure(t *testing.T) {
	f := func(s string) {
		t.Helper()
		v, err := parsePositiveNumber(s)
		if err == nil {
			t.Fatalf("expecting non-nil error in parsePositiveNumber(%q); got result %v", s, v)
		}
	}
	f("")
	f("0xqwert")
	f("foobar")
	f("234.foobar")
	f("123e")
	f("1233Ebc")
	f("12.34E+abc")
	f("12.34e-")
	f("12.weKB")
}

func TestIsSpecialIntegerPrefix(t *testing.T) {
	f := func(s string, resultExpected bool) {
		t.Helper()
		result := isSpecialIntegerPrefix(s)
		if result != resultExpected {
			t.Fatalf("unexpected result for isSpecialIntegerPrefix(%q); got %v; want %v", s, result, resultExpected)
		}
	}
	f("", false)
	f("1", false)
	f("0", false)

	// octal numbers
	f("03", true)
	f("0o1", true)
	f("0O12", true)

	// binary numbers
	f("0b1110", true)
	f("0B0", true)

	// hex number
	f("0x1ffa", true)
	f("0X4", true)
}

func TestUnescapeIdent(t *testing.T) {
	f := func(s, resultExpected string) {
		t.Helper()
		result := unescapeIdent(s)
		if result != resultExpected {
			t.Fatalf("unexpected result for unescapeIdent(%q); got %q; want %q", s, result, resultExpected)
		}
	}
	f("", "")
	f("a", "a")
	f("\\", `\`)
	f(`\\`, `\`)
	f(`\foo\-bar`, `foo-bar`)
	f(`a\\\\b\"c\d`, `a\\b"cd`)
	f(`foo.bar:baz_123`, `foo.bar:baz_123`)
	f(`foo\ bar`, `foo bar`)
	f(`\x21`, `!`)
	f(`\X21`, `!`)
	f(`\x7Dfoo\x2Fbar\-\xqw\x`, "}foo/bar-\\xqw\\x")
	f(`\п\р\и\в\е\т123`, "привет123")
	f(`123`, `123`)
	f(`\123`, `123`)
	f(`привет\-\foo`, "привет-foo")
	f(`\u0965`, "\u0965")
	f(`\U0965`, "\u0965")
	f(`\u202c`, "\u202c")
	f(`\U202ca`, "\u202ca")
}

func TestAppendEscapedIdent(t *testing.T) {
	f := func(s, resultExpected string) {
		t.Helper()
		result := appendEscapedIdent(nil, s)
		if string(result) != resultExpected {
			t.Fatalf("unexpected result for appendEscapedIdent(%q); got %q; want %q", s, result, resultExpected)
		}
	}
	f(`a`, `a`)
	f(`a.b:c_23`, `a.b:c_23`)
	f(`a b-cd+dd\`, `a\ b\-cd\+dd\\`)
	f("a\x1E\x20\x7e", `a\x1e\ \~`)
	f("\x2e\x2e", `\..`)
	f("123", `\123`)
	f("+43.6", `\+43.6`)
	f("привет123(a-b)", `привет123\(a\-b\)`)
	f("\u0965", `\॥`)
	f("\u202c", `\u202c`)
}

func TestScanIdent(t *testing.T) {
	f := func(s, resultExpected string) {
		t.Helper()
		result := scanIdent(s)
		if result != resultExpected {
			t.Fatalf("unexpected result for scanIdent(%q): got %q; want %q", s, result, resultExpected)
		}
	}
	f("a", "a")
	f("foo.bar:baz_123", "foo.bar:baz_123")
	f("a+b", "a")
	f("foo()", "foo")
	f(`a\-b+c`, `a\-b`)
	f(`a\ b\\\ c\`, `a\ b\\\ c`)
	f(`\п\р\и\в\е\т123`, `\п\р\и\в\е\т123`)
	f(`привет123!foo`, `привет123`)
	f(`\1fooЫ+bar`, `\1fooЫ`)
	f(`\u7834*аа`, `\u7834`)
	f(`\U7834*аа`, `\U7834`)
	f(`\x7834*аа`, `\x7834`)
	f(`\X7834*аа`, `\X7834`)
	f(`a\x+b`, `a`)
	f(`a\x1+b`, `a`)
	f(`a\x12+b`, `a\x12`)
	f(`a\u+b`, `a`)
	f(`a\u1+b`, `a`)
	f(`a\u12+b`, `a`)
	f(`a\u123+b`, `a`)
	f(`a\u1234+b`, `a\u1234`)
	f("a\\\u202c", `a`)
}

func TestLexerNextPrev(t *testing.T) {
	var lex lexer
	lex.Init("foo bar baz")
	if lex.Token != "" {
		t.Fatalf("unexpected token got: %q; want %q", lex.Token, "")
	}
	if err := lex.Next(); err != nil {
		t.Fatalf("unexpeted error: %s", err)
	}
	if lex.Token != "foo" {
		t.Fatalf("unexpected token got: %q; want %q", lex.Token, "foo")
	}

	// Rewind before the first item.
	lex.Prev()
	if lex.Token != "" {
		t.Fatalf("unexpected token got: %q; want %q", lex.Token, "")
	}
	if err := lex.Next(); err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	if lex.Token != "foo" {
		t.Fatalf("unexpected token got: %q; want %q", lex.Token, "foo")
	}
	if err := lex.Next(); err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	if lex.Token != "bar" {
		t.Fatalf("unexpected token got: %q; want %q", lex.Token, "bar")
	}

	// Rewind to the first item.
	lex.Prev()
	if lex.Token != "foo" {
		t.Fatalf("unexpected token got: %q; want %q", lex.Token, "foo")
	}
	if err := lex.Next(); err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	if lex.Token != "bar" {
		t.Fatalf("unexpected token got: %q; want %q", lex.Token, "bar")
	}
	if err := lex.Next(); err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	if lex.Token != "baz" {
		t.Fatalf("unexpected token got: %q; want %q", lex.Token, "baz")
	}

	// Go beyond the token stream.
	if err := lex.Next(); err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	if lex.Token != "" {
		t.Fatalf("unexpected token got: %q; want %q", lex.Token, "")
	}
	if !isEOF(lex.Token) {
		t.Fatalf("expecting eof")
	}
	lex.Prev()
	if lex.Token != "baz" {
		t.Fatalf("unexpected token got: %q; want %q", lex.Token, "baz")
	}

	// Go multiple times lex.Next() beyond token stream.
	if err := lex.Next(); err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	if lex.Token != "" {
		t.Fatalf("unexpected token got: %q; want %q", lex.Token, "")
	}
	if !isEOF(lex.Token) {
		t.Fatalf("expecting eof")
	}
	if err := lex.Next(); err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	if lex.Token != "" {
		t.Fatalf("unexpected token got: %q; want %q", lex.Token, "")
	}
	if !isEOF(lex.Token) {
		t.Fatalf("expecting eof")
	}
	lex.Prev()
	if lex.Token != "" {
		t.Fatalf("unexpected token got: %q; want %q", lex.Token, "")
	}
	if !isEOF(lex.Token) {
		t.Fatalf("expecting eof")
	}
}

func TestLexerSuccess(t *testing.T) {
	var s string
	var expectedTokens []string

	// An empty string
	s = ""
	expectedTokens = nil
	testLexerSuccess(t, s, expectedTokens)

	// String with whitespace
	s = "  \n\t\r "
	expectedTokens = nil
	testLexerSuccess(t, s, expectedTokens)

	// Just metric name
	s = "metric"
	expectedTokens = []string{"metric"}
	testLexerSuccess(t, s, expectedTokens)

	// Metric name with spec chars
	s = ":foo.bar_"
	expectedTokens = []string{":foo.bar_"}
	testLexerSuccess(t, s, expectedTokens)

	// Metric name with window
	s = "metric[5m]  "
	expectedTokens = []string{"metric", "[", "5m", "]"}
	testLexerSuccess(t, s, expectedTokens)

	// Metric name with tag filters
	s = `  metric:12.34{a="foo", b != "bar", c=~ "x.+y", d !~ "zzz"}`
	expectedTokens = []string{`metric:12.34`, `{`, `a`, `=`, `"foo"`, `,`, `b`, `!=`, `"bar"`, `,`, `c`, `=~`, `"x.+y"`, `,`, `d`, `!~`, `"zzz"`, `}`}
	testLexerSuccess(t, s, expectedTokens)

	// Metric name with offset
	s = `   metric offset 10d   `
	expectedTokens = []string{`metric`, `offset`, `10d`}
	testLexerSuccess(t, s, expectedTokens)

	// Func call
	s = `sum  (  metric{x="y"  }  [5m] offset 10h)`
	expectedTokens = []string{`sum`, `(`, `metric`, `{`, `x`, `=`, `"y"`, `}`, `[`, `5m`, `]`, `offset`, `10h`, `)`}
	testLexerSuccess(t, s, expectedTokens)

	// Binary op
	s = `a+b or c % d and e unless f`
	expectedTokens = []string{`a`, `+`, `b`, `or`, `c`, `%`, `d`, `and`, `e`, `unless`, `f`}
	testLexerSuccess(t, s, expectedTokens)

	// Numbers
	s = `3+1.2-.23+4.5e5-78e-6+1.24e+45-NaN+Inf`
	expectedTokens = []string{`3`, `+`, `1.2`, `-`, `.23`, `+`, `4.5e5`, `-`, `78e-6`, `+`, `1.24e+45`, `-`, `NaN`, `+`, `Inf`}
	testLexerSuccess(t, s, expectedTokens)

	s = `12.34 * 0X34 + 0b11 + 0O77`
	expectedTokens = []string{`12.34`, `*`, `0X34`, `+`, `0b11`, `+`, `0O77`}
	testLexerSuccess(t, s, expectedTokens)

	// Strings
	s = `""''` + "``" + `"\\"  '\\'  "\"" '\''"\\\"\\"`
	expectedTokens = []string{`""`, `''`, "``", `"\\"`, `'\\'`, `"\""`, `'\''`, `"\\\"\\"`}
	testLexerSuccess(t, s, expectedTokens)

	// Various durations
	s = `m offset 123h`
	expectedTokens = []string{`m`, `offset`, `123h`}
	testLexerSuccess(t, s, expectedTokens)

	s = `m offset -1.23w-5h34.5m - 123`
	expectedTokens = []string{`m`, `offset`, `-`, `1.23w-5h34.5m`, `-`, `123`}
	testLexerSuccess(t, s, expectedTokens)

	s = "   `foo\\\\\\`бар`  "
	expectedTokens = []string{"`foo\\\\\\`бар`"}
	testLexerSuccess(t, s, expectedTokens)

	s = `# comment # sdf
		foobar # comment
		baz
		# yet another comment`
	expectedTokens = []string{"foobar", "baz"}
	testLexerSuccess(t, s, expectedTokens)
}

func testLexerSuccess(t *testing.T, s string, expectedTokens []string) {
	t.Helper()

	var lex lexer
	lex.Init(s)

	var tokens []string
	for {
		if err := lex.Next(); err != nil {
			t.Fatalf("unexpected error: %s", err)
		}
		if isEOF(lex.Token) {
			break
		}
		tokens = append(tokens, lex.Token)
	}
	if !reflect.DeepEqual(tokens, expectedTokens) {
		t.Fatalf("unexected tokens\ngot\n%q\nwant\n%q", tokens, expectedTokens)
	}
}

func TestLexerError(t *testing.T) {
	// Invalid identifier
	testLexerError(t, ".foo")

	// Incomplete string
	testLexerError(t, `"foobar`)
	testLexerError(t, `'`)
	testLexerError(t, "`")

	// Invalid numbers
	testLexerError(t, `.`)
	testLexerError(t, `12e`)
	testLexerError(t, `1.2e`)
	testLexerError(t, `1.2E+`)
	testLexerError(t, `1.2E-`)
}

func testLexerError(t *testing.T, s string) {
	t.Helper()

	var lex lexer
	lex.Init(s)
	for {
		if err := lex.Next(); err != nil {
			// Expected error
			break
		}
		if isEOF(lex.Token) {
			t.Fatalf("expecting error during parse")
		}
	}

	// Try calling Next again. It must return error.
	if err := lex.Next(); err == nil {
		t.Fatalf("expecting non-nil error")
	}
}

func TestPositiveDurationSuccess(t *testing.T) {
	f := func(s string, step, expectedD int64) {
		t.Helper()
		d, err := PositiveDurationValue(s, step)
		if err != nil {
			t.Fatalf("unexpected error: %s", err)
		}
		if d != expectedD {
			t.Fatalf("unexpected duration; got %d; want %d", d, expectedD)
		}
	}

	// Integer durations
	f("123ms", 42, 123)
	f("123s", 42, 123*1000)
	f("123m", 42, 123*60*1000)
	f("1h", 42, 1*60*60*1000)
	f("2d", 42, 2*24*60*60*1000)
	f("3w", 42, 3*7*24*60*60*1000)
	f("4y", 42, 4*365*24*60*60*1000)
	f("1i", 42*1000, 42*1000)
	f("3i", 42, 3*42)

	// Float durations
	f("123.45ms", 42, 123)
	f("0.234s", 42, 234)
	f("1.5s", 42, 1.5*1000)
	f("1.5m", 42, 1.5*60*1000)
	f("1.2h", 42, 1.2*60*60*1000)
	f("1.1d", 42, 1.1*24*60*60*1000)
	f("1.1w", 42, 1.1*7*24*60*60*1000)
	f("1.3y", 42, 1.3*365*24*60*60*1000)
	f("0.1i", 12340, 0.1*12340)

	// Floating-point durations without suffix.
	f("123", 45, 123000)
	f("1.23", 45, 1230)
	f("0.56", 12, 560)
	f(".523e2", 21, 52300)

	// Duration suffixes in mixed case.
	f("1Ms", 45, 1)
	f("1mS", 45, 1)
	f("1H", 45, 1*60*60*1000)
	f("1D", 45, 1*24*60*60*1000)
	f("1Y", 45, 1*365*24*60*60*1000)
}

func TestPositiveDurationError(t *testing.T) {
	f := func(s string) {
		t.Helper()
		d, err := PositiveDurationValue(s, 42)
		if err == nil {
			t.Fatalf("expecting non-nil error for duration %q", s)
		}
		if d != 0 {
			t.Fatalf("expecting zero duration; got %d", d)
		}
	}
	f("")
	f("foo")
	f("m")
	f("1.23mm")
	f("123q")
	f("-123s")
	f("1.23.4434s")
	f("1mi")
	f("1mb")

	// Too big duration
	f("10000000000y")

	// Uppercase M isn't a duration, but a 1e6 multiplier.
	// See https://github.com/VictoriaMetrics/VictoriaMetrics/issues/3664
	f("1M")
}

func TestDurationSuccess(t *testing.T) {
	f := func(s string, step, expectedD int64) {
		t.Helper()
		d, err := DurationValue(s, step)
		if err != nil {
			t.Fatalf("unexpected error: %s", err)
		}
		if d != expectedD {
			t.Fatalf("unexpected duration; got %d; want %d", d, expectedD)
		}
	}

	// Integer durations
	f("123ms", 42, 123)
	f("-123ms", 42, -123)
	f("123s", 42, 123*1000)
	f("-123s", 42, -123*1000)
	f("123m", 42, 123*60*1000)
	f("1h", 42, 1*60*60*1000)
	f("2d", 42, 2*24*60*60*1000)
	f("3w", 42, 3*7*24*60*60*1000)
	f("4y", 42, 4*365*24*60*60*1000)
	f("1i", 42*1000, 42*1000)
	f("3i", 42, 3*42)
	f("-3i", 42, -3*42)
	f("1m34s24ms", 42, 94024)
	f("1m-34s24ms", 42, 25976)
	f("-1m34s24ms", 42, -94024)
	f("-1m-34s24ms", 42, -94024)

	// Float durations
	f("34.54ms", 42, 34)
	f("-34.34ms", 42, -34)
	f("0.234s", 42, 234)
	f("-0.234s", 42, -234)
	f("1.5s", 42, 1.5*1000)
	f("1.5m", 42, 1.5*60*1000)
	f("1.2h", 42, 1.2*60*60*1000)
	f("1.1d", 42, 1.1*24*60*60*1000)
	f("1.1w", 42, 1.1*7*24*60*60*1000)
	f("1.3y", 42, 1.3*365*24*60*60*1000)
	f("-1.3y", 42, -1.3*365*24*60*60*1000)
	f("0.1i", 12340, 0.1*12340)
	f("1.5m3.4s2.4ms", 42, 93402)
	f("-1.5m3.4s2.4ms", 42, -93402)

	// Floating-point durations without suffix.
	f("123", 45, 123000)
	f("1.23", 45, 1230)
	f("-0.56", 12, -560)
	f("-.523e2", 21, -52300)

	// Duration suffix in mixed case.
	f("-1Ms", 10, -1)
	f("-2.5mS", 10, -2)
	f("-1mS", 10, -1)
	f("-1H", 10, -1*60*60*1000)
	f("-3.H", 10, -3*60*60*1000)
	f("1D", 10, 1*24*60*60*1000)
	f("-.1Y", 10, -0.1*365*24*60*60*1000)
}

func TestDurationError(t *testing.T) {
	f := func(s string) {
		t.Helper()
		d, err := DurationValue(s, 42)
		if err == nil {
			t.Fatalf("expecting non-nil error for duration %q", s)
		}
		if d != 0 {
			t.Fatalf("expecting zero duration; got %d", d)
		}
	}
	f("")
	f("foo")
	f("m")
	f("1.23mm")
	f("123q")
	f("-123q")
	f("-5.3mb")
	f("-5.3mi")

	// M isn't a duration, but a 1e6 multiplier.
	// See https://github.com/VictoriaMetrics/VictoriaMetrics/issues/3664
	f("-5.3M")
}
