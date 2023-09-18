package spdxexp

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateLicenses(t *testing.T) {
	tests := []struct {
		name            string
		inputLicenses   []string
		allValid        bool
		invalidLicenses []string
	}{
		{"All invalid", []string{"MTI", "Apche-2.0", "0xDEADBEEF", ""}, false, []string{"MTI", "Apche-2.0", "0xDEADBEEF", ""}},
		{"All valid", []string{"MIT", "Apache-2.0", "GPL-2.0"}, true, []string{}},
		{"Some invalid", []string{"MTI", "Apche-2.0", "GPL-2.0"}, false, []string{"MTI", "Apche-2.0"}},
		{"GPL-2.0", []string{"GPL-2.0"}, true, []string{}},
		{"GPL-2.0-only", []string{"GPL-2.0-only"}, true, []string{}},
		{"SPDX Expressions are valid", []string{
			"MIT AND APACHE-2.0",
			"MIT AND APCHE-2.0",
			"LGPL-2.1-only OR MIT OR BSD-3-Clause",
			"GPL-2.0-or-later WITH Bison-exception-2.2",
		}, false, []string{"MIT AND APCHE-2.0"}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			valid, invalidLicenses := ValidateLicenses(test.inputLicenses)
			assert.EqualValues(t, test.invalidLicenses, invalidLicenses)
			assert.Equal(t, test.allValid, valid)
		})
	}
}

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
		{"WITH modifiers", "GPL-2.0-or-later WITH Bison-exception-2.2", []string{"GPL-2.0-or-later"}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			licenses, err := ExtractLicenses(test.inputExpression)
			assert.ElementsMatch(t, test.extractedLicenses, licenses)
			assert.NoError(t, err)
		})
	}
}

// TestSatisfiesSingle lets you quickly test a single call to Satisfies with a specific license expression and allowed list of licenses.
// To test a different expression, change the expression, allowed licenses, and expected result in the function body.
// TO RUN: go test ./expression -run TestSatisfiesSingle
func TestSatisfiesSingle(t *testing.T) {
	// Update these to test a different expression.
	expression := "BSD-3-Clause AND GPL-2.0"
	allowedList := []string{"BSD-3-Clause", "GPL-2.0"}
	expectedResult := true

	// Run the test.
	actualResult, err := Satisfies(expression, allowedList)
	assert.NoError(t, err)
	assert.Equal(t, expectedResult, actualResult)
}

