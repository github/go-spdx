package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBuildLicenseRanges(t *testing.T) {
	previousFamilies := licenseRangeFamilies
	licenseRangeFamilies = []licenseRangeFamily{
		{base: "Example"},
		{base: "Example", suffix: "-special", inheritBaseVersions: true},
	}
	t.Cleanup(func() { licenseRangeFamilies = previousFamilies })

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

func TestCompareLicenseVersions(t *testing.T) {
	assert.Negative(t, compareLicenseVersions("1.3a", "1.3c"))
	assert.Negative(t, compareLicenseVersions("2.0", "10.0"))
	assert.Zero(t, compareLicenseVersions("2.0", "2.0"))
}
