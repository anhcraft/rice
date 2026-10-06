package main

import (
	"encoding/binary"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/anhcraft/rice/exec"
	"github.com/anhcraft/rice/exec/compiler"
	"github.com/anhcraft/rice/exec/conf"
	"github.com/anhcraft/rice/exec/vm"
	"github.com/anhcraft/rice/frontend"
)

const usage = `Usage: rice <command> [arguments]

Commands:
  run      compile or decode, then execute
  repl     interactive session
  build    compile a script to .ricebc
  disasm   print a module listing
  check    parse and compile; exit status only
  version  print the RICE bytecode version
`

func main() {
	if len(os.Args) < 2 {
		fmt.Fprint(os.Stderr, usage)
		os.Exit(2)
	}
	cmd := os.Args[1]
	args := os.Args[2:]
	var err error
	switch cmd {
	case "run":
		err = cmdRun(args)
	case "repl":
		err = cmdRepl(args)
	case "build":
		err = cmdBuild(args)
	case "disasm":
		err = cmdDisasm(args)
	case "check":
		err = cmdCheck(args)
	case "version":
		fmt.Println(vm.CurrentVersion)
		return
	case "-h", "-help", "--help", "help":
		fmt.Fprint(os.Stderr, usage)
		os.Exit(2)
	default:
		fmt.Fprintf(os.Stderr, "unknown command %q\n%s", cmd, usage)
		os.Exit(2)
	}
	if err != nil {
		printErr(err)
		os.Exit(1)
	}
}

func printErr(err error) {
	var re exec.RuntimeError
	if errors.As(err, &re) {
		fmt.Fprint(os.Stderr, re.Stacktrace())
		return
	}
	fmt.Fprintln(os.Stderr, err)
}

func readInput(path string) ([]byte, error) {
	if path == "" || path == "-" {
		return io.ReadAll(os.Stdin)
	}
	return os.ReadFile(path)
}

func isRicebc(data []byte) bool {
	return len(data) >= 4 && binary.BigEndian.Uint32(data[:4]) == vm.Magic
}

func loadModule(data []byte) (*vm.Module, error) {
	if isRicebc(data) {
		return vm.Decode(data)
	}
	return compileSource(string(data))
}

func compileSource(src string) (*vm.Module, error) {
	tokens, err := frontend.Tokenize(src)
	if err != nil {
		return nil, err
	}
	p := frontend.NewParser(tokens)
	stmts := p.Parse()
	if len(p.Errors()) > 0 {
		return nil, p.Errors()[0]
	}
	return compiler.Compile(stmts)
}

func newInterpreter(profile bool) *exec.Interpreter {
	cfg := conf.NewDefaultEnvConfig()
	if profile {
		cfg = cfg.EnableProfiling()
	}
	return exec.NewInterpreter(cfg)
}

func parseCmdFlags(fs *flag.FlagSet, args []string) error {
	if err := fs.Parse(args); err != nil {
		os.Exit(2)
	}
	return nil
}

func addRunFlags(fs *flag.FlagSet, profile *bool, timeout *time.Duration, scope *uint) {
	def := conf.NewDefaultRunConfig()
	fs.BoolVar(profile, "profile", false, "print a profiler report at exit")
	fs.DurationVar(timeout, "timeout", def.UserFuncTimeout, "timeout per user-defined function")
	fs.UintVar(scope, "scope-limit", uint(def.LexicalScopeLimit), "maximum lexical scope depth")
}

func runConfig(timeout time.Duration, scope uint) *conf.RunConfig {
	return conf.NewDefaultRunConfig().
		SetUserFuncTimeout(timeout).
		SetLexicalScopeLimit(uint32(scope))
}

func defaultRicebcName(src string) string {
	ext := filepath.Ext(src)
	if ext == "" {
		return src + ".ricebc"
	}
	return strings.TrimSuffix(src, ext) + ".ricebc"
}
