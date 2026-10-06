package vm

import (
	"encoding/binary"
	"fmt"
	"io"
	"math"

	"github.com/anhcraft/rice/exec/types"
	"github.com/anhcraft/rice/exec/types/values"
)

var be = binary.BigEndian

func (m *Module) Encode() ([]byte, error) {
	ver := m.Version.Max(m.RequiredVersion())
	var buf []byte
	buf = appendU32(buf, Magic)
	buf = appendU16(buf, ver.Major)
	buf = appendU16(buf, ver.Minor)

	if len(m.Constants) > 65535 || len(m.Spans) > 65535 || len(m.Hotspots) > 65535 || len(m.Functions) > 65535 {
		return nil, fmt.Errorf("module table exceeds u16 limit")
	}

	buf = appendU16(buf, uint16(len(m.Constants)))
	for _, c := range m.Constants {
		b, err := encodeConst(c)
		if err != nil {
			return nil, err
		}
		buf = append(buf, b...)
	}

	buf = appendU16(buf, uint16(len(m.Spans)))
	for _, s := range m.Spans {
		buf = appendU32(buf, s.Line)
		buf = appendU32(buf, s.Index)
		buf = appendU32(buf, s.EndLine)
		buf = appendU32(buf, s.EndIndex)
		buf = appendU16(buf, s.Parent)
		buf = appendU16(buf, s.Label)
	}

	buf = appendU16(buf, uint16(len(m.Hotspots)))
	for _, h := range m.Hotspots {
		buf = appendU16(buf, h.Label)
		buf = appendU16(buf, h.Span)
		buf = append(buf, h.Flags)
	}

	buf = appendU16(buf, uint16(len(m.Functions)))
	for _, fn := range m.Functions {
		b, err := encodeFunction(fn)
		if err != nil {
			return nil, err
		}
		buf = append(buf, b...)
	}

	buf = appendU16(buf, m.Entry)
	return buf, nil
}

func Decode(data []byte) (*Module, error) {
	r := &reader{b: data}
	magic, err := r.u32()
	if err != nil {
		return nil, err
	}
	if magic != Magic {
		return nil, fmt.Errorf("invalid magic: 0x%08x", magic)
	}
	major, err := r.u16()
	if err != nil {
		return nil, err
	}
	minor, err := r.u16()
	if err != nil {
		return nil, err
	}
	ver := Version{Major: major, Minor: minor}
	if err := versionLoadError(ver); err != nil {
		return nil, err
	}
	dec, ok := decoders[ver.Major]
	if !ok {
		return nil, fmt.Errorf("no decoder for RICE major %d", ver.Major)
	}

	m, err := dec(r, ver)
	if err != nil {
		return nil, err
	}
	m.Version = ver
	if err := verifyVersion(m); err != nil {
		return nil, err
	}
	if r.off != len(r.b) {
		return nil, fmt.Errorf("trailing bytes: %d leftover", len(r.b)-r.off)
	}
	if int(m.Entry) >= len(m.Functions) && len(m.Functions) > 0 {
		return nil, fmt.Errorf("entry %d out of range", m.Entry)
	}
	return m, nil
}

var decoders = map[uint16]func(*reader, Version) (*Module, error){
	1: decodeV1,
}

func decodeV1(r *reader, ver Version) (*Module, error) {
	m := &Module{Version: ver}
	nconst, err := r.u16()
	if err != nil {
		return nil, err
	}
	m.Constants = make([]types.Value, nconst)
	for i := 0; i < int(nconst); i++ {
		c, err := r.constant()
		if err != nil {
			return nil, fmt.Errorf("constant %d: %w", i, err)
		}
		m.Constants[i] = c
	}

	nspan, err := r.u16()
	if err != nil {
		return nil, err
	}
	m.Spans = make([]Span, nspan)
	for i := 0; i < int(nspan); i++ {
		sp := Span{}
		if sp.Line, err = r.u32(); err != nil {
			return nil, err
		}
		if sp.Index, err = r.u32(); err != nil {
			return nil, err
		}
		if sp.EndLine, err = r.u32(); err != nil {
			return nil, err
		}
		if sp.EndIndex, err = r.u32(); err != nil {
			return nil, err
		}
		if sp.Parent, err = r.u16(); err != nil {
			return nil, err
		}
		if sp.Label, err = r.u16(); err != nil {
			return nil, err
		}
		m.Spans[i] = sp
	}

	nhot, err := r.u16()
	if err != nil {
		return nil, err
	}
	m.Hotspots = make([]Hotspot, nhot)
	for i := 0; i < int(nhot); i++ {
		h := Hotspot{}
		if h.Label, err = r.u16(); err != nil {
			return nil, err
		}
		if h.Span, err = r.u16(); err != nil {
			return nil, err
		}
		if h.Flags, err = r.u8(); err != nil {
			return nil, err
		}
		m.Hotspots[i] = h
	}

	nfn, err := r.u16()
	if err != nil {
		return nil, err
	}
	m.Functions = make([]*Function, nfn)
	for i := 0; i < int(nfn); i++ {
		fn, err := r.function()
		if err != nil {
			return nil, fmt.Errorf("function %d: %w", i, err)
		}
		m.Functions[i] = fn
	}

	if m.Entry, err = r.u16(); err != nil {
		return nil, err
	}
	return m, nil
}

