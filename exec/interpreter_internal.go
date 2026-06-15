package exec

import (
	"context"
	"errors"
	"math"

	"github.com/anhcraft/rice/exec/ast"
	"github.com/anhcraft/rice/exec/ast/opr"
	"github.com/anhcraft/rice/exec/mem"
	"github.com/anhcraft/rice/exec/types"
	"github.com/anhcraft/rice/exec/types/values"
)

func (i *Interpreter) VisitLiteralExpr(expr *ast.LiteralExpr) (types.Value, error) {
	switch e := expr.Value.(type) {
	case nil:
		return nil, nil
	case int64:
		return values.Int(e), nil
	case float64:
		return values.Float(e), nil
	case bool:
		return values.Bool(e), nil
	case string:
		return values.String(e), nil
	}
	return nil, i.throw(expr, "unsupported literal of type %T", expr.Value)
}

func (i *Interpreter) VisitIdentifierExpr(expr *ast.IdentifierExpr) (types.Value, error) {
	return values.Identifier(expr.Value), nil
}

func (i *Interpreter) VisitFuncLiteralExpr(expr *ast.FuncLiteralExpr) (types.Value, error) {
	body := expr.Body
	astNilCheck(body)

	handle := func(ctx context.Context, self *values.Func, site values.CallSite, args []types.Value) (types.Value, error) {
		n := len(args)

		if (self.Variadic() && n < self.Arity()-1) || (!self.Variadic() && n < self.Arity()) {
			return nil, i.throwCall(site, "too few arguments supplied to a call of %s", self)
		} else if !self.Variadic() && n > self.Arity() {
			return nil, i.throwCall(site, "too many arguments supplied to a call of %s", self)
		}

		// for each call, we create a fresh scope nested in the scope that the function literal was declared
		// this prevents retaining the state across multiple function calls
		closure := self.Closure().(*mem.LexicalScope)
		closure = mem.NewLexicalScope(closure)

		i.env.PushFrameWithScope(closure, site)

		i.functionDepth++
		callCtx, cancel := context.WithTimeout(ctx, i.userFuncTimeout)
		i.ctx = callCtx

		defer func() {
			cancel()
			i.ctx = ctx
			i.functionDepth--
			i.env.PopFrame() // implicit scope release
		}()

		if err := i.checkScopeThrottle(); err != nil {
			return nil, i.throw(expr, "function-literal evaluation gets interrupted").causedBy(err)
		}

		if self.Variadic() {
			lastParamIdx := self.Arity() - 1

			for k := 0; k < lastParamIdx; k++ {
				closure.Define(self.Param(k), args[k], false)
			}

			hole := make([]types.Value, n-lastParamIdx)

			for k := 0; k < len(hole); k++ {
				hole[k] = args[k+lastParamIdx]
			}

			closure.Define(self.Param(lastParamIdx), values.ListOf(hole), false)
		} else {
			for k := 0; k < self.Arity(); k++ {
				closure.Define(self.Param(k), args[k], false)
			}
		}

		res, err := i.VisitBlockExpr(body)
		var ret ReturnSignal
		if errors.As(err, &ret) {
			return ret.Result, nil
		}
		return res, err
	}

	params := make([]values.Identifier, len(expr.Params))
	for k, param := range expr.Params {
		params[k] = values.Identifier(param.Value)
	}

	// captures reference to the current scope; we permit changes to captured variables between
	// when the function literal is defined to when an actual call happens
	// this matches behavior of Go, JS, Python, etc
	closure := i.env.CurrentFrame().CurrentScope()

	return values.NewFunc(params, expr.Variadic, closure, handle), nil
}

func (i *Interpreter) VisitSelectorExpr(expr *ast.SelectorExpr) (types.Value, error) {
	astNilCheck(expr.Object)
	astNilCheck(expr.Target)

	obj, err := i.eval(expr.Object)
	if err != nil {
		return nil, i.throw(expr.Object, "cannot eval object").causedBy(err)
	}

	idx, err := i.evalc(expr.Target, ExceptId)
	if err != nil {
		return nil, i.throw(expr.Target, "cannot eval target").causedBy(err)
	}
	id := idx.(values.Identifier)

	if tries, ok := i.typeBoundFuncPkg[obj.Type()]; ok {
		if trie, ok := tries[id]; ok {
			return buildNativeFuncSet(obj, id, trie, i.nativeFuncTimeout), nil
		}
	}

	if v, ok := obj.(values.IndexedCollection); ok {
		return values.NewSelector(v, id), nil
	}

	return nil, i.throw(expr.Object, "object of type %T is not indexed collection", obj)
}

