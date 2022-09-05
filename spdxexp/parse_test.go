package spdxexp

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParse(t *testing.T) {
	tests := []struct {
		name       string
		expression string
		node       *Node
		nodestr    string
		err        error
	}{
		{"single license",
			"MIT",
			&Node{
				role: LICENSE_NODE,
				exp:  nil,
				lic: &licenseNodePartial{
					license: "MIT", hasPlus: false,
					hasException: false, exception: ""},
				ref: nil,
			},
			"MIT", nil},
		{"two licenses using AND",
			"MIT AND Apache-2.0",
			&Node{
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
			"{ LEFT: MIT and RIGHT: Apache-2.0 }", nil},
		{"two licenses using OR",
			"MIT OR Apache-2.0",
			&Node{
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
			"{ LEFT: MIT or RIGHT: Apache-2.0 }", nil},
		{"kitchen sink",
			"   (MIT AND Apache-1.0+)   OR   DocumentRef-spdx-tool-1.2:LicenseRef-MIT-Style-2 OR (GPL-2.0 WITH Bison-exception-2.2)",
			&Node{
				role: EXPRESSION_NODE,
				exp: &expressionNodePartial{
					left: &Node{
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
								role: LICENSE_NODE,
								exp:  nil,
								lic: &licenseNodePartial{
									license:      "Apache-1.0",
									hasPlus:      true,
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
						role: EXPRESSION_NODE,
						exp: &expressionNodePartial{
							left: &Node{
								role: LICENSEREF_NODE,
								exp:  nil,
								lic:  nil,
								ref: &referenceNodePartial{
									hasDocumentRef: true,
									documentRef:    "spdx-tool-1.2",
									licenseRef:     "MIT-Style-2",
								},
							},
							conjunction: "or",
							right: &Node{
								role: LICENSE_NODE,
								exp:  nil,
								lic: &licenseNodePartial{
									license:      "GPL-2.0",
									hasPlus:      false,
									hasException: true,
									exception:    "Bison-exception-2.2",
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
			"{ LEFT: { LEFT: MIT and RIGHT: Apache-1.0+ } or RIGHT: { LEFT: DocumentRef-spdx-tool-1.2:LicenseRef-MIT-Style-2 or RIGHT: GPL-2.0 with Bison-exception-2.2 } }", nil,
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			startNode, err := Parse(test.expression)

			require.Equal(t, test.err, err)
			if test.err != nil {
				// when error, check that returned node is nil
				var nilNode *Node = nil
				assert.Equal(t, nilNode, startNode, "Expected nil node when error occurs.")
				return
			}

			// ref found, check token values are as expected
			assert.Equal(t, test.node, startNode)
			assert.Equal(t, test.nodestr, startNode.String())
		})
	}
}

func TestParseTokens(t *testing.T) {
	tests := []struct {
		name    string
		tokens  *tokenStream
		node    *Node
		nodestr string
		err     error
	}{
		{"single license",
			getLicenseTokens(0),
			&Node{
				role: LICENSE_NODE,
				exp:  nil,
				lic: &licenseNodePartial{
					license: "MIT", hasPlus: false,
					hasException: false, exception: ""},
				ref: nil,
			},
			"MIT", nil},
		{"two licenses using AND",
			getAndClauseTokens(0),
			&Node{
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
			"{ LEFT: MIT and RIGHT: Apache-2.0 }", nil},
		{"two licenses using OR",
			getOrClauseTokens(0),
			&Node{
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
			"{ LEFT: MIT or RIGHT: Apache-2.0 }", nil},
		{"kitchen sink",
			getKitchSinkTokens(0),
			&Node{
				role: EXPRESSION_NODE,
				exp: &expressionNodePartial{
					left: &Node{
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
								role: LICENSE_NODE,
								exp:  nil,
								lic: &licenseNodePartial{
									license:      "Apache-1.0",
									hasPlus:      true,
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
						role: EXPRESSION_NODE,
						exp: &expressionNodePartial{
							left: &Node{
								role: LICENSEREF_NODE,
								exp:  nil,
								lic:  nil,
								ref: &referenceNodePartial{
									hasDocumentRef: true,
									documentRef:    "spdx-tool-1.2",
									licenseRef:     "MIT-Style-2",
								},
							},
							conjunction: "or",
							right: &Node{
								role: LICENSE_NODE,
								exp:  nil,
								lic: &licenseNodePartial{
									license:      "GPL-2.0",
									hasPlus:      false,
									hasException: true,
									exception:    "Bison-exception-2.2",
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
			"{ LEFT: { LEFT: MIT and RIGHT: Apache-1.0+ } or RIGHT: { LEFT: DocumentRef-spdx-tool-1.2:LicenseRef-MIT-Style-2 or RIGHT: GPL-2.0 with Bison-exception-2.2 } }", nil,
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			startNode := test.tokens.parseTokens()

			require.Equal(t, test.err, test.tokens.err)
			if test.err != nil {
				// when error, check that returned node is nil
				var nilNode *Node = nil
				assert.Equal(t, nilNode, startNode, "Expected nil node when error occurs.")
				return
			}

			// ref found, check token values are as expected
			assert.Equal(t, test.node, startNode)
			assert.Equal(t, test.nodestr, startNode.String())
		})
	}
}

func TestHasMoreTokens(t *testing.T) {
	tests := []struct {
		name   string
		tokens *tokenStream
		result bool
	}{
		{"at start", getAndClauseTokens(0), true},
		{"at middle", getAndClauseTokens(1), true},
		{"at end", getAndClauseTokens(2), true},
		{"past end", getAndClauseTokens(3), false},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.result, test.tokens.hasMore())
		})
	}
}

func TestPeek(t *testing.T) {
	tests := []struct {
		name   string
		tokens *tokenStream
		token  *token
	}{
		{"at start", getAndClauseTokens(0), &(token{role: LICENSE_TOKEN, value: "MIT"})},
		{"at middle", getAndClauseTokens(1), &(token{role: OPERATOR_TOKEN, value: "AND"})},
		{"at end", getAndClauseTokens(2), &(token{role: LICENSE_TOKEN, value: "Apache-2.0"})},
		{"past end", getAndClauseTokens(3), nil},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.token, test.tokens.peek())
		})
	}
}

func TestNext(t *testing.T) {
	tests := []struct {
		name     string
		tokens   *tokenStream
		newIndex int
		err      error
	}{
		{"at start", getAndClauseTokens(0), 1, nil},
		{"at middle", getAndClauseTokens(1), 2, nil},
		{"at end", getAndClauseTokens(2), 3, nil},
		{"past end", getAndClauseTokens(3), 3, errors.New("read past end of tokens")},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			test.tokens.next()
			assert.Equal(t, test.newIndex, test.tokens.index)
			require.Equal(t, test.err, test.tokens.err)
		})
	}
}

// // TODO: func parseParenthesizedExpression(tokens *[]token, index int) (*Node, int, error) {
// // TODO: func parseAtom(tokens *[]token, index int) (*Node, int, error) {
// // TODO: func parseExpression(tokens *[]token, index int) (*Node, int, error) {
// // TODO: func parseAnd(tokens *[]token, index int) (*Node, int, error) {
// // TODO: func parseLicenseRef(tokens *[]token, index int) (*Node, int, error) {
// // TODO: func parseLicense(tokens *[]token, index int) (*Node, int, error) {

func TestParseOperator(t *testing.T) {
	tests := []struct {
		name      string
		tokens    *tokenStream
		operator  string
		expectNil bool
		newIndex  int
	}{
		{"looking for WITH operator", getWithClauseTokens(1), "WITH", false, 2},
		{"looking for AND operator", getAndClauseTokens(1), "AND", false, 2},
		{"looking for OR operator", getOrClauseTokens(1), "OR", false, 2},
		{"looking for ( operator", getOrAndClauseTokens(2), "(", false, 3},
		{"looking for ) operator", getOrAndClauseTokens(6), ")", false, 7},
		{"looking for : operator", getColonClauseTokens(1), ":", false, 2},
		{"looking for + operator", getPlusClauseTokens(1), "+", false, 2},
		{"looking for OR operator, but got AND", getAndClauseTokens(1), "OR", true, 1},
		{"looking for OR operator, but got LICENSE", getOrClauseTokens(0), "OR", true, 0},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			token := test.tokens.parseOperator(test.operator)
			require.Equal(t, test.newIndex, test.tokens.index)
			if test.expectNil {
				// returned token is nil if it isn't an operator or it is a different operator
				var nil_token *string
				assert.Equal(t, nil_token, token)
			} else {
				// index advances when token is the expected operator
				assert.Equal(t, test.operator, *token)
			}
		})
	}
}

// func parseWith(tokens *[]token, index int) (*string, int, error) {
func TestParseWith(t *testing.T) {
	tests := []struct {
		name      string
		tokens    *tokenStream
		exception string
		expectNil bool
		newIndex  int
		err       error
	}{
		{"WITH followed by EXCEPTION", getWithClauseTokens(1), "Bison-exception-2.2", false, 2, nil},
		{"WITH not followed by EXCEPTION", getInvalidWithClauseTokens(1), "", true, 2, errors.New("expected exception after 'WITH'")},
		{"not with", getOrClauseTokens(1), "", true, 1, nil},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			exceptionLicense := test.tokens.parseWith()
			assert.Equal(t, test.newIndex, test.tokens.index)

			require.Equal(t, test.err, test.tokens.err)
			if test.expectNil {
				// exception license is nil when error occurs or WITH operator is not found
				var nilString *string = nil
				assert.Equal(t, nilString, exceptionLicense)
				return
			}

			// WITH found, check exceptionLicense value
			assert.Equal(t, test.exception, *exceptionLicense)
		})
	}
}

// TODO: func (n *Node) String() string {

func getLicenseTokens(index int) *tokenStream {
	var tokens []token
	tokens = append(tokens, token{role: LICENSE_TOKEN, value: "MIT"})
	return getTokenStream(tokens, index)
}

func getWithClauseTokens(index int) *tokenStream {
	var tokens []token
	tokens = append(tokens, token{role: LICENSE_TOKEN, value: "MIT"})
	tokens = append(tokens, token{role: OPERATOR_TOKEN, value: "WITH"})
	tokens = append(tokens, token{role: EXCEPTION_TOKEN, value: "Bison-exception-2.2"})
	return getTokenStream(tokens, index)
}

func getInvalidWithClauseTokens(index int) *tokenStream {
	var tokens []token
	tokens = append(tokens, token{role: LICENSE_TOKEN, value: "MIT"})
	tokens = append(tokens, token{role: OPERATOR_TOKEN, value: "WITH"})
	tokens = append(tokens, token{role: LICENSE_TOKEN, value: "Apache-2.0"})
	return getTokenStream(tokens, index)
}

func getAndClauseTokens(index int) *tokenStream {
	var tokens []token
	tokens = append(tokens, token{role: LICENSE_TOKEN, value: "MIT"})
	tokens = append(tokens, token{role: OPERATOR_TOKEN, value: "AND"})
	tokens = append(tokens, token{role: LICENSE_TOKEN, value: "Apache-2.0"})
	return getTokenStream(tokens, index)
}

func getOrClauseTokens(index int) *tokenStream {
	var tokens []token
	tokens = append(tokens, token{role: LICENSE_TOKEN, value: "MIT"})
	tokens = append(tokens, token{role: OPERATOR_TOKEN, value: "OR"})
	tokens = append(tokens, token{role: LICENSE_TOKEN, value: "Apache-2.0"})
	return getTokenStream(tokens, index)
}

func getOrAndClauseTokens(index int) *tokenStream {
	var tokens []token
	tokens = append(tokens, token{role: LICENSE_TOKEN, value: "Apache-2.0"})
	tokens = append(tokens, token{role: OPERATOR_TOKEN, value: "OR"})
	tokens = append(tokens, token{role: OPERATOR_TOKEN, value: "("})
	tokens = append(tokens, token{role: LICENSE_TOKEN, value: "MIT"})
	tokens = append(tokens, token{role: OPERATOR_TOKEN, value: "AND"})
	tokens = append(tokens, token{role: LICENSE_TOKEN, value: "Apache 2.0"})
	tokens = append(tokens, token{role: OPERATOR_TOKEN, value: ")"})
	return getTokenStream(tokens, index)
}

func getColonClauseTokens(index int) *tokenStream {
	var tokens []token
	tokens = append(tokens, token{role: DOCUMENTREF_TOKEN, value: "spdx-tool-1.2"})
	tokens = append(tokens, token{role: OPERATOR_TOKEN, value: ":"})
	tokens = append(tokens, token{role: LICENSEREF_TOKEN, value: "MIT-Style-2"})
	return getTokenStream(tokens, index)
}

func getPlusClauseTokens(index int) *tokenStream {
	var tokens []token
	tokens = append(tokens, token{role: LICENSE_TOKEN, value: "Apache-1.0"})
	tokens = append(tokens, token{role: OPERATOR_TOKEN, value: "+"})
	tokens = append(tokens, token{role: OPERATOR_TOKEN, value: "OR"})
	tokens = append(tokens, token{role: LICENSE_TOKEN, value: "MIT"})
	return getTokenStream(tokens, index)
}

func getKitchSinkTokens(index int) *tokenStream {
	var tokens []token
	tokens = append(tokens, token{role: OPERATOR_TOKEN, value: "("})
	tokens = append(tokens, token{role: LICENSE_TOKEN, value: "MIT"})
	tokens = append(tokens, token{role: OPERATOR_TOKEN, value: "AND"})
	tokens = append(tokens, token{role: LICENSE_TOKEN, value: "Apache-1.0"})
	tokens = append(tokens, token{role: OPERATOR_TOKEN, value: "+"})
	tokens = append(tokens, token{role: OPERATOR_TOKEN, value: ")"})
	tokens = append(tokens, token{role: OPERATOR_TOKEN, value: "OR"})
	tokens = append(tokens, token{role: DOCUMENTREF_TOKEN, value: "spdx-tool-1.2"})
	tokens = append(tokens, token{role: OPERATOR_TOKEN, value: ":"})
	tokens = append(tokens, token{role: LICENSEREF_TOKEN, value: "MIT-Style-2"})
	tokens = append(tokens, token{role: OPERATOR_TOKEN, value: "OR"})
	tokens = append(tokens, token{role: OPERATOR_TOKEN, value: "("})
	tokens = append(tokens, token{role: LICENSE_TOKEN, value: "GPL-2.0"})
	tokens = append(tokens, token{role: OPERATOR_TOKEN, value: "WITH"})
	tokens = append(tokens, token{role: EXCEPTION_TOKEN, value: "Bison-exception-2.2"})
	tokens = append(tokens, token{role: OPERATOR_TOKEN, value: ")"})
	return getTokenStream(tokens, index)
}

func getTokenStream(tokens []token, index int) *tokenStream {
	return &tokenStream{
		tokens: tokens,
		index:  index,
		err:    nil,
	}
}
