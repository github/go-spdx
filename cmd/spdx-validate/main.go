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
	Long: `spdx-validate reads newline-separated SPDX license expressions and validates them.

It reads from stdin by default, or from a file specified with -f/--file.
Blank lines are skipped. Exits 0 if all expressions are valid, or 1 if any
are invalid.

Examples:
  echo "MIT" | spdx-validate
  printf "MIT\nApache-2.0\n" | spdx-validate
  spdx-validate -f licenses.txt`,
	RunE: func(cmd *cobra.Command, args []string) error {
		var r io.Reader = os.Stdin
		if filePath != "" {
			f, err := os.Open(filePath)
			if err != nil {
				return fmt.Errorf("unable to open file: %w", err)
			}
			defer f.Close()
			r = f
		}
		ok, err := validateExpressions(r, os.Stderr)
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
