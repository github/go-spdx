package spdxexp

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestScan(t *testing.T) {
	tests := []struct {
		name       string
		expression string
		tokens     []token
		err        error
	}{
		{"single license", "MIT",
			[]token{
				{role: LICENSE_TOKEN, value: "MIT"},
			}, nil},
		{"two licenses using AND", "MIT AND Apache-2.0",
			[]token{
				{role: LICENSE_TOKEN, value: "MIT"},
				{role: OPERATOR_TOKEN, value: "AND"},
				{role: LICENSE_TOKEN, value: "Apache-2.0"},
			}, nil},
		{"two licenses using OR inside paren", "(MIT OR Apache-2.0)",
			[]token{
				{role: OPERATOR_TOKEN, value: "("},
				{role: LICENSE_TOKEN, value: "MIT"},
				{role: OPERATOR_TOKEN, value: "OR"},
				{role: LICENSE_TOKEN, value: "Apache-2.0"},
				{role: OPERATOR_TOKEN, value: ")"},
			}, nil},
		{"kitchen sink", "   (MIT AND Apache-1.0+)   OR   DocumentRef-spdx-tool-1.2:LicenseRef-MIT-Style-2 OR (GPL-2.0 WITH Bison-exception-2.2)",
			[]token{
				{role: OPERATOR_TOKEN, value: "("},
				{role: LICENSE_TOKEN, value: "MIT"},
				{role: OPERATOR_TOKEN, value: "AND"},
				{role: LICENSE_TOKEN, value: "Apache-1.0"},
				{role: OPERATOR_TOKEN, value: "+"},
				{role: OPERATOR_TOKEN, value: ")"},
				{role: OPERATOR_TOKEN, value: "OR"},
				{role: DOCUMENTREF_TOKEN, value: "spdx-tool-1.2"},
				{role: OPERATOR_TOKEN, value: ":"},
				{role: LICENSEREF_TOKEN, value: "MIT-Style-2"},
				{role: OPERATOR_TOKEN, value: "OR"},
				{role: OPERATOR_TOKEN, value: "("},
				{role: LICENSE_TOKEN, value: "GPL-2.0"},
				{role: OPERATOR_TOKEN, value: "WITH"},
				{role: EXCEPTION_TOKEN, value: "Bison-exception-2.2"},
				{role: OPERATOR_TOKEN, value: ")"},
			}, nil},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			tokens, err := scan(test.expression)

			require.Equal(t, test.err, err)
			if test.err != nil {
				// tokens should be nil when error occurs
				var nilTokens *[]token = nil
				assert.Equal(t, nilTokens, tokens, "Expected nil token array when error occurs.")
				return
			}

			// scan completed, check tokens
			assert.Equal(t, test.tokens, tokens)
		})
	}
}

func TestHasMoreSource(t *testing.T) {
	tests := []struct {
		name   string
		exp    *expressionStream
		result bool
	}{
		{"at start", getExpressionStream("MIT OR Apache-2.0", 0), true},
		{"at middle", getExpressionStream("MIT OR Apache-2.0", 3), true},
		{"at end", getExpressionStream("MIT OR Apache-2.0", len("MIT OR Apache-2.0")), false},
		{"past end", getExpressionStream("MIT OR Apache-2.0", 50), false},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.result, test.exp.hasMore())
		})
	}
}

func TestParseToken(t *testing.T) {
	tests := []struct {
		name     string
		exp      *expressionStream
		token    *token
		newIndex int
		err      error
	}{
		{"operator found", getExpressionStream("MIT AND Apache-2.0", 4),
			&token{role: OPERATOR_TOKEN, value: "AND"}, 7, nil},
		{"operator error", getExpressionStream("Apache-1.0 + OR MIT", 11),
			nil, 11, errors.New("unexpected space before +")},
		{"document ref found", getExpressionStream("DocumentRef-spdx-tool-1.2:LicenseRef-MIT-Style-2", 0),
			&token{role: DOCUMENTREF_TOKEN, value: "spdx-tool-1.2"}, 25, nil},
		{"document ref error", getExpressionStream("DocumentRef-!23", 0),
			nil, 12, errors.New("expected id at offset 12")},
		{"license ref found", getExpressionStream("DocumentRef-spdx-tool-1.2:LicenseRef-MIT-Style-2", 26),
			&token{role: LICENSEREF_TOKEN, value: "MIT-Style-2"}, 48, nil},
		{"license ref error", getExpressionStream("LicenseRef-!23", 0),
			nil, 11, errors.New("expected id at offset 11")},
		{"identifier found", getExpressionStream("MIT AND Apache-2.0", 8),
			&token{role: LICENSE_TOKEN, value: "Apache-2.0"}, 18, nil},
		{"identifier error", getExpressionStream("NON-EXISTENT-LICENSE", 0),
			nil, 0, errors.New("unexpected 'N' at offset 0")},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			tokn := test.exp.parseToken()
			assert.Equal(t, test.newIndex, test.exp.index)

			require.Equal(t, test.err, test.exp.err)
			if test.err != nil {
				// token is nil when error occurs or token is not recognized
				var nilToken *token = nil
				assert.Equal(t, nilToken, tokn)
				return
			}

			// token recognized, check token value
			assert.Equal(t, test.token, tokn)
		})
	}
}

