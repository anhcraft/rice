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
	"github.com/anhcraft/rice/exec/fun"
	"github.com/anhcraft/rice/exec/stdlib"
	"github.com/anhcraft/rice/exec/types"
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
		{name: "Break and Continue", filename: "break_continue.rice", expected: values.Bool(true)},
		{name: "Nested Loop Control (break/continue/return)", filename: "nested_loop_control.rice", expected: values.Bool(true)},
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

		// --- Object literal ---
		{name: "Object Literal", filename: "object_literal.rice", expected: values.Bool(true)},

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

// TestCustomPackageIntegration verifies the new package integration features.
func TestCustomPackageIntegration(t *testing.T) {
	// Helper: parse and run a script with a given interpreter config.
	run := func(t *testing.T, it *Interpreter, script string) (types.Value, error) {
		t.Helper()
		tokens, err := frontend.Tokenize(script)
		if err != nil {
			t.Fatalf("Tokenize failed: %v", err)
		}
		parser := frontend.NewParser(tokens)
		ast := parser.Parse()
		if len(parser.Errors()) > 0 {
			t.Fatalf("Parse failed: %v", parser.Errors()[0])
		}
		return it.Interpret(context.Background(), ast, conf.NewDefaultRunConfig())
	}

	t.Run("disable io package", func(t *testing.T) {
		it := NewInterpreter(conf.NewDefaultEnvConfig().
			DisableNamespacedPackage("io"))
		// print should be unavailable, this should error
		_, err := run(t, it, `print("hello")`)
		if err == nil {
			t.Fatal("expected error when calling print after disabling io, got nil")
		}
	})

	t.Run("disable error package", func(t *testing.T) {
		it := NewInterpreter(conf.NewDefaultEnvConfig().
			DisableNamespacedPackage("error"))
		_, err := run(t, it, `throw("test")`)
		if err == nil {
			t.Fatal("expected error when calling throw() after disabling error pkg, got nil")
		}
	})

	t.Run("disable io but keep other globals", func(t *testing.T) {
		it := NewInterpreter(conf.NewDefaultEnvConfig().
			DisableNamespacedPackage("io"))
		// typeof() from stdlib/type should still work
		val, err := run(t, it, `typeof(list.of(1,2,3))`)
		if err != nil {
			t.Fatalf("typeof() should still work after disabling io, got: %v", err)
		}
		if val == nil {
			t.Fatal("expected non-nil result from typeof()")
		}
	})

	t.Run("disable non-existent package is no-op", func(t *testing.T) {
		it := NewInterpreter(conf.NewDefaultEnvConfig().
			DisableNamespacedPackage("nonexistent_pkg"))
		// typeof should still be available
		val, err := run(t, it, `typeof("hello")`)
		if err != nil {
			t.Fatalf("typeof() should work, got: %v", err)
		}
		if val == nil {
			t.Fatal("expected non-nil result")
		}
	})

	t.Run("disable io does not affect namespaced packages", func(t *testing.T) {
		it := NewInterpreter(conf.NewDefaultEnvConfig().
			DisableNamespacedPackage("io"))
		// strings package should still work
		val, err := run(t, it, `strings.toUpper("hello")`)
		if err != nil {
			t.Fatalf("strings.toUpper should work after disabling io, got: %v", err)
		}
		expected := values.String("HELLO")
		if !reflect.DeepEqual(expected, val) {
			t.Errorf("expected %v, got %v", expected, val)
		}
	})

	t.Run("add custom namespaced package", func(t *testing.T) {
		pkg := fun.FunctionPackage{
			"greet": {stdlib.Define(func(name values.String) (types.Value, error) {
				return values.String("Hello, " + string(name) + "!"), nil
			})},
		}
		it := NewInterpreter(conf.NewDefaultEnvConfig().
			AddNamespacedFunctionPackage("tools", pkg))
		val, err := run(t, it, `tools.greet("World")`)
		if err != nil {
			t.Fatalf("tools.greet should work, got: %v", err)
		}
		expected := values.String("Hello, World!")
		if !reflect.DeepEqual(expected, val) {
			t.Errorf("expected %v, got %v", expected, val)
		}
	})

	t.Run("strict stdlib mode with enabled subset", func(t *testing.T) {
		it := NewInterpreter(conf.NewDefaultEnvConfig().
			SetStrictStdlibMode(true).
			EnableNamespacedPackage("math")) // only math is available
		// math should work
		val, err := run(t, it, `math.floor(3.7)`)
		if err != nil {
			t.Fatalf("math.floor should work in strict mode, got: %v", err)
		}
		if val != values.Float(3) {
			t.Errorf("expected 3, got %v", val)
		}
		// strings should NOT be available
		_, err = run(t, it, `strings.toUpper("x")`)
		if err == nil {
			t.Fatal("expected error when calling strings.toUpper in strict mode without enabling strings")
		}
	})

	t.Run("strict stdlib mode with custom package", func(t *testing.T) {
		pkg := fun.FunctionPackage{
			"halve": {stdlib.Define(func(x values.Float) (types.Value, error) {
				return values.Float(x / 2), nil
			})},
		}
		it := NewInterpreter(conf.NewDefaultEnvConfig().
			SetStrictStdlibMode(true).
			EnableNamespacedPackage("math").
			AddGlobalFunctionPackage(pkg))
		// math should work
		val, err := run(t, it, `math.abs(-5)`)
		if err != nil {
			t.Fatalf("math.abs should work, got: %v", err)
		}
		if !reflect.DeepEqual(values.Int(5), val) {
			t.Errorf("expected 5, got %v", val)
		}
		// custom global function should work
		val, err = run(t, it, `halve(10.0)`)
		if err != nil {
			t.Fatalf("halve should work, got: %v", err)
		}
		if !reflect.DeepEqual(values.Float(5.0), val) {
			t.Errorf("expected 5.0, got %v", val)
		}
		// strings should NOT be available
		_, err = run(t, it, `strings.toUpper("x")`)
		if err == nil {
			t.Fatal("expected error when calling strings.toUpper in strict mode without enabling strings")
		}
	})

	t.Run("disable type-bound string package", func(t *testing.T) {
		it := NewInterpreter(conf.NewDefaultEnvConfig().
			DisableTypeBoundPackage(types.String))
		// type-bound string methods like .toUpper() should not be available
		_, err := run(t, it, `const x = "hello"; x.toUpper()`)
		if err == nil {
			t.Fatal("expected error when calling string.toUpper after disabling type-bound string pkg")
		}
	})

	t.Run("backward compatibility - default config works", func(t *testing.T) {
		it := NewInterpreter(conf.NewDefaultEnvConfig())
		val, err := run(t, it, `math.abs(-42)`)
		if err != nil {
			t.Fatalf("math.abs should work with default config, got: %v", err)
		}
		if !reflect.DeepEqual(values.Int(42), val) {
			t.Errorf("expected 42, got %v", val)
		}
		// io should work
		_, err = run(t, it, `print("hello")`)
		if err != nil {
			t.Fatalf("print should work with default config, got: %v", err)
		}
	})
}
