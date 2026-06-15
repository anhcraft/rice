package exec

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/anhcraft/rice/exec/conf"
	"github.com/anhcraft/rice/exec/types/values"
	"github.com/anhcraft/rice/frontend"
)

func TestInterpreterScripts(t *testing.T) {
	testCases := []struct {
		name          string
		filename      string
		expected      any
		expectError   bool
		errorContains string
	}{
		// --- Existing tests (unchanged) ---
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

		// --- Core language features ---
		{name: "Unicode Literals", filename: "unicode.rice", expected: values.Bool(true)},
		{name: "Scientific Notation", filename: "scientific_notation.rice", expected: values.Bool(true)},
		{name: "Modulo Operator", filename: "modulo_ops.rice", expected: values.Bool(true)},
		{name: "Increment/Decrement", filename: "inc_dec.rice", expected: values.Bool(true)},
		{name: "Break and Continue", filename: "break_continue.rice", expected: values.Bool(true)},
		{name: "Return Expression", filename: "return_expr.rice", expected: values.Bool(true)},
		{name: "Assignment Expression", filename: "assignment.rice", expected: values.Bool(true)},
		{name: "Anonymous Functions", filename: "anonymous_func.rice", expected: values.Bool(true)},
		{name: "Closures", filename: "closures.rice", expected: values.Bool(true)},
		{name: "Varargs", filename: "varargs.rice", expected: values.Bool(true)},
		{name: "Spread Operator", filename: "spread.rice", expected: values.Bool(true)},
		{name: "Parenthesized Expressions", filename: "paren_expr.rice", expected: values.Bool(true)},
		{name: "If-Else Chain", filename: "if_else_chain.rice", expected: values.Bool(true)},
		{name: "Operator Precedence", filename: "operator_precedence.rice", expected: values.Bool(true)},
		{name: "For Loop Clauses", filename: "for_clauses.rice", expected: values.Bool(true)},
		{name: "Block Expression", filename: "block_expr.rice", expected: values.Bool(true)},
		{name: "Nested Shadowing", filename: "nested_shadowing.rice", expected: values.Bool(true)},
		{name: "Postfix Chaining", filename: "postfix_chaining.rice", expected: values.Bool(true)},
		{name: "Null Literal", filename: "null_literal.rice", expected: values.Bool(true)},

		// --- Collection operations ---
		{name: "String Operations", filename: "string_ops.rice", expected: values.Bool(true)},
		{name: "List Operations", filename: "list_ops.rice", expected: values.Bool(true)},
		{name: "Set Operations", filename: "set_ops.rice", expected: values.Bool(true)},
		{name: "Map Operations", filename: "map_ops.rice", expected: values.Bool(true)},
		{name: "Element Access", filename: "element_access.rice", expected: values.Bool(true)},
		{name: "Selector Expressions", filename: "selector.rice", expected: values.Bool(true)},

		// --- Type-bound functions ---
		{name: "Type-Bound Functions", filename: "type_bound_funcs.rice", expected: values.Bool(true)},

		// --- Standard library ---
		{name: "Type Conversions", filename: "type_conversions.rice", expected: values.Bool(true)},
		{name: "Math Functions", filename: "math_funcs.rice", expected: values.Bool(true)},
		{name: "Error Handling", filename: "error_handling.rice", expected: values.Bool(true)},
		{name: "JSON Operations", filename: "json_ops.rice", expected: values.Bool(true)},
		{name: "DateTime Now", filename: "datetime_now.rice", expected: values.Bool(true)},
		{name: "String Native Functions", filename: "strings_native.rice", expected: values.Bool(true)},
		{name: "I/O Functions", filename: "io_funcs.rice", expected: values.Bool(true)},

		// --- Functional programming ---
		{name: "Functional Programming", filename: "functional_programming.rice", expected: values.Bool(true)},

		// --- Implicit coercion ---
		{name: "Implicit Coercion", filename: "implicit_coercion.rice", expected: values.Bool(true)},
	}

	it := NewInterpreter(conf.NewDefaultEnvConfig())
	runConf := conf.NewDefaultRunConfig()

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

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
