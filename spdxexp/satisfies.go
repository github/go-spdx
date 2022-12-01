package spdxexp

import (
	"errors"
	"sort"
)

// ValidateLicenses checks if given licenses are valid according to spdx.
// Returns true if all licenses are valid; otherwise, false.
// Returns all the invalid licenses contained in the `licenses` argument.
func ValidateLicenses(licenses []string) (bool, []string) {
	valid := true
	invalidLicenses := []string{}
	for _, license := range licenses {
		if _, err := parse(license); err != nil {
			valid = false
			invalidLicenses = append(invalidLicenses, license)
		}
	}
	return valid, invalidLicenses
}

// Satisfies determines if the allowed list of licenses satisfies the test license expression.
// Returns true if allowed list satisfies test license expression; otherwise, false.
// Returns error if error occurs during processing.
func Satisfies(testExpression string, allowedList []string) (bool, error) {
	expressionNode, err := parse(testExpression)
	if err != nil {
		return false, err
	}
	if len(allowedList) == 0 {
		return false, errors.New("allowedList requires at least one element, but is empty")
	}
	allowedNodes, err := stringsToNodes(allowedList)
	if err != nil {
		return false, err
	}
	sortAndDedup(allowedNodes)

	expandedExpression := expressionNode.expand(true)

	for _, expressionPart := range expandedExpression {
		if isCompatible(expressionPart, allowedNodes) {
			// return once any expressionPart is compatible with the allow list
			// * each part is an array of licenses that are ANDed, meaning all have to be on the allowedList
			// * the parts are ORed, meaning only one of the parts need to be compatible
			return true, nil
		}
	}
	return false, nil
}

// stringsToNodes converts an array of single license strings to to an array of license nodes.
func stringsToNodes(licenseStrings []string) ([]*node, error) {
	nodes := make([]*node, len(licenseStrings))
	for i, s := range licenseStrings {
		node, err := parse(s)
		if err != nil {
			return nil, err
		}
		if node.isExpression() {
			return nil, errors.New("expressions are not supported in the allowedList")
		}
		nodes[i] = node
	}
	return nodes, nil
}

// isCompatible checks if expressionPart is compatible with allowed list.
// Expression part is an array of licenses that are ANDed together.
// Allowed is an array of licenses that can fulfill the expression.
func isCompatible(expressionPart, allowed []*node) bool {
	for _, expLicense := range expressionPart {
		compatible := false
		for _, allowedLicense := range allowed {
			nodes := &nodePair{firstNode: expLicense, secondNode: allowedLicense}
			if nodes.licensesAreCompatible() || nodes.licenseRefsAreCompatible() {
				compatible = true
				break
			}
		}
		if !compatible {
			// no compatible license found for one of the required licenses
			return false
		}
	}
	// found a compatible license in test for each required license
	return true
}

// expand will expand the given expression into an equivalent array representing ANDed licenses
// grouped in an array and ORed licenses each in a separate array.
//
// Example:
//
//	License node: "MIT" becomes [["MIT"]]
//	OR Expression: "MIT OR Apache-2.0" becomes [["MIT"], ["Apache-2.0"]]
//	AND Expression: "MIT AND Apache-2.0" becomes [["MIT", "Apache-2.0"]]
//	OR-AND Expression: "MIT OR Apache-2.0 AND GPL-2.0" becomes [["MIT"], ["Apache-2.0", "GPL-2.0"]]
//	OR(AND) Expression: "MIT OR (Apache-2.0 AND GPL-2.0)" becomes [["MIT"], ["Apache-2.0", "GPL-2.0"]]
//	AND-OR Expression: "MIT AND Apache-2.0 OR GPL-2.0" becomes [["Apache-2.0", "MIT], ["GPL-2.0"]]
//	AND(OR) Expression: "MIT AND (Apache-2.0 OR GPL-2.0)" becomes [["Apache-2.0", "MIT], ["GPL-2.0", "MIT"]]
//	OR-AND-OR Expression: "MIT OR ISC AND Apache-2.0 OR GPL-2.0" becomes
//	    [["MIT"], ["Apache-2.0", "ISC"], ["GPL-2.0"]]
//	(OR)AND(OR) Expression: "(MIT OR ISC) AND (Apache-2.0 OR GPL-2.0)" becomes
//	    [["Apache-2.0", "MIT"], ["GPL-2.0", "MIT"], ["Apache-2.0", "ISC"], ["GPL-2.0", "ISC"]]
//	OR(AND)OR Expression: "MIT OR (ISC AND Apache-2.0) OR GPL-2.0" becomes
//	    [["MIT"], ["Apache-2.0", "ISC"], ["GPL-2.0"]]
//	AND-OR-AND Expression: "MIT AND ISC OR Apache-2.0 AND GPL-2.0" becomes
//	    [["ISC", "MIT"], ["Apache-2.0", "GPL-2.0"]]
//	(AND)OR(AND) Expression: "(MIT AND ISC) OR (Apache-2.0 AND GPL-2.0)" becomes
//	    [["ISC", "MIT"], ["Apache-2.0", "GPL-2.0"]]
//	AND(OR)AND Expression: "MIT AND (ISC OR Apache-2.0) AND GPL-2.0" becomes
//	    [["GPL-2.0", "ISC", "MIT"], ["Apache-2.0", "GPL-2.0", "MIT"]]
func (n *node) expand(withDeepSort bool) [][]*node {
	if n.isLicense() || n.isLicenseRef() {
		return [][]*node{{n}}
	}

	var expanded [][]*node
	if n.isOrExpression() {
		expanded = n.expandOr()
	} else {
		expanded = n.expandAnd()
	}

	if withDeepSort {
		expanded = deepSort(expanded)
	}
	return expanded
}

