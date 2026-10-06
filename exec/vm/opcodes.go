package vm

// Opcode numbers are part of the RICE wire format (docs/bytecode.md).
// Do not renumber; assign new opcodes to free slots and bump the minor version.
type Opcode uint8

const (
	OpPush Opcode = 1
	OpPop  Opcode = 2
	OpDup  Opcode = 3
	OpDup2 Opcode = 4
	OpRot3 Opcode = 5

	OpLoadLocal          Opcode = 16
	OpStoreLocal         Opcode = 17
	OpDefineLocal        Opcode = 18
	OpDefineConstLocal   Opcode = 19
	OpLoadUpvalue        Opcode = 20
	OpStoreUpvalue       Opcode = 21
	OpDefineUpvalue      Opcode = 22
	OpDefineConstUpvalue Opcode = 23
	OpCloseUpvalue       Opcode = 24

	OpLoadGlobal        Opcode = 32
	OpStoreGlobal       Opcode = 33
	OpDefineGlobal      Opcode = 34
	OpDefineConstGlobal Opcode = 35

	OpJump            Opcode = 48
	OpJumpBackward    Opcode = 49
	OpJumpIfFalse     Opcode = 50
	OpJumpIfFalsey    Opcode = 51
	OpJumpIfBoolFalse Opcode = 52
	OpJumpIfBoolTrue  Opcode = 53

	OpNegate      Opcode = 64
	OpNot         Opcode = 65
	OpAdd         Opcode = 66
	OpSubtract    Opcode = 67
	OpMultiply    Opcode = 68
	OpDivide      Opcode = 69
	OpModulo      Opcode = 70
	OpEqual       Opcode = 71
	OpLess        Opcode = 72
	OpMore        Opcode = 73
	OpLessOrEqual Opcode = 74
	OpMoreOrEqual Opcode = 75
	OpAnd         Opcode = 76
	OpOr          Opcode = 77
	OpIncrement   Opcode = 78

	OpArray      Opcode = 96
	OpMap        Opcode = 97
	OpGetIndex   Opcode = 98
	OpSetIndex   Opcode = 99
	OpLoadField  Opcode = 100
	OpStoreField Opcode = 101

	OpClosure  Opcode = 112
	OpCall     Opcode = 113
	OpArgs     Opcode = 114
	OpArg      Opcode = 115
	OpSpread   Opcode = 116
	OpCallArgs Opcode = 117
	OpReturn   Opcode = 118

	OpBegin          Opcode = 128
	OpJumpIfEnd      Opcode = 129
	OpPointer        Opcode = 130
	OpIncrementIndex Opcode = 131
	OpEnd            Opcode = 132

	OpEnter Opcode = 144
	OpLeave Opcode = 145
)

const (
	Magic uint32 = 0x52494345 // 'RICE'
	None  uint16 = 0xFFFF
)

const (
	HotspotOpensScope byte = 1 << 0
	FuncVariadic      byte = 1 << 0
)

const (
	ConstNull   byte = 1
	ConstBool   byte = 2
	ConstInt    byte = 3
	ConstFloat  byte = 4
	ConstString byte = 5
)

type opcodeMeta struct {
	name  string
	since Version
}

type flagMeta struct {
	bit   byte
	since Version
}