func TestSatisfies(t *testing.T) {
	tests := []struct {
		name           string
		repoExpression string
		allowedList    []string
		satisfied      bool
		err            error
	}{
		// TODO: Test error conditions (e.g. GPL is an invalid license, Apachie + has invalid + operator)
		// regression tests from spdx-satisfies.js - comments for satisfies function
		// TODO: Commented out tests are not yet supported.
		{"MIT satisfies [MIT]", "MIT", []string{"MIT"}, true, nil},
		{"miT satisfies [MIT]", "miT", []string{"MIT"}, true, nil},
		{"MIT satisfies [mit]", "MIT", []string{"mit"}, true, nil},
		{"! MIT satisfies [Apache-2.0]", "MIT", []string{"Apache-2.0"}, false, nil},
		{"err - <empty expression> satisfies MIT", "", []string{"MIT"}, false,
			errors.New("parse error - cannot parse empty string")},
		{"err - MIT satisfies <empty allow list>", "MIT", []string{}, false,
			errors.New("allowedList requires at least one element, but is empty")},
		{"err - invalid license", "NON-EXISTENT-LICENSE", []string{"MIT", "Apache-2.0"}, false,
			errors.New("unknown license 'NON-EXISTENT-LICENSE' at offset 0")},
		{"err - invalid license in allowed list", "MIT", []string{"NON-EXISTENT-LICENSE", "Apache-2.0"}, false,
			errors.New("unknown license 'NON-EXISTENT-LICENSE' at offset 0")},

		{"MIT satisfies [MIT, Apache-2.0]", "MIT", []string{"MIT", "Apache-2.0"}, true, nil},
		{"MIT OR Apache-2.0 satisfies [MIT]", "MIT OR Apache-2.0", []string{"MIT"}, true, nil},
		{"! GPL-2.0 satisfies [MIT, Apache-2.0]", "GPL-2.0", []string{"MIT", "Apache-2.0"}, false, nil},
		{"! MIT OR Apache-2.0 satisfies [GPL-2.0]", "MIT OR Apache-2.0", []string{"GPL-2.0"}, false, nil},

		{"Apache-2.0 AND MIT satisfies [MIT, APACHE-2.0]", "Apache-2.0 AND MIT", []string{"MIT", "APACHE-2.0"}, true, nil},
		{"apache-2.0 AND mit satisfies [MIT, APACHE-2.0]", "apache-2.0 AND mit", []string{"MIT", "APACHE-2.0"}, true, nil},
		{"Apache-2.0 AND MIT satisfies [MIT, Apache-2.0]", "Apache-2.0 AND MIT", []string{"MIT", "Apache-2.0"}, true, nil},
		{"MIT AND Apache-2.0 satisfies [MIT, Apache-2.0]", "MIT AND Apache-2.0", []string{"MIT", "Apache-2.0"}, true, nil},
		{"! MIT AND Apache-2.0 satisfies [MIT]", "MIT AND Apache-2.0", []string{"MIT"}, false, nil},
		{"! GPL-2.0 satisfies [MIT, Apache-2.0]", "GPL-2.0", []string{"MIT", "Apache-2.0"}, false, nil},

		{"MIT AND Apache-2.0 satisfies [MIT, Apache-1.0, Apache-2.0]", "MIT AND Apache-2.0", []string{"MIT", "Apache-1.0", "Apache-2.0"}, true, nil},

		{"Apache-1.0+ satisfies [Apache-2.0]", "Apache-1.0+", []string{"Apache-2.0"}, true, nil},
		{"Apache-1.0+ satisfies [Apache-2.0+]", "Apache-1.0+", []string{"Apache-2.0+"}, true, nil}, // TODO: Fails here but passes js
		{"! Apache-1.0 satisfies [Apache-2.0+]", "Apache-1.0", []string{"Apache-2.0+"}, false, nil},
		{"Apache-2.0 satisfies [Apache-2.0+]", "Apache-2.0", []string{"Apache-2.0+"}, true, nil},
		{"! Apache-3.0 satisfies [Apache-2.0+]", "Apache-3.0", []string{"Apache-2.0+"}, false, errors.New("unknown license 'Apache-3.0' at offset 0")},

		{"! Apache-1.0 satisfies [Apache-2.0-or-later]", "Apache-1.0", []string{"Apache-2.0-or-later"}, false, nil},
		{"Apache-2.0 satisfies [Apache-2.0-or-later]", "Apache-2.0", []string{"Apache-2.0-or-later"}, true, nil},
		{"! Apache-3.0 satisfies [Apache-2.0-or-later]", "Apache-3.0", []string{"Apache-2.0-or-later"}, false, errors.New("unknown license 'Apache-3.0' at offset 0")},

		{"! Apache-1.0 satisfies [Apache-2.0-only]", "Apache-1.0", []string{"Apache-2.0-only"}, false, nil},
		{"Apache-2.0 satisfies [Apache-2.0-only]", "Apache-2.0", []string{"Apache-2.0-only"}, true, nil},
		{"! Apache-3.0 satisfies [Apache-2.0-only]", "Apache-3.0", []string{"Apache-2.0-only"}, false, errors.New("unknown license 'Apache-3.0' at offset 0")},

		// regression tests from spdx-satisfies.js - assert statements in README
		// TODO: Commented out tests are not yet supported.
		{"MIT satisfies [MIT]", "MIT", []string{"MIT"}, true, nil},

		{"MIT satisfies [ISC, MIT]", "MIT", []string{"ISC", "MIT"}, true, nil},
		{"Zlib satisfies [ISC, MIT, Zlib]", "Zlib", []string{"ISC", "MIT", "Zlib"}, true, nil},
		{"! GPL-3.0 satisfies [ISC, MIT]", "GPL-3.0", []string{"ISC", "MIT"}, false, nil},
		{"GPL-2.0 satisfies [GPL-2.0+]", "GPL-2.0", []string{"GPL-2.0+"}, true, nil},                 // TODO: Fails here but passes js
		{"GPL-2.0 satisfies [GPL-2.0-or-later]", "GPL-2.0", []string{"GPL-2.0-or-later"}, true, nil}, // TODO: Fails here and js
		{"GPL-3.0 satisfies [GPL-2.0+]", "GPL-3.0", []string{"GPL-2.0+"}, true, nil},
		{"GPL-1.0-or-later satisfies [GPL-2.0-or-later]", "GPL-1.0-or-later", []string{"GPL-2.0-or-later"}, true, nil},
		{"GPL-1.0+ satisfies [GPL-2.0+]", "GPL-1.0+", []string{"GPL-2.0+"}, true, nil},
		{"! GPL-1.0 satisfies [GPL-2.0+]", "GPL-1.0", []string{"GPL-2.0+"}, false, nil},
		{"GPL-2.0-only satisfies [GPL-2.0-only]", "GPL-2.0-only", []string{"GPL-2.0-only"}, true, nil},
		{"GPL-2.0 satisfies [GPL-2.0-only]", "GPL-2.0", []string{"GPL-2.0-only"}, true, nil},
		{"GPL-2.0 AND GPL-2.0-only satisfies [GPL-2.0-only]", "GPL-2.0 AND GPL-2.0-only", []string{"GPL-2.0-only"}, true, nil},
		{"GPL-3.0-only satisfies [GPL-2.0+]", "GPL-3.0-only", []string{"GPL-2.0+"}, true, nil},

		{"! GPL-2.0 satisfies [GPL-2.0+ WITH Bison-exception-2.2]",
			"GPL-2.0", []string{"GPL-2.0+ WITH Bison-exception-2.2"}, false, nil},
		{"GPL-3.0 WITH Bison-exception-2.2 satisfies [GPL-2.0+ WITH Bison-exception-2.2]",
			"GPL-3.0 WITH Bison-exception-2.2", []string{"GPL-2.0+ WITH Bison-exception-2.2"}, true, nil},

		{"(MIT OR GPL-2.0) satisfies [ISC, MIT]", "(MIT OR GPL-2.0)", []string{"ISC", "MIT"}, true, nil},
		{"(MIT AND GPL-2.0) satisfies [MIT, GPL-2.0]", "(MIT AND GPL-2.0)", []string{"MIT", "GPL-2.0"}, true, nil},
		{"MIT AND GPL-2.0 AND ISC satisfies [MIT, GPL-2.0, ISC]",
			"MIT AND GPL-2.0 AND ISC", []string{"MIT", "GPL-2.0", "ISC"}, true, nil},
		{"MIT AND GPL-2.0 AND ISC satisfies [ISC, GPL-2.0, MIT]",
			"MIT AND GPL-2.0 AND ISC", []string{"ISC", "GPL-2.0", "MIT"}, true, nil},
		{"(MIT OR GPL-2.0) AND ISC satisfies [MIT, ISC]",
			"(MIT OR GPL-2.0) AND ISC", []string{"MIT", "ISC"}, true, nil},
		{"MIT AND ISC satisfies [MIT, GPL-2.0, ISC]",
			"MIT AND ISC", []string{"MIT", "GPL-2.0", "ISC"}, true, nil},
		{"(MIT OR Apache-2.0) AND (ISC OR GPL-2.0) satisfies [Apache-2.0, ISC]",
			"(MIT OR Apache-2.0) AND (ISC OR GPL-2.0)", []string{"Apache-2.0", "ISC"}, true, nil},
		{"(MIT AND GPL-2.0) satisfies [MIT, GPL-2.0]",
			"(MIT AND GPL-2.0)", []string{"MIT", "GPL-2.0"}, true, nil},
		{"(MIT AND GPL-2.0) satisfies [GPL-2.0, MIT]",
			"(MIT AND GPL-2.0)", []string{"GPL-2.0", "MIT"}, true, nil},
		{"MIT satisfies [GPL-2.0, MIT, MIT, ISC]",
			"MIT", []string{"GPL-2.0", "MIT", "MIT", "ISC"}, true, nil},
		{"MIT AND ICU satisfies [MIT, GPL-2.0, ISC, Apache-2.0, ICU]",
			"MIT AND ICU", []string{"MIT", "GPL-2.0", "ISC", "Apache-2.0", "ICU"}, true, nil}, // TODO: This says true and the js version returns true, but it shouldn't.
		{"! (MIT AND GPL-2.0) satisfies [ISC, GPL-2.0]",
			"(MIT AND GPL-2.0)", []string{"ISC", "GPL-2.0"}, false, nil},
		{"! MIT AND (GPL-2.0 OR ISC) satisfies [MIT]",
			"MIT AND (GPL-2.0 OR ISC)", []string{"MIT"}, false, nil},
		{"! (MIT OR Apache-2.0) AND (ISC OR GPL-2.0) satisfies [MIT]",
			"(MIT OR Apache-2.0) AND (ISC OR GPL-2.0)", []string{"MIT"}, false, nil},
		{"licenseRef is expression",
			"LicenseRef-X-BSD-3-Clause-Golang", []string{"MIT", "Apache-2.0", "LicenseRef-X-BSD-3-Clause-Golang"}, true, nil},
		{"licenseRef in expression",
			"MIT AND LicenseRef-X-BSD-3-Clause-Golang", []string{"MIT", "Apache-2.0", "LicenseRef-X-BSD-3-Clause-Golang"}, true, nil},
		{"licenseRef not in expression",
			"MIT AND Apache-2.0", []string{"MIT", "Apache-2.0", "LicenseRef-X-BSD-3-Clause-Golang"}, true, nil},
		{"licenseRef not allowed",
			"MIT AND LicenseRef-X-BSD-3-Clause-Golang", []string{"MIT", "Apache-2.0"}, false, nil},
		{"licenseRef allowed, but OTHER is not allowed",
			"(BSD-3-Clause AND OTHER) OR (BSD-3-Clause AND LicenseRef-X-BSD-3-Clause-Golang)",
			[]string{"MIT", "Apache-2.0", "LicenseRef-X-BSD-3-Clause-Golang"}, false,
			errors.New("unknown license 'OTHER' at offset 18")},
		{"licenseRef with documentRef is expression",
			"DocumentRef-spdx-tool-1.2:LicenseRef-X-BSD-3-Clause-Golang",
			[]string{"MIT", "Apache-2.0", "DocumentRef-spdx-tool-1.2:LicenseRef-X-BSD-3-Clause-Golang"}, true, nil},
		{"licenseRef with documentRef in expression",
			"MIT AND DocumentRef-spdx-tool-1.2:LicenseRef-X-BSD-3-Clause-Golang",
			[]string{"MIT", "Apache-2.0", "DocumentRef-spdx-tool-1.2:LicenseRef-X-BSD-3-Clause-Golang"}, true, nil},
		{"licenseRef with documentRef not in expression",
			"MIT AND Apache-2.0",
			[]string{"MIT", "Apache-2.0", "DocumentRef-spdx-tool-1.2:LicenseRef-X-BSD-3-Clause-Golang"}, true, nil},
		{"licenseRef with documentRef not allowed",
			"MIT AND DocumentRef-spdx-tool-1.2:LicenseRef-X-BSD-3-Clause-Golang",
			[]string{"MIT", "Apache-2.0"}, false, nil},
		{"licenseRef allowed, but documentRef not allowed",
			"MIT AND DocumentRef-spdx-tool-1.2:LicenseRef-X-BSD-3-Clause-Golang",
			[]string{"MIT", "Apache-2.0", "LicenseRef-X-BSD-3-Clause-Golang"}, false, nil},
		{"licenseRef alone not allowed, but with documentRef allowed",
			"MIT AND LicenseRef-X-BSD-3-Clause-Golang",
			[]string{"MIT", "Apache-2.0", "DocumentRef-spdx-tool-1.2:LicenseRef-X-BSD-3-Clause-Golang"}, false, nil},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			satisfied, err := Satisfies(test.repoExpression, test.allowedList)
			assert.Equal(t, test.err, err)
			assert.Equal(t, test.satisfied, satisfied)
		})
	}
}

func TestExpand(t *testing.T) {
	// TODO: Add tests for licenses that include plus and/or exception.
	// TODO: Add tests for license ref and document ref.
	tests := []struct {
		name   string
		node   *node
		result [][]*node
	}{
		{singleLicense().name, singleLicense().node, singleLicense().sorted},
		{orExpression().name, orExpression().node, orExpression().sorted},
		{orAndExpression().name, orAndExpression().node, orAndExpression().sorted},
		{orOPAndCPExpression().name, orOPAndCPExpression().node, orOPAndCPExpression().sorted},
		{andOrExpression().name, andOrExpression().node, andOrExpression().sorted},
		{orAndOrExpression().name, orAndOrExpression().node, orAndOrExpression().sorted},
		{orOPAndCPOrExpression().name, orOPAndCPOrExpression().node, orOPAndCPOrExpression().sorted},
		{orOrOrExpression().name, orOrOrExpression().node, orOrOrExpression().sorted},
		{andOrAndExpression().name, andOrAndExpression().node, andOrAndExpression().sorted},
		{oPAndCPOrOPAndCPExpression().name, oPAndCPOrOPAndCPExpression().node, oPAndCPOrOPAndCPExpression().sorted},
		{andExpression().name, andExpression().node, andExpression().sorted},
		{andOPOrCPExpression().name, andOPOrCPExpression().node, andOPOrCPExpression().sorted},
		{oPOrCPAndOPOrCPExpression().name, oPOrCPAndOPOrCPExpression().node, oPOrCPAndOPOrCPExpression().sorted},
		{andOPOrCPAndExpression().name, andOPOrCPAndExpression().node, andOPOrCPAndExpression().sorted},
		{andAndAndExpression().name, andAndAndExpression().node, andAndAndExpression().sorted},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			expandResult := test.node.expand(true)
			assert.Equal(t, test.result, expandResult)
		})
	}
}

