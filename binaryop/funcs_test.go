package binaryop

import (
	"testing"
)

func TestAnd(t *testing.T) {
	f := func(left, right float64, expectedResult bool) {
		t.Helper()
		res := And(left, right)
		if res != expectedResult {
			t.Fatalf("unexpected result: %t, get: %t", res, expectedResult)
		}
	}
	f(2, -1, true)
	f(1, 2, true)
	f(0, 1, false)
	f(1, 0, false)
	f(2, 0, false)
}

func TestOr(t *testing.T) {
	f := func(left, right float64, expectedResult bool) {
		t.Helper()
		res := Or(left, right)
		if res != expectedResult {
			t.Fatalf("unexpected result: %t, get: %t", res, expectedResult)
		}
	}
	f(2, -1, true)
	f(1, 2, true)
	f(0, -1, true)
	f(1, 0, true)
	f(0, 0, false)
}
