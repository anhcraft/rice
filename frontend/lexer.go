package frontend

import (
	"errors"
	"math"
	"strconv"
	"strings"
	"unicode"
)

var (
	unexpectedCharErr     = errors.New("unexpected character")
	stringUnclosedErr     = errors.New("unclosed string")
	invalidNumberFormat   = errors.New("invalid number format")
	invalidCodepoint      = errors.New("invalid codepoint")
	invalidEscapeSequence = errors.New("invalid escape sequence")
)

var (
	identifiablePredicate = func(ch rune) bool {
		return (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || ch == '_' || (ch >= '0' && ch <= '9')
	}
	hexadecimalPredicate = func(ch rune) bool {
		return (ch >= 'a' && ch <= 'f') || (ch >= 'A' && ch <= 'F') || (ch >= '0' && ch <= '9')
	}
	asciiDigitsPredicate = func(ch rune) bool {
		return ch >= '0' && ch <= '9'
	}
	codepoint16Len = 4
	codepoint32Len = 8
	uint64Lim      = 20
	idLim          = 64
)

type Lexer struct {
	input []rune

	// the cursor on the current character
	cursor int

	// line (1-indexed)
	line int
}

func NewLexerFromString(input string) *Lexer {
	return &Lexer{input: []rune(input), line: 1}
}

func (lexer *Lexer) eof() bool {
	return lexer.cursor >= len(lexer.input)
}

func (lexer *Lexer) next() {
	if !lexer.eof() {
		ch := lexer.peek()
		if ch == '\n' { // newline character is still considered as part of the current line
			lexer.line++
		}
		lexer.cursor++
	}
}

func (lexer *Lexer) peek() rune {
	if lexer.eof() {
		return 0
	}
	return lexer.input[lexer.cursor]
}

func (lexer *Lexer) skipWhitespace() {
	for !lexer.eof() {
		ch := lexer.peek()
		if ch == ' ' || ch == '\t' || ch == '\r' || ch == '\n' {
			lexer.next()
		} else {
			break
		}
	}
}

// collectSequence collect the next sequence
func (lexer *Lexer) collectSequence(lim int, predicate func(ch rune) bool) string {
	offset := math.MaxInt

	for !lexer.eof() && lim > 0 {
		ch := lexer.peek()
		if predicate(ch) {
			offset = min(offset, lexer.cursor)
			lexer.next()
			if lexer.cursor-offset == lim {
				break
			}
		} else {
			break
		}
	}

	if offset == math.MaxInt {
		return ""
	}
	return string(lexer.input[offset:lexer.cursor])
}

// takeString takes the next inline string `"str"` (including the head and tail double quotes)
func (lexer *Lexer) takeString() (*Token, error) {
	offset := lexer.capturePos()
	lexer.next() // unquote opening

	var tok *Token
	builder := &strings.Builder{}
	escaped := false

	for !lexer.eof() {
		ch := lexer.peek()

		if ch == '\n' || (!escaped && ch == '"') {
			break // handle later
		}

		lexer.next()

		if !escaped && ch == '\\' {
			escaped = true
			continue
		} else if escaped {
			escaped = false
			if ch == '\\' {
				builder.WriteRune('\\')
				continue
			}
			if ch == '"' {
				builder.WriteRune('"')
				continue
			}
			if ch == 't' {
				builder.WriteString("\t")
				continue
			}
			if ch == 'r' {
				builder.WriteString("\r")
				continue
			}
			if ch == 'b' {
				builder.WriteString("\b")
				continue
			}
			if ch == 'f' {
				builder.WriteString("\f")
				continue
			}
			if ch == 'n' {
				builder.WriteString("\n")
				continue
			}
			if ch == 'u' { // no surrogate pairs support
				codePointStr := lexer.collectSequence(codepoint16Len, hexadecimalPredicate)
				if len(codePointStr) < codepoint16Len {
					return tok, lexer.buildSyntaxError(offset, invalidCodepoint)
				}
				codePoint, err := strconv.ParseInt(codePointStr, 16, 32)
				if err != nil {
					return tok, lexer.buildSyntaxError(offset, invalidCodepoint)
				}
				builder.WriteRune(rune(codePoint))
				continue
			}
			if ch == 'U' {
				codePointStr := lexer.collectSequence(codepoint32Len, hexadecimalPredicate)
				if len(codePointStr) < codepoint32Len {
					return tok, lexer.buildSyntaxError(offset, invalidCodepoint)
				}
				codePoint, err := strconv.ParseInt(codePointStr, 16, 32)
				if err != nil {
					return tok, lexer.buildSyntaxError(offset, invalidCodepoint)
				}
				builder.WriteRune(rune(codePoint))
				continue
			}

			return tok, lexer.buildSyntaxError(offset, invalidEscapeSequence)
		}

		builder.WriteRune(ch)
	}

	if lexer.eof() || lexer.peek() != '"' {
		return tok, lexer.buildSyntaxError(offset, stringUnclosedErr)
	}

	lexer.next() // unquote closing

	return lexer.buildToken(offset, stringLiteral, builder.String()), nil
}

// takeNumber takes the next non-negative number with pattern: ^[\d]+(\.[\d]+((e|E)(+|-)?[\d]+)?)?
// It returns int64 when two properties are both satisfied:
// - The decimal part does not exist
// - The exponent part (if exists) is non-negative
// Otherwise, float64 is returned
func (lexer *Lexer) takeNumber() (*Token, error) {
	var tok *Token
	offset := lexer.capturePos()

	// 0: next is integer part (non-negative integer)
	// 1: next is decimal part (non-negative integer)
	// 2: next is the exponent (integer)
	state := uint(0)

	// support either int64 or float64
	integerPart := uint64(0)
	decimalPart := uint64(0)
	hasDecimalPart := false // special marking because decimalPart is unsigned
	expPart := int64(0)

	/*
		State machine:
		[0] e.g. 1, 2
		[0,1] e.g. 1.3, 2.7
		[0,1,2] e.g. 1.3e8, 4.1e-9
		[0,2] e.g. 1e3, 2e-7
	*/
	for !lexer.eof() {
		if state == 0 {
			str := lexer.collectSequence(uint64Lim, asciiDigitsPredicate)
			if str == "" {
				return tok, lexer.buildSyntaxError(offset, invalidNumberFormat)
			}
			val, err := strconv.ParseUint(str, 10, 64)
			if err != nil {
				return tok, lexer.buildSyntaxError(offset, invalidNumberFormat)
			}
			integerPart = val

			if lexer.peek() == '.' {
				lexer.next()
				state = 1
				hasDecimalPart = true
				if lexer.eof() { // must have at least 1 char following the dot
					return tok, lexer.buildSyntaxError(offset, invalidNumberFormat)
				}
			} else if lexer.peek() == 'e' || lexer.peek() == 'E' {
				lexer.next()
				state = 2
				if lexer.eof() { // must have at least 1 char following the `e`
					return tok, lexer.buildSyntaxError(offset, invalidNumberFormat)
				}
			} else {
				break
			}
		} else if state == 1 {
			str := lexer.collectSequence(uint64Lim, asciiDigitsPredicate)
			if str == "" {
				return tok, lexer.buildSyntaxError(offset, invalidNumberFormat)
			}
			val, err := strconv.ParseUint(str, 10, 64)
			if err != nil {
				return tok, lexer.buildSyntaxError(offset, invalidNumberFormat)
			}
			decimalPart = val

			if lexer.peek() == 'e' || lexer.peek() == 'E' {
				lexer.next()
				state = 2
				if lexer.eof() { // must have at least 1 char following the `e`
					return tok, lexer.buildSyntaxError(offset, invalidNumberFormat)
				}
			} else {
				break
			}
		} else if state == 2 {
			sign := int64(1)
			if lexer.peek() == '-' {
				lexer.next()
				sign = -1
			} else if lexer.peek() == '+' {
				lexer.next()
			}

			str := lexer.collectSequence(uint64Lim, asciiDigitsPredicate)
			if str == "" {
				return tok, lexer.buildSyntaxError(offset, invalidNumberFormat)
			}
			val, err := strconv.ParseUint(str, 10, 64)
			if err != nil {
				return tok, lexer.buildSyntaxError(offset, invalidNumberFormat)
			}
			expPart = sign * int64(val)
			break
		} else {
			break
		}
	}

	val, err := convertNumber(integerPart, decimalPart, hasDecimalPart, expPart)
	if err != nil {
		return tok, lexer.buildSyntaxError(offset, err)
	}

	switch val.(type) {
	case int64:
		return lexer.buildToken(offset, integerLiteral, val), nil
	case float64:
		return lexer.buildToken(offset, floatLiteral, val), nil
	default:
		panic("unsupported number type")
	}
}

func (lexer *Lexer) NextToken() (*Token, error) {
	var tok *Token

	lexer.skipWhitespace()

	offset := lexer.capturePos()
	ch := lexer.peek()

	switch ch {
	case 0: // End of input
		tok = lexer.buildToken(offset, eof, nil)
	case '#':
		lexer.next()
		start := lexer.cursor
		for !lexer.eof() && lexer.peek() != '\n' {
			lexer.next()
		}
		end := lexer.cursor
		tok = lexer.buildToken(offset, comment, string(lexer.input[start:end]))
	case '+':
		lexer.next()
		if lexer.peek() == '+' {
			lexer.next()
			tok = lexer.buildToken(offset, increment, nil)
		} else {
			tok = lexer.buildToken(offset, plus, nil)
		}
	case '-':
		lexer.next()
		if lexer.peek() == '-' {
			lexer.next()
			tok = lexer.buildToken(offset, decrement, nil)
		} else {
			tok = lexer.buildToken(offset, minus, nil)
		}
	case '*':
		lexer.next()
		tok = lexer.buildToken(offset, star, nil)
	case '%':
		lexer.next()
		tok = lexer.buildToken(offset, percent, nil)
	case '/':
		lexer.next()
		tok = lexer.buildToken(offset, slash, nil)
	case '(':
		lexer.next()
		tok = lexer.buildToken(offset, lparen, nil)
	case ')':
		lexer.next()
		tok = lexer.buildToken(offset, rparen, nil)
	case '[':
		lexer.next()
		tok = lexer.buildToken(offset, lbracket, nil)
	case ']':
		lexer.next()
		tok = lexer.buildToken(offset, rbracket, nil)
	case '{':
		lexer.next()
		tok = lexer.buildToken(offset, lbrace, nil)
	case '}':
		lexer.next()
		tok = lexer.buildToken(offset, rbrace, nil)
	case ',':
		lexer.next()
		tok = lexer.buildToken(offset, comma, nil)
	case ';':
		lexer.next()
		tok = lexer.buildToken(offset, semicolon, nil)
	case '.':
		lexer.next()
		if lexer.peek() == '.' {
			lexer.next()
			if lexer.peek() == '.' {
				lexer.next()
				tok = lexer.buildToken(offset, ellipsis, nil)
			} else {
				return tok, lexer.buildSyntaxError(offset, unexpectedCharErr)
			}
		} else {
			tok = lexer.buildToken(offset, dot, nil)
		}
	case '=':
		lexer.next()
		if lexer.peek() == '=' {
			lexer.next()
			tok = lexer.buildToken(offset, equalEqual, nil)
		} else {
			tok = lexer.buildToken(offset, equal, nil)
		}
	case '!':
		lexer.next()
		if lexer.peek() == '=' {
			lexer.next()
			tok = lexer.buildToken(offset, bangEqual, nil)
		} else {
			tok = lexer.buildToken(offset, bang, nil)
		}
	case '>':
		lexer.next()
		if lexer.peek() == '=' {
			lexer.next()
			tok = lexer.buildToken(offset, gte, nil)
		} else {
			tok = lexer.buildToken(offset, gt, nil)
		}
	case '<':
		lexer.next()
		if lexer.peek() == '=' {
			lexer.next()
			tok = lexer.buildToken(offset, lte, nil)
		} else {
			tok = lexer.buildToken(offset, lt, nil)
		}
	case '&':
		lexer.next()
		if lexer.peek() == '&' {
			lexer.next()
			tok = lexer.buildToken(offset, and, nil)
		} else {
			return tok, lexer.buildSyntaxError(offset, unexpectedCharErr)
		}
	case '|':
		lexer.next()
		if lexer.peek() == '|' {
			lexer.next()
			tok = lexer.buildToken(offset, or, nil)
		} else {
			return tok, lexer.buildSyntaxError(offset, unexpectedCharErr)
		}
	case '"':
		return lexer.takeString()
	default:
		if unicode.IsDigit(ch) {
			return lexer.takeNumber()
		}

		word := lexer.collectSequence(idLim, identifiablePredicate)
		if word == "null" {
			tok = lexer.buildToken(offset, nullLiteral, nil)
			break
		} else if word == "true" {
			tok = lexer.buildToken(offset, booleanLiteral, true)
			break
		} else if word == "false" {
			tok = lexer.buildToken(offset, booleanLiteral, false)
			break
		} else if word == "if" {
			tok = lexer.buildToken(offset, ifKeyword, nil)
			break
		} else if word == "else" {
			tok = lexer.buildToken(offset, elseKeyword, nil)
		} else if word == "var" {
			tok = lexer.buildToken(offset, varKeyword, nil)
		} else if word == "const" {
			tok = lexer.buildToken(offset, constKeyword, nil)
		} else if word == "func" {
			tok = lexer.buildToken(offset, funcKeyword, nil)
		} else if word == "for" {
			tok = lexer.buildToken(offset, forKeyword, nil)
		} else if word == "in" {
			tok = lexer.buildToken(offset, inKeyword, nil)
		} else if word == "break" {
			tok = lexer.buildToken(offset, breakKeyword, nil)
		} else if word == "continue" {
			tok = lexer.buildToken(offset, continueKeyword, nil)
		} else if word == "return" {
			tok = lexer.buildToken(offset, returnKeyword, nil)
		} else if len(word) > 0 {
			tok = lexer.buildToken(offset, identifier, word)
		}
	}

	if tok == nil || tok.tokenType == illegal {
		return tok, lexer.buildSyntaxError(offset, unexpectedCharErr)
	}

	return tok, nil
}

func (lexer *Lexer) buildToken(offset Pos, tokenType TokenType, literal any) *Token {
	return &Token{
		tokenType: tokenType,
		lexeme:    lexer.input[offset.Index:lexer.cursor],
		literal:   literal,
		start:     offset,
		end:       lexer.capturePos(),
	}
}

func (lexer *Lexer) buildSyntaxError(offset Pos, err error) SyntaxError {
	return SyntaxError{Err: err, Start: offset, End: lexer.capturePos()}
}

func (lexer *Lexer) capturePos() Pos {
	return Pos{Line: lexer.line, Index: lexer.cursor}
}
