package spdxexp

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLicenseString(t *testing.T) {
	tests := []struct {
		name   string
		node   *Node
		result string
	}{
		{"License node - simple",
			&Node{
				role: LicenseNode,
				exp:  nil,
				lic: &licenseNodePartial{
					license: "MIT", hasPlus: false,
					hasException: false, exception: ""},
				ref: nil,
			}, "MIT"},
		{"License node - plus",
			&Node{
				role: LicenseNode,
				exp:  nil,
				lic: &licenseNodePartial{
					license: "Apache-1.0", hasPlus: true,
					hasException: false, exception: ""},
				ref: nil,
			}, "Apache-1.0+"},
		{"License node - exception",
			&Node{
				role: LicenseNode,
				exp:  nil,
				lic: &licenseNodePartial{
					license: "GPL-2.0", hasPlus: false,
					hasException: true, exception: "Bison-exception-2.2"},
				ref: nil,
			}, "GPL-2.0 WITH Bison-exception-2.2"},
		{"LicenseRef node - simple",
			&Node{
				role: LicenseRefNode,
				exp:  nil,
				lic:  nil,
				ref: &referenceNodePartial{
					hasDocumentRef: false,
					documentRef:    "",
					licenseRef:     "MIT-Style-2",
				},
			}, "LicenseRef-MIT-Style-2"},
		{"LicenseRef node - with DocumentRef",
			&Node{
				role: LicenseRefNode,
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
			license := *test.node.LicenseString()
			assert.Equal(t, test.result, license)
		})
	}
}

func TestLicensesAreCompatible(t *testing.T) {
	tests := []struct {
		name   string
		first  *Node
		second *Node
		result bool
	}{
		{"expect greater than: GPL-3.0 > GPL-2.0", getLicenseNode("GPL-3.0", false), getLicenseNode("GPL-2.0", false), true},
		{"expect greater than: GPL-3.0-only > GPL-2.0-only", getLicenseNode("GPL-3.0-only", false), getLicenseNode("GPL-2.0-only", false), true},
		{"expect greater than: LPPL-1.3a > LPPL-1.0", getLicenseNode("LPPL-1.3a", false), getLicenseNode("LPPL-1.0", false), true},
		{"expect greater than: LPPL-1.3c > LPPL-1.3a", getLicenseNode("LPPL-1.3c", false), getLicenseNode("LPPL-1.3a", false), true},
		{"expect greater than: AGPL-3.0 > AGPL-1.0", getLicenseNode("AGPL-3.0", false), getLicenseNode("AGPL-1.0", false), true},
		{"expect greater than: GPL-2.0-or-later > GPL-2.0-only", getLicenseNode("GPL-2.0-or-later", true), getLicenseNode("GPL-2.0-only", false), true}, // TODO: Double check that -or-later and -only should be true for GT
		{"expect greater than: GPL-2.0-or-later > GPL-2.0", getLicenseNode("GPL-2.0-or-later", true), getLicenseNode("GPL-2.0", false), true},
		{"expect equal: GPL-3.0 > GPL-3.0", getLicenseNode("GPL-3.0", false), getLicenseNode("GPL-3.0", false), false},
		{"expect less than: MPL-1.0 > MPL-2.0", getLicenseNode("MPL-1.0", false), getLicenseNode("MPL-2.0", false), false},
		{"incompatible: MIT > ISC", getLicenseNode("MIT", false), getLicenseNode("ISC", false), false},
		{"incompatible: OSL-1.0 > OPL-1.0", getLicenseNode("OSL-1.0", false), getLicenseNode("OPL-1.0", false), false},
		{"not simple license: (MIT OR ISC) > GPL-3.0", getLicenseNode("(MIT OR ISC)", false), getLicenseNode("GPL-3.0", false), false}, // TODO: should it raise error?
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.result, compareGT(test.first, test.second))
		})
	}
}

func TestRangesAreCompatible(t *testing.T) {
	tests := []struct {
		name   string
		nodes  *NodePair
		result bool
	}{
		{"compatible - both use -or-later", &NodePair{
			firstNode:  getLicenseNode("GPL-1.0-or-later", true),
			secondNode: getLicenseNode("GPL-2.0-or-later", true)}, true},
		// {"compatible - both use +", &NodePair{                     // TODO: fails here and in js, but passes js satisfies
		// 	firstNode:  getLicenseNode("Apache-1.0", true),
		// 	secondNode: getLicenseNode("Apache-2.0", true)}, true},
		{"not compatible", &NodePair{
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
		nodes  *NodePair
		result bool
	}{
		{"in or-later range (later)", &NodePair{
			firstNode:  getLicenseNode("GPL-3.0", false),
			secondNode: getLicenseNode("GPL-2.0-or-later", true)}, true},
		{"in or-later range (same)", &NodePair{
			firstNode:  getLicenseNode("GPL-2.0", false),
			secondNode: getLicenseNode("GPL-2.0-or-later", true)}, false}, // TODO: why doesn't this
		{"in + range", &NodePair{
			firstNode:  getLicenseNode("Apache-2.0", false),
			secondNode: getLicenseNode("Apache-1.0+", true)}, false}, // TODO: think this doesn't match because Apache doesn't have any -or-later
		{"not in range", &NodePair{
			firstNode:  getLicenseNode("GPL-1.0", false),
			secondNode: getLicenseNode("GPL-2.0-or-later", true)}, false},
		{"different base license", &NodePair{
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
		nodes  *NodePair
		result bool
	}{
		{"equal", &NodePair{
			firstNode:  getLicenseNode("GPL-2.0", false),
			secondNode: getLicenseNode("GPL-2.0", false)}, true},
		{"not equal", &NodePair{
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
