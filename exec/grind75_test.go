package exec

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/anhcraft/rice/exec/conf"
	"github.com/anhcraft/rice/frontend"
)

func TestGrind75TestSuite(t *testing.T) {
	suiteDir := filepath.Join("testdata", "grind75")

	entries, err := os.ReadDir(suiteDir)
	if err != nil {
		t.Fatalf("Failed to read testsuite directory '%s': %v", suiteDir, err)
	}

	var riceFiles []string
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".rice") {
			riceFiles = append(riceFiles, entry.Name())
		}
	}

	if len(riceFiles) == 0 {
		t.Fatal("No .rice files found in testsuite directory")
	}

	it := NewInterpreter(conf.NewDefaultEnvConfig())
	runConf := conf.NewDefaultRunConfig()

	for _, filename := range riceFiles {
		t.Run(filename, func(t *testing.T) {
			t.Parallel()

			scriptPath := filepath.Join(suiteDir, filename)
			scriptBytes, readErr := os.ReadFile(scriptPath)
			if readErr != nil {
				t.Fatalf("Failed to read test file '%s': %v", scriptPath, readErr)
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

			_, interpretErr := it.Interpret(context.Background(), ast, runConf)

			if interpretErr != nil {
				var re RuntimeError
				if errors.As(interpretErr, &re) {
					t.Errorf("Script failed with RuntimeError:\n%s", re.Stacktrace())
				} else {
					t.Errorf("Script failed with error: %v", interpretErr)
				}
			}
		})
	}
}
