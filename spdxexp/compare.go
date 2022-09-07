package spdxexp

func compareGT(first *Node, second *Node) bool {
	firstRange := GetLicenseRange(*first.License())
	secondRange := GetLicenseRange(*second.License())

	if !sameLicenseGroup(firstRange, secondRange) {
		return false
	}
	return firstRange.Location[VersionGroup] > secondRange.Location[VersionGroup]
}

func compareLT(first *Node, second *Node) bool {
	firstRange := GetLicenseRange(*first.License())
	secondRange := GetLicenseRange(*second.License())

	if !sameLicenseGroup(firstRange, secondRange) {
		return false
	}
	return firstRange.Location[VersionGroup] < secondRange.Location[VersionGroup]
}

func compareEQ(first *Node, second *Node) bool {
	firstRange := GetLicenseRange(*first.License())
	secondRange := GetLicenseRange(*second.License())

	if !sameLicenseGroup(firstRange, secondRange) {
		return false
	}
	return firstRange.Location[VersionGroup] == secondRange.Location[VersionGroup]
}

func sameLicenseGroup(firstRange *LicenseRange, secondRange *LicenseRange) bool {
	if firstRange == nil || secondRange == nil || firstRange.Location[LicenseGroup] != secondRange.Location[LicenseGroup] {
		return false
	}
	return true
}
