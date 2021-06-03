/*
Package vala is a simple, extensible, library to make argument
validation in Go palatable.

This package uses the fluent programming style to provide
simultaneously more robust and more terse parameter validation.

	BeginValidation().Validate(
		IsNotNil(a, "a"),
		IsNotNil(b, "b"),
		IsNotNil(c, "c"),
	).CheckAndPanic().Validate( // Panic will occur here if a, b, or c are nil.
		HasLen(a.Items, 50, "a.Items"),
		GreaterThan(b.UserCount, 0, "b.UserCount"),
		Equals(c.Name, "Vala", "c.name"),
		Not(Equals(c.FriendlyName, "Foo", "c.FriendlyName")),
	).Check()

Notice how checks can be tiered.

Vala is also extensible. As long as a function conforms to the Checker
specification, you can pass it into the Validate method:

	func ReportFitsRepository(report *Report, repository *Repository) Checker {
		return func() (passes bool, err error) {

			err = fmt.Errorf("A %s report does not belong in a %s repository.", report.Type, repository.Type)
			passes = (repository.Type == report.Type)
			return passes, err
		}
	}

	func AuthorCanUpload(authorName string, repository *Repository) Checker {
		return func() (passes bool, err error) {
			err = fmt.Errorf("%s does not have access to this repository.", authorName)
			passes = !repository.AuthorCanUpload(authorName)
			return passes, err
		}
	}

	func AuthorIsCollaborator(authorName string, report *Report) Checker {
		return func() (passes bool, err error) {

			err = fmt.Errorf("The given author was not one of the collaborators for this report.")
			for _, collaboratorName := range report.Collaborators() {
				if collaboratorName == authorName {
					passes = true
					break
				}
			}

			return passes, err
		}
	}

	func HandleReport(authorName string, report *Report, repository *Repository) {

		BeginValidation().Validate(
			AuthorIsCollaborator(authorName, report),
			AuthorCanUpload(authorName, repository),
			ReportFitsRepository(report, repository),
		).CheckAndPanic()
	}
*/
package vala

import (
	"fmt"
	"reflect"
	"strings"
)

func validationFactory(numErrors int) *Validation {
	return &Validation{make([]error, 0, numErrors)}
}

// Validation contains all the errors from performing Checkers, and is
// the fluent type off which all Validation methods hang.
type Validation struct {
	Errors []error
}

// BeginValidation begins a validation check.
func BeginValidation() *Validation {
	return nil
}

// Check aggregates all checker errors into a single error and returns
// this error.
func (val *Validation) Check() error {
	if val == nil || len(val.Errors) <= 0 {
		return nil
	}

	return val.constructErrorMessage()
}

// CheckAndPanic aggregates all checker errors into a single error and
// panics with this error.
func (val *Validation) CheckAndPanic() *Validation {
	if val == nil || len(val.Errors) <= 0 {
		return val
	}

	panic(val.constructErrorMessage())
}

// CheckSetErrorAndPanic aggregates any Errors produced by the
// Checkers into a single error, and sets the address of retError to
// this, and panics. The canonical use-case of this is to pass in the
// address of an error you would like to return, and then to catch the
// panic and do nothing.
func (val *Validation) CheckSetErrorAndPanic(retError *error) *Validation {
	if val == nil || len(val.Errors) <= 0 {
		return val
	}

	*retError = val.constructErrorMessage()
	panic(*retError)
}

// Validate runs all of the checkers passed in and collects errors
// into an internal collection. To take action on these errors, call
// one of the Check* methods.
func (val *Validation) Validate(checkers ...Checker) *Validation {

	for _, checker := range checkers {
		if pass, err := checker(); !pass {
			if val == nil {
				val = validationFactory(1)
			}

			val.Errors = append(val.Errors, err)
		}
	}

	return val
}

