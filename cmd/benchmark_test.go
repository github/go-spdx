package main

import (
	"fmt"
	"math"
	"os"
	"strings"
	"testing"

	"github.com/github/go-spdx/v2/spdxexp"
)

func TestMain(m *testing.M) {
	fmt.Fprintln(os.Stdout, "Benchmark output columns (Go 'go test -bench'):")
	fmt.Fprintln(os.Stdout, "- BenchmarkName-<GOMAXPROCS>: which benchmark ran and with how many OS threads")
	fmt.Fprintln(os.Stdout, "- iters: number of iterations (b.N) executed")
	fmt.Fprintln(os.Stdout, "- ns/op: average time per iteration")
	fmt.Fprintln(os.Stdout, "- B/op: bytes allocated per iteration (shown with -benchmem)")
	fmt.Fprintln(os.Stdout, "- allocs/op: allocations per iteration (shown with -benchmem)")
	fmt.Fprintln(os.Stdout, "")

	code := m.Run()

	// Compute an observed relative scale factor using the benchmark functions.
	// This is separate from the `go test -bench ...` results (which are printed
	// above), but it gives a concrete, machine-specific ratio to show at a glance.
	eq := testing.Benchmark(BenchmarkStringEqualityMIT)
	val := testing.Benchmark(BenchmarkValidateLicensesMIT)

	// Prefer a floating-point ns/op average for display so sub-nanosecond results
	// don't get rounded to 0.
	eqNsAvg := 0.0
	valNsAvg := 0.0
	if eq.N > 0 {
		eqNsAvg = float64(eq.T.Nanoseconds()) / float64(eq.N)
	}
	if val.N > 0 {
		valNsAvg = float64(val.T.Nanoseconds()) / float64(val.N)
	}
	formatNsAvg := func(ns float64) string {
		if ns < 10 {
			return fmt.Sprintf("~%.1f ns/op", ns)
		}
		rounded := int64(math.Round(ns))
		return fmt.Sprintf("~%s ns/op", formatWithCommas(rounded))
	}
	formatScale := func(val, baseline float64) string {
		if baseline <= 0 {
			return "n/a"
		}
		ratio := val / baseline
		if ratio < 1 {
			ratio = 1
		}

		// Round to 2 significant digits to match the practical precision of these
		// measurements (e.g. 9597 -> 9600, 843 -> 840).
		rounded := 0.0
		if ratio >= 10 {
			magnitude := math.Pow(10, math.Floor(math.Log10(ratio))-1) // keep 2 sig digits
			rounded = math.Round(ratio/magnitude) * magnitude
		} else {
			// For very small ratios, keep a single decimal place.
			rounded = math.Round(ratio*10) / 10
		}

		if rounded >= 10 {
			return fmt.Sprintf("~%sx", formatWithCommas(int64(rounded)))
		}
		if rounded == math.Trunc(rounded) {
			return fmt.Sprintf("~%dx", int64(rounded))
		}
		return fmt.Sprintf("~%.1fx", rounded)
	}
	nsOpEq := formatNsAvg(eqNsAvg)
	nsOpVal := formatNsAvg(valNsAvg)
	scaleVal := formatScale(valNsAvg, eqNsAvg)

	fmt.Fprintln(os.Stdout, "\nScalability summary (at a glance)")

	col1 := 28
	col2 := 14
	col3 := 44

	line := func() {
		fmt.Fprintf(os.Stdout, "+-%s-+-%s-+-%s-+\n", strings.Repeat("-", col1), strings.Repeat("-", col2), strings.Repeat("-", col3))
	}
	row := func(c1, c2, c3 string) {
		fmt.Fprintf(os.Stdout, "| %-*s | %-*s | %-*s |\n", col1, c1, col2, c2, col3, c3)
	}

	line()
	row("Characteristic", "MIT==MIT", "ValidateLicenses([\"MIT\"]) ")
	line()
	row("ns/op average", nsOpEq, nsOpVal)
	row("Scale", "1x", scaleVal)
	row("Time per check", "O(1)", "~O(M*L) (parse each of M licenses)")
	row("Memory per check", "O(1)", "~O(M) allocs (see B/op, allocs/op)")
	line()
	fmt.Fprintln(os.Stdout, "")
	fmt.Fprintln(os.Stdout, "Measurement tip: for strict comparisons, keep ops/run equal (-benchtime=1000x) and increase repeats (-count=10+) then compare with benchstat.")

	os.Exit(code)
}

func formatWithCommas(n int64) string {
	s := fmt.Sprintf("%d", n)
	if len(s) <= 3 {
		return s
	}

	var b strings.Builder
	pre := len(s) % 3
	if pre == 0 {
		pre = 3
	}
	b.WriteString(s[:pre])
	for i := pre; i < len(s); i += 3 {
		b.WriteByte(',')
		b.WriteString(s[i : i+3])
	}
	return b.String()
}

// Benchmark summary (scalability-focused)
//
// BenchmarkStringEqualityMIT measures a constant-time operation: comparing two
// already-in-memory short string literals. This is O(1) time, ~0 allocations,
// and scales linearly only with how many comparisons you do.
//
// BenchmarkValidateLicensesMIT measures SPDX license validation via parsing.
// Even for a single license, this is substantially heavier because it creates
// parser structures and does work proportional to the license string length.
//
// Scalability implications:
//   - If you validate M licenses, ValidateLicenses is ~O(M) calls to parse(), so
//     total cost grows roughly linearly with M (and with average string length).
//   - If license strings are expressions, runtime also scales with expression
//     complexity (more tokens/nodes) and may allocate more.
//   - The string equality baseline stays near O(1) per comparison with minimal
//     memory traffic.
//
// In practice, for “at scale” validation (large M, long expressions, repeated
// checks), the dominant lever is avoiding repeated parsing (e.g., parse once and
// reuse/caching parsed nodes) rather than micro-optimizing string comparisons.
func BenchmarkValidateLicensesMIT(b *testing.B) {
	b.ReportAllocs()

	licenses := []string{"MIT"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		valid, invalid := spdxexp.ValidateLicenses(licenses)
		if !valid || len(invalid) != 0 {
			b.Fatalf("expected MIT to be valid; valid=%v invalid=%v", valid, invalid)
		}
	}
}

func BenchmarkStringEqualityMIT(b *testing.B) {
	b.ReportAllocs()

	v1 := "MIT"
	v2 := "MIT"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if v1 != v2 {
			b.Fatal("unexpected string inequality")
		}
	}
}
