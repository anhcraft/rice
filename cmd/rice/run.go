package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/anhcraft/rice/exec/types"
	"github.com/anhcraft/rice/exec/vm"
)

func cmdRun(args []string) error {
	fs := flag.NewFlagSet("run", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	var profile bool
	var timeout time.Duration
	var scope uint
	addRunFlags(fs, &profile, &timeout, &scope)
	if err := parseCmdFlags(fs, args); err != nil {
		return err
	}
	path := ""
	if fs.NArg() > 0 {
		path = fs.Arg(0)
	}
	data, err := readInput(path)
	if err != nil {
		return err
	}
	mod, err := loadModule(data)
	if err != nil {
		return err
	}
	it := newInterpreter(profile)
	val, err := it.ExecuteModule(context.Background(), mod, runConfig(timeout, scope))
	if err != nil {
		return err
	}
	printValue(val)
	if profile {
		fmt.Fprint(os.Stderr, it.Profiler().Report())
	}
	return nil
}

func cmdBuild(args []string) error {
	fs := flag.NewFlagSet("build", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	out := fs.String("o", "", "output .ricebc path")
	if err := parseCmdFlags(fs, args); err != nil {
		return err
	}
	if fs.NArg() < 1 {
		return fmt.Errorf("usage: rice build [-o out.ricebc] file.rice")
	}
	path := fs.Arg(0)
	data, err := readInput(path)
	if err != nil {
		return err
	}
	if isRicebc(data) {
		return fmt.Errorf("%s is already a RICE module", path)
	}
	mod, err := compileSource(string(data))
	if err != nil {
		return err
	}
	raw, err := mod.Encode()
	if err != nil {
		return err
	}
	dest := *out
	if dest == "" {
		if path == "" || path == "-" {
			return fmt.Errorf("rice build: -o is required when reading stdin")
		}
		dest = defaultRicebcName(path)
	}
	return os.WriteFile(dest, raw, 0644)
}

func cmdDisasm(args []string) error {
	fs := flag.NewFlagSet("disasm", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	if err := parseCmdFlags(fs, args); err != nil {
		return err
	}
	path := ""
	if fs.NArg() > 0 {
		path = fs.Arg(0)
	}
	data, err := readInput(path)
	if err != nil {
		return err
	}
	mod, err := loadModule(data)
	if err != nil {
		return err
	}
	fmt.Print(mod.Disassemble())
	return nil
}

func cmdCheck(args []string) error {
	fs := flag.NewFlagSet("check", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	if err := parseCmdFlags(fs, args); err != nil {
		return err
	}
	if fs.NArg() < 1 {
		return fmt.Errorf("usage: rice check file.rice")
	}
	data, err := readInput(fs.Arg(0))
	if err != nil {
		return err
	}
	if isRicebc(data) {
		_, err = vm.Decode(data)
		return err
	}
	_, err = compileSource(string(data))
	return err
}

func printValue(val types.Value) {
	if val != nil {
		fmt.Println(val)
	}
}
