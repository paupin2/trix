package trix

import (
	"fmt"
	"reflect"
	"testing"
)

func testEqualString(t *testing.T, actual, expected interface{}) {
	t.Helper()
	if fmt.Sprint(actual) != fmt.Sprint(expected) {
		t.Errorf(`Expected "%v", got "%v"`, expected, actual)
	}
}

func testDeepEqual(t *testing.T, actual, expected interface{}) {
	t.Helper()
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf(`Expected "%v" (%T), got "%v" (%T)`, expected, expected, actual, actual)
	}
}

func testError(t *testing.T, err error, expected string) {
	t.Helper()
	if expected == "" {
		if err != nil {
			t.Errorf(`Expected not error, got "%v"`, err)
		}

	} else if err == nil {
		t.Errorf(`Expected error "%s", got none`, expected)
	} else if actual := err.Error(); actual != expected {
		t.Errorf(`Expected error "%s", got "%s"`, expected, actual)
	}
}

func testTrue(t *testing.T, value bool) {
	t.Helper()
	if !value {
		t.Errorf(`Expected true, got "%v"`, value)
	}
}
