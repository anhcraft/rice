package ast

import (
	"fmt"
	"github.com/anhcraft/rice/exec/ast/opr"
	"github.com/anhcraft/rice/exec/types"
	"strings"
)

// Note: use prefix notation for stringification; avoid ambiguity

type Node interface {
	fmt.Stringer
	node()
}

type Stmt interface {
	Node
	Accept(v Visitor) (types.Value, error)
	StartPos() Pos
	EndPos() Pos
}

type SimpleStmt interface {
	Stmt
	simple()
}

type Expr interface {
	SimpleStmt
}

// BlockLike allows omitting semicolon following the closing brace
type BlockLike interface {
	blockLike()
}

// Hotspot get recorded in profiling
type Hotspot interface {
	Stmt
	Label() string
}

type BaseNode struct {
	Start Pos
	End   Pos // exclusive
}

func (b BaseNode) node() {}

func (b BaseNode) StartPos() Pos {
	return b.Start
}

func (b BaseNode) EndPos() Pos {
	return b.End
}

// --- Dummy nodes ---

var _ Stmt = Root{}

type Root struct {
	Stmt
	pos Pos
}

func (r Root) String() string {
	return "root"
}

func (r Root) node() {}

func (r Root) Accept(_ Visitor) (types.Value, error) {
	return nil, nil
}

func (r Root) StartPos() Pos {
	return r.pos
}

func (r Root) EndPos() Pos {
	return r.pos
}

func (r Root) Label() string {
	return "Root"
}

var _ Expr = InvalidExpr{}
var Invalid = InvalidExpr{}

type InvalidExpr struct {
}

func (i InvalidExpr) String() string {
	return "Invalid"
}

func (i InvalidExpr) node() {}

func (i InvalidExpr) Accept(_ Visitor) (types.Value, error) {
	return nil, nil
}

func (i InvalidExpr) StartPos() Pos {
	return Pos{}
}

func (i InvalidExpr) EndPos() Pos {
	return Pos{}
}

func (i InvalidExpr) simple() {}

// --- Statement ---

type DeclareStmt struct {
	BaseNode
	Const  bool
	Target *IdentifierExpr
	Value  Expr
}

func (d *DeclareStmt) simple() {}

func (d *DeclareStmt) Accept(v Visitor) (types.Value, error) { return v.VisitDeclareStmt(d) }

func (d *DeclareStmt) String() string {
	if d.Const {
		return fmt.Sprintf("(Declare-const %s %s)", d.Target, d.Value)
	}
	return fmt.Sprintf("(Declare %s %s)", d.Target, d.Value)
}

type ForStmt struct {
	BaseNode
	Init SimpleStmt
	Cond Expr
	Post SimpleStmt
	Body *BlockExpr
}

func (f *ForStmt) Label() string {
	return "ForLoop"
}

func (f *ForStmt) blockLike() {}

func (f *ForStmt) Accept(v Visitor) (types.Value, error) { return v.VisitForStmt(f) }

func (f *ForStmt) String() string {
	var builder strings.Builder
	builder.WriteString("(For ")

	if f.Init != nil {
		builder.WriteString(f.Init.String())
	} else {
		builder.WriteString("()")
	}

	builder.WriteString(" ")

	if f.Cond != nil {
		builder.WriteString(f.Cond.String())
	} else {
		builder.WriteString("()")
	}

	builder.WriteString(" ")

	if f.Post != nil {
		builder.WriteString(f.Post.String())
	} else {
		builder.WriteString("()")
	}

	builder.WriteString(" ")
	builder.WriteString(f.Body.String())
	builder.WriteString(")")

	return builder.String()
}

type ForInStmt struct {
	BaseNode
	Key   *IdentifierExpr
	Value Expr
	Body  *BlockExpr
}

func (f *ForInStmt) Label() string {
	return "ForInLoop"
}

func (f *ForInStmt) blockLike() {}

func (f *ForInStmt) Accept(v Visitor) (types.Value, error) { return v.VisitForInStmt(f) }

