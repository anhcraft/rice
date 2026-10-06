package conformance

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"testing"

	"github.com/anhcraft/rice/exec"
	"github.com/anhcraft/rice/exec/ast"
	"github.com/anhcraft/rice/exec/compiler"
	"github.com/anhcraft/rice/exec/conf"
	"github.com/anhcraft/rice/exec/types"
	"github.com/anhcraft/rice/exec/types/values"
	"github.com/anhcraft/rice/exec/vm"
	"github.com/anhcraft/rice/frontend"
)

var update = flag.Bool("update", false, "update golden .ricebc files and expected.txt")

func TestRoundTripTestdata(t *testing.T) {
	root := filepath.Join("..", "testdata")
	entries, err := os.ReadDir(root)
	if err != nil {
		t.Fatal(err)
	}
	goldDir := "testdata"
	if err := os.MkdirAll(goldDir, 0755); err != nil {
		t.Fatal(err)
	}

	wantResults := loadExpected(t)
	got := map[string]string{}

	for _, e := range entries {
		if e.IsDir() || filepath.Ext(e.Name()) != ".rice" {
			continue
		}
		name := e.Name()
		t.Run(name, func(t *testing.T) {
			src, err := os.ReadFile(filepath.Join(root, name))
			if err != nil {
				t.Fatal(err)
			}
			stmts, err := parseSource(string(src))
			if err != nil {
				t.Skip(err.Error())
			}
			mod, err := compiler.Compile(stmts)
			if err != nil {
				line := formatExpected(name, nil, err)
				got[name] = line
				if !*update {
					checkExpected(t, wantResults, name, line)
				}
				return
			}
			raw, err := mod.Encode()
			if err != nil {
				t.Fatal(err)
			}
			decoded, err := vm.Decode(raw)
			if err != nil {
				t.Fatal(err)
			}
			again, err := decoded.Encode()
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(raw, again) {
				t.Fatal("encode/decode mismatch")
			}

			gold := filepath.Join(goldDir, strings.TrimSuffix(name, ".rice")+".ricebc")
			if *update {
				if err := os.WriteFile(gold, raw, 0644); err != nil {
					t.Fatal(err)
				}
			} else {
				want, err := os.ReadFile(gold)
				if err != nil {
					t.Fatalf("missing golden %s (run go test ./exec/conformance -update)", gold)
				}
				if !bytes.Equal(raw, want) {
					t.Fatalf("golden mismatch for %s", name)
				}
			}

			it := exec.NewInterpreter(conf.NewDefaultEnvConfig())
			val, runErr := it.ExecuteModule(context.Background(), decoded, conf.NewDefaultRunConfig())
			line := formatExpected(name, val, runErr)
			got[name] = line
			if !*update {
				checkExpected(t, wantResults, name, line)
			}
		})
	}

	if *update {
		writeExpected(t, "expected.txt", got)
	}
}

func TestVectorModules(t *testing.T) {
	dir := "vectors"
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	cfg := conf.NewDefaultRunConfig()
	for _, e := range entries {
		if e.IsDir() || filepath.Ext(e.Name()) != ".rice" {
			continue
		}
		name := e.Name()
		t.Run(name, func(t *testing.T) {
			src, err := os.ReadFile(filepath.Join(dir, name))
			if err != nil {
				t.Fatal(err)
			}
			stmts, err := parseSource(string(src))
			if err != nil {
				t.Fatal(err)
			}
			mod, err := compiler.Compile(stmts)
			if err != nil {
				t.Fatal(err)
			}
			raw, err := mod.Encode()
			if err != nil {
				t.Fatal(err)
			}
			gold := filepath.Join(dir, strings.TrimSuffix(name, ".rice")+".ricebc")
			if *update {
				if err := os.WriteFile(gold, raw, 0644); err != nil {
					t.Fatal(err)
				}
			} else {
				want, err := os.ReadFile(gold)
				if err != nil {
					t.Fatalf("missing golden %s (run go test ./exec/conformance -update)", gold)
				}
				if !bytes.Equal(raw, want) {
					t.Fatalf("golden mismatch for %s", name)
				}
			}
			it := exec.NewInterpreter(conf.NewDefaultEnvConfig())
			val, err := it.ExecuteModule(context.Background(), mod, cfg)
			if err != nil {
				t.Fatal(err)
			}
			if val != values.Bool(true) {
				t.Fatalf("got %v (%T)", val, val)
			}
		})
	}
}

