package compiler

import (
	"fmt"

	"github.com/anhcraft/rice/exec/ast"
	"github.com/anhcraft/rice/exec/types"
	"github.com/anhcraft/rice/exec/types/values"
	"github.com/anhcraft/rice/exec/vm"
)

type Error struct {
	Msg   string
	Start ast.Pos
	End   ast.Pos
}

func (e Error) Error() string {
	return fmt.Sprintf("%s (L%d-L%d)", e.Msg, e.Start.Line, e.End.Line)
}

const (
	resLocal = iota
	resUpvalue
	resGlobal
)

type uvKey struct {
	isLocal bool
	index   uint16
}

type blockScope struct {
	names map[string]int
	bound map[string]bool
}

type loopCtx struct {
	stayRegions   int
	breakJumps    []int
	continueJumps []int
	iterSlot      int
	perIter       bool
}

type fnState struct {
	parent      *fnState
	index       int
	arity       int
	variadic    bool
	code        []vm.Opcode
	args        []int32
	spans       []uint16
	slotNames   []string
	scopes      []blockScope
	upvalues    []vm.UpvalueDesc
	uvNames     []string
	uvMap       map[uvKey]int
	captured    map[int]bool
	loops       []*loopCtx
	regionDepth int
	stack       int
	maxStack    int
}

type compiler struct {
	mod     *vm.Module
	fn      *fnState
	curSpan uint16
	consts  map[constKey]int
	errs    []error
}

type constKey struct {
	tag byte
	i   int64
	s   string
}

func Compile(stmts []ast.Stmt) (*vm.Module, error) {
	c := &compiler{
		mod: &vm.Module{
			Constants: []types.Value{},
			Spans:     []vm.Span{},
			Hotspots:  []vm.Hotspot{},
			Functions: []*vm.Function{},
		},
		curSpan: vm.None,
		consts:  make(map[constKey]int),
	}
	c.beginFunction(nil, false, nil)
	c.compileTop(stmts)
	c.finishFunction()
	c.mod.Entry = 0
	if len(c.errs) > 0 {
		return c.mod, c.errs[0]
	}
	if len(c.mod.Functions[0].Bytecode) == 0 {
		c.fn = &fnState{index: 0}
		// unreachable
	}
	return c.mod, nil
}

func (c *compiler) errorf(node ast.Stmt, msg string, args ...any) {
	start, end := ast.Pos{}, ast.Pos{}
	if node != nil {
		start, end = node.StartPos(), node.EndPos()
	}
	c.errs = append(c.errs, Error{Msg: fmt.Sprintf(msg, args...), Start: start, End: end})
}

func (c *compiler) beginFunction(params []*ast.IdentifierExpr, variadic bool, parent *fnState) {
	idx := len(c.mod.Functions)
	c.mod.Functions = append(c.mod.Functions, &vm.Function{})
	c.fn = &fnState{
		parent:   parent,
		index:    idx,
		arity:    len(params),
		variadic: variadic,
		uvMap:    make(map[uvKey]int),
		captured: make(map[int]bool),
	}
	c.beginScope()
	for _, p := range params {
		c.declare(p.Value)
	}
}

func (c *compiler) finishFunction() {
	fn := c.fn
	names := make([]uint16, len(fn.slotNames))
	for i, n := range fn.slotNames {
		if n == "" {
			names[i] = vm.None
		} else {
			names[i] = uint16(c.addString(n))
		}
	}
	f := c.mod.Functions[fn.index]
	f.Variadic = fn.variadic
	f.Arity = uint16(fn.arity)
	max := uint16(fn.maxStack)
	if max == 0 {
		max = 1
	}
	f.MaxStack = max
	f.SlotNames = names
	f.Upvalues = fn.upvalues
	f.Bytecode = fn.code
	f.Arguments = fn.args
	f.Spans = fn.spans
}

func (c *compiler) beginScope() {
	c.fn.scopes = append(c.fn.scopes, blockScope{names: make(map[string]int), bound: make(map[string]bool)})
}