func TestExpandOr(t *testing.T) {
	tests := []struct {
		name     string
		node     *node
		expanded [][]*node
	}{
		{orExpression().name, orExpression().node, orExpression().expanded},
		{orAndExpression().name, orAndExpression().node, orAndExpression().expanded},
		{orOPAndCPExpression().name, orOPAndCPExpression().node, orOPAndCPExpression().expanded},
		{andOrExpression().name, andOrExpression().node, andOrExpression().expanded},
		{orAndOrExpression().name, orAndOrExpression().node, orAndOrExpression().expanded},
		{orOPAndCPOrExpression().name, orOPAndCPOrExpression().node, orOPAndCPOrExpression().expanded},
		{orOrOrExpression().name, orOrOrExpression().node, orOrOrExpression().expanded},
		{andOrAndExpression().name, andOrAndExpression().node, andOrAndExpression().expanded},
		{oPAndCPOrOPAndCPExpression().name, oPAndCPOrOPAndCPExpression().node, oPAndCPOrOPAndCPExpression().expanded},

		// TODO: Uncomment kitchen sink test when license plus, exception, license ref, and document ref are supported.
		// {"kitchen sink",
		// 	// "   (MIT AND Apache-1.0+)   OR   DocumentRef-spdx-tool-1.2:LicenseRef-MIT-Style-2 OR (GPL-2.0 WITH Bison-exception-2.2)",
		// 	&node{
		// 		role: expressionNode,
		// 		exp: &expressionNodePartial{
		// 			left: &node{
		// 				role: expressionNode,
		// 				exp: &expressionNodePartial{
		// 					left: &node{
		// 						role: licenseNode,
		// 						exp:  nil,
		// 						lic: &licenseNodePartial{
		// 							license:      "MIT",
		// 							hasPlus:      false,
		// 							hasException: false,
		// 							exception:    "",
		// 						},
		// 						ref: nil,
		// 					},
		// 					conjunction: "and",
		// 					right: &node{
		// 						role: licenseNode,
		// 						exp:  nil,
		// 						lic: &licenseNodePartial{
		// 							license:      "Apache-1.0",
		// 							hasPlus:      true,
		// 							hasException: false,
		// 							exception:    "",
		// 						},
		// 						ref: nil,
		// 					},
		// 				},
		// 				lic: nil,
		// 				ref: nil,
		// 			},
		// 			conjunction: "or",
		// 			right: &node{
		// 				role: expressionNode,
		// 				exp: &expressionNodePartial{
		// 					left: &node{
		// 						role: licenseRefNode,
		// 						exp:  nil,
		// 						lic:  nil,
		// 						ref: &referenceNodePartial{
		// 							hasDocumentRef: true,
		// 							documentRef:    "spdx-tool-1.2",
		// 							licenseRef:     "MIT-Style-2",
		// 						},
		// 					},
		// 					conjunction: "or",
		// 					right: &node{
		// 						role: licenseNode,
		// 						exp:  nil,
		// 						lic: &licenseNodePartial{
		// 							license:      "GPL-2.0",
		// 							hasPlus:      false,
		// 							hasException: true,
		// 							exception:    "Bison-exception-2.2",
		// 						},
		// 						ref: nil,
		// 					},
		// 				},
		// 				lic: nil,
		// 				ref: nil,
		// 			},
		// 		},
		// 		lic: nil,
		// 		ref: nil,
		// 	},
		// 	[][]string{{"MIT", "Apache-1.0+"}, {"DocumentRef-spdx-tool-1.2:LicenseRef-MIT-Style-2"}, {"GPL-2.0 with Bison-exception-2.2"}}},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			expanded := test.node.expandOr()
			assert.Equal(t, test.expanded, expanded)
		})
	}
}

func TestExpandAnd(t *testing.T) {
	tests := []struct {
		name     string
		node     *node
		expanded [][]*node
	}{
		{andExpression().name, andExpression().node, andExpression().expanded},
		{andOPOrCPExpression().name, andOPOrCPExpression().node, andOPOrCPExpression().expanded},
		{oPOrCPAndOPOrCPExpression().name, oPOrCPAndOPOrCPExpression().node, oPOrCPAndOPOrCPExpression().expanded},
		{andOPOrCPAndExpression().name, andOPOrCPAndExpression().node, andOPOrCPAndExpression().expanded},
		{andAndAndExpression().name, andAndAndExpression().node, andAndAndExpression().expanded},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			expandAndResult := test.node.expandAnd()
			assert.Equal(t, test.expanded, expandAndResult)
		})
	}
}

type testCaseData struct {
	name       string
	expression string
	node       *node
	expanded   [][]*node
	sorted     [][]*node
}

func singleLicense() testCaseData {
	return testCaseData{
		name:       "Single License",
		expression: "MIT",
		node: &node{
			role: licenseNode,
			exp:  nil,
			lic: &licenseNodePartial{
				license: "MIT", hasPlus: false,
				hasException: false, exception: ""},
			ref: nil,
		},
		expanded: [][]*node{
			{
				{
					role: licenseNode,
					exp:  nil,
					lic: &licenseNodePartial{
						license: "MIT", hasPlus: false,
						hasException: false, exception: ""},
					ref: nil,
				},
			},
		},
		sorted: [][]*node{
			{
				{
					role: licenseNode,
					exp:  nil,
					lic: &licenseNodePartial{
						license: "MIT", hasPlus: false,
						hasException: false, exception: ""},
					ref: nil,
				},
			},
		},
	}
}
func orExpression() testCaseData {
	return testCaseData{
		name:       "OR Expression",
		expression: "MIT OR Apache-2.0",
		node: &node{
			role: expressionNode,
			exp: &expressionNodePartial{
				left: &node{
					role: licenseNode,
					exp:  nil,
					lic: &licenseNodePartial{
						license: "MIT", hasPlus: false,
						hasException: false, exception: ""},
					ref: nil,
				},
				conjunction: "or",
				right: &node{
					role: licenseNode,
					exp:  nil,
					lic: &licenseNodePartial{
						license: "Apache-2.0", hasPlus: false,
						hasException: false, exception: ""},
					ref: nil,
				},
			},
			lic: nil,
			ref: nil,
		},
		expanded: [][]*node{
			{
				{
					role: licenseNode,
					exp:  nil,
					lic: &licenseNodePartial{
						license: "MIT", hasPlus: false,
						hasException: false, exception: ""},
					ref: nil,
				},
			},
			{
				{
					role: licenseNode,
					exp:  nil,
					lic: &licenseNodePartial{
						license: "Apache-2.0", hasPlus: false,
						hasException: false, exception: ""},
					ref: nil,
				},
			},
		},
		sorted: [][]*node{
			{
				{
					role: licenseNode,
					exp:  nil,
					lic: &licenseNodePartial{
						license: "Apache-2.0", hasPlus: false,
						hasException: false, exception: ""},
					ref: nil,
				},
			},
			{
				{
					role: licenseNode,
					exp:  nil,
					lic: &licenseNodePartial{
						license: "MIT", hasPlus: false,
						hasException: false, exception: ""},
					ref: nil,
				},
			},
		},
	}
}