func (i *Interpreter) VisitElementAccessExpr(expr *ast.ElementAccessExpr) (types.Value, error) {
	astNilCheck(expr.Object)
	astNilCheck(expr.Index)

	obj, err := i.eval(expr.Object)
	if err != nil {
		return nil, i.throw(expr.Object, "cannot eval object").causedBy(err)
	}

	if v, ok := obj.(values.IndexedCollection); ok {
		idx, err := i.eval(expr.Index)
		if err != nil {
			return nil, i.throw(expr.Index, "cannot eval index").causedBy(err)
		}

		return values.NewSelector(v, idx), nil
	}

	return nil, i.throw(expr.Object, "object of type %T is not indexed collection", obj)
}

func (i *Interpreter) VisitCallExpr(expr *ast.CallExpr) (types.Value, error) {
	astNilCheck(expr.Callee)

	if err := i.checkContext(); err != nil {
		return nil, i.throw(expr, "call evaluation gets interrupted").causedBy(err)
	}

	i.profiler.Start(expr)
	defer func() {
		i.profiler.End()
	}()

	callee, err := i.eval(expr.Callee)
	if err != nil {
		return nil, i.throw(expr.Callee, "cannot eval callee").causedBy(err)
	}

	if callable, ok := callee.(values.Callable); ok {
		args := make([]types.Value, 0, len(expr.Arguments))

		for j, arg := range expr.Arguments {
			val, err := i.eval(arg.Value)

			if err != nil {
				return nil, i.throw(arg.Value, "cannot eval args[%d]", j).causedBy(err)
			}

			if arg.Spread {
				if coll, ok := val.(values.Collection); ok {
					for value := range coll.Iterate() {
						args = append(args, value)
					}
				} else {
					return nil, i.throw(arg.Value, "cannot spread args[%d] because type %T is not collection", j, val)
				}
			} else {
				args = append(args, val)
			}
		}

		return callable.Call(i.ctx, values.CallSite{Caller: "CallExpr", StartPos: expr.StartPos(), EndPos: expr.EndPos()}, args)
	}

	return nil, i.throw(expr.Callee, "callee of type %T is not callable", callee)
}

func (i *Interpreter) VisitDeclareStmt(expr *ast.DeclareStmt) (types.Value, error) {
	astNilCheck(expr.Target)
	astNilCheck(expr.Value)

	target, err := i.evalc(expr.Target, ExceptId|ExceptSel)
	if err != nil {
		return nil, i.throw(expr.Target, "cannot eval declaration target").causedBy(err)
	}

	val, err := i.eval(expr.Value)
	if err != nil {
		return nil, i.throw(expr.Value, "cannot eval declaration value").causedBy(err)
	}

	if id, ok := target.(values.Identifier); ok {
		ok = i.env.Define(id, val, expr.Const)
		if !ok {
			return nil, i.throw(expr, "cannot redeclare %q", id)
		}
	} else {
		return nil, i.throw(expr.Target, "target of type %T is not declarable", target)
	}

	return nil, nil
}

func (i *Interpreter) VisitForStmt(expr *ast.ForStmt) (types.Value, error) {
	i.env.EnterScope()
	i.profiler.Start(expr)
	defer func() {
		i.profiler.End()
		i.env.ExitScope()
	}()
	if err := i.checkScopeThrottle(); err != nil {
		return nil, i.throw(expr, "for-loop evaluation gets interrupted").causedBy(err)
	}

	// init is executed in a new lexical scope
	if _, err := i.eval(expr.Init); err != nil {
		return nil, i.throw(expr.Init, "cannot eval for-loop init").causedBy(err)
	}

	return func() (types.Value, error) {
		var err error

		i.loopDepth++
		defer func() {
			i.loopDepth--
		}()

		for {
			if err = i.checkContext(); err != nil {
				return nil, i.throw(expr, "iteration gets interrupted").causedBy(err)
			}

			if expr.Cond != nil {
				cond, err := i.eval(expr.Cond)
				if err != nil {
					return nil, i.throw(expr.Cond, "cannot eval for-loop condition").causedBy(err)
				}

				if b, err := values.AsBool(cond); err != nil {
					return nil, i.throw(expr.Cond, "cannot implicitly convert condition of type %T to Bool", cond).causedBy(err)
				} else if !b {
					break
				}
			}

			// new lexical scope per each iteration (including the body + post iteration)
			if _, err = i.execBlockExpr(expr.Body, nil, func() (types.Value, error) {
				return i.eval(expr.Post)
			}); err != nil {
				var continueSignal ContinueSignal
				if errors.As(err, &continueSignal) {
					continue
				}
				var breakSignal BreakSignal
				if errors.As(err, &breakSignal) {
					break
				}
				return nil, i.throw(expr, "iteration gets interrupted").causedBy(err)
			}
		}

		return nil, nil
	}()
}

