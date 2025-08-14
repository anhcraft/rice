package values

import (
	"errors"
	"iter"
	"rice/exec/types"
)

var _ NamespacedLayer = NativeFunctionLayer{}

type NativeFunctionLayer map[Identifier]NativeFunctionSet

func (n NativeFunctionLayer) Element(id types.Value) (types.Value, error) {
	if key, ok := id.(Identifier); ok {
		if v, ok := n[key]; ok {
			return v, nil
		} else {
			return nil, nil
		}
	} else {
		return nil, elementNotIdErr
	}
}

func (n NativeFunctionLayer) Size() Int {
	return Int(len(n))
}

func (n NativeFunctionLayer) Keys() []types.Value {
	k := make([]types.Value, len(n))
	j := 0
	for i := range n {
		k[j] = i
		j++
	}
	return k
}

func (n NativeFunctionLayer) Iterate() iter.Seq[types.Value] {
	return func(yield func(types.Value) bool) {
		for _, v := range n {
			if !yield(v) {
				return
			}
		}
	}
}

func (n NativeFunctionLayer) PutElement(id types.Value, item types.Value) error {
	return errors.New("unsupported operation")
}
