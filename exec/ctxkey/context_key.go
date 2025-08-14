//go:generate stringer -type ContextKey
package ctxkey

type ContextKey int

const (
	// SessionId (uint16) a positive monotonically-increasing session id per Interpreter; eventually wraparound
	SessionId ContextKey = iota

	// LoggingOutput (io.Writer) the output to write out
	LoggingOutput

	// Env (*mem.Environment) the environment
	Env
)
