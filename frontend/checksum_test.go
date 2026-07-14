package frontend

import (
	"testing"
)

func checksum(t *testing.T, input string) string {
	t.Helper()
	tokens, err := Tokenize(input)
	if err != nil {
		t.Fatalf("Tokenize(%q) failed: %v", input, err)
	}
	return Checksum(tokens)
}

func TestChecksumIgnoresWhitespaceAndComments(t *testing.T) {
	cases := []struct {
		name string
		a    string
		b    string
	}{
		{"spaces", "a + b", "a+b"},
		{"newlines", "a\n+\nb", "a + b"},
		{"tabs", "a\t+\tb", "a + b"},
		{"comments", "a + b # comment", "a + b"},
		{"comments with newlines", "a + b\n#comment\n", "a + b"},
		{"mixed", "  a   +   b  # trailing", "a+b"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if checksum(t, tc.a) != checksum(t, tc.b) {
				t.Errorf("expected equal checksums for %q and %q", tc.a, tc.b)
			}
		})
	}
}

func TestChecksumDistinguishesSemantics(t *testing.T) {
	cases := []struct {
		name string
		a    string
		b    string
	}{
		{"identifiers", "a + b", "x + y"},
		{"integer literals", "1 + 2", "1 + 3"},
		{"float vs int", "1", "1.0"},
		{"operators", "a == b", "a != b"},
		{"keywords", "if a { b }", "for a { b }"},
		{"booleans", "true", "false"},
		{"trailing semicolon", "a; b", "a; b;"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if checksum(t, tc.a) == checksum(t, tc.b) {
				t.Errorf("expected different checksums for %q and %q", tc.a, tc.b)
			}
		})
	}
}

func TestChecksumLiteralNormalization(t *testing.T) {
	cases := []struct {
		name string
		a    string
		b    string
	}{
		{"integer leading zeros", "01", "1"},
		{"string same content", `"hello"`, `"hello"`},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if checksum(t, tc.a) != checksum(t, tc.b) {
				t.Errorf("expected equal checksums for %q and %q", tc.a, tc.b)
			}
		})
	}
}

func TestChecksumStable(t *testing.T) {
	a := checksum(t, "func add(a, b) { return a + b }")
	b := checksum(t, "func add(a,b){return a+b}")
	c := checksum(t, "func add(a, b) {\n  return a + b\n}")

	if a != b || b != c {
		t.Errorf("expected identical checksums for formatting variants, got %q, %q, %q", a, b, c)
	}
}
