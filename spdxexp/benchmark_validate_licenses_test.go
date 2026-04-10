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
	{"MIT", []string{"MIT"}},
	{"mit", []string{"mit"}},
	{"Apache-2.0", []string{"Apache-2.0"}},
	{"Zed", []string{"Zed"}},
	{"MIT AND Apache-2.0", []string{"MIT", "Apache-2.0"}},
	{"MIT AND Apache-2.0 OR Zed", []string{"MIT", "Apache-2.0", "Zed"}},
	{"GPL-2.0-or-later", []string{"GPL-2.0-or-later"}},
	{"GPL-2.0+", []string{"GPL-2.0+"}},
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
