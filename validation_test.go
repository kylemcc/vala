package vala

import (
	"testing"
)

func TestPanicIsIssued(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.FailNow()
		}
	}()

	BeginValidation().Validate(
		Equals("foo", "bar", "foo"),
	).CheckAndPanic()
}

func TestErrorReturns(t *testing.T) {

	err := BeginValidation().Validate(
		Equals("foo", "bar", "foo"),
	).Check()

	if err == nil {
		t.FailNow()
	}
}

func TestSetError(t *testing.T) {
	var returnErr error
	defer func() {
		if r := recover(); r == nil {
			t.FailNow()
		}

		if returnErr == nil {
			t.FailNow()
		}
	}()

	BeginValidation().Validate(
		Equals("foo", "bar", "foo"),
	).CheckSetErrorAndPanic(&returnErr)

	t.Error("We should have never reached this.")
	t.FailNow()
}

func TestNot(t *testing.T) {

	err := BeginValidation().Validate(
		Not(Equals("foo", "bar", "foo")),
	).Check()

	if err != nil {
		t.Error("Received an unexpected error.")
		t.FailNow()
	}

	err = BeginValidation().Validate(
		Not(Equals("foo", "foo", "varName")),
	).Check()

	if err == nil {
		t.Error("Expected an error.")
		t.Fail()
	}
}

func TestEquals(t *testing.T) {

	err := BeginValidation().Validate(
		Equals("foo", "bar", "foo"),
	).Check()

	if err == nil {
		t.FailNow()
	}

	err = BeginValidation().Validate(
		Equals("foo", "foo", "foo"),
	).Check()

	if err != nil {
		t.FailNow()
	}
}

func TestIsNil(t *testing.T) {

	err := BeginValidation().Validate(
		IsNotNil("foo", "foo"),
		IsNotNil(t, "t"),
	).Check()

	if err != nil {
		t.Error("Received an unexpected error.")
		t.FailNow()
	}

	var nilSlice []string

	err = BeginValidation().Validate(
		Not(IsNotNil(nil, "foo")),
		Not(IsNotNil(nilSlice, "nilSlice")),
	).Check()

	if err != nil {
		t.Errorf("Received an unexpected error: %v", err)
		t.FailNow()
	}
}

func TestHasLen(t *testing.T) {

	err := BeginValidation().Validate(
		HasLen([]int{1, 2}, 2, "tmpA"),
		HasLen([]int{}, 0, "tmpB"),
		HasLen("1", 1, "tmpC"),
	).Check()

	if err != nil {
		t.Errorf("Received an unexpected error: %v", err)
		t.FailNow()
	}

	err = BeginValidation().Validate(
		HasLen("", 1, "tmpC"),
	).Check()

	if err == nil {
		t.Errorf("Expected an error.")
		t.FailNow()
	}
}

func TestGreaterThan(t *testing.T) {

	err := BeginValidation().Validate(
		GreaterThan(1, 0, "tmpA"),
	).Check()

	if err != nil {
		t.Errorf("Received an unexpected error: %v", err)
		t.FailNow()
	}

	err = BeginValidation().Validate(
		GreaterThan(0, 1, "tmpC"),
	).Check()

	if err == nil {
		t.Errorf("Expected an error.")
		t.FailNow()
	}
}

func TestStringNotEmpty(t *testing.T) {

	err := BeginValidation().Validate(
		StringNotEmpty("", "tmpA"),
	).Check()

	if err == nil {
		t.Errorf("Expected an error.")
		t.FailNow()
	}
}

func TestOr(t *testing.T) {
	cases := []struct {
		checkers []Checker
		pass     bool
	}{
		{[]Checker{HasLen("short", 10, "test"), StringNotEmpty("", "test")}, false},
		{[]Checker{HasLen("1234567890", 10, "test"), StringNotEmpty("", "test")}, true},
		{[]Checker{HasLen("short", 10, "test"), StringNotEmpty("foo", "test")}, true},
	}

	for i, c := range cases {
		err := BeginValidation().Validate(Or(c.checkers...)).Check()
		if (err == nil) != c.pass {
			t.Errorf("Or checker #%d returned incorrect result: %v\n", i, err)
		}
	}
}

func TestAnd(t *testing.T) {
	cases := []struct {
		checkers []Checker
		pass     bool
	}{
		{[]Checker{HasLen("short", 10, "test"), StringNotEmpty("", "test")}, false},
		{[]Checker{HasLen("1234567890", 10, "test"), StringNotEmpty("test", "test")}, true},
		{[]Checker{HasLen("short", 10, "test"), StringNotEmpty("foo", "test")}, false},
	}

	for i, c := range cases {
		err := BeginValidation().Validate(And(c.checkers...)).Check()
		if (err == nil) != c.pass {
			t.Errorf("And checker #%d returned incorrect result: %v\n", i, err)
		}
	}
}
