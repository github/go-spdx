package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBuildLicenseRanges(t *testing.T) {
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

	licenses := []License{
		{LicenseID: "Example-10.0"},
		{LicenseID: "Example-2.0-only"},
		{LicenseID: "Example-1.0-or-later"},
		{LicenseID: "Example-1.0"},
		{LicenseID: "Example-2.0"},
		{LicenseID: "Example-3.0-special"},
		{LicenseID: "Example-1.0-special"},
		{LicenseID: "Unrelated-2.0"},
	}

	assert.Equal(t, [][][]string{
		{{"Example-1.0"}, {"Example-2.0", "Example-2.0-only"}, {"Example-10.0"}},
		{{"Example-1.0-special"}, {"Example-2.0"}, {"Example-3.0-special"}, {"Example-10.0"}},
	}, buildLicenseRanges(licenses))
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

	assert.Equal(t, []licenseRangeFamily{
		{base: "Apache"},
		{base: "ASWF"},
		{base: "Example"},
		{base: "Special", suffix: "-Clause"},
	}, buildLicenseRangeFamilies(licenses, included, excluded))
}

func TestCompareLicenseVersions(t *testing.T) {
	assert.Negative(t, compareLicenseVersions("1.3a", "1.3c"))
	assert.Negative(t, compareLicenseVersions("2.0", "10.0"))
	assert.Zero(t, compareLicenseVersions("2.0", "2.0"))
}
