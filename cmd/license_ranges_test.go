package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateLicenseRanges(t *testing.T) {
	configureExampleLicenseRangeFamilies(t)

	outputPath := filepath.Join(t.TempDir(), "license_ranges.go")
	licenses := []License{
		{LicenseID: "Example-10.0"},
		{LicenseID: "Example-2.0-only"},
		{LicenseID: "Example-1.0-or-later"},
		{LicenseID: "Example-1.0"},
		{LicenseID: "Example-2.0"},
		{LicenseID: "Example-3.0-special"},
		{LicenseID: "Example-1.0-special"},
		{LicenseID: "SingleVersion-2.0"},
	}

	require.NoError(t, generateLicenseRanges(licenses, outputPath))
	// #nosec G304 -- path is created within t.TempDir for this test.
	generated, err := os.ReadFile(outputPath)
	require.NoError(t, err)
	golden, err := os.ReadFile("testdata/license_ranges.golden")
	require.NoError(t, err)

	assert.Equal(t, string(golden), string(generated))
}

func TestBuildLicenseRanges(t *testing.T) {
	configureExampleLicenseRangeFamilies(t)

	licenses := []License{
		{LicenseID: "Example-10.0"},
		{LicenseID: "Example-2.0-only"},
		{LicenseID: "Example-1.0-or-later"},
		{LicenseID: "Example-1.0"},
		{LicenseID: "Example-2.0"},
		{LicenseID: "Example-3.0-special"},
		{LicenseID: "Example-1.0-special"},
		{LicenseID: "SingleVersion-2.0"},
	}

	licenseRanges := buildLicenseRanges(licenses)

	assert.Equal(t, [][][]string{
		{{"Example-1.0"}, {"Example-2.0", "Example-2.0-only"}, {"Example-10.0"}},
		{{"Example-1.0-special"}, {"Example-2.0"}, {"Example-3.0-special"}, {"Example-10.0"}},
	}, licenseRanges)

	assert.NotContains(t, licenseRanges, [][]string{{"SingleVersion-2.0"}})
}

func configureExampleLicenseRangeFamilies(t *testing.T) {
	previousIncluded := includedLicenseRangeFamilies
	previousExcluded := excludedLicenseRangeFamilyBases
	includedLicenseRangeFamilies = []licenseRangeFamily{
		{base: "Example"},
		{base: "Example", suffix: "-special", inheritBaseVersions: true},
	}
	excludedLicenseRangeFamilyBases = nil
	t.Cleanup(func() {
		includedLicenseRangeFamilies = previousIncluded
		excludedLicenseRangeFamilyBases = previousExcluded
	})
}

func TestBuildLicenseRangeFamilies(t *testing.T) {
	licenses := []License{
		{LicenseID: "Apache-1.0"},
		{LicenseID: "Apache-2.0"},
		{LicenseID: "ASWF-1.0"},
		{LicenseID: "ASWF-2.0"},
		{LicenseID: "Example-1.0"},
		{LicenseID: "Example-2.0"},
		{LicenseID: "Excluded-1.0"},
		{LicenseID: "Excluded-2.0"},
		{LicenseID: "Singleton-1.0"},
		{LicenseID: "Special-2-Clause"},
		{LicenseID: "Special-3-Clause"},
	}
	included := []licenseRangeFamily{{base: "Special", suffix: "-Clause"}}
	excluded := map[string]struct{}{"Excluded": {}}

	tests := []struct {
		name                 string
		licenses             []License
		included             []licenseRangeFamily
		excluded             map[string]struct{}
		licenseRangeFamilies []licenseRangeFamily
	}{
		{"no version", []License{{LicenseID: "Example"}}, nil, nil, []licenseRangeFamily{}},
		{"only one version", []License{{LicenseID: "Singleton-1.0"}}, nil, nil, []licenseRangeFamily{}},
		{"multiple versions", []License{{LicenseID: "Example-1.0"}, {LicenseID: "Example-2.0"}}, nil, nil, []licenseRangeFamily{{base: "Example"}}},
		{"included with suffix", []License{{LicenseID: "Special-2-Clause"}, {LicenseID: "Special-3-Clause"}}, included, nil, []licenseRangeFamily{{base: "Special", suffix: "-Clause"}}},
		{"excluded base", []License{{LicenseID: "Excluded-1.0"}, {LicenseID: "Excluded-2.0"}}, nil, excluded, []licenseRangeFamily{}},
		{"multiple families", licenses, included, excluded, []licenseRangeFamily{
			{base: "Apache"},
			{base: "ASWF"},
			{base: "Example"},
			{base: "Special", suffix: "-Clause"},
		}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			licenseRangeFamilies := buildLicenseRangeFamilies(test.licenses, test.included, test.excluded)
			assert.Equal(t, test.licenseRangeFamilies, licenseRangeFamilies)
		})
	}
}

func TestCompareLicenseVersions(t *testing.T) {
	assert.Negative(t, compareLicenseVersions("1.3a", "1.3c"))
	assert.Negative(t, compareLicenseVersions("2.0", "10.0"))
	assert.Zero(t, compareLicenseVersions("2.0", "2.0"))
}
