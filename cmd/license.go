package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
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

	// create two slices of license IDs, one for deprecated and one for not deprecated
	var deprecatedLicenseIDs []string
	var nonDeprecatedLicenseIDs []string
	for _, l := range licenseData.Licenses {
		if l.IsDeprecated {
			deprecatedLicenseIDs = append(deprecatedLicenseIDs, l.LicenseID)
		} else {
			nonDeprecatedLicenseIDs = append(nonDeprecatedLicenseIDs, l.LicenseID)
		}
	}

	// save deprecated license IDs followed by a comma with one per line in file deprecated_license_ids.txt
	deprecatedLicenseIDsTxt := []byte{}
	for _, id := range deprecatedLicenseIDs {
		deprecatedLicenseIDsTxt = append(deprecatedLicenseIDsTxt, []byte("		\"")...)
		deprecatedLicenseIDsTxt = append(deprecatedLicenseIDsTxt, []byte(id)...)
		deprecatedLicenseIDsTxt = append(deprecatedLicenseIDsTxt, []byte("\",")...)
		deprecatedLicenseIDsTxt = append(deprecatedLicenseIDsTxt, []byte("\n")...)
	}
	err = ioutil.WriteFile("deprecated_license_ids.txt", deprecatedLicenseIDsTxt, 0600)
	if err != nil {
		return err
	}

	// save deprecated license IDs to json array in file deprecated_license_ids.json
	deprecatedLicenseIDsJSON, err := json.Marshal(deprecatedLicenseIDs)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile("deprecated_license_ids.json", deprecatedLicenseIDsJSON, 0600)
	if err != nil {
		return err
	}

	// save non-deprecated license IDs followed by a comma with one per line in file license_ids.txt
	nonDeprecatedLicenseIDsTxt := []byte{}
	for _, id := range nonDeprecatedLicenseIDs {
		nonDeprecatedLicenseIDsTxt = append(nonDeprecatedLicenseIDsTxt, []byte("		\"")...)
		nonDeprecatedLicenseIDsTxt = append(nonDeprecatedLicenseIDsTxt, []byte(id)...)
		nonDeprecatedLicenseIDsTxt = append(nonDeprecatedLicenseIDsTxt, []byte("\",")...)
		nonDeprecatedLicenseIDsTxt = append(nonDeprecatedLicenseIDsTxt, []byte("\n")...)
	}
	err = ioutil.WriteFile("license_ids.txt", nonDeprecatedLicenseIDsTxt, 0600)
	if err != nil {
		return err
	}

	// save non-deprecated license IDs to json array in file license_ids.json
	nonDeprecatedLicenseIDsJSON, err := json.Marshal(nonDeprecatedLicenseIDs)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile("license_ids.json", nonDeprecatedLicenseIDsJSON, 0600)
	if err != nil {
		return err
	}
	return nil
}
