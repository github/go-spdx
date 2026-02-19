package spdxexp

import (
	"flag"
	"fmt"
	"math"
	"os"
	"strings"
	"testing"
	"time"
)

const benchmarkRepeatsForSummary = 4

// Fixed baselines for the Scale column in the summary tables.
// Using constants makes the scale values comparable across runs/branches.
const (
	benchmarkScaleBaselineValidateLicensesNsOp = 5.0
	benchmarkScaleBaselineSatisfiesNsOp        = 1500.0
)

func TestMain(m *testing.M) {
	// When TestMain is present, it's safest to explicitly parse flags before
	// inspecting any -test.* settings.
	if !flag.Parsed() {
		flag.Parse()
	}

	benchPattern := ""
	benchFlag := flag.Lookup("test.bench")
	if benchFlag != nil {
		benchPattern = benchFlag.Value.String()
	}

	shouldPrintBenchOutput := benchPattern != ""
	if shouldPrintBenchOutput {
		// Benchmarks are executed as part of summary table generation (via
		// testing.Benchmark). Suppress the default go test benchmark execution so
		// we don't run benchmarks twice in a single invocation.
		if benchFlag != nil {
			_ = benchFlag.Value.Set("$^")
		}

		fmt.Fprintln(os.Stdout, "Benchmark summary tables:")
		fmt.Fprintln(os.Stdout, "- ns/op average: average time per operation")
		fmt.Fprintln(os.Stdout, "- Scale: relative to a fixed baseline per table")
		fmt.Fprintln(os.Stdout, "")
	}

	code := m.Run()

	if shouldPrintBenchOutput {
		validateRows := withScaleColumn(computeValidateLicensesBenchmarkTableRows(benchmarkRepeatsForSummary), benchmarkScaleBaselineValidateLicensesNsOp)
		printBenchmarkTable(os.Stdout, "Benchmark ValidateLicenses", validateRows, benchmarkScaleBaselineValidateLicensesNsOp)

		satisfiesRows := withScaleColumn(computeSatisfiesBenchmarkTableRows(benchmarkRepeatsForSummary), benchmarkScaleBaselineSatisfiesNsOp)
		printBenchmarkTable(os.Stdout, "Benchmark Satisfies", satisfiesRows, benchmarkScaleBaselineSatisfiesNsOp)
	}

	os.Exit(code)
}

type benchmarkTableRow struct {
	label   string
	nsOpAvg float64
	scale   string
}

func withScaleColumn(rows []benchmarkTableRow, benchmarkScaleBaselineNsOp float64) []benchmarkTableRow {
	if len(rows) == 0 {
		return rows
	}

	for i := range rows {
		rows[i].scale = formatScale(rows[i].nsOpAvg, benchmarkScaleBaselineNsOp)
	}
	return rows
}

func formatScale(ns, baseline float64) string {
	if ns <= 0 || baseline <= 0 {
		return "n/a"
	}

	ratio := ns / baseline
	if ratio >= 0.95 && ratio <= 1.05 {
		return "1x"
	}

	if ratio >= 10 {
		return fmt.Sprintf("~%sx", formatWithCommas(int64(math.Round(ratio))))
	}

	return fmt.Sprintf("~%.1fx", math.Round(ratio*10)/10)
}

func runBenchmarkNsAvg(repeats int, fn func(b *testing.B)) float64 {
	if repeats <= 0 {
		repeats = 1
	}

	sum := 0.0
	count := 0
	for i := 0; i < repeats; i++ {
		res := testing.Benchmark(fn)
		if res.N <= 0 {
			continue
		}
		ns := float64(res.T.Nanoseconds()) / float64(res.N)
		sum += ns
		count++
	}
	if count == 0 {
		return 0
	}
	return sum / float64(count)
}

func printBenchmarkTable(w *os.File, title string, rows []benchmarkTableRow, benchmarkScaleBaselineNsOp float64) {
	header1 := title
	header2 := "ns/op average"
	header3 := fmt.Sprintf("Scale (%dns/op=1x)", int(benchmarkScaleBaselineNsOp))

	col1 := len(header1)
	for _, r := range rows {
		if len(r.label) > col1 {
			col1 = len(r.label)
		}
	}

	formatNsAvg := func(r benchmarkTableRow) string {
		num := nsNumberString(r.nsOpAvg)
		return fmt.Sprintf("~%s ns/op", num)
	}

	col2 := len(header2)
	for _, r := range rows {
		if l := len(formatNsAvg(r)); l > col2 {
			col2 = l
		}
	}

	col3 := len(header3)
	for _, r := range rows {
		if len(r.scale) > col3 {
			col3 = len(r.scale)
		}
	}

	line := func() {
		fmt.Fprintf(w, "+-%s-+-%s-+-%s-+\n", strings.Repeat("-", col1), strings.Repeat("-", col2), strings.Repeat("-", col3))
	}
	row := func(c1, c2, c3 string) {
		fmt.Fprintf(w, "| %-*s | %-*s | %-*s |\n", col1, c1, col2, c2, col3, c3)
	}

	line()
	row(header1, header2, header3)
	line()
	for _, r := range rows {
		ns := formatNsAvg(r)
		row(r.label, ns, r.scale)
	}
	line()
	fmt.Fprintln(w, "")
}

func nsNumberString(ns float64) string {
	if ns <= 0 {
		return "0"
	}
	if ns < float64(10*time.Microsecond.Nanoseconds()) {
		return fmt.Sprintf("%.1f", ns)
	}
	return formatWithCommas(int64(ns + 0.5))
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
