package values

import (
	"github.com/anhcraft/rice/exec/types"
	"github.com/anhcraft/rice/lib/set"
	"iter"
)

var _ = types.Set.DefineType((*Set)(nil))
var _ Collection = (*Set)(nil)

type Set struct {
	hs *set.Set[types.Value]
}

func NewSet() *Set {
	return &Set{hs: set.NewSet[types.Value]()}
}

func SetOf[E types.Value](arr []E) *Set {
	st := NewSet()
	for _, v := range arr {
		st.hs.Add(v)
	}
	return st
}

func (s *Set) String() string {
	return "Set"
}

func (s *Set) Type() types.Type {
	return types.Set
}

func (s *Set) Size() Int {
	return Int(s.hs.Size())
}

func (s *Set) Iterate() iter.Seq[types.Value] {
	return s.hs.Iterate()
}

func (s *Set) Add(item types.Value) {
	s.hs.Add(item)
}

func (s *Set) Contain(item types.Value) Bool {
	return Bool(s.hs.Has(item))
}

func (s *Set) Remove(item types.Value) {
	s.hs.Remove(item)
}

func (s *Set) AsList() *List {
	list := NewList()
	for v := range s.hs.Iterate() {
		list.Append(v)
	}
	return list
}
