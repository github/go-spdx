package spdxexp

import (
	"fmt"
)

// Return the string representation of the license or license ref.
// TODO: Original had "NOASSERTION".  Does that still apply?
func (node *Node) licenseString() *string {
	switch node.role {
	case LICENSE_NODE:
		license := *node.License()
		if node.HasPlus() {
			license += "+"
		}
		if node.HasException() {
			license += " WITH " + *node.Exception()
		}
		return &license
	case LICENSEREF_NODE:
		license := "LicenseRef-" + *node.LicenseRef()
		if node.HasDocumentRef() {
			license = "DocumentRef-" + *node.DocumentRef() + ":" + license
		}
		return &license
	}
	return nil
}

// // Flatten the given expression into an array of all licenses mentioned in the expression.
// func flatten (expression) {
//   const expanded = Array.from(expandInner(expression))
//   const flattened = expanded.reduce(func (result, clause) {
//     return Object.assign(result, clause)
//   }, {})
//   return sort([flattened])[0]
// }

// Expand the given expression into an equivalent array representing ANDed licenses
// grouped in an array and ORed licenses each in a separate array.
//
// Example:
//   License Node: "MIT" becomes [["MIT"]]
//   OR Expression: "MIT OR Apache-2.0" becomes [["MIT"], ["Apache-2.0"]]
//   AND Expression: "MIT AND Apache-2.0" becomes [["MIT", "Apache-2.0"]]
//   OR-AND Expression: "MIT OR Apache-2.0 AND GPL-2.0" becomes [["MIT"], ["Apache-2.0", "GPL-2.0"]]
//   OR(AND) Expression: "MIT OR (Apache-2.0 AND GPL-2.0)" becomes [["MIT"], ["Apache-2.0", "GPL-2.0"]]
//   AND-OR Expression: "MIT AND Apache-2.0 OR GPL-2.0" becomes [["Apache-2.0", "MIT], ["GPL-2.0"]]
//   AND(OR) Expression: "MIT AND (Apache-2.0 OR GPL-2.0)" becomes [["Apache-2.0", "MIT], ["GPL-2.0", "MIT"]]
//   OR-AND-OR Expression: "MIT OR ISC AND Apache-2.0 OR GPL-2.0" becomes
//       [["MIT"], ["Apache-2.0", "ISC"], ["GPL-2.0"]]
//   (OR)AND(OR) Expression: "(MIT OR ISC) AND (Apache-2.0 OR GPL-2.0)" becomes
//       [["Apache-2.0", "MIT"], ["GPL-2.0", "MIT"], ["Apache-2.0", "ISC"], ["GPL-2.0", "ISC"]]
//   OR(AND)OR Expression: "MIT OR (ISC AND Apache-2.0) OR GPL-2.0" becomes
//       [["MIT"], ["Apache-2.0", "ISC"], ["GPL-2.0"]]
//   AND-OR-AND Expression: "MIT AND ISC OR Apache-2.0 AND GPL-2.0" becomes
//       [["ISC", "MIT"], ["Apache-2.0", "GPL-2.0"]]
//   (AND)OR(AND) Expression: "(MIT AND ISC) OR (Apache-2.0 AND GPL-2.0)" becomes
//       [["ISC", "MIT"], ["Apache-2.0", "GPL-2.0"]]
//   AND(OR)AND Expression: "MIT AND (ISC OR Apache-2.0) AND GPL-2.0" becomes
//       [["GPL-2.0", "ISC", "MIT"], ["Apache-2.0", "GPL-2.0", "MIT"]]
func (node *Node) expand() [][]string {
	if node.IsLicense() || node.IsLicenseRef() {
		var result [][]string
		license := []string{*node.licenseString()}
		return append(result, license)
	}

	// TODO: Need to deep-sort results
	if node.IsOrExpression() {
		return node.expandOr()
	}
	return node.expandAnd()
}

// Expand the given expression into an equivalent array representing ORed licenses each in a separate array.
//
// Example:
//   OR Expression: "MIT OR Apache-2.0" becomes [["MIT"], ["Apache-2.0"]]
func (node *Node) expandOr() [][]string {
	var result [][]string
	if node.Left().IsLicense() {
		left := []string{*node.Left().licenseString()}
		result = append(result, left)
	} else if node.Left().IsExpression() {
		if node.Left().IsOrExpression() {
			left := node.Left().expandOr()
			result = append(result, left...)
		} else if node.Left().IsAndExpression() {
			left := node.Left().expandAnd()[0]
			result = append(result, left)
		}
	}
	if node.Right().IsLicense() {
		right := []string{*node.Right().licenseString()}
		result = append(result, right)
	} else if node.Right().IsExpression() {
		if node.Right().IsOrExpression() {
			right := node.Right().expandOr()
			result = append(result, right...)
		} else if node.Right().IsAndExpression() {
			right := node.Right().expandAnd()[0]
			result = append(result, right)
		}
	}
	return result
}

