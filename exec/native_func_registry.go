package exec

import (
	"context"
	"fmt"
	"os"
	"reflect"
	"time"

	"github.com/anhcraft/rice/exec/ctxkey"
	"github.com/anhcraft/rice/exec/fun"
	"github.com/anhcraft/rice/exec/mem"
	_datetime "github.com/anhcraft/rice/exec/stdlib/datetime"
	_error "github.com/anhcraft/rice/exec/stdlib/error"
	_io "github.com/anhcraft/rice/exec/stdlib/io"
	_json "github.com/anhcraft/rice/exec/stdlib/json"
	_list "github.com/anhcraft/rice/exec/stdlib/list"
	_map "github.com/anhcraft/rice/exec/stdlib/map"
	_math "github.com/anhcraft/rice/exec/stdlib/math"
	_set "github.com/anhcraft/rice/exec/stdlib/set"
	_string "github.com/anhcraft/rice/exec/stdlib/string"
	_type "github.com/anhcraft/rice/exec/stdlib/type"
	"github.com/anhcraft/rice/exec/types"
	"github.com/anhcraft/rice/exec/types/values"
)

// NamespacedFunctionPackageList e.g. {string: [], io: []}
type NamespacedFunctionPackageList = map[values.Identifier]fun.FunctionPackage

// TypeboundFunctionPackageList e.g. {string: [], int: [], list: []}
type TypeboundFunctionPackageList = map[types.Type]fun.FunctionPackage

// CompiledFunctionPackage e.g. {substr: "(string, off, [end])", slice: "(list, start, [end])"}
type CompiledFunctionPackage = map[values.Identifier]*fun.ParamTrie
type CompiledNamespacedFunctionPackageList = map[values.Identifier]CompiledFunctionPackage
type CompiledTypeboundFunctionPackageList = map[types.Type]CompiledFunctionPackage

var standardNamespacedPackages = NamespacedFunctionPackageList{
	"":         union(_error.Functions, _io.Functions, _type.Functions),
	"strings":  _string.Functions,
	"math":     _math.Functions,
	"list":     _list.Functions,
	"set":      _set.Functions,
	"map":      _map.Functions,
	"datetime": _datetime.Functions,
	"json":     _json.Functions,
}

var standardTypeboundPackages = TypeboundFunctionPackageList{
	types.String: _string.Functions,
	types.Int:    _math.Functions,
	types.Float:  _math.Functions,
	types.List:   _list.Functions,
	types.Set:    _set.Functions,
	types.Map:    _map.Functions,
}

var printFunctionSignatures = false

func compileNamespacedPkg(packageNamespaces NamespacedFunctionPackageList) CompiledNamespacedFunctionPackageList {
	compiled := make(CompiledNamespacedFunctionPackageList)

	for ns, functionList := range packageNamespaces {
		tries := make(CompiledFunctionPackage)

		for id, function := range functionList {
			trie := fun.NewParamTrie(string(id))
			for _, def := range function {
				err := trie.Register(def)
				if err != nil {
					panic(fmt.Errorf("failed global func init at func %s: %w", id, err))
				}

				if printFunctionSignatures {
					if ns != "" {
						printFunction("%s.", ns)
					}
					printFunction("%s%s\n", id, def.ReadableString())
				}
			}
			tries[id] = trie
		}

		compiled[ns] = tries
	}

	return compiled
}

func compileTypeboundPkg(packageNamespaces TypeboundFunctionPackageList) CompiledTypeboundFunctionPackageList {
	compiled := make(CompiledTypeboundFunctionPackageList)

	for argType, functionList := range packageNamespaces {
		at := fun.NewArgType(0, argType)
		compiled[argType] = make(CompiledFunctionPackage)

		for id, function := range functionList {
			trie := fun.NewParamTrie(string(id))

			for _, def := range function {
				ok := false

				if def.SizeOfArgs() > 0 {
					arg0 := def.Arg(0)

					for _, paramType := range arg0 {
						if paramType.CanAccept(at) || paramType.CanContainMultiOf(at) {
							ok = true
							break
						}
					}
				}

				if !ok {
					continue
				}

				err := trie.Register(def)
				if err != nil {
					panic(fmt.Errorf("failed typebound init at argType %v func %s%s: %w", argType, id, def, err))
				}

				if printFunctionSignatures {
					printFunction("(value of type %s).%s%s\n", argType, id, def.ReadableString())
				}
			}

			compiled[argType][id] = trie
		}
	}

	return compiled
}

func union(sets ...fun.FunctionPackage) fun.FunctionPackage {
	result := make(fun.FunctionPackage)

	for _, s := range sets {
		for k, v := range s {
			result[k] = v
		}
	}

	return result
}

func buildNativeFuncSet(boundValue types.Value, id values.Identifier, pt *fun.ParamTrie, execTimeout time.Duration) values.NativeFunctionSet {
	// NOTE: environment data should be passed from context to decouple the interpreter
	// (for future reusability and optimization)

	return values.NewNativeFunctionSet(
		boundValue,
		func(ctx context.Context, self values.NativeFunctionSet, site values.CallSite, args []types.Value) (types.Value, error) {
			var argValues []reflect.Value

			if boundValue == nil {
				argValues = make([]reflect.Value, len(args))
				for k, arg := range args {
					argValues[k] = reflect.ValueOf(arg)
				}
			} else {
				argValues = make([]reflect.Value, len(args)+1)
				argValues[0] = reflect.ValueOf(boundValue)
				for k, arg := range args {
					argValues[k+1] = reflect.ValueOf(arg)
				}
			}

			lookup, err := pt.MatchHandler(argValues)
			if err != nil {
				return nil, fmt.Errorf("no matching signature for native function %s", id)
			}

			if lookup.Contextual {
				callCtx, _ := context.WithTimeout(ctx, execTimeout)

				argValues = append([]reflect.Value{reflect.ValueOf(callCtx)}, argValues...)
			}

			// Fix up nil args: reflect.ValueOf(untyped nil) produces a zero Value that
			// reflect.Call rejects. Replace zero Values with a typed nil (types.Value interface).
			typeOfValue := reflect.TypeOf((*types.Value)(nil)).Elem()
			for k, arg := range argValues {
				if !arg.IsValid() {
					argValues[k] = reflect.Zero(typeOfValue)
				}
			}

			env := ctx.Value(ctxkey.Env).(*mem.Environment)
			env.PushFrame(site)
			defer env.PopFrame()

			out := lookup.Handler.Call(argValues)

			if !out[1].IsNil() {
				return nil, out[1].Interface().(error)
			}

			if out[0].IsNil() {
				return nil, nil
			} else {
				return out[0].Interface().(types.Value), nil
			}
		},
	)
}

func printFunction(s string, args ...any) {
	f, err := os.OpenFile("function_signatures.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Println("Error opening file:", err)
		return
	}
	defer f.Close()

	if _, err := f.WriteString(fmt.Sprintf(s, args...)); err != nil {
		fmt.Println("Error writing to file:", err)
	}
}
