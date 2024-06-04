package main

import (
	"encoding/json"
	"fmt"
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
	var exceptionLicenseIDs []string
	for _, e := range exceptionData.Exceptions {
		if !e.IsDeprecated {
			exceptionLicenseIDs = append(exceptionLicenseIDs, e.LicenseID)
		}
	}

	// generate the GetExceptions() function in get_exceptions.go
	getExceptionsContents := []byte(`package spdxlicenses

func GetExceptions() []string {
	return []string{
`)
	for _, id := range exceptionLicenseIDs {
		getExceptionsContents = append(getExceptionsContents, `		"`+id+`",
`...)
	}
	getExceptionsContents = append(getExceptionsContents, `	}
}
`...)

	err = os.WriteFile("../spdxexp/spdxlicenses/get_exceptions.go", getExceptionsContents, 0600)
	if err != nil {
		return err
	}
	fmt.Println("Writing `../spdxexp/spdxlicenses/get_exceptions.go`... COMPLETE")

	return nil
}
