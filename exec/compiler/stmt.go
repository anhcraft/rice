package compiler

import (
	"fmt"

	"github.com/anhcraft/rice/exec/ast"
	"github.com/anhcraft/rice/exec/ast/opr"
	"github.com/anhcraft/rice/exec/vm"
)

func producesValue(s ast.Stmt) bool {
	switch s.(type) {
	case *ast.DeclareStmt, *ast.ForStmt, *ast.ForInStmt, *ast.ControlStmt:
		return false
	default:
		return true
	}
}

func (c *compiler) compileTop(stmts []ast.Stmt) {
	if len(stmts) == 0 {
		c.emitPushNull()
		c.emit(vm.OpReturn, 0)
		return
	}
	for i, s := range stmts {
		keep := i == len(stmts)-1
		c.compileStmt(s, keep, true)
	}
	last := stmts[len(stmts)-1]
	if !producesValue(last) {
		if _, ok := last.(*ast.ControlStmt); !ok {
			c.emitPushNull()
		}
	}
	if _, ok := last.(*ast.ControlStmt); !ok {
		c.emit(vm.OpReturn, 0)
	}
}

func (c *compiler) compileStmt(s ast.Stmt, keepValue, topLevel bool) {
	if s == nil {
		if keepValue {
			c.emitPushNull()
		}
		return
	}
	switch n := s.(type) {
	case *ast.DeclareStmt:
		c.compileDeclare(n, topLevel)
		if keepValue {
			c.emitPushNull()
		}
	case *ast.ForStmt:
		c.compileFor(n)
		if keepValue {
			c.emitPushNull()
		}
	case *ast.ForInStmt:
		c.compileForIn(n)
		if keepValue {
			c.emitPushNull()
		}
	case *ast.ControlStmt:
		c.compileControl(n)
	case *ast.IncDecStmt:
		c.compileIncDec(n)
		if !keepValue {
			c.emit(vm.OpPop, 0)
		}
	default:
		if ex, ok := s.(ast.Expr); ok {
			c.compileExpr(ex)
			if !keepValue {
				c.emit(vm.OpPop, 0)
			}
		} else {
			c.errorf(s, "unsupported statement %T", s)
		}
	}
}

func (c *compiler) compileDeclare(d *ast.DeclareStmt, topLevel bool) {
	c.withSpan(d.Value, "cannot eval declaration value", func() {
		c.compileExpr(d.Value)
	})
	if topLevel && len(c.fn.scopes) == 1 && c.fn.parent == nil {
		op := vm.OpDefineGlobal
		if d.Const {
			op = vm.OpDefineConstGlobal
		}
		c.emit(op, int32(c.addString(d.Target.Value)))
		return
	}
	sc := &c.fn.scopes[len(c.fn.scopes)-1]
	if sc.bound[d.Target.Value] {
		c.errorf(d, "cannot redeclare %q", d.Target.Value)
	}
	sc.bound[d.Target.Value] = true
	kind, idx := c.resolveFn(c.fn, d.Target.Value)
	if kind != resLocal {
		idx = c.declare(d.Target.Value)
	}
	op := vm.OpDefineLocal
	if d.Const {
		op = vm.OpDefineConstLocal
	}
	c.emit(op, int32(idx))
}

func (c *compiler) compileBlock(b *ast.BlockExpr) {
	if b == nil {
		c.emitPushNull()
		return
	}
	prev := c.pushSpan(b, "")
	c.emitEnter(c.addHotspot("Block", b, vm.HotspotOpensScope))
	c.beginScope()
	c.hoist(b.Statements)
	if len(b.Statements) == 0 {
		c.emitPushNull()
	} else {
		for i, s := range b.Statements {
			keep := i == len(b.Statements)-1
			c.withSpan(s, fmt.Sprintf("cannot eval block statement #%d", i+1), func() {
				c.compileStmt(s, keep, false)
			})
		}
		last := b.Statements[len(b.Statements)-1]
		if !producesValue(last) {
			if _, ok := last.(*ast.ControlStmt); !ok {
				c.emitPushNull()
			}
		}
	}
	c.endScope()
	c.emitLeave()
	c.curSpan = prev
}

