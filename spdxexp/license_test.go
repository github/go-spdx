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
			assert.Equal(t, test.result, ActiveLicense(test.id))
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
			assert.Equal(t, test.result, DeprecatedLicense(test.id))
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
			assert.Equal(t, test.result, ExceptionLicense(test.id))
		})
	}
}

func TestGetLicenseRange(t *testing.T) {
	tests := []struct {
		name         string
		id           string
		licenseRange *LicenseRange
	}{
		{"no multi-element ranges", "Apache-2.0", &LicenseRange{
			Licenses: []string{"Apache-2.0"},
			Location: map[uint8]int{LICENSE_GROUP: 2, VERSION_GROUP: 2, LICENSE_INDEX: 0}}},
		{"multi-element ranges", "GFDL-1.2-only", &LicenseRange{
			Licenses: []string{"GFDL-1.2", "GFDL-1.2-only"},
			Location: map[uint8]int{LICENSE_GROUP: 18, VERSION_GROUP: 1, LICENSE_INDEX: 1}}},
		{"no range", "Bison-exception-2.2", nil},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			licenseRange := GetLicenseRange(test.id)
			assert.Equal(t, test.licenseRange, licenseRange)
		})
	}
}
