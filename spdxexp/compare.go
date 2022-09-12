package spdxexp

func compareGT(first *node, second *node) bool {
	firstRange := getLicenseRange(*first.license())
	secondRange := getLicenseRange(*second.license())

	if !sameLicenseGroup(firstRange, secondRange) {
		return false
	}
	return firstRange.location[versionGroup] > secondRange.location[versionGroup]
}

func compareLT(first *node, second *node) bool {
	firstRange := getLicenseRange(*first.license())
	secondRange := getLicenseRange(*second.license())

	if !sameLicenseGroup(firstRange, secondRange) {
		return false
	}
	return firstRange.location[versionGroup] < secondRange.location[versionGroup]
}

func compareEQ(first *node, second *node) bool {
	firstRange := getLicenseRange(*first.license())
	secondRange := getLicenseRange(*second.license())

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
