package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/anhcraft/rice/exec"
	"github.com/anhcraft/rice/exec/ast"
	"github.com/anhcraft/rice/exec/conf"
	"github.com/anhcraft/rice/exec/types"
	"github.com/anhcraft/rice/frontend"
)

func cmdRepl(args []string) error {
	fs := flag.NewFlagSet("repl", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	debug := fs.Bool("debug", false, "print tokens, checksum, and AST")
	var profile bool
	var timeout time.Duration
	var scope uint
	addRunFlags(fs, &profile, &timeout, &scope)
	if err := parseCmdFlags(fs, args); err != nil {
		return err
	}

	it := newInterpreter(profile)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	messages := make(chan string)
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		runReplInterpreter(ctx, it, messages, *debug, runConfig(timeout, scope))
		if profile {
			fmt.Fprint(os.Stderr, it.Profiler().Report())
		}
	}()

	fmt.Println("Rice REPL")
	fmt.Println("Enter code to execute. Type 'q' or press Ctrl+D to exit.")
	fmt.Println(strings.Repeat("-", 60))
	fmt.Print("> ")

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.EqualFold(line, "q") {
			break
		}
		if line == "" {
			fmt.Print("> ")
			continue
		}
		messages <- line
	}
	if err := scanner.Err(); err != nil && err != io.EOF {
		fmt.Fprintf(os.Stderr, "Error reading from stdin: %v\n", err)
	}

	close(messages)
	wg.Wait()
	return nil
}

func runReplInterpreter(ctx context.Context, it *exec.Interpreter, messages <-chan string, debug bool, cfg *conf.RunConfig) {
	stmtProducer := func(yield func(ast.Stmt) bool) {
		for msg := range messages {
			tokens, err := frontend.Tokenize(msg)
			if err != nil {
				fmt.Printf("Lexing Error: %v\n", err)
				fmt.Print("> ")
				continue
			}
			if debug {
				fmt.Print("* Tokens: ")
				for _, token := range tokens {
					fmt.Print(token.Type(), " ")
				}
				fmt.Print("\n")
				fmt.Printf("* Checksum: %s\n", frontend.Checksum(tokens))
			}

			parser := frontend.NewParser(tokens)
			tree := parser.Parse()
			if len(parser.Errors()) > 0 {
				for i, syntaxError := range parser.Errors() {
					fmt.Printf("Parsing Error #%d: %v\n", i+1, syntaxError)
				}
				fmt.Print("> ")
				continue
			}

			for _, stmt := range tree {
				if debug {
					fmt.Printf("* AST: %s\n", stmt)
				}
				if !yield(stmt) {
					return
				}
			}
		}
	}

	_, _ = it.InterpretStream(ctx, stmtProducer, cfg, func(val types.Value, err error) bool {
		if err != nil {
			printErr(err)
		} else {
			printValue(val)
		}
		fmt.Println()
		fmt.Print("> ")
		return true
	})
}
