package frontend

import (
	"fmt"
	"github.com/anhcraft/rice/exec/ast"
)

type Pos struct {
	Index int // the index in terms of runes (0-indexed)
	Line  int // the ordinal in terms of line (1-indexed)
}

func (p Pos) ast() ast.Pos {
	return ast.Pos{Index: p.Index, Line: p.Line}
}

type Token struct {
	tokenType TokenType
	lexeme    []rune
	literal   any
	start     Pos
	end       Pos
}

func (t *Token) Type() TokenType {
	return t.tokenType
}

func (t *Token) Lexeme() []rune {
	return t.lexeme
}

func (t *Token) Literal() any {
	return t.literal
}

type SyntaxError struct {
	Err   error
	Start Pos
	End   Pos // (exclusive)
}

func (e SyntaxError) Error() string {
	return fmt.Sprintf("%v at %d:%d line %d:%d", e.Err, e.Start.Index, e.End.Index, e.Start.Line, e.End.Line)
}

func (e SyntaxError) Unwrap() error {
	return e.Err
}
