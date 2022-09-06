package spdxexp

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSatisfies(t *testing.T) {
	tests := []struct {
		name      string
		firstExp  string
		secondExp string
		satisfied bool
		err       error
	}{
		// regression tests from spdx-satisfies.js - comments for satisfies function
		// TODO: Commented out tests are not yet supported.
		{"MIT satisfies MIT is true", "MIT", "MIT", true, nil},
		{"MIT satisfies Apache-2.0 is false", "MIT", "Apache-2.0", false, nil},

		// 		// {"MIT satisfies MIT OR Apache-2.0 is true", "MIT", "MIT OR Apache-2.0", true, nil},
		// 		// {"MIT OR Apache-2.0 satisfies MIT is true", "MIT OR Apache-2.0", "MIT", true, nil},
		// 		// {"GPL satisfies MIT OR Apache-2.0 is false", "GPL", "MIT OR Apache-2.0", false, nil},
		// 		// {"MIT OR Apache-2.0 satisfies GPL is false", "MIT OR Apache-2.0", "GPL", false, nil},

		// 		// {"Apache-2.0 AND MIT satisfies MIT AND Apache-2.0 is true", "Apache-2.0 AND MIT", "MIT AND Apache-2.0", true, nil},
		// 		// {"MIT AND Apache-2.0 satisfies MIT AND Apache-2.0 is true", "MIT AND Apache-2.0", "MIT AND Apache-2.0", true, nil},
		// 		// {"MIT satisfies MIT AND Apache-2.0 is false", "MIT", "MIT AND Apache-2.0", false, nil},
		// 		// {"MIT AND Apache-2.0 satisfies MIT is false", "MIT AND Apache-2.0", "MIT", false, nil},
		// 		// {"GPL satisfies MIT AND Apache-2.0 is false", "GPL", "MIT AND Apache-2.0", false, nil},

		// 		// {"MIT AND Apache-2.0 satisfies MIT AND (Apache-1.0 OR Apache-2.0)", "MIT AND Apache-2.0", "MIT AND (Apache-1.0 OR Apache-2.0)", true, nil},

		// 		// {"Apache-1.0+ satisfies Apache-2.0+ is true", "Apache-1.0+", "Apache-2.0+", true, nil}, // TODO: why does this fail here but passes js?
		// 		{"Apache-1.0 satisfies Apache-2.0+ is false", "Apache-1.0", "Apache-2.0+", false, nil},
		// 		{"Apache-2.0 satisfies Apache-2.0+ is true", "Apache-2.0", "Apache-2.0+", true, nil},
		// 		// {"Apache-3.0 satisfies Apache-2.0+ is true", "Apache-3.0", "Apache-2.0+", true, nil}, // TODO: gets error b/c Apache-3.0 doesn't exist -- need better error message

		// 		{"Apache-1.0 satisfies Apache-2.0-or-later is false", "Apache-1.0", "Apache-2.0-or-later", false, nil},
		// 		{"Apache-2.0 satisfies Apache-2.0-or-later is true", "Apache-2.0", "Apache-2.0-or-later", true, nil},
		// 		// {"Apache-3.0 satisfies Apache-2.0-or-later is true", "Apache-3.0", "Apache-2.0-or-later", true, nil},

		// 		{"Apache-1.0 satisfies Apache-2.0-only is false", "Apache-1.0", "Apache-2.0-only", false, nil},
		// 		{"Apache-2.0 satisfies Apache-2.0-only is true", "Apache-2.0", "Apache-2.0-only", true, nil},
		// 		// {"Apache-3.0 satisfies Apache-2.0-only is false", "Apache-3.0", "Apache-2.0-only", false, nil},

		// 		// regression tests from spdx-satisfies.js - assert statements in README
		// 		// TODO: Commented out tests are not yet supported.
		// 		{"MIT satisfies MIT", "MIT", "MIT", true, nil},

		// 		// {"MIT satisfies (ISC OR MIT)", "MIT", "(ISC OR MIT)", true, nil},
		// 		// {"Zlib satisfies (ISC OR (MIT OR Zlib))", "Zlib", "(ISC OR (MIT OR Zlib))", true, nil},
		// 		// {"GPL-3.0 !satisfies (ISC OR MIT)", "GPL-3.0", "(ISC OR MIT)", false, nil},

		// 		// {"GPL-2.0 satisfies GPL-2.0+", "GPL-2.0", "GPL-2.0+", true, nil}, // TODO: why does this fail here but passes js?
		// 		// {"GPL-2.0 satisfies GPL-2.0-or-later", "GPL-2.0", "GPL-2.0-or-later", true, nil}, // TODO: why does this fail here but passes js?
		// 		{"GPL-3.0 satisfies GPL-2.0+", "GPL-3.0", "GPL-2.0+", true, nil},
		// 		{"GPL-1.0-or-later satisfies GPL-2.0-or-later", "GPL-1.0-or-later", "GPL-2.0-or-later", true, nil},
		// 		{"GPL-1.0+ satisfies GPL-2.0+", "GPL-1.0+", "GPL-2.0+", true, nil},
		// 		{"GPL-1.0 !satisfies GPL-2.0+", "GPL-1.0", "GPL-2.0+", false, nil},
		// 		{"GPL-2.0-only satisfies GPL-2.0-only", "GPL-2.0-only", "GPL-2.0-only", true, nil},
		// 		{"GPL-3.0-only satisfies GPL-2.0+", "GPL-3.0-only", "GPL-2.0+", true, nil},

		// 		// {"GPL-2.0 !satisfies GPL-2.0+ WITH Bison-exception-2.2",
		// 		// 	"GPL-2.0", "GPL-2.0+ WITH Bison-exception-2.2", false, nil},
		// 		// {"GPL-3.0 WITH Bison-exception-2.2 satisfies GPL-2.0+ WITH Bison-exception-2.2",
		// 		// 	"GPL-3.0 WITH Bison-exception-2.2", "GPL-2.0+ WITH Bison-exception-2.2", true, nil},

		// 		// {"(MIT OR GPL-2.0) satisfies (ISC OR MIT)", "(MIT OR GPL-2.0)", "(ISC OR MIT)", true, nil},
		// 		// {"(MIT AND GPL-2.0) satisfies (MIT AND GPL-2.0)", "(MIT AND GPL-2.0)", "(MIT AND GPL-2.0)", true, nil},
		// 		// {"MIT AND GPL-2.0 AND ISC satisfies MIT AND GPL-2.0 AND ISC",
		// 		// 	"MIT AND GPL-2.0 AND ISC", "MIT AND GPL-2.0 AND ISC", true, nil},
		// 		// {"MIT AND GPL-2.0 AND ISC satisfies ISC AND GPL-2.0 AND MIT",
		// 		// 	"MIT AND GPL-2.0 AND ISC", "ISC AND GPL-2.0 AND MIT", true, nil},
		// 		// {"(MIT OR GPL-2.0) AND ISC satisfies MIT AND ISC",
		// 		// 	"(MIT OR GPL-2.0) AND ISC", "MIT AND ISC", true, nil},
		// 		// {"MIT AND ISC satisfies (MIT OR GPL-2.0) AND ISC",
		// 		// 	"MIT AND ISC", "(MIT OR GPL-2.0) AND ISC", true, nil},
		// 		// {"MIT AND ISC satisfies (MIT AND GPL-2.0) OR ISC",
		// 		// 	"MIT AND ISC", "(MIT AND GPL-2.0) OR ISC", true, nil},
		// 		// {"(MIT OR Apache-2.0) AND (ISC OR GPL-2.0) satisfies Apache-2.0 AND ISC",
		// 		// 	"(MIT OR Apache-2.0) AND (ISC OR GPL-2.0)", "Apache-2.0 AND ISC", true, nil},
		// 		// {"(MIT OR Apache-2.0) AND (ISC OR GPL-2.0) satisfies Apache-2.0 OR ISC",
		// 		// 	"(MIT OR Apache-2.0) AND (ISC OR GPL-2.0)", "Apache-2.0 OR ISC", true, nil},
		// 		// {"(MIT AND GPL-2.0) satisfies (MIT OR GPL-2.0)",
		// 		// 	"(MIT AND GPL-2.0)", "(MIT OR GPL-2.0)", true, nil},
		// 		// {"(MIT AND GPL-2.0) satisfies (GPL-2.0 AND MIT)",
		// 		// 	"(MIT AND GPL-2.0)", "(GPL-2.0 AND MIT)", true, nil},
		// 		// {"MIT satisfies (GPL-2.0 OR MIT) AND (MIT OR ISC)",
		// 		// 	"MIT", "(GPL-2.0 OR MIT) AND (MIT OR ISC)", true, nil},
		// 		// {"MIT AND ICU satisfies (MIT AND GPL-2.0) OR (ISC AND (Apache-2.0 OR ICU))",
		// 		// 	"MIT AND ICU", "(MIT AND GPL-2.0) OR (ISC AND (Apache-2.0 OR ICU))", true, nil},
		// 		// {"(MIT AND GPL-2.0) !satisfies (ISC OR GPL-2.0)",
		// 		// 	"(MIT AND GPL-2.0)", "(ISC OR GPL-2.0)", false, nil},
		// 		// {"MIT AND (GPL-2.0 OR ISC) !satisfies MIT",
		// 		// 	"MIT AND (GPL-2.0 OR ISC)", "MIT", false, nil},
		// 		// {"(MIT OR Apache-2.0) AND (ISC OR GPL-2.0) !satisfies MIT",
		// 		// 	"(MIT OR Apache-2.0) AND (ISC OR GPL-2.0)", "MIT", false, nil},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			satisfied, err := Satisfies(test.firstExp, test.secondExp)
			assert.Equal(t, test.err, err)
			assert.Equal(t, test.satisfied, satisfied)
		})
	}
}

