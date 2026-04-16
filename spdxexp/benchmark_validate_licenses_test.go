package spdxexp

import (
	"fmt"
	"testing"
)

type validateLicensesBenchmarkScenario struct {
	name         string
	testLicenses []string
}

var validateLicensesBenchmarkScenarios = []validateLicensesBenchmarkScenario{
	// Scenario order is used as-is in the summary table.
	{"MIT--exact", []string{"MIT"}},
	{"mit--caseinsensitive", []string{" mit  "}},
	{"Apache-2.0--active-early", []string{"Apache-2.0"}},
	{"Zed--active-end", []string{"Zed"}},
	{"MIT AND Apache-2.0--complex", []string{"MIT", "Apache-2.0"}},
	{"MIT AND Apache-2.0 OR Zed--complex", []string{"MIT", "Apache-2.0", "Zed"}},
	{"BSD-2-Clause-FreeBSD--deprecated", []string{"BSD-2-Clause-FreeBSD"}},
	{"GPL-2.0-or-later--range", []string{"GPL-2.0-or-later"}},
	{"Apache-1.0+--plus-range", []string{"Apache-1.0+"}},
	{"LicenseRef-scancode-adobe-postscript", []string{"LicenseRef-scancode-adobe-postscript"}},
	{"DocumentRef-spdx-tool-1.2:LicenseRef-MIT-Style-2", []string{"DocumentRef-spdx-tool-1.2:LicenseRef-MIT-Style-2"}},
}

func BenchmarkValidateLicenses(b *testing.B) {
	for _, scenario := range validateLicensesBenchmarkScenarios {
		scenario := scenario
		b.Run(scenario.name, func(b *testing.B) {
			benchmarkValidateLicensesScenario(b, scenario.testLicenses)
		})
	}
}

func benchmarkValidateLicensesScenario(b *testing.B, licenses []string) {
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		valid, invalidLicenses := ValidateLicenses(licenses)
		if !valid || len(invalidLicenses) != 0 {
			b.Fatalf("ValidateLicenses(%v) returned valid=%v invalid=%v", licenses, valid, invalidLicenses)
		}
	}
}

func computeValidateLicensesBenchmarkTableRows(repeats int) []benchmarkTableRow {
	rows := make([]benchmarkTableRow, 0, len(validateLicensesBenchmarkScenarios))

	for _, scenario := range validateLicensesBenchmarkScenarios {
		scenario := scenario
		avg := runBenchmarkNsAvg(repeats, func(b *testing.B) {
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				valid, invalidLicenses := ValidateLicenses(scenario.testLicenses)
				if !valid || len(invalidLicenses) != 0 {
					panic(fmt.Sprintf("ValidateLicenses scenario %q failed: valid=%v invalid=%v", scenario.name, valid, invalidLicenses))
				}
			}
		})
		rows = append(rows, benchmarkTableRow{label: scenario.name, nsOpAvg: avg})
	}

	return rows
}
