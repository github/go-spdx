package spdxexp

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestActiveLicense(t *testing.T) {
	tests := []struct {
		name   string
		id     string
		result bool
	}{
		{"active license", "Apache-2.0", true},
		{"deprecated license", "GFDL-1.3", false},
		{"exception license", "Bison-exception-2.2", false},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.result, activeLicense(test.id))
		})
	}
}

func TestDeprecatedLicense(t *testing.T) {
	tests := []struct {
		name   string
		id     string
		result bool
	}{
		{"active license", "Apache-2.0", false},
		{"deprecated license", "GFDL-1.3", true},
		{"exception license", "Bison-exception-2.2", false},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.result, deprecatedLicense(test.id))
		})
	}
}

func TestExceptionLicense(t *testing.T) {
	tests := []struct {
		name   string
		id     string
		result bool
	}{
		{"active license", "Apache-2.0", false},
		{"deprecated license", "GFDL-1.3", false},
		{"exception license", "Bison-exception-2.2", true},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.result, exceptionLicense(test.id))
		})
	}
}

func TestGetLicenseRange(t *testing.T) {
	tests := []struct {
		name         string
		id           string
		licenseRange *licenseRange
	}{
		{"no multi-element ranges", "Apache-2.0", &licenseRange{
			licenses: []string{"Apache-2.0"},
			location: map[uint8]int{licenseGroup: 2, versionGroup: 2, licenseIndex: 0}}},
		{"multi-element ranges", "GFDL-1.2-only", &licenseRange{
			licenses: []string{"GFDL-1.2", "GFDL-1.2-only"},
			location: map[uint8]int{licenseGroup: 18, versionGroup: 1, licenseIndex: 1}}},
		{"no range", "Bison-exception-2.2", nil},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			licenseRange := getLicenseRange(test.id)
			assert.Equal(t, test.licenseRange, licenseRange)
		})
	}
}