func orAndExpression() testCaseData {
	return testCaseData{
		name:       "OR-AND Expression",
		expression: "MIT OR Apache-2.0 AND GPL-2.0",
		node: &node{
			role: expressionNode,
			exp: &expressionNodePartial{
				left: &node{
					role: licenseNode,
					exp:  nil,
					lic: &licenseNodePartial{
						license: "MIT", hasPlus: false,
						hasException: false, exception: ""},
					ref: nil,
				},
				conjunction: "or",
				right: &node{
					exp: &expressionNodePartial{
						left: &node{
							role: licenseNode,
							exp:  nil,
							lic: &licenseNodePartial{
								license:      "Apache-2.0",
								hasPlus:      false,
								hasException: false,
								exception:    "",
							},
							ref: nil,
						},
						conjunction: "and",
						right: &node{
							role: licenseNode,
							exp:  nil,
							lic: &licenseNodePartial{
								license:      "GPL-2.0",
								hasPlus:      false,
								hasException: false,
								exception:    "",
							},
							ref: nil,
						},
					},
					lic: nil,
					ref: nil,
				},
			},
			lic: nil,
			ref: nil,
		},
		expanded: [][]*node{
			{
				{
					role: licenseNode,
					exp:  nil,
					lic: &licenseNodePartial{
						license: "MIT", hasPlus: false,
						hasException: false, exception: ""},
					ref: nil,
				},
			},
			{
				{
					role: licenseNode,
					exp:  nil,
					lic: &licenseNodePartial{
						license: "Apache-2.0", hasPlus: false,
						hasException: false, exception: ""},
					ref: nil,
				},
				{
					role: licenseNode,
					exp:  nil,
					lic: &licenseNodePartial{
						license: "GPL-2.0", hasPlus: false,
						hasException: false, exception: ""},
					ref: nil,
				},
			},
		},
		sorted: [][]*node{
			{
				{
					role: licenseNode,
					exp:  nil,
					lic: &licenseNodePartial{
						license: "Apache-2.0", hasPlus: false,
						hasException: false, exception: ""},
					ref: nil,
				},
				{
					role: licenseNode,
					exp:  nil,
					lic: &licenseNodePartial{
						license: "GPL-2.0", hasPlus: false,
						hasException: false, exception: ""},
					ref: nil,
				},
			},
			{
				{
					role: licenseNode,
					exp:  nil,
					lic: &licenseNodePartial{
						license: "MIT", hasPlus: false,
						hasException: false, exception: ""},
					ref: nil,
				},
			},
		},
	}
}

func orOPAndCPExpression() testCaseData {
	return testCaseData{
		name:       "OR(AND) Expression",
		expression: "MIT OR (Apache-2.0 AND GPL-2.0)",
		node: &node{
			role: expressionNode,
			exp: &expressionNodePartial{
				left: &node{
					role: licenseNode,
					exp:  nil,
					lic: &licenseNodePartial{
						license: "MIT", hasPlus: false,
						hasException: false, exception: ""},
					ref: nil,
				},
				conjunction: "or",
				right: &node{
					exp: &expressionNodePartial{
						left: &node{
							role: licenseNode,
							exp:  nil,
							lic: &licenseNodePartial{
								license:      "Apache-2.0",
								hasPlus:      false,
								hasException: false,
								exception:    "",
							},
							ref: nil,
						},
						conjunction: "and",
						right: &node{
							role: licenseNode,
							exp:  nil,
							lic: &licenseNodePartial{
								license:      "GPL-2.0",
								hasPlus:      false,
								hasException: false,
								exception:    "",
							},
							ref: nil,
						},
					},
					lic: nil,
					ref: nil,
				},
			},
			lic: nil,
			ref: nil,
		},
		expanded: [][]*node{
			{
				{
					role: licenseNode,
					exp:  nil,
					lic: &licenseNodePartial{
						license: "MIT", hasPlus: false,
						hasException: false, exception: ""},
					ref: nil,
				},
			},
			{
				{
					role: licenseNode,
					exp:  nil,
					lic: &licenseNodePartial{
						license: "Apache-2.0", hasPlus: false,
						hasException: false, exception: ""},
					ref: nil,
				},
				{
					role: licenseNode,
					exp:  nil,
					lic: &licenseNodePartial{
						license: "GPL-2.0", hasPlus: false,
						hasException: false, exception: ""},
					ref: nil,
				},
			},
		},
		sorted: [][]*node{
			{
				{
					role: licenseNode,
					exp:  nil,
					lic: &licenseNodePartial{
						license: "Apache-2.0", hasPlus: false,
						hasException: false, exception: ""},
					ref: nil,
				},
				{
					role: licenseNode,
					exp:  nil,
					lic: &licenseNodePartial{
						license: "GPL-2.0", hasPlus: false,
						hasException: false, exception: ""},
					ref: nil,
				},
			},
			{
				{
					role: licenseNode,
					exp:  nil,
					lic: &licenseNodePartial{
						license: "MIT", hasPlus: false,
						hasException: false, exception: ""},
					ref: nil,
				},
			},
		},
	}
}

func andOrExpression() testCaseData {
	return testCaseData{
		name:       "AND-OR Expression",
		expression: "MIT AND Apache-2.0 OR GPL-2.0",
		node: &node{
			role: expressionNode,
			exp: &expressionNodePartial{
				left: &node{
					exp: &expressionNodePartial{
						left: &node{
							role: licenseNode,
							exp:  nil,
							lic: &licenseNodePartial{
								license:      "MIT",
								hasPlus:      false,
								hasException: false,
								exception:    "",
							},
							ref: nil,
						},
						conjunction: "and",
						right: &node{
							role: licenseNode,
							exp:  nil,
							lic: &licenseNodePartial{
								license:      "Apache-2.0",
								hasPlus:      false,
								hasException: false,
								exception:    "",
							},
							ref: nil,
						},
					},
					lic: nil,
					ref: nil,
				},
				conjunction: "or",
				right: &node{
					role: licenseNode,
					exp:  nil,
					lic: &licenseNodePartial{
						license: "GPL-2.0", hasPlus: false,
						hasException: false, exception: ""},
					ref: nil,
				},
			},
			lic: nil,
			ref: nil,
		},
		expanded: [][]*node{
			{
				{
					role: licenseNode,
					exp:  nil,
					lic: &licenseNodePartial{
						license: "MIT", hasPlus: false,
						hasException: false, exception: ""},
					ref: nil,
				},
				{
					role: licenseNode,
					exp:  nil,
					lic: &licenseNodePartial{
						license: "Apache-2.0", hasPlus: false,
						hasException: false, exception: ""},
					ref: nil,
				},
			},
			{
				{
					role: licenseNode,
					exp:  nil,
					lic: &licenseNodePartial{
						license: "GPL-2.0", hasPlus: false,
						hasException: false, exception: ""},
					ref: nil,
				},
			},
		},
		sorted: [][]*node{
			{
				{
					role: licenseNode,
					exp:  nil,
					lic: &licenseNodePartial{
						license: "Apache-2.0", hasPlus: false,
						hasException: false, exception: ""},
					ref: nil,
				},
				{
					role: licenseNode,
					exp:  nil,
					lic: &licenseNodePartial{
						license: "MIT", hasPlus: false,
						hasException: false, exception: ""},
					ref: nil,
				},
			},
			{
				{
					role: licenseNode,
					exp:  nil,
					lic: &licenseNodePartial{
						license: "GPL-2.0", hasPlus: false,
						hasException: false, exception: ""},
					ref: nil,
				},
			},
		},
	}
}

func orAndOrExpression() testCaseData {
	return testCaseData{
		name:       "OR-AND-OR Expression",
		expression: "MIT OR ISC AND Apache-2.0 OR GPL-2.0",
		node: &node{
			role: expressionNode,
			exp: &expressionNodePartial{
				left: &node{
					role: licenseNode,
					exp:  nil,
					lic: &licenseNodePartial{
						license:      "MIT",
						hasPlus:      false,
						hasException: false,
						exception:    "",
					},
					ref: nil,
				},
				conjunction: "or",
				right: &node{
					exp: &expressionNodePartial{
						left: &node{
							role: expressionNode,
							exp: &expressionNodePartial{
								left: &node{
									role: licenseNode,
									exp:  nil,
									lic: &licenseNodePartial{
										license:      "ISC",
										hasPlus:      false,
										hasException: false,
										exception:    "",
									},
									ref: nil,
								},
								conjunction: "and",
								right: &node{
									role: licenseNode,
									exp:  nil,
									lic: &licenseNodePartial{
										license:      "Apache-2.0",
										hasPlus:      false,
										hasException: false,
										exception:    "",
									},
									ref: nil,
								},
							},
							lic: nil,
							ref: nil,
						},
						conjunction: "or",
						right: &node{
							role: licenseNode,
							exp:  nil,
							lic: &licenseNodePartial{
								license:      "GPL-2.0",
								hasPlus:      false,
								hasException: false,
								exception:    "",
							},
							ref: nil,
						},
					},
					lic: nil,
					ref: nil,
				},
			},
			lic: nil,
			ref: nil,
		},
		expanded: [][]*node{
			{
				{
					role: licenseNode,
					exp:  nil,
					lic: &licenseNodePartial{
						license: "MIT", hasPlus: false,
						hasException: false, exception: ""},
					ref: nil,
				},
			},
			{
				{
					role: licenseNode,
					exp:  nil,
					lic: &licenseNodePartial{
						license: "ISC", hasPlus: false,
						hasException: false, exception: ""},
					ref: nil,
				},
				{
					role: licenseNode,
					exp:  nil,
					lic: &licenseNodePartial{
						license: "Apache-2.0", hasPlus: false,
						hasException: false, exception: ""},
					ref: nil,
				},
			},
			{
				{
					role: licenseNode,
					exp:  nil,
					lic: &licenseNodePartial{
						license: "GPL-2.0", hasPlus: false,
						hasException: false, exception: ""},
					ref: nil,
				},
			},
		},
		sorted: [][]*node{
			{
				{
					role: licenseNode,
					exp:  nil,
					lic: &licenseNodePartial{
						license: "Apache-2.0", hasPlus: false,
						hasException: false, exception: ""},
					ref: nil,
				},
				{
					role: licenseNode,
					exp:  nil,
					lic: &licenseNodePartial{
						license: "ISC", hasPlus: false,
						hasException: false, exception: ""},
					ref: nil,
				},
			},
			{
				{
					role: licenseNode,
					exp:  nil,
					lic: &licenseNodePartial{
						license: "GPL-2.0", hasPlus: false,
						hasException: false, exception: ""},
					ref: nil,
				},
			},
			{
				{
					role: licenseNode,
					exp:  nil,
					lic: &licenseNodePartial{
						license: "MIT", hasPlus: false,
						hasException: false, exception: ""},
					ref: nil,
				},
			},
		},
	}
}

