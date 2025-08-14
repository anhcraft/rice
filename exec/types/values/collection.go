package values

import (
	"github.com/anhcraft/rice/exec/types"
	"iter"
)

type Collection interface {
	Size() Int
	Iterate() iter.Seq[types.Value]
}

type IndexedCollection interface {
	Collection
	Keys() []types.Value
	Element(id types.Value) (types.Value, error)
	PutElement(id types.Value, item types.Value) error
}
