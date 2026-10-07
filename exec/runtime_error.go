package exec

import (
	"errors"
	"fmt"
	"strings"

	"github.com/anhcraft/rice/exec/ast"
	"github.com/anhcraft/rice/exec/types/values"
)

type RuntimeError struct {
	message string

	// source the caller of the stack frame where the error occurred
	// this is a reliable key to distinct stack frame without keeping references to mem.StackFrame
	source values.CallSite

	// start the starting position of the relevant AST node
	start ast.Pos
	// end the ending position of the relevant AST node
	end ast.Pos

	cause error

	line    int
	column  int
	snippet string
}

// ErrorFrame is one node in a RuntimeError causal chain.
type ErrorFrame struct {
	Message   string
	Caller    string
	Internal  bool
	Start     ast.Pos
	End       ast.Pos
	CallStart ast.Pos
	CallEnd   ast.Pos
	Line      int
	Column    int
	Snippet   string
}

func (re RuntimeError) causedBy(err error) RuntimeError {
	re.cause = err
	return re
}

func (re RuntimeError) Unwrap() error {
	return re.cause
}

func (re RuntimeError) Error() string {
	return re.asFrame().formatError()
}

func (re RuntimeError) asFrame() ErrorFrame {
	return ErrorFrame{
		Message:   re.message,
		Caller:    re.source.Caller,
		Internal:  re.source.Internal,
		Start:     re.start,
		End:       re.end,
		CallStart: re.source.StartPos,
		CallEnd:   re.source.EndPos,
		Line:      re.line,
		Column:    re.column,
		Snippet:   re.snippet,
	}
}

func (f ErrorFrame) formatError() string {
	if f.Snippet != "" {
		return fmt.Sprintf("%s (%d:%d)%s", f.Message, f.Line, f.Column+1, f.Snippet)
	}
	if f.Internal {
		return fmt.Sprintf("%s (internal)", f.Message)
	}
	return f.Message
}

func (f ErrorFrame) sameSite(other ErrorFrame) bool {
	return f.Caller == other.Caller &&
		f.Internal == other.Internal &&
		f.CallStart.Line == other.CallStart.Line &&
		f.CallStart.Index == other.CallStart.Index &&
		f.CallEnd.Line == other.CallEnd.Line &&
		f.CallEnd.Index == other.CallEnd.Index
}

func (f ErrorFrame) siteString() string {
	if f.Internal {
		return fmt.Sprintf("%s (internal)", f.Caller)
	}
	start, end := f.CallStart, f.CallEnd
	if start.Line == 0 && end.Line == 0 {
		start, end = f.Start, f.End
	}
	return fmt.Sprintf("%s (%s-%s)", f.Caller, start.String(), end.String())
}

// Frames walks the causal chain from the outermost frame to the innermost.
func (re RuntimeError) Frames() []ErrorFrame {
	var frames []ErrorFrame
	var cur error = re
	for cur != nil {
		var r RuntimeError
		if errors.As(cur, &r) {
			frames = append(frames, r.asFrame())
			cur = r.cause
			continue
		}
		frames = append(frames, ErrorFrame{Message: cur.Error()})
		break
	}
	return frames
}

// Bind attaches source text and fills line, column, and a caret snippet on this
// error and any RuntimeError causes. The VM does not keep script text.
func (re RuntimeError) Bind(source string) RuntimeError {
	re.line, re.column, re.snippet = bindPos(source, re.start.Index)
	if re.cause != nil {
		var inner RuntimeError
		if errors.As(re.cause, &inner) {
			re.cause = inner.Bind(source)
		}
	}
	return re
}

func (re RuntimeError) Stacktrace() string {
	return buildErrorStacktrace(re)
}

var tabReplacer = strings.NewReplacer("\t", " ")

func bindPos(source string, index int) (line, column int, snippet string) {
	if source == "" {
		return 0, 0, ""
	}
	line = 1
	column = 0
	var runeCount, lineStart int
	for i, r := range source {
		if runeCount == index {
			break
		}
		if r == '\n' {
			lineStart = i + 1
			line++
			column = 0
		} else {
			column++
		}
		runeCount++
	}
	lineEnd := strings.IndexByte(source[lineStart:], '\n')
	if lineEnd < 0 {
		lineEnd = len(source)
	} else {
		lineEnd += lineStart
	}
	if lineStart == lineEnd {
		return line, column, ""
	}

	const prefix = "\n | "
	srcLine := source[lineStart:lineEnd]
	var b strings.Builder
	b.Grow(2*len(prefix) + len(srcLine) + column + 1)
	b.WriteString(prefix)
	_, _ = tabReplacer.WriteString(&b, srcLine)
	b.WriteString(prefix)
	for i := 0; i < column; i++ {
		b.WriteByte('.')
	}
	b.WriteByte('^')
	return line, column, b.String()
}
