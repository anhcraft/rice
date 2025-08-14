package list

import (
	"context"
	"fmt"
	"rice/exec/fun"
	"rice/exec/stdlib"
	"rice/exec/types"
	"rice/exec/types/values"
)

var Functions = fun.FunctionPackage{
	"new":       {stdlib.Define(New)},
	"of":        {stdlib.Define(Of)},
	"prepend":   {stdlib.Define(Prepend)},
	"append":    {stdlib.Define(Append)},
	"include":   {stdlib.Define(Include)},
	"index":     {stdlib.Define(Index)},
	"lastIndex": {stdlib.Define(LastIndex)},
	"sort":      {stdlib.Define(Sort)},
	"reverse":   {stdlib.Define(Reverse)},
	"map":       {stdlib.Define(Map)},
	"filter":    {stdlib.Define(Filter)},
	"removeAt":  {stdlib.Define(RemoveAt)},
	"removeAll": {stdlib.Define(RemoveAll)},
	"slice":     {stdlib.Define(Slice), stdlib.Define(Slice0)},
}

// New allocates a new, empty list.
//
// @return (*values.List): A new list.
func New() (types.Value, error) {
	return values.NewList(), nil
}

// Of allocates a new list with supplied values.
//
// @param items (...types.Value): The items.
// @return (*values.List): A new list.
func Of(items ...types.Value) (types.Value, error) {
	li := values.NewList()
	li.AppendAll(items)
	return li, nil
}

// Prepend prepends one or multiple items to the head of the list.
//
// @param li (*values.List): The list to modify.
// @param items (...types.Value): The items to prepend.
// @return (*values.List): The list itself.
func Prepend(li *values.List, items ...types.Value) (types.Value, error) {
	li.PrependAll(items)
	return li, nil
}

// Append appends one or multiple items to the end of the list.
//
// @param li (*values.List): The list to modify.
// @param items (...types.Value): The items to append.
// @return (*values.List): The list itself.
func Append(li *values.List, items ...types.Value) (types.Value, error) {
	li.AppendAll(items)
	return li, nil
}

// Include checks if a list contains a given value.
// Equality is determined by the `==` operator on the underlying values.
//
// @param li (*values.List): The list to search in.
// @param valueToFind (types.Value): The value to search for.
// @return (values.Bool): `true` if the value is found, otherwise `false`.
func Include(li *values.List, valueToFind types.Value) (types.Value, error) {
	size := li.Size()
	for i := values.Int(0); i < size; i++ {
		if li.At(i) == valueToFind {
			return values.Bool(true), nil
		}
	}
	return values.Bool(false), nil
}

// Index finds the first index of a value in a list.
//
// @param li (*values.List): The list to search in.
// @param valueToFind (types.Value): The value to search for.
// @return (values.Int): The zero-based index of the first occurrence, or -1 if not found.
func Index(li *values.List, valueToFind types.Value) (types.Value, error) {
	size := li.Size()
	for i := values.Int(0); i < size; i++ {
		if li.At(i) == valueToFind {
			return i, nil
		}
	}
	return values.Int(-1), nil
}

// LastIndex finds the last index of a value in a list.
//
// @param li (*values.List): The list to search in.
// @param valueToFind (types.Value): The value to search for.
// @return (values.Int): The zero-based index of the last occurrence, or -1 if not found.
func LastIndex(li *values.List, valueToFind types.Value) (types.Value, error) {
	for i := li.Size() - 1; i >= 0; i-- {
		if li.At(i) == valueToFind {
			return i, nil
		}
	}
	return values.Int(-1), nil
}

