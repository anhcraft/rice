package _type

import (
	"github.com/anhcraft/rice/exec/types"
	"github.com/anhcraft/rice/exec/types/values"
	"reflect"
	"testing"
)

func TestTypeof(t *testing.T) {
	testCases := []struct {
		name    string
		input   types.Value
		want    types.Value
		wantErr bool
	}{
		{"int type", values.Int(123), values.String("Int"), false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := Typeof(tc.input)

			if (err != nil) != tc.wantErr {
				t.Errorf("Typeof() error = %v, wantErr %v", err, tc.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("Typeof() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestLen(t *testing.T) {
	testCases := []struct {
		name    string
		input   types.Value
		want    types.Value
		wantErr bool
	}{
		{"string length", values.String("hello"), values.Int(5), false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := Len(tc.input)

			if (err != nil) != tc.wantErr {
				t.Errorf("Len() error = %v, wantErr %v", err, tc.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("Len() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestInt(t *testing.T) {
	testCases := []struct {
		name    string
		input   types.Value
		want    types.Value
		wantErr bool
	}{
		{"from float", values.Float(123.7), values.Int(123), false},
		{"from valid string", values.String("456"), values.Int(456), false},
		{"from invalid string", values.String("abc"), values.Int(0), true},
		{"from bool true", values.Bool(true), values.Int(1), false},
		{"from bool false", values.Bool(false), values.Int(0), false},
		{"from nil", nil, values.Int(0), false},
		{"from int", values.Int(789), values.Int(789), false},
		{"from negative float", values.Float(-5.5), values.Int(-5), false},
		{"unsupported type func", &values.Func{}, values.Int(0), true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := Int(tc.input)

			if (err != nil) != tc.wantErr {
				t.Errorf("Int() error = %v, wantErr %v", err, tc.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("Int() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestFloat(t *testing.T) {
	testCases := []struct {
		name    string
		input   types.Value
		want    types.Value
		wantErr bool
	}{
		{"from int", values.Int(123), values.Float(123.0), false},
		{"from valid string", values.String("456.78"), values.Float(456.78), false},
		{"from invalid string", values.String("abc"), values.Float(0), true},
		{"from bool true", values.Bool(true), values.Float(1.0), false},
		{"from bool false", values.Bool(false), values.Float(0.0), false},
		{"from nil", nil, values.Float(0.0), false},
		{"from float", values.Float(9.87), values.Float(9.87), false},
		{"unsupported type func", &values.Func{}, values.Float(0.0), true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := Float(tc.input)

			if (err != nil) != tc.wantErr {
				t.Errorf("Float() error = %v, wantErr %v", err, tc.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("Float() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestBool(t *testing.T) {
	testCases := []struct {
		name    string
		input   types.Value
		want    types.Value
		wantErr bool
	}{
		{"from positive int", values.Int(1), values.Bool(true), false},
		{"from zero int", values.Int(0), values.Bool(false), false},
		{"from negative int", values.Int(-1), values.Bool(true), false},
		{"from positive float", values.Float(0.1), values.Bool(true), false},
		{"from zero float", values.Float(0.0), values.Bool(false), false},
		{"from string 'true'", values.String("true"), values.Bool(true), false},
		{"from string 'false'", values.String("false"), values.Bool(false), false},
		{"from string '1'", values.String("1"), values.Bool(true), false},
		{"from string '0'", values.String("0"), values.Bool(false), false},
		{"from invalid string", values.String("text"), values.Bool(false), true},
		{"from nil", nil, values.Bool(false), false},
		{"from bool", values.Bool(true), values.Bool(true), false},
		{"unsupported type func", &values.Func{}, values.Bool(false), true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := Bool(tc.input)

			if (err != nil) != tc.wantErr {
				t.Errorf("Bool() error = %v, wantErr %v", err, tc.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("Bool() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestString(t *testing.T) {
	testCases := []struct {
		name    string
		input   types.Value
		want    types.Value
		wantErr bool
	}{
		{"from int", values.Int(123), values.String("123"), false},
		{"from float", values.Float(45.6), values.String("45.6"), false},
		{"from bool true", values.Bool(true), values.String("true"), false},
		{"from bool false", values.Bool(false), values.String("false"), false},
		{"from nil", nil, values.String(""), false},
		{"from string", values.String("hello"), values.String("hello"), false},
		{
			"from func",
			values.NewFunc([]values.Identifier{}, false, nil, nil),
			values.String("func()"),
			false,
		},
		{
			"from func a",
			values.NewFunc([]values.Identifier{"a"}, false, nil, nil),
			values.String("func(a)"),
			false,
		},
		{
			"from func a...",
			values.NewFunc([]values.Identifier{"a"}, true, nil, nil),
			values.String("func(a...)"),
			false},
		{
			"from func a,b...",
			values.NewFunc([]values.Identifier{"a", "b"}, true, nil, nil),
			values.String("func(a,b...)"),
			false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := String(tc.input)

			if (err != nil) != tc.wantErr {
				t.Errorf("String() error = %v, wantErr %v", err, tc.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("String() = %v, want %v", got, tc.want)
			}
		})
	}
}