func (i *Interpreter) VisitForInStmt(expr *ast.ForInStmt) (types.Value, error) {
	astNilCheck(expr.Key)
	astNilCheck(expr.Value)
	astNilCheck(expr.Body)

	i.env.EnterScope()
	i.profiler.Start(expr)
	defer func() {
		i.profiler.End()
		i.env.ExitScope()
	}()
	if err := i.checkScopeThrottle(); err != nil {
		return nil, i.throw(expr, "for-in evaluation gets interrupted").causedBy(err)
	}

	key, err := i.evalc(expr.Key, ExceptId|ExceptSel)
	if err != nil {
		return nil, i.throw(expr.Key, "cannot eval for-in key").causedBy(err)
	}

	if id, ok := key.(values.Identifier); ok {
		val, err := i.eval(expr.Value)
		if err != nil {
			return nil, i.throw(expr.Value, "cannot eval for-in value").causedBy(err)
		}

		if c, ok := val.(values.Collection); ok {
			return func() (types.Value, error) {
				var err error

				i.loopDepth++
				defer func() {
					i.loopDepth--
				}()

				for value := range c.Iterate() {
					if err = i.checkContext(); err != nil {
						return nil, i.throw(expr, "iteration gets interrupted").causedBy(err)
					}

					// new lexical scope per each iteration (including pre-iteration + the body)
					if _, err = i.execBlockExpr(expr.Body, func() (types.Value, error) {
						ok = i.env.Define(id, value, false)
						if !ok {
							return nil, i.throw(expr, "cannot declare for..in variable %q of type %T", id, value)
						}
						return nil, nil
					}, nil); err != nil {
						return nil, i.throw(expr, "iteration gets interrupted").causedBy(err)
					}
				}

				return nil, nil
			}()
		} else {
			return nil, i.throw(expr.Value, "value of type %T is not collection", val)
		}
	} else {
		return nil, i.throw(expr.Key, "key of type %T is not identifier", key)
	}
}

func (i *Interpreter) VisitControlStmt(expr *ast.ControlStmt) (types.Value, error) {
	switch expr.Op {
	case opr.Return:
		val, err := i.eval(expr.Value)
		if err != nil {
			return nil, i.throw(expr.Value, "cannot eval return-value").causedBy(err)
		}
		return nil, ReturnSignal{Result: val}
	case opr.Break:
		return nil, BreakSignal{}
	case opr.Continue:
		return nil, ContinueSignal{}
	default:
		panic("unimplemented")
	}
}

func (i *Interpreter) VisitIncDecStmt(expr *ast.IncDecStmt) (types.Value, error) {
	astNilCheck(expr.Target)

	target, err := i.evalc(expr.Target, ExceptId|ExceptSel)
	if err != nil {
		return nil, i.throw(expr.Target, "cannot eval target").causedBy(err)
	}

	incrementer := func(num types.Value, delta values.Int) (types.Value, bool) {
		switch n := num.(type) {
		case values.Int:
			return n + delta, true
		case values.Float:
			return n + values.Float(delta), true
		}
		return nil, false
	}

	delta := values.Int(1)
	if expr.Op == opr.Dec {
		delta = values.Int(-1)
	}

	var res types.Value

	if sel, ok := target.(values.Selector); ok {
		val, err := sel.Get()
		if err != nil {
			return nil, i.throw(expr.Target, "cannot read value from target selector").causedBy(err)
		}

		newVal, ok := incrementer(val, delta)
		if !ok {
			return nil, i.throw(expr.Target, "value of type %T is not numeric", val)
		}

		if err = sel.Put(newVal); err != nil {
			return nil, i.throw(expr, "cannot assign value to target selector").causedBy(err)
		}

		if expr.Pre {
			res = newVal
		} else {
			res = val
		}
	} else if id, ok := target.(values.Identifier); ok {
		val, ok := i.env.Get(id)
		if !ok {
			return nil, i.throw(expr.Target, "unknown identifier %q", id)
		}

		newVal, ok := incrementer(val, delta)
		if !ok {
			return nil, i.throw(expr.Target, "value of type %T is not numeric", val)
		}

		if err = i.env.Assign(id, newVal); err != nil {
			return nil, i.throw(expr, "cannot assign value to target identifier").causedBy(err)
		}

		if expr.Pre {
			res = newVal
		} else {
			res = val
		}
	} else {
		return nil, i.throw(expr.Target, "target of type %T is not assignable", target)
	}

	return res, nil
}