// Sort sorts the list in-place.
//
// @param li (*values.List): The list to sort.
// @param lambda (values.Func): A comparator function that takes two arguments (a, b)
//
//	and returns `true` if `a` should come before `b`.
//
// @return (*values.List): The list itself.
func Sort(ctx context.Context, li *values.List, lambda *values.Func) (types.Value, error) {
	size := li.Size()
	if size < 2 {
		return li, nil
	}

	var sortErr error
	li.Sort(func(i, j int) bool {
		if sortErr != nil {
			return false
		}

		resultVal, err := lambda.Call(ctx, values.InternalCallSite, []types.Value{
			li.At(values.Int(i)),
			li.At(values.Int(j)),
		})
		if err != nil {
			sortErr = err
			return false
		}

		isLess, err := values.AsBool(resultVal)
		if err != nil {
			sortErr = fmt.Errorf("comparator must return a boolean, but got %s", resultVal.Type())
			return false
		}
		return bool(isLess)
	})

	if sortErr != nil {
		return nil, sortErr
	}

	return li, nil
}

// Reverse Reverses the list in-place.
//
// @param li (*values.List): The list to reverse.
//
// @return (*values.List): The list itself.
func Reverse(li *values.List) (types.Value, error) {
	li.Reverse()
	return li, nil
}

// Map creates a new list populated with the results of calling a provided function on every element in the calling list.
//
// @param li (*values.List): The list to iterate over.
// @param lambda (values.Func): The function to call for each element. It receives the element as an argument.
// @return (*values.List): A new list with the mapped elements.
func Map(ctx context.Context, li *values.List, lambda *values.Func) (types.Value, error) {
	var err error

	clone := values.NewList()
	for i := values.Int(0); i < li.Size(); i++ {
		v := li.At(i)
		v, err = lambda.Call(ctx, values.InternalCallSite, []types.Value{v})
		if err != nil {
			return nil, err
		}
		clone.Append(v)
	}

	return clone, nil
}

// Filter creates a new list with all elements that pass the test implemented by the provided function.
//
// @param li (*values.List): The list to filter.
// @param lambda (values.Func): The function to test each element. It receives one element and should return
//
//	a value convertible to a boolean.
//
// @return (*values.List): A new list with the elements that passed the test.
func Filter(ctx context.Context, li *values.List, lambda *values.Func) (types.Value, error) {
	newList := values.NewList()
	size := li.Size()

	for i := values.Int(0); i < size; i++ {
		element := li.At(i)
		resultVal, err := lambda.Call(ctx, values.InternalCallSite, []types.Value{element})
		if err != nil {
			return nil, err
		}

		shouldKeep, err := values.AsBool(resultVal)
		if err != nil {
			return nil, fmt.Errorf("filter predicate must return a boolean, but got %s", resultVal.Type())
		}

		if shouldKeep {
			newList.Append(element)
		}
	}

	return newList, nil
}

// RemoveAt removes an element at a specific index from a list.
//
// @param li (*values.List): The list to modify.
// @param idx (values.Int): The index of the element to remove.
// @return (*values.List): The list itself.
func RemoveAt(li *values.List, idx values.Int) (types.Value, error) {
	return li, li.RemoveAt(idx)
}

// RemoveAll removes all occurrences of a given item from the list.
//
// @param li (*values.List): The list to modify.
// @param item (types.Value): The item to remove.
// @return (values.Int): The number of elements removed.
func RemoveAll(li *values.List, item types.Value) (types.Value, error) {
	return li.RemoveAll(item), nil
}

// Slice extracts a section of a list and returns it as a new list.
// The original list is not modified.
//
// @param li (*values.List): The list to slice.
// @param start (values.Int): The zero-based index at which to begin extraction.
// @param end (values.Int): The zero-based index before which to end extraction.
// @return (*values.List): A new list containing the extracted elements.
func Slice(li *values.List, start values.Int, end values.Int) (types.Value, error) {
	size := li.Size()

	if start < 0 || end > size || start > end {
		return nil, fmt.Errorf("index out of bounds for slice operation: start=%d, end=%d, size=%d", start, end, size)
	}

	newList := values.NewList()
	for i := start; i < end; i++ {
		newList.Append(li.At(i))
	}

	return newList, nil
}

// Slice0 is an overload for Slice that extracts from a start index to the end of the list.
//
// @param li (*values.List): The list to slice.
// @param start (values.Int): The zero-based index at which to begin extraction.
// @return (*values.List): A new list containing the extracted elements.
func Slice0(li *values.List, start values.Int) (types.Value, error) {
	return Slice(li, start, li.Size())
}
