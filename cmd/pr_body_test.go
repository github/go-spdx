package main

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBuildPRBody(t *testing.T) {
	previousLicenses := LicenseData{
		Version: "3.28.0",
		Licenses: []License{
			{LicenseID: "Example-1.0"},
			{LicenseID: "Removed-1.0"},
			{LicenseID: "Now-Deprecated"},
			{LicenseID: "Reactivated", IsDeprecated: true},
		},
	}
	currentLicenses := LicenseData{
		Version: "3.29.0",
		Licenses: []License{
			{LicenseID: "Example-1.0"},
			{LicenseID: "Example-2.0"},
			{LicenseID: "Now-Deprecated", IsDeprecated: true},
			{LicenseID: "Reactivated"},
		},
	}
	previousExceptions := ExceptionData{Exceptions: []Exception{{LicenseID: "Old-exception"}}}
	currentExceptions := ExceptionData{Exceptions: []Exception{{LicenseID: "New-exception"}}}

	body := buildPRBody(previousLicenses, currentLicenses, previousExceptions, currentExceptions)

	assert.True(t, strings.HasPrefix(body, "> [!WARNING]\n"))
	assert.Contains(t, body, "open a stacked PR with `auto-update-licenses` as its base")
	assert.Contains(t, body, "- **Previous SPDX License List Version:** `3.28.0`")
	assert.Contains(t, body, "- **New SPDX License List Version:** `3.29.0`")
	assert.Contains(t, body, "- **Added:** `Example-2.0`")
	assert.Contains(t, body, "- **Removed:** `Removed-1.0`")
	assert.Contains(t, body, "- **Added:** `New-exception`")
	assert.Contains(t, body, "- **Removed:** `Old-exception`")
	assert.Contains(t, body, "- **Newly deprecated licenses:** `Now-Deprecated`")
	assert.Contains(t, body, "- **Reactivated licenses:** `Reactivated`")
	assert.Contains(t, body, "- **Added:** `Example-<version>`")
	assert.Contains(t, body, "Human review required: license range family changed")
}
