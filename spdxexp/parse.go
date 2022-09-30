package spdxexp

import (
	"errors"
	"strings"
)

// The ABNF grammar in the spec is totally ambiguous.
//
// This parser follows the operator precedence defined in the
// `Order of Precedence and Parentheses` section.

type tokenStream struct {
	tokens []token
	index  int
	err    error
}

func parse(source string, options *Options) (*node, error) {
	options = processOptions(options)

	if len(source) == 0 {
		return nil, errors.New("parse error - cannot parse empty string")
	}
	tokens, err := scan(source, options)
	if err != nil {
		return nil, err
	}
	tokns := &tokenStream{tokens: tokens, index: 0, err: nil}
	return tokns.parseTokens(), tokns.err
}

func (t *tokenStream) parseTokens() *node {
	node := t.parseExpression()
	if t.err != nil {
		return nil
	}
	if node == nil || t.hasMore() {
		t.err = errors.New("syntax error")
		return nil
	}
	return node
}

// Return true if there is another token to process; otherwise, return false.
func (t *tokenStream) hasMore() bool {
	return t.index < len(t.tokens)
}

// Return the value of the next token without advancing the index.
func (t *tokenStream) peek() *token {
	if t.hasMore() {
		token := t.tokens[t.index]
		return &token
	}
	return nil
}

// Advance the index to the next token.
func (t *tokenStream) next() {
	if !t.hasMore() {
		t.err = errors.New("read past end of tokens")
		return
	}
	t.index++
}

func (t *tokenStream) parseParenthesizedExpression() *node {
	openParen := t.parseOperator("(")
	if openParen == nil {
		// paren not found
		return nil
	}

	expr := t.parseExpression()
	if t.err != nil {
		return nil
	}

	closeParen := t.parseOperator(")")
	if closeParen == nil {
		t.err = errors.New("expected ')'")
		return nil
	}

	return expr
}

func (t *tokenStream) parseAtom() *node {
	parenNode := t.parseParenthesizedExpression()
	if t.err != nil {
		return nil
	}
	if parenNode != nil {
		return parenNode
	}

	refNode := t.parseLicenseRef()
	if t.err != nil {
		return nil
	}
	if refNode != nil {
		return refNode
	}

	licenseNode := t.parseLicense()
	if t.err != nil {
		return nil
	}
	if licenseNode != nil {
		return licenseNode
	}

	t.err = errors.New("expected node, but found none")
	return nil
}

func (t *tokenStream) parseExpression() *node {
	left := t.parseAnd()
	if t.err != nil {
		return nil
	}
	if left == nil {
		return nil
	}
	if !t.hasMore() {
		// expression found and no more tokens to process
		return left
	}

	operator := t.parseOperator("OR")
	if operator == nil {
		return left
	}
	op := strings.ToLower(*operator)

	right := t.parseExpression()
	if t.err != nil {
		return nil
	}
	if right == nil {
		t.err = errors.New("expected expression, but found none")
		return nil
	}

	return &(node{
		role: expressionNode,
		exp: &(expressionNodePartial{
			left:        left,
			conjunction: op,
			right:       right,
		}),
	})
}

// Return a node representation of an atomic value or an AND expression.  If a malformed
// atomic value or expression is found, an error is returned.  Advances the index if a
// valid atomic value or a valid expression is found.
func (t *tokenStream) parseAnd() *node {
	left := t.parseAtom()
	if t.err != nil {
		return nil
	}
	if left == nil {
		return nil
	}
	if !t.hasMore() {
		// atomic token found and no more tokens to process
		return left
	}

	operator := t.parseOperator("AND")
	if operator == nil {
		return left
	}

	right := t.parseAnd()
	if t.err != nil {
		return nil
	}
	if right == nil {
		t.err = errors.New("expected expression, but found none")
		return nil
	}

	exp := expressionNodePartial{left: left, conjunction: "and", right: right}

	return &(node{
		role: expressionNode,
		exp:  &exp,
	})
}