func encodeConst(c types.Value) ([]byte, error) {
	if c == nil {
		return []byte{ConstNull}, nil
	}
	switch v := c.(type) {
	case values.Bool:
		b := byte(0)
		if v {
			b = 1
		}
		return []byte{ConstBool, b}, nil
	case values.Int:
		buf := []byte{ConstInt}
		buf = appendI64(buf, int64(v))
		return buf, nil
	case values.Float:
		buf := []byte{ConstFloat}
		buf = appendU64(buf, math.Float64bits(float64(v)))
		return buf, nil
	case values.String:
		s := []byte(string(v))
		buf := []byte{ConstString}
		buf = appendU32(buf, uint32(len(s)))
		buf = append(buf, s...)
		return buf, nil
	default:
		return nil, fmt.Errorf("unencodable constant type %T", c)
	}
}

func encodeFunction(fn *Function) ([]byte, error) {
	if len(fn.Upvalues) > 255 {
		return nil, fmt.Errorf("too many upvalues")
	}
	if len(fn.Bytecode) != len(fn.Arguments) || len(fn.Bytecode) != len(fn.Spans) {
		return nil, fmt.Errorf("function array length mismatch")
	}
	if len(fn.SlotNames) > 65535 {
		return nil, fmt.Errorf("too many slots")
	}

	var buf []byte
	flags := byte(0)
	if fn.Variadic {
		flags |= FuncVariadic
	}
	buf = append(buf, flags)
	buf = appendU16(buf, fn.Arity)
	buf = appendU16(buf, fn.MaxStack)
	buf = appendU16(buf, uint16(len(fn.SlotNames)))
	for _, n := range fn.SlotNames {
		buf = appendU16(buf, n)
	}
	buf = appendU16(buf, uint16(len(fn.Upvalues)))
	for _, u := range fn.Upvalues {
		loc := byte(0)
		if u.IsLocal {
			loc = 1
		}
		buf = append(buf, loc)
		buf = appendU16(buf, u.Index)
	}
	buf = appendU32(buf, uint32(len(fn.Bytecode)))
	for i, op := range fn.Bytecode {
		if !ValidOpcode(op) {
			return nil, fmt.Errorf("unknown opcode %d", op)
		}
		buf = append(buf, byte(op))
		buf = appendI32(buf, fn.Arguments[i])
	}
	for _, s := range fn.Spans {
		buf = appendU16(buf, s)
	}
	return buf, nil
}

type reader struct {
	b   []byte
	off int
}

func (r *reader) remain() int { return len(r.b) - r.off }

func (r *reader) u8() (byte, error) {
	if r.remain() < 1 {
		return 0, io.ErrUnexpectedEOF
	}
	v := r.b[r.off]
	r.off++
	return v, nil
}

func (r *reader) u16() (uint16, error) {
	if r.remain() < 2 {
		return 0, io.ErrUnexpectedEOF
	}
	v := be.Uint16(r.b[r.off:])
	r.off += 2
	return v, nil
}

func (r *reader) u32() (uint32, error) {
	if r.remain() < 4 {
		return 0, io.ErrUnexpectedEOF
	}
	v := be.Uint32(r.b[r.off:])
	r.off += 4
	return v, nil
}

func (r *reader) i32() (int32, error) {
	u, err := r.u32()
	return int32(u), err
}

