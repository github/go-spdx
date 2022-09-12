package spdxexp

func getLicenseNode(license string, hasPlus bool) *node {
	return &node{
		role: licenseNode,
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
