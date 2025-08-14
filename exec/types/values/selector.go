package values

import (
	"fmt"
	"rice/exec/types"
)

var _ = types.Selector.DefineType(Selector{})

type Selector struct {
	object Collection
	elem   types.Value
}

func NewSelector(o Collection, e types.Value) Selector {
	return Selector{object: o, elem: e}
}

func (s Selector) Type() types.Type {
	return types.Selector
}

func (s Selector) Object() Collection {
	return s.object
}

func (s Selector) Elem() types.Value {
	return s.elem
}

func (s Selector) Get() (types.Value, error) {
	if o, ok := s.object.(IndexedCollection); ok {
		return o.Element(s.elem)
	}
	return nil, fmt.Errorf("type %T is not indexable", s.object)
}

func (s Selector) Put(v types.Value) error {
	if o, ok := s.object.(IndexedCollection); ok {
		return o.PutElement(s.elem, v)
	}
	return fmt.Errorf("type %T is not indexable", s.object)
}

func (s Selector) String() string {
	return fmt.Sprintf("Selector(%v, %v)", s.object, s.elem)
}
