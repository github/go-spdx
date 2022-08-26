package spdxexp

func getLicenseNode(license string, hasPlus bool) *Node {
	return &Node{
		role: LICENSE_NODE,
		exp:  nil,
		lic: &licenseNodePartial{
			license:      license,
			hasPlus:      false,
			hasException: false,
			exception:    "",
		},
		ref: nil,
	}
}