func (c *compiler) compileFor(n *ast.ForStmt) {
	prev := c.pushSpan(n, "")
	headerStay := c.fn.regionDepth + 1
	c.emitEnter(c.addHotspot("ForLoop", n, vm.HotspotOpensScope))
	c.beginScope()

	loop := &loopCtx{stayRegions: headerStay, iterSlot: -1}
	perDecl, _ := n.Init.(*ast.DeclareStmt)
	if perDecl != nil {
		loop.perIter = true
		loop.iterSlot = c.declare(perDecl.Target.Value)
		c.compileDeclare(perDecl, false)
	} else if n.Init != nil {
		c.withSpan(n.Init, "cannot eval for-loop init", func() {
			c.compileStmt(n.Init, false, false)
		})
	}
	c.fn.loops = append(c.fn.loops, loop)

	condL := len(c.fn.code)
	hasCond := n.Cond != nil
	exitJ := 0
	if hasCond {
		c.withSpan(n.Cond, "cannot eval for-loop condition", func() {
			c.compileExpr(n.Cond)
		})
		exitJ = c.emit(vm.OpJumpIfFalsey, 0)
		c.emit(vm.OpPop, 0)
	}

	if n.Body != nil {
		c.compileBlock(n.Body)
		c.emit(vm.OpPop, 0)
	}

	contL := len(c.fn.code)
	c.patchList(loop.continueJumps, contL)
	if loop.perIter && loop.iterSlot >= 0 && c.fn.captured[loop.iterSlot] {
		c.emit(vm.OpCloseUpvalue, int32(loop.iterSlot))
	}
	if n.Post != nil {
		c.withSpan(n.Post, "cannot eval for-loop post", func() {
			c.compileStmt(n.Post, false, false)
		})
	}
	c.emit(vm.OpJumpBackward, int32(c.calcBackward(condL)))

	if hasCond {
		c.patchJumpTo(exitJ, len(c.fn.code))
		c.emit(vm.OpPop, 0)
	}
	c.patchList(loop.breakJumps, len(c.fn.code))

	c.fn.loops = c.fn.loops[:len(c.fn.loops)-1]
	c.endScope()
	c.emitLeave()
	c.curSpan = prev
}

func (c *compiler) compileForIn(n *ast.ForInStmt) {
	prev := c.pushSpan(n, "")
	headerStay := c.fn.regionDepth + 1
	c.emitEnter(c.addHotspot("ForInLoop", n, vm.HotspotOpensScope))
	c.beginScope()
	keySlot := c.declare(n.Key.Value)
	c.emitPushNull()
	c.emit(vm.OpDefineLocal, int32(keySlot))

	c.withSpan(n.Value, "cannot eval for-in value", func() {
		c.compileExpr(n.Value)
	})
	c.emit(vm.OpBegin, 0)

	loop := &loopCtx{stayRegions: headerStay, iterSlot: keySlot, perIter: true}
	c.fn.loops = append(c.fn.loops, loop)

	head := len(c.fn.code)
	endJ := c.emit(vm.OpJumpIfEnd, 0)
	c.emit(vm.OpPointer, 0)
	c.emit(vm.OpStoreLocal, int32(keySlot))

	if n.Body != nil {
		c.compileBlock(n.Body)
		c.emit(vm.OpPop, 0)
	}

	contL := len(c.fn.code)
	c.patchList(loop.continueJumps, contL)
	if c.fn.captured[keySlot] {
		c.emit(vm.OpCloseUpvalue, int32(keySlot))
	}
	c.emit(vm.OpIncrementIndex, 0)
	c.emit(vm.OpJumpBackward, int32(c.calcBackward(head)))

	exitL := len(c.fn.code)
	c.patchJumpTo(endJ, exitL)
	c.patchList(loop.breakJumps, exitL)
	c.emit(vm.OpEnd, 0)

	c.fn.loops = c.fn.loops[:len(c.fn.loops)-1]
	c.endScope()
	c.emitLeave()
	c.curSpan = prev
}

func (c *compiler) compileControl(n *ast.ControlStmt) {
	switch n.Op {
	case opr.Return:
		if c.fn.parent == nil {
			c.errorf(n, "return outside function")
			c.emitPushNull()
			c.emit(vm.OpReturn, 0)
			return
		}
		if n.Value != nil {
			c.withSpan(n.Value, "cannot eval return-value", func() {
				c.compileExpr(n.Value)
			})
		} else {
			c.emitPushNull()
		}
		c.emit(vm.OpReturn, 0)
	case opr.Break:
		if len(c.fn.loops) == 0 {
			c.errorf(n, "break outside loop")
			return
		}
		lp := c.fn.loops[len(c.fn.loops)-1]
		c.closeIter(lp)
		c.emitJumpLeaves(c.fn.regionDepth - lp.stayRegions)
		j := c.emit(vm.OpJump, 0)
		lp.breakJumps = append(lp.breakJumps, j)
	case opr.Continue:
		if len(c.fn.loops) == 0 {
			c.errorf(n, "continue outside loop")
			return
		}
		lp := c.fn.loops[len(c.fn.loops)-1]
		c.closeIter(lp)
		c.emitJumpLeaves(c.fn.regionDepth - lp.stayRegions)
		j := c.emit(vm.OpJump, 0)
		lp.continueJumps = append(lp.continueJumps, j)
	default:
		c.errorf(n, "unknown control %v", n.Op)
	}
}