func orOPAndCPOrExpression() testCaseData {
	return testCaseData{
		name:       "OR(AND)OR Expression",
		expression: "MIT OR (ISC AND Apache-2.0) OR GPL-2.0",
		node: &node{
			role: expressionNode,
			exp: &expressionNodePartial{
				left: &node{
					role: licenseNode,
					exp:  nil,
					lic: &licenseNodePartial{
						license:      "MIT",
						hasPlus:      false,
						hasException: false,
						exception:    "",
					},
					ref: nil,
				},
				conjunction: "or",
				right: &node{
					exp: &expressionNodePartial{
						left: &node{
							role: expressionNode,
							exp: &expressionNodePartial{
								left: &node{
									role: licenseNode,
									exp:  nil,
									lic: &licenseNodePartial{
										license:      "ISC",
										hasPlus:      false,
										hasException: false,
										exception:    "",
									},
									ref: nil,
								},
								conjunction: "and",
								right: &node{
									role: licenseNode,
									exp:  nil,
									lic: &licenseNodePartial{
										license:      "Apache-2.0",
										hasPlus:      false,
										hasException: false,
										exception:    "",
									},
									ref: nil,
								},
							},
							lic: nil,
							ref: nil,
						},
						conjunction: "or",
						right: &node{
							role: licenseNode,
							exp:  nil,
							lic: &licenseNodePartial{
								license:      "GPL-2.0",
								hasPlus:      false,
								hasException: false,
								exception:    "",
							},
							ref: nil,
						},
					},
					lic: nil,
					ref: nil,
				},
			},
			lic: nil,
			ref: nil,
		},
		expanded: [][]*node{
			{
				{
					role: licenseNode,
					exp:  nil,
					lic: &licenseNodePartial{
						license: "MIT", hasPlus: false,
						hasException: false, exception: ""},
					ref: nil,
				},
			},
			{
				{
					role: licenseNode,
					exp:  nil,
					lic: &licenseNodePartial{
						license: "ISC", hasPlus: false,
						hasException: false, exception: ""},
					ref: nil,
				},
				{
					role: licenseNode,
					exp:  nil,
					lic: &licenseNodePartial{
						license: "Apache-2.0", hasPlus: false,
						hasException: false, exception: ""},
					ref: nil,
				},
			},
			{
				{
					role: licenseNode,
					exp:  nil,
					lic: &licenseNodePartial{
						license: "GPL-2.0", hasPlus: false,
						hasException: false, exception: ""},
					ref: nil,
				},
			},
		},

		sorted: [][]*node{
			{
				{
					role: licenseNode,
					exp:  nil,
					lic: &licenseNodePartial{
						license: "Apache-2.0", hasPlus: false,
						hasException: false, exception: ""},
					ref: nil,
				},
				{
					role: licenseNode,
					exp:  nil,
					lic: &licenseNodePartial{
						license: "ISC", hasPlus: false,
						hasException: false, exception: ""},
					ref: nil,
				},
			},
			{
				{
					role: licenseNode,
					exp:  nil,
					lic: &licenseNodePartial{
						license: "GPL-2.0", hasPlus: false,
						hasException: false, exception: ""},
					ref: nil,
				},
			},
			{
				{
					role: licenseNode,
					exp:  nil,
					lic: &licenseNodePartial{
						license: "MIT", hasPlus: false,
						hasException: false, exception: ""},
					ref: nil,
				},
			},
		},
	}
}

func orOrOrExpression() testCaseData {
	return testCaseData{
		name:       "OR-OR-OR Expression",
		expression: "MIT OR ISC OR Apache-2.0 OR GPL-2.0",
		node: &node{
			role: expressionNode,
			exp: &expressionNodePartial{
				left: &node{
					role: licenseNode,
					exp:  nil,
					lic: &licenseNodePartial{
						license:      "MIT",
						hasPlus:      false,
						hasException: false,
						exception:    "",
					},
					ref: nil,
				},
				conjunction: "or",
				right: &node{
					exp: &expressionNodePartial{
						left: &node{
							role: licenseNode,
							exp:  nil,
							lic: &licenseNodePartial{
								license:      "ISC",
								hasPlus:      false,
								hasException: false,
								exception:    "",
							},
							ref: nil,
						},
						conjunction: "or",
						right: &node{
							exp: &expressionNodePartial{
								left: &node{
									role: licenseNode,
									exp:  nil,
									lic: &licenseNodePartial{
										license:      "Apache-2.0",
										hasPlus:      false,
										hasException: false,
										exception:    "",
									},
									ref: nil,
								},
								conjunction: "or",
								right: &node{
									role: licenseNode,
									exp:  nil,
									lic: &licenseNodePartial{
										license:      "GPL-2.0",
										hasPlus:      false,
										hasException: false,
										exception:    "",
									},
									ref: nil,
								},
							},
							lic: nil,
							ref: nil,
						},
					},
					lic: nil,
					ref: nil,
				},
			},
			lic: nil,
			ref: nil,
		},
		expanded: [][]*node{
			{
				{
					role: licenseNode,
					exp:  nil,
					lic: &licenseNodePartial{
						license: "MIT", hasPlus: false,
						hasException: false, exception: ""},
					ref: nil,
				},
			},
			{
				{
					role: licenseNode,
					exp:  nil,
					lic: &licenseNodePartial{
						license: "ISC", hasPlus: false,
						hasException: false, exception: ""},
					ref: nil,
				},
			},
			{
				{
					role: licenseNode,
					exp:  nil,
					lic: &licenseNodePartial{
						license: "Apache-2.0", hasPlus: false,
						hasException: false, exception: ""},
					ref: nil,
				},
			},
			{
				{
					role: licenseNode,
					exp:  nil,
					lic: &licenseNodePartial{
						license: "GPL-2.0", hasPlus: false,
						hasException: false, exception: ""},
					ref: nil,
				},
			},
		},

		sorted: [][]*node{
			{
				{
					role: licenseNode,
					exp:  nil,
					lic: &licenseNodePartial{
						license: "Apache-2.0", hasPlus: false,
						hasException: false, exception: ""},
					ref: nil,
				},
			},
			{
				{
					role: licenseNode,
					exp:  nil,
					lic: &licenseNodePartial{
						license: "GPL-2.0", hasPlus: false,
						hasException: false, exception: ""},
					ref: nil,
				},
			},
			{
				{
					role: licenseNode,
					exp:  nil,
					lic: &licenseNodePartial{
						license: "ISC", hasPlus: false,
						hasException: false, exception: ""},
					ref: nil,
				},
			},
			{
				{
					role: licenseNode,
					exp:  nil,
					lic: &licenseNodePartial{
						license: "MIT", hasPlus: false,
						hasException: false, exception: ""},
					ref: nil,
				},
			},
		},
	}
}

