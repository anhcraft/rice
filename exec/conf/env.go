package conf

import (
	"io"
	"os"
	"rice/exec/fun"
	"rice/exec/types"
	"rice/exec/types/values"
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