func (c *compiler) closeIter(lp *loopCtx) {
	if lp.perIter && lp.iterSlot >= 0 && c.fn.captured[lp.iterSlot] {
		c.emit(vm.OpCloseUpvalue, int32(lp.iterSlot))
	}
}

func (c *compiler) compileIncDec(n *ast.IncDecStmt) {
	delta := int32(1)
	if n.Op == opr.Dec {
		delta = -1
	}
	switch t := n.Target.(type) {
	case *ast.IdentifierExpr:
		c.compileIdentLoad(t.Value)
		if !n.Pre {
			c.emit(vm.OpDup, 0)
		}
		c.emit(vm.OpIncrement, delta)
		if n.Pre {
			c.emit(vm.OpDup, 0)
		}
		c.compileIdentStore(t.Value)
	case *ast.ElementAccessExpr:
		c.withSpan(t.Object, "cannot eval object", func() { c.compileExpr(t.Object) })
		c.withSpan(t.Index, "cannot eval index", func() { c.compileExpr(t.Index) })
		if n.Pre {
			c.emit(vm.OpDup2, 0)
			c.emit(vm.OpGetIndex, 0)
			c.emit(vm.OpIncrement, delta)
			c.emit(vm.OpSetIndex, 0)
		} else {
			tmpObj := c.allocTemp()
			tmpIdx := c.allocTemp()
			c.emit(vm.OpStoreLocal, int32(tmpIdx))
			c.emit(vm.OpStoreLocal, int32(tmpObj))
			c.emit(vm.OpLoadLocal, int32(tmpObj))
			c.emit(vm.OpLoadLocal, int32(tmpIdx))
			c.emit(vm.OpGetIndex, 0)
			c.emit(vm.OpDup, 0)
			c.emit(vm.OpIncrement, delta)
			c.emit(vm.OpLoadLocal, int32(tmpObj))
			c.emit(vm.OpLoadLocal, int32(tmpIdx))
			c.emit(vm.OpRot3, 0)
			c.emit(vm.OpSetIndex, 0)
			c.emit(vm.OpPop, 0)
		}
	case *ast.SelectorExpr:
		c.withSpan(t.Object, "cannot eval object", func() { c.compileExpr(t.Object) })
		name := int32(c.addString(t.Target.Value))
		if n.Pre {
			c.emit(vm.OpDup, 0)
			c.emit(vm.OpLoadField, name)
			c.emit(vm.OpIncrement, delta)
			c.emit(vm.OpStoreField, name)
		} else {
			tmpObj := c.allocTemp()
			tmpNew := c.allocTemp()
			c.emit(vm.OpDup, 0)
			c.emit(vm.OpStoreLocal, int32(tmpObj))
			c.emit(vm.OpLoadField, name)
			c.emit(vm.OpDup, 0)
			c.emit(vm.OpIncrement, delta)
			c.emit(vm.OpStoreLocal, int32(tmpNew))
			c.emit(vm.OpLoadLocal, int32(tmpObj))
			c.emit(vm.OpLoadLocal, int32(tmpNew))
			c.emit(vm.OpStoreField, name)
			c.emit(vm.OpPop, 0)
		}
	default:
		c.errorf(n, "target of type %T is not assignable", n.Target)
		c.emitPushNull()
	}
}

func (c *compiler) compileExpr(e ast.Expr) {
	if e == nil {
		c.emitPushNull()
		return
	}
	switch n := e.(type) {
	case *ast.LiteralExpr:
		c.compileLiteral(n)
	case *ast.IdentifierExpr:
		c.compileIdentLoad(n.Value)
	case *ast.AssignExpr:
		c.compileAssign(n)
	case *ast.BinaryExpr:
		c.compileBinary(n)
	case *ast.UnaryExpr:
		c.compileUnary(n)
	case *ast.CallExpr:
		c.compileCall(n)
	case *ast.ElementAccessExpr:
		c.withSpan(n.Object, "cannot eval object", func() { c.compileExpr(n.Object) })
		c.withSpan(n.Index, "cannot eval index", func() { c.compileExpr(n.Index) })
		c.emit(vm.OpGetIndex, 0)
	case *ast.SelectorExpr:
		c.withSpan(n.Object, "cannot eval object", func() { c.compileExpr(n.Object) })
		c.emit(vm.OpLoadField, int32(c.addString(n.Target.Value)))
	case *ast.BlockExpr:
		c.compileBlock(n)
	case *ast.IfExpr:
		c.compileIf(n)
	case *ast.FuncLiteralExpr:
		c.compileFunc(n)
	case *ast.ObjectLiteralExpr:
		c.compileObject(n)
	case *ast.ArrayLiteralExpr:
		c.compileArray(n)
	default:
		c.errorf(e, "unsupported expression %T", e)
		c.emitPushNull()
	}
}