func (c *compiler) endScope() {
	c.fn.scopes = c.fn.scopes[:len(c.fn.scopes)-1]
}

func (c *compiler) declare(name string) int {
	sc := &c.fn.scopes[len(c.fn.scopes)-1]
	if sl, ok := sc.names[name]; ok {
		return sl
	}
	sl := len(c.fn.slotNames)
	c.fn.slotNames = append(c.fn.slotNames, name)
	sc.names[name] = sl
	return sl
}

func (c *compiler) allocTemp() int {
	sl := len(c.fn.slotNames)
	c.fn.slotNames = append(c.fn.slotNames, "")
	return sl
}

func (c *compiler) hoist(stmts []ast.Stmt) {
	for _, s := range stmts {
		if d, ok := s.(*ast.DeclareStmt); ok {
			c.declare(d.Target.Value)
		}
	}
}

func (c *compiler) resolveFn(fn *fnState, name string) (kind, idx int) {
	for i := len(fn.scopes) - 1; i >= 0; i-- {
		if sl, ok := fn.scopes[i].names[name]; ok {
			return resLocal, sl
		}
	}
	for i, n := range fn.uvNames {
		if n == name {
			return resUpvalue, i
		}
	}
	if fn.parent != nil {
		k, i := c.resolveFn(fn.parent, name)
		switch k {
		case resLocal:
			fn.parent.captured[i] = true
			return resUpvalue, c.addUV(fn, true, uint16(i), name)
		case resUpvalue:
			return resUpvalue, c.addUV(fn, false, uint16(i), name)
		}
	}
	return resGlobal, c.addString(name)
}

func (c *compiler) addUV(fn *fnState, isLocal bool, index uint16, name string) int {
	key := uvKey{isLocal, index}
	if i, ok := fn.uvMap[key]; ok {
		return i
	}
	i := len(fn.upvalues)
	if i >= 255 {
		c.errorf(nil, "too many upvalues")
	}
	fn.upvalues = append(fn.upvalues, vm.UpvalueDesc{IsLocal: isLocal, Index: index})
	fn.uvNames = append(fn.uvNames, name)
	fn.uvMap[key] = i
	return i
}

func (c *compiler) emit(op vm.Opcode, arg int32) int {
	c.fn.code = append(c.fn.code, op)
	c.fn.args = append(c.fn.args, arg)
	c.fn.spans = append(c.fn.spans, c.curSpan)
	c.fn.stack += stackEffect(op, arg)
	if c.fn.stack < 0 {
		c.fn.stack = 0
	}
	if c.fn.stack > c.fn.maxStack {
		c.fn.maxStack = c.fn.stack
	}
	return len(c.fn.code)
}

func (c *compiler) patchJump(placeholder int) {
	offset := len(c.fn.code) - placeholder
	c.fn.args[placeholder-1] = int32(offset)
}

func (c *compiler) patchJumpTo(placeholder, target int) {
	offset := target - placeholder
	c.fn.args[placeholder-1] = int32(offset)
}

func (c *compiler) calcBackward(to int) int {
	return len(c.fn.code) + 1 - to
}

func (c *compiler) patchList(jumps []int, target int) {
	for _, p := range jumps {
		c.patchJumpTo(p, target)
	}
}

func (c *compiler) pushSpan(node ast.Stmt, label string) uint16 {
	prev := c.curSpan
	lab := vm.None
	if label != "" {
		lab = uint16(c.addString(label))
	}
	parent := prev
	idx := uint16(len(c.mod.Spans))
	c.mod.Spans = append(c.mod.Spans, vm.Span{
		Line:     uint32(node.StartPos().Line),
		Index:    uint32(node.StartPos().Index),
		EndLine:  uint32(node.EndPos().Line),
		EndIndex: uint32(node.EndPos().Index),
		Parent:   parent,
		Label:    lab,
	})
	c.curSpan = idx
	return prev
}

