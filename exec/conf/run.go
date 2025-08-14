package conf

import (
	"github.com/anhcraft/rice/exec/types"
	"github.com/anhcraft/rice/exec/types/values"
	"time"
)

// RunConfig represents configuration to the Interpreter per each execution
type RunConfig struct {
	UserFuncTimeout   time.Duration
	LexicalScopeLimit uint32
	Variables         map[values.Identifier]types.Value
	Constants         map[values.Identifier]types.Value
}

func NewDefaultRunConfig() *RunConfig {
	return &RunConfig{UserFuncTimeout: 2 * time.Second, LexicalScopeLimit: 1 << 8}
}

func (r *RunConfig) DefineVariable(key values.Identifier, value types.Value) *RunConfig {
	if r.Variables == nil {
		r.Variables = make(map[values.Identifier]types.Value)
	}
	r.Variables[key] = value
	return r
}

func (r *RunConfig) DefineConstant(key values.Identifier, value types.Value) *RunConfig {
	if r.Constants == nil {
		r.Constants = make(map[values.Identifier]types.Value)
	}
	r.Constants[key] = value
	return r
}

func (r *RunConfig) SetUserFuncTimeout(o time.Duration) *RunConfig {
	r.UserFuncTimeout = o
	return r
}

func (r *RunConfig) SetLexicalScopeLimit(lim uint32) *RunConfig {
	r.LexicalScopeLimit = lim
	return r
}