func (c *compiler) compileLiteral(n *ast.LiteralExpr) {
	switch v := n.Value.(type) {
	case nil:
		c.emitPushNull()
	case int64:
		c.emit(vm.OpPush, int32(c.addInt(v)))
	case float64:
		c.emit(vm.OpPush, int32(c.addFloat(v)))
	case bool:
		c.emitPushBool(v)
	case string:
		c.emit(vm.OpPush, int32(c.addString(v)))
	default:
		c.errorf(n, "unsupported literal of type %T", n.Value)
		c.emitPushNull()
	}
}

func (c *compiler) compileIdentLoad(name string) {
	kind, idx := c.resolveFn(c.fn, name)
	switch kind {
	case resLocal:
		c.emit(vm.OpLoadLocal, int32(idx))
	case resUpvalue:
		c.emit(vm.OpLoadUpvalue, int32(idx))
	default:
		c.emit(vm.OpLoadGlobal, int32(idx))
	}
}

func (c *compiler) compileIdentStore(name string) {
	kind, idx := c.resolveFn(c.fn, name)
	switch kind {
	case resLocal:
		c.emit(vm.OpStoreLocal, int32(idx))
	case resUpvalue:
		c.emit(vm.OpStoreUpvalue, int32(idx))
	default:
		c.emit(vm.OpStoreGlobal, int32(idx))
	}
}

func (c *compiler) compileAssign(n *ast.AssignExpr) {
	c.withSpan(n.Value, "cannot eval assignment value", func() {
		switch t := n.Target.(type) {
		case *ast.IdentifierExpr:
			c.compileExpr(n.Value)
			c.emit(vm.OpDup, 0)
			c.compileIdentStore(t.Value)
		case *ast.ElementAccessExpr:
			c.withSpan(t.Object, "cannot eval object", func() { c.compileExpr(t.Object) })
			c.withSpan(t.Index, "cannot eval index", func() { c.compileExpr(t.Index) })
			c.compileExpr(n.Value)
			c.emit(vm.OpSetIndex, 0)
		case *ast.SelectorExpr:
			c.withSpan(t.Object, "cannot eval object", func() { c.compileExpr(t.Object) })
			c.compileExpr(n.Value)
			c.emit(vm.OpStoreField, int32(c.addString(t.Target.Value)))
		default:
			c.errorf(n.Target, "target of type %T is not assignable", n.Target)
			c.compileExpr(n.Value)
		}
	})
}

func (c *compiler) compileBinary(n *ast.BinaryExpr) {
	if n.Op == opr.And || n.Op == opr.Or {
		c.compileLogical(n)
		return
	}
	c.withSpan(n.Left, "cannot eval left operand", func() { c.compileExpr(n.Left) })
	c.withSpan(n.Right, "cannot eval right operand", func() { c.compileExpr(n.Right) })
	if n.Op == opr.Neq {
		c.emit(vm.OpEqual, 0)
		c.emit(vm.OpNot, 0)
		return
	}
	c.emit(binOp(n.Op), 0)
}

func (c *compiler) compileLogical(n *ast.BinaryExpr) {
	c.withSpan(n.Left, "cannot eval left operand", func() { c.compileExpr(n.Left) })
	jmp := 0
	if n.Op == opr.And {
		jmp = c.emit(vm.OpJumpIfBoolFalse, 0)
	} else {
		jmp = c.emit(vm.OpJumpIfBoolTrue, 0)
	}
	c.withSpan(n.Right, "cannot eval right operand", func() { c.compileExpr(n.Right) })
	if n.Op == opr.And {
		c.emit(vm.OpAnd, 0)
	} else {
		c.emit(vm.OpOr, 0)
	}
	c.patchJump(jmp)
}

