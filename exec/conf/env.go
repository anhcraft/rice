package conf

import (
	"github.com/anhcraft/rice/exec/fun"
	"github.com/anhcraft/rice/exec/types"
	"github.com/anhcraft/rice/exec/types/values"
	"io"
	"os"
	"time"
)

// EnvConfig represents the global configuration across executions of the same Interpreter
// these settings could not be made specific in RunConfig:
// - ExtraNamespacedFuncPkg, ExtraTypeBoundFuncPkg: they gets compiled on initialization of Interpreter
// - NativeFuncTimeout: they are hardcoded in compiled type-bound functions
// - ProfilerEnabled: the profiler is init once
// - LoggingOutput: used in init process
type EnvConfig struct {
	ExtraNamespacedFuncPkg map[values.Identifier][]fun.FunctionPackage
	ExtraTypeBoundFuncPkg  map[types.Type][]fun.FunctionPackage
	NativeFuncTimeout      time.Duration
	PreAllocatedFrames     uint16
	ProfilerEnabled        bool
	LoggingOutput          io.Writer

	// DisableNamespacedPackages lists sub-package identifiers whose standard
	// packages should be excluded from the interpreter environment.
	// e.g. {"io"} disables the global io functions (print, printf, etc.).
	DisableNamespacedPackages []values.Identifier

	// DisableTypeBoundPackages lists types whose standard type-bound
	// packages should be excluded. e.g. {types.String} disables string methods.
	DisableTypeBoundPackages []types.Type

	// OverrideNamespacedPackages lists sub-package identifiers for which
	// custom packages may silently replace standard packages of the same ID.
	// Without this, merging a custom package over an existing stdlib sub-package
	// logs a conflict warning.
	OverrideNamespacedPackages []values.Identifier

	// StrictStdlibMode, when true, disables all standard library packages.
	// Only packages explicitly listed in EnableNamespacedPackages and those
	// added via AddNamespacedFunctionPackage/AddGlobalFunctionPackage are available.
	StrictStdlibMode bool

	// EnableNamespacedPackages lists the stdlib sub-package identifiers to
	// enable when StrictStdlibMode is true. Ignored when StrictStdlibMode is false.
	EnableNamespacedPackages []values.Identifier
}

func NewDefaultEnvConfig() *EnvConfig {
	return &EnvConfig{LoggingOutput: os.Stdout, NativeFuncTimeout: time.Second, PreAllocatedFrames: 8}
}

func (e *EnvConfig) AddNamespacedFunctionPackage(ns values.Identifier, pkg fun.FunctionPackage) *EnvConfig {
	if e.ExtraNamespacedFuncPkg == nil {
		e.ExtraNamespacedFuncPkg = make(map[values.Identifier][]fun.FunctionPackage)
	}
	list, ok := e.ExtraNamespacedFuncPkg[ns]
	if !ok {
		list = make([]fun.FunctionPackage, 0)
		e.ExtraNamespacedFuncPkg[ns] = list
	}
	e.ExtraNamespacedFuncPkg[ns] = append(list, pkg)
	return e
}

func (e *EnvConfig) AddGlobalFunctionPackage(pkg fun.FunctionPackage) *EnvConfig {
	return e.AddNamespacedFunctionPackage("", pkg)
}

func (e *EnvConfig) AddTypeBoundFunctionPackage(t types.Type, pkg fun.FunctionPackage) *EnvConfig {
	if e.ExtraTypeBoundFuncPkg == nil {
		e.ExtraTypeBoundFuncPkg = make(map[types.Type][]fun.FunctionPackage)
	}
	list, ok := e.ExtraTypeBoundFuncPkg[t]
	if !ok {
		list = make([]fun.FunctionPackage, 0)
		e.ExtraTypeBoundFuncPkg[t] = list
	}
	e.ExtraTypeBoundFuncPkg[t] = append(list, pkg)
	return e
}

func (e *EnvConfig) SetNativeFuncTimeout(o time.Duration) *EnvConfig {
	e.NativeFuncTimeout = o
	return e
}

func (e *EnvConfig) SetPreAllocatedFrames(n uint16) *EnvConfig {
	e.PreAllocatedFrames = n
	return e
}

func (e *EnvConfig) EnableProfiling() *EnvConfig {
	e.ProfilerEnabled = true
	return e
}

func (e *EnvConfig) SetLoggingOutput(o io.Writer) *EnvConfig {
	e.LoggingOutput = o
	return e
}

// DisableNamespacedPackage adds a sub-package identifier to the exclusion list.
// Standard functions belonging to this sub-package will not be loaded.
// e.g. DisableNamespacedPackage("io") removes print, printf, println, printlnf.
func (e *EnvConfig) DisableNamespacedPackage(pkgID values.Identifier) *EnvConfig {
	e.DisableNamespacedPackages = append(e.DisableNamespacedPackages, pkgID)
	return e
}

// DisableTypeBoundPackage adds a type to the type-bound exclusion list.
// Standard type-bound functions for this type will not be loaded.
func (e *EnvConfig) DisableTypeBoundPackage(t types.Type) *EnvConfig {
	e.DisableTypeBoundPackages = append(e.DisableTypeBoundPackages, t)
	return e
}

// OverrideNamespacedPackage marks a sub-package so that custom functions
// may silently replace its standard definitions without a conflict warning.
func (e *EnvConfig) OverrideNamespacedPackage(pkgID values.Identifier) *EnvConfig {
	e.OverrideNamespacedPackages = append(e.OverrideNamespacedPackages, pkgID)
	return e
}

// SetStrictStdlibMode enables or disables strict stdlib mode.
// When true, only packages listed via EnableNamespacedPackage are available.
func (e *EnvConfig) SetStrictStdlibMode(v bool) *EnvConfig {
	e.StrictStdlibMode = v
	return e
}

// EnableNamespacedPackage adds a sub-package identifier to the whitelist
// used when StrictStdlibMode is true.
func (e *EnvConfig) EnableNamespacedPackage(pkgID values.Identifier) *EnvConfig {
	e.EnableNamespacedPackages = append(e.EnableNamespacedPackages, pkgID)
	return e
}