func (i *Interpreter) VisitAssignExpr(expr *ast.AssignExpr) (types.Value, error) {
	astNilCheck(expr.Target)
	astNilCheck(expr.Value)

	target, err := i.evalc(expr.Target, ExceptId|ExceptSel)
	if err != nil {
		return nil, i.throw(expr.Target, "cannot eval assignment target").causedBy(err)
	}

	val, err := i.eval(expr.Value)
	if err != nil {
		return nil, i.throw(expr.Value, "cannot eval assignment value").causedBy(err)
	}

	if sel, ok := target.(values.Selector); ok {
		err = sel.Put(val)
		if err != nil {
			return nil, i.throw(expr, "cannot assign value to target selector").causedBy(err)
		}
	} else if id, ok := target.(values.Identifier); ok {
		err = i.env.Assign(id, val)
		if err != nil {
			return nil, i.throw(expr, "cannot assign value to target identifier").causedBy(err)
		}
	} else {
		return nil, i.throw(expr.Target, "target of type %T is not assignable", target)
	}

	return val, nil
}

func (i *Interpreter) VisitBinaryExpr(expr *ast.BinaryExpr) (types.Value, error) {
	astNilCheck(expr.Left)
	astNilCheck(expr.Right)

	var left types.Value
	var isLeftPrimitive bool
	var right types.Value
	var isRightPrimitive bool
	var err error

	if left, err = i.eval(expr.Left); err != nil {
		return nil, i.throw(expr.Left, "cannot eval left operand").causedBy(err)
	} else if _, ok := left.(values.Primitive); ok {
		isLeftPrimitive = true
	}

	if right, err = i.eval(expr.Right); err != nil {
		return nil, i.throw(expr.Right, "cannot eval right operand").causedBy(err)
	} else if _, ok := right.(values.Primitive); ok {
		isRightPrimitive = true
	}

	if left == nil || right == nil || !isLeftPrimitive || !isRightPrimitive {
		if expr.Op == opr.Eq {
			return values.Bool(left == right), nil
		} else if expr.Op == opr.Neq {
			return values.Bool(left != right), nil
		}
		return nil, i.throw(expr, "cannot eval %T %s %T", left, expr.Op, right)
	}

	left, right, err = values.ConvertPrimitiveImplicitly(left.(values.Primitive), right.(values.Primitive))

	if err != nil {
		return nil, i.throw(expr, "failed implicit type conversion of %T and %T", left, right).causedBy(err)
	}

	if expr.Op == opr.Eq {
		return values.Bool(left == right), nil
	} else if expr.Op == opr.Neq {
		return values.Bool(left != right), nil
	}

	switch left.(type) {
	case values.Int:
		switch expr.Op {
		case opr.Sum:
			return left.(values.Int) + right.(values.Int), nil
		case opr.Sub:
			return left.(values.Int) - right.(values.Int), nil
		case opr.Multi:
			return left.(values.Int) * right.(values.Int), nil
		case opr.Div:
			return left.(values.Int) / right.(values.Int), nil
		case opr.Rem:
			return left.(values.Int) % right.(values.Int), nil
		case opr.Gt:
			return values.Bool(left.(values.Int) > right.(values.Int)), nil
		case opr.Gte:
			return values.Bool(left.(values.Int) >= right.(values.Int)), nil
		case opr.Lt:
			return values.Bool(left.(values.Int) < right.(values.Int)), nil
		case opr.Lte:
			return values.Bool(left.(values.Int) <= right.(values.Int)), nil
		default:
			goto fail
		}
	case values.Float:
		switch expr.Op {
		case opr.Sum:
			return left.(values.Float) + right.(values.Float), nil
		case opr.Sub:
			return left.(values.Float) - right.(values.Float), nil
		case opr.Multi:
			return left.(values.Float) * right.(values.Float), nil
		case opr.Div:
			return left.(values.Float) / right.(values.Float), nil
		case opr.Rem:
			return values.Float(math.Mod(float64(left.(values.Float)), float64(right.(values.Float)))), nil
		case opr.Gt:
			return values.Bool(left.(values.Float) > right.(values.Float)), nil
		case opr.Gte:
			return values.Bool(left.(values.Float) >= right.(values.Float)), nil
		case opr.Lt:
			return values.Bool(left.(values.Float) < right.(values.Float)), nil
		case opr.Lte:
			return values.Bool(left.(values.Float) <= right.(values.Float)), nil
		default:
			goto fail
		}
	case values.Bool:
		switch expr.Op {
		case opr.And:
			return left.(values.Bool) && right.(values.Bool), nil
		case opr.Or:
			return left.(values.Bool) || right.(values.Bool), nil
		default:
			goto fail
		}
	case values.String:
		switch expr.Op {
		case opr.Sum:
			return left.(values.String) + right.(values.String), nil
		case opr.Gt:
			return values.Bool(left.(values.String) > right.(values.String)), nil
		case opr.Gte:
			return values.Bool(left.(values.String) >= right.(values.String)), nil
		case opr.Lt:
			return values.Bool(left.(values.String) < right.(values.String)), nil
		case opr.Lte:
			return values.Bool(left.(values.String) <= right.(values.String)), nil
		default:
			goto fail
		}
	}

fail: // T(left)=T(right)
	return nil, i.throw(expr, "cannot eval %T %s %T", left, expr.Op, right)
}

