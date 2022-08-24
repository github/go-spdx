package spdxexp

/* Translation to GO from javascript code: https://github.com/clearlydefined/spdx-expression-parse.js/blob/master/scan.js */

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
)

type expressionStream struct {
	expression string
	index      int
	err        error
}

type token struct {
	role  tokenrole
	value string
}

type tokenrole int64

const (
	OPERATOR_TOKEN tokenrole = iota
	DOCUMENTREF_TOKEN
	LICENSEREF_TOKEN
	LICENSE_TOKEN
	EXCEPTION_TOKEN
)

// Scan expression gathering valid SPDX expression tokens.  Returns error if any tokens are invalid.
func scan(expression string) ([]token, error) {
	var tokens []token
	var token *token

	exp := &expressionStream{expression: expression, index: 0, err: nil}

	for exp.hasMore() {
		exp.skipWhitespace()
		if !exp.hasMore() {
			break
		}

		token = exp.parseToken()
		if exp.err != nil {
			// stop processing at first error and return
			return nil, exp.err
		}

		if token == nil {
			// TODO: shouldn't happen ???
			return nil, errors.New("got nil token when expecting more")
		}

		tokens = append(tokens, *token)
	}
	return tokens, nil
}

// Determine if expression has more to process.
func (exp *expressionStream) hasMore() bool {
	return exp.index < len(exp.expression)
}

// Try to read the next token starting at index. Returns error if no token is recognized.
func (exp *expressionStream) parseToken() *token {
	// Ordering matters
	op := exp.readOperator()
	if exp.err != nil {
		return nil
	}
	if op != nil {
		return op
	}

	dref := exp.readDocumentRef()
	if exp.err != nil {
		return nil
	}
	if dref != nil {
		return dref
	}

	lref := exp.readLicenseRef()
	if exp.err != nil {
		return nil
	}
	if lref != nil {
		return lref
	}

	identifier := exp.readIdentifier()
	if exp.err != nil {
		return nil
	}
	if identifier != nil {
		return identifier
	}

	errmsg := fmt.Sprintf("unexpected '%c' at offset %d", exp.expression[exp.index], exp.index)
	exp.err = errors.New(errmsg)
	return nil
}

// Read more from expression if the next substring starting at index matches the regex pattern.
func (exp *expressionStream) readRegex(pattern string) string {
	expressionSlice := exp.expression[exp.index:]

	r, _ := regexp.Compile(pattern)
	i := r.FindStringIndex(expressionSlice)
	if i != nil && i[1] > 0 && i[0] == 0 {
		// match found in expression at index
		exp.index += i[1]
		return expressionSlice[0:i[1]]
	}
	return ""
}

// Read more from expression if the substring starting at index is the next expected string.
func (exp *expressionStream) read(next string) string {
	expressionSlice := exp.expression[exp.index:]

	if strings.HasPrefix(expressionSlice, next) {
		// next found in expression at index
		exp.index += len(next)
		return next
	}
	return ""
}

// Skip whitespace in expression starting at index
func (exp *expressionStream) skipWhitespace() {
	exp.readRegex("[ ]*")
}

// Read operator in expression starting at index if it exists
func (exp *expressionStream) readOperator() *token {
	possibilities := []string{"WITH", "AND", "OR", "(", ")", ":", "+"}

	var op string
	for _, p := range possibilities {
		op = exp.read(p)
		if len(op) > 0 {
			break
		}
	}
	if len(op) == 0 {
		// not an error if an operator isn't found
		return nil
	}

	if op == "+" && exp.index > 1 && exp.expression[exp.index-2:exp.index-1] == " " {
		exp.err = errors.New("unexpected space before +")
		exp.index -= 1
		return nil
	}

	return &token{role: OPERATOR_TOKEN, value: op}
}

// Get id from expression starting at index.  Raise error if id not found.
func (exp *expressionStream) readID() string {
	id := exp.readRegex("[A-Za-z0-9-.]+")
	if len(id) == 0 {
		errmsg := fmt.Sprintf("expected id at offset %d", exp.index)
		exp.err = errors.New(errmsg)
		return ""
	}
	return id
}

// Read DocumentRef in expression starting at index if it exists. Raise error if found and id doesn't follow.
func (exp *expressionStream) readDocumentRef() *token {
	ref := exp.read("DocumentRef-")
	if len(ref) == 0 {
		// not an error if a DocumentRef isn't found
		return nil
	}

	id := exp.readID()
	if exp.err != nil {
		return nil
	}
	return &token{role: DOCUMENTREF_TOKEN, value: id}
}

// Read LicenseRef in expression starting at index if it exists. Raise error if found and id doesn't follow.
func (exp *expressionStream) readLicenseRef() *token {
	ref := exp.read("LicenseRef-")
	if len(ref) == 0 {
		// not an error if a LicenseRef isn't found
		return nil
	}

	id := exp.readID()
	if exp.err != nil {
		return nil
	}
	return &token{role: LICENSEREF_TOKEN, value: id}
}

// Read a LICENSE/EXCEPTION in expression starting at index if it exists. Raise error if found and id doesn't follow.
func (exp *expressionStream) readIdentifier() *token {
	// because readID matches broadly, save the index so it can be reset if an actual license is not found
	index := exp.index

	id := exp.readID()
	if exp.err != nil {
		return nil
	}

	if ActiveLicense(id) || DeprecatedLicense(id) {
		return &token{role: LICENSE_TOKEN, value: id}
	} else if ExceptionLicense(id) {
		return &token{role: EXCEPTION_TOKEN, value: id}
	}

	// license not found in indices
	exp.index = index
	return nil
}