// Expand the given expression into an equivalent array representing ANDed licenses
// grouped in an array.  When an ORed expression is combined with AND, the ORed
// expressions are combined with the ANDed expressions.
//
// Example:
//   AND Expression: "MIT AND Apache-2.0" becomes [["MIT", "Apache-2.0"]]
//   AND(OR) Expression: "MIT AND (Apache-2.0 OR GPL-2.0)" becomes [["Apache-2.0", "MIT], ["GPL-2.0", "MIT"]]
// See more examples under func expand.
func (node *Node) expandAnd() [][]string {
	var left [][]string
	if node.Left().IsLicense() {
		left = append(left, []string{*node.Left().licenseString()})
	} else if node.Left().IsExpression() {
		if node.Left().IsAndExpression() {
			left = node.Left().expandAnd()
		} else if node.Left().IsOrExpression() {
			left = node.Left().expandOr()
		}
	}
	var right [][]string
	if node.Right().IsLicense() {
		right = append(right, []string{*node.Right().licenseString()})
	} else if node.Right().IsExpression() {
		if node.Right().IsAndExpression() {
			right = node.Right().expandAnd()
		} else if node.Right().IsOrExpression() {
			right = node.Right().expandOr()
		}
	}

	if len(left) > 1 || len(right) > 1 {
		// an OR expression has been processed
		// somewhere on the left and/or right node path
		return appendLeftRight(left, right)
	}

	// only AND expressions have been processed
	return mergeLeftRight(left, right)
}

// Append results from expanding the right expression into the results
// from expanding the left expression.  When at least one of the left/right
// nodes includes an OR expression, the values are spread across at times
// producing more results than exists in the left or right results.
//
// Example:
//   left: {{"MIT"}} right: {{"ISC"}, {"Apache-2.0"}} becomes
//     {{"MIT", "ISC"}, {"MIT", "Apache-2.0"}}
func appendLeftRight(left, right [][]string) [][]string {
	var result [][]string
	for _, r := range right {
		for _, l := range left {
			tmp := append(l, r...)
			result = append(result, tmp)
		}
	}
	return result
}

// Merge results from expanding left and right expressions.
// When neither left/right nodes includes an OR expression, the values
// are merged left and right results.
//
// Example:
//   left: {{"MIT"}} right: {{"ISC", "Apache-2.0"}} becomes
//     {{"MIT", "ISC", "Apache-2.0"}}
func mergeLeftRight(left, right [][]string) [][]string {
	for _, r := range right {
		for j, l := range left {
			left[j] = append(l, r...)
		}
	}
	return left
}

// func sort (licenseList) {
//   var sortedLicenseLists = licenseList
//     .filter(func (e) { return Object.keys(e).length })
//     .map(func (e) { return Object.keys(e).sort() })
//   return sortedLicenseLists.map(func (list, i) {
//     return list.map(func (license) { return licenseList[i][license] })
//   })
// }

// // func isANDCompatible (one string, two string) bool {
// //   return one.every(func (o) {
// //     return two.some(func (t) { return licensesAreCompatible(o, t) })
// //   })
// // }

// Determine if first expression satisfies second expression.
//
// Examples:
//   "MIT" satisfies "MIT" is true
//
//   "MIT" satisfies "MIT OR Apache-2.0" is true
//   "MIT OR Apache-2.0" satisfies "MIT" is true
//   "GPL" satisfies "MIT OR Apache-2.0" is false
//   "MIT OR Apache-2.0" satisfies "GPL" is false
//
//   "Apache-2.0 AND MIT" satisfies "MIT AND Apache-2.0" is true
//   "MIT AND Apache-2.0" satisfies "MIT AND Apache-2.0" is true
//   "MIT" satisfies "MIT AND Apache-2.0" is false
//   "MIT AND Apache-2.0" satisfies "MIT" is false
//   "GPL" satisfies "MIT AND Apache-2.0" is false
//
//   "MIT AND Apache-2.0" satisfies "MIT AND (Apache-1.0 OR Apache-2.0)"
//
//   "Apache-1.0" satisfies "Apache-2.0+" is false
//   "Apache-2.0" satisfies "Apache-2.0+" is true
//   "Apache-3.0" satisfies "Apache-2.0+" is true
//
//   "Apache-1.0" satisfies "Apache-2.0-or-later" is false
//   "Apache-2.0" satisfies "Apache-2.0-or-later" is true
//   "Apache-3.0" satisfies "Apache-2.0-or-later" is true
//
//   "Apache-1.0" satisfies "Apache-2.0-only" is false
//   "Apache-2.0" satisfies "Apache-2.0-only" is true
//   "Apache-3.0" satisfies "Apache-2.0-only" is false
//
func satisfies(firstExp string, secondExp string) (bool, error) {
	firstTree, err := Parse(firstExp)
	if err != nil {
		return false, err
	}

	secondTree, err := Parse(secondExp)
	if err != nil {
		return false, err
	}

	nodes := &NodePair{firstNode: firstTree, secondNode: secondTree}
	if firstTree.IsLicense() && secondTree.IsLicense() {
		return nodes.LicensesAreCompatible(), nil
	}

	firstExpanded := firstTree.expand()
	fmt.Println("firstExpanded: ", firstExpanded)
	// secondFlattened := flatten(secondNormalized)

	// satisfactionFunc := func(o string) bool { return isAndCompatible(o, secondFlattened) }
	// satisfaction := some(firstExpanded, satisfactionFunc)

	// return one.some(satisfactionFunc)
	// return satisfaction

	// TODO: Stubbed
	return false, nil
}