// Return a node representation of a License Reference.  If a malformed license reference is
// found, an error is returned.  Advances the index if a valid license reference is found.
func (t *tokenStream) parseLicenseRef() *node {
	ref := referenceNodePartial{documentRef: "", hasDocumentRef: false, licenseRef: ""}

	token := t.peek()
	if token.role == documentRefToken {
		ref.documentRef = token.value
		ref.hasDocumentRef = true
		t.next()

		operator := t.parseOperator(":")
		if operator == nil {
			t.err = errors.New("expected ':' after 'DocumentRef-...'")
			return nil
		}
	}

	token = t.peek()
	if token.role != licenseRefToken && ref.hasDocumentRef {
		t.err = errors.New("expected 'LicenseRef-...' after 'DocumentRef-...'")
		return nil
	} else if token.role != licenseRefToken {
		// not found is not an error as long as DocumentRef and : weren't the previous tokens
		return nil
	}

	ref.licenseRef = token.value
	t.next()

	return &(node{
		role: licenseRefNode,
		ref:  &ref,
	})
}

// Return a node representation of a License.  If a malformed license is found,
// an error is returned.  Advances the index if a valid license is found.
func (t *tokenStream) parseLicense() *node {
	token := t.peek()
	if token.role != licenseToken && token.role != extensionToken {
		return nil
	}
	t.next()

	lic := licenseNodePartial{
		license:      token.value,
		hasPlus:      false,
		hasException: false,
		exception:    ""}

	// for licenses that specifically support -or-later, a `+` operator token isn't expected to be present
	if strings.HasSuffix(token.value, "-or-later") {
		lic.hasPlus = true
	}

	if t.hasMore() {
		// use new var idx to avoid creating a new var index
		operator := t.parseOperator("+")
		if operator != nil {
			lic.hasPlus = true
		}

		if t.hasMore() {
			exception := t.parseWith()
			if t.err != nil {
				return nil
			}
			if exception != nil {
				lic.hasException = true
				lic.exception = *exception
				t.next()
			}
		}
	}

	return &(node{
		role: licenseNode,
		lic:  &lic,
	})
}

// Return the operator's value (e.g. AND, OR, WITH) if the current token is an OPERATOR.
// Advances the index if the operator is found.
func (t *tokenStream) parseOperator(operator string) *string {
	token := t.peek()
	if token.role == operatorToken && token.value == operator {
		t.next()
		return &(token.value)
	}
	// requested operator not found
	return nil
}

// Get the exception license when the WITH operator is found.
// Return without advancing the index if the current token is not the WITH operator.
// Raise an error if the WITH operator is not followed by and EXCEPTION license.
func (t *tokenStream) parseWith() *string {
	operator := t.parseOperator("WITH")
	if operator == nil {
		// WITH not found is not an error
		return nil
	}

	token := t.peek()
	if token.role != exceptionToken {
		t.err = errors.New("expected exception after 'WITH'")
		return nil
	}

	return &(token.value)
}

// Returns a human readable representation of the node tree.
func (n *node) string() string {
	switch n.role {
	case expressionNode:
		return expressionString(*n.exp)
	case licenseNode:
		return licenseString(*n.lic)
	case licenseRefNode:
		return referenceString(*n.ref)
	}
	return ""
}

func expressionString(exp expressionNodePartial) string {
	s := "{ LEFT: " + exp.left.string() + " "
	s += exp.conjunction
	s += " RIGHT: " + exp.right.string() + " }"
	return s
}

func licenseString(lic licenseNodePartial) string {
	s := lic.license
	if lic.hasPlus {
		s += "+"
	}
	if lic.hasException {
		s += " with " + lic.exception
	}
	return s
}

func referenceString(ref referenceNodePartial) string {
	s := ""
	if ref.hasDocumentRef {
		s = "DocumentRef-" + ref.documentRef + ":"
	}
	s += "LicenseRef-" + ref.licenseRef
	return s
}
