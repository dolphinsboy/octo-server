package util

import (
	"testing"
)

func TestObjToStrUint32(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected string
	}{
		{"uint32_zero", uint32(0), "0"},
		{"uint32_small", uint32(123), "123"},
		{"uint32_max", uint32(4294967295), "4294967295"},
		{"int", int(42), "42"},
		{"int64", int64(9223372036854775807), "9223372036854775807"},
		{"string", "hello", "hello"},
		// float32 tests - ensure proper formatting, not "%!s(float32=...)"
		{"float32_zero", float32(0), "0"},
		{"float32_int", float32(123), "123"},
		{"float32_decimal", float32(3.14), "3.14"},
		{"float32_negative", float32(-99.5), "-99.5"},
		// float64 tests
		{"float64_zero", float64(0), "0"},
		{"float64_int", float64(456), "456"},
		{"float64_decimal", float64(2.718281828), "2.718281828"},
		{"float64_negative", float64(-123.456), "-123.456"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := objToStr(tt.input)
			if result != tt.expected {
				t.Errorf("objToStr(%v) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}
