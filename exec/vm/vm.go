package vm

import (
	"context"
	"fmt"

	"github.com/anhcraft/rice/exec/ast"
	"github.com/anhcraft/rice/exec/types"
	"github.com/anhcraft/rice/exec/types/values"
)

type slot struct {
	val      types.Value
	defined  bool
	constant bool
}

type upvalue struct {
	sl       *slot
	closed   types.Value
	isClosed bool
	constant bool
}

func (u *upvalue) get() (types.Value, bool, bool) {
	if u.isClosed {
		return u.closed, true, u.constant
	}
	return u.sl.val, u.sl.defined, u.sl.constant
}

func (u *upvalue) set(v types.Value) error {
	if u.isClosed {
		if u.constant {
			return fmt.Errorf("cannot assign to constant")
		}
		u.closed = v
		return nil
	}
	if u.sl.constant {
		return fmt.Errorf("cannot assign to constant")
	}
	if !u.sl.defined {
		return fmt.Errorf("unknown variable")
	}
	u.sl.val = v
	return nil
}

func (u *upvalue) define(v types.Value, constant bool) error {
	if u.isClosed {
		if u.definedClosed() {
			return fmt.Errorf("cannot redeclare")
		}
		u.closed = v
		u.constant = constant
		return nil
	}
	if u.sl.defined {
		return fmt.Errorf("cannot redeclare")
	}
	u.sl.val = v
	u.sl.defined = true
	u.sl.constant = constant
	return nil
}

func (u *upvalue) definedClosed() bool { return true }

func (u *upvalue) close() {
	if u.isClosed {
		return
	}
	u.closed = u.sl.val
	u.constant = u.sl.constant
	u.isClosed = true
	u.sl = nil
}

type iterScope struct {
	items []types.Value
	index int
}

type region struct {
	opensScope bool
}

type frame struct {
	module     *Module
	fn         *Function
	fnIndex    int
	slots      []slot
	upvalues   []*upvalue
	openUV     map[int]*upvalue
	ip         int
	stackBase  int
	regionBase int
	iterBase   int
	scopeBase  uint32
	profDepth  int
	site       values.CallSite
}

type closureEnv struct {
	module   *Module
	fnIndex  int
	upvalues []*upvalue
	vm       *VM
}

type VM struct {
	Stack []types.Value
	host  Host

	frames     []*frame
	regions    []region
	iters      []iterScope
	scopeDepth uint32
	argBuf     []types.Value

	debug   bool
	step    chan struct{}
	curr    chan DebugPos
	dbgDone chan struct{}
}

func (vm *VM) Run(mod *Module, host Host) (types.Value, error) {
	vm.host = host
	if vm.Stack == nil {
		vm.Stack = make([]types.Value, 0, 16)
	} else {
		vm.Stack = vm.Stack[:0]
	}
	vm.frames = vm.frames[:0]
	vm.regions = vm.regions[:0]
	vm.iters = vm.iters[:0]
	vm.scopeDepth = 0
	vm.argBuf = vm.argBuf[:0]
	val, err := vm.execFn(mod, int(mod.Entry), nil, nil, values.RootCallSite)
	vm.debugFinish()
	return val, err
}

func (vm *VM) invokeClosure(ctx context.Context, self *values.Func, site values.CallSite, args []types.Value) (types.Value, error) {
	env := self.Closure().(*closureEnv)
	prev := vm.host.Context()
	callCtx, cancel := context.WithTimeout(ctx, vm.host.UserFuncTimeout())
	vm.host.SetContext(callCtx)
	defer func() {
		cancel()
		vm.host.SetContext(prev)
	}()
	return vm.execFn(env.module, env.fnIndex, args, env.upvalues, site)
}