func (val *Validation) constructErrorMessage() error {
	if len(val.Errors) == 1 {
		return fmt.Errorf("parameter validation failed: %s", val.Errors[0])
	}

	errorStrings := make([]string, 0, len(val.Errors))
	for _, e := range val.Errors {
		errorStrings = append(errorStrings, e.Error())
	}
	return fmt.Errorf(
		"parameter validation failed:\n\t%s",
		strings.Join(errorStrings, "\n\t"),
	)
}

//
// Checker functions
//

// Checker defines the type of function which can represent a Vala
// checker.  If the Checker fails, returns false with a corresponding
// error message. If the Checker succeeds, returns true, but _also_
// returns an error message. This helps to support the Not function.
type Checker func() (checkerIsTrue bool, err error)

// Not returns the inverse of any Checker passed in.
func Not(checker Checker) Checker {

	return func() (passed bool, err error) {
		if passed, err = checker(); passed {
			return false, fmt.Errorf("Not(%s)", err)
		}

		return true, nil
	}
}

// Equals performs a basic == on the given parameters and fails if
// they are not equal.
func Equals(lhs, rhs interface{}, paramName string) Checker {

	return func() (pass bool, err error) {
		return (lhs == rhs), fmt.Errorf("Parameters were not equal: %v, %v", lhs, rhs)
	}
}

// IsNotNil checks to see if the value passed in is nil. This Checker
// attempts to check the most performant things first, and then
// degrade into the less-performant, but accurate checks for nil.
func IsNotNil(obtained interface{}, paramName string) Checker {
	return func() (isNotNil bool, err error) {

		if obtained == nil {
			isNotNil = false
		} else if str, ok := obtained.(string); ok {
			isNotNil = str != ""
		} else {
			switch v := reflect.ValueOf(obtained); v.Kind() {
			case
				reflect.Chan,
				reflect.Func,
				reflect.Interface,
				reflect.Map,
				reflect.Ptr,
				reflect.Slice:
				isNotNil = !v.IsNil()
			default:
				panic("Vala is unable to check this type for nilability at this time.")
			}
		}

		return isNotNil, fmt.Errorf("Parameter was nil: %v", paramName)
	}
}

// HasLen checks to ensure the given argument is the desired length.
func HasLen(param interface{}, desiredLength int, paramName string) Checker {

	return func() (hasLen bool, err error) {
		hasLen = desiredLength == reflect.ValueOf(param).Len()
		return hasLen, fmt.Errorf("Parameter did not contain the correct number of elements: %v", paramName)
	}
}

// GreaterThan checks to ensure the given argument is greater than the
// given value.
func GreaterThan(param int, comparativeVal int, paramName string) Checker {

	return func() (isGreaterThan bool, err error) {
		if isGreaterThan = param > comparativeVal; !isGreaterThan {
			err = fmt.Errorf(
				"Parameter's length was not greater than:  %s(%d) < %d",
				paramName,
				param,
				comparativeVal)
		}

		return isGreaterThan, err
	}
}

// StringNotEmpty checks to ensure the given string is not empty.
func StringNotEmpty(obtained, paramName string) Checker {
	return func() (isNotEmpty bool, err error) {
		isNotEmpty = obtained != ""
		err = fmt.Errorf("Parameter is an empty string: %s", paramName)
		return
	}
}

// Or executes multiple Checkers and makes sure one is valid
func Or(checkers ...Checker) Checker {
	return func() (valid bool, err error) {
		msgs := make([]string, 0, len(checkers))
		for _, c := range checkers {
			v, e := c()
			if v {
				return true, nil
			}
			msgs = append(msgs, e.Error())
		}
		return false, fmt.Errorf("all checks failed:\n\t%s", strings.Join(msgs, "\n\t"))
	}
}

// And executes multiple Checkers and makes sure all are valid
func And(checkers ...Checker) Checker {
	return func() (valid bool, err error) {
		for _, checker := range checkers {
			if pass, err := checker(); !pass {
				return false, err
			}
		}
		return true, nil
	}
}