func andOrAndExpression() testCaseData {
	return testCaseData{
		name:       "AND-OR-AND Expression",
		expression: "MIT AND ISC OR Apache-2.0 AND GPL-2.0",
		node: &node{
			role: expressionNode,
			exp: &expressionNodePartial{
				left: &node{
					exp: &expressionNodePartial{
						left: &node{
							role: licenseNode,
							exp:  nil,
							lic: &licenseNodePartial{
								license:      "MIT",
								hasPlus:      false,
								hasException: false,
								exception:    "",
							},
							ref: nil,
						},
						conjunction: "and",
						right: &node{
							role: licenseNode,
							exp:  nil,
							lic: &licenseNodePartial{
								license:      "ISC",
								hasPlus:      false,
								hasException: false,
								exception:    "",
							},
							ref: nil,
						},
					},
					lic: nil,
					ref: nil,
				},
				conjunction: "or",
				right: &node{
					exp: &expressionNodePartial{
						left: &node{
							role: licenseNode,
							exp:  nil,
							lic: &licenseNodePartial{
								license:      "Apache-2.0",
								hasPlus:      false,
								hasException: false,
								exception:    "",
							},
							ref: nil,
						},
						conjunction: "and",
						right: &node{
							role: licenseNode,
							exp:  nil,
							lic: &licenseNodePartial{
								license:      "GPL-2.0",
								hasPlus:      false,
								hasException: false,
								exception:    "",
							},
							ref: nil,
						},
					},
					lic: nil,
					ref: nil,
				},
			},
			lic: nil,
			ref: nil,
		},
		expanded: [][]*node{
			{
				{
					role: licenseNode,
					exp:  nil,
					lic: &licenseNodePartial{
						license: "MIT", hasPlus: false,
						hasException: false, exception: ""},
					ref: nil,
				},
				{
					role: licenseNode,
					exp:  nil,
					lic: &licenseNodePartial{
						license: "ISC", hasPlus: false,
						hasException: false, exception: ""},
					ref: nil,
				},
			},
			{
				{
					role: licenseNode,
					exp:  nil,
					lic: &licenseNodePartial{
						license: "Apache-2.0", hasPlus: false,
						hasException: false, exception: ""},
					ref: nil,
				},
				{
					role: licenseNode,
					exp:  nil,
					lic: &licenseNodePartial{
						license: "GPL-2.0", hasPlus: false,
						hasException: false, exception: ""},
					ref: nil,
				},
			},
		},

		sorted: [][]*node{
			{
				{
					role: licenseNode,
					exp:  nil,
					lic: &licenseNodePartial{
						license: "Apache-2.0", hasPlus: false,
						hasException: false, exception: ""},
					ref: nil,
				},
				{
					role: licenseNode,
					exp:  nil,
					lic: &licenseNodePartial{
						license: "GPL-2.0", hasPlus: false,
						hasException: false, exception: ""},
					ref: nil,
				},
			},
			{
				{
					role: licenseNode,
					exp:  nil,
					lic: &licenseNodePartial{
						license: "ISC", hasPlus: false,
						hasException: false, exception: ""},
					ref: nil,
				},
				{
					role: licenseNode,
					exp:  nil,
					lic: &licenseNodePartial{
						license: "MIT", hasPlus: false,
						hasException: false, exception: ""},
					ref: nil,
				},
			},
		},
	}
}

func oPAndCPOrOPAndCPExpression() testCaseData {
	return testCaseData{
		name:       "(AND)OR(AND) Expression",
		expression: "(MIT AND ISC) OR (Apache-2.0 AND GPL-2.0)",
		node: &node{
			role: expressionNode,
			exp: &expressionNodePartial{
				left: &node{
					exp: &expressionNodePartial{
						left: &node{
							role: licenseNode,
							exp:  nil,
							lic: &licenseNodePartial{
								license:      "MIT",
								hasPlus:      false,
								hasException: false,
								exception:    "",
							},
							ref: nil,
						},
						conjunction: "and",
						right: &node{
							role: licenseNode,
							exp:  nil,
							lic: &licenseNodePartial{
								license:      "ISC",
								hasPlus:      false,
								hasException: false,
								exception:    "",
							},
							ref: nil,
						},
					},
					lic: nil,
					ref: nil,
				},
				conjunction: "or",
				right: &node{
					exp: &expressionNodePartial{
						left: &node{
							role: licenseNode,
							exp:  nil,
							lic: &licenseNodePartial{
								license:      "Apache-2.0",
								hasPlus:      false,
								hasException: false,
								exception:    "",
							},
							ref: nil,
						},
						conjunction: "and",
						right: &node{
							role: licenseNode,
							exp:  nil,
							lic: &licenseNodePartial{
								license:      "GPL-2.0",
								hasPlus:      false,
								hasException: false,
								exception:    "",
							},
							ref: nil,
						},
					},
					lic: nil,
					ref: nil,
				},
			},
			lic: nil,
			ref: nil,
		},
		expanded: [][]*node{
			{
				{
					role: licenseNode,
					exp:  nil,
					lic: &licenseNodePartial{
						license: "MIT", hasPlus: false,
						hasException: false, exception: ""},
					ref: nil,
				},
				{
					role: licenseNode,
					exp:  nil,
					lic: &licenseNodePartial{
						license: "ISC", hasPlus: false,
						hasException: false, exception: ""},
					ref: nil,
				},
			},
			{
				{
					role: licenseNode,
					exp:  nil,
					lic: &licenseNodePartial{
						license: "Apache-2.0", hasPlus: false,
						hasException: false, exception: ""},
					ref: nil,
				},
				{
					role: licenseNode,
					exp:  nil,
					lic: &licenseNodePartial{
						license: "GPL-2.0", hasPlus: false,
						hasException: false, exception: ""},
					ref: nil,
				},
			},
		},

		sorted: [][]*node{
			{
				{
					role: licenseNode,
					exp:  nil,
					lic: &licenseNodePartial{
						license: "Apache-2.0", hasPlus: false,
						hasException: false, exception: ""},
					ref: nil,
				},
				{
					role: licenseNode,
					exp:  nil,
					lic: &licenseNodePartial{
						license: "GPL-2.0", hasPlus: false,
						hasException: false, exception: ""},
					ref: nil,
				},
			},
			{
				{
					role: licenseNode,
					exp:  nil,
					lic: &licenseNodePartial{
						license: "ISC", hasPlus: false,
						hasException: false, exception: ""},
					ref: nil,
				},
				{
					role: licenseNode,
					exp:  nil,
					lic: &licenseNodePartial{
						license: "MIT", hasPlus: false,
						hasException: false, exception: ""},
					ref: nil,
				},
			},
		},
	}
}

func andExpression() testCaseData {
	return testCaseData{
		name:       "AND Expression",
		expression: "MIT AND Apache-2.0",
		node: &node{
			role: expressionNode,
			exp: &expressionNodePartial{
				left: &node{
					role: licenseNode,
					exp:  nil,
					lic: &licenseNodePartial{
						license: "MIT", hasPlus: false,
						hasException: false, exception: ""},
					ref: nil,
				},
				conjunction: "and",
				right: &node{
					role: licenseNode,
					exp:  nil,
					lic: &licenseNodePartial{
						license: "Apache-2.0", hasPlus: false,
						hasException: false, exception: ""},
					ref: nil,
				},
			},
			lic: nil,
			ref: nil,
		},
		expanded: [][]*node{
			{
				{
					role: licenseNode,
					exp:  nil,
					lic: &licenseNodePartial{
						license: "MIT", hasPlus: false,
						hasException: false, exception: ""},
					ref: nil,
				},
				{
					role: licenseNode,
					exp:  nil,
					lic: &licenseNodePartial{
						license: "Apache-2.0", hasPlus: false,
						hasException: false, exception: ""},
					ref: nil,
				},
			},
		},

		sorted: [][]*node{
			{
				{
					role: licenseNode,
					exp:  nil,
					lic: &licenseNodePartial{
						license: "Apache-2.0", hasPlus: false,
						hasException: false, exception: ""},
					ref: nil,
				},
				{
					role: licenseNode,
					exp:  nil,
					lic: &licenseNodePartial{
						license: "MIT", hasPlus: false,
						hasException: false, exception: ""},
					ref: nil,
				},
			},
		},
	}
}

