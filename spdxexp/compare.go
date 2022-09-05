package spdxexp

func compareGT(first *Node, second *Node) bool {
	firstRange := GetLicenseRange(*first.License())
	secondRange := GetLicenseRange(*second.License())

	if !sameLicenseGroup(firstRange, secondRange) {
		return false
	}
	return firstRange.Location[VERSION_GROUP] > secondRange.Location[VERSION_GROUP]
}

func compareLT(first *Node, second *Node) bool {
	firstRange := GetLicenseRange(*first.License())
	secondRange := GetLicenseRange(*second.License())

	if !sameLicenseGroup(firstRange, secondRange) {
		return false
	}
	return firstRange.Location[VERSION_GROUP] < secondRange.Location[VERSION_GROUP]
}

func compareEQ(first *Node, second *Node) bool {
	firstRange := GetLicenseRange(*first.License())
	secondRange := GetLicenseRange(*second.License())

	if !sameLicenseGroup(firstRange, secondRange) {
		return false
	}
	return firstRange.Location[VERSION_GROUP] == secondRange.Location[VERSION_GROUP]
}

func sameLicenseGroup(firstRange *LicenseRange, secondRange *LicenseRange) bool {
	if firstRange == nil || secondRange == nil || firstRange.Location[LICENSE_GROUP] != secondRange.Location[LICENSE_GROUP] {
		return false
	}
	return true
}
