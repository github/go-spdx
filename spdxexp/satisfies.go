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

	for _, s1d := range firstExpanded {
		if isANDCompatible(s1d, secondFlattened) {
			return true, nil
		}
	}
	return false, nil
}

func isANDCompatible(required, test []*Node) bool {
	for _, r := range required {
		compatible := false
		for _, t := range test {
			nodes := &NodePair{firstNode: r, secondNode: t}
			if nodes.LicensesAreCompatible() {
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
func (node *Node) flatten() []*Node {
	var flattened []*Node
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
func (node *Node) expand(withDeepSort bool) [][]*Node {
	if node.IsLicense() || node.IsLicenseRef() {
		return [][]*Node{{node}}
	}

	var expanded [][]*Node
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
func (node *Node) expandOr() [][]*Node {
	var result [][]*Node
	if node.Left().IsLicense() {
		result = append(result, []*Node{node.Left()})
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
		result = append(result, []*Node{node.Right()})
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
func (node *Node) expandAnd() [][]*Node {
	var left [][]*Node
	if node.Left().IsLicense() {
		left = append(left, []*Node{node.Left()})
	} else if node.Left().IsExpression() {
		if node.Left().IsAndExpression() {
			left = node.Left().expandAnd()
		} else if node.Left().IsOrExpression() {
			left = node.Left().expandOr()
		}
	}
	var right [][]*Node
	if node.Right().IsLicense() {
		right = append(right, []*Node{node.Right()})
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
func appendLeftRight(left, right [][]*Node) [][]*Node {
	var result [][]*Node
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
func mergeLeftRight(left, right [][]*Node) [][]*Node {
	for _, r := range right {
		for j, l := range left {
			left[j] = append(l, r...)
		}
	}
	return left
}

// Sort and dedup an array of license nodes.
func sortAndDedup(nodes []*Node) []*Node {
	if len(nodes) <= 1 {
		return nodes
	}

	SortLicenses(nodes)
	prev := 1
	for curr := 1; curr < len(nodes); curr++ {
		if *nodes[curr-1].LicenseString() != *nodes[curr].LicenseString() {
			nodes[prev] = nodes[curr]
			prev++
		}
	}

	return nodes[:prev]
}

// Sort two-dimensional array of license nodes.  Internal arrays are sorted first.
// Then each array of nodes are sorted relative to the other arrays.
//
// Example:
//   BEFORE {{"MIT", "GPL-2.0"}, {"ISC", "Apache-2.0"}}
//   AFTER  {{"Apache-2.0", "ISC"}, {"GPL-2.0", "MIT"}}
func deepSort(nodes2d [][]*Node) [][]*Node {
	if len(nodes2d) == 0 || len(nodes2d) == 1 && len(nodes2d[0]) <= 1 {
		return nodes2d
	}

	// sort each array internally
	// Example:
	//   BEFORE {{"MIT", "GPL-2.0"}, {"ISC", "Apache-2.0"}}
	//   AFTER  {{"GPL-2.0", "MIT"}, {"Apache-2.0", "ISC"}}
	for _, nodes := range nodes2d {
		if len(nodes) > 1 {
			SortLicenses(nodes)
		}
	}

	// sort arrays relative to each other
	// Example:
	//   BEFORE {{"GPL-2.0", "MIT"}, {"Apache-2.0", "ISC"}}
	//   AFTER  {{"Apache-2.0", "ISC"}, {"GPL-2.0", "MIT"}}
	sort.Slice(nodes2d, func(i, j int) bool {
		// TODO: Consider refactor to map nodes to LicenseString before processing.
		for k, _ := range nodes2d[j] {
			if k >= len(nodes2d[i]) {
				// if the first k elements are equal and the second array is
				// longer than the first, the first is considered less than
				return true
			}
			iLicense := *nodes2d[i][k].LicenseString()
			jLicense := *nodes2d[j][k].LicenseString()
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
