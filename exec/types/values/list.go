package values

import (
	"iter"
	"rice/exec/types"
	"sort"
)

var _ = types.List.DefineType((*List)(nil))
var _ IndexedCollection = (*List)(nil)

type List struct {
	arr []types.Value
}

func NewList() *List {
	return &List{arr: make([]types.Value, 0)}
}

func ListOf[E types.Value](arr []E) *List {
	cop := make([]types.Value, len(arr))
	for i := 0; i < len(arr); i++ {
		cop[i] = arr[i]
	}
	return &List{arr: cop}
}

func (l *List) String() string {
	return "List"
}

func (l *List) Type() types.Type {
	return types.List
}

func (l *List) Element(id types.Value) (types.Value, error) {
	if i, ok := id.(Int); ok {
		if i < 0 || int(i) >= len(l.arr) {
			return nil, indexOutOfBoundErr
		}
		return l.arr[i], nil
	} else {
		return nil, elementNotIntErr
	}
}

func (l *List) Size() Int {
	return Int(len(l.arr))
}

func (l *List) Keys() []types.Value {
	k := make([]types.Value, len(l.arr))
	for i := 0; i < len(k); i++ {
		k[i] = Int(i)
	}
	return k
}

func (l *List) Iterate() iter.Seq[types.Value] {
	return func(yield func(types.Value) bool) {
		for _, v := range l.arr {
			if !yield(v) {
				return
			}
		}
	}
}

func (l *List) PutElement(id types.Value, item types.Value) error {
	if i, ok := id.(Int); ok {
		if i < 0 || int(i) >= len(l.arr) {
			return indexOutOfBoundErr
		}
		l.arr[i] = item
		return nil
	} else {
		return elementNotIntErr
	}
}

func (l *List) At(id Int) types.Value {
	return l.arr[id]
}

func (l *List) Append(v types.Value) {
	l.arr = append(l.arr, v)
}

func (l *List) PrependAll(v []types.Value) {
	l.arr = append(v, l.arr...)
}

func (l *List) AppendAll(v []types.Value) {
	l.arr = append(l.arr, v...)
}

func (l *List) Sort(cmp func(i, j int) bool) {
	sort.Slice(l.arr, cmp)
}

func (l *List) Reverse() {
	n := len(l.arr)
	for i := 0; i < n/2; i++ {
		j := n - i - 1
		l.arr[i], l.arr[j] = l.arr[j], l.arr[i]
	}
}

func (l *List) RemoveAt(i Int) error {
	if i < 0 || int(i) >= len(l.arr) {
		return indexOutOfBoundErr
	}
	clone := make([]types.Value, i)
	for i := 0; i < len(clone); i++ {
		clone[i] = l.arr[i]
	}
	l.arr = append(clone, l.arr[i+1:]...)
	return nil
}

func (l *List) RemoveAll(v types.Value) Int {
	deleted := 0
	clone := make([]types.Value, 0, len(l.arr))
	for i := 0; i < len(l.arr); i++ {
		if l.arr[i] == v {
			deleted++
			continue
		}
		clone = append(clone, l.arr[i])
	}
	l.arr = clone
	return Int(deleted)
}
