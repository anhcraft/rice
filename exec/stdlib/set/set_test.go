package set

import (
	"context"
	"errors"
	"github.com/anhcraft/rice/exec/types"
	"github.com/anhcraft/rice/exec/types/values"
	"testing"
)

func mockFunc(delegate values.FuncDelegate) *values.Func {
	return values.NewFunc([]values.Identifier{"x"}, false, nil, delegate)
}

func assertSetContents(t *testing.T, st *values.Set, expected ...types.Value) {
	t.Helper()

	if int(st.Size()) != len(expected) {
		t.Errorf("expected set size %d, but got %d", len(expected), st.Size())
	}

	for _, v := range expected {
		if !st.Contain(v) {
			t.Errorf("expected set to contain %v, but it did not", v)
		}
	}
}

func TestNew(t *testing.T) {
	t.Run("should create a new empty set", func(t *testing.T) {
		v, err := New()
		if err != nil {
			t.Fatalf("New() returned an unexpected error: %v", err)
		}

		st, ok := v.(*values.Set)
		if !ok {
			t.Fatalf("New() should return a *values.Set, but got %T", v)
		}

		if st.Size() != 0 {
			t.Errorf("expected new set to be empty, but size was %d", st.Size())
		}
	})
}

func TestAdd(t *testing.T) {
	st := values.NewSet()
	v1 := values.Int(10)
	v2 := values.String("hello")
	v3 := values.Int(10)

	t.Run("add single item", func(t *testing.T) {
		res, err := Add(st, v1)
		if err != nil {
			t.Fatalf("Add() returned an unexpected error: %v", err)
		}
		if res != st {
			t.Errorf("Add() should return the same set instance")
		}
		assertSetContents(t, st, v1)
	})

	t.Run("add multiple items including duplicates", func(t *testing.T) {
		res, err := Add(st, v2, v3)
		if err != nil {
			t.Fatalf("Add() returned an unexpected error: %v", err)
		}
		if res != st {
			t.Errorf("Add() should return the same set instance")
		}
		assertSetContents(t, st, v1, v2)
	})

	t.Run("add no items", func(t *testing.T) {
		initialSize := st.Size()
		_, err := Add(st)
		if err != nil {
			t.Fatalf("Add() with no items returned an unexpected error: %v", err)
		}
		if st.Size() != initialSize {
			t.Errorf("Add() with no items should not change the set size")
		}
	})
}

func TestInclude(t *testing.T) {
	st := values.NewSet()
	v1 := values.Int(42)
	v2 := values.String("world")
	st.Add(v1)
	st.Add(v2)

	testCases := []struct {
		name     string
		value    types.Value
		expected values.Bool
	}{
		{"existing int", v1, true},
		{"existing string", v2, true},
		{"non-existing int", values.Int(99), false},
		{"non-existing string", values.String("goodbye"), false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := Include(st, tc.value)
			if err != nil {
				t.Fatalf("Include() returned an unexpected error: %v", err)
			}

			b, ok := result.(values.Bool)
			if !ok {
				t.Fatalf("Include() should return a values.Bool, got %T", result)
			}

			if b != tc.expected {
				t.Errorf("expected %v, got %v", tc.expected, b)
			}
		})
	}
}

func TestMap(t *testing.T) {
	st := values.NewSet()
	st.Add(values.Int(1))
	st.Add(values.Int(2))
	st.Add(values.Int(3))

	t.Run("successful map operation", func(t *testing.T) {
		doubleFunc := mockFunc(func(_ context.Context, self *values.Func, site values.CallSite, args []types.Value) (types.Value, error) {
			i := args[0].(values.Int)
			return i * 2, nil
		})

		result, err := Map(context.Background(), st, doubleFunc)
		if err != nil {
			t.Fatalf("Map() returned an unexpected error: %v", err)
		}

		newSet, ok := result.(*values.Set)
		if !ok {
			t.Fatalf("Map() should return a *values.Set, got %T", result)
		}

		if newSet == st {
			t.Errorf("Map() should return a new set instance, not modify in-place")
		}

		assertSetContents(t, newSet, values.Int(2), values.Int(4), values.Int(6))
		assertSetContents(t, st, values.Int(1), values.Int(2), values.Int(3)) // Original unchanged
	})

	t.Run("map with lambda error", func(t *testing.T) {
		expectedErr := errors.New("lambda failed")
		errorFunc := mockFunc(func(_ context.Context, self *values.Func, site values.CallSite, args []types.Value) (types.Value, error) {
			return nil, expectedErr
		})

		_, err := Map(context.Background(), st, errorFunc)
		if !errors.Is(err, expectedErr) {
			t.Errorf("Map() did not propagate the error from the lambda. Expected '%v', got '%v'", expectedErr, err)
		}
	})
}

