package spdxexp

type NodePair struct {
	firstNode  *Node
	secondNode *Node
}

type nodeRole uint8

const (
	EXPRESSION_NODE nodeRole = iota
	LICENSEREF_NODE
	LICENSE_NODE
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
	return node.role == EXPRESSION_NODE
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

func (node *Node) Right() *Node {
	if !node.IsExpression() {
		return nil
	}
	return node.exp.right
}

func (node *Node) IsLicense() bool {
	return node.role == LICENSE_NODE
}

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
	return node.role == LICENSEREF_NODE
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

// ---------------------- Comparator Methods ----------------------

// Return true if two licenses are compatible; otherwise, false.
func (nodes *NodePair) LicensesAreCompatible() bool {
	if !nodes.firstNode.IsLicense() || !nodes.secondNode.IsLicense() {
		return false
	}
	if nodes.secondNode.HasPlus() {
		if nodes.firstNode.HasPlus() {
			// first+, second+
			return nodes.rangesAreCompatible()
		} else {
			// first, second+
			return nodes.identifierInRange()
		}
	} else {
		if nodes.firstNode.HasPlus() {
			// first+, second
			rev_nodes := &NodePair{firstNode: nodes.secondNode, secondNode: nodes.firstNode}
			return rev_nodes.identifierInRange()
		} else {
			// first, second
			return nodes.licensesExactlyEqual()
		}
	}
}

// Return true if two licenses are compatible in the context of their ranges; otherwise, false.
func (nodes *NodePair) rangesAreCompatible() bool {
	if nodes.licensesExactlyEqual() {
		// licenses specify ranges exactly the same
		return true
	}

	firstLicense := *nodes.firstNode.License()
	secondLicense := *nodes.secondNode.License()

	firstLicenseRange := GetLicenseRange(firstLicense)
	secondLicenseRange := GetLicenseRange(secondLicense)

	return licenseInRange(firstLicense, secondLicenseRange.Licenses) &&
		licenseInRange(secondLicense, firstLicenseRange.Licenses)
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
func (nodes *NodePair) identifierInRange() bool {
	simpleLicense := nodes.firstNode
	plusLicense := nodes.secondNode

	return compareGT(simpleLicense, plusLicense) ||
		compareEQ(simpleLicense, plusLicense)
}

// Return true if the licenses are the same; otherwise, false
func (nodes *NodePair) licensesExactlyEqual() bool {
	return *nodes.firstNode.License() == *nodes.secondNode.License()
}
