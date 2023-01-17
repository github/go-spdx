/*
Extracts license, deprecation, and exception ids from the official spdx license list data.
The source data needs to be manually updated by copying the licenses.json file from
https://github.com/spdx/license-list-data/blob/main/json/licenses.json and exceptions.json
file from https://github.com/spdx/license-list-data/blob/main/json/exceptions.json.

After running the extract command, the license_ids.json, deprecated_ids.json, and exception_ids.json
files will be overwritten with the extracted ids.  These license ids can then be used to update the
spdxexp/license.go file.

Command to run all extractions (run command from the /cmd directory):

	go run . extract -l -e

Usage options:

	-h: prints this help message
	-l: Extract license ids
	-e: Extract exception ids
*/
package main
