package set

import (
	"context"
	"fmt"
	"github.com/anhcraft/rice/exec/fun"
	"github.com/anhcraft/rice/exec/stdlib"
	"github.com/anhcraft/rice/exec/types"
	"github.com/anhcraft/rice/exec/types/values"
)

var Functions = fun.FunctionPackage{
	"new":     {stdlib.Define(New)},
	"of":      {stdlib.Define(Of)},
	"add":     {stdlib.Define(Add)},
	"include": {stdlib.Define(Include)},
	"map":     {stdlib.Define(Map)},
	"filter":  {stdlib.Define(Filter)},
	"remove":  {stdlib.Define(Remove)},
}

// New allocates a new, empty set.
//
// @return (*values.Set): A new set.
func New() (types.Value, error) {
	return values.NewSet(), nil
}

// Of allocates a new set with supplied values.
//
// @param items (...types.Value): The items.
// @return (*values.Set): A new set.
func Of(items ...types.Value) (types.Value, error) {
	st := values.NewSet()
	for _, item := range items {
		st.Add(item)
	}
	return st, nil
}

// Add adds one or multiple items to the set.
//
// @param li (*values.Set): The set to modify.
// @param items (...types.Value): The items to add.
// @return (*values.Set): The set itself.
func Add(st *values.Set, items ...types.Value) (types.Value, error) {
	for _, item := range items {
		st.Add(item)
	}
	return st, nil
}

// Include checks if a set contains a given value.
// Equality is determined by the `==` operator on the underlying values.
//
// @param li (*values.Set): The set to search in.
// @param valueToFind (types.Value): The value to search for.
// @return (values.Bool): `true` if the value is found, otherwise `false`.
func Include(st *values.Set, valueToFind types.Value) (types.Value, error) {
	return st.Contain(valueToFind), nil
}

// Map creates a new set populated with the results of calling a provided function on every element in the calling set.
//
// @param li (*values.Set): The set to iterate over.
// @param lambda (values.Func): The function to call for each element. It receives the element as an argument.
// @return (*values.Set): A new set with the mapped elements.
func Map(ctx context.Context, st *values.Set, lambda *values.Func) (types.Value, error) {
	clone := values.NewSet()
	for v := range st.Iterate() {
		r, err := lambda.Call(ctx, values.InternalCallSite, []types.Value{v})
		if err != nil {
			return nil, err
		}
		clone.Add(r)
	}

	return clone, nil
}

// Filter creates a new set with all elements that pass the test implemented by the provided function.
//
// @param li (*values.Set): The set to filter.
// @param lambda (values.Func): The function to test each element. It receives one element and should return
//
//	a value convertible to a boolean.
//
// @return (*values.Set): A new set with the elements that passed the test.
func Filter(ctx context.Context, st *values.Set, lambda *values.Func) (types.Value, error) {
	newSet := values.NewSet()

	for v := range st.Iterate() {
		r, err := lambda.Call(ctx, values.InternalCallSite, []types.Value{v})
		if err != nil {
			return nil, err
		}

		shouldKeep, err := values.AsBool(r)
		if err != nil {
			return nil, fmt.Errorf("filter predicate must return a boolean, but got %s", r.Type())
		}

		if shouldKeep {
			newSet.Add(v)
		}
	}

	return newSet, nil
}

// Remove removes the given item from the set.
//
// @param li (*values.Set): The set to modify.
// @param item (types.Value): The item to remove.
// @return (*values.Set): The set itself.
func Remove(st *values.Set, item types.Value) (types.Value, error) {
	st.Remove(item)
	return st, nil
}