func (vm *VM) execFn(mod *Module, fnIndex int, args []types.Value, uvs []*upvalue, site values.CallSite) (val types.Value, err error) {
	if fnIndex < 0 || fnIndex >= len(mod.Functions) {
		return nil, fmt.Errorf("function index %d out of range", fnIndex)
	}
	fn := mod.Functions[fnIndex]

	n := len(args)
	arity := int(fn.Arity)
	if (fn.Variadic && n < arity-1) || (!fn.Variadic && n < arity) {
		return nil, vm.callFault(site, "too few arguments supplied to a call of %s", formatFunc(fn, mod))
	}
	if !fn.Variadic && n > arity {
		return nil, vm.callFault(site, "too many arguments supplied to a call of %s", formatFunc(fn, mod))
	}

	slots := make([]slot, len(fn.SlotNames))
	for i, name := range fn.SlotNames {
		if name == None {
			slots[i].defined = true
		}
	}

	fr := &frame{
		module:     mod,
		fn:         fn,
		fnIndex:    fnIndex,
		slots:      slots,
		upvalues:   uvs,
		openUV:     make(map[int]*upvalue),
		stackBase:  len(vm.Stack),
		regionBase: len(vm.regions),
		iterBase:   len(vm.iters),
		scopeBase:  vm.scopeDepth,
		profDepth:  vm.host.ProfilerDepth(),
		site:       site,
	}

	vm.scopeDepth++
	if err := vm.checkScope(); err != nil {
		vm.scopeDepth--
		return nil, vm.callFault(site, "function-literal evaluation gets interrupted").causedBy(err)
	}

	if fn.Variadic {
		last := arity - 1
		for k := 0; k < last; k++ {
			fr.slots[k] = slot{val: args[k], defined: true}
		}
		hole := make([]types.Value, n-last)
		copy(hole, args[last:])
		if last >= 0 && last < len(fr.slots) {
			fr.slots[last] = slot{val: values.ListOf(hole), defined: true}
		}
	} else {
		for k := 0; k < arity; k++ {
			fr.slots[k] = slot{val: args[k], defined: true}
		}
	}

	vm.frames = append(vm.frames, fr)

	defer func() {
		if r := recover(); r != nil {
			switch e := r.(type) {
			case Fault:
				err = e
			case error:
				err = vm.wrapErr(e)
			default:
				err = vm.wrapErr(fmt.Errorf("%v", r))
			}
		}
		vm.leaveFrame(fr)
	}()

	return vm.interpret(fr), nil
}

func (vm *VM) leaveFrame(fr *frame) {
	for _, uv := range fr.openUV {
		uv.close()
	}
	for len(vm.regions) > fr.regionBase {
		vm.leaveRegion()
	}
	vm.host.ProfilerUnwind(fr.profDepth)
	vm.scopeDepth = fr.scopeBase
	vm.iters = vm.iters[:fr.iterBase]
	vm.Stack = vm.Stack[:fr.stackBase]
	if n := len(vm.frames); n > 0 && vm.frames[n-1] == fr {
		vm.frames = vm.frames[:n-1]
	}
}

func (vm *VM) frame() *frame {
	return vm.frames[len(vm.frames)-1]
}

