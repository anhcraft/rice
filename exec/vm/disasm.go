package vm

import (
	"bytes"
	"fmt"
	"io"
	"text/tabwriter"
)

func (m *Module) Disassemble() string {
	var buf bytes.Buffer
	w := tabwriter.NewWriter(&buf, 0, 0, 2, ' ', 0)
	m.DisassembleWriter(w)
	_ = w.Flush()
	return buf.String()
}

func (m *Module) DisassembleWriter(w io.Writer) {
	fmt.Fprintf(w, "; version %s  entry %d  consts %d  spans %d  hotspots %d  funcs %d\n",
		m.Version.Max(m.RequiredVersion()), m.Entry, len(m.Constants), len(m.Spans), len(m.Hotspots), len(m.Functions))
	for i, c := range m.Constants {
		fmt.Fprintf(w, "; const %d\t%T\t%v\n", i, c, c)
	}
	for fi, fn := range m.Functions {
		fmt.Fprintf(w, "\n; function %d arity=%d variadic=%t maxstack=%d slots=%d upvalues=%d\n",
			fi, fn.Arity, fn.Variadic, fn.MaxStack, len(fn.SlotNames), len(fn.Upvalues))
		ip := 0
		for ip < len(fn.Bytecode) {
			op := fn.Bytecode[ip]
			arg := fn.Arguments[ip]
			spanIdx := uint16(None)
			if ip < len(fn.Spans) {
				spanIdx = fn.Spans[ip]
			}
			fmt.Fprintf(w, "%d\t%s\t<%d>\t%s\n", ip, op, arg, m.argHint(fn, op, arg, spanIdx))
			ip++
		}
	}
}

func (m *Module) argHint(fn *Function, op Opcode, arg int32, spanIdx uint16) string {
	hint := ""
	switch op {
	case OpPush:
		if arg >= 0 && int(arg) < len(m.Constants) {
			hint = fmt.Sprintf("%v", m.Constants[arg])
		}
	case OpLoadLocal, OpStoreLocal, OpDefineLocal, OpDefineConstLocal, OpCloseUpvalue:
		if arg >= 0 && int(arg) < len(fn.SlotNames) {
			hint = m.ConstString(fn.SlotNames[arg])
		}
	case OpLoadGlobal, OpStoreGlobal, OpDefineGlobal, OpDefineConstGlobal, OpLoadField, OpStoreField:
		if arg >= 0 && int(arg) < len(m.Constants) {
			hint = m.ConstString(uint16(arg))
		}
	case OpJump, OpJumpIfFalse, OpJumpIfFalsey, OpJumpIfBoolFalse, OpJumpIfBoolTrue, OpJumpIfEnd:
		hint = fmt.Sprintf("→ %d", int(spanIdx) /* placeholder */)
		_ = spanIdx
		hint = fmt.Sprintf("→ %+d", arg)
	case OpJumpBackward:
		hint = fmt.Sprintf("← %d", arg)
	case OpEnter:
		if arg >= 0 && int(arg) < len(m.Hotspots) {
			hint = m.ConstString(m.Hotspots[arg].Label)
		}
	case OpClosure:
		hint = fmt.Sprintf("fn %d", arg)
	}
	if spanIdx != None && int(spanIdx) < len(m.Spans) {
		sp := m.Spans[spanIdx]
		lab := m.SpanLabel(sp)
		if lab != "" {
			if hint != "" {
				hint += " "
			}
			hint += lab
		}
	}
	return hint
}
