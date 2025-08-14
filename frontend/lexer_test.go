package frontend

import (
	"errors"
	"math"
	"reflect"
	"testing"
)

func TestLexer(t *testing.T) {
	t.Run("collectSequence", collectSequence)
	t.Run("takeString", testTakeString)
	t.Run("takeNumber", testTakeNumber)
	t.Run("NextToken", testNextToken)
}

func collectSequence(t *testing.T) {
	lookup := func(p string) func(ch rune) bool {
		switch p {
		case "identifiable":
			return identifiablePredicate
		case "hexadecimal":
			return hexadecimalPredicate
		case "asciiDigits":
			return asciiDigitsPredicate
		}
		panic("unsupported")
	}

	testCases := loadTestCases[struct {
		Name            string `json:"name,omitempty"`
		Input           string `json:"input,omitempty"`
		Limit           int    `json:"limit,omitempty"`
		Predicate       string `json:"predicate,omitempty"`
		WantSeq         string `json:"wantSeq,omitempty"`
		WantFinalCursor int    `json:"wantFinalCursor,omitempty"`
		WantFinalLine   int    `json:"wantFinalLine,omitempty"`
	}](t, "lexer-collect-sequence.json")

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			lexer := NewLexerFromString(tc.Input)
			got := lexer.collectSequence(tc.Limit, lookup(tc.Predicate))

			if tc.WantSeq != got {
				t.Errorf("collectSequence() for input %q, predicate %q, limit %d\nwant seq: %q, got: %q",
					tc.Input, tc.Predicate, tc.Limit, tc.WantSeq, got)
			}

			if tc.WantFinalCursor != lexer.cursor {
				t.Errorf("collectSequence() for input %q, predicate %q, limit %d\nwant final cursor: %q, got: %q",
					tc.Input, tc.Predicate, tc.Limit, tc.WantFinalCursor, lexer.cursor)
			}

			if tc.WantFinalLine != lexer.line {
				t.Errorf("collectSequence() for input %q, predicate %q, limit %d\nwant final line: %q, got: %q",
					tc.Input, tc.Predicate, tc.Limit, tc.WantFinalLine, lexer.line)
			}
		})
	}
}

func testTakeString(t *testing.T) {
	lookupErr := func(name string) error {
		if name == "stringUnclosedErr" {
			return stringUnclosedErr
		} else if name == "invalidCodepoint" {
			return invalidCodepoint
		} else if name == "invalidEscapeSequence" {
			return invalidEscapeSequence
		}
		panic("unknown error " + name + "; invalid testdata?")
	}

	testCases := loadTestCases[struct {
		Name        string `json:"name,omitempty"`
		Input       string `json:"input,omitempty"`
		WantLexeme  string `json:"wantLexeme,omitempty"`
		WantLiteral any    `json:"wantLiteral,omitempty"`
		WantError   string `json:"wantError,omitempty"`
	}](t, "lexer-take-string.json")

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			lexer := NewLexerFromString(tc.Input)
			tok, err := lexer.takeString()

			if err != nil {
				if tc.WantError == "" {
					t.Errorf("takeString() for input %q\nexpect none error, got: %q", tc.Input, err)
				} else if !errors.Is(err, lookupErr(tc.WantError)) {
					t.Errorf("takeString() for input %q\nerror want: %q, got %q", tc.Input, tc.WantError, err)
				}

				return
			}

			if tc.WantError != "" {
				t.Errorf("takeString() for input %q\nerror want: %q, got none", tc.Input, tc.WantError)
			}

			if tc.WantLexeme != string(tok.lexeme) {
				t.Errorf("takeString() for input %q\nlexeme want: %q, got: %q",
					tc.Input, tc.WantLexeme, string(tok.lexeme))
			}

			if tc.WantLiteral != tok.literal {
				t.Errorf("takeString() for input %q\nliteral want: %q, got: %q",
					tc.Input, tc.WantLiteral, tok.literal)
			}
		})
	}
}