func (vm *VM) interpret(fr *frame) types.Value {
	code := fr.fn.Bytecode
	args := fr.fn.Arguments
	for fr.ip < len(code) {
		if debug && vm.debug {
			<-vm.step
		}

		op := code[fr.ip]
		arg := args[fr.ip]
		fr.ip++

		switch op {
		case OpPush:
			vm.push(fr.module.Constants[arg])
		case OpPop:
			vm.pop()
		case OpDup:
			vm.push(vm.current())
		case OpDup2:
			a := vm.Stack[len(vm.Stack)-2]
			b := vm.Stack[len(vm.Stack)-1]
			vm.push(a)
			vm.push(b)
		case OpRot3:
			n := len(vm.Stack)
			vm.Stack[n-3], vm.Stack[n-2], vm.Stack[n-1] = vm.Stack[n-2], vm.Stack[n-1], vm.Stack[n-3]

		case OpLoadLocal:
			sl := &fr.slots[arg]
			if !sl.defined {
				panic(vm.fault("unresolved reference %q", vm.slotName(fr, int(arg))))
			}
			vm.push(sl.val)
		case OpStoreLocal:
			v := vm.pop()
			sl := &fr.slots[arg]
			if sl.constant {
				panic(vm.fault("cannot assign to constant %q", vm.slotName(fr, int(arg))))
			}
			if !sl.defined {
				panic(vm.fault("unknown variable %q", vm.slotName(fr, int(arg))))
			}
			sl.val = v
		case OpDefineLocal, OpDefineConstLocal:
			v := vm.pop()
			name := vm.slotName(fr, int(arg))
			if name != "" && vm.host.IsNamespaceEntry(name) {
				panic(vm.fault("cannot declare %q because it conflicts with a built-in function or namespace", name))
			}
			sl := &fr.slots[arg]
			sl.val = v
			sl.defined = true
			sl.constant = op == OpDefineConstLocal

		case OpLoadUpvalue:
			v, defined, _ := fr.upvalues[arg].get()
			if !defined {
				panic(vm.fault("unresolved reference %q", vm.upvalueName(fr, int(arg))))
			}
			vm.push(v)
		case OpStoreUpvalue:
			v := vm.pop()
			if err := fr.upvalues[arg].set(v); err != nil {
				panic(vm.fault("%s", err.Error()))
			}
		case OpDefineUpvalue, OpDefineConstUpvalue:
			v := vm.pop()
			name := vm.upvalueName(fr, int(arg))
			if name != "" && vm.host.IsNamespaceEntry(name) {
				panic(vm.fault("cannot declare %q because it conflicts with a built-in function or namespace", name))
			}
			if err := fr.upvalues[arg].define(v, op == OpDefineConstUpvalue); err != nil {
				panic(vm.fault("cannot redeclare %q", name))
			}
		case OpCloseUpvalue:
			if uv, ok := fr.openUV[int(arg)]; ok {
				uv.close()
				delete(fr.openUV, int(arg))
			}

		case OpLoadGlobal:
			name := vm.constString(fr, arg)
			v, err := vm.host.LoadGlobal(name)
			if err != nil {
				panic(vm.wrapErr(err))
			}
			vm.push(v)
		case OpStoreGlobal:
			v := vm.pop()
			name := vm.constString(fr, arg)
			if err := vm.host.StoreGlobal(name, v); err != nil {
				panic(vm.wrapErr(err))
			}
		case OpDefineGlobal, OpDefineConstGlobal:
			v := vm.pop()
			name := vm.constString(fr, arg)
			if err := vm.host.DefineGlobal(name, v, op == OpDefineConstGlobal); err != nil {
				panic(vm.wrapErr(err))
			}

		case OpJump:
			fr.ip += int(arg)
		case OpJumpBackward:
			if err := vm.host.CheckContext(); err != nil {
				panic(vm.wrapErr(fmt.Errorf("iteration gets interrupted: %w", err)))
			}
			fr.ip -= int(arg)
		case OpJumpIfFalse:
			v := vm.current()
			b, ok := v.(values.Bool)
			if !ok {
				panic(vm.fault("expect if-condition to be Bool but got %T", v))
			}
			if !b {
				fr.ip += int(arg)
			}
		case OpJumpIfFalsey:
			v := vm.current()
			b, err := values.AsBool(v)
			if err != nil {
				panic(vm.wrapErr(fmt.Errorf("cannot implicitly convert condition of type %T to Bool: %w", v, err)))
			}
			if !b {
				fr.ip += int(arg)
			}
		case OpJumpIfBoolFalse:
			if b, ok := vm.current().(values.Bool); ok && !bool(b) {
				fr.ip += int(arg)
			}
		case OpJumpIfBoolTrue:
			if b, ok := vm.current().(values.Bool); ok && bool(b) {
				fr.ip += int(arg)
			}

		case OpNegate:
			v, err := unaryNegate(vm.pop())
			if err != nil {
				panic(vm.wrapErr(err))
			}
			vm.push(v)
		case OpNot:
			v, err := unaryNot(vm.pop())
			if err != nil {
				panic(vm.wrapErr(err))
			}
			vm.push(v)
		case OpAdd, OpSubtract, OpMultiply, OpDivide, OpModulo, OpEqual, OpLess, OpMore, OpLessOrEqual, OpMoreOrEqual, OpAnd, OpOr:
			right := vm.pop()
			left := vm.pop()
			v, err := binaryOp(op, left, right)
			if err != nil {
				panic(vm.wrapErr(err))
			}
			vm.push(v)
		case OpIncrement:
			v, err := increment(vm.pop(), values.Int(arg))
			if err != nil {
				panic(vm.wrapErr(err))
			}
			vm.push(v)

		case OpArray:
			n := int(arg)
			elems := make([]types.Value, n)
			for i := n - 1; i >= 0; i-- {
				elems[i] = vm.pop()
			}
			vm.push(values.ListOf(elems))
		case OpMap:
			n := int(arg)
			pairs := make([]types.Value, n*2)
			for i := n*2 - 1; i >= 0; i-- {
				pairs[i] = vm.pop()
			}
			m := values.NewMap()
			for i := 0; i < n; i++ {
				k := pairs[i*2]
				v := pairs[i*2+1]
				if id, ok := k.(values.Identifier); ok {
					k = values.String(id)
				}
				m.Put(k, v)
			}
			vm.push(m)
		case OpGetIndex:
			idx := vm.pop()
			obj := vm.pop()
			vm.push(vm.getIndex(obj, idx))
		case OpSetIndex:
			val := vm.pop()
			idx := vm.pop()
			obj := vm.pop()
			vm.setIndex(obj, idx, val)
			vm.push(val)
		case OpLoadField:
			obj := vm.pop()
			name := values.Identifier(vm.constString(fr, arg))
			vm.push(vm.loadField(obj, name))
		case OpStoreField:
			val := vm.pop()
			obj := vm.pop()
			name := values.Identifier(vm.constString(fr, arg))
			vm.storeField(obj, name, val)
			vm.push(val)

		case OpClosure:
			vm.push(vm.makeClosure(fr, int(arg)))
		case OpCall:
			argc := int(arg)
			cargs := make([]types.Value, argc)
			for i := argc - 1; i >= 0; i-- {
				cargs[i] = vm.pop()
			}
			callee := vm.pop()
			vm.push(vm.call(fr, callee, cargs))
		case OpArgs:
			vm.argBuf = vm.argBuf[:0]
		case OpArg:
			vm.argBuf = append(vm.argBuf, vm.pop())
		case OpSpread:
			v := vm.pop()
			coll, ok := v.(values.Collection)
			if !ok {
				panic(vm.fault("cannot spread because type %T is not collection", v))
			}
			for item := range coll.Iterate() {
				vm.argBuf = append(vm.argBuf, item)
			}
		case OpCallArgs:
			callee := vm.pop()
			cargs := append([]types.Value(nil), vm.argBuf...)
			vm.argBuf = vm.argBuf[:0]
			vm.push(vm.call(fr, callee, cargs))
		case OpReturn:
			v := vm.pop()
			return v

		case OpBegin:
			v := vm.pop()
			coll, ok := v.(values.Collection)
			if !ok {
				panic(vm.fault("value of type %T is not collection", v))
			}
			var items []types.Value
			for item := range coll.Iterate() {
				items = append(items, item)
			}
			vm.iters = append(vm.iters, iterScope{items: items, index: 0})
		case OpJumpIfEnd:
			it := &vm.iters[len(vm.iters)-1]
			if it.index >= len(it.items) {
				fr.ip += int(arg)
			}
		case OpPointer:
			it := &vm.iters[len(vm.iters)-1]
			vm.push(it.items[it.index])
		case OpIncrementIndex:
			vm.iters[len(vm.iters)-1].index++
		case OpEnd:
			vm.iters = vm.iters[:len(vm.iters)-1]

		case OpEnter:
			vm.enter(fr, int(arg))
		case OpLeave:
			vm.leaveRegion()

		default:
			panic(vm.fault("unknown opcode %#x", op))
		}

		if debug && vm.debug {
			vm.curr <- DebugPos{Func: fr.fnIndex, IP: fr.ip}
		}
	}

	if len(vm.Stack) > fr.stackBase {
		return vm.pop()
	}
	return nil
}