func TestLicenseString(t *testing.T) {
	tests := []struct {
		name   string
		node   *Node
		result string
	}{
		{"License node - simple",
			&Node{
				role: LICENSE_NODE,
				exp:  nil,
				lic: &licenseNodePartial{
					license: "MIT", hasPlus: false,
					hasException: false, exception: ""},
				ref: nil,
			}, "MIT"},
		{"License node - plus",
			&Node{
				role: LICENSE_NODE,
				exp:  nil,
				lic: &licenseNodePartial{
					license: "Apache-1.0", hasPlus: true,
					hasException: false, exception: ""},
				ref: nil,
			}, "Apache-1.0+"},
		{"License node - exception",
			&Node{
				role: LICENSE_NODE,
				exp:  nil,
				lic: &licenseNodePartial{
					license: "GPL-2.0", hasPlus: false,
					hasException: true, exception: "Bison-exception-2.2"},
				ref: nil,
			}, "GPL-2.0 WITH Bison-exception-2.2"},
		{"LicenseRef node - simple",
			&Node{
				role: LICENSEREF_NODE,
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
				role: LICENSEREF_NODE,
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
			license := *test.node.licenseString()
			assert.Equal(t, test.result, license)
		})
	}
}

func TestFlatten(t *testing.T) {
	tests := []struct {
		name      string
		node      *Node
		flattened []string
	}{
		{"License node", // "MIT"
			&Node{
				role: LICENSE_NODE,
				exp:  nil,
				lic: &licenseNodePartial{
					license: "MIT", hasPlus: false,
					hasException: false, exception: ""},
				ref: nil,
			},
			[]string{"MIT"}},
		{orExpression().name, orExpression().node, orExpression().flattened},
		{orAndExpression().name, orAndExpression().node, orAndExpression().flattened},
		{or_And_Expression().name, or_And_Expression().node, or_And_Expression().flattened},
		{andOrExpression().name, andOrExpression().node, andOrExpression().flattened},
		{orAndOrExpression().name, orAndOrExpression().node, orAndOrExpression().flattened},
		{or_And_OrExpression().name, or_And_OrExpression().node, or_And_OrExpression().flattened},
		{orOrOrExpression().name, orOrOrExpression().node, orOrOrExpression().flattened},
		{andOrAndExpression().name, andOrAndExpression().node, andOrAndExpression().flattened},
		{_and_Or_And_Expression().name, _and_Or_And_Expression().node, _and_Or_And_Expression().flattened},
		{andExpression().name, andExpression().node, andExpression().flattened},
		{and_Or_Expression().name, and_Or_Expression().node, and_Or_Expression().flattened},
		{_or_And_Or_Expression().name, _or_And_Or_Expression().node, _or_And_Or_Expression().flattened},
		{and_Or_AndExpression().name, and_Or_AndExpression().node, and_Or_AndExpression().flattened},
		{andAndAndExpression().name, andAndAndExpression().node, andAndAndExpression().flattened},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			flattened := test.node.flatten()
			assert.Equal(t, test.flattened, flattened)
		})
	}
}