var opcodeInfo = map[Opcode]opcodeMeta{
	OpPush: {name: "OpPush", since: v1_0}, OpPop: {name: "OpPop", since: v1_0},
	OpDup: {name: "OpDup", since: v1_0}, OpDup2: {name: "OpDup2", since: v1_0},
	OpRot3:      {name: "OpRot3", since: v1_0},
	OpLoadLocal: {name: "OpLoadLocal", since: v1_0}, OpStoreLocal: {name: "OpStoreLocal", since: v1_0},
	OpDefineLocal: {name: "OpDefineLocal", since: v1_0}, OpDefineConstLocal: {name: "OpDefineConstLocal", since: v1_0},
	OpLoadUpvalue: {name: "OpLoadUpvalue", since: v1_0}, OpStoreUpvalue: {name: "OpStoreUpvalue", since: v1_0},
	OpDefineUpvalue: {name: "OpDefineUpvalue", since: v1_0}, OpDefineConstUpvalue: {name: "OpDefineConstUpvalue", since: v1_0},
	OpCloseUpvalue: {name: "OpCloseUpvalue", since: v1_0},
	OpLoadGlobal:   {name: "OpLoadGlobal", since: v1_0}, OpStoreGlobal: {name: "OpStoreGlobal", since: v1_0},
	OpDefineGlobal: {name: "OpDefineGlobal", since: v1_0}, OpDefineConstGlobal: {name: "OpDefineConstGlobal", since: v1_0},
	OpJump: {name: "OpJump", since: v1_0}, OpJumpBackward: {name: "OpJumpBackward", since: v1_0},
	OpJumpIfFalse: {name: "OpJumpIfFalse", since: v1_0}, OpJumpIfFalsey: {name: "OpJumpIfFalsey", since: v1_0},
	OpJumpIfBoolFalse: {name: "OpJumpIfBoolFalse", since: v1_0}, OpJumpIfBoolTrue: {name: "OpJumpIfBoolTrue", since: v1_0},
	OpNegate: {name: "OpNegate", since: v1_0}, OpNot: {name: "OpNot", since: v1_0},
	OpAdd: {name: "OpAdd", since: v1_0}, OpSubtract: {name: "OpSubtract", since: v1_0},
	OpMultiply: {name: "OpMultiply", since: v1_0}, OpDivide: {name: "OpDivide", since: v1_0},
	OpModulo: {name: "OpModulo", since: v1_0}, OpEqual: {name: "OpEqual", since: v1_0},
	OpLess: {name: "OpLess", since: v1_0}, OpMore: {name: "OpMore", since: v1_0},
	OpLessOrEqual: {name: "OpLessOrEqual", since: v1_0}, OpMoreOrEqual: {name: "OpMoreOrEqual", since: v1_0},
	OpAnd: {name: "OpAnd", since: v1_0}, OpOr: {name: "OpOr", since: v1_0},
	OpIncrement: {name: "OpIncrement", since: v1_0},
	OpArray:     {name: "OpArray", since: v1_0}, OpMap: {name: "OpMap", since: v1_0},
	OpGetIndex: {name: "OpGetIndex", since: v1_0}, OpSetIndex: {name: "OpSetIndex", since: v1_0},
	OpLoadField: {name: "OpLoadField", since: v1_0}, OpStoreField: {name: "OpStoreField", since: v1_0},
	OpClosure: {name: "OpClosure", since: v1_0}, OpCall: {name: "OpCall", since: v1_0},
	OpArgs: {name: "OpArgs", since: v1_0}, OpArg: {name: "OpArg", since: v1_0},
	OpSpread: {name: "OpSpread", since: v1_0}, OpCallArgs: {name: "OpCallArgs", since: v1_0},
	OpReturn: {name: "OpReturn", since: v1_0},
	OpBegin:  {name: "OpBegin", since: v1_0}, OpJumpIfEnd: {name: "OpJumpIfEnd", since: v1_0},
	OpPointer: {name: "OpPointer", since: v1_0}, OpIncrementIndex: {name: "OpIncrementIndex", since: v1_0},
	OpEnd:   {name: "OpEnd", since: v1_0},
	OpEnter: {name: "OpEnter", since: v1_0}, OpLeave: {name: "OpLeave", since: v1_0},
}

var constTagSince = map[byte]Version{
	ConstNull:   v1_0,
	ConstBool:   v1_0,
	ConstInt:    v1_0,
	ConstFloat:  v1_0,
	ConstString: v1_0,
}

var hotspotFlagSince = []flagMeta{
	{bit: HotspotOpensScope, since: v1_0},
}

var funcFlagSince = []flagMeta{
	{bit: FuncVariadic, since: v1_0},
}

func (op Opcode) String() string {
	if m, ok := opcodeInfo[op]; ok {
		return m.name
	}
	return "OpUnknown"
}

func opcodeSince(op Opcode) (Version, bool) {
	m, ok := opcodeInfo[op]
	return m.since, ok
}

func ValidOpcode(op Opcode) bool {
	_, ok := opcodeSince(op)
	return ok
}