func TestReadRegex(t *testing.T) {
	tests := []struct {
		name     string
		exp      *expressionStream
		pattern  string
		match    string
		newIndex int
	}{
		{"regex to skip leading blank in middle", getExpressionStream("MIT OR Apache-2.0", 3),
			"[ ]*", " ", 4},
		{"regex for id", getExpressionStream("LicenseRef-MIT-Style-1", 11),
			"[A-Za-z0-9-.]+", "MIT-Style-1", 22},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			match := test.exp.readRegex(test.pattern)
			assert.Equal(t, test.match, match)
			assert.Equal(t, test.newIndex, test.exp.index)
		})
	}
}

func TestRead(t *testing.T) {
	tests := []struct {
		name     string
		exp      *expressionStream
		next     string
		match    string
		newIndex int
	}{
		{"at first - match word", getExpressionStream("MIT OR Apache-2.0", 0), "MIT", "MIT", 3},
		{"at middle - match operator", getExpressionStream("MIT OR Apache-2.0", 4), "OR", "OR", 6},
		{"at middle - match last word", getExpressionStream("MIT OR Apache-2.0", 7), "Apache-2.0", "Apache-2.0", 17},
		{"at first - no match", getExpressionStream("MIT OR Apache-2.0", 0), "GPL", "", 0},
		{"at middle - no match for operator", getExpressionStream("MIT OR Apache-2.0", 4), "AND", "", 4},
		{"at middle - no match last word", getExpressionStream("MIT OR Apache-2.0", 7), "GPL", "", 7},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			match := test.exp.read(test.next)
			assert.Equal(t, test.match, match)
			assert.Equal(t, test.newIndex, test.exp.index)
		})
	}
}

func TestSkipWhitespace(t *testing.T) {
	tests := []struct {
		name     string
		exp      *expressionStream
		newIndex int
	}{
		{"at first - no blanks", getExpressionStream("MIT OR Apache-2.0", 0), 0},
		{"at first - with blanks", getExpressionStream("  MIT OR Apache-2 .0", 0), 2},
		{"at middle - no blanks", getExpressionStream("MIT OR Apache-2.0", 4), 4},
		{"at middle - with blanks", getExpressionStream("MIT OR Apache-2.0", 3), 4},
		{"at end - no blanks", getExpressionStream("MIT OR GPL", 10), 10},
		{"at end - with blanks", getExpressionStream("MIT OR GPL  ", 10), 12},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			test.exp.skipWhitespace()
			assert.Equal(t, test.newIndex, test.exp.index)
		})
	}
}

func TestReadOperator(t *testing.T) {
	tests := []struct {
		name     string
		exp      *expressionStream
		operator *token
		newIndex int
		err      error
	}{
		{"WITH operator", getExpressionStream("MIT WITH Bison-exception-2.2", 4),
			&token{role: OPERATOR_TOKEN, value: "WITH"}, 8, nil},
		{"AND operator", getExpressionStream("MIT AND Apache-2.0", 4),
			&token{role: OPERATOR_TOKEN, value: "AND"}, 7, nil},
		{"OR operator", getExpressionStream("MIT OR Apache-2.0", 4),
			&token{role: OPERATOR_TOKEN, value: "OR"}, 6, nil},
		{"( operator", getExpressionStream("(MIT OR Apache-2.0)", 0),
			&token{role: OPERATOR_TOKEN, value: "("}, 1, nil},
		{") operator", getExpressionStream("(MIT OR Apache-2.0)", 18),
			&token{role: OPERATOR_TOKEN, value: ")"}, 19, nil},
		{": operator", getExpressionStream("DocumentRef-spdx-tool-1.2:LicenseRef-MIT-Style-2", 25),
			&token{role: OPERATOR_TOKEN, value: ":"}, 26, nil},
		{"plus operator - correctly used", getExpressionStream("Apache-1.0+ OR MIT", 10),
			&token{role: OPERATOR_TOKEN, value: "+"}, 11, nil},
		{"plus operator - with preceding space", getExpressionStream("Apache-1.0 + OR MIT", 11),
			nil, 11, errors.New("unexpected space before +")},
		{"operator not found", getExpressionStream("MIT AND Apache-2.0", 8),
			nil, 8, nil},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			operator := test.exp.readOperator()
			assert.Equal(t, test.newIndex, test.exp.index)
			require.Equal(t, test.err, test.exp.err)
			assert.Equal(t, test.operator, operator)
		})
	}
}