// expandOr expands the given expression into an equivalent array representing ORed licenses each in a separate array.
//
// Example:
//
//	OR Expression: "MIT OR Apache-2.0" becomes [["MIT"], ["Apache-2.0"]]
func (n *node) expandOr() [][]*node {
	var result [][]*node
	result = expandOrTerm(n.left(), result)
	result = expandOrTerm(n.right(), result)
	return result
}

// expandOrTerm expands the terms of an OR expression.
func expandOrTerm(term *node, result [][]*node) [][]*node {
	if term.isLicense() {
		result = append(result, []*node{term})
	} else if term.isExpression() {
		if term.isOrExpression() {
			left := term.expandOr()
			result = append(result, left...)
		} else if term.isAndExpression() {
			left := term.expandAnd()[0]
			result = append(result, left)
		}
	}
	return result
}

// expandAnd expands the given expression into an equivalent array representing ANDed licenses
// grouped in an array.  When an ORed expression is combined with AND, the ORed
// expressions are combined with the ANDed expressions.
//
// Example:
//
//	AND Expression: "MIT AND Apache-2.0" becomes [["MIT", "Apache-2.0"]]
//	AND(OR) Expression: "MIT AND (Apache-2.0 OR GPL-2.0)" becomes [["Apache-2.0", "MIT], ["GPL-2.0", "MIT"]]
//
// See more examples under func expand.
func (n *node) expandAnd() [][]*node {
	left := expandAndTerm(n.left())
	right := expandAndTerm(n.right())

	if len(left) > 1 || len(right) > 1 {
		// an OR expression has been processed
		// somewhere on the left and/or right node path
		return appendTerms(left, right)
	}

	// only AND expressions have been processed
	return mergeTerms(left, right)
}

// expandAndTerm expands the terms of an AND expression.
func expandAndTerm(term *node) [][]*node {
	var result [][]*node
	if term.isLicense() || term.isLicenseRef() {
		result = append(result, []*node{term})
	} else if term.isExpression() {
		if term.isAndExpression() {
			result = term.expandAnd()
		} else if term.isOrExpression() {
			result = term.expandOr()
		}
	}
	return result
}

// appendTerms appends results from expanding the right expression into the results
// from expanding the left expression.  When at least one of the left/right
// nodes includes an OR expression, the values are spread across at times
// producing more results than exists in the left or right results.
//
// Example:
//
//	left: {{"MIT"}} right: {{"ISC"}, {"Apache-2.0"}} becomes
//	  {{"MIT", "ISC"}, {"MIT", "Apache-2.0"}}
func appendTerms(left, right [][]*node) [][]*node {
	var result [][]*node
	for _, r := range right {
		for _, l := range left {
			tmp := l
			tmp = append(tmp, r...)
			result = append(result, tmp)
		}
	}
	return result
}

// mergeTerms merges results from expanding left and right expressions.
// When neither left/right nodes includes an OR expression, the values
// are merged left and right results.
//
// Example:
//
//	left: {{"MIT"}} right: {{"ISC", "Apache-2.0"}} becomes
//	  {{"MIT", "ISC", "Apache-2.0"}}
func mergeTerms(left, right [][]*node) [][]*node {
	results := left
	for _, r := range right {
		for j, l := range results {
			results[j] = append(l, r...)
		}
	}
	return results
}

// sortAndDedup sorts an array of license nodes and then removes duplicates.
func sortAndDedup(nodes []*node) []*node {
	if len(nodes) <= 1 {
		return nodes
	}

	sortLicenses(nodes)
	prev := 1
	for curr := 1; curr < len(nodes); curr++ {
		if *nodes[curr-1].reconstructedLicenseString() != *nodes[curr].reconstructedLicenseString() {
			nodes[prev] = nodes[curr]
			prev++
		}
	}

	return nodes[:prev]
}

// deepSort sorts a two-dimensional array of license nodes.  Internal arrays are sorted first.
// Then each array of nodes are sorted relative to the other arrays.
//
// Example:
//
//	BEFORE {{"MIT", "GPL-2.0"}, {"ISC", "Apache-2.0"}}
//	AFTER  {{"Apache-2.0", "ISC"}, {"GPL-2.0", "MIT"}}
func deepSort(nodes2d [][]*node) [][]*node {
	if len(nodes2d) == 0 || len(nodes2d) == 1 && len(nodes2d[0]) <= 1 {
		return nodes2d
	}

	// sort each array internally
	// Example:
	//   BEFORE {{"MIT", "GPL-2.0"}, {"ISC", "Apache-2.0"}}
	//   AFTER  {{"GPL-2.0", "MIT"}, {"Apache-2.0", "ISC"}}
	for _, nodes := range nodes2d {
		if len(nodes) > 1 {
			sortLicenses(nodes)
		}
	}

	// sort arrays relative to each other
	// Example:
	//   BEFORE {{"GPL-2.0", "MIT"}, {"Apache-2.0", "ISC"}}
	//   AFTER  {{"Apache-2.0", "ISC"}, {"GPL-2.0", "MIT"}}
	sort.Slice(nodes2d, func(i, j int) bool {
		// TODO: Consider refactor to map nodes to licenseString before processing.
		for k := range nodes2d[j] {
			if k >= len(nodes2d[i]) {
				// if the first k elements are equal and the second array is
				// longer than the first, the first is considered less than
				return true
			}
			iLicense := *nodes2d[i][k].reconstructedLicenseString()
			jLicense := *nodes2d[j][k].reconstructedLicenseString()
			if iLicense != jLicense {
				// when elements are not equal, return true if first is less than
				return iLicense < jLicense
			}
		}
		// all elements are equal, return false to avoid a swap
		return false
	})

	return nodes2d
}
