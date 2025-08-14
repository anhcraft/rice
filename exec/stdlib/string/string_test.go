package string

import (
	"reflect"
	"rice/exec/types"
	"rice/exec/types/values"
	"testing"
)

func TestFormat(t *testing.T) {
	testCases := []struct {
		name    string
		format  values.String
		values  []any
		want    any
		wantErr bool
	}{
		{
			name:   "format with string and int",
			format: "Hello %s, your number is %d",
			values: []any{values.String("World"), values.Int(42)},
			want:   values.String("Hello World, your number is 42"),
		},
		{
			name:   "format with float and bool",
			format: "Value: %.2f, Status: %t",
			values: []any{values.Float(3.14159), values.Bool(true)},
			want:   values.String("Value: 3.14, Status: true"),
		},
		{
			name:   "format with no values",
			format: "Just a static string.",
			values: []any{},
			want:   values.String("Just a static string."),
		},
		{
			name:   "format with nil value",
			format: "Value is %v",
			values: []any{nil},
			want:   values.String("Value is <nil>"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := Format(tc.format, tc.values...)

			if (err != nil) != tc.wantErr {
				t.Errorf("Format() error = %v, wantErr %v", err, tc.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("Format() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestTrim(t *testing.T) {
	testCases := []struct {
		name  string
		input values.String
		want  any
	}{
		{"leading and trailing spaces", "  hello world  ", values.String("hello world")},
		{"only leading spaces", "  hello", values.String("hello")},
		{"only trailing spaces", "hello  ", values.String("hello")},
		{"tabs and newlines", "\t\n hello \n\t", values.String("hello")},
		{"no spaces to trim", "hello", values.String("hello")},
		{"empty string", "", values.String("")},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got, _ := Trim(tc.input)
			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("Trim() = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestToUpper(t *testing.T) {
	testCases := []struct {
		name  string
		input values.String
		want  any
	}{
		{"mixed case", "HeLlO wOrLd", values.String("HELLO WORLD")},
		{"all lowercase", "hello", values.String("HELLO")},
		{"already uppercase", "HELLO", values.String("HELLO")},
		{"with numbers and symbols", "Hello 123!", values.String("HELLO 123!")},
		{"empty string", "", values.String("")},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got, _ := ToUpper(tc.input)
			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("ToUpper() = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestToLower(t *testing.T) {
	testCases := []struct {
		name  string
		input values.String
		want  any
	}{
		{"mixed case", "HeLlO wOrLd", values.String("hello world")},
		{"all uppercase", "HELLO", values.String("hello")},
		{"already lowercase", "hello", values.String("hello")},
		{"with numbers and symbols", "Hello 123!", values.String("hello 123!")},
		{"empty string", "", values.String("")},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got, _ := ToLower(tc.input)
			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("ToLower() = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestInclude(t *testing.T) {
	testCases := []struct {
		name string
		str  values.String
		sub  values.String
		want any
	}{
		{"substring exists", "hello world", "world", values.Bool(true)},
		{"substring does not exist", "hello world", "earth", values.Bool(false)},
		{"substring is identical", "hello", "hello", values.Bool(true)},
		{"empty substring", "hello", "", values.Bool(true)},
		{"empty string", "", "a", values.Bool(false)},
		{"both empty", "", "", values.Bool(true)},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got, _ := Include(tc.str, tc.sub)
			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("Include() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestIndex(t *testing.T) {
	testCases := []struct {
		name string
		str  values.String
		sub  values.String
		want any
	}{
		{"first occurrence", "hello world", "o", values.Int(4)},
		{"substring not found", "hello world", "z", values.Int(-1)},
		{"substring at start", "hello", "he", values.Int(0)},
		{"empty substring", "hello", "", values.Int(0)},
		{"empty string", "", "a", values.Int(-1)},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got, _ := Index(tc.str, tc.sub)
			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("Index() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestLastIndex(t *testing.T) {
	testCases := []struct {
		name string
		str  values.String
		sub  values.String
		want any
	}{
		{"last occurrence", "hello world", "o", values.Int(7)},
		{"substring not found", "hello world", "z", values.Int(-1)},
		{"single occurrence", "hello world", "h", values.Int(0)},
		{"empty substring", "hello", "", values.Int(5)},
		{"empty string", "", "a", values.Int(-1)},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got, _ := LastIndex(tc.str, tc.sub)
			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("LastIndex() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestSubstr(t *testing.T) {
	testCases := []struct {
		name  string
		s     values.String
		start values.Int
		end   values.Int
		want  any
	}{
		{"basic substring", "hello world", 0, 5, values.String("hello")},
		{"substring from middle", "hello world", 6, 11, values.String("world")},
		{"unicode characters", "你好世界", 0, 2, values.String("你好")},
		{"start equals end", "hello", 2, 2, values.String("")},
		{"start greater than end", "hello", 3, 1, values.String("")},
		{"start out of bounds (negative)", "hello", -1, 3, values.String("")},
		{"start out of bounds (large)", "hello", 10, 12, values.String("")},
		{"end out of bounds (large)", "hello", 0, 20, values.String("hello")},
		{"end out of bounds (negative)", "hello", 2, -1, values.String("")},
		{"empty string", "", 0, 0, values.String("")},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got, _ := Substr(tc.s, tc.start, tc.end)
			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("Substr() = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestSplit(t *testing.T) {
	testCases := []struct {
		name      string
		s         values.String
		separator values.String
		want      any
	}{
		{"basic split", "a,b,c", ",", values.ListOf([]values.String{"a", "b", "c"})},
		{"no separator found", "abc", ",", values.ListOf([]values.String{"abc"})},
		{"empty string", "", ",", values.ListOf([]values.String{""})},
		{"empty separator", "abc", "", values.ListOf([]values.String{"a", "b", "c"})},
		{"separator at start and end", ",a,b,", ",", values.ListOf([]values.String{"", "a", "b", ""})},
		{"multiple char separator", "a--b--c", "--", values.ListOf([]values.String{"a", "b", "c"})},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got, _ := Split(tc.s, tc.separator)
			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("Split() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestJoin(t *testing.T) {
	testCases := []struct {
		name      string
		separator values.String
		parts     []types.Value
		want      any
		wantErr   bool
	}{
		{
			name:      "join multiple strings",
			separator: ",",
			parts:     []types.Value{values.String("a"), values.String("b"), values.String("c")},
			want:      values.String("a,b,c"),
		},
		{
			name:      "join mixed types",
			separator: "-",
			parts:     []types.Value{values.String("id"), values.Int(123), values.Bool(true)},
			want:      values.String("id-123-true"),
		},
		{
			name:      "join with empty separator",
			separator: "",
			parts:     []types.Value{values.String("a"), values.String("b"), values.String("c")},
			want:      values.String("abc"),
		},
		{
			name:      "join single part",
			separator: ",",
			parts:     []types.Value{values.String("a")},
			want:      values.String("a"),
		},
		{
			name:      "join no parts",
			separator: ",",
			parts:     []types.Value{},
			want:      values.String(""),
		},
		{
			name:      "join with empty parts",
			separator: ",",
			parts:     []types.Value{values.String("a"), values.String(""), values.String("c")},
			want:      values.String("a,,c"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := Join(tc.separator, tc.parts...)

			if (err != nil) != tc.wantErr {
				t.Errorf("Join() error = %v, wantErr %v", err, tc.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("Join() = %v, want %v", got, tc.want)
			}
		})
	}
}
