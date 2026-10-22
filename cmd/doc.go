/*
Extracts license, deprecation, and exception ids from the official spdx license list data.
The source data needs to be manually updated by copying the licenses.json file from
https://github.com/spdx/license-list-data/blob/main/json/licenses.json and exceptions.json
file from https://github.com/spdx/license-list-data/blob/main/json/exceptions.json.

After running the extract command, the generated license, deprecated license, exception, and license
range Go files in spdxexp/spdxlicenses will be overwritten with data derived from the SPDX license list.

Command to run all extractions (run command from the /cmd directory):

	cd cmd
	go run . extract -l -e

Usage options:

	-h: prints this help message
	-l: Extract licenses and ranges
	-e: Extract exception ids
*/
package main
