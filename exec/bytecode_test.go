package exec

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/anhcraft/rice/exec/compiler"
	"github.com/anhcraft/rice/exec/conf"
	"github.com/anhcraft/rice/exec/types"
	"github.com/anhcraft/rice/exec/types/values"
	"github.com/anhcraft/rice/frontend"
)

func errText(err error) string {
	if err == nil {
		return ""
	}
	var re RuntimeError
	if errors.As(err, &re) {
		return re.Stacktrace() + re.Error()
	}
	return err.Error()
}
func runScript(t *testing.T, script string) (types.Value, error) {
	t.Helper()
	tokens, err := frontend.Tokenize(script)
	if err != nil {
		t.Fatalf("tokenize: %v", err)
	}
	p := frontend.NewParser(tokens)
	tree := p.Parse()
	if len(p.Errors()) > 0 {
		t.Fatalf("parse: %v", p.Errors()[0])
	}
	it := NewInterpreter(conf.NewDefaultEnvConfig())
	return it.Interpret(context.Background(), tree, conf.NewDefaultRunConfig())
}

func TestBytecodeDeliberateChanges(t *testing.T) {
	t.Run("hoisted inner x hides outer until defined", func(t *testing.T) {
		_, err := runScript(t, `
			var x = 10;
			{
				const g = func() { return x; };
				g();
				var x = 2;
			}
		`)
		if err == nil || !strings.Contains(errText(err), "unresolved") {
			t.Fatalf("expected unresolved, got %v", err)
		}
	})

	t.Run("parameter shadows builtin", func(t *testing.T) {
		val, err := runScript(t, `const f = func(print) { return print; }; f(42)`)
		if err != nil {
			t.Fatal(err)
		}
		if val != values.Int(42) {
			t.Fatalf("got %v", val)
		}
	})

	t.Run("break outside loop is compile error", func(t *testing.T) {
		_, err := runScript(t, `break;`)
		var ce compiler.Error
		if !errors.As(err, &ce) {
			t.Fatalf("expected compile error, got %v", err)
		}
	})

	t.Run("continue outside loop is compile error", func(t *testing.T) {
		_, err := runScript(t, `continue;`)
		var ce compiler.Error
		if !errors.As(err, &ce) {
			t.Fatalf("expected compile error, got %v", err)
		}
	})

	t.Run("return at script level is compile error", func(t *testing.T) {
		_, err := runScript(t, `return 1;`)
		var ce compiler.Error
		if !errors.As(err, &ce) {
			t.Fatalf("expected compile error, got %v", err)
		}
	})

	t.Run("integer division by zero", func(t *testing.T) {
		_, err := runScript(t, `1 / 0`)
		if err == nil || !strings.Contains(errText(err), "division by zero") {
			t.Fatalf("expected division by zero, got %v", err)
		}
	})
}

func TestBytecodeOps(t *testing.T) {
	t.Run("short circuit non-bool left", func(t *testing.T) {
		val, err := runScript(t, `0 && true`)
		if err != nil {
			t.Fatal(err)
		}
		if val != values.Bool(false) {
			t.Fatalf("got %v (%T)", val, val)
		}
	})

	t.Run("if without else is false", func(t *testing.T) {
		val, err := runScript(t, `if false { 1 }`)
		if err != nil {
			t.Fatal(err)
		}
		if val != values.Bool(false) {
			t.Fatalf("got %v (%T)", val, val)
		}
	})

	t.Run("increment string fails", func(t *testing.T) {
		_, err := runScript(t, `var s = "a"; s++`)
		if err == nil || !strings.Contains(errText(err), "not numeric") {
			t.Fatalf("expected numeric error, got %v", err)
		}
	})

	t.Run("list map calls vm closure", func(t *testing.T) {
		val, err := runScript(t, `list.of(1,2,3).map(func(x){ x * 2 })[1]`)
		if err != nil {
			t.Fatal(err)
		}
		if val != values.Int(4) {
			t.Fatalf("got %v", val)
		}
	})

	t.Run("scope limit", func(t *testing.T) {
		it := NewInterpreter(conf.NewDefaultEnvConfig())
		tokens, err := frontend.Tokenize(`{ { { { { { { { { { 1 } } } } } } } } } }`)
		if err != nil {
			t.Fatal(err)
		}
		p := frontend.NewParser(tokens)
		tree := p.Parse()
		if len(p.Errors()) > 0 {
			t.Fatal(p.Errors()[0])
		}
		_, err = it.Interpret(context.Background(), tree, conf.NewDefaultRunConfig().SetLexicalScopeLimit(8))
		if err == nil || !strings.Contains(errText(err), "lexical scope limit") {
			t.Fatalf("expected scope limit, got %v", err)
		}
	})
}

func TestInterpretStreamSequential(t *testing.T) {
	val, err := runScript(t, `var n = 1; n + 2`)
	if err != nil {
		t.Fatal(err)
	}
	if val != values.Int(3) {
		t.Fatalf("got %v", val)
	}
}

func TestNestedErrorTrace(t *testing.T) {
	script := `
		const f = func() { missing };
		const g = func() { f() };
		g()
	`
	_, err := runScript(t, script)
	if err == nil {
		t.Fatal("expected error")
	}
	var re RuntimeError
	if !errors.As(err, &re) {
		t.Fatalf("expected RuntimeError, got %T %v", err, err)
	}
	st := re.Stacktrace()
	if !strings.Contains(st, "unresolved") {
		t.Fatalf("trace:\n%s", st)
	}
	frames := re.Frames()
	if len(frames) < 2 {
		t.Fatalf("expected a call chain, got %d frames:\n%s", len(frames), st)
	}
	if re.Unwrap() == nil {
		t.Fatal("expected Unwrap to return the cause")
	}
	var inner RuntimeError
	if !errors.As(re.Unwrap(), &inner) {
		t.Fatalf("expected Unwrap RuntimeError, got %T", re.Unwrap())
	}
	callees := 0
	for _, f := range frames {
		if f.Caller == "func()" {
			callees++
		}
	}
	if callees < 2 {
		t.Fatalf("expected caller frames for f and g, got %d:\n%s", callees, st)
	}

	bound := re.Bind(script)
	bst := bound.Stacktrace()
	if !strings.Contains(bst, "unresolved") || !strings.Contains(bst, "^") {
		t.Fatalf("bound trace:\n%s", bst)
	}
	last := bound.Frames()[len(bound.Frames())-1]
	if last.Snippet == "" || !strings.Contains(last.Snippet, "^") {
		t.Fatalf("expected caret snippet on the root cause: %+v", last)
	}
	if !strings.Contains(bst, "(") || last.Line < 1 {
		t.Fatalf("expected (line:column) on the root cause: %+v\n%s", last, bst)
	}
}

func TestCallbackErrorTrace(t *testing.T) {
	_, err := runScript(t, `list.of(1).map(func(x){ missing })`)
	if err == nil {
		t.Fatal("expected error")
	}
	var re RuntimeError
	if !errors.As(err, &re) {
		t.Fatalf("expected RuntimeError, got %T %v", err, err)
	}
	st := re.Stacktrace()
	if !strings.Contains(st, "unresolved") {
		t.Fatalf("trace:\n%s", st)
	}
	sawMap := false
	for _, f := range re.Frames() {
		if f.Caller == "map" && !f.Internal {
			sawMap = true
			break
		}
	}
	if !sawMap {
		t.Fatalf("expected a non-internal map call in the trace:\n%s", st)
	}
}
