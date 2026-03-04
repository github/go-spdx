package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/github/go-spdx/v2/spdxexp"
	"github.com/spf13/cobra"
)

var filePath string

var rootCmd = &cobra.Command{
	Use:   "spdx-validate",
	Short: "Validate SPDX license expressions",
	Long: `spdx-validate reads SPDX license expressions and validates them.

By default it reads a single expression from stdin. Use -f/--file to read
a newline-separated list of expressions from a file.

Exits 0 if all expressions are valid, or 1 if any are invalid.

Examples:
  echo "MIT" | spdx-validate
  echo "Apache-2.0 OR MIT" | spdx-validate
  spdx-validate -f licenses.txt`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if filePath != "" {
			f, err := os.Open(filePath)
			if err != nil {
				return fmt.Errorf("unable to open file: %w", err)
			}
			defer f.Close()
			ok, err := validateExpressions(f, os.Stderr)
			if err != nil {
				return err
			}
			if !ok {
				os.Exit(1)
			}
			return nil
		}
		ok, err := validateSingleExpression(os.Stdin, os.Stderr)
		if err != nil {
			return err
		}
		if !ok {
			os.Exit(1)
		}
		return nil
	},
	SilenceUsage:  true,
	SilenceErrors: true,
}

func init() {
	rootCmd.Flags().StringVarP(&filePath, "file", "f", "", "path to a newline-separated file of SPDX expressions")
}

// validateSingleExpression reads one line from r, validates it as an SPDX
// expression, and writes an error message to w if invalid. Returns (true, nil)
// when valid, (false, nil) when invalid, or (false, err) on read errors.
func validateSingleExpression(r io.Reader, w io.Writer) (bool, error) {
	scanner := bufio.NewScanner(r)
	if !scanner.Scan() {
		return false, fmt.Errorf("no input provided")
	}
	input := strings.TrimSpace(scanner.Text())
	if input == "" {
		return false, fmt.Errorf("empty input")
	}

	valid, _ := spdxexp.ValidateLicenses([]string{input})
	if !valid {
		fmt.Fprintf(w, "invalid SPDX expression: %q\n", input)
		return false, nil
	}
	return true, nil
}

// validateExpressions reads newline-separated SPDX expressions from r,
// validates each one, and writes error messages to w for any that are invalid.
// Returns (true, nil) when all are valid, (false, nil) when any are invalid, or
// (false, err) on read errors or when no expressions are found.
func validateExpressions(r io.Reader, w io.Writer) (bool, error) {
	scanner := bufio.NewScanner(r)
	lineNum := 0
	failures := 0

	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		valid, _ := spdxexp.ValidateLicenses([]string{line})
		if !valid {
			failures++
			fmt.Fprintf(w, "line %d: invalid SPDX expression: %q\n", lineNum, line)
		}
	}

	if err := scanner.Err(); err != nil {
		return false, fmt.Errorf("error reading file: %w", err)
	}

	if lineNum == 0 || (lineNum > 0 && failures == lineNum) {
		return false, fmt.Errorf("no valid expressions found")
	}

	if failures > 0 {
		fmt.Fprintf(w, "%d of %d expressions failed validation\n", failures, lineNum)
		return false, nil
	}

	return true, nil
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
