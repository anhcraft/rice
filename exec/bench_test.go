//go:build benchmark

package exec

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/anhcraft/rice/exec/conf"
	"github.com/anhcraft/rice/frontend"
)

// Helper to tokenize, parse, and interpret a .rice script string.
// Returns the elapsed time reported by the profiler.
func benchScript(b *testing.B, script string) {
	it := NewInterpreter(conf.NewDefaultEnvConfig().EnableProfiling())
	runConf := conf.NewDefaultRunConfig()

	tokens, err := frontend.Tokenize(script)
	if err != nil {
		b.Fatalf("Tokenize failed: %v", err)
	}

	parser := frontend.NewParser(tokens)
	ast := parser.Parse()
	if len(parser.Errors()) > 0 {
		b.Fatalf("Parsing failed: %v", parser.Errors()[0])
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err = it.Interpret(context.Background(), ast, runConf)
		if err != nil {
			b.Fatalf("Interpret failed: %v", err)
		}
		it.Profiler().Reset()
	}
}

// Helper to load and run a .rice file as a benchmark.
func benchFile(b *testing.B, filename string) {
	scriptPath := filepath.Join("testdata", filename)
	scriptBytes, err := os.ReadFile(scriptPath)
	if err != nil {
		b.Fatalf("Failed to read file '%s': %v", scriptPath, err)
	}
	benchScript(b, string(scriptBytes))
}

// BenchmarkFibonacci measures recursive function-call overhead (fib(20)).
func BenchmarkFibonacci(b *testing.B) {
	benchScript(b, `
		const fib = func(n) {
			if n < 2 {
				return n;
			}
			return fib(n - 1) + fib(n - 2);
		};
		fib(20);
	`)
}

// BenchmarkForLoop measures tight-loop iteration overhead (0..9999 sum).
func BenchmarkForLoop(b *testing.B) {
	benchScript(b, `
		var sum = 0;
		for (var i = 0; i < 10000; i++) {
			sum = sum + i;
		}
		sum;
	`)
}

// BenchmarkFunctional measures list/map higher-order function overhead.
func BenchmarkFunctional(b *testing.B) {
	benchScript(b, `
		const data = list.of(
			map.of("name", "Alice",   "score", 85),
			map.of("name", "Bob",     "score", 42),
			map.of("name", "Charlie", "score", 73),
			map.of("name", "Diana",   "score", 91),
			map.of("name", "Eve",     "score", 58)
		);
		const passing = list.filter(data, func(student) {
			return student["score"] >= 60;
		});
		list.map(passing, func(student) {
			return student["name"];
		});
	`)
}

// BenchmarkArithmetic measures arithmetic & comparison operator throughput.
func BenchmarkArithmetic(b *testing.B) {
	benchScript(b, `
		var a = 0;
		for (var i = 0; i < 5000; i++) {
			a = a + i * 2 - i / 2;
		}
		a;
	`)
}

// BenchmarkStringOps measures string concatenation and comparison throughput.
func BenchmarkStringOps(b *testing.B) {
	benchScript(b, `
		var s = "";
		for (var i = 0; i < 1000; i++) {
			s = s + "a";
		}
		len(s);
	`)
}

// BenchmarkRecursionDeep measures moderately deep recursion (depth 15, fib style).
func BenchmarkRecursionDeep(b *testing.B) {
	benchScript(b, `
		const fib = func(n) {
			if n < 2 {
				return n;
			}
			return fib(n - 1) + fib(n - 2);
		};
		fib(15);
	`)
}

// runBenchmarksWithProfiler is a helper that executes a script once with profiling
// enabled and returns the profiler report string. This is used to extract a single-run
// elapsed time for the README table.
func runOnceWithProfiler(script string) (string, error) {
	it := NewInterpreter(conf.NewDefaultEnvConfig().EnableProfiling())
	runConf := conf.NewDefaultRunConfig()

	tokens, err := frontend.Tokenize(script)
	if err != nil {
		return "", fmt.Errorf("tokenize: %w", err)
	}

	parser := frontend.NewParser(tokens)
	ast := parser.Parse()
	if len(parser.Errors()) > 0 {
		return "", fmt.Errorf("parse: %v", parser.Errors()[0])
	}

	_, err = it.Interpret(context.Background(), ast, runConf)
	if err != nil {
		return "", fmt.Errorf("interpret: %w", err)
	}

	return it.Profiler().Report(), nil
}

// printBenchmarkProfiles prints single-run profiler reports for the key benchmarks.
// Use: go test -run TestBenchmarkProfiles -v ./exec/
func TestBenchmarkProfiles(t *testing.T) {
	tests := []struct {
		name   string
		script string
	}{
		{"Fibonacci(20)", `
			const fib = func(n) {
				if n < 2 {
					return n;
				}
				return fib(n - 1) + fib(n - 2);
			};
			fib(20);
		`},
		{"ForLoop(10000)", `
			var sum = 0;
			for (var i = 0; i < 10000; i++) {
				sum = sum + i;
			}
			sum;
		`},
		{"Functional", `
			const data = list.of(
				map.of("name", "Alice",   "score", 85),
				map.of("name", "Bob",     "score", 42),
				map.of("name", "Charlie", "score", 73),
				map.of("name", "Diana",   "score", 91),
				map.of("name", "Eve",     "score", 58)
			);
			const passing = list.filter(data, func(student) {
				return student["score"] >= 60;
			});
			list.map(passing, func(student) {
				return student["name"];
			});
		`},
		{"Arithmetic(5000)", `
			var a = 0;
			for (var i = 0; i < 5000; i++) {
				a = a + i * 2 - i / 2;
			}
			a;
		`},
		{"StringConcat(1000)", `
			var s = "";
			for (var i = 0; i < 1000; i++) {
				s = s + "a";
			}
			len(s);
		`},
	}

	separator := strings.Repeat("-", 70)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			report, err := runOnceWithProfiler(tt.script)
			if err != nil {
				t.Fatalf("Failed: %v", err)
			}
			fmt.Println(separator)
			fmt.Printf("Profile: %s\n", tt.name)
			fmt.Println(separator)
			fmt.Print(report)
		})
	}
}
