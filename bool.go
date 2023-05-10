package tools

import (
	"fmt"
	"strconv"
	"strings"
)

// checkState checks if the given value indicates a specific state based on the provided condition.
// If the state cannot be determined, it returns the provided default value.
func checkState(v interface{}, condition bool, deflt bool) bool {
	s := strings.ToLower(fmt.Sprintf("%v", v))

	switch s {
	case "1", "on", "yes", "y", "enabled", "active", "true", "t", "+":
		return condition
	case "0", "off", "no", "n", "disabled", "inactive", "false", "f", "-":
		return !condition
	}

	if n, err := strconv.ParseInt(s, 10, 64); err == nil {
		if condition {
			return n > 0
		}
		return n <= 0
	}
	return deflt
}

// IsOn checks if the given value indicates an enabled state.
// If the state cannot be determined, it returns the provided default value.
func IsOn(v interface{}, deflt bool) bool {
	return checkState(v, true, deflt)
}

// IsOff checks if the given value indicates a disabled state.
// If the state cannot be determined, it returns the provided default value.
func IsOff(v interface{}, deflt bool) bool {
	return checkState(v, false, deflt)
}
