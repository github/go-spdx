package spdxexp

import "sort"

type nodePair struct {
	firstNode  *Node
	secondNode *Node
}

type nodeRole uint8

const (
	ExpressionNode nodeRole = iota
	LicenseRefNode
	LicenseNode
)

type Node struct {
	role nodeRole
	exp  *expressionNodePartial
	lic  *licenseNodePartial
	ref  *referenceNodePartial
}

type expressionNodePartial struct {
	left        *Node
	conjunction string
	right       *Node
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

func (node *Node) IsExpression() bool {
	return node.role == ExpressionNode
}

func (node *Node) IsOrExpression() bool {
	if !node.IsExpression() {
		return false
	}
	return node.exp.conjunction == "or"
}

func (node *Node) IsAndExpression() bool {
	if !node.IsExpression() {
		return false
	}
	return node.exp.conjunction == "and"
}

func (node *Node) Left() *Node {
	if !node.IsExpression() {
		return nil
	}
	return node.exp.left
}

func (node *Node) Conjunction() *string {
	if !node.IsExpression() {
		return nil
	}
	return &(node.exp.conjunction)
}

func (node *Node) Right() *Node {
	if !node.IsExpression() {
		return nil
	}
	return node.exp.right
}

func (node *Node) IsLicense() bool {
	return node.role == LicenseNode
}

// Return the value of the license field.
// See also reconstructedLicenseString()
func (node *Node) License() *string {
	if !node.IsLicense() {
		return nil
	}
	return &(node.lic.license)
}

func (node *Node) Exception() *string {
	if !node.HasException() {
		return nil
	}
	return &(node.lic.exception)
}

func (node *Node) HasPlus() bool {
	if !node.IsLicense() {
		return false
	}
	return node.lic.hasPlus
}

func (node *Node) HasException() bool {
	if !node.IsLicense() {
		return false
	}
	return node.lic.hasException
}

func (node *Node) IsLicenseRef() bool {
	return node.role == LicenseRefNode
}

func (node *Node) LicenseRef() *string {
	if !node.IsLicenseRef() {
		return nil
	}
	return &(node.ref.licenseRef)
}

func (node *Node) DocumentRef() *string {
	if !node.HasDocumentRef() {
		return nil
	}
	return &(node.ref.documentRef)
}

func (node *Node) HasDocumentRef() bool {
	if !node.IsLicenseRef() {
		return false
	}
	return node.ref.hasDocumentRef
}

// Return the string representation of the license or license ref.
// TODO: Original had "NOASSERTION".  Does that still apply?
func (node *Node) reconstructedLicenseString() *string {
	switch node.role {
	case LicenseNode:
		license := *node.License()
		if node.HasPlus() {
			license += "+"
		}
		if node.HasException() {
			license += " WITH " + *node.Exception()
		}
		return &license
	case LicenseRefNode:
		license := "LicenseRef-" + *node.LicenseRef()
		if node.HasDocumentRef() {
			license = "DocumentRef-" + *node.DocumentRef() + ":" + license
		}
		return &license
	}
	return nil
}

// Sort an array of license and license reference nodes alphebetically based
// on their reconstructedLicenseString() representation.  The sort function does not expect
// expression nodes, but if one is in the nodes list, it will sort to the end.
func sortLicenses(nodes []*Node) {
	sort.Slice(nodes, func(i, j int) bool {
		if nodes[j].IsExpression() {
			// push second license toward end by saying first license is less than
			return true
		}
		if nodes[i].IsExpression() {
			// push first license toward end by saying second license is less than
			return false
		}
		return *nodes[i].reconstructedLicenseString() < *nodes[j].reconstructedLicenseString()
	})
}

// ---------------------- Comparator Methods ----------------------

// Return true if two licenses are compatible; otherwise, false.
func (nodes *nodePair) licensesAreCompatible() bool {
	if !nodes.firstNode.IsLicense() || !nodes.secondNode.IsLicense() {
		return false
	}
	if nodes.secondNode.HasPlus() {
		if nodes.firstNode.HasPlus() {
			// first+, second+
			return nodes.rangesAreCompatible()
		}
		// first, second+
		return nodes.identifierInRange()
	}
	// else secondNode does not have plus
	if nodes.firstNode.HasPlus() {
		// first+, second
		revNodes := &nodePair{firstNode: nodes.secondNode, secondNode: nodes.firstNode}
		return revNodes.identifierInRange()
	}
	// first, second
	return nodes.licensesExactlyEqual()
}

// Return true if two licenses are compatible in the context of their ranges; otherwise, false.
func (nodes *nodePair) rangesAreCompatible() bool {
	if nodes.licensesExactlyEqual() {
		// licenses specify ranges exactly the same
		return true
	}

	firstLicense := *nodes.firstNode.License()
	secondLicense := *nodes.secondNode.License()

	firstLicenseRange := GetLicenseRange(firstLicense)
	secondLicenseRange := GetLicenseRange(secondLicense)

	return licenseInRange(firstLicense, secondLicenseRange.licenses) &&
		licenseInRange(secondLicense, firstLicenseRange.licenses)
}

// Return true if license is found in licenseRange; otherwise, false
func licenseInRange(simpleLicense string, licenseRange []string) bool {
	for _, testLicense := range licenseRange {
		if simpleLicense == testLicense {
			return true
		}
	}
	return false
}

// Return true if the (first) simple license is in range of the (second) ranged license; otherwise, false.
func (nodes *nodePair) identifierInRange() bool {
	simpleLicense := nodes.firstNode
	plusLicense := nodes.secondNode

	return compareGT(simpleLicense, plusLicense) ||
		compareEQ(simpleLicense, plusLicense)
}

// Return true if the licenses are the same; otherwise, false
func (nodes *nodePair) licensesExactlyEqual() bool {
	return *nodes.firstNode.reconstructedLicenseString() == *nodes.secondNode.reconstructedLicenseString()
}
