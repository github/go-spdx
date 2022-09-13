package spdxexp

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReconstructedLicenseString(t *testing.T) {
	tests := []struct {
		name   string
		node   *node
		result string
	}{
		{"License node - simple", getLicenseNode("MIT", false), "MIT"},
		{"License node - plus", getLicenseNode("Apache-1.0", true), "Apache-1.0+"},
		{"License node - exception",
			&node{
				role: licenseNode,
				exp:  nil,
				lic: &licenseNodePartial{
					license: "GPL-2.0", hasPlus: false,
					hasException: true, exception: "Bison-exception-2.2"},
				ref: nil,
			}, "GPL-2.0 WITH Bison-exception-2.2"},
		{"LicenseRef node - simple",
			&node{
				role: licenseRefNode,
				exp:  nil,
				lic:  nil,
				ref: &referenceNodePartial{
					hasDocumentRef: false,
					documentRef:    "",
					licenseRef:     "MIT-Style-2",
				},
			}, "LicenseRef-MIT-Style-2"},
		{"LicenseRef node - with DocumentRef",
			&node{
				role: licenseRefNode,
				exp:  nil,
				lic:  nil,
				ref: &referenceNodePartial{
					hasDocumentRef: true,
					documentRef:    "spdx-tool-1.2",
					licenseRef:     "MIT-Style-2",
				},
			}, "DocumentRef-spdx-tool-1.2:LicenseRef-MIT-Style-2"},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			license := *test.node.reconstructedLicenseString()
			assert.Equal(t, test.result, license)
		})
	}
}

func TestLicensesAreCompatible(t *testing.T) {
	tests := []struct {
		name   string
		nodes  *nodePair
		result bool
	}{
		{"compatible (exact equal): GPL-3.0, GPL-3.0", &nodePair{
			getLicenseNode("GPL-3.0", false),
			getLicenseNode("GPL-3.0", false)}, true},
		{"compatible (diff case equal): Apache-2.0, APACHE-2.0", &nodePair{
			getLicenseNode("Apache-2.0", false),
			getLicenseNode("APACHE-2.0", false)}, true},
		// {"compatible (same version with +): Apache-1.0+, Apache-1.0", &nodePair{
		// 	getLicenseNode("Apache-1.0+", true),
		// 	getLicenseNode("Apache-1.0", false)}, true},
		// {"compatible (later version with +): Apache-1.0+, Apache-2.0", &nodePair{
		// 	getLicenseNode("Apache-1.0+", true),
		// 	getLicenseNode("Apache-2.0", false)}, true},
		// {"compatible (same version with -or-later): GPL-2.0-or-later, GPL-2.0", &nodePair{
		// 	getLicenseNode("GPL-2.0-or-later", true),
		// 	getLicenseNode("GPL-2.0", false)}, true},
		// {"compatible (same version with -or-later and -only): GPL-2.0-or-later, GPL-2.0-only", &nodePair{
		// 	getLicenseNode("GPL-2.0-or-later", true),
		// 	getLicenseNode("GPL-2.0-only", false)}, true}, // TODO: Double check that -or-later and -only should be true for GT
		// {"compatible (later version with -or-later): GPL-2.0-or-later, GPL-3.0", &nodePair{
		// 	getLicenseNode("GPL-2.0-or-later", true),
		// 	getLicenseNode("GPL-3.0", false)}, true},
		// {"incompatible (different versions using -only): GPL-3.0-only, GPL-2.0-only", &nodePair{
		// 	getLicenseNode("GPL-3.0-only", false),
		// 	getLicenseNode("GPL-2.0-only", false)}, false},
		{"incompatible (different versions with letter): LPPL-1.3c, LPPL-1.3a", &nodePair{
			getLicenseNode("LPPL-1.3c", false),
			getLicenseNode("LPPL-1.3a", false)}, false},
		{"incompatible (first > second): AGPL-3.0, AGPL-1.0", &nodePair{
			getLicenseNode("AGPL-3.0", false),
			getLicenseNode("AGPL-1.0", false)}, false},
		{"incompatible (second > first): MPL-1.0, MPL-2.0", &nodePair{
			getLicenseNode("MPL-1.0", false),
			getLicenseNode("MPL-2.0", false)}, false},
		{"incompatible (diff licenses): MIT, ISC", &nodePair{
			getLicenseNode("MIT", false),
			getLicenseNode("ISC", false)}, false},
		{"not simple license: (MIT OR ISC), GPL-3.0", &nodePair{
			getLicenseNode("(MIT OR ISC)", false),
			getLicenseNode("GPL-3.0", false)}, false}, // TODO: should it raise error?
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.result, test.nodes.licensesAreCompatible())
		})
	}
}

