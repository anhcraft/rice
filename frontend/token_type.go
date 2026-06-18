//go:generate stringer -type TokenType
package frontend

type TokenType int

const (
	illegal TokenType = iota

	// identifiers

	identifier

	// keywords

	ifKeyword
	elseKeyword
	varKeyword
	constKeyword
	funcKeyword
	forKeyword
	inKeyword
	breakKeyword
	continueKeyword
	returnKeyword

	// literals

	integerLiteral
	floatLiteral
	stringLiteral
	booleanLiteral
	nullLiteral // "null"

	// operators

	equal      // "="
	equalEqual // "=="
	bang       // "!"
	bangEqual  // "!="
	gt         // ">"
	gte        // ">="
	lt         // "<"
	lte        // "<="
	and        // "&&"
	or         // "||"

	// punctuation

	plus      // "+"
	minus     // "-"
	star      // "*"
	slash     // "/"
	percent   // "%"
	lparen    // "("
	rparen    // ")"
	lbracket  // "["
	rbracket  // "]"
	lbrace    // "{"
	rbrace    // "}"
	comma     // ","
	semicolon // ";"
	colon     // ":"
	dot       // "."
	ellipsis  // "..."
	increment // "++"
	decrement // "--"

	// special
	eof
	comment
)