func (vm *VM) enter(fr *frame, hi int) {
	if err := vm.host.CheckContext(); err != nil {
		panic(vm.wrapErr(fmt.Errorf("evaluation gets interrupted: %w", err)))
	}
	h := fr.module.Hotspots[hi]
	if h.OpensScope() {
		vm.scopeDepth++
		if err := vm.checkScope(); err != nil {
			panic(vm.wrapErr(err))
		}
	}
	lab := fr.module.ConstString(h.Label)
	pos := ast.Pos{}
	if int(h.Span) < len(fr.module.Spans) {
		sp := fr.module.Spans[h.Span]
		pos = ast.Pos{Index: int(sp.Index), Line: int(sp.Line)}
	}
	vm.host.ProfilerStart(lab, pos)
	vm.regions = append(vm.regions, region{opensScope: h.OpensScope()})
}

func (vm *VM) leaveRegion() {
	if len(vm.regions) == 0 {
		return
	}
	r := vm.regions[len(vm.regions)-1]
	vm.regions = vm.regions[:len(vm.regions)-1]
	vm.host.ProfilerEnd()
	if r.opensScope && vm.scopeDepth > 0 {
		vm.scopeDepth--
	}
}

func (vm *VM) checkScope() error {
	if vm.scopeDepth > vm.host.ScopeLimit() {
		return fmt.Errorf("reached lexical scope limit of %d", vm.host.ScopeLimit())
	}
	return nil
}