func (c *compiler) withSpan(node ast.Stmt, label string, fn func()) {
	if node == nil {
		fn()
		return
	}
	prev := c.pushSpan(node, label)
	fn()
	c.curSpan = prev
}

func (c *compiler) addHotspot(label string, node ast.Stmt, flags byte) int32 {
	sp := c.curSpan
	if node != nil && (sp == vm.None || len(c.mod.Spans) == 0) {
		c.pushSpan(node, "")
		sp = c.curSpan
	}
	idx := int32(len(c.mod.Hotspots))
	c.mod.Hotspots = append(c.mod.Hotspots, vm.Hotspot{
		Label: uint16(c.addString(label)),
		Span:  sp,
		Flags: flags,
	})
	return idx
}

func (c *compiler) emitEnter(h int32) {
	c.emit(vm.OpEnter, h)
	c.fn.regionDepth++
}

func (c *compiler) emitLeave() {
	c.emit(vm.OpLeave, 0)
	c.fn.regionDepth--
}

func (c *compiler) emitJumpLeaves(n int) {
	for i := 0; i < n; i++ {
		c.emit(vm.OpLeave, 0)
	}
}

func (c *compiler) addNull() int {
	return c.addConst(constKey{tag: vm.ConstNull}, nil)
}

func (c *compiler) addBool(b bool) int {
	v := int64(0)
	if b {
		v = 1
	}
	return c.addConst(constKey{tag: vm.ConstBool, i: v}, values.Bool(b))
}

func (c *compiler) addInt(i int64) int {
	return c.addConst(constKey{tag: vm.ConstInt, i: i}, values.Int(i))
}

func (c *compiler) addFloat(f float64) int {
	return c.addConst(constKey{tag: vm.ConstFloat, s: fmt.Sprintf("%b", f)}, values.Float(f))
}

func (c *compiler) addString(s string) int {
	return c.addConst(constKey{tag: vm.ConstString, s: s}, values.String(s))
}

func (c *compiler) addConst(k constKey, v types.Value) int {
	if i, ok := c.consts[k]; ok {
		return i
	}
	i := len(c.mod.Constants)
	if i > 65535 {
		c.errorf(nil, "too many constants")
	}
	c.mod.Constants = append(c.mod.Constants, v)
	c.consts[k] = i
	return i
}

func (c *compiler) emitPushNull() {
	c.emit(vm.OpPush, int32(c.addNull()))
}

func (c *compiler) emitPushBool(b bool) {
	c.emit(vm.OpPush, int32(c.addBool(b)))
}

func stackEffect(op vm.Opcode, arg int32) int {
	switch op {
	case vm.OpPush, vm.OpDup, vm.OpLoadLocal, vm.OpLoadUpvalue, vm.OpLoadGlobal, vm.OpPointer, vm.OpClosure:
		return 1
	case vm.OpDup2:
		return 2
	case vm.OpPop, vm.OpStoreLocal, vm.OpDefineLocal, vm.OpDefineConstLocal,
		vm.OpStoreUpvalue, vm.OpDefineUpvalue, vm.OpDefineConstUpvalue,
		vm.OpStoreGlobal, vm.OpDefineGlobal, vm.OpDefineConstGlobal,
		vm.OpArg, vm.OpSpread, vm.OpReturn, vm.OpBegin:
		return -1
	case vm.OpAdd, vm.OpSubtract, vm.OpMultiply, vm.OpDivide, vm.OpModulo,
		vm.OpEqual, vm.OpLess, vm.OpMore, vm.OpLessOrEqual, vm.OpMoreOrEqual,
		vm.OpAnd, vm.OpOr, vm.OpGetIndex:
		return -1
	case vm.OpSetIndex:
		return -2
	case vm.OpStoreField:
		return -1
	case vm.OpArray:
		return 1 - int(arg)
	case vm.OpMap:
		return 1 - int(arg)*2
	case vm.OpCall:
		return -int(arg)
	case vm.OpCallArgs:
		return 0
	default:
		return 0
	}
}