func (i *Interpreter) VisitUnaryExpr(expr *ast.UnaryExpr) (types.Value, error) {
	astNilCheck(expr.Right)

	right, err := i.eval(expr.Right)
	if err != nil {
		return nil, i.throw(expr.Right, "cannot eval unary operand").causedBy(err)
	}

	if v, ok := right.(values.Bool); ok && expr.Op == opr.Inv {
		return !v, nil
	}
	if v, ok := right.(values.Int); ok && expr.Op == opr.Neg {
		return -v, nil
	}
	if v, ok := right.(values.Float); ok && expr.Op == opr.Neg {
		return -v, nil
	}

	return nil, i.throw(expr, "cannot match operator %q to operand of type %T", expr.Op, expr.Right)
}

func (i *Interpreter) execBlockExpr(expr *ast.BlockExpr, pre func() (types.Value, error), post func() (types.Value, error)) (types.Value, error) {
	if err := i.checkContext(); err != nil {
		return nil, i.throw(expr, "block evaluation gets interrupted").causedBy(err)
	}

	i.env.EnterScope()
	i.profiler.Start(expr)
	defer func() {
		i.profiler.End()
		i.env.ExitScope()
	}()
	if err := i.checkScopeThrottle(); err != nil {
		return nil, i.throw(expr, "block evaluation gets interrupted").causedBy(err)
	}

	var res types.Value
	var err error

	if pre != nil {
		res, err = pre()
		if err != nil {
			return nil, i.throw(expr, "cannot eval pre-block").causedBy(err)
		}
	}

	for idx, stmt := range expr.Statements {
		res, err = i.eval(stmt)

		if err != nil {
			if _, ok := err.(Signal); ok {
				return nil, err
			}
			return nil, i.throw(stmt, "cannot eval block statement #%d", idx+1).causedBy(err)
		}
	}

	if post != nil {
		res, err = post()
		if err != nil {
			return nil, i.throw(expr, "cannot eval post-block").causedBy(err)
		}
	}

	return res, nil
}

func (i *Interpreter) VisitBlockExpr(expr *ast.BlockExpr) (types.Value, error) {
	return i.execBlockExpr(expr, nil, nil)
}

func (i *Interpreter) VisitIfExpr(expr *ast.IfExpr) (types.Value, error) {
	astNilCheck(expr.Condition)

	condition, err := i.eval(expr.Condition)
	if err != nil {
		return nil, i.throw(expr.Condition, "cannot eval if-condition").causedBy(err)
	}

	if cond, ok := condition.(values.Bool); ok {
		if cond {
			return i.VisitBlockExpr(expr.ThenBranch)
		} else if expr.ElseBranch != nil {
			return i.eval(expr.ElseBranch)
		}

		return values.Bool(false), nil
	}

	return nil, i.throw(expr.Condition, "expect if-condition to be Bool but got %T", condition)
}
