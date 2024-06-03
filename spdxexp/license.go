package spdxexp

import (
	"strings"

	"github.com/github/go-spdx/v2/spdxexp/spdxlicenses"
)

// activeLicense returns true if the id is an active license.
func activeLicense(id string) (bool, string) {
	return inLicenseList(spdxlicenses.GetLicenses(), id)
}

// deprecatedLicense returns true if the id is a deprecated license.
func deprecatedLicense(id string) (bool, string) {
	return inLicenseList(spdxlicenses.GetDeprecated(), id)
}

// exceptionLicense returns true if the id is an exception license.
func exceptionLicense(id string) (bool, string) {
	return inLicenseList(spdxlicenses.GetExceptions(), id)
}

// inLicenseList looks for id in the list of licenses.  The check is case-insensitive (e.g. "mit" will match "MIT").
func inLicenseList(licenses []string, id string) (bool, string) {
	for _, license := range licenses {
		if strings.EqualFold(license, id) {
			return true, license
		}
	}
	return false, id
}

const (
	licenseGroup uint8 = iota
	versionGroup
	licenseIndex
)

type licenseRange struct {
	licenses []string
	location map[uint8]int // licenseGroup, versionGroup, licenseIndex
}

// getLicenseRange returns a range of licenses from licenseRanges
func getLicenseRange(id string) *licenseRange {
	simpleID := simplifyLicense(id)
	allRanges := spdxlicenses.LicenseRanges()
	for i, licenseGrp := range allRanges {
		for j, versionGrp := range licenseGrp {
			for k, license := range versionGrp {
				if simpleID == license {
					location := map[uint8]int{
						licenseGroup: i,
						versionGroup: j,
						licenseIndex: k,
					}
					return &licenseRange{
						licenses: versionGrp,
						location: location,
					}
				}
			}
		}
	}
	return nil
}

func simplifyLicense(id string) string {
	if strings.HasSuffix(id, "-or-later") {
		return id[0 : len(id)-9]
	}
	return id
}
