package exec

import (
	"github.com/anhcraft/rice/exec/types"
)

type Signal interface {
	signal()
}

type BreakSignal struct {
	RuntimeError
}

func (s BreakSignal) signal() {}

func (s BreakSignal) Error() string {
	return "BreakSignal"
}

type ContinueSignal struct {
	RuntimeError
}

func (s ContinueSignal) signal() {}

func (s ContinueSignal) Error() string {
	return "ContinueSignal"
}

type ReturnSignal struct {
	RuntimeError
	Result types.Value
}

func (s ReturnSignal) signal() {}

func (s ReturnSignal) Error() string {
	return "ReturnSignal"
}
