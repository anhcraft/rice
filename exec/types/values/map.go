package values

import (
	"iter"
	"rice/exec/types"
)

var _ = types.Map.DefineType((*Map)(nil))
var _ IndexedCollection = (*Map)(nil)

type Map struct {
	hmap map[types.Value]types.Value
}

func NewMap() *Map {
	return &Map{hmap: make(map[types.Value]types.Value)}
}

func (m *Map) String() string {
	return "Map"
}

func (m *Map) Type() types.Type {
	return types.Map
}

func (m *Map) Size() Int {
	return Int(len(m.hmap))
}

func (m *Map) Iterate() iter.Seq[types.Value] {
	return func(yield func(types.Value) bool) {
		for k, v := range m.hmap {
			entry := []types.Value{k, v}
			if !yield(ListOf(entry)) {
				return
			}
		}
	}
}

func (m *Map) Put(key types.Value, value types.Value) {
	m.hmap[key] = value
}

func (m *Map) Get(key types.Value) (types.Value, Bool) {
	val, ok := m.hmap[key]
	return val, Bool(ok)
}

func (m *Map) Remove(key types.Value) {
	delete(m.hmap, key)
}

func (m *Map) Keys() []types.Value {
	ks := make([]types.Value, 0)
	for k := range m.hmap {
		ks = append(ks, k)
	}
	return ks
}

func (m *Map) Element(id types.Value) (types.Value, error) {
	if ident, ok := id.(Identifier); ok {
		id = String(ident)
	}
	return m.hmap[id], nil
}

func (m *Map) PutElement(id types.Value, item types.Value) error {
	m.hmap[id] = item
	return nil
}
