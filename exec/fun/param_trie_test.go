package fun

import (
	"context"
	"fmt"
	"reflect"
	"rice/exec/types"
	"rice/exec/types/values"
	"strings"
	"testing"
)

// --- Test Functions ---

func sumInts(s []values.Int) values.Int {
	var sum values.Int
	for _, v := range s {
		sum += v
	}
	return sum
}

func multiplyAndSumInts(m values.Int, s []values.Int) values.Int {
	return m * sumInts(s)
}

func identityAny(v any) any               { return v }
func identityInt(v values.Int) values.Int { return v }

func processIntOrFloat(v any) values.String {
	switch v.(type) {
	case values.Int:
		return "processed_int"
	case values.Float:
		return "processed_float"
	}
	return "unknown"
}

func processNested(arr [][]values.Int) values.Int {
	var sum values.Int
	for _, sub := range arr {
		sum += sumInts(sub)
	}
	return sum
}

func variadicSum(nums ...values.Int) values.Int { return sumInts(nums) }
func noArgs() values.String                     { return "no_args_executed" }
func oneString(s values.String) values.String   { return s + "_processed" }

func variadicAnySum(args ...any) values.Int {
	var sum values.Int
	for _, arg := range args {
		if v, ok := arg.(values.Int); ok {
			sum += v
		}
	}
	return sum
}

func variadicStringAndInts(s values.String, nums ...values.Int) values.String {
	sum := sumInts(nums)
	return values.String(fmt.Sprintf("%s:%d", s, sum))
}

func sliceOfAny(v []any) values.Int {
	return values.Int(len(v))
}

func contextualNoArgs(ctx context.Context) values.String {
	return "no_args_executed"
}

func contextualOneString(ctx context.Context, s values.String) values.String {
	return s + "_processed"
}

func contextualVariadicSum(ctx context.Context, nums ...values.Int) values.Int {
	return sumInts(nums)
}

// --- Test Main ---

