package spdxexp

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Regression for #165: expandOrTerm used to drop LicenseRef leaves, so an OR
// alternative that was only a LicenseRef never participated in Satisfies.
func TestSatisfiesLicenseRefOR(t *testing.T) {
	cases := []struct {
		name     string
		expr     string
		allowed  []string
		want     bool
	}{
		{
			name:    "LicenseRef alternative of OR is allowed",
			expr:    "MIT OR LicenseRef-x",
			allowed: []string{"LicenseRef-x"},
			want:    true,
		},
		{
			name:    "LicenseRef first in OR is allowed",
			expr:    "LicenseRef-x OR MIT",
			allowed: []string{"LicenseRef-x"},
			want:    true,
		},
		{
			name:    "list-id OR still works",
			expr:    "MIT OR ISC",
			allowed: []string{"ISC"},
			want:    true,
		},
		{
			name:    "AND requiring a LicenseRef OR does not pass on MIT alone",
			expr:    "MIT AND (LicenseRef-a OR LicenseRef-b)",
			allowed: []string{"MIT"},
			want:    false,
		},
		{
			name:    "AND with LicenseRef OR passes when both sides allowed",
			expr:    "MIT AND (LicenseRef-a OR LicenseRef-b)",
			allowed: []string{"MIT", "LicenseRef-a"},
			want:    true,
		},
		{
			name:    "AND with list-id OR still rejects MIT alone",
			expr:    "MIT AND (ISC OR BSD-3-Clause)",
			allowed: []string{"MIT"},
			want:    false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := Satisfies(tc.expr, tc.allowed)
			require.NoError(t, err)
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestExpandOrKeepsLicenseRef(t *testing.T) {
	node, err := parse("MIT OR LicenseRef-x")
	require.NoError(t, err)
	expanded := node.expand(true)
	require.Len(t, expanded, 2)

	seenRef, seenMIT := false, false
	for _, part := range expanded {
		require.Len(t, part, 1)
		if part[0].isLicenseRef() {
			seenRef = true
			assert.Equal(t, "x", *part[0].licenseRef())
		}
		if part[0].isLicense() && part[0].string() == "MIT" {
			seenMIT = true
		}
	}
	assert.True(t, seenRef, "LicenseRef-x missing from expand")
	assert.True(t, seenMIT, "MIT missing from expand")
}
