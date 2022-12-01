package expression

import (
	"sort"
	"strings"
)

type nodePair struct {
	firstNode  *node
	secondNode *node
}

type nodeRole uint8

const (
	expressionNode nodeRole = iota
	licenseRefNode
	licenseNode
)

type node struct {
	role nodeRole
	exp  *expressionNodePartial
	lic  *licenseNodePartial
	ref  *referenceNodePartial
}

type expressionNodePartial struct {
	left        *node
	conjunction string
	right       *node
}

type licenseNodePartial struct {
	license      string
	hasPlus      bool
	hasException bool
	exception    string
}

type referenceNodePartial struct {
	hasDocumentRef bool
	documentRef    string
	licenseRef     string
}

// ---------------------- Helper Methods ----------------------

func (n *node) isExpression() bool {
	return n.role == expressionNode
}

func (n *node) isOrExpression() bool {
	if !n.isExpression() {
		return false
	}
	return *n.conjunction() == "or"
}

func (n *node) isAndExpression() bool {
	if !n.isExpression() {
		return false
	}
	return *n.conjunction() == "and"
}

func (n *node) left() *node {
	if !n.isExpression() {
		return nil
	}
	return n.exp.left
}

func (n *node) conjunction() *string {
	if !n.isExpression() {
		return nil
	}
	return &(n.exp.conjunction)
}

func (n *node) right() *node {
	if !n.isExpression() {
		return nil
	}
	return n.exp.right
}

func (n *node) isLicense() bool {
	return n.role == licenseNode
}

// Return the value of the license field.
// See also reconstructedLicenseString()
func (n *node) license() *string {
	if !n.isLicense() {
		return nil
	}
	return &(n.lic.license)
}

func (n *node) exception() *string {
	if !n.hasException() {
		return nil
	}
	return &(n.lic.exception)
}

func (n *node) hasPlus() bool {
	if !n.isLicense() {
		return false
	}
	return n.lic.hasPlus
}

func (n *node) hasException() bool {
	if !n.isLicense() {
		return false
	}
	return n.lic.hasException
}

func (n *node) isLicenseRef() bool {
	return n.role == licenseRefNode
}

func (n *node) licenseRef() *string {
	if !n.isLicenseRef() {
		return nil
	}
	return &(n.ref.licenseRef)
}

func (n *node) documentRef() *string {
	if !n.hasDocumentRef() {
		return nil
	}
	return &(n.ref.documentRef)
}

func (n *node) hasDocumentRef() bool {
	if !n.isLicenseRef() {
		return false
	}
	return n.ref.hasDocumentRef
}

// reconstructedLicenseString returns the string representation of the license or license ref.
// TODO: Original had "NOASSERTION".  Does that still apply?
func (n *node) reconstructedLicenseString() *string {
	switch n.role {
	case licenseNode:
		license := *n.license()
		if n.hasPlus() {
			license += "+"
		}
		if n.hasException() {
			license += " WITH " + *n.exception()
		}
		return &license
	case licenseRefNode:
		license := "LicenseRef-" + *n.licenseRef()
		if n.hasDocumentRef() {
			license = "DocumentRef-" + *n.documentRef() + ":" + license
		}
		return &license
	}
	return nil
}

// Sort an array of license and license reference nodes alphebetically based
// on their reconstructedLicenseString() representation.  The sort function does not expect
// expression nodes, but if one is in the nodes list, it will sort to the end.
func sortLicenses(nodes []*node) {
	sort.Slice(nodes, func(i, j int) bool {
		if nodes[j].isExpression() {
			// push second license toward end by saying first license is less than
			return true
		}
		if nodes[i].isExpression() {
			// push first license toward end by saying second license is less than
			return false
		}
		return *nodes[i].reconstructedLicenseString() < *nodes[j].reconstructedLicenseString()
	})
}

// ---------------------- Comparator Methods ----------------------

// Return true if two licenses are compatible; otherwise, false.
func (nodes *nodePair) licensesAreCompatible() bool {
	if !nodes.firstNode.isLicense() || !nodes.secondNode.isLicense() {
		return false
	}
	if nodes.secondNode.hasPlus() {
		if nodes.firstNode.hasPlus() {
			// first+, second+
			return nodes.rangesAreCompatible()
		}
		// first, second+
		return nodes.identifierInRange()
	}
	// else secondNode does not have plus
	if nodes.firstNode.hasPlus() {
		// first+, second
		revNodes := &nodePair{firstNode: nodes.secondNode, secondNode: nodes.firstNode}
		return revNodes.identifierInRange()
	}
	// first, second
	return nodes.licensesExactlyEqual()
}

func (nodes *nodePair) licenseRefsAreCompatible() bool {
	if !nodes.firstNode.isLicenseRef() || !nodes.secondNode.isLicenseRef() {
		return false
	}

	compatible := *nodes.firstNode.licenseRef() == *nodes.secondNode.licenseRef()
	compatible = compatible && (nodes.firstNode.hasDocumentRef() == nodes.secondNode.hasDocumentRef())
	if compatible && nodes.firstNode.hasDocumentRef() {
		compatible = compatible && (*nodes.firstNode.documentRef() == *nodes.secondNode.documentRef())
	}
	return compatible
}

// Return true if two licenses are compatible in the context of their ranges; otherwise, false.
func (nodes *nodePair) rangesAreCompatible() bool {
	if nodes.licensesExactlyEqual() {
		// licenses specify ranges exactly the same (e.g. Apache-1.0+, Apache-1.0+)
		return true
	}

	firstNode := *nodes.firstNode
	secondNode := *nodes.secondNode

	firstRange := getLicenseRange(*firstNode.license())
	secondRange := getLicenseRange(*secondNode.license())

	// When both licenses allow later versions (i.e. hasPlus==true), being in the same license
	// group is sufficient for compatibility, as long as, any exception is also compatible
	// Example: All Apache licenses (e.g. Apache-1.0, Apache-2.0) are in the same license group
	return sameLicenseGroup(firstRange, secondRange) && nodes.exceptionsAreCompatible()
}

// identifierInRange returns true if the (first) simple license is in range of the (second)
// ranged license; otherwise, false.
func (nodes *nodePair) identifierInRange() bool {
	simpleLicense := nodes.firstNode
	plusLicense := nodes.secondNode

	if !compareGT(simpleLicense, plusLicense) && !compareEQ(simpleLicense, plusLicense) {
		return false
	}

	// With simpleLicense >= plusLicense, licenses are compatible, as long as, any exception
	// is also compatible
	return nodes.exceptionsAreCompatible()

}

// exceptionsAreCompatible returns true if neither license has an exception or they have
// the same exception; otherwise, false
func (nodes *nodePair) exceptionsAreCompatible() bool {
	firstNode := *nodes.firstNode
	secondNode := *nodes.secondNode

	if !firstNode.hasException() && !secondNode.hasException() {
		// if neither has an exception, then licenses are compatible
		return true
	}

	if firstNode.hasException() != secondNode.hasException() {
		// if one has and exception and the other does not, then the license are NOT compatible
		return false
	}

	return *nodes.firstNode.exception() == *nodes.secondNode.exception()

}

// Return true if the licenses are the same; otherwise, false
func (nodes *nodePair) licensesExactlyEqual() bool {
	return strings.EqualFold(*nodes.firstNode.reconstructedLicenseString(), *nodes.secondNode.reconstructedLicenseString())
}