func andOPOrCPExpression() testCaseData {
	return testCaseData{
		name:       "AND(OR) Expression",
		expression: "MIT AND (Apache-2.0 OR GPL-2.0)",
		node: &node{
			role: expressionNode,
			exp: &expressionNodePartial{
				left: &node{
					role: licenseNode,
					exp:  nil,
					lic: &licenseNodePartial{
						license:      "MIT",
						hasPlus:      false,
						hasException: false,
						exception:    "",
					},
					ref: nil,
				},
				conjunction: "and",
				right: &node{
					exp: &expressionNodePartial{
						left: &node{
							role: licenseNode,
							exp:  nil,
							lic: &licenseNodePartial{
								license:      "Apache-2.0",
								hasPlus:      false,
								hasException: false,
								exception:    "",
							},
							ref: nil,
						},
						conjunction: "or",
						right: &node{
							role: licenseNode,
							exp:  nil,
							lic: &licenseNodePartial{
								license:      "GPL-2.0",
								hasPlus:      false,
								hasException: false,
								exception:    "",
							},
							ref: nil,
						},
					},
					lic: nil,
					ref: nil,
				},
			},
			lic: nil,
			ref: nil,
		},
		expanded: [][]*node{
			{
				{
					role: licenseNode,
					exp:  nil,
					lic: &licenseNodePartial{
						license: "MIT", hasPlus: false,
						hasException: false, exception: ""},
					ref: nil,
				},
				{
					role: licenseNode,
					exp:  nil,
					lic: &licenseNodePartial{
						license: "Apache-2.0", hasPlus: false,
						hasException: false, exception: ""},
					ref: nil,
				},
			},
			{
				{
					role: licenseNode,
					exp:  nil,
					lic: &licenseNodePartial{
						license: "MIT", hasPlus: false,
						hasException: false, exception: ""},
					ref: nil,
				},
				{
					role: licenseNode,
					exp:  nil,
					lic: &licenseNodePartial{
						license: "GPL-2.0", hasPlus: false,
						hasException: false, exception: ""},
					ref: nil,
				},
			},
		},

		sorted: [][]*node{
			{
				{
					role: licenseNode,
					exp:  nil,
					lic: &licenseNodePartial{
						license: "Apache-2.0", hasPlus: false,
						hasException: false, exception: ""},
					ref: nil,
				},
				{
					role: licenseNode,
					exp:  nil,
					lic: &licenseNodePartial{
						license: "MIT", hasPlus: false,
						hasException: false, exception: ""},
					ref: nil,
				},
			},
			{
				{
					role: licenseNode,
					exp:  nil,
					lic: &licenseNodePartial{
						license: "GPL-2.0", hasPlus: false,
						hasException: false, exception: ""},
					ref: nil,
				},
				{
					role: licenseNode,
					exp:  nil,
					lic: &licenseNodePartial{
						license: "MIT", hasPlus: false,
						hasException: false, exception: ""},
					ref: nil,
				},
			},
		},
	}
}

func oPOrCPAndOPOrCPExpression() testCaseData {
	return testCaseData{
		name:       "(OR)AND(OR) Expression",
		expression: "(MIT OR ISC) AND (Apache-2.0 OR GPL-2.0)",
		node: &node{
			role: expressionNode,
			exp: &expressionNodePartial{
				left: &node{
					exp: &expressionNodePartial{
						left: &node{
							role: licenseNode,
							exp:  nil,
							lic: &licenseNodePartial{
								license:      "MIT",
								hasPlus:      false,
								hasException: false,
								exception:    "",
							},
							ref: nil,
						},
						conjunction: "or",
						right: &node{
							role: licenseNode,
							exp:  nil,
							lic: &licenseNodePartial{
								license:      "ISC",
								hasPlus:      false,
								hasException: false,
								exception:    "",
							},
							ref: nil,
						},
					},
					lic: nil,
					ref: nil,
				},
				conjunction: "and",
				right: &node{
					exp: &expressionNodePartial{
						left: &node{
							role: licenseNode,
							exp:  nil,
							lic: &licenseNodePartial{
								license:      "Apache-2.0",
								hasPlus:      false,
								hasException: false,
								exception:    "",
							},
							ref: nil,
						},
						conjunction: "or",
						right: &node{
							role: licenseNode,
							exp:  nil,
							lic: &licenseNodePartial{
								license:      "GPL-2.0",
								hasPlus:      false,
								hasException: false,
								exception:    "",
							},
							ref: nil,
						},
					},
					lic: nil,
					ref: nil,
				},
			},
			lic: nil,
			ref: nil,
		},
		expanded: [][]*node{
			{
				{
					role: licenseNode,
					exp:  nil,
					lic: &licenseNodePartial{
						license: "MIT", hasPlus: false,
						hasException: false, exception: ""},
					ref: nil,
				},
				{
					role: licenseNode,
					exp:  nil,
					lic: &licenseNodePartial{
						license: "Apache-2.0", hasPlus: false,
						hasException: false, exception: ""},
					ref: nil,
				},
			},
			{
				{
					role: licenseNode,
					exp:  nil,
					lic: &licenseNodePartial{
						license: "ISC", hasPlus: false,
						hasException: false, exception: ""},
					ref: nil,
				},
				{
					role: licenseNode,
					exp:  nil,
					lic: &licenseNodePartial{
						license: "Apache-2.0", hasPlus: false,
						hasException: false, exception: ""},
					ref: nil,
				},
			},
			{
				{
					role: licenseNode,
					exp:  nil,
					lic: &licenseNodePartial{
						license: "MIT", hasPlus: false,
						hasException: false, exception: ""},
					ref: nil,
				},
				{
					role: licenseNode,
					exp:  nil,
					lic: &licenseNodePartial{
						license: "GPL-2.0", hasPlus: false,
						hasException: false, exception: ""},
					ref: nil,
				},
			},
			{
				{
					role: licenseNode,
					exp:  nil,
					lic: &licenseNodePartial{
						license: "ISC", hasPlus: false,
						hasException: false, exception: ""},
					ref: nil,
				},
				{
					role: licenseNode,
					exp:  nil,
					lic: &licenseNodePartial{
						license: "GPL-2.0", hasPlus: false,
						hasException: false, exception: ""},
					ref: nil,
				},
			},
		},

		sorted: [][]*node{
			{
				{
					role: licenseNode,
					exp:  nil,
					lic: &licenseNodePartial{
						license: "Apache-2.0", hasPlus: false,
						hasException: false, exception: ""},
					ref: nil,
				},
				{
					role: licenseNode,
					exp:  nil,
					lic: &licenseNodePartial{
						license: "ISC", hasPlus: false,
						hasException: false, exception: ""},
					ref: nil,
				},
			},
			{
				{
					role: licenseNode,
					exp:  nil,
					lic: &licenseNodePartial{
						license: "Apache-2.0", hasPlus: false,
						hasException: false, exception: ""},
					ref: nil,
				},
				{
					role: licenseNode,
					exp:  nil,
					lic: &licenseNodePartial{
						license: "MIT", hasPlus: false,
						hasException: false, exception: ""},
					ref: nil,
				},
			},
			{
				{
					role: licenseNode,
					exp:  nil,
					lic: &licenseNodePartial{
						license: "GPL-2.0", hasPlus: false,
						hasException: false, exception: ""},
					ref: nil,
				},
				{
					role: licenseNode,
					exp:  nil,
					lic: &licenseNodePartial{
						license: "ISC", hasPlus: false,
						hasException: false, exception: ""},
					ref: nil,
				},
			},
			{
				{
					role: licenseNode,
					exp:  nil,
					lic: &licenseNodePartial{
						license: "GPL-2.0", hasPlus: false,
						hasException: false, exception: ""},
					ref: nil,
				},
				{
					role: licenseNode,
					exp:  nil,
					lic: &licenseNodePartial{
						license: "MIT", hasPlus: false,
						hasException: false, exception: ""},
					ref: nil,
				},
			},
		},
	}
}

