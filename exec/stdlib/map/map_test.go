package _map

import (
	"context"
	"fmt"
	"github.com/anhcraft/rice/exec/types"
	"github.com/anhcraft/rice/exec/types/values"
	"reflect"
	"sort"
	"strings"
	"testing"
)

func newTestMap(kvs ...types.Value) *values.Map {
	m := values.NewMap()
	for i := 0; i < len(kvs); i += 2 {
		m.Put(kvs[i], kvs[i+1])
	}
	return m
}

func valueToComparable(v types.Value) string {
	if v == nil {
		return "<nil>"
	}
	switch val := v.(type) {
	case *values.List:
		var parts []string
		for item := range val.Iterate() {
			parts = append(parts, valueToComparable(item))
		}
		return "[" + strings.Join(parts, ", ") + "]"
	case values.String:
		return fmt.Sprintf("%q", string(val))
	case values.Int:
		return fmt.Sprintf("%d", val)
	case values.Bool:
		return fmt.Sprintf("%t", bool(val))
	case values.Float:
		return fmt.Sprintf("%f", float64(val))
	default:
		return fmt.Sprintf("%v", val)
	}
}

func assertUnorderedListEquals(t *testing.T, got, want *values.List) {
	t.Helper()
	if got.Size() != want.Size() {
		t.Fatalf("list size mismatch: got %d, want %d", got.Size(), want.Size())
	}

	wantMap := make(map[string]int)
	for item := range want.Iterate() {
		key := valueToComparable(item)
		wantMap[key]++
	}

	gotMap := make(map[string]int)
	for item := range got.Iterate() {
		key := valueToComparable(item)
		gotMap[key]++
	}

	if !reflect.DeepEqual(wantMap, gotMap) {
		var gotStrings []string
		for k, v := range gotMap {
			gotStrings = append(gotStrings, fmt.Sprintf("%q (count: %d)", k, v))
		}
		var wantStrings []string
		for k, v := range wantMap {
			wantStrings = append(wantStrings, fmt.Sprintf("%q (count: %d)", k, v))
		}
		sort.Strings(gotStrings)
		sort.Strings(wantStrings)

		t.Errorf("lists do not contain the same elements.\n\ngot:\n\t%s\n\nwant:\n\t%s",
			strings.Join(gotStrings, "\n\t"),
			strings.Join(wantStrings, "\n\t"))
	}
}

func TestNew(t *testing.T) {
	result, err := New()
	if err != nil {
		t.Fatalf("New() returned an error: %v", err)
	}

	m, ok := result.(*values.Map)
	if !ok {
		t.Fatalf("New() did not return a *values.Map, got %T", result)
	}

	if m.Size() != 0 {
		t.Errorf("Expected new map to have size 0, got %d", m.Size())
	}
}

func TestPut(t *testing.T) {
	t.Run("Put single item", func(t *testing.T) {
		m := values.NewMap()
		key := values.String("hello")
		val := values.Int(123)

		result, err := Put(m, key, val)
		if err != nil {
			t.Fatalf("Put() returned an error: %v", err)
		}

		if result != m {
			t.Fatal("Put() should return the same map instance")
		}

		if m.Size() != 1 {
			t.Fatalf("Expected map size to be 1, got %d", m.Size())
		}

		gotVal, found := m.Get(key)
		if !found {
			t.Fatal("Key 'hello' not found after Put")
		}
		if !reflect.DeepEqual(gotVal, val) {
			t.Errorf("Got value %v, want %v", gotVal, val)
		}
	})

	t.Run("Put multiple items", func(t *testing.T) {
		m := values.NewMap()
		k1, v1 := values.String("a"), values.Int(1)
		k2, v2 := values.String("b"), values.Int(2)

		_, err := Put(m, k1, v1, k2, v2)
		if err != nil {
			t.Fatalf("Put() returned an error: %v", err)
		}

		if m.Size() != 2 {
			t.Fatalf("Expected map size to be 2, got %d", m.Size())
		}
		gotV1, _ := m.Get(k1)
		gotV2, _ := m.Get(k2)
		if !reflect.DeepEqual(gotV1, v1) || !reflect.DeepEqual(gotV2, v2) {
			t.Error("Values not set correctly for multiple put")
		}
	})

	t.Run("Put uneven arguments", func(t *testing.T) {
		m := values.NewMap()
		_, err := Put(m, values.String("a"))
		if err == nil {
			t.Fatal("Expected an error for uneven arguments, but got nil")
		}
	})
}

func TestRemove(t *testing.T) {
	k1, v1 := values.String("a"), values.Int(1)
	k2, v2 := values.String("b"), values.Int(2)
	k3, v3 := values.String("c"), values.Int(3)
	m := newTestMap(k1, v1, k2, v2, k3, v3)

	result, err := Remove(m, k1, k3)
	if err != nil {
		t.Fatalf("Remove() returned an error: %v", err)
	}

	if result != m {
		t.Fatal("Remove() should return the same map instance")
	}

	if m.Size() != 1 {
		t.Fatalf("Expected map size to be 1 after removing 2 keys, got %d", m.Size())
	}

	_, foundA := m.Get(k1)
	_, foundC := m.Get(k3)
	valB, foundB := m.Get(k2)

	if foundA || foundC {
		t.Error("Removed keys should not be found")
	}
	if !bool(foundB) || !reflect.DeepEqual(valB, v2) {
		t.Error("Key 'b' should not have been removed")
	}
}