func (vm *VM) call(fr *frame, callee types.Value, args []types.Value) types.Value {
	if err := vm.host.CheckContext(); err != nil {
		panic(vm.wrapErr(fmt.Errorf("call evaluation gets interrupted: %w", err)))
	}
	callable, ok := callee.(values.Callable)
	if !ok {
		panic(vm.fault("callee of type %T is not callable", callee))
	}
	sp := vm.span(fr, fr.ip-1)
	site := values.CallSite{Caller: calleeName(callee), StartPos: sp.start(), EndPos: sp.end()}

	vm.scopeDepth++
	if err := vm.checkScope(); err != nil {
		vm.scopeDepth--
		panic(vm.wrapErr(err))
	}
	defer func() { vm.scopeDepth-- }()

	res, err := callable.Call(vm.host.Context(), site, args)
	if err != nil {
		panic(vm.noteCall(err, site))
	}
	return res
}

func (vm *VM) makeClosure(fr *frame, fnIndex int) *values.Func {
	fn := fr.module.Functions[fnIndex]
	uvs := make([]*upvalue, len(fn.Upvalues))
	for i, d := range fn.Upvalues {
		if d.IsLocal {
			uvs[i] = fr.capture(int(d.Index))
		} else {
			uvs[i] = fr.upvalues[d.Index]
		}
	}
	params := make([]values.Identifier, fn.Arity)
	for i := 0; i < int(fn.Arity); i++ {
		if i < len(fn.SlotNames) {
			params[i] = values.Identifier(fr.module.ConstString(fn.SlotNames[i]))
		}
	}
	env := &closureEnv{module: fr.module, fnIndex: fnIndex, upvalues: uvs, vm: vm}
	return values.NewFunc(params, fn.Variadic, env, func(ctx context.Context, self *values.Func, site values.CallSite, args []types.Value) (types.Value, error) {
		return env.vm.invokeClosure(ctx, self, site, args)
	})
}

func (fr *frame) capture(slotIdx int) *upvalue {
	if uv, ok := fr.openUV[slotIdx]; ok {
		return uv
	}
	uv := &upvalue{sl: &fr.slots[slotIdx]}
	fr.openUV[slotIdx] = uv
	return uv
}

func (vm *VM) getIndex(obj, idx types.Value) types.Value {
	if obj == nil {
		panic(vm.fault("cannot access element on nil"))
	}
	coll, ok := obj.(values.IndexedCollection)
	if !ok {
		panic(vm.fault("object of type %T is not indexed collection", obj))
	}
	v, err := coll.Element(idx)
	if err != nil {
		panic(vm.wrapErr(err))
	}
	return v
}

func (vm *VM) setIndex(obj, idx, val types.Value) {
	if obj == nil {
		panic(vm.fault("cannot access element on nil"))
	}
	coll, ok := obj.(values.IndexedCollection)
	if !ok {
		panic(vm.fault("object of type %T is not indexed collection", obj))
	}
	if err := coll.PutElement(idx, val); err != nil {
		panic(vm.wrapErr(err))
	}
}

func (vm *VM) loadField(obj types.Value, name values.Identifier) types.Value {
	if obj == nil {
		panic(vm.fault("cannot select on nil"))
	}
	if v, ok := vm.host.TypeBound(obj, name); ok {
		return v
	}
	coll, ok := obj.(values.IndexedCollection)
	if !ok {
		panic(vm.fault("object of type %T is not indexed collection", obj))
	}
	v, err := coll.Element(name)
	if err != nil {
		panic(vm.wrapErr(err))
	}
	return v
}

func (vm *VM) storeField(obj types.Value, name values.Identifier, val types.Value) {
	if obj == nil {
		panic(vm.fault("cannot select on nil"))
	}
	coll, ok := obj.(values.IndexedCollection)
	if !ok {
		panic(vm.fault("object of type %T is not indexed collection", obj))
	}
	if err := coll.PutElement(name, val); err != nil {
		panic(vm.wrapErr(err))
	}
}

func (vm *VM) push(v types.Value) { vm.Stack = append(vm.Stack, v) }

func (vm *VM) pop() types.Value {
	n := len(vm.Stack)
	if n == 0 {
		op := Opcode(0)
		ip := -1
		if len(vm.frames) > 0 {
			fr := vm.frame()
			ip = fr.ip - 1
			if ip >= 0 && ip < len(fr.fn.Bytecode) {
				op = fr.fn.Bytecode[ip]
			}
		}
		panic(fmt.Sprintf("stack underflow at ip %d (%s)", ip, op))
	}
	v := vm.Stack[n-1]
	vm.Stack = vm.Stack[:n-1]
	return v
}