func (f *ForInStmt) String() string {
	return fmt.Sprintf("(ForIn %s %s %s)", f.Key, f.Value, f.Body)
}

type ControlStmt struct {
	BaseNode
	OpPos Pos
	Op    opr.OpType

	// Value used for "break <value>", "return <value>"
	Value Expr
}

func (c *ControlStmt) Accept(v Visitor) (types.Value, error) { return v.VisitControlStmt(c) }

func (c *ControlStmt) String() string {
	if c.Value == nil {
		return fmt.Sprintf("(Control %s)", c.Op)
	}
	return fmt.Sprintf("(Control %s %s)", c.Op, c.Value)
}

type IncDecStmt struct {
	BaseNode
	Pre    bool
	OpPos  Pos
	Op     opr.OpType
	Target Expr
}

func (i *IncDecStmt) simple() {}

func (i *IncDecStmt) Accept(v Visitor) (types.Value, error) { return v.VisitIncDecStmt(i) }

func (i *IncDecStmt) String() string {
	nodeName := "Post-incDec"
	if i.Pre {
		nodeName = "Pre-incDec"
	}
	return fmt.Sprintf("(%s %s %s)", nodeName, i.Op, i.Target)
}

// --- Expressions ---
// NOTE: Stringify must use Prefix notation

type AssignExpr struct {
	BaseNode
	Target Expr
	Value  Expr
}

func (a *AssignExpr) simple() {}

func (a *AssignExpr) Accept(v Visitor) (types.Value, error) { return v.VisitAssignExpr(a) }

func (a *AssignExpr) String() string {
	return fmt.Sprintf("(Assign %s %s)", a.Target, a.Value)
}

type BinaryExpr struct {
	BaseNode
	Left  Expr
	OpPos Pos
	Op    opr.OpType
	Right Expr
}

func (b *BinaryExpr) simple() {}

func (b *BinaryExpr) Accept(v Visitor) (types.Value, error) { return v.VisitBinaryExpr(b) }

func (b *BinaryExpr) String() string {
	return fmt.Sprintf("(%s %s %s)", b.Op, b.Left, b.Right)
}

type UnaryExpr struct {
	BaseNode
	OpPos Pos
	Op    opr.OpType
	Right Expr
}

func (u *UnaryExpr) simple() {}

func (u *UnaryExpr) Accept(v Visitor) (types.Value, error) { return v.VisitUnaryExpr(u) }

func (u *UnaryExpr) String() string {
	return fmt.Sprintf("(%s %s)", u.Op, u.Right)
}

type CallExpr struct {
	BaseNode
	Callee    Expr
	Arguments []CallExprArg
}

type CallExprArg struct {
	Value  Expr
	Spread bool
}

func (c *CallExpr) Label() string {
	return "Call"
}

func (c *CallExpr) simple() {}

func (c *CallExpr) Accept(v Visitor) (types.Value, error) { return v.VisitCallExpr(c) }

func (c *CallExpr) String() string {
	var builder strings.Builder
	builder.WriteString("(Call ")
	builder.WriteString(c.Callee.String())
	for _, arg := range c.Arguments {
		builder.WriteString(" ")
		if arg.Spread {
			builder.WriteString("...")
		}
		builder.WriteString(arg.Value.String())
	}
	builder.WriteString(")")
	return builder.String()
}

type ElementAccessExpr struct {
	BaseNode
	Object Expr
	Index  Expr
}

func (e *ElementAccessExpr) simple() {}

func (e *ElementAccessExpr) Accept(v Visitor) (types.Value, error) {
	return v.VisitElementAccessExpr(e)
}

func (e *ElementAccessExpr) String() string {
	return fmt.Sprintf("(Get %s %s)", e.Object, e.Index)
}

type SelectorExpr struct {
	BaseNode
	Object Expr
	Target *IdentifierExpr
}

func (s *SelectorExpr) simple() {}

func (s *SelectorExpr) Accept(v Visitor) (types.Value, error) { return v.VisitSelectorExpr(s) }

func (s *SelectorExpr) String() string {
	return fmt.Sprintf("(Sel %s %s)", s.Object, s.Target)
}

