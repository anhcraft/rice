package ast

import (
	"github.com/anhcraft/rice/exec/types"
)

type Visitor interface {
	VisitDeclareStmt(expr *DeclareStmt) (types.Value, error)
	VisitForStmt(expr *ForStmt) (types.Value, error)
	VisitForInStmt(expr *ForInStmt) (types.Value, error)
	VisitControlStmt(expr *ControlStmt) (types.Value, error)
	VisitIncDecStmt(expr *IncDecStmt) (types.Value, error)
	VisitAssignExpr(expr *AssignExpr) (types.Value, error)
	VisitBinaryExpr(expr *BinaryExpr) (types.Value, error)
	VisitUnaryExpr(expr *UnaryExpr) (types.Value, error)
	VisitLiteralExpr(expr *LiteralExpr) (types.Value, error)
	VisitBlockExpr(expr *BlockExpr) (types.Value, error)
	VisitIfExpr(expr *IfExpr) (types.Value, error)
	VisitCallExpr(expr *CallExpr) (types.Value, error)
	VisitElementAccessExpr(expr *ElementAccessExpr) (types.Value, error)
	VisitSelectorExpr(expr *SelectorExpr) (types.Value, error)
	VisitIdentifierExpr(expr *IdentifierExpr) (types.Value, error)
	VisitFuncLiteralExpr(expr *FuncLiteralExpr) (types.Value, error)
}
