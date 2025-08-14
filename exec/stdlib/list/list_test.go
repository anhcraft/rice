package list

import (
	"context"
	"fmt"
	"rice/exec/types"
	"rice/exec/types/values"
	"testing"
)

func newMockFunc(c func(site values.CallSite, args []types.Value) (types.Value, error)) *values.Func {
	return values.NewFunc(
		[]values.Identifier{"a", "b"},
		false,
		nil,
		func(_ context.Context, self *values.Func, site values.CallSite, args []types.Value) (types.Value, error) {
			return c(site, args)
		},
	)
}

func assertListEquals(t *testing.T, li *values.List, expected ...types.Value) {
	t.Helper()
	if li.Size() != values.Int(len(expected)) {
		t.Fatalf("expected list size %d, but got %d", len(expected), li.Size())
	}
	for i, exp := range expected {
		if li.At(values.Int(i)) != exp {
			t.Errorf("expected value at index %d to be %v, but got %v", i, exp, li.At(values.Int(i)))
		}
	}
}

func TestNew(t *testing.T) {
	val, err := New()
	if err != nil {
		t.Fatalf("New() returned an unexpected error: %v", err)
	}
	li, ok := val.(*values.List)
	if !ok {
		t.Fatalf("New() should return *values.List, but got %T", val)
	}
	if li.Size() != 0 {
		t.Errorf("expected new list to have size 0, but got %d", li.Size())
	}
}

func TestPrepend(t *testing.T) {
	li := values.ListOf([]values.String{"c"})
	_, err := Prepend(li, values.String("a"), values.String("b"))
	if err != nil {
		t.Fatalf("Prepend() returned an unexpected error: %v", err)
	}
	assertListEquals(t, li, values.String("a"), values.String("b"), values.String("c"))
}

func TestAppend(t *testing.T) {
	li := values.ListOf([]values.String{"a"})
	_, err := Append(li, values.String("b"), values.String("c"))
	if err != nil {
		t.Fatalf("Append() returned an unexpected error: %v", err)
	}
	assertListEquals(t, li, values.String("a"), values.String("b"), values.String("c"))
}

func TestInclude(t *testing.T) {
	li := values.ListOf([]types.Value{values.String("a"), values.Int(1), values.Bool(true)})
	testCases := []struct {
		name     string
		value    types.Value
		expected values.Bool
	}{
		{"String Exists", values.String("a"), true},
		{"Int Exists", values.Int(1), true},
		{"Value Does Not Exist", values.String("z"), false},
		{"Different Type", values.Float(1.0), false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := Include(li, tc.value)
			if err != nil {
				t.Fatalf("Include() returned an unexpected error: %v", err)
			}
			if result != tc.expected {
				t.Errorf("expected %v, but got %v", tc.expected, result)
			}
		})
	}
}

func TestIndex(t *testing.T) {
	li := values.ListOf([]types.Value{values.String("a"), values.Int(1), values.String("a")})
	testCases := []struct {
		name     string
		value    types.Value
		expected values.Int
	}{
		{"First Occurrence", values.String("a"), 0},
		{"Middle Value", values.Int(1), 1},
		{"Value Does Not Exist", values.String("z"), -1},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := Index(li, tc.value)
			if err != nil {
				t.Fatalf("Index() returned an unexpected error: %v", err)
			}
			if result != tc.expected {
				t.Errorf("expected index %v, but got %v", tc.expected, result)
			}
		})
	}
}

func TestLastIndex(t *testing.T) {
	li := values.ListOf([]types.Value{values.String("a"), values.Int(1), values.String("a")})
	testCases := []struct {
		name     string
		value    types.Value
		expected values.Int
	}{
		{"Last Occurrence", values.String("a"), 2},
		{"Middle Value", values.Int(1), 1},
		{"Value Does Not Exist", values.String("z"), -1},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := LastIndex(li, tc.value)
			if err != nil {
				t.Fatalf("LastIndex() returned an unexpected error: %v", err)
			}
			if result != tc.expected {
				t.Errorf("expected index %v, but got %v", tc.expected, result)
			}
		})
	}
}

func TestSort(t *testing.T) {
	li := values.ListOf([]types.Value{values.Int(3), values.Int(1), values.Int(2)})
	comparator := newMockFunc(func(ctxsite values.CallSite, args []types.Value) (types.Value, error) {
		a := args[0].(values.Int)
		b := args[1].(values.Int)
		return values.Bool(a < b), nil
	})

	_, err := Sort(context.Background(), li, comparator)
	if err != nil {
		t.Fatalf("Sort() returned an unexpected error: %v", err)
	}
	assertListEquals(t, li, values.Int(1), values.Int(2), values.Int(3))

	t.Run("SortErrorInLambda", func(t *testing.T) {
		li := values.ListOf([]types.Value{values.Int(2), values.Int(1)})
		errorComparator := newMockFunc(func(site values.CallSite, args []types.Value) (types.Value, error) {
			return nil, fmt.Errorf("comparison failed")
		})
		_, err := Sort(context.Background(), li, errorComparator)
		if err == nil {
			t.Error("expected an error from Sort when lambda fails, but got nil")
		}
	})
}

