//go:build !solution

package testequal

// AssertEqual checks that expected and actual are equal.
//
// Marks caller function as having failed but continues execution.
//
// Returns true iff arguments are equal.
func AssertEqual(t T, expected, actual interface{}, msgAndArgs ...interface{}) bool {
	t.Helper()
	if equal(expected, actual) {
		return true
	}
	if len(msgAndArgs) > 0 {
		t.Errorf(msgAndArgs[0].(string), msgAndArgs[1:]...)
	} else {
		t.Errorf("")
	}
	return false
}

// AssertNotEqual checks that expected and actual are not equal.
//
// Marks caller function as having failed but continues execution.
//
// Returns true iff arguments are not equal.
func AssertNotEqual(t T, expected, actual interface{}, msgAndArgs ...interface{}) bool {
	t.Helper()
	if equal(expected, actual) {
		if len(msgAndArgs) > 0 {
			t.Errorf(msgAndArgs[0].(string), msgAndArgs[1:]...)
		} else {
			t.Errorf("")
		}
		return false
	}
	return true
}

// RequireEqual does the same as AssertEqual but fails caller test immediately.
func RequireEqual(t T, expected, actual interface{}, msgAndArgs ...interface{}) {
	t.Helper()
	if !equal(expected, actual) {
		if len(msgAndArgs) > 0 {
			t.Errorf(msgAndArgs[0].(string), msgAndArgs[1:]...)
		} else {
			t.Errorf("")
		}
		t.FailNow()
	}
}

// RequireNotEqual does the same as AssertNotEqual but fails caller test immediately.
func RequireNotEqual(t T, expected, actual interface{}, msgAndArgs ...interface{}) {
	t.Helper()
	if equal(expected, actual) {
		if len(msgAndArgs) > 0 {
			t.Errorf(msgAndArgs[0].(string), msgAndArgs[1:]...)
		} else {
			t.Errorf("")
		}
		t.FailNow()
	}
}

func equal(expected interface{}, actual interface{}) bool {
	switch exp := expected.(type) {
	case int, int8, int16, int32, int64, uint8, uint16, uint32, uint64, string:
		return exp == actual
	case []int:
		act, ok := actual.([]int)
		if !ok {
			return false
		}
		return equalIntSlices(exp, act)
	case map[string]string:
		act, ok := actual.(map[string]string)
		if !ok {
			return false
		}
		return equalMaps(exp, act)
	case []byte:
		act, ok := actual.([]byte)
		if !ok {
			return false
		}
		return equalBytes(exp, act)
	default:
		return false
	}
}

func equalIntSlices(a []int, b []int) bool {
	aNil := a == nil
	bNil := b == nil
	if aNil != bNil {
		return false
	}
	if len(a) != len(b) {
		return false
	}
	for i := range len(a) {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func equalBytes(a []byte, b []byte) bool {
	aNil := a == nil
	bNil := b == nil
	if aNil != bNil {
		return false
	}
	if len(a) != len(b) {
		return false
	}
	for i := range len(a) {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func equalMaps(a map[string]string, b map[string]string) bool {
	aNil := a == nil
	bNil := b == nil
	if aNil != bNil {
		return false
	}
	if len(a) != len(b) {
		return false
	}
	for k := range a {
		bV, exists := b[k]
		if !exists {
			return false
		}
		if a[k] != bV {
			return false
		}
	}
	return true
}
