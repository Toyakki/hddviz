package main_test

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
)

// assert fails the test if the condition is false.
func assert(tb testing.TB, condition bool, msg string, v ...interface{}) {
	if !condition {
		_, file, line, _ := runtime.Caller(1)
		fmt.Printf("\033[31m%s:%d: "+msg+"\033[39m\n\n", append([]interface{}{filepath.Base(file), line}, v...)...)
		tb.FailNow()
	}
}

// // ok fails the test if an err is not nil.
// func ok(tb testing.TB, err error) {
// 	if err != nil {
// 		_, file, line, _ := runtime.Caller(1)
// 		fmt.Printf("\033[31m%s:%d: unexpected error: %s\033[39m\n\n", filepath.Base(file), line, err.Error())
// 		tb.FailNow()
// 	}
// }

// // equals fails the test if exp is not equal to act.
//
//	func equals(tb testing.TB, exp, act interface{}) {
//		if !reflect.DeepEqual(exp, act) {
//			_, file, line, _ := runtime.Caller(1)
//			fmt.Printf("\033[31m%s:%d:\n\n\texp: %#v\n\n\tgot: %#v\033[39m\n\n", filepath.Base(file), line, exp, act)
//			tb.FailNow()
//		}
//	}
func requireExitCode(t *testing.T, err error, out []byte, exit_code int) {
	t.Helper()
	assert(t, err != nil, "expected exit code %d, got sucess; output = %s", exit_code, string(out))
	ee, ok := err.(*exec.ExitError)
	assert(t, ok, "expected *exec.ExitError, got %T (%v); output=%s", err, err, string(out))

	assert(t, ee.ExitCode() == exit_code, "expected exit code %d, got %d; output=%s", exit_code, ee.ExitCode(), string(out))
}

func TestMain_Invalid_Flag(t *testing.T) {
	cmd := exec.Command("go", "run", ".", "-non-existent-flag")
	out, err := cmd.CombinedOutput()
	requireExitCode(t, err, out, 1)
}

func TestMain_Invalid_Root(t *testing.T) {
	cmd := exec.Command("go", "run", ".", "-root", "/path/non_existent")
	out, err := cmd.CombinedOutput()
	requireExitCode(t, err, out, 1)
}
