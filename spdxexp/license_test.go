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

func TestLicenseRange(t *testing.T) {
	tests := []struct {
		name   string
		id     string
		result []string
	}{
		{"single range", "Apache-2.0",
			[]string{"Apache-1.0", "Apache-1.1", "Apache-2.0"}},
		{"multiple ranges", "GFDL-1.2-only",
			[]string{"GFDL-1.2", "GFDL-1.2-only"}},
		{"no range", "Bison-exception-2.2",
			[]string{"Bison-exception-2.2"}}, // TODO: should this return empty array?
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.result, LicenseRange(test.id))
		})
	}
}
