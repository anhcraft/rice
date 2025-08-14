package fun

import (
	"cmp"
	"reflect"
	"slices"
)

type ParamNode struct {
	argType  ArgType
	handler  reflect.Value
	children map[ArgType]*ParamNode

	// 0: hasHandler; 1: checkpoint; 2: contextual handler
	flag uint8

	// childrenNav denoting the precedence of children
	childrenNav []ArgType
}

func NewParamNode(argType ArgType) *ParamNode {
	var zero reflect.Value
	return &ParamNode{argType: argType, handler: zero, flag: 0, children: nil, childrenNav: nil}
}

func (p *ParamNode) ArgType() ArgType {
	return p.argType
}

func (p *ParamNode) Handler() reflect.Value {
	return p.handler
}

func (p *ParamNode) HasHandler() bool {
	return p.flag&1 != 0
}

func (p *ParamNode) markHasHandler() {
	p.flag |= 1
}

func (p *ParamNode) SupportCheckpoint() bool {
	return p.flag&2 != 0
}

func (p *ParamNode) markCheckpointSupport() {
	p.flag |= 2
}

func (p *ParamNode) IsContextual() bool {
	return p.flag&4 != 0
}

func (p *ParamNode) markContextual() {
	p.flag |= 4
}

func (p *ParamNode) GetOrCreateChild(arg ArgType, creator func() *ParamNode) *ParamNode {
	if p.children == nil {
		p.children = make(map[ArgType]*ParamNode)
		p.childrenNav = make([]ArgType, 0)
	}

	if _, ok := p.children[arg]; ok {
		return p.children[arg]
	}

	p.children[arg] = creator()
	p.childrenNav = append(p.childrenNav, arg)

	// TODO optimize
	slices.SortedFunc[ArgType](slices.Values(p.childrenNav), func(pi, pj ArgType) int {
		// Sort by dimension, then type ID
		// A special treat is to (any, 0) which accepts any kind of value regardless of dimension
		if pi.MatchAnyType() {
			pi = maxArgType // let (any, 0) have the lowest precedence
		}
		if pj.MatchAnyType() {
			pj = maxArgType
		}
		return cmp.Compare(pi, pj)
	})

	return p.children[arg]
}
