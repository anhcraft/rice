package _map

import (
	"context"
	"fmt"
	"rice/exec/fun"
	"rice/exec/stdlib"
	"rice/exec/types"
	"rice/exec/types/values"
)

var Functions = fun.FunctionPackage{
	"new":        {stdlib.Define(New)},
	"of":         {stdlib.Define(Of)},
	"put":        {stdlib.Define(Put)},
	"remove":     {stdlib.Define(Remove)},
	"includeKey": {stdlib.Define(IncludeKey)},
	"keys":       {stdlib.Define(Keys)},
	"values":     {stdlib.Define(Values)},
	"entries":    {stdlib.Define(Entries)},
	"map":        {stdlib.Define(Map)},
	"filter":     {stdlib.Define(Filter)},
}

// New creates a new, empty map.
//
// @return (*values.Map): A new, empty map.
func New() (types.Value, error) {
	return values.NewMap(), nil
}

// Of allocates a new map with supplied entries.
//
// @param kvs (...types.Value): A sequence of key-value pairs.
// @return (*values.Map): A new map.
func Of(kvs ...types.Value) (types.Value, error) {
	m := values.NewMap()
	if len(kvs)%2 != 0 {
		return nil, fmt.Errorf("put requires an even number of arguments for key-value pairs, but got %d", len(kvs))
	}
	for i := 0; i < len(kvs); i += 2 {
		key := kvs[i]
		value := kvs[i+1]
		m.Put(key, value)
	}
	return m, nil
}

// Put adds one or more key-value pairs to the map.
// The arguments must be provided in pairs (key1, value1, key2, value2, ...).
//
// @param m (*values.Map): The map to modify.
// @param kvs (...types.Value): A sequence of key-value pairs.
// @return (*values.Map): The modified map.
func Put(m *values.Map, kvs ...types.Value) (types.Value, error) {
	if len(kvs)%2 != 0 {
		return nil, fmt.Errorf("put requires an even number of arguments for key-value pairs, but got %d", len(kvs))
	}
	for i := 0; i < len(kvs); i += 2 {
		key := kvs[i]
		value := kvs[i+1]
		m.Put(key, value)
	}
	return m, nil
}

// Remove removes one or more entries from the map by their keys.
//
// @param m (*values.Map): The map to modify.
// @param keys (...types.Value): The keys to remove.
// @return (*values.Map): The modified map.
func Remove(m *values.Map, keys ...types.Value) (types.Value, error) {
	for _, key := range keys {
		m.Remove(key)
	}
	return m, nil
}

// IncludeKey checks if the map contains a given key.
//
// @param m (*values.Map): The map to check.
// @param key (types.Value): The key to search for.
// @return (values.Bool): `true` if the key exists, otherwise `false`.
func IncludeKey(m *values.Map, key types.Value) (types.Value, error) {
	_, found := m.Get(key)
	return found, nil
}

// Keys returns a set of all keys in the map. The order is not guaranteed.
//
// @param m (*values.Map): The map to get keys from.
// @return (*values.List): A set containing all keys.
func Keys(m *values.Map) (types.Value, error) {
	keySet := values.NewSet()
	for entry := range m.Iterate() {
		key := entry.(*values.List).At(0)
		keySet.Add(key)
	}
	return keySet, nil
}

// Values returns a list of all values in the map. The order is not guaranteed.
//
// @param m (*values.Map): The map to get values from.
// @return (*values.List): A list containing all values.
func Values(m *values.Map) (types.Value, error) {
	valueList := values.NewList()
	for entry := range m.Iterate() {
		value := entry.(*values.List).At(1)
		valueList.Append(value)
	}
	return valueList, nil
}

// Entries returns a list of all entries in the map. Each entry is a list of `[key, value]`.
// The order is not guaranteed.
//
// @param m (*values.Map): The map to get entries from.
// @return (*values.List): A list of entries.
func Entries(m *values.Map) (types.Value, error) {
	entryList := values.NewList()
	for entry := range m.Iterate() {
		entryList.Append(entry)
	}
	return entryList, nil
}

// Map creates a new map with the results of calling a provided function on every entry.
//
// @param m (*values.Map): The map to iterate over.
// @param lambda (*values.Func): A function that accepts an entry (`*values.List` of `[key, value]`)
//
//	and returns a new entry (`*values.List` of `[newKey, newValue]`).
//
// @return (*values.Map): A new map with the transformed entries.
func Map(ctx context.Context, m *values.Map, lambda *values.Func) (types.Value, error) {
	newMap := values.NewMap()
	for entry := range m.Iterate() {
		result, err := lambda.Call(ctx, values.InternalCallSite, []types.Value{entry})
		if err != nil {
			return nil, err
		}

		newEntry, ok := result.(*values.List)
		if !ok || newEntry.Size() != 2 {
			return nil, fmt.Errorf("map lambda must return a List of [key, value], but got %s", result.Type())
		}

		newMap.Put(newEntry.At(0), newEntry.At(1))
	}
	return newMap, nil
}

// Filter creates a new map with all entries that pass a test.
//
// @param m (*values.Map): The map to filter.
// @param lambda (*values.Func): A predicate function that accepts an entry (`*values.List` of `[key, value]`)
//
//	and returns a boolean value.
//
// @return (*values.Map): A new map containing only the entries that passed the test.
func Filter(ctx context.Context, m *values.Map, lambda *values.Func) (types.Value, error) {
	newMap := values.NewMap()
	for entry := range m.Iterate() {
		result, err := lambda.Call(ctx, values.InternalCallSite, []types.Value{entry})
		if err != nil {
			return nil, err
		}

		shouldKeep, err := values.AsBool(result)
		if err != nil {
			return nil, fmt.Errorf("filter predicate must return a boolean, but got %s", result.Type())
		}

		if shouldKeep {
			entryList := entry.(*values.List)
			newMap.Put(entryList.At(0), entryList.At(1))
		}
	}
	return newMap, nil
}
