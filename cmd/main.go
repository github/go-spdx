package main

import (
	"flag"
	"fmt"
	"os"
)

func main() {
	cmd := "extract"

	if len(os.Args) > 1 {
		cmd = os.Args[1]
	}
	argsRemainder := []string{}
	if len(os.Args) > 2 {
		argsRemainder = os.Args[2:]
	}

	flagSet := flag.NewFlagSet("run", flag.ExitOnError)
	shouldExtractLicenses := flagSet.Bool("l", false, "Should licenses and ranges be extracted?")
	extractExceptions := flagSet.Bool("e", false, "Should exception ids be extracted?")
	help := flagSet.Bool("h", false, "Show help")

	err := flagSet.Parse(argsRemainder)
	if err != nil {
		fmt.Printf("error parsing flags for run: %v\n", err)
		os.Exit(1)
	}

	switch cmd {
	case "extract":
		if *help || (!*shouldExtractLicenses && !*extractExceptions) {
			writeHelpMessage()
			os.Exit(0)
		}
		if *shouldExtractLicenses {
			fmt.Println("-------------------------")
			fmt.Println("Extracting licenses and ranges...")
			err := extractLicenses()
			if err != nil {
				fmt.Printf("error extracting licenses and ranges: %v\n", err)
				os.Exit(1)
			}
			fmt.Println("Done!")
		}
		if *extractExceptions {
			fmt.Println("---------------------------")
			fmt.Println("Extracting exception ids...")
			err := extractExceptionLicenseIDs()
			if err != nil {
				fmt.Printf("error extracting exception ids: %v\n", err)
				os.Exit(1)
			}
			fmt.Println("Done!")
			fmt.Println("---------------------------")
		}
	default:
		writeHelpMessage()
		os.Exit(0)
	}
}

func writeHelpMessage() {
	fmt.Println("")
	fmt.Println("Extracts license, deprecation, and exception ids from the official spdx license list data.")
	fmt.Println("The source data needs to be manually updated by copying the licenses.json file from")
	fmt.Println("https://github.com/spdx/license-list-data/blob/main/json/licenses.json and exceptions.json")
	fmt.Println("file from https://github.com/spdx/license-list-data/blob/main/json/exceptions.json.")
	fmt.Println("")
	fmt.Println("After running the extract command, the generated license, deprecated license, exception, and license")
	fmt.Println("range Go files in spdxexp/spdxlicenses will be overwritten with data derived from the SPDX license list.")
	fmt.Println("")
	fmt.Println("Command to run all extractions (run command from the /cmd directory):")
	fmt.Println("  `go run . extract -l -e`")
	fmt.Println("")
	fmt.Println("Usage options:")
	fmt.Println("  -h: prints this help message")
	fmt.Println("  -l: Extract licenses and ranges")
	fmt.Println("  -e: Extract exception ids")
	fmt.Println("")
	os.Exit(0)
}
