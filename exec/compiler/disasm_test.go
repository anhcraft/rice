package compiler

import (
	"strings"
	"testing"

	"github.com/anhcraft/rice/frontend"
)

func TestDisassembleSnapshot(t *testing.T) {
	tokens, err := frontend.Tokenize("const x = 1 + 2; x")
	if err != nil {
		t.Fatal(err)
	}
	p := frontend.NewParser(tokens)
	stmts := p.Parse()
	if len(p.Errors()) > 0 {
		t.Fatal(p.Errors()[0])
	}
	mod, err := Compile(stmts)
	if err != nil {
		t.Fatal(err)
	}
	d := mod.Disassemble()
	for _, want := range []string{"OpAdd", "OpDefineConstGlobal", "OpLoadGlobal", "OpReturn", "OpPush"} {
		if !strings.Contains(d, want) {
			t.Fatalf("disassembly missing %s:\n%s", want, d)
		}
	}
}
