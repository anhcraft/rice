//go:generate stringer -type OpType
package opr

type OpType int

const (
	// Unary

	Neg OpType = iota
	Inv
	Inc
	Dec
	Spread

	// Binary

	Sum
	Sub
	Multi
	Div
	Rem
	Eq
	Neq
	Gt  // ">"
	Gte // ">="
	Lt  // "<"
	Lte // "<="
	And // "&&"
	Or  // "||"

	// Flow control

	Return
	Continue
	Break
)