func TestReadId(t *testing.T) {
	tests := []struct {
		name     string
		exp      *expressionStream
		id       string
		newIndex int
	}{
		{"valid numeric id", getExpressionStream("LicenseRef-23", 11), "23", 13},
		{"valid id with dashes", getExpressionStream("LicenseRef-MIT-Style-1", 11), "MIT-Style-1", 22},
		{"valid id with period", getExpressionStream("DocumentRef-spdx-tool-1.2:LicenseRef-MIT-Style-2", 12), "spdx-tool-1.2", 25},
		{"invalid starts with non-supported character", getExpressionStream("LicenseRef-!23", 11), "", 11},
		{"invalid non-supported character in middle", getExpressionStream("LicenseRef-2!3", 11), "2", 12},
		{"invalid ends with non-supported character", getExpressionStream("LicenseRef-23!", 11), "23", 13},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			id := test.exp.readID()
			// check values if there isn't an error
			assert.Equal(t, test.id, id)
			assert.Equal(t, test.newIndex, test.exp.index)
		})
	}
}

func TestReadDocumentRef(t *testing.T) {
	tests := []struct {
		name     string
		exp      *expressionStream
		ref      *token
		newIndex int
		err      error
	}{
		{"valid document ref with id", getExpressionStream("DocumentRef-spdx-tool-1.2:LicenseRef-MIT-Style-2", 0), &token{role: DOCUMENTREF_TOKEN, value: "spdx-tool-1.2"}, 25, nil},
		{"document ref not found", getExpressionStream("DocumentRef-spdx-tool-1.2:LicenseRef-MIT-Style-2", 26), nil, 26, nil},
		{"invalid document ref with bad id", getExpressionStream("DocumentRef-!23", 0), nil, 12, errors.New("expected id at offset 12")},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			ref := test.exp.readDocumentRef()
			assert.Equal(t, test.newIndex, test.exp.index)

			require.Equal(t, test.err, test.exp.err)
			if test.err != nil {
				// ref should be nil when error occurs or a ref is not found
				var nilToken *token = nil
				assert.Equal(t, nilToken, ref, "Expected nil token when error occurs.")
				return
			}

			// ref found, check ref value
			assert.Equal(t, test.ref, ref)
		})
	}
}

func TestReadLicenseRef(t *testing.T) {
	tests := []struct {
		name     string
		exp      *expressionStream
		ref      *token
		newIndex int
		err      error
	}{
		{"valid license ref with id", getExpressionStream("DocumentRef-spdx-tool-1.2:LicenseRef-MIT-Style-2", 26), &token{role: LICENSEREF_TOKEN, value: "MIT-Style-2"}, 48, nil},
		{"license ref not found", getExpressionStream("DocumentRef-spdx-tool-1.2:LicenseRef-MIT-Style-2", 0), nil, 0, nil},
		{"invalid license ref with bad id", getExpressionStream("LicenseRef-!23", 0), nil, 11, errors.New("expected id at offset 11")},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			ref := test.exp.readLicenseRef()
			assert.Equal(t, test.newIndex, test.exp.index)

			require.Equal(t, test.err, test.exp.err)
			if test.err != nil {
				// ref should be nil when error occurs or a ref is not found
				var nilToken *token = nil
				assert.Equal(t, nilToken, ref)
				return
			}

			// ref found, check ref value
			assert.Equal(t, test.ref, ref)
		})
	}
}

func TestReadIdentifier(t *testing.T) {
	tests := []struct {
		name     string
		exp      *expressionStream
		license  *token
		newIndex int
		err      error
	}{
		{"active license", getExpressionStream("MIT", 0), &token{role: LICENSE_TOKEN, value: "MIT"}, 3, nil},
		{"deprecated license", getExpressionStream("LGPL-2.1", 0), &token{role: LICENSE_TOKEN, value: "LGPL-2.1"}, 8, nil},
		{"exception license", getExpressionStream("GPL-CC-1.0", 0), &token{role: EXCEPTION_TOKEN, value: "GPL-CC-1.0"}, 10, nil},
		{"invalid license", getExpressionStream("NON-EXISTENT-LICENSE", 0), nil, 0, nil}, // TODO: should this return an error?
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			license := test.exp.readIdentifier()
			assert.Equal(t, test.newIndex, test.exp.index)

			require.Equal(t, test.err, test.exp.err)
			if test.err != nil {
				// license should be nil when error occurs or a license is not found
				var nilToken *token = nil
				assert.Equal(t, nilToken, license)
				return
			}

			// license found, check license value
			assert.Equal(t, test.license, license)
		})
	}
}

func getExpressionStream(expression string, index int) *expressionStream {
	return &expressionStream{
		expression: expression,
		index:      index,
		err:        nil,
	}
}