func andOPOrCPAndExpression() testCaseData {
	return testCaseData{
		name:       "AND(OR)AND Expression",
		expression: "MIT AND (ISC OR Apache-2.0) AND GPL-2.0",
		node: &node{
			role: expressionNode,
			exp: &expressionNodePartial{
				left: &node{
					role: licenseNode,
					exp:  nil,
					lic: &licenseNodePartial{
						license:      "MIT",
						hasPlus:      false,
						hasException: false,
						exception:    "",
					},
					ref: nil,
				},
				conjunction: "and",
				right: &node{
					role: expressionNode,
					exp: &expressionNodePartial{
						left: &node{
							role: expressionNode,
							exp: &expressionNodePartial{
								left: &node{
									role: licenseNode,
									exp:  nil,
									lic: &licenseNodePartial{
										license:      "ISC",
										hasPlus:      false,
										hasException: false,
										exception:    "",
									},
									ref: nil,
								},
								conjunction: "or",
								right: &node{
									role: licenseNode,
									exp:  nil,
									lic: &licenseNodePartial{
										license:      "Apache-2.0",
										hasPlus:      false,
										hasException: false,
										exception:    "",
									},
									ref: nil,
								},
							},
							lic: nil,
							ref: nil,
						},
						conjunction: "and",
						right: &node{
							role: licenseNode,
							exp:  nil,
							lic: &licenseNodePartial{
								license:      "GPL-2.0",
								hasPlus:      false,
								hasException: false,
								exception:    "",
							},
							ref: nil,
						},
					},
					lic: nil,
					ref: nil,
				},
			},
			lic: nil,
			ref: nil,
		},
		expanded: [][]*node{
			{
				{
					role: licenseNode,
					exp:  nil,
					lic: &licenseNodePartial{
						license: "MIT", hasPlus: false,
						hasException: false, exception: ""},
					ref: nil,
				},
				{
					role: licenseNode,
					exp:  nil,
					lic: &licenseNodePartial{
						license: "ISC", hasPlus: false,
						hasException: false, exception: ""},
					ref: nil,
				},
				{
					role: licenseNode,
					exp:  nil,
					lic: &licenseNodePartial{
						license: "GPL-2.0", hasPlus: false,
						hasException: false, exception: ""},
					ref: nil,
				},
			},
			{
				{
					role: licenseNode,
					exp:  nil,
					lic: &licenseNodePartial{
						license: "MIT", hasPlus: false,
						hasException: false, exception: ""},
					ref: nil,
				},
				{
					role: licenseNode,
					exp:  nil,
					lic: &licenseNodePartial{
						license: "Apache-2.0", hasPlus: false,
						hasException: false, exception: ""},
					ref: nil,
				},
				{
					role: licenseNode,
					exp:  nil,
					lic: &licenseNodePartial{
						license: "GPL-2.0", hasPlus: false,
						hasException: false, exception: ""},
					ref: nil,
				},
			},
		},

		sorted: [][]*node{
			{
				{
					role: licenseNode,
					exp:  nil,
					lic: &licenseNodePartial{
						license: "Apache-2.0", hasPlus: false,
						hasException: false, exception: ""},
					ref: nil,
				},
				{
					role: licenseNode,
					exp:  nil,
					lic: &licenseNodePartial{
						license: "GPL-2.0", hasPlus: false,
						hasException: false, exception: ""},
					ref: nil,
				},
				{
					role: licenseNode,
					exp:  nil,
					lic: &licenseNodePartial{
						license: "MIT", hasPlus: false,
						hasException: false, exception: ""},
					ref: nil,
				},
			},
			{
				{
					role: licenseNode,
					exp:  nil,
					lic: &licenseNodePartial{
						license: "GPL-2.0", hasPlus: false,
						hasException: false, exception: ""},
					ref: nil,
				},
				{
					role: licenseNode,
					exp:  nil,
					lic: &licenseNodePartial{
						license: "ISC", hasPlus: false,
						hasException: false, exception: ""},
					ref: nil,
				},
				{
					role: licenseNode,
					exp:  nil,
					lic: &licenseNodePartial{
						license: "MIT", hasPlus: false,
						hasException: false, exception: ""},
					ref: nil,
				},
			},
		},
	}
}

func andAndAndExpression() testCaseData {
	return testCaseData{
		name:       "AND-AND-AND Expression",
		expression: "MIT AND ISC AND Apache-2.0 AND GPL-2.0",
		node: &node{
			role: expressionNode,
			exp: &expressionNodePartial{
				left: &node{
					role: licenseNode,
					exp:  nil,
					lic: &licenseNodePartial{
						license:      "MIT",
						hasPlus:      false,
						hasException: false,
						exception:    "",
					},
					ref: nil,
				},
				conjunction: "and",
				right: &node{
					role: expressionNode,
					exp: &expressionNodePartial{
						left: &node{
							role: licenseNode,
							exp:  nil,
							lic: &licenseNodePartial{
								license:      "ISC",
								hasPlus:      false,
								hasException: false,
								exception:    "",
							},
							ref: nil,
						},
						conjunction: "and",
						right: &node{
							role: expressionNode,
							exp: &expressionNodePartial{
								left: &node{
									role: licenseNode,
									exp:  nil,
									lic: &licenseNodePartial{
										license:      "Apache-2.0",
										hasPlus:      false,
										hasException: false,
										exception:    "",
									},
									ref: nil,
								},
								conjunction: "and",
								right: &node{
									role: licenseNode,
									exp:  nil,
									lic: &licenseNodePartial{
										license:      "GPL-2.0",
										hasPlus:      false,
										hasException: false,
										exception:    "",
									},
									ref: nil,
								},
							},
							lic: nil,
							ref: nil,
						},
					},
					lic: nil,
					ref: nil,
				},
			},
			lic: nil,
			ref: nil,
		},
		expanded: [][]*node{
			{
				{
					role: licenseNode,
					exp:  nil,
					lic: &licenseNodePartial{
						license: "MIT", hasPlus: false,
						hasException: false, exception: ""},
					ref: nil,
				},
				{
					role: licenseNode,
					exp:  nil,
					lic: &licenseNodePartial{
						license: "ISC", hasPlus: false,
						hasException: false, exception: ""},
					ref: nil,
				},
				{
					role: licenseNode,
					exp:  nil,
					lic: &licenseNodePartial{
						license: "Apache-2.0", hasPlus: false,
						hasException: false, exception: ""},
					ref: nil,
				},
				{
					role: licenseNode,
					exp:  nil,
					lic: &licenseNodePartial{
						license: "GPL-2.0", hasPlus: false,
						hasException: false, exception: ""},
					ref: nil,
				},
			},
		},

		sorted: [][]*node{
			{
				{
					role: licenseNode,
					exp:  nil,
					lic: &licenseNodePartial{
						license: "Apache-2.0", hasPlus: false,
						hasException: false, exception: ""},
					ref: nil,
				},
				{
					role: licenseNode,
					exp:  nil,
					lic: &licenseNodePartial{
						license: "GPL-2.0", hasPlus: false,
						hasException: false, exception: ""},
					ref: nil,
				},
				{
					role: licenseNode,
					exp:  nil,
					lic: &licenseNodePartial{
						license: "ISC", hasPlus: false,
						hasException: false, exception: ""},
					ref: nil,
				},
				{
					role: licenseNode,
					exp:  nil,
					lic: &licenseNodePartial{
						license: "MIT", hasPlus: false,
						hasException: false, exception: ""},
					ref: nil,
				},
			},
		},
	}
}

func ExampleSatisfies_singleLicense() {
	fmt.Println(Satisfies("MIT", []string{"MIT"}))
	// Output: true <nil>
}

func ExampleSatisfies_or() {
	fmt.Println(Satisfies("MIT OR Apache-2.0", []string{"MIT"}))
	// Output: true <nil>
}

func ExampleSatisfies_orNotFound() {
	fmt.Println(Satisfies("MIT OR Apache-2.0", []string{"GPL-2.0"}))
	// Output: false <nil>
}

func ExampleSatisfies_and() {
	fmt.Println(Satisfies("Apache-2.0 AND MIT", []string{"MIT", "Apache-2.0"}))
	// Output: true <nil>
}

func ExampleSatisfies_andNotFound() {
	fmt.Println(Satisfies("MIT AND Apache-2.0", []string{"MIT"}))
	// Output: false <nil>
}

func ExampleSatisfies_plus() {
	fmt.Println(Satisfies("Apache-2.0", []string{"Apache-1.0+"}))
	// Output: true <nil>
}

func ExampleSatisfies_plusNotFound() {
	fmt.Println(Satisfies("Apache-1.0", []string{"Apache-2.0+"}))
	// Output: false <nil>
}

func ExampleSatisfies_orLater() {
	fmt.Println(Satisfies("Apache-2.0", []string{"Apache-1.0-or-later"}))
	// Output: true <nil>
}

func ExampleSatisfies_orLaterNotFound() {
	fmt.Println(Satisfies("Apache-1.0", []string{"Apache-2.0-or-later"}))
	// Output: false <nil>
}

func ExampleSatisfies_only() {
	fmt.Println(Satisfies("Apache-1.0", []string{"Apache-1.0-only"}))
	// Output: true <nil>
}

func ExampleSatisfies_onlyNotFound() {
	fmt.Println(Satisfies("Apache-2.0", []string{"Apache-1.0-only"}))
	// Output: false <nil>
}

func ExampleSatisfies_extraAllowedLicense() {
	fmt.Println(Satisfies("MIT AND Apache-2.0", []string{"MIT", "Apache-1.0", "Apache-2.0"}))
	// Output: true <nil>
}

func ExampleSatisfies_errorUnknownLicense() {
	fmt.Println(Satisfies("GPL", []string{"GPL"}))
	// Output: false unknown license 'GPL' at offset 0
}

func ExampleValidateLicenses_allGood() {
	fmt.Println(ValidateLicenses([]string{"MIT", "Apache-2.0", "GPL-2.0"}))
	// Output: true []
}

func ExampleValidateLicenses_oneBad() {
	fmt.Println(ValidateLicenses([]string{"MIT", "Apache-2.0", "GPL"}))
	// Output: false [GPL]
}

func ExampleValidateLicenses_allBad() {
	fmt.Println(ValidateLicenses([]string{"MTI", "Apache--2.0", "GPL"}))
	// Output: false [MTI Apache--2.0 GPL]
}