func TestFilter(t *testing.T) {
	st := values.NewSet()
	st.Add(values.Int(1))
	st.Add(values.Int(2))
	st.Add(values.Int(3))
	st.Add(values.Int(4))

	t.Run("successful filter operation", func(t *testing.T) {
		isEvenFunc := mockFunc(func(_ context.Context, self *values.Func, site values.CallSite, args []types.Value) (types.Value, error) {
			i := args[0].(values.Int)
			return values.Bool(i%2 == 0), nil
		})

		result, err := Filter(context.Background(), st, isEvenFunc)
		if err != nil {
			t.Fatalf("Filter() returned an unexpected error: %v", err)
		}

		newSet, ok := result.(*values.Set)
		if !ok {
			t.Fatalf("Filter() should return a *values.Set, got %T", result)
		}

		if newSet == st {
			t.Errorf("Filter() should return a new set instance, not modify in-place")
		}

		assertSetContents(t, newSet, values.Int(2), values.Int(4))
		assertSetContents(t, st, values.Int(1), values.Int(2), values.Int(3), values.Int(4)) // Original unchanged
	})

	t.Run("filter with lambda error", func(t *testing.T) {
		expectedErr := errors.New("lambda failed")
		errorFunc := mockFunc(func(_ context.Context, self *values.Func, site values.CallSite, args []types.Value) (types.Value, error) {
			return nil, expectedErr
		})

		_, err := Filter(context.Background(), st, errorFunc)
		if !errors.Is(err, expectedErr) {
			t.Errorf("Filter() did not propagate the error from the lambda. Expected '%v', got '%v'", expectedErr, err)
		}
	})

	t.Run("filter with non-boolean predicate", func(t *testing.T) {
		nonBoolFunc := mockFunc(func(_ context.Context, self *values.Func, site values.CallSite, args []types.Value) (types.Value, error) {
			return values.String("not a bool"), nil
		})

		_, err := Filter(context.Background(), st, nonBoolFunc)
		expectedErrStr := "filter predicate must return a boolean, but got String"
		if err == nil || err.Error() != expectedErrStr {
			t.Errorf("Filter() did not return the correct error for a non-boolean predicate. Expected '%s', got '%v'", expectedErrStr, err)
		}
	})
}

func TestRemove(t *testing.T) {
	st := values.NewSet()
	v1 := values.Int(100)
	v2 := values.String("item")
	st.Add(v1)
	st.Add(v2)

	t.Run("remove existing item", func(t *testing.T) {
		res, err := Remove(st, v1)
		if err != nil {
			t.Fatalf("Remove() returned an unexpected error: %v", err)
		}
		if res != st {
			t.Errorf("Remove() should return the same set instance")
		}
		assertSetContents(t, st, v2)
	})

	t.Run("remove non-existing item", func(t *testing.T) {
		// v1 was already removed, so it's non-existing now
		_, err := Remove(st, v1)
		if err != nil {
			t.Fatalf("Remove() returned an unexpected error: %v", err)
		}
		assertSetContents(t, st, v2)
	})
}

func TestFreeze(t *testing.T) {
	t.Run("Freeze returns same set", func(t *testing.T) {
		st := values.NewSet()
		st.Add(values.Int(1))
		st.Add(values.Int(2))
		result, err := Freeze(st)
		if err != nil {
			t.Fatalf("Freeze() returned an unexpected error: %v", err)
		}
		if result != st {
			t.Error("Freeze() should return the same set instance")
		}
		if !st.IsFrozen() {
			t.Error("set should be frozen after Freeze()")
		}
	})

	t.Run("Add on frozen set returns error", func(t *testing.T) {
		st := values.NewSet()
		st.Add(values.Int(1))
		Freeze(st)
		_, err := Add(st, values.Int(2))
		if err == nil {
			t.Error("expected frozen error from Add, got nil")
		}
	})

	t.Run("Remove on frozen set returns error", func(t *testing.T) {
		st := values.NewSet()
		st.Add(values.Int(1))
		Freeze(st)
		_, err := Remove(st, values.Int(1))
		if err == nil {
			t.Error("expected frozen error from Remove, got nil")
		}
	})

	t.Run("Non-mutating operations work on frozen set", func(t *testing.T) {
		st := values.NewSet()
		st.Add(values.Int(1))
		st.Add(values.Int(2))
		Freeze(st)

		found, err := Include(st, values.Int(1))
		if err != nil || found != values.Bool(true) {
			t.Errorf("Include() should work on frozen set: err=%v, found=%v", err, found)
		}

		result, err := Map(context.Background(), st, mockFunc(func(_ context.Context, self *values.Func, site values.CallSite, args []types.Value) (types.Value, error) {
			return args[0].(values.Int) * 2, nil
		}))
		if err != nil {
			t.Errorf("Map() should work on frozen set: %v", err)
		}
		_ = result

		result, err = Filter(context.Background(), st, mockFunc(func(_ context.Context, self *values.Func, site values.CallSite, args []types.Value) (types.Value, error) {
			return values.Bool(true), nil
		}))
		if err != nil {
			t.Errorf("Filter() should work on frozen set: %v", err)
		}
	})
}