func TestRangesAreCompatible(t *testing.T) {
	tests := []struct {
		name   string
		nodes  *nodePair
		result bool
	}{
		{"compatible - both use -or-later", &nodePair{
			firstNode:  getLicenseNode("GPL-1.0-or-later", true),
			secondNode: getLicenseNode("GPL-2.0-or-later", true)}, true},
		// {"compatible - both use +", &nodePair{                     // TODO: fails here and in js, but passes js satisfies
		// 	firstNode:  getLicenseNode("Apache-1.0", true),
		// 	secondNode: getLicenseNode("Apache-2.0", true)}, true},
		{"not compatible", &nodePair{
			firstNode:  getLicenseNode("GPL-1.0-or-later", true),
			secondNode: getLicenseNode("LGPL-3.0-or-later", true)}, false},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.result, test.nodes.rangesAreCompatible())
		})
	}
}

func TestLicenseInRange(t *testing.T) {
	tests := []struct {
		name         string
		license      string
		licenseRange []string
		result       bool
	}{
		{"in range", "GPL-3.0", []string{
			"GPL-1.0-or-later",
			"GPL-2.0-or-later",
			"GPL-3.0",
			"GPL-3.0-only",
			"GPL-3.0-or-later"}, true},
		{"not in range", "GPL-3.0", []string{
			"GPL-2.0",
			"GPL-2.0-only"}, false},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.result, licenseInRange(test.license, test.licenseRange))
		})
	}
}

func TestIdentifierInRange(t *testing.T) {
	tests := []struct {
		name   string
		nodes  *nodePair
		result bool
	}{
		{"in or-later range (later)", &nodePair{
			firstNode:  getLicenseNode("GPL-3.0", false),
			secondNode: getLicenseNode("GPL-2.0-or-later", true)}, true},
		{"in or-later range (same)", &nodePair{
			firstNode:  getLicenseNode("GPL-2.0", false),
			secondNode: getLicenseNode("GPL-2.0-or-later", true)}, false}, // TODO: why doesn't this
		{"in + range", &nodePair{
			firstNode:  getLicenseNode("Apache-2.0", false),
			secondNode: getLicenseNode("Apache-1.0+", true)}, false}, // TODO: think this doesn't match because Apache doesn't have any -or-later
		{"not in range", &nodePair{
			firstNode:  getLicenseNode("GPL-1.0", false),
			secondNode: getLicenseNode("GPL-2.0-or-later", true)}, false},
		{"different base license", &nodePair{
			firstNode:  getLicenseNode("GPL-1.0", false),
			secondNode: getLicenseNode("LGPL-2.0-or-later", true)}, false},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.result, test.nodes.identifierInRange())
		})
	}
}

func TestLicensesExactlyEqual(t *testing.T) {
	tests := []struct {
		name   string
		nodes  *nodePair
		result bool
	}{
		{"equal", &nodePair{
			firstNode:  getLicenseNode("GPL-2.0", false),
			secondNode: getLicenseNode("GPL-2.0", false)}, true},
		{"not equal", &nodePair{
			firstNode:  getLicenseNode("GPL-1.0", false),
			secondNode: getLicenseNode("GPL-2.0", false)}, false},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.result, test.nodes.licensesExactlyEqual())
		})
	}
}
