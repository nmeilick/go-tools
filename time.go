package tools

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var (
	// Regex patterns for duration units and for checking validity of duration units.
	reDurationUnit       = regexp.MustCompile(`(\d+(?:\.\d+)?)([a-zA-Zµ]+)`)
	reValidDurationUnits = regexp.MustCompile("^[+-]?(" + reDurationUnit.String() + ")+$")
)

// ParseDuration takes a string representing a duration and returns its equivalent time.Duration.
// It supports different units like seconds, minutes, hours, days, weeks and years.
func ParseDuration(input string) (time.Duration, error) {
	// Remove all whitespace and lowercase the given duration
	cleaned := strings.ToLower(strings.Join(strings.Fields(input), ""))

	// Check if the cleaned duration string is a valid duration
	if !reValidDurationUnits.MatchString(cleaned) {
		return 0, fmt.Errorf("invalid duration: %q", input)
	}

	// Extract the number and unit parts of the duration
	matches := reDurationUnit.FindAllStringSubmatch(cleaned, -1)

	// Accumulate the total duration by iterating over each part
	var total time.Duration
	for _, match := range matches {
		value, err := strconv.ParseFloat(match[1], 64)
		if err != nil {
			return 0, fmt.Errorf("invalid number in duration: %v", err)
		}

		unit := strings.ToLower(match[2])
		var d time.Duration
		switch unit {
		case "ns", "nsec", "nsecs", "nanosecond", "nanoseconds":
			d = time.Nanosecond
		case "µs", "musec", "musecs", "microsecond", "microseconds":
			d = time.Microsecond
		case "ms", "msec", "msecs", "millisecond", "milliconds":
			d = time.Millisecond
		case "s", "sec", "secs", "second", "seconds":
			d = time.Second
		case "m", "min", "mins", "minute", "minutes":
			d = time.Minute
		case "h", "hr", "hrs", "hour", "hours":
			d = time.Hour
		case "d", "day", "days":
			d = time.Hour * 24
		case "w", "wk", "wks", "week", "weeks":
			d = time.Hour * 24 * 7
		case "y", "yr", "yrs", "year", "years":
			d = time.Hour * 24 * 365
		default:
			return 0, fmt.Errorf("invalid unit %q in duration", unit)
		}

		total += time.Duration(value * float64(d))
	}

	if strings.HasPrefix(cleaned, "-") {
		total = -total
	}

	return total, nil
}

// ParseDurationWithDefaultUnit is similar to ParseDuration but it accepts a default unit.
// If the input string is a simple float, it assumes the default unit.
func ParseDurationWithDefaultUnit(input, defaultUnit string) (time.Duration, error) {
	// Remove all whitespace and lowercase the given duration
	cleaned := strings.ToLower(strings.Join(strings.Fields(input), ""))

	// Try to parse the cleaned input as a float
	_, err := strconv.ParseFloat(cleaned, 64)
	if err != nil {
		// If the parsing fails, the input likely includes units, so we use ParseDuration
		return ParseDuration(input)
	}
	// If the parsing succeeds, the input is a number without units, so we append the default unit and use ParseDuration
	return ParseDuration(cleaned + defaultUnit)
}

// FormatDuration takes a time.Duration and formats it as a string in the format used by ParseDuration.
// The format is the largest suitable unit followed by the next largest, and so on.
// For example, 25 hours 90 seconds is formatted as "1d1h1m30s".
func FormatDuration(d time.Duration) string {
	var parts []string

	// Check if the duration is negative
	neg := d < 0
	if neg {
		// If it is, make it positive for formatting and add a "-" prefix
		d = -d
		parts = append(parts, "-")
	}

	// Break down the duration into years, weeks, days, hours, minutes, and seconds
	years := int(d.Hours() / 24 / 365)
	d -= time.Duration(years) * 365 * 24 * time.Hour

	weeks := int(d.Hours() / 24 / 7)
	d -= time.Duration(weeks) * 7 * 24 * time.Hour

	days := int(d.Hours() / 24)
	d -= time.Duration(days) * 24 * time.Hour

	// Construct the string representation using the largest units possible
	if years > 0 {
		parts = append(parts, fmt.Sprintf("%dy", years))
	}
	if weeks > 0 {
		parts = append(parts, fmt.Sprintf("%dw", weeks))
	}
	if days > 0 {
		parts = append(parts, fmt.Sprintf("%dd", days))
	}

	if d > 0 {
		// Let time.Duration.String() do the rest of the work
		s := d.String()
		s = strings.TrimPrefix(strings.TrimPrefix(s, "0h"), "0m")
		if s != "" {
			parts = append(parts, s)
		}
	}

	// If no parts were added, the duration is 0
	if len(parts) == 0 {
		return "0s"
	}

	// Join all the parts with no separator
	return strings.Join(parts, "")
}
