package profiler

import (
	"rice/exec/ast"
	"rice/lib/stack"
	"time"
)

type Profiler interface {
	Start(node ast.Hotspot)
	End()
	Reset()
	Report() string
}

var _ Profiler = Muted{}
var _ Profiler = &Impl{}

type Muted struct{}

func NewMuted() Muted {
	return Muted{}
}

func (m Muted) Start(_ ast.Hotspot) {}

func (m Muted) End() {}

func (m Muted) Reset() {}

func (m Muted) Report() string {
	return "(Muted)"
}

type Impl struct {
	// stack profile records in the current call stack
	stack *stack.Stack[*ProfileRecord]

	// depth how depth the current call stack is
	depth int

	root *ProfileRecord

	// pool reusable profile record pool; gives better performance than sync.Pool
	pool *stack.Stack[*ProfileRecord]
}

func NewImpl() *Impl {
	return &Impl{
		stack: stack.New[*ProfileRecord](16),
		pool:  stack.New[*ProfileRecord](16),
	}
}

func (p *Impl) getOrAllocateRecord() *ProfileRecord {
	if p.pool.IsEmpty() {
		return &ProfileRecord{
			Children: make([]*ProfileRecord, 0, 4),
		}
	} else {
		v, _ := p.pool.Pop()
		return v
	}
}

func (p *Impl) getRecord(label string, pos ast.Pos, depth int, start time.Time) *ProfileRecord {
	record := p.getOrAllocateRecord()

	record.Label = label
	record.Pos = pos
	record.Depth = depth
	record.StartTime = start

	return record
}

func (p *Impl) Start(node ast.Hotspot) {
	p.depth++
	record := p.getRecord(node.Label(), node.StartPos(), p.depth, time.Now())

	if !p.stack.IsEmpty() {
		top, ok := p.stack.Peek()
		if ok {
			top.AddChild(record)
		}
	} else {
		p.root = record
	}
	p.stack.Push(record)
}

func (p *Impl) End() {
	if p.stack.IsEmpty() {
		panic("End() called on empty profiler")
	}

	top, _ := p.stack.Pop()
	top.StopClock()
	p.depth--
}

func (p *Impl) putRecord(record *ProfileRecord) {
	for _, child := range record.Children {
		p.putRecord(child)
	}
	record.Reset()
	p.pool.Push(record)
}

func (p *Impl) Reset() {
	if p.root != nil {
		p.putRecord(p.root)
	}
	p.root = nil
	p.depth = 0
	p.stack.Clear()
}
