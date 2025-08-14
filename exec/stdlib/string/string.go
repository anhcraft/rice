package string

import (
	"fmt"
	"github.com/anhcraft/rice/exec/fun"
	"github.com/anhcraft/rice/exec/stdlib"
	"github.com/anhcraft/rice/exec/types"
	"github.com/anhcraft/rice/exec/types/values"
	"strings"
)

var Functions = fun.FunctionPackage{
	"format":    {stdlib.Define(Format)},
	"trim":      {stdlib.Define(Trim)},
	"toUpper":   {stdlib.Define(ToUpper)},
	"toLower":   {stdlib.Define(ToLower)},
	"include":   {stdlib.Define(Include)},
	"index":     {stdlib.Define(Index)},
	"lastIndex": {stdlib.Define(LastIndex)},
	"substr":    {stdlib.Define(Substr), stdlib.Define(Substr0)},
	"split":     {stdlib.Define(Split)},
	"join":      {stdlib.Define(Join)},
}

// Format formats according to a format specifier.
func Format(format values.String, args ...any) (types.Value, error) {
	return values.String(fmt.Sprintf(string(format), args...)), nil
}

// Trim removes leading and trailing white space from a string.
func Trim(s values.String) (types.Value, error) {
	return values.String(strings.TrimSpace(string(s))), nil
}

// ToUpper converts a string to its uppercase representation.
func ToUpper(s values.String) (types.Value, error) {
	return values.String(strings.ToUpper(string(s))), nil
}

// ToLower converts a string to its lowercase representation.
func ToLower(s values.String) (types.Value, error) {
	return values.String(strings.ToLower(string(s))), nil
}

// Include checks if a substring exists in the given string.
func Include(str values.String, sub values.String) (types.Value, error) {
	return values.Bool(strings.Contains(string(str), string(sub))), nil
}

// Index searches for the first occurrence of the substring in the given string; otherwise, -1 if not exist.
func Index(str values.String, sub values.String) (types.Value, error) {
	return values.Int(strings.Index(string(str), string(sub))), nil
}

// LastIndex searches for the last occurrence of the substring in the given string; otherwise, -1 if not exist.
func LastIndex(str values.String, sub values.String) (types.Value, error) {
	return values.Int(strings.LastIndex(string(str), string(sub))), nil
}

// Substr extracts a substring from a string; end is exclusive.
func Substr(s values.String, start values.Int, end values.Int) (types.Value, error) {
	runes := []rune(s)
	runeLen := len(runes)

	if start < 0 || int(start) > runeLen {
		return values.String(""), nil
	}

	if end < 0 {
		end = 0
	}
	if int(end) > runeLen {
		end = values.Int(runeLen)
	}

	if start > end {
		return values.String(""), nil
	}

	return values.String(runes[start:end]), nil
}

func Substr0(s values.String, start values.Int) (types.Value, error) {
	return Substr(s, start, values.Int(len(s)))
}

// Split divides a string into an array of substrings based on a separator.
func Split(s values.String, separator values.String) (types.Value, error) {
	parts := strings.Split(string(s), string(separator))
	result := make([]types.Value, len(parts))
	for i, p := range parts {
		result[i] = values.String(p)
	}
	return values.ListOf(result), nil
}

// Join concatenates elements of an array into a single string, separated by a given separator.
func Join(separator values.String, parts ...types.Value) (types.Value, error) {
	if len(parts) == 0 {
		return values.String(""), nil
	}

	sep := string(separator)
	stringParts := make([]string, len(parts))

	for i, p := range parts {
		stringParts[i] = string(values.AsString(p))
	}

	return values.String(strings.Join(stringParts, sep)), nil
}
