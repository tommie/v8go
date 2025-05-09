package v8go_test

import "testing"

func fatalIf(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
}

func recoverPanic(f func()) (recovered interface{}) {
	defer func() {
		recovered = recover()
	}()
	f()
	return nil
}

// errorsJoin is like [errors.Join] in the standard library. But this library
// supports Go 1.19, which doesn't have errors.Join.
func errorsJoin(errs ...error) error {
	for _, err := range errs {
		if err != nil {
			return err
		}
	}
	return nil
}