func testTakeNumber(t *testing.T) {
	lookupErr := func(name string) error {
		if name == "invalidNumberFormat" {
			return invalidNumberFormat
		} else if name == "errIntegerOverflow" {
			return errIntegerOverflow
		} else if name == "errFloat64Overflow" {
			return errFloat64Overflow
		}
		panic("unknown error " + name + "; invalid testdata?")
	}

	type testCaseDef struct {
		Name            string `json:"name,omitempty"`
		Input           string `json:"input,omitempty"`
		WantLexeme      string `json:"wantLexeme,omitempty"`
		WantLiteral     any    `json:"wantLiteral,omitempty"`
		WantLiteralType any    `json:"wantLiteralType,omitempty"`
		WantError       string `json:"wantError,omitempty"`
	}

	testCases := loadTestCases[testCaseDef](t, "lexer-take-number.json")

	// additional test cases for input outside JSON number range (float64)
	hardcodedCases := []testCaseDef{
		{
			Name:            "Boundary - Max int64",
			Input:           "9223372036854775807",
			WantLexeme:      "9223372036854775807",
			WantLiteral:     int64(9223372036854775807),
			WantLiteralType: "int64",
		},
		{
			Name:            "Boundary - Max int64 with exponent",
			Input:           "9223372036854775807e0",
			WantLexeme:      "9223372036854775807e0",
			WantLiteral:     int64(9223372036854775807),
			WantLiteralType: "int64",
		},
		{
			Name:      "Overflow - int64",
			Input:     "9223372036854775808",
			WantError: "errIntegerOverflow",
		},
		{
			Name:      "Overflow - int64 via exponent",
			Input:     "9223372036854775807e1",
			WantError: "errIntegerOverflow",
		},
		{
			Name:      "Overflow - float64",
			Input:     "1.8e308",
			WantError: "errFloat64Overflow",
		},
	}

	testCases = append(testCases, hardcodedCases...)

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			lexer := NewLexerFromString(tc.Input)
			tok, err := lexer.takeNumber()

			if err != nil {
				if tc.WantError == "" {
					t.Errorf("takeNumber() for input %q\nexpect none error, got: %q", tc.Input, err)
				} else if !errors.Is(err, lookupErr(tc.WantError)) {
					t.Errorf("takeNumber() for input %q\nerror want: %q, got %q", tc.Input, tc.WantError, err)
				}

				return
			}

			if tc.WantError != "" {
				t.Errorf("takeNumber() for input %q\nerror want: %q, got none", tc.Input, tc.WantError)
			}

			if tc.WantLexeme != string(tok.lexeme) {
				t.Errorf("takeNumber() for input %q\nlexeme want: %q, got: %q",
					tc.Input, tc.WantLexeme, string(tok.lexeme))
			}

			if tc.WantLiteralType != "" && tc.WantLiteralType != reflect.TypeOf(tok.literal).Name() {
				t.Errorf("takeNumber() for input %q\nliteral type want: %q, got: %q",
					tc.Input, tc.WantLiteralType, reflect.TypeOf(tok.literal))
			}

			if tc.WantLiteralType == "float64" && math.Abs(tc.WantLiteral.(float64)-tok.literal.(float64)) > 1e-9 {
				t.Errorf("takeNumber() for input %q\nfloat64 literal want: %f, got: %f",
					tc.Input, tc.WantLiteral, tok.literal)
			}

			if tc.WantLiteralType == "int64" {
				var wantLit int64

				if reflect.TypeOf(tc.WantLiteral).Name() == "int64" {
					wantLit = tc.WantLiteral.(int64)
				} else {
					wantLit = int64(tc.WantLiteral.(float64))
				}

				if wantLit != tok.literal.(int64) {
					t.Errorf("takeNumber() for input %q\nint64 literal want: %d, got: %d",
						tc.Input, tc.WantLiteral, tok.literal)
				}
			}
		})
	}
}

