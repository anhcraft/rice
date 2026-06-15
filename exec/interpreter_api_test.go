package exec

import (
	"context"
	"errors"
	"github.com/anhcraft/rice/exec/conf"
	"github.com/anhcraft/rice/exec/types/values"
	"github.com/anhcraft/rice/frontend"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestInterpreterScripts(t *testing.T) {
	testCases := []struct {
		name          string
		filename      string
		expected      any
		expectError   bool
		errorContains string
	}{
		{name: "Literals", filename: "literals.rice", expected: nil},
		{name: "Arithmetic Operators", filename: "arithmetic_ops.rice", expected: values.Float(10.5)},
		{name: "Comparison Operators", filename: "comparison_ops.rice", expected: values.Bool(true)},
		{name: "Logical Operators", filename: "logical_ops.rice", expected: values.Bool(true)},
		{name: "Variables", filename: "variables.rice", expected: values.Int(30)},
		{name: "If Expression", filename: "if_expression.rice", expected: values.String("Positive")},
		{name: "Block Scope", filename: "block_scope.rice", expected: values.Int(100)},
		{name: "For Loop (C-style & While)", filename: "for_loop.rice", expected: values.Int(10)},
		{name: "For-In Loop", filename: "for_in_loop.rice", expected: values.Int(6)},
		{name: "User-Defined Function", filename: "user_func.rice", expected: values.Int(25)},
		{name: "Recursion (Fibonacci)", filename: "recursion.rice", expected: values.Int(55)},
		{name: "Short-Circuit Evaluation", filename: "short_circuit.rice", expected: values.Bool(true)},
	}

	it := NewInterpreter(conf.NewDefaultEnvConfig())
	runConf := conf.NewDefaultRunConfig()

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			scriptPath := filepath.Join("testdata", tc.filename)
			scriptBytes, err := os.ReadFile(scriptPath)
			if err != nil {
				t.Fatalf("Failed to read test file '%s': %v", scriptPath, err)
			}
			script := string(scriptBytes)

			tokens, tokenizeErr := frontend.Tokenize(script)
			if tokenizeErr != nil {
				t.Fatalf("Tokenize failed: %v", tokenizeErr)
			}

			parser := frontend.NewParser(tokens)
			ast := parser.Parse()
			if len(parser.Errors()) > 0 {
				t.Fatalf("Parsing failed: %v", parser.Errors()[0])
			}

			actual, interpretErr := it.Interpret(context.Background(), ast, runConf)

			if tc.expectError {
				if interpretErr == nil {
					t.Errorf("Expected an error, but got none. Output: %v", actual)
					return
				}
				if tc.errorContains != "" && !strings.Contains(interpretErr.Error(), tc.errorContains) {
					t.Errorf("Expected error to contain '%s', but got: %v", tc.errorContains, interpretErr)
				}
			} else {
				if interpretErr != nil {
					var re RuntimeError
					if errors.As(interpretErr, &re) {
						t.Errorf("Expected no error, but got RuntimeError:\n%s", re.Stacktrace())
					} else {
						t.Errorf("Expected no error, but got: %v", interpretErr)
					}
					return
				}
				if !reflect.DeepEqual(tc.expected, actual) {
					t.Errorf("Mismatched output.\nExpected: %v (%T)\nActual:   %v (%T)",
						tc.expected, tc.expected, actual, actual)
				}
			}
		})
	}
}