func TestReverse(t *testing.T) {
	li := values.ListOf([]values.String{"c", "b", "a"})
	_, err := Reverse(li)
	if err != nil {
		t.Fatalf("Reverse() returned an unexpected error: %v", err)
	}
	assertListEquals(t, li, values.String("a"), values.String("b"), values.String("c"))
}

func TestMap(t *testing.T) {
	li := values.ListOf([]types.Value{values.Int(1), values.Int(2), values.Int(3)})
	mapper := newMockFunc(func(site values.CallSite, args []types.Value) (types.Value, error) {
		val := args[0].(values.Int)
		return val * 2, nil
	})

	result, err := Map(context.Background(), li, mapper)
	if err != nil {
		t.Fatalf("Map() returned an unexpected error: %v", err)
	}

	newList := result.(*values.List)
	assertListEquals(t, newList, values.Int(2), values.Int(4), values.Int(6))
	// Ensure original list is not modified
	assertListEquals(t, li, values.Int(1), values.Int(2), values.Int(3))
}

func TestFilter(t *testing.T) {
	li := values.ListOf([]types.Value{values.Int(1), values.Int(2), values.Int(3), values.Int(4)})
	predicate := newMockFunc(func(site values.CallSite, args []types.Value) (types.Value, error) {
		val := args[0].(values.Int)
		return values.Bool(val%2 == 0), nil
	})

	result, err := Filter(context.Background(), li, predicate)
	if err != nil {
		t.Fatalf("Filter() returned an unexpected error: %v", err)
	}

	newList := result.(*values.List)
	assertListEquals(t, newList, values.Int(2), values.Int(4))
	// Ensure original list is not modified
	assertListEquals(t, li, values.Int(1), values.Int(2), values.Int(3), values.Int(4))
}

func TestRemoveAt(t *testing.T) {
	li := values.ListOf([]types.Value{values.String("a"), values.String("b"), values.String("c")})
	_, err := RemoveAt(li, values.Int(1))
	if err != nil {
		t.Fatalf("RemoveAt() returned an unexpected error: %v", err)
	}
	assertListEquals(t, li, values.String("a"), values.String("c"))

	t.Run("IndexOutOfBounds", func(t *testing.T) {
		_, err := RemoveAt(li, values.Int(99))
		if err == nil {
			t.Error("expected an error for out-of-bounds index, but got nil")
		}
	})
}

func TestRemoveAll(t *testing.T) {
	li := values.ListOf([]types.Value{values.String("a"), values.String("b"), values.String("a"), values.String("c")})
	count, err := RemoveAll(li, values.String("a"))
	if err != nil {
		t.Fatalf("RemoveAll() returned an unexpected error: %v", err)
	}
	if count != values.Int(2) {
		t.Errorf("expected to remove 2 items, but got %d", count)
	}
	assertListEquals(t, li, values.String("b"), values.String("c"))
}

func TestSlice(t *testing.T) {
	li := values.ListOf([]types.Value{values.Int(10), values.Int(20), values.Int(30), values.Int(40), values.Int(50)})

	t.Run("ValidSlice", func(t *testing.T) {
		result, err := Slice(li, 1, 4)
		if err != nil {
			t.Fatalf("Slice() returned an unexpected error: %v", err)
		}
		newList := result.(*values.List)
		assertListEquals(t, newList, values.Int(20), values.Int(30), values.Int(40))
		// Ensure original list is not modified
		assertListEquals(t, li, values.Int(10), values.Int(20), values.Int(30), values.Int(40), values.Int(50))
	})

	t.Run("InvalidSlice", func(t *testing.T) {
		_, err := Slice(li, 3, 2)
		if err == nil {
			t.Error("expected an error for invalid slice (start > end), but got nil")
		}
	})
}

func TestSlice0(t *testing.T) {
	li := values.ListOf([]types.Value{values.Int(10), values.Int(20), values.Int(30), values.Int(40), values.Int(50)})

	t.Run("ValidSlice", func(t *testing.T) {
		result, err := Slice0(li, 2)
		if err != nil {
			t.Fatalf("Slice0() returned an unexpected error: %v", err)
		}
		newList := result.(*values.List)
		assertListEquals(t, newList, values.Int(30), values.Int(40), values.Int(50))
	})

	t.Run("InvalidSlice", func(t *testing.T) {
		_, err := Slice0(li, -1)
		if err == nil {
			t.Error("expected an error for invalid slice (start < 0), but got nil")
		}
	})
}
