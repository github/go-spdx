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
		{"active license - direct match", "Apache-2.0", true},
		{"active license - all upper", "APACHE-2.0", true},
		{"active license - all lower", "apache-2.0", true},
		{"active license - mixed case", "apACHe-2.0", true},
		{"deprecated license - direct match", "eCos-2.0", false},
		{"deprecated license - all upper", "ECOS-2.0", false},
		{"deprecated license - all lower", "ecos-2.0", false},
		{"deprecated license - mixed case", "ECos-2.0", false},
		{"exception license - direct match", "Bison-exception-2.2", false},
		{"exception license - all upper", "BISON-EXCEPTION-2.2", false},
		{"exception license - all lower", "bison-exception-2.2", false},
		{"exception license - mixed case", "BisoN-Exception-2.2", false},
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
		{"active license - direct match", "Apache-2.0", false},
		{"active license - all upper", "APACHE-2.0", false},
		{"active license - all lower", "apache-2.0", false},
		{"active license - mixed case", "apACHe-2.0", false},
		{"deprecated license - direct match", "eCos-2.0", true},
		{"deprecated license - all upper", "ECOS-2.0", true},
		{"deprecated license - all lower", "ecos-2.0", true},
		{"deprecated license - mixed case", "ECos-2.0", true},
		{"exception license - direct match", "Bison-exception-2.2", false},
		{"exception license - all upper", "BISON-EXCEPTION-2.2", false},
		{"exception license - all lower", "bison-exception-2.2", false},
		{"exception license - mixed case", "BisoN-Exception-2.2", false},
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
		{"active license - direct match", "Apache-2.0", false},
		{"active license - all upper", "APACHE-2.0", false},
		{"active license - all lower", "apache-2.0", false},
		{"active license - mixed case", "apACHe-2.0", false},
		{"deprecated license - direct match", "eCos-2.0", false},
		{"deprecated license - all upper", "ECOS-2.0", false},
		{"deprecated license - all lower", "ecos-2.0", false},
		{"deprecated license - mixed case", "ECos-2.0", false},
		{"exception license - direct match", "Bison-exception-2.2", true},
		{"exception license - all upper", "BISON-EXCEPTION-2.2", true},
		{"exception license - all lower", "bison-exception-2.2", true},
		{"exception license - mixed case", "BisoN-Exception-2.2", true},
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