// --- Structural Nodes ---

type BlockExpr struct {
	BaseNode
	Statements []Stmt
}

func (b *BlockExpr) Label() string {
	return "Block"
}

func (b *BlockExpr) simple() {}

func (b *BlockExpr) blockLike() {}

func (b *BlockExpr) Accept(v Visitor) (types.Value, error) { return v.VisitBlockExpr(b) }

func (b *BlockExpr) String() string {
	var builder strings.Builder
	builder.WriteString("(Block")
	for _, expr := range b.Statements {
		builder.WriteString(" ")
		builder.WriteString(expr.String())
	}
	builder.WriteString(")")
	return builder.String()
}

type IfExpr struct {
	BaseNode
	Condition  Expr
	ThenBranch *BlockExpr
	ElseBranch Expr
}

func (i *IfExpr) simple() {}

func (i *IfExpr) blockLike() {}

func (i *IfExpr) Accept(v Visitor) (types.Value, error) { return v.VisitIfExpr(i) }

func (i *IfExpr) String() string {
	if i.ElseBranch == nil {
		return fmt.Sprintf("(If %s %s)", i.Condition, i.ThenBranch)
	}
	return fmt.Sprintf("(If %s %s %s)", i.Condition, i.ThenBranch, i.ElseBranch)
}

// --- Primary & Literal Nodes ---

type LiteralExpr struct {
	BaseNode
	Value any
}

func (l *LiteralExpr) simple() {}

func (l *LiteralExpr) Accept(v Visitor) (types.Value, error) { return v.VisitLiteralExpr(l) }

func (l *LiteralExpr) String() string {
	switch v := l.Value.(type) {
	case string:
		return fmt.Sprintf("(Str %q)", v) // quoting is enough
	case bool:
		return fmt.Sprintf("(Bool %t)", v)
	case int64:
		return fmt.Sprintf("(Int %d)", v)
	case float64:
		return fmt.Sprintf("(Float %g)", v)
	case nil:
		return "(Null)"
	default:
		return fmt.Sprintf("%#v", v)
	}
}

type FuncLiteralExpr struct {
	BaseNode
	Params   []*IdentifierExpr
	Body     *BlockExpr
	Variadic bool
}

func (f *FuncLiteralExpr) blockLike() {}

func (f *FuncLiteralExpr) simple() {}

func (f *FuncLiteralExpr) Accept(v Visitor) (types.Value, error) {
	return v.VisitFuncLiteralExpr(f)
}

func (f *FuncLiteralExpr) String() string {
	var builder strings.Builder
	builder.WriteString("(Func")
	for _, arg := range f.Params {
		builder.WriteString(" ")
		builder.WriteString(arg.String())
	}
	if f.Variadic {
		builder.WriteString("...")
	}
	builder.WriteString(" ")
	builder.WriteString(f.Body.String())
	builder.WriteString(")")
	return builder.String()
}

type IdentifierExpr struct {
	BaseNode
	Value string
}

func (i *IdentifierExpr) simple() {}

func (i *IdentifierExpr) Accept(v Visitor) (types.Value, error) { return v.VisitIdentifierExpr(i) }

func (i *IdentifierExpr) String() string {
	return fmt.Sprintf("(Id %s)", i.Value)
}

type ObjectLiteralEntry struct {
	Key   Expr
	Value Expr
}

type ObjectLiteralExpr struct {
	BaseNode
	Entries []ObjectLiteralEntry
}

func (o *ObjectLiteralExpr) simple()    {}
func (o *ObjectLiteralExpr) blockLike() {}

func (o *ObjectLiteralExpr) Accept(v Visitor) (types.Value, error) {
	return v.VisitObjectLiteralExpr(o)
}

func (o *ObjectLiteralExpr) String() string {
	var builder strings.Builder
	builder.WriteString("(Object")
	for _, entry := range o.Entries {
		builder.WriteString(" ")
		builder.WriteString(entry.Key.String())
		builder.WriteString(" ")
		builder.WriteString(entry.Value.String())
	}
	builder.WriteString(")")
	return builder.String()
}
