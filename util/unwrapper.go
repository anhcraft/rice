package util

import (
	"fmt"

	"github.com/anhcraft/rice/exec/types"
	"github.com/anhcraft/rice/exec/types/values"
)

func Unwrap(v types.Value) (any, error) {
	switch v.Type() {
	case types.Int:
		return int64(v.(values.Int)), nil
	case types.Float:
		return float64(v.(values.Float)), nil
	case types.Bool:
		return bool(v.(values.Bool)), nil
	case types.String:
		return string(v.(values.String)), nil
	case types.List:
		return UnwrapCollection(v.(values.Collection))
	case types.Set:
		return UnwrapCollection(v.(values.Collection))
	case types.Map:
		return UnwrapIndexedCollection(v.(values.IndexedCollection))
	case types.Func:
		return nil, nil
	case types.Identifier:
		return string(v.(values.Identifier)), nil
	case types.Namespace:
		return UnwrapIndexedCollection(v.(values.IndexedCollection))
	case types.NativeFuncSet:
		return nil, nil
	case types.Selector:
		return UnwrapSelector(v.(values.Selector))
	}
	return nil, nil
}

func UnwrapCollection(collection values.Collection) (any, error) {
	list := make([]any, collection.Size())
	i := 0
	for elem := range collection.Iterate() {
		v, e := Unwrap(elem)
		if e != nil {
			return nil, e
		}
		list[i] = v
		i++
	}
	return list, nil
}

func UnwrapIndexedCollection(ic values.IndexedCollection) (any, error) {
	hmap := make(map[string]any)
	for _, key := range ic.Keys() {
		v, e := ic.Element(key)
		if e != nil {
			return nil, e
		}
		v2, e2 := Unwrap(v)
		if e2 != nil {
			return nil, e2
		}
		hmap[fmt.Sprint(key)] = v2
	}
	return hmap, nil
}

func UnwrapSelector(selector values.Selector) (any, error) {
	v, e := selector.Get()
	if e != nil {
		return nil, e
	}
	return Unwrap(v)
}