func (r *reader) i64() (int64, error) {
	if r.remain() < 8 {
		return 0, io.ErrUnexpectedEOF
	}
	v := int64(be.Uint64(r.b[r.off:]))
	r.off += 8
	return v, nil
}

func (r *reader) u64() (uint64, error) {
	if r.remain() < 8 {
		return 0, io.ErrUnexpectedEOF
	}
	v := be.Uint64(r.b[r.off:])
	r.off += 8
	return v, nil
}

func (r *reader) bytes(n int) ([]byte, error) {
	if n < 0 || r.remain() < n {
		return nil, io.ErrUnexpectedEOF
	}
	v := r.b[r.off : r.off+n]
	r.off += n
	return v, nil
}

func (r *reader) constant() (types.Value, error) {
	tag, err := r.u8()
	if err != nil {
		return nil, err
	}
	switch tag {
	case ConstNull:
		return nil, nil
	case ConstBool:
		b, err := r.u8()
		if err != nil {
			return nil, err
		}
		return values.Bool(b != 0), nil
	case ConstInt:
		i, err := r.i64()
		if err != nil {
			return nil, err
		}
		return values.Int(i), nil
	case ConstFloat:
		u, err := r.u64()
		if err != nil {
			return nil, err
		}
		return values.Float(math.Float64frombits(u)), nil
	case ConstString:
		n, err := r.u32()
		if err != nil {
			return nil, err
		}
		s, err := r.bytes(int(n))
		if err != nil {
			return nil, err
		}
		return values.String(s), nil
	default:
		return nil, fmt.Errorf("unknown constant tag %d", tag)
	}
}

func (r *reader) function() (*Function, error) {
	fn := &Function{}
	flags, err := r.u8()
	if err != nil {
		return nil, err
	}
	fn.Variadic = flags&FuncVariadic != 0
	if fn.Arity, err = r.u16(); err != nil {
		return nil, err
	}
	if fn.MaxStack, err = r.u16(); err != nil {
		return nil, err
	}
	nslot, err := r.u16()
	if err != nil {
		return nil, err
	}
	fn.SlotNames = make([]uint16, nslot)
	for i := 0; i < int(nslot); i++ {
		if fn.SlotNames[i], err = r.u16(); err != nil {
			return nil, err
		}
	}
	nuv, err := r.u16()
	if err != nil {
		return nil, err
	}
	if nuv > 255 {
		return nil, fmt.Errorf("too many upvalues")
	}
	fn.Upvalues = make([]UpvalueDesc, nuv)
	for i := 0; i < int(nuv); i++ {
		loc, err := r.u8()
		if err != nil {
			return nil, err
		}
		idx, err := r.u16()
		if err != nil {
			return nil, err
		}
		fn.Upvalues[i] = UpvalueDesc{IsLocal: loc != 0, Index: idx}
	}
	ncode, err := r.u32()
	if err != nil {
		return nil, err
	}
	fn.Bytecode = make([]Opcode, ncode)
	fn.Arguments = make([]int32, ncode)
	for i := 0; i < int(ncode); i++ {
		op, err := r.u8()
		if err != nil {
			return nil, err
		}
		if !ValidOpcode(Opcode(op)) {
			return nil, fmt.Errorf("unknown opcode %d", op)
		}
		fn.Bytecode[i] = Opcode(op)
		if fn.Arguments[i], err = r.i32(); err != nil {
			return nil, err
		}
	}
	fn.Spans = make([]uint16, ncode)
	for i := 0; i < int(ncode); i++ {
		if fn.Spans[i], err = r.u16(); err != nil {
			return nil, err
		}
	}
	return fn, nil
}

func appendU16(b []byte, v uint16) []byte {
	var tmp [2]byte
	be.PutUint16(tmp[:], v)
	return append(b, tmp[:]...)
}

func appendU32(b []byte, v uint32) []byte {
	var tmp [4]byte
	be.PutUint32(tmp[:], v)
	return append(b, tmp[:]...)
}

func appendI32(b []byte, v int32) []byte {
	return appendU32(b, uint32(v))
}

func appendU64(b []byte, v uint64) []byte {
	var tmp [8]byte
	be.PutUint64(tmp[:], v)
	return append(b, tmp[:]...)
}

func appendI64(b []byte, v int64) []byte {
	return appendU64(b, uint64(v))
}
