package math

import (
	"github.com/anhcraft/rice/exec/types"
	"github.com/anhcraft/rice/exec/types/values"
	"math"
	"reflect"
	"testing"
)

func TestMax(t *testing.T) {
	nan := values.Float(math.NaN())
	testCases := []struct {
		name    string
		input   []types.Value
		want    any
		wantErr bool
	}{
		{"all ints", []types.Value{values.Int(1), values.Int(5), values.Int(3)}, values.Int(5), false},
		{"all floats", []types.Value{values.Float(1.1), values.Float(5.5), values.Float(3.3)}, values.Float(5.5), false},
		{"mixed int and float, float is max", []types.Value{values.Int(1), values.Float(5.5), values.Int(3)}, values.Float(5.5), false},
		{"mixed int and float, int is max", []types.Value{values.Int(10), values.Float(5.5), values.Int(3)}, values.Int(10), false},
		{"with negative numbers", []types.Value{values.Int(-1), values.Float(-5.5), values.Int(-3)}, values.Int(-1), false},
		{"with non-numeric types", []types.Value{values.Int(1), values.String("ignore"), values.Int(5)}, values.Int(5), false},
		{"with NaN", []types.Value{values.Int(1), values.Float(10.0), nan}, nan, false},
		{"no arguments", []types.Value{}, nil, false},
		{"only non-numeric arguments", []types.Value{values.String("a"), values.Bool(false)}, nil, false},
		{"single int", []types.Value{values.Int(42)}, values.Int(42), false},
		{"single float", []types.Value{values.Float(42.5)}, values.Float(42.5), false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := Max(tc.input...)

			if (err != nil) != tc.wantErr {
				t.Errorf("Max() error = %v, wantErr %v", err, tc.wantErr)
				return
			}

			// Special handling for NaN since NaN != NaN
			if gotNaN, ok := got.(values.Float); ok && math.IsNaN(float64(gotNaN)) {
				if wantNaN, ok := tc.want.(values.Float); !ok || !math.IsNaN(float64(wantNaN)) {
					t.Errorf("Max() = %v, want %v", got, tc.want)
				}
			} else if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("Max() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestMin(t *testing.T) {
	nan := values.Float(math.NaN())
	testCases := []struct {
		name    string
		input   []types.Value
		want    any
		wantErr bool
	}{
		{"all ints", []types.Value{values.Int(1), values.Int(5), values.Int(3)}, values.Int(1), false},
		{"all floats", []types.Value{values.Float(1.1), values.Float(5.5), values.Float(3.3)}, values.Float(1.1), false},
		{"mixed int and float, float is min", []types.Value{values.Int(10), values.Float(5.5), values.Int(8)}, values.Float(5.5), false},
		{"mixed int and float, int is min", []types.Value{values.Int(1), values.Float(5.5), values.Int(3)}, values.Int(1), false},
		{"with negative numbers", []types.Value{values.Int(-1), values.Float(-5.5), values.Int(-3)}, values.Float(-5.5), false},
		{"with non-numeric types", []types.Value{values.Int(1), values.String("ignore"), values.Int(5)}, values.Int(1), false},
		{"with NaN", []types.Value{values.Int(1), values.Float(10.0), nan}, nan, false},
		{"no arguments", []types.Value{}, nil, false},
		{"only non-numeric arguments", []types.Value{values.String("a"), values.Bool(false)}, nil, false},
		{"single int", []types.Value{values.Int(42)}, values.Int(42), false},
		{"single float", []types.Value{values.Float(42.5)}, values.Float(42.5), false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := Min(tc.input...)

			if (err != nil) != tc.wantErr {
				t.Errorf("Min() error = %v, wantErr %v", err, tc.wantErr)
				return
			}

			if gotNaN, ok := got.(values.Float); ok && math.IsNaN(float64(gotNaN)) {
				if wantNaN, ok := tc.want.(values.Float); !ok || !math.IsNaN(float64(wantNaN)) {
					t.Errorf("Min() = %v, want %v", got, tc.want)
				}
			} else if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("Min() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestAbs(t *testing.T) {
	testCases := []struct {
		name    string
		input   types.Value
		want    types.Value
		wantErr bool
	}{
		{"positive int", values.Int(10), values.Int(10), false},
		{"negative int", values.Int(-10), values.Int(10), false},
		{"zero int", values.Int(0), values.Int(0), false},
		{"positive float", values.Float(10.5), values.Float(10.5), false},
		{"negative float", values.Float(-10.5), values.Float(10.5), false},
		{"zero float", values.Float(0.0), values.Float(0.0), false},
		{"invalid type", values.String("abc"), nil, true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := Abs(tc.input)
			if (err != nil) != tc.wantErr {
				t.Errorf("Abs() error = %v, wantErr %v", err, tc.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("Abs() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestSqrt(t *testing.T) {
	nan := values.Float(math.NaN())
	testCases := []struct {
		name    string
		input   types.Value
		want    types.Value
		wantErr bool
	}{
		{"perfect square int", values.Int(9), values.Float(3.0), false},
		{"non-perfect square int", values.Int(2), values.Float(math.Sqrt(2)), false},
		{"float", values.Float(16.0), values.Float(4.0), false},
		{"zero", values.Int(0), values.Float(0.0), false},
		{"negative number", values.Int(-4), nan, false},
		{"bool true", values.Bool(true), values.Float(1.0), false},
		{"invalid type", values.String("abc"), nil, true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := Sqrt(tc.input)
			if (err != nil) != tc.wantErr {
				t.Errorf("Sqrt() error = %v, wantErr %v", err, tc.wantErr)
				return
			}
			if gotNaN, ok := got.(values.Float); ok && math.IsNaN(float64(gotNaN)) {
				if wantNaN, ok := tc.want.(values.Float); !ok || !math.IsNaN(float64(wantNaN)) {
					t.Errorf("Sqrt() = %v, want %v", got, tc.want)
				}
			} else if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("Sqrt() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestFloor(t *testing.T) {
	testCases := []struct {
		name    string
		input   types.Value
		want    types.Value
		wantErr bool
	}{
		{"positive float", values.Float(3.7), values.Float(3.0), false},
		{"negative float", values.Float(-3.7), values.Float(-4.0), false},
		{"integer float", values.Float(3.0), values.Float(3.0), false},
		{"int", values.Int(5), values.Int(5), false},
		{"bool true", values.Bool(true), values.Int(1), false},
		{"invalid type", values.String("abc"), nil, true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := Floor(tc.input)
			if (err != nil) != tc.wantErr {
				t.Errorf("Floor() error = %v, wantErr %v", err, tc.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("Floor() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestCeil(t *testing.T) {
	testCases := []struct {
		name    string
		input   types.Value
		want    types.Value
		wantErr bool
	}{
		{"positive float", values.Float(3.2), values.Float(4.0), false},
		{"negative float", values.Float(-3.7), values.Float(-3.0), false},
		{"integer float", values.Float(3.0), values.Float(3.0), false},
		{"int", values.Int(5), values.Int(5), false},
		{"bool true", values.Bool(true), values.Int(1), false},
		{"invalid type", values.String("abc"), nil, true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := Ceil(tc.input)
			if (err != nil) != tc.wantErr {
				t.Errorf("Ceil() error = %v, wantErr %v", err, tc.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("Ceil() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestPow(t *testing.T) {
	testCases := []struct {
		name     string
		base     types.Value
		exponent types.Value
		want     types.Value
		wantErr  bool
	}{
		{"int base, int exponent", values.Int(2), values.Int(3), values.Float(8.0), false},
		{"float base, int exponent", values.Float(2.5), values.Int(2), values.Float(6.25), false},
		{"int base, float exponent", values.Int(4), values.Float(0.5), values.Float(2.0), false},
		{"negative base, odd exponent", values.Int(-2), values.Int(3), values.Float(-8.0), false},
		{"negative base, even exponent", values.Int(-2), values.Int(2), values.Float(4.0), false},
		{"zero exponent", values.Int(10), values.Int(0), values.Float(1.0), false},
		{"zero base", values.Int(0), values.Int(5), values.Float(0.0), false},
		{"invalid base", values.String("a"), values.Int(2), nil, true},
		{"invalid exponent", values.Int(2), values.String("b"), nil, true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := Pow(tc.base, tc.exponent)
			if (err != nil) != tc.wantErr {
				t.Errorf("Pow() error = %v, wantErr %v", err, tc.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("Pow() = %v, want %v", got, tc.want)
			}
		})
	}
}
