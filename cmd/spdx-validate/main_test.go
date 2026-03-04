package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// --- Tests for validateSingleExpression (stdin path) ---

func TestValidateSingleExpression_Valid(t *testing.T) {
	tests := []string{
		"MIT",
		"Apache-2.0",
		"BSD-3-Clause",
		"Apache-2.0 OR MIT",
		"MIT AND ISC",
		"GPL-3.0-only WITH Classpath-exception-2.0",
	}
	for _, expr := range tests {
		t.Run(expr, func(t *testing.T) {
			r := strings.NewReader(expr + "\n")
			var w bytes.Buffer
			ok, err := validateSingleExpression(r, &w)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !ok {
				t.Errorf("expected valid, got invalid; stderr: %s", w.String())
			}
			if w.Len() != 0 {
				t.Errorf("expected no stderr output, got: %s", w.String())
			}
		})
	}
}

func TestValidateSingleExpression_Invalid(t *testing.T) {
	tests := []string{
		"BOGUS-LICENSE",
		"NOT-A-REAL-ID",
		"MIT ANDOR Apache-2.0",
	}
	for _, expr := range tests {
		t.Run(expr, func(t *testing.T) {
			r := strings.NewReader(expr + "\n")
			var w bytes.Buffer
			ok, err := validateSingleExpression(r, &w)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if ok {
				t.Error("expected invalid, got valid")
			}
			if !strings.Contains(w.String(), "invalid SPDX expression") {
				t.Errorf("expected error message in output, got: %s", w.String())
			}
			if !strings.Contains(w.String(), expr) {
				t.Errorf("expected expression %q in output, got: %s", expr, w.String())
			}
		})
	}
}

func TestValidateSingleExpression_EmptyInput(t *testing.T) {
	r := strings.NewReader("")
	var w bytes.Buffer
	ok, err := validateSingleExpression(r, &w)
	if err == nil {
		t.Fatal("expected error for empty input, got nil")
	}
	if ok {
		t.Error("expected ok=false for empty input")
	}
	if !strings.Contains(err.Error(), "no input provided") {
		t.Errorf("expected 'no input provided' error, got: %v", err)
	}
}

func TestValidateSingleExpression_BlankLine(t *testing.T) {
	r := strings.NewReader("   \n")
	var w bytes.Buffer
	ok, err := validateSingleExpression(r, &w)
	if err == nil {
		t.Fatal("expected error for blank input, got nil")
	}
	if ok {
		t.Error("expected ok=false for blank input")
	}
	if !strings.Contains(err.Error(), "empty input") {
		t.Errorf("expected 'empty input' error, got: %v", err)
	}
}

// --- Tests for validateExpressions (file path) ---

func TestValidateExpressions_AllValid(t *testing.T) {
	input := "MIT\nApache-2.0\nBSD-3-Clause OR MIT\n"
	r := strings.NewReader(input)
	var w bytes.Buffer
	ok, err := validateExpressions(r, &w)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ok {
		t.Errorf("expected all valid, got invalid; stderr: %s", w.String())
	}
	if w.Len() != 0 {
		t.Errorf("expected no stderr output, got: %s", w.String())
	}
}

func TestValidateExpressions_SomeInvalid(t *testing.T) {
	input := "MIT\nNOT-A-LICENSE\nApache-2.0\nALSO-BOGUS\n"
	r := strings.NewReader(input)
	var w bytes.Buffer
	ok, err := validateExpressions(r, &w)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ok {
		t.Error("expected invalid result, got valid")
	}
	output := w.String()
	if !strings.Contains(output, `"NOT-A-LICENSE"`) {
		t.Errorf("expected NOT-A-LICENSE in output, got: %s", output)
	}
	if !strings.Contains(output, `"ALSO-BOGUS"`) {
		t.Errorf("expected ALSO-BOGUS in output, got: %s", output)
	}
	if !strings.Contains(output, "line 2:") {
		t.Errorf("expected 'line 2:' in output, got: %s", output)
	}
	if !strings.Contains(output, "line 4:") {
		t.Errorf("expected 'line 4:' in output, got: %s", output)
	}
	if !strings.Contains(output, "2 of 4 expressions failed") {
		t.Errorf("expected summary in output, got: %s", output)
	}
}

func TestValidateExpressions_AllInvalid(t *testing.T) {
	input := "BOGUS-1\nBOGUS-2\n"
	r := strings.NewReader(input)
	var w bytes.Buffer
	ok, err := validateExpressions(r, &w)
	if err == nil {
		t.Fatal("expected error when all expressions are invalid, got nil")
	}
	if ok {
		t.Error("expected ok=false")
	}
	if !strings.Contains(err.Error(), "no valid expressions found") {
		t.Errorf("expected 'no valid expressions found' error, got: %v", err)
	}
}

func TestValidateExpressions_EmptyFile(t *testing.T) {
	r := strings.NewReader("")
	var w bytes.Buffer
	ok, err := validateExpressions(r, &w)
	if err == nil {
		t.Fatal("expected error for empty file, got nil")
	}
	if ok {
		t.Error("expected ok=false for empty file")
	}
	if !strings.Contains(err.Error(), "no valid expressions found") {
		t.Errorf("expected 'no valid expressions found' error, got: %v", err)
	}
}

func TestValidateExpressions_SkipsBlankLines(t *testing.T) {
	input := "\nMIT\n\n\nApache-2.0\n\n"
	r := strings.NewReader(input)
	var w bytes.Buffer
	ok, err := validateExpressions(r, &w)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ok {
		t.Errorf("expected all valid, got invalid; stderr: %s", w.String())
	}
}

// --- Integration test using a temp file ---

func TestValidateExpressions_FromTempFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "licenses.txt")

	content := "MIT\nApache-2.0\nBSD-2-Clause\n"
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}

	f, err := os.Open(path)
	if err != nil {
		t.Fatalf("failed to open temp file: %v", err)
	}
	defer f.Close()

	var w bytes.Buffer
	ok, err := validateExpressions(f, &w)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ok {
		t.Errorf("expected all valid from file, got invalid; stderr: %s", w.String())
	}
}

func TestValidateExpressions_FromTempFileWithFailures(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "licenses.txt")

	content := "MIT\nINVALID-1\nApache-2.0\nINVALID-2\nBSD-2-Clause\n"
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}

	f, err := os.Open(path)
	if err != nil {
		t.Fatalf("failed to open temp file: %v", err)
	}
	defer f.Close()

	var w bytes.Buffer
	ok, err := validateExpressions(f, &w)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ok {
		t.Error("expected invalid result from file with bad entries")
	}
	output := w.String()
	if !strings.Contains(output, `"INVALID-1"`) {
		t.Errorf("expected INVALID-1 in output, got: %s", output)
	}
	if !strings.Contains(output, `"INVALID-2"`) {
		t.Errorf("expected INVALID-2 in output, got: %s", output)
	}
	if !strings.Contains(output, "2 of 5 expressions failed") {
		t.Errorf("expected summary in output, got: %s", output)
	}
}
