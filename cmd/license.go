package main

import (
	"encoding/json"
	"fmt"
	"os"
)

type LicenseData struct {
	Version  string    `json:"licenseListVersion"`
	Licenses []License `json:"licenses"`
}

type License struct {
	Reference       string   `json:"reference"`
	IsDeprecated    bool     `json:"isDeprecatedLicenseId"`
	DetailsURL      string   `json:"detailsUrl"`
	ReferenceNumber int      `json:"referenceNumber"`
	Name            string   `json:"name"`
	LicenseID       string   `json:"licenseId"`
	SeeAlso         []string `json:"seeAlso"`
	IsOsiApproved   bool     `json:"isOsiApproved"`
}

// extractLicenseIDs reads the official licenses.json file copied from spdx/license-list-data
// and writes two files, license_ids.json and deprecated_license_ids.json, containing just
// the license IDs and deprecated license IDs, respectively.  It returns an error if it
// encounters one.
func extractLicenseIDs() error {
	// open file
	file, err := os.Open("licenses.json")
	if err != nil {
		fmt.Println(err)
		return err
	}

	// read in all licenses marshalled into a slice of license structs
	var licenseData LicenseData
	err = json.NewDecoder(file).Decode(&licenseData)
	if err != nil {
		fmt.Println(err)
		return err
	}

	// create two slices of license IDs, one for deprecated and one for active
	var activeLicenseIDs []string
	var deprecatedLicenseIDs []string
	for _, l := range licenseData.Licenses {
		if l.IsDeprecated {
			deprecatedLicenseIDs = append(deprecatedLicenseIDs, l.LicenseID)
		} else {
			activeLicenseIDs = append(activeLicenseIDs, l.LicenseID)
		}
	}

	// generate the GetLicenses() function in get_licenses.go
	getLicensesContents := []byte(`package spdxlicenses

func GetLicenses() []string {
	return []string{
`)
	for _, id := range activeLicenseIDs {
		getLicensesContents = append(getLicensesContents, `		"`+id+`",
`...)
	}
	getLicensesContents = append(getLicensesContents, `	}
}
`...)

	err = os.WriteFile("../spdxexp/spdxlicenses/get_licenses.go", getLicensesContents, 0600)
	if err != nil {
		return err
	}
	fmt.Println("Writing `../spdxexp/spdxlicenses/get_licenses.go`... COMPLETE")

	// generate the GetDeprecated() function in get_deprecated.go
	getDeprecatedContents := []byte(`package spdxlicenses

func GetDeprecated() []string {
	return []string{
`)
	for _, id := range deprecatedLicenseIDs {
		getDeprecatedContents = append(getDeprecatedContents, `		"`+id+`",
`...)
	}
	getDeprecatedContents = append(getDeprecatedContents, `	}
}
`...)

	err = os.WriteFile("../spdxexp/spdxlicenses/get_deprecated.go", getDeprecatedContents, 0600)
	if err != nil {
		return err
	}
	fmt.Println("Writing `../spdxexp/spdxlicenses/get_deprecated.go`... COMPLETE")

	return nil
}
