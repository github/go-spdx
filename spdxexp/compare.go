package spdxexp

func compareGT(first *Node, second *Node) bool {
	firstRange := GetLicenseRange(*first.License())
	secondRange := GetLicenseRange(*second.License())

	if !sameLicenseGroup(firstRange, secondRange) {
		return false
	}
	return firstRange.location[versionGroup] > secondRange.location[versionGroup]
}

func compareLT(first *Node, second *Node) bool {
	firstRange := GetLicenseRange(*first.License())
	secondRange := GetLicenseRange(*second.License())

	if !sameLicenseGroup(firstRange, secondRange) {
		return false
	}
	return firstRange.location[versionGroup] < secondRange.location[versionGroup]
}

func compareEQ(first *Node, second *Node) bool {
	firstRange := GetLicenseRange(*first.License())
	secondRange := GetLicenseRange(*second.License())

	if !sameLicenseGroup(firstRange, secondRange) {
		return false
	}
	return firstRange.location[versionGroup] == secondRange.location[versionGroup]
}

func sameLicenseGroup(firstRange *licenseRange, secondRange *licenseRange) bool {
	if firstRange == nil || secondRange == nil || firstRange.location[licenseGroup] != secondRange.location[licenseGroup] {
		return false
	}
	return true
}
