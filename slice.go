package tools

import (
	"strings"
	"unicode"
)

type MapFunc[T any] func(T) T

// Map applies the given mapping functions sequentially to the values and returns the result.
func Map[T any](vv []T, ff ...MapFunc[T]) (mapped []T) {
	for _, v := range vv {
		for _, f := range ff {
			v = f(v)
		}
		mapped = append(mapped, v)
	}
	return
}

type SelectFunc[T any] func(T) bool

// Select returns these values for which any of the given select functions return true.
// If no select function is given, IsNotZero will be used to filter out empty elements.
func Select[T comparable](vv []T, ff ...SelectFunc[T]) (selected []T) {
	if len(ff) == 0 {
		ff = append(ff, IsNotZero[T])
	}
	for _, v := range vv {
		for _, f := range ff {
			if f != nil && f(v) {
				selected = append(selected, v)
				break
			}
		}
	}
	return
}

// IsZero checks whether the given value has the default value for its type.
func IsZero[T comparable](v T) bool {
	var zero T
	return v == zero
	//return !reflect.ValueOf(v).IsNil() && reflect.DeepEqual(v, reflect.Zero(reflect.TypeOf(v)).Interface())
}

// IsNotZero checks whether the given value is different from the default value for its type.
func IsNotZero[T comparable](v T) bool {
	return !IsZero(v)
}

// Unique returns a slice with all duplicates removed.
func Unique[T comparable](vv []T) (result []T) {
	if vv != nil {
		result = []T{}
		seen := map[T]bool{}
		for _, v := range vv {
			if !seen[v] {
				seen[v] = true
				result = append(result, v)
			}
		}
	}
	return
}

type Lenable interface {
	~string | ~[]string | ~[]any | ~[]bool |
		~[]int | ~[]int8 | ~[]int16 | ~[]int32 | ~[]int64 | ~[]uint | ~[]uint8 | ~[]uint16 | ~[]uint32 | ~[]uint64 |
		~[]uintptr | ~[]float32 | ~[]float64 | ~[]complex64 | ~[]complex128
}

// Prune returns a slice with all empty elements removed.
func Prune[T Lenable](vv []T) (pruned []T) {
	if vv != nil {
		pruned = []T{}
		for _, v := range vv {
			if len(v) > 0 {
				pruned = append(pruned, v)
			}
		}
	}
	return
}

// Either returns the first non-empty element of the given list.
func Either[T Lenable](vv ...T) T {
	for _, v := range vv {
		if len(v) > 0 {
			return v
		}
	}
	// Return the default value for the given type
	var zero T
	return zero
}

// Includes returns true if the slice contains the given element.
func Includes[T comparable](vv []T, v T) bool {
	for i := range vv {
		if vv[i] == v {
			return true
		}
	}
	return false
}

// Tokens splits to given string types at whitespace or comma and returns lower-cased unique values.
func Tokens[T ~string](vv ...T) []T {
	split := func(r rune) bool {
		return unicode.IsSpace(r) || r == ','
	}

	tokens := []T{}
	seen := map[string]bool{}
	for _, v := range vv {
		for _, s := range strings.FieldsFunc(string(v), split) {
			if s = strings.ToLower(s); !seen[s] {
				seen[s] = true
				tokens = append(tokens, T(s))
			}
		}
	}
	return tokens
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
