package binaryop

import "testing"

func TestAnd(t *testing.T) {
	f := func(left, right, expectedResult float64) {
		t.Helper()
		res := And(left, right)
		if res != expectedResult {
			t.Fatalf("unexpected result: %1.f, get: %1.f", res, expectedResult)
		}
	}
	f(2, -1, 1)
	f(1, 2, 1)
	f(0, 1, 0)
	f(1, 0, 0)
	f(2, 0, 0)
}

func TestOr(t *testing.T) {
	f := func(left, right, expectedResult float64) {
		t.Helper()
		res := Or(left, right)
		if res != expectedResult {
			t.Fatalf("unexpected result: %1.f, get: %1.f", res, expectedResult)
		}
	}
	f(2, -1, 1)
	f(1, 2, 1)
	f(0, -1, 1)
	f(1, 0, 1)
	f(0, 0, 0)
}
