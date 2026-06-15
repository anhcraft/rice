package main

import (
	"bytes"
	"context"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/anhcraft/rice/exec"
	"github.com/anhcraft/rice/exec/conf"
	"github.com/anhcraft/rice/frontend"
)

//go:embed index.html
var indexHTML string

//go:embed monarch.js
var monarchJS string

//go:embed hello.rice
var exampleHello string

//go:embed functions.rice
var exampleFunctions string

//go:embed loops.rice
var exampleLoops string

//go:embed conditionals.rice
var exampleConditionals string

//go:embed maps.rice
var exampleMaps string

type example struct {
	Label string `json:"label"`
	Code  string `json:"code"`
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", handleIndex)
	mux.HandleFunc("/execute", handleExecute)
	mux.HandleFunc("/monarch.js", handleMonarch)
	mux.HandleFunc("/examples", handleExamples)

	fmt.Printf("Rice Playground listening on http://localhost:%s\n", port)
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		fmt.Fprintf(os.Stderr, "server error: %v\n", err)
		os.Exit(1)
	}
}

func handleIndex(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(indexHTML))
}

func handleMonarch(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/javascript")
	w.Write([]byte(monarchJS))
}

func handleExamples(w http.ResponseWriter, r *http.Request) {
	examples := map[string]example{
		"hello":        {Label: "Hello World", Code: exampleHello},
		"functions":    {Label: "Functions", Code: exampleFunctions},
		"loops":        {Label: "For Loops", Code: exampleLoops},
		"conditionals": {Label: "Conditionals", Code: exampleConditionals},
		"maps":         {Label: "Maps", Code: exampleMaps},
	}
	writeJSON(w, http.StatusOK, examples)
}

type executeRequest struct {
	Code string `json:"code"`
}

type executeResponse struct {
	Output string `json:"output"`
	Error  string `json:"error,omitempty"`
}

func handleExecute(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, executeResponse{Error: "method not allowed"})
		return
	}

	var req executeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, executeResponse{Error: "invalid JSON body"})
		return
	}

	output, err := runRiceCode(req.Code)
	if err != nil {
		writeJSON(w, http.StatusOK, executeResponse{Output: output, Error: err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, executeResponse{Output: output})
}

func runRiceCode(code string) (string, error) {
	tokens, err := frontend.Tokenize(code)
	if err != nil {
		return "", fmt.Errorf("tokenization error: %w", err)
	}

	parser := frontend.NewParser(tokens)
	astTree := parser.Parse()

	if len(parser.Errors()) > 0 {
		var errs []string
		for i, e := range parser.Errors() {
			errs = append(errs, fmt.Sprintf("#%d: %v", i+1, e))
		}
		return "", fmt.Errorf("parse errors:\n%s", strings.Join(errs, "\n"))
	}

	var buf bytes.Buffer
	cfg := conf.NewDefaultEnvConfig()
	cfg.LoggingOutput = &buf

	it := exec.NewInterpreter(cfg)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	runCfg := conf.NewDefaultRunConfig()

	_, err = it.Interpret(ctx, astTree, runCfg)
	output := buf.String()

	if err != nil {
		var re exec.RuntimeError
		if errors.As(err, &re) {
			return output, fmt.Errorf("%s", re.Stacktrace())
		}
		return output, err
	}

	return output, nil
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}
