package spdxexp

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExtractLicenses(t *testing.T) {
	tests := []struct {
		name              string
		inputExpression   string
		extractedLicenses []string
	}{
		{"Single license", "MIT", []string{"MIT"}},
		{"AND'ed licenses", "MIT AND Apache-2.0", []string{"MIT", "Apache-2.0"}},
		{"AND'ed & OR'ed licenses", "(MIT AND Apache-2.0) OR GPL-3.0", []string{"GPL-3.0", "MIT", "Apache-2.0"}},
		{"ONLY modifiers", "LGPL-2.1-only OR MIT OR BSD-3-Clause", []string{"MIT", "BSD-3-Clause", "LGPL-2.1-only"}},
		{"WITH modifiers", "GPL-2.0-or-later WITH Bison-exception-2.2", []string{"GPL-2.0-or-later+ WITH Bison-exception-2.2"}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			licenses, err := ExtractLicenses(test.inputExpression)
			assert.ElementsMatch(t, test.extractedLicenses, licenses)
			assert.NoError(t, err)
		})
	}
}
