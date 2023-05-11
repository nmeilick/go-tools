package tools

import (
	"reflect"
	"sort"
	"strings"
	"unicode"

	"golang.org/x/exp/constraints"
)

type MapFunc[T any] func(T) T

// Map applies mapping functions sequentially to each value and returns the result.
func Map[T any](values []T, funcs ...MapFunc[T]) []T {
	if values == nil {
		return nil
	}

	mapped := make([]T, len(values))
	copy(mapped, values)
	for i := range mapped {
		for _, f := range funcs {
			mapped[i] = f(mapped[i])
		}
	}
	return mapped
}

type SelectFunc[T any] func(T) bool

// Select returns these values for which any of the given select functions return true.
// If no select function is given, IsNotZero will be used to filter out empty elements.
func Select[T any](values []T, funcs ...SelectFunc[T]) []T {
	if values == nil {
		return nil
	}

	if len(funcs) == 0 {
		funcs = append(funcs, IsNotZero[T])
	}

	selected := []T{}
	for _, v := range values {
		for _, f := range funcs {
			if f != nil && f(v) {
				selected = append(selected, v)
				break
			}
		}
	}
	return selected
}

// IsZero checks whether the given value has the default value for its type.
func IsZero[T any](v T) bool {
	return isZero(v)
}

// IsNotZero checks whether the given value is different from the default value for its type.
func IsNotZero[T any](v T) bool {
	return !IsZero(v)
}

func isZero(i interface{}) bool {
	v := reflect.ValueOf(i)
	switch v.Kind() {
	case reflect.Func, reflect.Map, reflect.Slice:
		return v.IsNil()
	case reflect.Array:
		z := true
		for i := 0; i < v.Len(); i++ {
			z = z && isZero(v.Index(i).Interface())
		}
		return z
	case reflect.Struct:
		z := true
		for i := 0; i < v.NumField(); i++ {
			z = z && isZero(v.Field(i).Interface())
		}
		return z
	case reflect.Ptr:
		if v.IsNil() {
			return true
		}
		return isZero(v.Elem().Interface())
	}
	// Compare other types directly:
	z := reflect.Zero(v.Type())
	return v.Interface() == z.Interface()
}

// Unique returns a copy of the slice with all duplicates removed.
func Unique[T comparable](values []T) []T {
	if values == nil {
		return nil
	}

	result := []T{}
	seen := map[T]bool{}
	for _, v := range values {
		if !seen[v] {
			seen[v] = true
			result = append(result, v)
		}
	}
	return result
}

// Tokens splits the given values at whitespace or comma and returns lower-cased unique values.
func Tokens[T ~string](values ...T) []T {
	split := func(r rune) bool {
		return unicode.IsSpace(r) || r == ','
	}

	tokens := []T{}
	seen := map[string]bool{}
	for _, v := range values {
		for _, s := range strings.FieldsFunc(string(v), split) {
			if s = strings.ToLower(s); !seen[s] {
				seen[s] = true
				tokens = append(tokens, T(s))
			}
		}
	}
	return tokens
}

// FirstNonEmpty returns the first non-empty element of the given list.
// To use a fallback value, put it as the last element.
func FirstNonEmpty[T any](values ...T) T {
	for _, v := range values {
		if !isZero(v) {
			return v
		}
	}
	// Return the default value for the given type
	var zero T
	return zero
}

// Includes returns true if the slice contains the given element.
func Includes[T comparable](values []T, value T) bool {
	for i := range values {
		if values[i] == value {
			return true
		}
	}
	return false
}

// Minus is a generic function that returns a new slice including only those elements of the first input slice
// that are not present in the second input slice.
//
// Example usage:
//   s1 := []int{1, 2, 3, 4}
//   s2 := []int{1, 3}
//   result := Minus(s1, s2)  // Output: [2, 4]
func Minus[T comparable](s1, s2 []T) []T {
	result := []T{}

	m := ToMapWithValue(s2, true)
	for _, v := range s1 {
		if _, ok := m[v]; !ok {
			v := v
			result = append(result, v)
		}
	}
	return result
}

// Merge returns a new slice that includes all elements of all input slices. Duplicates are removed.
//
// Example usage:
//   s1 := []int{1, 2}
//   s2 := []int{2, 3, 4}
//   result := Merge(s1, s2)  // Output: [1, 2, 3, 4]
func Merge[T comparable](slices ...[]T) []T {
	seen := map[T]bool{}
	result := []T{}
	for _, s := range slices {
		for _, v := range s {
			if !seen[v] {
				v := v
				seen[v] = true
				result = append(result, v)
			}
		}
	}
	return result
}

// Sort returns a sorted copy of a slice.
//
// Example usage:
//   s := []int{3, 2, 4, 1}
//   result := Sort(s)  // Output: [1, 2, 3, 4]
func Sort[T constraints.Ordered](values []T) []T {
	result := make([]T, len(values))
	copy(result, values)
	sort.Slice(result, func(i, j int) bool { return result[i] < result[j] })
	return result
}

// SortNatural returns a naturally sorted copy of a slice of string type.
//
// Example usage:
//   s := []string{"v1.10.3", "v1.5.1", "v1.10.1"}
//   result := Sort(s)  // Output: ["v1.5.1", "v1.10.1", "v1.10.3"]
func SortNatural[T ~string](values []T, ignoreCase bool) []T {
	result := make([]T, len(values))
	copy(result, values)
	sort.Slice(result,
		func(i, j int) bool {
			si, sj := string(result[i]), string(result[j])
			if ignoreCase {
				si, sj = strings.ToLower(si), strings.ToLower(sj)
			}
			return naturalLess(si, sj)
		},
	)

	return result
}

func naturalLess(s1, s2 string) bool {
	var chunk1, chunk2 string
	var isNum1, isNum2 bool

	for {
		if chunk1, isNum1, s1 = getChunk(s1); chunk1 == "" {
			return true
		}

		if chunk2, isNum2, s2 = getChunk(s2); chunk2 == "" {
			return false
		}

		if chunk1 != chunk2 {
			if isNum1 && isNum2 {
				if len(chunk1) != len(chunk2) {
					return len(chunk1) < len(chunk2)
				}
				return chunk1 < chunk2
			}
			return chunk1 < chunk2
		}
	}
}

func getChunk(s string) (string, bool, string) {
	var text, digits strings.Builder
	for _, r := range s {
		if unicode.IsDigit(r) {
			if text.Len() > 0 {
				break
			}
			digits.WriteRune(r)
		} else {
			if digits.Len() > 0 {
				break
			}
			text.WriteRune(r)
		}
	}

	if digits.Len() > 0 {
		s = s[digits.Len():]
		chunk := strings.TrimLeft(digits.String(), "0") // remove leading zeroes as we'll use a string comparison
		if chunk == "" {
			chunk = "0"
		}
		return chunk, true, s
	}
	chunk := strings.TrimSpace(text.String())
	return chunk, false, s[text.Len():]
}

// Define a function type that takes a Key and returns a Key and Value.
type KeyValueGenerator[K comparable, V any] func(K) (K, V)

// ToMap returns a map where each entry is the result of the generator function.
// Empty keys are ignored.
func ToMap[K comparable, V any](keys []K, gen KeyValueGenerator[K, V]) map[K]V {
	m := map[K]V{}
	var zero K
	for _, key := range keys {
		key, val := gen(key)
		if key != zero {
			m[key] = val
		}
	}
	return m
}

// ToMapWithValue returns a map with each key set to the given value.
func ToMapWithValue[K comparable, V any](keys []K, v V) map[K]V {
	return ToMap(keys, func(k K) (K, V) { return k, v })
}
