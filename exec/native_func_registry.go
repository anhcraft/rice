package exec

import (
	"context"
	"fmt"
	"os"
	"reflect"
	"time"

	"github.com/anhcraft/rice/exec/conf"
	"github.com/anhcraft/rice/exec/ctxkey"
	"github.com/anhcraft/rice/exec/fun"
	"github.com/anhcraft/rice/exec/mem"
	_datetime "github.com/anhcraft/rice/exec/stdlib/datetime"
	_duration "github.com/anhcraft/rice/exec/stdlib/duration"
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

// StandardNamespacedPackageEntry tracks a stdlib sub-package with its identity
// so that individual sub-packages can be selectively disabled or overridden.
type StandardNamespacedPackageEntry struct {
	PkgID values.Identifier     // unique sub-package identifier, e.g. "io", "error", "type"
	Pkg   fun.FunctionPackage
}

// StandardNamespacedPackageGroups is the decomposed representation of
// standard namespaced packages. Each namespace (key) maps to a list of
// individually-addressable sub-package entries.
var StandardNamespacedPackageGroups = map[values.Identifier][]StandardNamespacedPackageEntry{
	"": {
		{PkgID: "error", Pkg: _error.Functions},
		{PkgID: "io", Pkg: _io.Functions},
		{PkgID: "type", Pkg: _type.Functions},
	},
	"strings":  {{PkgID: "strings", Pkg: _string.Functions}},
	"math":     {{PkgID: "math", Pkg: _math.Functions}},
	"list":     {{PkgID: "list", Pkg: _list.Functions}},
	"set":      {{PkgID: "set", Pkg: _set.Functions}},
	"map":      {{PkgID: "map", Pkg: _map.Functions}},
	"datetime": {{PkgID: "datetime", Pkg: _datetime.Functions}},
	"duration": {{PkgID: "duration", Pkg: _duration.Functions}},
	"json":     {{PkgID: "json", Pkg: _json.Functions}},
}

// buildStandardNamespacedPackages flattens StandardNamespacedPackageGroups into
// a NamespacedFunctionPackageList, optionally excluding packages listed in disabledIDs.
func buildStandardNamespacedPackages(disabledIDs map[values.Identifier]bool) NamespacedFunctionPackageList {
	result := make(NamespacedFunctionPackageList)
	for ns, entries := range StandardNamespacedPackageGroups {
		var toMerge []fun.FunctionPackage
		for _, entry := range entries {
			if disabledIDs != nil && disabledIDs[entry.PkgID] {
				continue
			}
			toMerge = append(toMerge, entry.Pkg)
		}
		if len(toMerge) > 0 {
			result[ns] = union(toMerge...)
		}
	}
	return result
}

// buildStandardNamespacedPackagesWithWhitelist flattens StandardNamespacedPackageGroups
// into a NamespacedFunctionPackageList, including only entries whose PkgID appears
// in the enabledIDs set. If enabledIDs is nil, all entries are included.
func buildStandardNamespacedPackagesWithWhitelist(enabledIDs map[values.Identifier]bool) NamespacedFunctionPackageList {
	result := make(NamespacedFunctionPackageList)
	for ns, entries := range StandardNamespacedPackageGroups {
		var toMerge []fun.FunctionPackage
		for _, entry := range entries {
			if enabledIDs != nil && !enabledIDs[entry.PkgID] {
				continue
			}
			toMerge = append(toMerge, entry.Pkg)
		}
		if len(toMerge) > 0 {
			result[ns] = union(toMerge...)
		}
	}
	return result
}

// buildIdentifierSet converts a slice of identifiers to a lookup set.
func buildIdentifierSet(ids []values.Identifier) map[values.Identifier]bool {
	if len(ids) == 0 {
		return nil
	}
	s := make(map[values.Identifier]bool, len(ids))
	for _, id := range ids {
		s[id] = true
	}
	return s
}

// buildTypeSet converts a slice of types to a lookup set.
func buildTypeSet(ids []types.Type) map[types.Type]bool {
	if len(ids) == 0 {
		return nil
	}
	s := make(map[types.Type]bool, len(ids))
	for _, id := range ids {
		s[id] = true
	}
	return s
}

// buildEffectiveNamespacedPkgs computes the final namespaced package list by
// applying the EnvConfig's disable/strict/override logic on top of the standard
// packages and then merging in any extra packages.
func buildEffectiveNamespacedPkgs(cfg *conf.EnvConfig) NamespacedFunctionPackageList {
	var namespacedPkgs NamespacedFunctionPackageList

	if cfg.StrictStdlibMode {
		namespacedPkgs = buildStandardNamespacedPackagesWithWhitelist(
			buildIdentifierSet(cfg.EnableNamespacedPackages),
		)
	} else {
		namespacedPkgs = buildStandardNamespacedPackages(
			buildIdentifierSet(cfg.DisableNamespacedPackages),
		)
	}

	// Merge extras
	if cfg.ExtraNamespacedFuncPkg != nil && len(cfg.ExtraNamespacedFuncPkg) > 0 {
		if namespacedPkgs == nil {
			namespacedPkgs = make(NamespacedFunctionPackageList)
		}
		overrideSet := buildIdentifierSet(cfg.OverrideNamespacedPackages)
		for k, v := range cfg.ExtraNamespacedFuncPkg {
			existing := namespacedPkgs[k]
			merged := union(v...)
			if existing != nil && len(existing) > 0 {
				// Check for conflicts
				for funcName := range merged {
					if _, exists := existing[funcName]; exists {
						if overrideSet == nil || !overrideSet[k] {
							if cfg.LoggingOutput != nil {
								fmt.Fprintf(cfg.LoggingOutput,
									"[rice] WARNING: custom package overrides standard function %q in namespace %q. "+
										"Add OverrideNamespacedPackage(%q) to suppress this warning.\n",
									funcName, k, k)
							}
						}
					}
				}
			}
			namespacedPkgs[k] = union(existing, merged)
		}
	}

	return namespacedPkgs
}

// buildEffectiveTypeboundPkgs computes the final type-bound package list by
// applying the EnvConfig's disable logic on top of the standard packages
// and then merging in any extra packages.
func buildEffectiveTypeboundPkgs(cfg *conf.EnvConfig) TypeboundFunctionPackageList {
	disabledSet := buildTypeSet(cfg.DisableTypeBoundPackages)

	typeboundPkgs := make(TypeboundFunctionPackageList)
	for t, pkg := range standardTypeboundPackages {
		if disabledSet != nil && disabledSet[t] {
			continue
		}
		typeboundPkgs[t] = pkg
	}

	if cfg.ExtraTypeBoundFuncPkg != nil && len(cfg.ExtraTypeBoundFuncPkg) > 0 {
		for k, v := range cfg.ExtraTypeBoundFuncPkg {
			typeboundPkgs[k] = union(typeboundPkgs[k], union(v...))
		}
	}

	return typeboundPkgs
}

// standardNamespacedPackages is the legacy flat representation, kept for
// backward compatibility. It is built from StandardNamespacedPackageGroups
// with no exclusions.
var standardNamespacedPackages = buildStandardNamespacedPackages(nil)

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