func testNextToken(t *testing.T) {
	testCases := []struct {
		name        string
		input       string
		wantType    TokenType
		wantLexeme  string
		wantLiteral any
		wantError   error
	}{
		// --- Basic and Whitespace ---
		{
			name:        "Empty input",
			input:       "",
			wantType:    eof,
			wantLexeme:  "",
			wantLiteral: nil,
			wantError:   nil,
		},
		{
			name:        "Whitespace only",
			input:       "   \t\n\r ",
			wantType:    eof,
			wantLexeme:  "",
			wantLiteral: nil,
			wantError:   nil,
		},
		{
			name:        "Leading whitespace",
			input:       "  +",
			wantType:    plus,
			wantLexeme:  "+",
			wantLiteral: nil,
			wantError:   nil,
		},

		// --- Single-Character Tokens ---
		{name: "Plus", input: "+", wantType: plus, wantLexeme: "+"},
		{name: "Increment", input: "++", wantType: increment, wantLexeme: "++"},
		{name: "Minus", input: "-", wantType: minus, wantLexeme: "-"},
		{name: "Decrement", input: "--", wantType: decrement, wantLexeme: "--"},
		{name: "Star", input: "*", wantType: star, wantLexeme: "*"},
		{name: "Slash", input: "/", wantType: slash, wantLexeme: "/"},
		{name: "Percent", input: "%", wantType: percent, wantLexeme: "%"},
		{name: "Left Paren", input: "(", wantType: lparen, wantLexeme: "("},
		{name: "Right Paren", input: ")", wantType: rparen, wantLexeme: ")"},
		{name: "Left Bracket", input: "[", wantType: lbracket, wantLexeme: "["},
		{name: "Right Bracket", input: "]", wantType: rbracket, wantLexeme: "]"},
		{name: "Left Brace", input: "{", wantType: lbrace, wantLexeme: "{"},
		{name: "Right Brace", input: "}", wantType: rbrace, wantLexeme: "}"},
		{name: "Comma", input: ",", wantType: comma, wantLexeme: ","},
		{name: "Semicolon", input: ";", wantType: semicolon, wantLexeme: ";"},
		{name: "Dot", input: ".", wantType: dot, wantLexeme: "."},
		{name: "Ellipsis", input: "...", wantType: ellipsis, wantLexeme: "..."},

		// --- One-or-Two-Character Tokens ---
		{name: "Equal", input: "=", wantType: equal, wantLexeme: "="},
		{name: "Equal Equal", input: "==", wantType: equalEqual, wantLexeme: "=="},
		{name: "Bang", input: "!", wantType: bang, wantLexeme: "!"},
		{name: "Bang Equal", input: "!=", wantType: bangEqual, wantLexeme: "!="},
		{name: "Greater Than", input: ">", wantType: gt, wantLexeme: ">"},
		{name: "Greater Than or Equal", input: ">=", wantType: gte, wantLexeme: ">="},
		{name: "Less Than", input: "<", wantType: lt, wantLexeme: "<"},
		{name: "Less Than or Equal", input: "<=", wantType: lte, wantLexeme: "<="},

		// --- Mandatory Two-Character Tokens & Errors ---
		{name: "Logical And", input: "&&", wantType: and, wantLexeme: "&&"},
		{
			name:      "Stray Ampersand",
			input:     "&",
			wantError: unexpectedCharErr,
		},
		{name: "Logical Or", input: "||", wantType: or, wantLexeme: "||"},
		{
			name:      "Stray Pipe",
			input:     "|",
			wantError: unexpectedCharErr,
		},

		// --- Comments ---
		{
			name:        "Comment until newline",
			input:       "# this is a comment\n+",
			wantType:    comment,
			wantLexeme:  "# this is a comment",
			wantLiteral: " this is a comment",
		},
		{
			name:        "Comment until EOF",
			input:       "# this is a comment at EOF",
			wantType:    comment,
			wantLexeme:  "# this is a comment at EOF",
			wantLiteral: " this is a comment at EOF",
		},
		{
			name:        "Empty comment",
			input:       "#\n",
			wantType:    comment,
			wantLexeme:  "#",
			wantLiteral: "",
		},

		// --- Delegated Tokenization (String and Number) ---
		{
			name:        "String literal",
			input:       `"hello world"`,
			wantType:    stringLiteral,
			wantLexeme:  `"hello world"`,
			wantLiteral: "hello world",
		},
		{
			name:        "Integer literal",
			input:       "12345",
			wantType:    integerLiteral,
			wantLexeme:  "12345",
			wantLiteral: int64(12345),
		},
		{
			name:        "Float literal",
			input:       "123.45",
			wantType:    floatLiteral,
			wantLexeme:  "123.45",
			wantLiteral: 123.45,
		},

		// --- Keywords ---
		{
			name:        "Null keyword",
			input:       "null",
			wantType:    nullLiteral,
			wantLexeme:  "null",
			wantLiteral: nil,
		},
		{
			name:        "True keyword",
			input:       "true",
			wantType:    booleanLiteral,
			wantLexeme:  "true",
			wantLiteral: true,
		},
		{
			name:        "False keyword",
			input:       "false",
			wantType:    booleanLiteral,
			wantLexeme:  "false",
			wantLiteral: false,
		},
		{
			name:       "If keyword",
			input:      "if",
			wantType:   ifKeyword,
			wantLexeme: "if",
		},
		{
			name:       "Else keyword",
			input:      "else",
			wantType:   elseKeyword,
			wantLexeme: "else",
		},
		{
			name:       "Const keyword",
			input:      "const",
			wantType:   constKeyword,
			wantLexeme: "const",
		},
		{
			name:       "Var keyword",
			input:      "var",
			wantType:   varKeyword,
			wantLexeme: "var",
		},
		{
			name:       "For keyword",
			input:      "for",
			wantType:   forKeyword,
			wantLexeme: "for",
		},
		{
			name:       "In keyword",
			input:      "in",
			wantType:   inKeyword,
			wantLexeme: "in",
		},
		{
			name:       "Break keyword",
			input:      "break",
			wantType:   breakKeyword,
			wantLexeme: "break",
		},
		{
			name:       "Continue keyword",
			input:      "continue",
			wantType:   continueKeyword,
			wantLexeme: "continue",
		},
		{
			name:       "Return keyword",
			input:      "return",
			wantType:   returnKeyword,
			wantLexeme: "return",
		},
		{
			name:       "Func keyword",
			input:      "func",
			wantType:   funcKeyword,
			wantLexeme: "func",
		},

		// --- Identifiers ---
		{
			name:        "Simple identifier",
			input:       "variableName",
			wantType:    identifier,
			wantLexeme:  "variableName",
			wantLiteral: "variableName",
		},
		{
			name:        "Identifier with numbers",
			input:       "var123",
			wantType:    identifier,
			wantLexeme:  "var123",
			wantLiteral: "var123",
		},
		{
			name:        "Identifier starting with a keyword",
			input:       "falsehood",
			wantType:    identifier,
			wantLexeme:  "falsehood",
			wantLiteral: "falsehood",
		},
		{
			name:        "Identifier with underscore",
			input:       "my_identifier",
			wantType:    identifier,
			wantLexeme:  "my_identifier",
			wantLiteral: "my_identifier",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			lexer := NewLexerFromString(tc.input)
			tok, err := lexer.NextToken()

			if err != nil {
				if tc.wantError == nil {
					t.Errorf("takeNumber() for input %q\nexpect none error, got: %q", tc.input, err)
				} else if !errors.Is(err, tc.wantError) {
					t.Errorf("takeNumber() for input %q\nerror want: %q, got %q", tc.input, tc.wantError, err)
				}

				return
			}

			if tok == nil {
				t.Fatalf("NextToken() for input %q returned a nil token", tc.input)
			}

			if tc.wantType != tok.tokenType {
				t.Errorf("NextToken() for input %q\ntype want: %q, got: %q",
					tc.input, tc.wantType, tok.tokenType)
			}

			if tc.wantLexeme != string(tok.lexeme) {
				t.Errorf("NextToken() for input %q\nlexeme want: %q, got: %q",
					tc.input, tc.wantLexeme, string(tok.lexeme))
			}

			if !reflect.DeepEqual(tc.wantLiteral, tok.literal) {
				t.Errorf("NextToken() for input %q\nliteral want: %v (%T), got: %v (%T)",
					tc.input, tc.wantLiteral, tc.wantLiteral, tok.literal, tok.literal)
			}
		})
	}
}
