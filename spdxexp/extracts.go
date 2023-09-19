package spdxexp

// ExtractLicenses extracts licenses from the given expression without duplicates.
// Returns an array of licenses or error if error occurs during processing.
func ExtractLicenses(expression string) ([]string, error) {
	node, err := parse(expression)
	if err != nil {
		return nil, err
	}

	expanded := node.expand(true)
	licenses := make([]string, 0)
	allLicenses := flatten(expanded)
	for _, licenseNode := range allLicenses {
		licenses = append(licenses, *licenseNode.reconstructedLicenseString())
	}

	licenses = removeDuplicateStrings(licenses)

	return licenses, nil
}
