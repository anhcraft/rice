package frontend

import (
	"errors"
	"rice/lib/set"
	"strings"
	"testing"
)

type parserTestCase struct {
	Name              string   `json:"name"`
	Input             string   `json:"input"`
	ExpectedOutput    []string `json:"expectedOutput"`
	ExpectedErrorName string   `json:"expectedErrorName"`
}

var errorNameToError = map[string]error{
	"expectIdErr":                      expectIdErr,
	"expectEqualRhsErr":                expectEqualRhsErr,
	"expectReferenceErr":               expectReferenceErr,
	"expectValueErr":                   expectValueErr,
	"expectValueLparenForErr":          expectValueLparenForErr,
	"expectPartSeparatorErr":           expectPartSeparatorErr,
	"expectValuePartSeparatorErr":      expectValuePartSeparatorErr,
	"expectSimpleStmtPartSeparatorErr": expectSimpleStmtPartSeparatorErr,
	"expectRbracketElemAccessErr":      expectRbracketElemAccessErr,
	"expectRparenGroupErr":             expectRparenGroupErr,
	"expectSimpleStmtRparenErr":        expectSimpleStmtRparenErr,
	"expectRparenForClauseErr":         expectRparenForClauseErr,
	"expectValueRparenCallArgsErr":     expectValueRparenCallArgsErr,
	"expectCommaRparenCallArgsErr":     expectCommaRparenCallArgsErr,
	"expectIdRparenFuncParamsErr":      expectIdRparenFuncParamsErr,
	"expectCommaRparenFuncParamsErr":   expectCommaRparenFuncParamsErr,
	"expectLbraceBlockErr":             expectLbraceBlockErr,
	"expectRbraceBlockErr":             expectRbraceBlockErr,
	"expectLbraceOrIfBranchErr":        expectLbraceOrIfBranchErr,
	"expectLparenFuncErr":              expectLparenFuncErr,
	"expectStmtTerminatorErr":          expectStmtTerminatorErr,
	"genericUnexpectedTokenErr":        genericUnexpectedTokenErr,
	"excessiveIfDepthErr":              excessiveIfDepthErr,
	"excessiveCallArgsSizeErr":         excessiveCallArgsSizeErr,
	"excessiveFuncParamsSizeErr":       excessiveFuncParamsSizeErr,
	"invalidAssignmentTargetErr":       invalidAssignmentTargetErr,
}

func TestParser(t *testing.T) {
	testCases := loadTestCases[parserTestCase](t, "parser-parse.json")

	errorFound := set.NewSet[string]()

	for _, v := range testCases {
		errorFound.Add(v.ExpectedErrorName)
	}

	for err := range errorNameToError {
		if !errorFound.Has(err) {
			t.Logf("missing test case for error %q", err)
		}
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			tokens, err := Tokenize(tc.Input)
			if err != nil {
				t.Errorf("Tokenize() error = %v", err)
			}

			parser := NewParser(tokens)

			expressions := parser.Parse()

			hasError := len(parser.errors) > 0
			expectsError := tc.ExpectedErrorName != ""

			if expectsError {
				if !hasError {
					t.Fatalf("Expected error '%s' but got none.", tc.ExpectedErrorName)
				}

				expectedErr, ok := errorNameToError[tc.ExpectedErrorName]
				if !ok {
					t.Fatalf("Unknown error name in test case: '%s'", tc.ExpectedErrorName)
				}

				actualErr := parser.errors[0]
				if !errors.Is(actualErr, expectedErr) {
					t.Errorf("Expected error type '%v' but got '%v'", expectedErr, actualErr)
				}
			} else {
				if hasError {
					t.Fatalf("Expected no errors, but got %d: %v", len(parser.errors), parser.errors[0])
				}
				if expressions == nil && len(tc.ExpectedOutput) > 0 {
					t.Fatal("Expected expressions but parser returned nil")
				}

				var actualOutput []string
				for _, expr := range expressions {
					actualOutput = append(actualOutput, expr.String())
				}

				expectedStr := strings.Join(tc.ExpectedOutput, "\n")
				actualStr := strings.Join(actualOutput, "\n")

				if expectedStr != actualStr {
					t.Errorf("AST mismatch:\n\n--- Expected ---\n%s\n\n--- Actual ---\n%s\n", expectedStr, actualStr)
				}
			}
		})
	}
}