func (c *compiler) compileUnary(n *ast.UnaryExpr) {
	c.withSpan(n.Right, "cannot eval unary operand", func() { c.compileExpr(n.Right) })
	switch n.Op {
	case opr.Neg:
		c.emit(vm.OpNegate, 0)
	case opr.Inv:
		c.emit(vm.OpNot, 0)
	default:
		c.errorf(n, "unknown unary %v", n.Op)
	}
}

func (c *compiler) compileCall(n *ast.CallExpr) {
	prev := c.pushSpan(n, "")
	c.emitEnter(c.addHotspot("Call", n, 0))
	c.withSpan(n.Callee, "cannot eval callee", func() { c.compileExpr(n.Callee) })
	spread := false
	for _, a := range n.Arguments {
		if a.Spread {
			spread = true
			break
		}
	}
	if spread {
		c.emit(vm.OpArgs, 0)
		for j, a := range n.Arguments {
			c.withSpan(a.Value, fmt.Sprintf("cannot eval args[%d]", j), func() {
				c.compileExpr(a.Value)
			})
			if a.Spread {
				c.emit(vm.OpSpread, 0)
			} else {
				c.emit(vm.OpArg, 0)
			}
		}
		c.emit(vm.OpCallArgs, 0)
	} else {
		for j, a := range n.Arguments {
			c.withSpan(a.Value, fmt.Sprintf("cannot eval args[%d]", j), func() {
				c.compileExpr(a.Value)
			})
		}
		c.emit(vm.OpCall, int32(len(n.Arguments)))
	}
	c.emitLeave()
	c.curSpan = prev
}

func (c *compiler) compileIf(n *ast.IfExpr) {
	c.withSpan(n.Condition, "cannot eval if-condition", func() {
		c.compileExpr(n.Condition)
	})
	elseJ := c.emit(vm.OpJumpIfFalse, 0)
	c.emit(vm.OpPop, 0)
	c.compileBlock(n.ThenBranch)
	endJ := c.emit(vm.OpJump, 0)
	c.patchJump(elseJ)
	c.emit(vm.OpPop, 0)
	if n.ElseBranch != nil {
		c.compileExpr(n.ElseBranch)
	} else {
		c.emitPushBool(false)
	}
	c.patchJump(endJ)
}

func (c *compiler) compileFunc(n *ast.FuncLiteralExpr) {
	saved := c.fn
	savedSpan := c.curSpan
	c.beginFunction(n.Params, n.Variadic, saved)
	c.compileBlock(n.Body)
	c.emit(vm.OpReturn, 0)
	idx := c.fn.index
	c.finishFunction()
	c.fn = saved
	c.curSpan = savedSpan
	c.emit(vm.OpClosure, int32(idx))
}

func (c *compiler) compileObject(n *ast.ObjectLiteralExpr) {
	for i, e := range n.Entries {
		switch k := e.Key.(type) {
		case *ast.IdentifierExpr:
			c.emit(vm.OpPush, int32(c.addString(k.Value)))
		case *ast.LiteralExpr:
			if s, ok := k.Value.(string); ok {
				c.emit(vm.OpPush, int32(c.addString(s)))
			} else {
				c.errorf(e.Key, "object literal key must be string, got %T", k.Value)
				c.emitPushNull()
			}
		default:
			c.errorf(e.Key, "object literal key must be identifier or string literal, got %T", e.Key)
			c.emitPushNull()
		}
		c.withSpan(e.Value, fmt.Sprintf("cannot eval object literal value at entry #%d", i), func() {
			c.compileExpr(e.Value)
		})
	}
	c.emit(vm.OpMap, int32(len(n.Entries)))
}

func (c *compiler) compileArray(n *ast.ArrayLiteralExpr) {
	for i, e := range n.Elements {
		c.withSpan(e, fmt.Sprintf("cannot eval array literal element at index #%d", i), func() {
			c.compileExpr(e)
		})
	}
	c.emit(vm.OpArray, int32(len(n.Elements)))
}

func binOp(op opr.OpType) vm.Opcode {
	switch op {
	case opr.Sum:
		return vm.OpAdd
	case opr.Sub:
		return vm.OpSubtract
	case opr.Multi:
		return vm.OpMultiply
	case opr.Div:
		return vm.OpDivide
	case opr.Rem:
		return vm.OpModulo
	case opr.Eq:
		return vm.OpEqual
	case opr.Gt:
		return vm.OpMore
	case opr.Gte:
		return vm.OpMoreOrEqual
	case opr.Lt:
		return vm.OpLess
	case opr.Lte:
		return vm.OpLessOrEqual
	case opr.And:
		return vm.OpAnd
	case opr.Or:
		return vm.OpOr
	default:
		return vm.OpPop
	}
}
