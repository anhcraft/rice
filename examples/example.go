package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"rice/exec"
	"rice/exec/conf"
	"rice/exec/types/values"
	"rice/frontend"
	"strings"
)

const scriptPath = "./examples/process-dataset.rice"

func main() {
	scriptBytes, err := os.ReadFile(scriptPath)
	if err != nil {
		panic(fmt.Errorf("failed to read test file '%s': %v", scriptPath, err))
	}
	script := string(scriptBytes)

	tokens, tokenizeErr := frontend.Tokenize(script)
	if tokenizeErr != nil {
		panic(fmt.Errorf("tokenize failed: %v", tokenizeErr))
	}

	parser := frontend.NewParser(tokens)
	ast := parser.Parse()
	if len(parser.Errors()) > 0 {
		panic(fmt.Errorf("parsing failed: %v", parser.Errors()[0]))
	}

	it := exec.NewInterpreter(conf.NewDefaultEnvConfig().EnableProfiling())

	result, err := it.Interpret(context.Background(), ast, conf.NewDefaultRunConfig().
		DefineConstant("NUM_RECORDS", values.Int(10000)).
		DefineConstant("CATEGORIES", values.ListOf([]values.String{"Electronics", "Books", "Home Goods", "Apparel", "Toys"})))

	fmt.Println()
	fmt.Println(strings.Repeat("-", 55))
	fmt.Println()

	if err != nil {
		fmt.Println("ERROR:")
		var re exec.RuntimeError
		if errors.As(err, &re) {
			fmt.Println(re.Stacktrace())
		} else {
			fmt.Println(err)
		}
		return
	}

	if result == nil {
		fmt.Println("RESULT:")
		fmt.Println("(none)")
	} else {
		fmt.Printf("RESULT (type %T):\n", result)
		fmt.Println(result)
	}

	fmt.Println()
	fmt.Println(strings.Repeat("-", 55))
	fmt.Println()

	fmt.Println("PROFILING:")
	fmt.Println(it.Profiler().Report())
}
