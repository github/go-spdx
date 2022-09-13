package spdxexp

func getLicenseNode(license string, hasPlus bool) *node {
	return &node{
		role: licenseNode,
		exp:  nil,
		lic: &licenseNodePartial{
			license:      license,
			hasPlus:      hasPlus,
			hasException: false,
			exception:    "",
		},
		ref: nil,
	}
}
