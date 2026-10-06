package vm

import (
	"fmt"

	"github.com/anhcraft/rice/exec/types"
	"github.com/anhcraft/rice/exec/types/values"
)

// Module is a decoded Rice bytecode module.
type Module struct {
	Version   Version // declared version; Encode stamps max(Version, RequiredVersion())
	Constants []types.Value
	Spans     []Span
	Hotspots  []Hotspot
	Functions []*Function
	Entry     uint16
}

type Span struct {
	Line     uint32
	Index    uint32
	EndLine  uint32
	EndIndex uint32
	Parent   uint16 // None if no parent
	Label    uint16 // string const index, or None
}

type Hotspot struct {
	Label uint16 // string const index
	Span  uint16
	Flags byte
}

func (h Hotspot) OpensScope() bool {
	return h.Flags&HotspotOpensScope != 0
}

type UpvalueDesc struct {
	IsLocal bool
	Index   uint16
}

type Function struct {
	Variadic  bool
	Arity     uint16
	MaxStack  uint16
	SlotNames []uint16 // string const or None for temps
	Upvalues  []UpvalueDesc
	Bytecode  []Opcode
	Arguments []int32
	Spans     []uint16
}

func (m *Module) ConstString(idx uint16) string {
	if idx == None || int(idx) >= len(m.Constants) {
		return ""
	}
	if s, ok := m.Constants[idx].(values.String); ok {
		return string(s)
	}
	return fmt.Sprint(m.Constants[idx])
}

func (m *Module) SpanLabel(sp Span) string {
	return m.ConstString(sp.Label)
}
