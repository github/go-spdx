package spdxexp

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestActiveLicense(t *testing.T) {
	tests := []struct {
		name     string
		inputID  string
		outputID string
		result   bool
	}{
		{"active license - direct match", "Apache-2.0", "Apache-2.0", true},
		{"active license - all upper", "APACHE-2.0", "Apache-2.0", true},
		{"active license - all lower", "apache-2.0", "Apache-2.0", true},
		{"active license - mixed case", "apACHe-2.0", "Apache-2.0", true},
		{"deprecated license - direct match", "eCos-2.0", "eCos-2.0", false},
		{"deprecated license - all upper", "ECOS-2.0", "eCos-2.0", false},
		{"deprecated license - all lower", "ecos-2.0", "eCos-2.0", false},
		{"deprecated license - mixed case", "ECos-2.0", "eCos-2.0", false},
		{"exception license - direct match", "Bison-exception-2.2", "Bison-exception-2.2", false},
		{"exception license - all upper", "BISON-EXCEPTION-2.2", "Bison-exception-2.2", false},
		{"exception license - all lower", "bison-exception-2.2", "Bison-exception-2.2", false},
		{"exception license - mixed case", "BisoN-Exception-2.2", "Bison-exception-2.2", false},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			result, license := activeLicense((test.inputID))
			assert.Equal(t, test.result, result)
			if result {
				// updated to the proper case if found
				assert.Equal(t, test.outputID, license)
			} else {
				// no change in case if not found
				assert.Equal(t, test.inputID, license)
			}
		})
	}
}

func TestDeprecatedLicense(t *testing.T) {
	tests := []struct {
		name     string
		inputID  string
		outputID string
		result   bool
	}{
		{"active license - direct match", "Apache-2.0", "Apache-2.0", false},
		{"active license - all upper", "APACHE-2.0", "Apache-2.0", false},
		{"active license - all lower", "apache-2.0", "Apache-2.0", false},
		{"active license - mixed case", "apACHe-2.0", "Apache-2.0", false},
		{"deprecated license - direct match", "eCos-2.0", "eCos-2.0", true},
		{"deprecated license - all upper", "ECOS-2.0", "eCos-2.0", true},
		{"deprecated license - all lower", "ecos-2.0", "eCos-2.0", true},
		{"deprecated license - mixed case", "ECos-2.0", "eCos-2.0", true},
		{"exception license - direct match", "Bison-exception-2.2", "Bison-exception-2.2", false},
		{"exception license - all upper", "BISON-EXCEPTION-2.2", "Bison-exception-2.2", false},
		{"exception license - all lower", "bison-exception-2.2", "Bison-exception-2.2", false},
		{"exception license - mixed case", "BisoN-Exception-2.2", "Bison-exception-2.2", false},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			result, license := deprecatedLicense((test.inputID))
			assert.Equal(t, test.result, result)
			if result {
				// updated to the proper case if found
				assert.Equal(t, test.outputID, license)
			} else {
				// no change in case if not found
				assert.Equal(t, test.inputID, license)
			}
		})
	}
}

func TestExceptionLicense(t *testing.T) {
	tests := []struct {
		name     string
		inputID  string
		outputID string
		result   bool
	}{
		{"active license - direct match", "Apache-2.0", "Apache-2.0", false},
		{"active license - all upper", "APACHE-2.0", "Apache-2.0", false},
		{"active license - all lower", "apache-2.0", "Apache-2.0", false},
		{"active license - mixed case", "apACHe-2.0", "Apache-2.0", false},
		{"deprecated license - direct match", "eCos-2.0", "eCos-2.0", false},
		{"deprecated license - all upper", "ECOS-2.0", "eCos-2.0", false},
		{"deprecated license - all lower", "ecos-2.0", "eCos-2.0", false},
		{"deprecated license - mixed case", "ECos-2.0", "eCos-2.0", false},
		{"exception license - direct match", "Bison-exception-2.2", "Bison-exception-2.2", true},
		{"exception license - all upper", "BISON-EXCEPTION-2.2", "Bison-exception-2.2", true},
		{"exception license - all lower", "bison-exception-2.2", "Bison-exception-2.2", true},
		{"exception license - mixed case", "BisoN-Exception-2.2", "Bison-exception-2.2", true},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			result, license := exceptionLicense((test.inputID))
			assert.Equal(t, test.result, result)
			if result {
				// updated to the proper case if found
				assert.Equal(t, test.outputID, license)
			} else {
				// no change in case if not found
				assert.Equal(t, test.inputID, license)
			}
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
			location: map[uint8]int{licenseGroup: 22, versionGroup: 1, licenseIndex: 1}}},
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
