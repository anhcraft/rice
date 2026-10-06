package vm

import (
	"bytes"
	"encoding/binary"
	"strings"
	"testing"

	"github.com/anhcraft/rice/exec/types"
	"github.com/anhcraft/rice/exec/types/values"
)

func sampleModule() *Module {
	return &Module{
		Constants: []types.Value{
			nil,
			values.Bool(true),
			values.Int(42),
			values.Float(1.5),
			values.String("hi"),
		},
		Spans: []Span{
			{Line: 1, Index: 0, EndLine: 1, EndIndex: 2, Parent: None, Label: 4},
		},
		Hotspots: []Hotspot{
			{Label: 4, Span: 0, Flags: HotspotOpensScope},
		},
		Functions: []*Function{
			{
				Variadic:  false,
				Arity:     0,
				MaxStack:  2,
				SlotNames: []uint16{None},
				Upvalues:  []UpvalueDesc{{IsLocal: true, Index: 0}},
				Bytecode:  []Opcode{OpPush, OpEnter, OpReturn},
				Arguments: []int32{2, 0, 0},
				Spans:     []uint16{0, 0, 0},
			},
		},
		Entry: 0,
	}
}

func TestModuleRoundTrip(t *testing.T) {
	mod := sampleModule()

	raw, err := mod.Encode()
	if err != nil {
		t.Fatal(err)
	}
	if binary.BigEndian.Uint32(raw[:4]) != Magic {
		t.Fatalf("magic 0x%08x", binary.BigEndian.Uint32(raw[:4]))
	}
	if binary.BigEndian.Uint16(raw[4:6]) != 1 || binary.BigEndian.Uint16(raw[6:8]) != 0 {
		t.Fatalf("version %d.%d", binary.BigEndian.Uint16(raw[4:6]), binary.BigEndian.Uint16(raw[6:8]))
	}
	got, err := Decode(raw)
	if err != nil {
		t.Fatal(err)
	}
	raw2, err := got.Encode()
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(raw, raw2) {
		t.Fatalf("encode not stable")
	}
	if got.Entry != 0 || len(got.Functions) != 1 {
		t.Fatalf("functions")
	}
	if got.Functions[0].Bytecode[0] != OpPush {
		t.Fatalf("opcode")
	}
	if s, ok := got.Constants[4].(values.String); !ok || s != "hi" {
		t.Fatalf("string const %v", got.Constants[4])
	}
	if got.Version != (Version{Major: 1, Minor: 0}) {
		t.Fatalf("decoded version %s", got.Version)
	}
	if got.Disassemble() == "" {
		t.Fatal("disassemble")
	}
	if !strings.Contains(got.Disassemble(), "version 1.0") {
		t.Fatalf("disasm missing version:\n%s", got.Disassemble())
	}
}

func TestRequiredVersionV10(t *testing.T) {
	mod := sampleModule()
	if got := mod.RequiredVersion(); got != v1_0 {
		t.Fatalf("RequiredVersion = %s, want 1.0", got)
	}
}

func TestDecodeRejectsNewerMajor(t *testing.T) {
	raw := []byte{0x52, 0x49, 0x43, 0x45, 0x00, 0x02, 0x00, 0x00}
	_, err := Decode(raw)
	if err == nil || !strings.Contains(err.Error(), "needs RICE 2.0") {
		t.Fatalf("got %v", err)
	}
}

func TestDecodeRejectsNewerMinor(t *testing.T) {
	raw := []byte{0x52, 0x49, 0x43, 0x45, 0x00, 0x01, 0x00, 0x01}
	_, err := Decode(raw)
	if err == nil || !strings.Contains(err.Error(), "needs RICE 1.1") {
		t.Fatalf("got %v", err)
	}
}

func TestDecodeAcceptsOlderMinor(t *testing.T) {
	prev := CurrentVersion
	CurrentVersion = Version{Major: 1, Minor: 1}
	t.Cleanup(func() { CurrentVersion = prev })

	raw, err := sampleModule().Encode()
	if err != nil {
		t.Fatal(err)
	}
	if binary.BigEndian.Uint16(raw[6:8]) != 0 {
		t.Fatalf("stamped minor %d, want 0", binary.BigEndian.Uint16(raw[6:8]))
	}
	got, err := Decode(raw)
	if err != nil {
		t.Fatal(err)
	}
	if got.Version != v1_0 {
		t.Fatalf("got %s", got.Version)
	}
}

func TestDecodeRejectsBelowMinMajor(t *testing.T) {
	prev := MinMajor
	MinMajor = 2
	t.Cleanup(func() { MinMajor = prev })

	raw, err := sampleModule().Encode()
	if err != nil {
		t.Fatal(err)
	}
	_, err = Decode(raw)
	if err == nil || !strings.Contains(err.Error(), "requires major >= 2") {
		t.Fatalf("got %v", err)
	}
}

func TestVerifyRejectsMislabeledOpcode(t *testing.T) {
	const fake Opcode = 200
	opcodeInfo[fake] = opcodeMeta{name: "OpFake", since: Version{Major: 1, Minor: 1}}
	t.Cleanup(func() { delete(opcodeInfo, fake) })

	mod := sampleModule()
	mod.Version = v1_0
	mod.Functions[0].Bytecode = []Opcode{fake, OpReturn}
	mod.Functions[0].Arguments = []int32{0, 0}
	mod.Functions[0].Spans = []uint16{0, 0}
	err := verifyVersion(mod)
	if err == nil || !strings.Contains(err.Error(), "requires RICE 1.1") {
		t.Fatalf("got %v", err)
	}
}

func TestSupports(t *testing.T) {
	v := Version{Major: 1, Minor: 2}
	if !v.Supports(Version{1, 0}) || !v.Supports(Version{1, 2}) {
		t.Fatal("should accept same-major older-or-equal minor")
	}
	if v.Supports(Version{1, 3}) {
		t.Fatal("should reject newer minor")
	}
	if v.Supports(Version{2, 0}) {
		t.Fatal("should reject newer major")
	}

	prev := MinMajor
	MinMajor = 2
	t.Cleanup(func() { MinMajor = prev })
	v2 := Version{Major: 2, Minor: 0}
	if v2.Supports(Version{1, 0}) {
		t.Fatal("should reject below MinMajor")
	}
}
