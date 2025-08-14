package main

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"rice/exec"
	"rice/exec/ast"
	"rice/exec/conf"
	"rice/exec/types"
	"rice/frontend"
	"strings"
	"sync"
)

func main() {
	it := exec.NewInterpreter(conf.NewDefaultEnvConfig())

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var wg sync.WaitGroup
	messages := make(chan string)
	wg.Add(1)

	go runInterpreter(ctx, it, &wg, messages)
	runREPL(messages)

	fmt.Println("\nShutting down...")
	close(messages)
	wg.Wait()

	fmt.Println("Exit.")
}

func runInterpreter(ctx context.Context, it *exec.Interpreter, wg *sync.WaitGroup, messages <-chan string) {
	defer wg.Done()

	stmtProducer := func(yield func(ast.Stmt) bool) {
		for msg := range messages {
			tokens, err := frontend.Tokenize(msg)
			if err != nil {
				fmt.Printf("Lexing Error: %v\n", err)
				fmt.Print("> ")
				continue
			}

			fmt.Print("* Tokens: ")
			for _, token := range tokens {
				fmt.Print(token.Type())
				fmt.Print(" ")
			}
			fmt.Print("\n")

			parser := frontend.NewParser(tokens)
			astTree := parser.Parse()

			if len(parser.Errors()) > 0 {
				for i, syntaxError := range parser.Errors() {
					fmt.Printf("Parsing Error #%d: %v\n", i+1, syntaxError)
				}
				fmt.Print("> ")
				continue
			}

			for _, stmt := range astTree {
				fmt.Print("* AST: ")
				fmt.Print(stmt)
				fmt.Print("\n")

				if !yield(stmt) {
					return
				}
			}
		}
	}

	runCfg := conf.NewDefaultRunConfig()
	_, _ = it.InterpretStream(ctx, stmtProducer, runCfg, func(val types.Value, err error) bool {
		if err != nil {
			var re exec.RuntimeError
			if errors.As(err, &re) {
				fmt.Println(re.Stacktrace())
			} else {
				fmt.Println(err)
			}
		} else if val != nil {
			fmt.Println(val)
		}
		fmt.Println()
		fmt.Print("> ")
		return true
	})

	fmt.Println("\n" + strings.Repeat("-", 80))
	fmt.Println("Interpreter session finished.")
}

func runREPL(messages chan<- string) {
	fmt.Println("Scripting REPL")
	fmt.Println("Enter code to execute. Type 'q' or press Ctrl+D to exit.")
	fmt.Println(strings.Repeat("-", 60))

	scanner := bufio.NewScanner(os.Stdin)

	fmt.Print("> ")
	for {

		if !scanner.Scan() {
			continue
		}

		line := strings.TrimSpace(scanner.Text())

		if strings.ToLower(line) == "q" {
			break
		}
		if line == "" {
			continue
		}

		messages <- line
	}

	if err := scanner.Err(); err != nil && err != io.EOF {
		fmt.Printf("Error reading from stdin: %v", err)
	}
}