func TestFrozenCompat(t *testing.T) {
	dir := filepath.Join("compat", "1.0")
	want := loadExpectedFile(t, filepath.Join(dir, "expected.txt"))
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	cfg := conf.NewDefaultRunConfig()
	var found int
	for _, e := range entries {
		if e.IsDir() || filepath.Ext(e.Name()) != ".ricebc" {
			continue
		}
		found++
		name := e.Name()
		srcName := strings.TrimSuffix(name, ".ricebc") + ".rice"
		t.Run(name, func(t *testing.T) {
			raw, err := os.ReadFile(filepath.Join(dir, name))
			if err != nil {
				t.Fatal(err)
			}
			mod, err := vm.Decode(raw)
			if err != nil {
				t.Fatal(err)
			}
			if mod.Version.Major != 1 || mod.Version.Minor != 0 {
				t.Fatalf("compat blob version %s, want 1.0", mod.Version)
			}
			it := exec.NewInterpreter(conf.NewDefaultEnvConfig())
			val, runErr := it.ExecuteModule(context.Background(), mod, cfg)
			line := formatExpected(srcName, val, runErr)
			checkExpected(t, want, srcName, line)
		})
	}
	if found == 0 {
		t.Fatal("no .ricebc files in compat/1.0")
	}
}

func TestFloatFormatVectors(t *testing.T) {
	cases := []struct {
		in   values.Float
		want string
	}{
		{1e6, "1e+06"},
		{1.0, "1"},
		{3.14, "3.14"},
	}
	for _, tc := range cases {
		got := string(values.AsString(tc.in))
		if got != tc.want {
			t.Errorf("string(%v) = %q, want %q", tc.in, got, tc.want)
		}
		ref := strconv.FormatFloat(float64(tc.in), 'g', -1, 64)
		if got != ref {
			t.Errorf("string(%v) = %q, FormatFloat = %q", tc.in, got, ref)
		}
	}
}

func TestParseVectors(t *testing.T) {
	i, err := values.AsInt(values.String("42"))
	if err != nil || i != 42 {
		t.Fatalf("int: %v %v", i, err)
	}
	f, err := values.AsFloat(values.String("1.5"))
	if err != nil || f != 1.5 {
		t.Fatalf("float: %v %v", f, err)
	}
	for _, s := range []string{"1", "t", "T", "TRUE", "true", "True"} {
		b, err := values.AsBool(values.String(s))
		if err != nil || !b {
			t.Fatalf("bool(%q) = %v, %v", s, b, err)
		}
	}
	for _, s := range []string{"0", "f", "F", "FALSE", "false", "False"} {
		b, err := values.AsBool(values.String(s))
		if err != nil || b {
			t.Fatalf("bool(%q) = %v, %v", s, b, err)
		}
	}
}

func TestStringOrderVectors(t *testing.T) {
	pairs := []struct {
		a, b values.String
		lt   bool
	}{
		{"a", "b", true},
		{"é", "ê", true},
		{"e\u0301", "\u00e9", true},
		{"b", "a", false},
	}
	for _, tc := range pairs {
		got := tc.a < tc.b
		if got != tc.lt {
			t.Errorf("%q < %q = %v, want %v", tc.a, tc.b, got, tc.lt)
		}
	}
}

func parseSource(src string) ([]ast.Stmt, error) {
	tokens, err := frontend.Tokenize(src)
	if err != nil {
		return nil, fmt.Errorf("tokenize: %w", err)
	}
	p := frontend.NewParser(tokens)
	stmts := p.Parse()
	if len(p.Errors()) > 0 {
		return nil, fmt.Errorf("parse error: %v", p.Errors()[0])
	}
	return stmts, nil
}

func formatExpected(name string, val types.Value, err error) string {
	if err != nil {
		return name + "\terror"
	}
	switch v := val.(type) {
	case nil:
		return name + "\tnull"
	case values.Bool:
		return name + "\tbool\t" + strconv.FormatBool(bool(v))
	case values.Int:
		return name + "\tint\t" + strconv.FormatInt(int64(v), 10)
	case values.Float:
		return name + "\tfloat\t" + strconv.FormatFloat(float64(v), 'g', -1, 64)
	case values.String:
		return name + "\tstring\t" + strconv.Quote(string(v))
	default:
		return name + "\tother\t" + fmt.Sprint(val)
	}
}

func loadExpected(t *testing.T) map[string]string {
	t.Helper()
	return loadExpectedFile(t, "expected.txt")
}

func loadExpectedFile(t *testing.T, path string) map[string]string {
	t.Helper()
	out := map[string]string{}
	data, err := os.ReadFile(path)
	if err != nil {
		if *update && path == "expected.txt" {
			return out
		}
		t.Fatalf("missing %s: %v", path, err)
	}
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		name, _, _ := strings.Cut(line, "\t")
		out[name] = line
	}
	return out
}

func checkExpected(t *testing.T, want map[string]string, name, got string) {
	t.Helper()
	line, ok := want[name]
	if !ok {
		t.Fatalf("no expected line for %s (run -update)", name)
	}
	if line != got {
		t.Fatalf("expected %q, got %q", line, got)
	}
}

func writeExpected(t *testing.T, path string, got map[string]string) {
	t.Helper()
	names := make([]string, 0, len(got))
	for n := range got {
		names = append(names, n)
	}
	sort.Strings(names)
	var b strings.Builder
	b.WriteString("# filename, kind (null|bool|int|float|string|error|other), optional payload\n")
	for _, n := range names {
		b.WriteString(got[n])
		b.WriteByte('\n')
	}
	if err := os.WriteFile(path, []byte(b.String()), 0644); err != nil {
		t.Fatal(err)
	}
}
