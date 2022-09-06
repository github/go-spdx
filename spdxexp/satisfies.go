package spdxexp

import (
	"fmt"
	"sort"
)

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
func Satisfies(firstExp string, secondExp string) (bool, error) {
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

	firstExpanded := firstTree.expand(true)
	fmt.Println("firstExpanded: ", firstExpanded)
	secondFlattened := secondTree.flatten()
	fmt.Println("secondFlattened: ", secondFlattened)

	// satisfactionFunc := func(o string) bool { return isAndCompatible(o, secondFlattened) }
	// satisfaction := some(firstExpanded, satisfactionFunc)

	// return one.some(satisfactionFunc)
	// return satisfaction

	// TODO: Stubbed
	return false, nil
}

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

// Flatten the given expression into an array of all licenses mentioned in the expression.
//
// Example:
//   License Node: "MIT" becomes ["MIT"]
//   OR Expression: "MIT OR Apache-2.0" becomes ["Apache-2.0", "MIT"]
//   AND Expression: "MIT AND Apache-2.0" becomes ["Apache-2.0", "MIT"]
//   OR-AND Expression: "MIT OR Apache-2.0 AND GPL-2.0" becomes ["Apache-2.0", "GPL-2.0", "MIT"]
//   OR(AND) Expression: "MIT OR (Apache-2.0 AND GPL-2.0)" becomes ["Apache-2.0", "GPL-2.0", "MIT"]
//   AND-OR Expression: "MIT AND Apache-2.0 OR GPL-2.0" becomes ["Apache-2.0", "GPL-2.0", "MIT"]
//   AND(OR) Expression: "MIT AND (Apache-2.0 OR GPL-2.0)" becomes ["Apache-2.0", "GPL-2.0", "MIT"]
//   OR-AND-OR Expression: "MIT OR ISC AND Apache-2.0 OR GPL-2.0" becomes
//       ["Apache-2.0", "GPL-2.0", "ISC", "MIT"]
//   (OR)AND(OR) Expression: "(MIT OR ISC) AND (Apache-2.0 OR GPL-2.0)" becomes
//       ["Apache-2.0", "GPL-2.0", "ISC", "MIT"]
//   OR(AND)OR Expression: "MIT OR (ISC AND Apache-2.0) OR GPL-2.0" becomes
//       ["Apache-2.0", "GPL-2.0", "ISC", "MIT"]
//   AND-OR-AND Expression: "MIT AND ISC OR Apache-2.0 AND GPL-2.0" becomes
//       ["Apache-2.0", "GPL-2.0", "ISC", "MIT"]
//   (AND)OR(AND) Expression: "(MIT AND ISC) OR (Apache-2.0 AND GPL-2.0)" becomes
//       ["Apache-2.0", "GPL-2.0", "ISC", "MIT"]
//   AND(OR)AND Expression: "MIT AND (ISC OR Apache-2.0) AND GPL-2.0" becomes
//       ["Apache-2.0", "GPL-2.0", "ISC", "MIT"]
func (node *Node) flatten() []string {
	var flattened []string
	expanded := node.expand(false)
	for _, licenses := range expanded {
		flattened = append(flattened, licenses...)
	}
	return sortAndDedup(flattened)
}

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
func (node *Node) expand(withDeepSort bool) [][]string {
	if node.IsLicense() || node.IsLicenseRef() {
		var result [][]string
		license := []string{*node.licenseString()}
		return append(result, license)
	}

	var expanded [][]string
	if node.IsOrExpression() {
		expanded = node.expandOr()
	} else {
		expanded = node.expandAnd()
	}

	if withDeepSort {
		expanded = deepSort(expanded)
	}
	return expanded
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

// Sort and dedup an array of strings.
func sortAndDedup(s []string) []string {
	if len(s) <= 1 {
		return s
	}

	sort.Strings(s)
	prev := 1
	for curr := 1; curr < len(s); curr++ {
		if s[curr-1] != s[curr] {
			s[prev] = s[curr]
			prev++
		}
	}

	return s[:prev]
}

func deepSort(s2d [][]string) [][]string {
	if len(s2d) == 0 || len(s2d) == 1 && len(s2d[0]) <= 1 {
		return s2d
	}

	// sort each array internally
	// Example:
	//   BEFORE {{"MIT", "GPL-2.0"}, {"ISC", "Apache-2.0"}}
	//   AFTER  {{"GPL-2.0", "MIT"}, {"Apache-2.0", "ISC"}}
	for _, s := range s2d {
		if len(s) > 1 {
			sort.Strings(s)
		}
	}

	// sort arrays relative to each other
	// Example:
	//   BEFORE {{"GPL-2.0", "MIT"}, {"Apache-2.0", "ISC"}}
	//   AFTER  {{"Apache-2.0", "ISC"}, {"GPL-2.0", "MIT"}}
	sort.Slice(s2d, func(i, j int) bool {
		for k, _ := range s2d[j] {
			if k >= len(s2d[i]) {
				// if the first k elements are equal and the second array is
				// longer than the first, the first is considered less than
				return true
			}
			if s2d[i][k] != s2d[j][k] {
				// when elements are not equal, return true if first is less than
				return s2d[i][k] < s2d[j][k]
			}
		}
		// all elements are equal, return false to avoid a swap
		return false
	})

	return s2d
}