func TestInclude(t *testing.T) {
	key := values.String("a")
	val := values.Int(1)
	m := newTestMap(key, val)

	t.Run("Include existing key", func(t *testing.T) {
		result, err := IncludeKey(m, key)
		if err != nil {
			t.Fatalf("Include() returned an error: %v", err)
		}
		if b, ok := result.(values.Bool); !ok || !bool(b) {
			t.Errorf("Expected true for existing key, got %v", result)
		}
	})

	t.Run("Include non-existing key", func(t *testing.T) {
		result, err := IncludeKey(m, values.String("non-existent"))
		if err != nil {
			t.Fatalf("Include() returned an error: %v", err)
		}
		if b, ok := result.(values.Bool); !ok || bool(b) {
			t.Errorf("Expected false for non-existing key, got %v", result)
		}
	})
}

func TestKeys(t *testing.T) {
	k1, v1 := values.String("a"), values.Int(1)
	k2, v2 := values.String("b"), values.Int(2)
	m := newTestMap(k1, v1, k2, v2)

	result, err := Keys(m)
	if err != nil {
		t.Fatalf("Keys() returned an error: %v", err)
	}

	list, ok := result.(*values.Set)
	if !ok {
		t.Fatalf("Keys() did not return a *values.Set, got %T", result)
	}

	want := values.ListOf([]types.Value{k1, k2})
	assertUnorderedListEquals(t, list.AsList(), want)
}

func TestValues(t *testing.T) {
	k1, v1 := values.String("a"), values.Int(1)
	k2, v2 := values.String("b"), values.Int(2)
	m := newTestMap(k1, v1, k2, v2)

	result, err := Values(m)
	if err != nil {
		t.Fatalf("Values() returned an error: %v", err)
	}

	list, ok := result.(*values.List)
	if !ok {
		t.Fatalf("Values() did not return a *values.List, got %T", result)
	}

	want := values.ListOf([]types.Value{v1, v2})
	assertUnorderedListEquals(t, list, want)
}

func TestEntries(t *testing.T) {
	k1, v1 := values.String("a"), values.Int(1)
	k2, v2 := values.String("b"), values.Int(2)
	m := newTestMap(k1, v1, k2, v2)

	result, err := Entries(m)
	if err != nil {
		t.Fatalf("Entries() returned an error: %v", err)
	}

	list, ok := result.(*values.List)
	if !ok {
		t.Fatalf("Entries() did not return a *values.List, got %T", result)
	}

	e1 := values.ListOf([]types.Value{k1, v1})
	e2 := values.ListOf([]types.Value{k2, v2})
	want := values.ListOf([]types.Value{e1, e2})

	assertUnorderedListEquals(t, list, want)
}

func TestMap(t *testing.T) {
	k1, v1 := values.String("a"), values.Int(1)
	k2, v2 := values.String("b"), values.Int(2)
	m := newTestMap(k1, v1, k2, v2)

	mapDelegate := func(_ context.Context, self *values.Func, site values.CallSite, args []types.Value) (types.Value, error) {
		entry := args[0].(*values.List)
		key := entry.At(0)
		val := entry.At(1).(values.Int)
		newVal := val * 2
		return values.ListOf([]types.Value{key, newVal}), nil
	}
	lambda := values.NewFunc([]values.Identifier{"entry"}, false, nil, mapDelegate)

	result, err := Map(context.Background(), m, lambda)
	if err != nil {
		t.Fatalf("Map() returned an error: %v", err)
	}

	newMap, ok := result.(*values.Map)
	if !ok {
		t.Fatalf("Map() did not return a *values.Map, got %T", result)
	}

	if newMap.Size() != 2 {
		t.Fatalf("Expected new map size to be 2, got %d", newMap.Size())
	}

	gotV1, _ := newMap.Get(k1)
	gotV2, _ := newMap.Get(k2)
	wantV1 := values.Int(2)
	wantV2 := values.Int(4)

	if !reflect.DeepEqual(gotV1, wantV1) {
		t.Errorf("Value for key 'a' is wrong. got %v, want %v", gotV1, wantV1)
	}
	if !reflect.DeepEqual(gotV2, wantV2) {
		t.Errorf("Value for key 'b' is wrong. got %v, want %v", gotV2, wantV2)
	}
}

func TestFilter(t *testing.T) {
	k1, v1 := values.String("a"), values.Int(1)
	k2, v2 := values.String("b"), values.Int(2)
	k3, v3 := values.String("c"), values.Int(3)
	m := newTestMap(k1, v1, k2, v2, k3, v3)

	filterDelegate := func(_ context.Context, self *values.Func, site values.CallSite, args []types.Value) (types.Value, error) {
		entry := args[0].(*values.List)
		val := entry.At(1).(values.Int)
		return values.Bool(val%2 == 0), nil
	}
	lambda := values.NewFunc([]values.Identifier{"entry"}, false, nil, filterDelegate)

	result, err := Filter(context.Background(), m, lambda)
	if err != nil {
		t.Fatalf("Filter() returned an error: %v", err)
	}

	newMap, ok := result.(*values.Map)
	if !ok {
		t.Fatalf("Filter() did not return a *values.Map, got %T", result)
	}

	if newMap.Size() != 1 {
		t.Fatalf("Expected new map size to be 1, got %d", newMap.Size())
	}

	_, foundA := newMap.Get(k1)
	valB, foundB := newMap.Get(k2)
	_, foundC := newMap.Get(k3)

	if foundA || foundC {
		t.Error("Keys 'a' and 'c' should have been filtered out")
	}
	if !bool(foundB) || !reflect.DeepEqual(valB, v2) {
		t.Errorf("Key 'b' was not found or has wrong value. Found: %v, Value: %v", foundB, valB)
	}
}