func TestExpand(t *testing.T) {
	// TODO: Add tests for licenses that include plus and/or exception.
	// TODO: Add tests for license ref and document ref.
	tests := []struct {
		name   string
		node   *Node
		result [][]string
	}{
		{"License node", // "MIT"
			&Node{
				role: LICENSE_NODE,
				exp:  nil,
				lic: &licenseNodePartial{
					license: "MIT", hasPlus: false,
					hasException: false, exception: ""},
				ref: nil,
			},
			[][]string{{"MIT"}}},
		{orExpression().name, orExpression().node, orExpression().sorted},
		{orAndExpression().name, orAndExpression().node, orAndExpression().sorted},
		{or_And_Expression().name, or_And_Expression().node, or_And_Expression().sorted},
		{andOrExpression().name, andOrExpression().node, andOrExpression().sorted},
		{orAndOrExpression().name, orAndOrExpression().node, orAndOrExpression().sorted},
		{or_And_OrExpression().name, or_And_OrExpression().node, or_And_OrExpression().sorted},
		{orOrOrExpression().name, orOrOrExpression().node, orOrOrExpression().sorted},
		{andOrAndExpression().name, andOrAndExpression().node, andOrAndExpression().sorted},
		{_and_Or_And_Expression().name, _and_Or_And_Expression().node, _and_Or_And_Expression().sorted},
		{andExpression().name, andExpression().node, andExpression().sorted},
		{and_Or_Expression().name, and_Or_Expression().node, and_Or_Expression().sorted},
		{_or_And_Or_Expression().name, _or_And_Or_Expression().node, _or_And_Or_Expression().sorted},
		{and_Or_AndExpression().name, and_Or_AndExpression().node, and_Or_AndExpression().sorted},
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
		node     *Node
		expanded [][]string
	}{
		{orExpression().name, orExpression().node, orExpression().expanded},
		{orAndExpression().name, orAndExpression().node, orAndExpression().expanded},
		{or_And_Expression().name, or_And_Expression().node, or_And_Expression().expanded},
		{andOrExpression().name, andOrExpression().node, andOrExpression().expanded},
		{orAndOrExpression().name, orAndOrExpression().node, orAndOrExpression().expanded},
		{or_And_OrExpression().name, or_And_OrExpression().node, or_And_OrExpression().expanded},
		{orOrOrExpression().name, orOrOrExpression().node, orOrOrExpression().expanded},
		{andOrAndExpression().name, andOrAndExpression().node, andOrAndExpression().expanded},
		{_and_Or_And_Expression().name, _and_Or_And_Expression().node, _and_Or_And_Expression().expanded},

		// TODO: Uncomment kitchen sink test when license plus, exception, license ref, and document ref are supported.
		// {"kitchen sink",
		// 	// "   (MIT AND Apache-1.0+)   OR   DocumentRef-spdx-tool-1.2:LicenseRef-MIT-Style-2 OR (GPL-2.0 WITH Bison-exception-2.2)",
		// 	&Node{
		// 		role: EXPRESSION_NODE,
		// 		exp: &expressionNodePartial{
		// 			left: &Node{
		// 				role: EXPRESSION_NODE,
		// 				exp: &expressionNodePartial{
		// 					left: &Node{
		// 						role: LICENSE_NODE,
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
		// 					right: &Node{
		// 						role: LICENSE_NODE,
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
		// 			right: &Node{
		// 				role: EXPRESSION_NODE,
		// 				exp: &expressionNodePartial{
		// 					left: &Node{
		// 						role: LICENSEREF_NODE,
		// 						exp:  nil,
		// 						lic:  nil,
		// 						ref: &referenceNodePartial{
		// 							hasDocumentRef: true,
		// 							documentRef:    "spdx-tool-1.2",
		// 							licenseRef:     "MIT-Style-2",
		// 						},
		// 					},
		// 					conjunction: "or",
		// 					right: &Node{
		// 						role: LICENSE_NODE,
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
		node     *Node
		expanded [][]string
	}{
		{andExpression().name, andExpression().node, andExpression().expanded},
		{and_Or_Expression().name, and_Or_Expression().node, and_Or_Expression().expanded},
		{_or_And_Or_Expression().name, _or_And_Or_Expression().node, _or_And_Or_Expression().expanded},
		{and_Or_AndExpression().name, and_Or_AndExpression().node, and_Or_AndExpression().expanded},
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
	node       *Node
	expanded   [][]string
	sorted     [][]string
	flattened  []string
}

func orExpression() testCaseData {
	return testCaseData{
		name:       "OR Expression",
		expression: "MIT OR Apache-2.0",
		node: &Node{
			role: EXPRESSION_NODE,
			exp: &expressionNodePartial{
				left: &Node{
					role: LICENSE_NODE,
					exp:  nil,
					lic: &licenseNodePartial{
						license: "MIT", hasPlus: false,
						hasException: false, exception: ""},
					ref: nil,
				},
				conjunction: "or",
				right: &Node{
					role: LICENSE_NODE,
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
		expanded:  [][]string{{"MIT"}, {"Apache-2.0"}},
		sorted:    [][]string{{"Apache-2.0"}, {"MIT"}},
		flattened: []string{"Apache-2.0", "MIT"},
	}
}

func orAndExpression() testCaseData {
	return testCaseData{
		name:       "OR-AND Expression",
		expression: "MIT OR Apache-2.0 AND GPL-2.0",
		node: &Node{
			role: EXPRESSION_NODE,
			exp: &expressionNodePartial{
				left: &Node{
					role: LICENSE_NODE,
					exp:  nil,
					lic: &licenseNodePartial{
						license: "MIT", hasPlus: false,
						hasException: false, exception: ""},
					ref: nil,
				},
				conjunction: "or",
				right: &Node{
					exp: &expressionNodePartial{
						left: &Node{
							role: LICENSE_NODE,
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
						right: &Node{
							role: LICENSE_NODE,
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
		expanded:  [][]string{{"MIT"}, {"Apache-2.0", "GPL-2.0"}},
		sorted:    [][]string{{"Apache-2.0", "GPL-2.0"}, {"MIT"}},
		flattened: []string{"Apache-2.0", "GPL-2.0", "MIT"},
	}
}

func or_And_Expression() testCaseData {
	return testCaseData{
		name:       "OR(AND) Expression",
		expression: "MIT OR (Apache-2.0 AND GPL-2.0)",
		node: &Node{
			role: EXPRESSION_NODE,
			exp: &expressionNodePartial{
				left: &Node{
					role: LICENSE_NODE,
					exp:  nil,
					lic: &licenseNodePartial{
						license: "MIT", hasPlus: false,
						hasException: false, exception: ""},
					ref: nil,
				},
				conjunction: "or",
				right: &Node{
					exp: &expressionNodePartial{
						left: &Node{
							role: LICENSE_NODE,
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
						right: &Node{
							role: LICENSE_NODE,
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
		expanded:  [][]string{{"MIT"}, {"Apache-2.0", "GPL-2.0"}},
		sorted:    [][]string{{"Apache-2.0", "GPL-2.0"}, {"MIT"}},
		flattened: []string{"Apache-2.0", "GPL-2.0", "MIT"},
	}
}

func andOrExpression() testCaseData {
	return testCaseData{
		name:       "AND-OR Expression",
		expression: "MIT AND Apache-2.0 OR GPL-2.0",
		node: &Node{
			role: EXPRESSION_NODE,
			exp: &expressionNodePartial{
				left: &Node{
					exp: &expressionNodePartial{
						left: &Node{
							role: LICENSE_NODE,
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
						right: &Node{
							role: LICENSE_NODE,
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
				right: &Node{
					role: LICENSE_NODE,
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
		expanded:  [][]string{{"MIT", "Apache-2.0"}, {"GPL-2.0"}},
		sorted:    [][]string{{"Apache-2.0", "MIT"}, {"GPL-2.0"}},
		flattened: []string{"Apache-2.0", "GPL-2.0", "MIT"},
	}
}

func orAndOrExpression() testCaseData {
	return testCaseData{
		name:       "OR-AND-OR Expression",
		expression: "MIT OR ISC AND Apache-2.0 OR GPL-2.0",
		node: &Node{
			role: EXPRESSION_NODE,
			exp: &expressionNodePartial{
				left: &Node{
					role: LICENSE_NODE,
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
				right: &Node{
					exp: &expressionNodePartial{
						left: &Node{
							role: EXPRESSION_NODE,
							exp: &expressionNodePartial{
								left: &Node{
									role: LICENSE_NODE,
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
								right: &Node{
									role: LICENSE_NODE,
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
						right: &Node{
							role: LICENSE_NODE,
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
		expanded:  [][]string{{"MIT"}, {"ISC", "Apache-2.0"}, {"GPL-2.0"}},
		sorted:    [][]string{{"Apache-2.0", "ISC"}, {"GPL-2.0"}, {"MIT"}},
		flattened: []string{"Apache-2.0", "GPL-2.0", "ISC", "MIT"},
	}
}

func or_And_OrExpression() testCaseData {
	return testCaseData{
		name:       "OR(AND)OR Expression",
		expression: "MIT OR (ISC AND Apache-2.0) OR GPL-2.0",
		node: &Node{
			role: EXPRESSION_NODE,
			exp: &expressionNodePartial{
				left: &Node{
					role: LICENSE_NODE,
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
				right: &Node{
					exp: &expressionNodePartial{
						left: &Node{
							role: EXPRESSION_NODE,
							exp: &expressionNodePartial{
								left: &Node{
									role: LICENSE_NODE,
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
								right: &Node{
									role: LICENSE_NODE,
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
						right: &Node{
							role: LICENSE_NODE,
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
		expanded:  [][]string{{"MIT"}, {"ISC", "Apache-2.0"}, {"GPL-2.0"}},
		sorted:    [][]string{{"Apache-2.0", "ISC"}, {"GPL-2.0"}, {"MIT"}},
		flattened: []string{"Apache-2.0", "GPL-2.0", "ISC", "MIT"},
	}
}

func orOrOrExpression() testCaseData {
	return testCaseData{
		name:       "OR-OR-OR Expression",
		expression: "MIT OR ISC OR Apache-2.0 OR GPL-2.0",
		node: &Node{
			role: EXPRESSION_NODE,
			exp: &expressionNodePartial{
				left: &Node{
					role: LICENSE_NODE,
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
				right: &Node{
					exp: &expressionNodePartial{
						left: &Node{
							role: LICENSE_NODE,
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
						right: &Node{
							exp: &expressionNodePartial{
								left: &Node{
									role: LICENSE_NODE,
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
								right: &Node{
									role: LICENSE_NODE,
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
		expanded:  [][]string{{"MIT"}, {"ISC"}, {"Apache-2.0"}, {"GPL-2.0"}},
		sorted:    [][]string{{"Apache-2.0"}, {"GPL-2.0"}, {"ISC"}, {"MIT"}},
		flattened: []string{"Apache-2.0", "GPL-2.0", "ISC", "MIT"},
	}
}

func andOrAndExpression() testCaseData {
	return testCaseData{
		name:       "AND-OR-AND Expression",
		expression: "MIT AND ISC OR Apache-2.0 AND GPL-2.0",
		node: &Node{
			role: EXPRESSION_NODE,
			exp: &expressionNodePartial{
				left: &Node{
					exp: &expressionNodePartial{
						left: &Node{
							role: LICENSE_NODE,
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
						right: &Node{
							role: LICENSE_NODE,
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
				right: &Node{
					exp: &expressionNodePartial{
						left: &Node{
							role: LICENSE_NODE,
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
						right: &Node{
							role: LICENSE_NODE,
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
		expanded:  [][]string{{"MIT", "ISC"}, {"Apache-2.0", "GPL-2.0"}},
		sorted:    [][]string{{"Apache-2.0", "GPL-2.0"}, {"ISC", "MIT"}},
		flattened: []string{"Apache-2.0", "GPL-2.0", "ISC", "MIT"},
	}
}

func _and_Or_And_Expression() testCaseData {
	return testCaseData{
		name:       "(AND)OR(AND) Expression",
		expression: "(MIT AND ISC) OR (Apache-2.0 AND GPL-2.0)",
		node: &Node{
			role: EXPRESSION_NODE,
			exp: &expressionNodePartial{
				left: &Node{
					exp: &expressionNodePartial{
						left: &Node{
							role: LICENSE_NODE,
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
						right: &Node{
							role: LICENSE_NODE,
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
				right: &Node{
					exp: &expressionNodePartial{
						left: &Node{
							role: LICENSE_NODE,
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
						right: &Node{
							role: LICENSE_NODE,
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
		expanded:  [][]string{{"MIT", "ISC"}, {"Apache-2.0", "GPL-2.0"}},
		sorted:    [][]string{{"Apache-2.0", "GPL-2.0"}, {"ISC", "MIT"}},
		flattened: []string{"Apache-2.0", "GPL-2.0", "ISC", "MIT"},
	}
}

func andExpression() testCaseData {
	return testCaseData{
		name:       "AND Expression",
		expression: "MIT AND Apache-2.0",
		node: &Node{
			role: EXPRESSION_NODE,
			exp: &expressionNodePartial{
				left: &Node{
					role: LICENSE_NODE,
					exp:  nil,
					lic: &licenseNodePartial{
						license: "MIT", hasPlus: false,
						hasException: false, exception: ""},
					ref: nil,
				},
				conjunction: "and",
				right: &Node{
					role: LICENSE_NODE,
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
		expanded:  [][]string{{"MIT", "Apache-2.0"}},
		sorted:    [][]string{{"Apache-2.0", "MIT"}},
		flattened: []string{"Apache-2.0", "MIT"},
	}
}

func and_Or_Expression() testCaseData {
	return testCaseData{
		name:       "AND(OR) Expression",
		expression: "MIT AND (Apache-2.0 OR GPL-2.0)",
		node: &Node{
			role: EXPRESSION_NODE,
			exp: &expressionNodePartial{
				left: &Node{
					role: LICENSE_NODE,
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
				right: &Node{
					exp: &expressionNodePartial{
						left: &Node{
							role: LICENSE_NODE,
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
						right: &Node{
							role: LICENSE_NODE,
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
		expanded:  [][]string{{"MIT", "Apache-2.0"}, {"MIT", "GPL-2.0"}},
		sorted:    [][]string{{"Apache-2.0", "MIT"}, {"GPL-2.0", "MIT"}},
		flattened: []string{"Apache-2.0", "GPL-2.0", "MIT"},
	}
}

func _or_And_Or_Expression() testCaseData {
	return testCaseData{
		name:       "(OR)AND(OR) Expression",
		expression: "(MIT OR ISC) AND (Apache-2.0 OR GPL-2.0)",
		node: &Node{
			role: EXPRESSION_NODE,
			exp: &expressionNodePartial{
				left: &Node{
					exp: &expressionNodePartial{
						left: &Node{
							role: LICENSE_NODE,
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
						right: &Node{
							role: LICENSE_NODE,
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
				right: &Node{
					exp: &expressionNodePartial{
						left: &Node{
							role: LICENSE_NODE,
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
						right: &Node{
							role: LICENSE_NODE,
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
		expanded:  [][]string{{"MIT", "Apache-2.0"}, {"ISC", "Apache-2.0"}, {"MIT", "GPL-2.0"}, {"ISC", "GPL-2.0"}},
		sorted:    [][]string{{"Apache-2.0", "ISC"}, {"Apache-2.0", "MIT"}, {"GPL-2.0", "ISC"}, {"GPL-2.0", "MIT"}},
		flattened: []string{"Apache-2.0", "GPL-2.0", "ISC", "MIT"},
	}
}

func and_Or_AndExpression() testCaseData {
	return testCaseData{
		name:       "AND(OR)AND Expression",
		expression: "MIT AND (ISC OR Apache-2.0) AND GPL-2.0",
		node: &Node{
			role: EXPRESSION_NODE,
			exp: &expressionNodePartial{
				left: &Node{
					role: LICENSE_NODE,
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
				right: &Node{
					role: EXPRESSION_NODE,
					exp: &expressionNodePartial{
						left: &Node{
							role: EXPRESSION_NODE,
							exp: &expressionNodePartial{
								left: &Node{
									role: LICENSE_NODE,
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
								right: &Node{
									role: LICENSE_NODE,
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
						right: &Node{
							role: LICENSE_NODE,
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
		expanded:  [][]string{{"MIT", "ISC", "GPL-2.0"}, {"MIT", "Apache-2.0", "GPL-2.0"}},
		sorted:    [][]string{{"Apache-2.0", "GPL-2.0", "MIT"}, {"GPL-2.0", "ISC", "MIT"}},
		flattened: []string{"Apache-2.0", "GPL-2.0", "ISC", "MIT"},
	}
}

func andAndAndExpression() testCaseData {
	return testCaseData{
		name:       "AND-AND-AND Expression",
		expression: "MIT AND ISC AND Apache-2.0 AND GPL-2.0",
		node: &Node{
			role: EXPRESSION_NODE,
			exp: &expressionNodePartial{
				left: &Node{
					role: LICENSE_NODE,
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
				right: &Node{
					role: EXPRESSION_NODE,
					exp: &expressionNodePartial{
						left: &Node{
							role: LICENSE_NODE,
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
						right: &Node{
							role: EXPRESSION_NODE,
							exp: &expressionNodePartial{
								left: &Node{
									role: LICENSE_NODE,
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
								right: &Node{
									role: LICENSE_NODE,
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
		expanded:  [][]string{{"MIT", "ISC", "Apache-2.0", "GPL-2.0"}},
		sorted:    [][]string{{"Apache-2.0", "GPL-2.0", "ISC", "MIT"}},
		flattened: []string{"Apache-2.0", "GPL-2.0", "ISC", "MIT"},
	}
}
