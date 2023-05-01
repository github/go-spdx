package main

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"reflect"
	"strings"
	"testing"
)

func TestExtractExceptionLicenseIDs(t *testing.T) {
	// Test 1: Test that the function can open and read the exceptions.json file.
	file, err := os.Open("exceptions.json")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer file.Close()

	// Test 2: Test that the function can unmarshal the exception data from the JSON file.
	var exceptionData ExceptionData
	err = json.NewDecoder(file).Decode(&exceptionData)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Test 3: Test that the function extracts only the non-deprecated exception license IDs.
	var nonDeprecatedExceptionLicenseIDs []string
	for _, e := range exceptionData.Exceptions {
		if !e.IsDeprecated {
			nonDeprecatedExceptionLicenseIDs = append(nonDeprecatedExceptionLicenseIDs, e.LicenseID)
		}
	}

	// Test 4: Test that the function writes the non-deprecated license IDs to exception_ids.json.
	err = extractExceptionLicenseIDs()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Verify that the contents of the file are correct.
	fileContents, err := ioutil.ReadFile("exception_ids.json")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var fileIDs []string
	err = json.Unmarshal(fileContents, &fileIDs)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Check if the generated slice of strings matches the original slice
	if !reflect.DeepEqual(fileIDs, nonDeprecatedExceptionLicenseIDs) {
		t.Errorf("generated IDs do not match original IDs:\n%v\n%v", fileIDs, nonDeprecatedExceptionLicenseIDs)
	}

	// Test 5: Test that the function writes the non-deprecated license IDs to exception_ids.txt.
	err = extractExceptionLicenseIDs()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Verify that the contents of the file are correct.
	fileContents, err = ioutil.ReadFile("exception_ids.txt")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	fileLines := strings.Split(string(fileContents), "\n")

	// Check each line for the correct format
	for _, line := range fileLines {
		if line == "" {
			continue
		}
		if !strings.HasPrefix(line, "		\"") || !strings.HasSuffix(line, "\",") {
			t.Errorf("unexpected line format: %v", line)
		}
	}
}
