package frontend

import (
	"encoding/json"
	"errors"
	"math"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func loadTestCases[T any](t *testing.T, filename string) []T {
	filePath := filepath.Join("testdata", filename)

	data, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("Failed to read test data file %s: %v", filePath, err)
	}

	var testCases []T
	err = json.Unmarshal(data, &testCases)
	if err != nil {
		t.Fatalf("Failed to unmarshal JSON from %s: %v", filePath, err)
	}

	return testCases
}

func TestConvertNumber(t *testing.T) {
	const float64EqualityThreshold = 1e-9

	testCases := []struct {
		name           string
		integerPart    uint64
		decimalPart    uint64
		hasDecimalPart bool
		expPart        int64
		wantVal        interface{}
		wantErr        error
	}{
		// --- Integer Path Success Cases ---
		{
			name:        "Simple Integer",
			integerPart: 123,
			decimalPart: 0,
			expPart:     0,
			wantVal:     int64(123),
			wantErr:     nil,
		},
		{
			name:        "Integer with Positive Exponent",
			integerPart: 45,
			decimalPart: 0,
			expPart:     3,
			wantVal:     int64(45000),
			wantErr:     nil,
		},
		{
			name:        "Max Int64",
			integerPart: uint64(math.MaxInt64),
			decimalPart: 0,
			expPart:     0,
			wantVal:     int64(math.MaxInt64),
			wantErr:     nil,
		},
		// --- Integer Path Error Cases ---
		{
			name:        "Integer Part Overflow",
			integerPart: uint64(math.MaxInt64) + 1,
			decimalPart: 0,
			expPart:     0,
			wantVal:     nil,
			wantErr:     errIntegerOverflow,
		},
		{
			name:        "Integer Exponentiation Overflow",
			integerPart: math.MaxInt64 / 5,
			decimalPart: 0,
			expPart:     1,
			wantVal:     nil,
			wantErr:     errIntegerOverflow,
		},
		// --- Float Path Success Cases ---
		{
			name:        "Simple Float",
			integerPart: 12,
			decimalPart: 345,
			expPart:     0,
			wantVal:     12.345,
			wantErr:     nil,
		},
		{
			name:        "Float with Positive Exponent",
			integerPart: 1,
			decimalPart: 23,
			expPart:     2,
			wantVal:     123.0,
			wantErr:     nil,
		},
		{
			name:        "Float with Negative Exponent",
			integerPart: 123,
			decimalPart: 45,
			expPart:     -3,
			wantVal:     0.12345,
			wantErr:     nil,
		},
		{
			name:        "Integer with Negative Exponent (becomes float)",
			integerPart: 987,
			decimalPart: 0,
			expPart:     -2,
			wantVal:     9.87,
			wantErr:     nil,
		},
		// --- Float Path Error Cases ---
		{
			name:        "Float Overflow",
			integerPart: 1,
			decimalPart: 2,
			expPart:     400, // 1.2e400
			wantVal:     nil,
			wantErr:     errFloat64Overflow,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			gotVal, gotErr := convertNumber(tc.integerPart, tc.decimalPart, tc.hasDecimalPart, tc.expPart)

			if !errors.Is(gotErr, tc.wantErr) {
				t.Errorf("convertNumber() error = %v, wantErr %v", gotErr, tc.wantErr)
				return
			}

			if tc.wantErr != nil {
				return
			}

			if reflect.TypeOf(gotVal) != reflect.TypeOf(tc.wantVal) {
				t.Errorf("convertNumber() type mismatch: got %T, want %T", gotVal, tc.wantVal)
				return
			}

			switch want := tc.wantVal.(type) {
			case int64:
				if got := gotVal.(int64); got != want {
					t.Errorf("convertNumber() got = %v, want %v", got, want)
				}
			case float64:
				if got := gotVal.(float64); math.Abs(got-want) > float64EqualityThreshold {
					t.Errorf("convertNumber() got = %v, want %v", got, want)
				}
			}
		})
	}
}