func TestParamTrie(t *testing.T) {
	// --- Test Helper Setup ---
	type testRegistry struct {
		*ParamTrie
		t *testing.T
	}

	newTestRegistry := func(t *testing.T) *testRegistry {
		return &testRegistry{
			ParamTrie: NewParamTrie(""),
			t:         t,
		}
	}

	register := func(r *testRegistry, id string, fn any) *FunctionDef {
		r.t.Helper()
		def, err := ScanFunction(fn)
		if err != nil {
			r.t.Fatalf("Failed to scan function %s: %v", id, err)
		}
		r.id = id
		err = r.Register(def)
		if err != nil {
			r.t.Fatalf("Failed to register function %s: %v", id, err)
		}
		return def
	}

	registerAndExpectError := func(r *testRegistry, id string, fn any, expectedErr string) {
		r.t.Helper()
		def, err := ScanFunction(fn)
		if err != nil {
			r.t.Fatalf("ScanFunction unexpectedly failed for %s: %v", id, err)
		}
		r.id = id
		err = r.Register(def)
		if err == nil {
			r.t.Errorf("Expected error when registering %s, but got none", id)
			return
		}
		if !strings.Contains(err.Error(), expectedErr) {
			r.t.Errorf("Expected registration error for %s to contain %q, but got %q", id, expectedErr, err.Error())
		}
	}

	call := func(r *testRegistry, id string, expectedResult any, args ...any) {
		r.t.Helper()
		argValues := make([]reflect.Value, len(args))
		for i, arg := range args {
			argValues[i] = reflect.ValueOf(arg)
		}

		res, err := r.MatchHandler(argValues)
		if err != nil {
			r.t.Errorf("GetHandler for %s with args (%T) failed: %v", id, args, err)
			return
		}
		if !res.Handler.IsValid() {
			r.t.Errorf("GetHandler for %s returned an invalid handler", id)
			return
		}

		if res.Contextual {
			argValues = append([]reflect.Value{reflect.ValueOf(context.TODO())}, argValues...)
		}

		results := res.Handler.Call(argValues)

		if len(results) == 0 {
			if expectedResult != nil {
				r.t.Errorf("Call to %s had no return value, expected %v", id, expectedResult)
			}
			return
		}

		result := results[0].Interface()
		if !reflect.DeepEqual(result, expectedResult) {
			r.t.Errorf("Call to %s produced %v, expected %v", id, result, expectedResult)
		}
	}

	callAndExpectError := func(r *testRegistry, id string, expectedErr string, args ...any) {
		r.t.Helper()
		r.id = id
		argValues := make([]reflect.Value, len(args))
		for i, arg := range args {
			if arg == nil {
				argValues[i] = reflect.ValueOf(nil)
			} else {
				argValues[i] = reflect.ValueOf(arg)
			}
		}

		_, err := r.MatchHandler(argValues)
		if err == nil {
			r.t.Errorf("Expected error for %s with args (%v), but got none", id, args)
			return
		}
		if !strings.Contains(err.Error(), expectedErr) {
			r.t.Errorf("Expected error for %s to contain %q, but got %q", id, expectedErr, err.Error())
		}
	}

	// --- Test Cases ---

	t.Run("Basic Registration and Calling", func(t *testing.T) {
		r := newTestRegistry(t)
		register(r, "math.sum", sumInts)
		register(r, "util.no_args", noArgs)
		register(r, "util.nested", processNested)

		call(r, "math.sum", values.Int(6), []values.Int{1, 2, 3})
		call(r, "util.no_args", values.String("no_args_executed"))
		call(r, "util.nested", values.Int(15), [][]values.Int{{1, 2, 3}, {4, 5}})
	})

	t.Run("Contextual Basic Registration and Calling", func(t *testing.T) {
		r := newTestRegistry(t)
		register(r, "util.one", contextualOneString)
		register(r, "util.no_args", contextualNoArgs)

		call(r, "util.one", values.String("one_arg_processed"), values.String("one_arg"))
		call(r, "util.no_args", values.String("no_args_executed"))
	})

	t.Run("Overload Resolution", func(t *testing.T) {
		r := newTestRegistry(t)
		register(r, "util.id", identityInt)
		register(r, "util.id", identityAny)

		t.Log("Testing specificity: util.id(Int) should pick identityInt over identityAny")
		call(r, "util.id", values.Int(42), values.Int(42))

		t.Log("Testing fallback: util.id(Float) should pick identityAny")
		call(r, "util.id", values.Float(3.14), values.Float(3.14))
	})

	t.Run("Custom Defined Arguments", func(t *testing.T) {
		r := newTestRegistry(t)
		r.id = "util.special"
		def, _ := ScanFunction(processIntOrFloat)
		def.DefineArg(0, NewArgType(0, types.Int), NewArgType(0, types.Float))
		if err := r.Register(def); err != nil {
			t.Fatal(err)
		}

		call(r, "util.special", values.String("processed_int"), values.Int(100))
		call(r, "util.special", values.String("processed_float"), values.Float(200.5))
		callAndExpectError(r, "util.special", "no matching signature", values.String("abc"))
	})

	t.Run("Variadic Function Resolution", func(t *testing.T) {
		t.Run("Specific Type Variadic", func(t *testing.T) {
			r := newTestRegistry(t)
			register(r, "vsum", variadicSum)

			call(r, "vsum", values.Int(0))
			call(r, "vsum", values.Int(10), values.Int(10))
			call(r, "vsum", values.Int(15), values.Int(1), values.Int(2), values.Int(12))
			callAndExpectError(r, "vsum", "no matching signature", values.Int(1), values.Float(2.0))
		})

		t.Run("Contextual Specific Type Variadic", func(t *testing.T) {
			r := newTestRegistry(t)
			register(r, "vsum", contextualVariadicSum)

			call(r, "vsum", values.Int(0))
			call(r, "vsum", values.Int(10), values.Int(10))
			call(r, "vsum", values.Int(15), values.Int(1), values.Int(2), values.Int(12))
			callAndExpectError(r, "vsum", "no matching signature", values.Int(1), values.Float(2.0))
		})

		t.Run("Any Type Variadic", func(t *testing.T) {
			r := newTestRegistry(t)
			register(r, "vany", variadicAnySum)

			call(r, "vany", values.Int(0))
			call(r, "vany", values.Int(5), values.Int(5))
			call(r, "vany", values.Int(15), values.Int(5), values.String("hello"), values.Bool(true), values.Int(10))
		})

		t.Run("Overload Precedence with Variadic", func(t *testing.T) {
			r := newTestRegistry(t)
			register(r, "multi", variadicStringAndInts)
			register(r, "multi", variadicAnySum)

			call(r, "multi", values.String("test:15"), values.String("test"), values.Int(5), values.Int(10))
			call(r, "multi", values.String("zero:0"), values.String("zero"))
			call(r, "multi", values.Int(25), values.Int(10), values.Int(15))
		})
	})

	t.Run("Registration Conflict Errors", func(t *testing.T) {
		t.Run("Duplicate Signature", func(t *testing.T) {
			r := newTestRegistry(t)
			register(r, "dup", oneString)
			registerAndExpectError(r, "dup", oneString, "has duplicated handler")
		})

		t.Run("Zero Variadic Conflict", func(t *testing.T) {
			r := newTestRegistry(t)
			register(r, "conflict", noArgs)
			registerAndExpectError(r, "conflict", variadicSum, "has duplicated handler (zero variadic)")
		})

		t.Run("Slice vs Variadic Conflict", func(t *testing.T) {
			r := newTestRegistry(t)
			register(r, "slice", sliceOfAny)
			registerAndExpectError(r, "slice", variadicAnySum, "has duplicated handler")
		})
	})

	t.Run("Lookup Error Handling", func(t *testing.T) {
		r := newTestRegistry(t)
		register(r, "math.sum", multiplyAndSumInts)

		t.Run("Function Not Found", func(t *testing.T) {
			callAndExpectError(r, "nonexistent.func", "no matching signature for function \"nonexistent.func\" with given arguments")
		})

		t.Run("Signature Not Found", func(t *testing.T) {
			callAndExpectError(r, "math.sum", "no matching signature", values.Bool(true))
			callAndExpectError(r, "math.sum", "no matching signature", values.Int(1))
		})
	})
}
