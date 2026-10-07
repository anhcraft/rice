//go:build rice_debug

package vm_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/anhcraft/rice/exec/ast"
	"github.com/anhcraft/rice/exec/compiler"
	"github.com/anhcraft/rice/exec/types"
	"github.com/anhcraft/rice/exec/types/values"
	"github.com/anhcraft/rice/exec/vm"
	"github.com/anhcraft/rice/frontend"
)

type stubHost struct {
	ctx context.Context
}

func (h *stubHost) LoadGlobal(name string) (types.Value, error) {
	return nil, fmt.Errorf("unresolved reference %q", name)
}
func (h *stubHost) StoreGlobal(string, types.Value) error        { return nil }
func (h *stubHost) DefineGlobal(string, types.Value, bool) error { return nil }
func (h *stubHost) IsNamespaceEntry(string) bool                 { return false }
func (h *stubHost) TypeBound(types.Value, values.Identifier) (types.Value, bool) {
	return nil, false
}
func (h *stubHost) ProfilerStart(string, ast.Pos) {}
func (h *stubHost) ProfilerEnd()                  {}
func (h *stubHost) ProfilerDepth() int            { return 0 }
func (h *stubHost) ProfilerUnwind(int)            {}
func (h *stubHost) Context() context.Context {
	if h.ctx == nil {
		return context.Background()
	}
	return h.ctx
}
func (h *stubHost) SetContext(ctx context.Context) { h.ctx = ctx }
func (h *stubHost) CheckContext() error            { return nil }
func (h *stubHost) ScopeLimit() uint32             { return 256 }
func (h *stubHost) UserFuncTimeout() time.Duration { return time.Second }

func TestDebugger(t *testing.T) {
	tokens, err := frontend.Tokenize(`1 + 2`)
	if err != nil {
		t.Fatal(err)
	}
	p := frontend.NewParser(tokens)
	tree := p.Parse()
	if len(p.Errors()) > 0 {
		t.Fatal(p.Errors()[0])
	}
	mod, err := compiler.Compile(tree)
	if err != nil {
		t.Fatal(err)
	}

	debug := (&vm.VM{}).Debug()
	got := make([]vm.DebugPos, 0)
	done := make(chan struct{})
	go func() {
		defer close(done)
		for pos := range debug.Position() {
			got = append(got, pos)
			debug.Step()
		}
	}()
	go debug.Step()

	val, err := debug.Run(mod, &stubHost{})
	if err != nil {
		t.Fatal(err)
	}
	<-done

	if val != values.Int(3) {
		t.Fatalf("got %v", val)
	}
	entry := int(mod.Entry)
	fn := mod.Functions[entry]
	if len(got) == 0 {
		t.Fatal("expected position events")
	}
	seen := make(map[int]bool)
	for _, pos := range got {
		if pos.Func != entry {
			t.Fatalf("unexpected func %d, want %d", pos.Func, entry)
		}
		if pos.IP < 0 || pos.IP > len(fn.Bytecode) {
			t.Fatalf("ip %d out of range 0..%d", pos.IP, len(fn.Bytecode))
		}
		seen[pos.IP] = true
	}
	if len(seen) == 0 {
		t.Fatal("positions did not cover any bytecode ip")
	}
}