func (vm *VM) current() types.Value {
	n := len(vm.Stack)
	if n == 0 {
		panic("stack underflow")
	}
	return vm.Stack[n-1]
}

func (vm *VM) constString(fr *frame, arg int32) string {
	return fr.module.ConstString(uint16(arg))
}

func (vm *VM) slotName(fr *frame, i int) string {
	if i >= 0 && i < len(fr.fn.SlotNames) {
		return fr.module.ConstString(fr.fn.SlotNames[i])
	}
	return ""
}

func (vm *VM) upvalueName(fr *frame, i int) string {
	return fmt.Sprintf("upvalue[%d]", i)
}

func (vm *VM) span(fr *frame, ip int) Span {
	if ip >= 0 && ip < len(fr.fn.Spans) {
		si := fr.fn.Spans[ip]
		if si != None && int(si) < len(fr.module.Spans) {
			return fr.module.Spans[si]
		}
	}
	return Span{Parent: None, Label: None}
}

func (s Span) start() ast.Pos { return ast.Pos{Index: int(s.Index), Line: int(s.Line)} }
func (s Span) end() ast.Pos   { return ast.Pos{Index: int(s.EndIndex), Line: int(s.EndLine)} }

func (vm *VM) site() values.CallSite {
	if len(vm.frames) == 0 {
		return values.RootCallSite
	}
	return vm.frame().site
}

func calleeName(callee types.Value) string {
	switch c := callee.(type) {
	case *values.Func:
		return c.String()
	case values.NativeFunctionSet:
		if c.Name() != "" {
			return string(c.Name())
		}
		return "NativeFunctionSet"
	default:
		return "CallExpr"
	}
}

func (vm *VM) noteCall(err error, site values.CallSite) error {
	if err == nil {
		return nil
	}
	if f, ok := err.(Fault); ok {
		if !f.Source.Equal(site) {
			return Fault{
				Message: site.Caller,
				Source:  site,
				Start:   site.StartPos,
				End:     site.EndPos,
			}.causedBy(f)
		}
		return f
	}
	return vm.wrapErr(err)
}

func (vm *VM) fault(msg string, args ...any) Fault {
	return vm.wrapErr(fmt.Errorf(msg, args...)).(Fault)
}

func (vm *VM) wrapErr(err error) error {
	if err == nil {
		return nil
	}
	if f, ok := err.(Fault); ok {
		return f
	}
	fr := (*frame)(nil)
	if len(vm.frames) > 0 {
		fr = vm.frame()
	}
	ip := 0
	mod := (*Module)(nil)
	if fr != nil {
		ip = fr.ip - 1
		if ip < 0 {
			ip = 0
		}
		mod = fr.module
	}
	sp := Span{Parent: None, Label: None}
	if fr != nil {
		sp = vm.span(fr, ip)
	}
	f := Fault{
		Message: err.Error(),
		Source:  vm.site(),
		Start:   sp.start(),
		End:     sp.end(),
	}
	if mod != nil {
		for p := sp.Parent; p != None && int(p) < len(mod.Spans); {
			ps := mod.Spans[p]
			lab := mod.SpanLabel(ps)
			if lab != "" {
				f = Fault{Message: lab, Source: vm.site(), Start: ps.start(), End: ps.end()}.causedBy(f)
			}
			if p == ps.Parent {
				break
			}
			p = ps.Parent
		}
	}
	return f
}

func (vm *VM) callFault(site values.CallSite, msg string, args ...any) Fault {
	return Fault{
		Message: fmt.Sprintf(msg, args...),
		Source:  site,
		Start:   site.StartPos,
		End:     site.EndPos,
	}
}

func formatFunc(fn *Function, mod *Module) string {
	params := make([]values.Identifier, fn.Arity)
	for i := 0; i < int(fn.Arity); i++ {
		if i < len(fn.SlotNames) {
			params[i] = values.Identifier(mod.ConstString(fn.SlotNames[i]))
		}
	}
	return values.NewFunc(params, fn.Variadic, nil, dummyDelegate).String()
}

func dummyDelegate(context.Context, *values.Func, values.CallSite, []types.Value) (types.Value, error) {
	return nil, nil
}
