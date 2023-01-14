package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
)

type ExceptionData struct {
	Version    string      `json:"licenseListVersion"`
	Exceptions []Exception `json:"exceptions"`
}

type Exception struct {
	Reference       string   `json:"reference"`
	IsDeprecated    bool     `json:"isDeprecatedLicenseId"`
	DetailsURL      string   `json:"detailsUrl"`
	ReferenceNumber int      `json:"referenceNumber"`
	Name            string   `json:"name"`
	LicenseID       string   `json:"licenseExceptionId"`
	SeeAlso         []string `json:"seeAlso"`
	IsOsiApproved   bool     `json:"isOsiApproved"`
}

// extractExceptionLicenseIDs read official exception licenses file copied from spdx/license-list-data
// and write file exception_license_ids.json containing just the non-deprecated exception license IDs.
// NOTE: For now, this function ignores the deprecated exception licenses.
func extractExceptionLicenseIDs() error {
	// open file
	file, err := os.Open("exceptions.json")
	if err != nil {
		fmt.Println(err)
		return err
	}

	// read in all licenses marshalled into a slice of exception structs
	var exceptionData ExceptionData
	err = json.NewDecoder(file).Decode(&exceptionData)
	if err != nil {
		return err
	}

	// create slice of exception license IDs that are not deprecated
	var nonDeprecatedExceptionLicenseIDs []string
	for _, e := range exceptionData.Exceptions {
		if !e.IsDeprecated {
			nonDeprecatedExceptionLicenseIDs = append(nonDeprecatedExceptionLicenseIDs, e.LicenseID)
		}
	}

	// save non-deprecated license IDs followed by a comma with one per line in file license_ids.txt
	nonDeprecatedExceptionLicenseIDsTxt := []byte{}
	for _, id := range nonDeprecatedExceptionLicenseIDs {
		nonDeprecatedExceptionLicenseIDsTxt = append(nonDeprecatedExceptionLicenseIDsTxt, []byte("		\"")...)
		nonDeprecatedExceptionLicenseIDsTxt = append(nonDeprecatedExceptionLicenseIDsTxt, []byte(id)...)
		nonDeprecatedExceptionLicenseIDsTxt = append(nonDeprecatedExceptionLicenseIDsTxt, []byte("\",")...)
		nonDeprecatedExceptionLicenseIDsTxt = append(nonDeprecatedExceptionLicenseIDsTxt, []byte("\n")...)
	}
	err = ioutil.WriteFile("exception_ids.txt", nonDeprecatedExceptionLicenseIDsTxt, 0600)
	if err != nil {
		return err
	}

	// save non-deprecated license IDs to json array in file exception_ids.json
	nonDeprecatedExceptionLicenseIDsJSON, err := json.Marshal(nonDeprecatedExceptionLicenseIDs)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile("exception_ids.json", nonDeprecatedExceptionLicenseIDsJSON, 0600)
	if err != nil {
		return err
	}
	return nil
}
