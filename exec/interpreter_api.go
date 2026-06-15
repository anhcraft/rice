package exec

import (
	"context"
	"io"
	"iter"
	"sync"
	"time"

	"github.com/anhcraft/rice/exec/ast"
	"github.com/anhcraft/rice/exec/conf"
	"github.com/anhcraft/rice/exec/ctxkey"
	"github.com/anhcraft/rice/exec/mem"
	"github.com/anhcraft/rice/exec/profiler"
	"github.com/anhcraft/rice/exec/types"
	"github.com/anhcraft/rice/exec/types/values"
)

var _ ast.Visitor = (*Interpreter)(nil)

// Interpreter a single-threaded interpreter
type Interpreter struct {
	// Global
	// -config
	profiler          profiler.Profiler
	loggingOutput     io.Writer
	nativeFuncTimeout time.Duration // for type-bound functions; in sync with compiled namespaced functions
	typeBoundFuncPkg  CompiledTypeboundFunctionPackageList

	// -state
	lock      sync.Mutex
	sessionId uint16
	env       *mem.Environment

	// Per-execution
	// -config
	ctx               context.Context
	userFuncTimeout   time.Duration
	lexicalScopeLimit uint32

	// -state
	functionDepth int
	loopDepth     int
	dirty         bool
}

func NewInterpreter(cfg *conf.EnvConfig) *Interpreter {
	it := &Interpreter{
		loggingOutput:     cfg.LoggingOutput,
		nativeFuncTimeout: cfg.NativeFuncTimeout,
		env:               mem.NewEnvironment(cfg.PreAllocatedFrames),
		typeBoundFuncPkg:  make(CompiledTypeboundFunctionPackageList),
	}

	if cfg.ProfilerEnabled {
		it.profiler = profiler.NewImpl()
	} else {
		it.profiler = profiler.NewMuted()
	}

	{
		namespacedPkgs := standardNamespacedPackages

		if cfg.ExtraNamespacedFuncPkg != nil && len(cfg.ExtraNamespacedFuncPkg) > 0 {
			namespacedPkgs = make(NamespacedFunctionPackageList)
			for k, v := range standardNamespacedPackages { // clone
				namespacedPkgs[k] = v
			}
			for k, v := range cfg.ExtraNamespacedFuncPkg {
				namespacedPkgs[k] = union(namespacedPkgs[k], union(v...))
			}
		}

		gt := compileNamespacedPkg(namespacedPkgs)

		for ns, tries := range gt {
			layer := make(map[values.Identifier]values.NativeFunctionSet)

			for id, trie := range tries {
				layer[id] = buildNativeFuncSet(nil, id, trie, cfg.NativeFuncTimeout)
			}

			if ns != "" {
				it.env.Namespace().RequirePath(ns).Merge(values.NativeFunctionLayer(layer))
			} else {
				it.env.Namespace().Merge(values.NativeFunctionLayer(layer))
			}
		}
	}

	{
		typeboundPkgs := standardTypeboundPackages

		if cfg.ExtraTypeBoundFuncPkg != nil && len(cfg.ExtraTypeBoundFuncPkg) > 0 {
			typeboundPkgs = make(TypeboundFunctionPackageList)
			for k, v := range standardTypeboundPackages { // clone
				typeboundPkgs[k] = v
			}
			for k, v := range cfg.ExtraTypeBoundFuncPkg {
				typeboundPkgs[k] = union(typeboundPkgs[k], union(v...))
			}
		}

		it.typeBoundFuncPkg = compileTypeboundPkg(typeboundPkgs)
	}

	return it
}

// Interpret entrypoint to evaluating the given script
func (i *Interpreter) Interpret(ctx context.Context, script []ast.Stmt, cfg *conf.RunConfig) (types.Value, error) {
	return i.InterpretStream(ctx, func(yield func(ast.Stmt) bool) {
		for _, v := range script {
			if !yield(v) {
				return
			}
		}
	}, cfg, nil)
}

// InterpretStream entrypoint to evaluating the given script from a stream
// For each statement evaluated, stmtCallback get called; its result determine whether to continue the execution
func (i *Interpreter) InterpretStream(ctx context.Context,
	script iter.Seq[ast.Stmt],
	cfg *conf.RunConfig,
	errorRecover func(val types.Value, err error) bool) (types.Value, error) {
	i.lock.Lock()
	defer i.lock.Unlock()

	if i.dirty {
		i.cleanUp()
	}

	i.sessionId++ // ensure positive
	i.ctx = ctx
	i.ctx = context.WithValue(i.ctx, ctxkey.SessionId, i.sessionId)
	i.ctx = context.WithValue(i.ctx, ctxkey.LoggingOutput, i.loggingOutput)
	i.ctx = context.WithValue(i.ctx, ctxkey.Env, i.env)
	i.userFuncTimeout = cfg.UserFuncTimeout
	i.lexicalScopeLimit = cfg.LexicalScopeLimit

	i.env.PushFrame(values.RootCallSite)
	i.profiler.Start(ast.Root{})
	defer func() {
		i.profiler.End()
		i.env.PopFrame()
		i.dirty = true
	}()

	for id, v := range cfg.Constants {
		i.env.Define(id, v, true)
	}

	for id, v := range cfg.Variables {
		if !i.env.Define(id, v, false) {
			err := i.env.Assign(id, v)
			if err != nil {
				return nil, err
			}
		}
	}

	var val types.Value
	var err error

	idx := 0
	for e := range script {
		val, err = i.eval(e)

		if err != nil {
			err = i.throw(e, "cannot eval script statement #%d", idx+1).causedBy(err)
		}

		if errorRecover != nil {
			if !errorRecover(val, err) {
				break
			}
		} else if err != nil {
			return nil, err
		}

		idx++
	}

	return val, err
}
