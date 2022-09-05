package spdxexp

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCompareGT(t *testing.T) {
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

func TestCompareEQ(t *testing.T) {
	tests := []struct {
		name   string
		first  *Node
		second *Node
		result bool
	}{
		{"expect greater than: GPL-3.0 == GPL-2.0", getLicenseNode("GPL-3.0", false), getLicenseNode("GPL-2.0", false), false},
		{"expect greater than: GPL-3.0-only == GPL-2.0-only", getLicenseNode("GPL-3.0-only", false), getLicenseNode("GPL-2.0-only", false), false},
		{"expect greater than: LPPL-1.3a == LPPL-1.0", getLicenseNode("LPPL-1.3a", false), getLicenseNode("LPPL-1.0", false), false},
		{"expect greater than: LPPL-1.3c == LPPL-1.3a", getLicenseNode("LPPL-1.3c", false), getLicenseNode("LPPL-1.3a", false), false},
		{"expect greater than: AGPL-3.0 == AGPL-1.0", getLicenseNode("AGPL-3.0", false), getLicenseNode("AGPL-1.0", false), false},
		{"expect greater than: GPL-2.0-or-later == GPL-2.0-only", getLicenseNode("GPL-2.0-or-later", true), getLicenseNode("GPL-2.0-only", false), false},
		{"expect greater than: GPL-2.0-or-later == GPL-2.0", getLicenseNode("GPL-2.0-or-later", true), getLicenseNode("GPL-2.0", false), false},
		{"expect equal: GPL-3.0 == GPL-3.0", getLicenseNode("GPL-3.0", false), getLicenseNode("GPL-3.0", false), true},
		{"expect less than: MPL-1.0 == MPL-2.0", getLicenseNode("MPL-1.0", false), getLicenseNode("MPL-2.0", false), false},
		{"incompatible: MIT == ISC", getLicenseNode("MIT", false), getLicenseNode("ISC", false), false},
		{"incompatible: OSL-1.0 == OPL-1.0", getLicenseNode("OSL-1.0", false), getLicenseNode("OPL-1.0", false), false},
		{"not simple license: (MIT OR ISC) == GPL-3.0", getLicenseNode("(MIT OR ISC)", false), getLicenseNode("GPL-3.0", false), false}, // TODO: should it raise error?
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.result, compareEQ(test.first, test.second))
		})
	}
}

func TestCompareLT(t *testing.T) {
	tests := []struct {
		name   string
		first  *Node
		second *Node
		result bool
	}{
		{"expect greater than: GPL-3.0 < GPL-2.0", getLicenseNode("GPL-3.0", false), getLicenseNode("GPL-2.0", false), false},
		{"expect greater than: GPL-3.0-only < GPL-2.0-only", getLicenseNode("GPL-3.0-only", false), getLicenseNode("GPL-2.0-only", false), false},
		{"expect greater than: LPPL-1.3a < LPPL-1.0", getLicenseNode("LPPL-1.3a", false), getLicenseNode("LPPL-1.0", false), false},
		{"expect greater than: LPPL-1.3c < LPPL-1.3a", getLicenseNode("LPPL-1.3c", false), getLicenseNode("LPPL-1.3a", false), false},
		{"expect greater than: AGPL-3.0 < AGPL-1.0", getLicenseNode("AGPL-3.0", false), getLicenseNode("AGPL-1.0", false), false},
		{"expect greater than: GPL-2.0-or-later < GPL-2.0-only", getLicenseNode("GPL-2.0-or-later", true), getLicenseNode("GPL-2.0-only", false), false},
		{"expect greater than: GPL-2.0-or-later == GPL-2.0", getLicenseNode("GPL-2.0-or-later", true), getLicenseNode("GPL-2.0", false), false},
		{"expect equal: GPL-3.0 < GPL-3.0", getLicenseNode("GPL-3.0", false), getLicenseNode("GPL-3.0", false), false},
		{"expect less than: MPL-1.0 < MPL-2.0", getLicenseNode("MPL-1.0", false), getLicenseNode("MPL-2.0", false), true},
		{"incompatible: MIT < ISC", getLicenseNode("MIT", false), getLicenseNode("ISC", false), false},
		{"incompatible: OSL-1.0 < OPL-1.0", getLicenseNode("OSL-1.0", false), getLicenseNode("OPL-1.0", false), false},
		{"not simple license: (MIT OR ISC) < GPL-3.0", getLicenseNode("(MIT OR ISC)", false), getLicenseNode("GPL-3.0", false), false}, // TODO: should it raise error?
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.result, compareLT(test.first, test.second))
		})
	}
}
