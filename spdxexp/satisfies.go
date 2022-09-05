package spdxexp

// "fmt"

// func licenseString (e spdx.Expression) {
//   if (e.hasOwnProperty("noassertion")) return "NOASSERTION"
//   if (e.license) return `${e.license}${e.plus ? "+" : ""}${e.exception ? ` WITH ${e.exception}` : ""}`
// }

// // Expand the given expression into an equivalent array where each member is an array of licenses AND"d
// // together and the members are OR"d together. For example, `(MIT OR ISC) AND GPL-3.0` expands to
// // `[[GPL-3.0 AND MIT], [ISC AND MIT]]`. Note that within each array of licenses, the entries are
// // normalized (sorted) by license name.
// func expand (expression spdx.Expression) {
//   return sort(Array.from(expandInner(expression)))
// }

// // Flatten the given expression into an array of all licenses mentioned in the expression.
// func flatten (expression) {
//   const expanded = Array.from(expandInner(expression))
//   const flattened = expanded.reduce(func (result, clause) {
//     return Object.assign(result, clause)
//   }, {})
//   return sort([flattened])[0]
// }

// func expandInner (expression spdx.Expression) spdx.Expression {

//   type := reflect.TypeOf(expression)
//   switch type {
//   case reflect.TypeOf(spdx.LicenseID{}):

//   case reflect.TypeOf(spdx.Or{}):
//   case reflect.TypeOf(spdx.And{}):
//   case reflect.TypeOf(spdx.Left{}):
//   case reflect.TypeOf(spdx.Right{}):

//   }
//   if (!expression.conjunction) return [{ [licenseString(expression)]: expression }]
//   if (expression.conjunction === "or") return expandInner(expression.left).concat(expandInner(expression.right))
//   if (expression.conjunction === "and") {
//     var left = expandInner(expression.left)
//     var right = expandInner(expression.right)
//     return left.reduce(func (result, l) {
//       right.forEach(func (r) { result.push(Object.assign({}, l, r)) })
//       return result
//     }, [])
//   }
// }

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

	// firstNormalized := firstTree   // normalizeGPLIdentifiers(firstTree)
	// secondNormalized := secondTree // normalizeGPLIdentifiers(secondTree)

	// firstExpanded := expand(firstNormalized)
	// secondFlattened := flatten(secondNormalized)

	// satisfactionFunc := func(o string) bool { return isAndCompatible(o, secondFlattened) }
	// satisfaction := some(firstExpanded, satisfactionFunc)

	// return one.some(satisfactionFunc)
	// return satisfaction

	// TODO: Stubbed
	return false, nil
}
