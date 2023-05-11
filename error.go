package tools

import (
	"fmt"
	"os"
)

// Fail formats according to a format specifier, writes to stderr and exits with code 1.
func Fail(format string, a ...interface{}) {
	if len(a) > 0 {
		format = fmt.Sprintf(format, a...)
	}
	fmt.Fprintln(os.Stderr, format)
	Exit(1)
}
