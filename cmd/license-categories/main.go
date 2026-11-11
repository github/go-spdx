// Command license-categories generates TypeScript data for the license-categories package.
package main

import (
	"archive/zip"
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"runtime/debug"
	"slices"
	"strings"
	"time"

	"github.com/github/go-spdx/v2/spdxexp/spdxlicenses"
)

const (
	scancodeCommit           = "5ee8765870cd2ed1cb345645be94a799ca584c90"
	defaultScancodeZipURL    = "https://codeload.github.com/aboutcode-org/scancode-licensedb/zip/" + scancodeCommit
	defaultScancodeZipSHA256 = "6b4b05f985bd7a520b30e7f8547778ae4cdd6760acc577ab11cebbde41f165bb"
	defaultTimeout           = 60 * time.Second
	maxZipBytes              = 100 * 1024 * 1024 // 100MB upper bound for the scancode zip download.
	maxZipEntryBytes         = 1 * 1024 * 1024   // 1MB upper bound per zip entry to guard against zip bombs.
	outputFileMode           = 0o644             // World-readable since the generated file is manually copied into other projects.
)

// excludedSpdxIDs contains SPDX IDs that should be ignored when generating license categories.
// Keep this hardcoded list explicit and reviewed.
var excludedSpdxIDs = map[string]struct{}{}

// spdxCategoryOverrides allows manually overriding category assignments for specific SPDX IDs.
// Keep this list explicit and reviewed for exceptions where upstream category data is not helpful.
var spdxCategoryOverrides = map[string]string{
	"any-OSI":                    "Permissive",       // "any-OSI" just means any OSI-approved license, so people could choose a permissive one from that set.
	"any-OSI-perl-modules":       "Permissive",       // same rationale as "any-OSI"
	"ESA-PL-strong-copyleft-2.4": "Copyleft",         // incorrectly categorized as Permissive
	"ESA-PL-weak-copyleft-2.4":   "Copyleft limited", // incorrectly categorized as Permissive
	"MakeIndex":                  "Copyleft",         // Unclear why the category didn't make it across: https://github.com/aboutcode-org/scancode-licensedb/blob/main/docs/makeindex.LICENSE
}

var categoryDescriptions = map[string]string{
	"Commercial":       "High risk",
	"Copyleft":         "High risk",
	"Non-commercial":   "High risk",
	"Copyleft limited": "Review recommended",
	"Free restricted":  "Review recommended",
	"Proprietary free": "Review recommended",
	"Permissive":       "Safe for most",
	"Public domain":    "Safe for most",
	"Source available": "Review recommended",
	"Unknown":          "Review recommended",
}

type licenseMetadata struct {
	Name      string
	Category  string
	SourceURL string
}

type scancodeLicense struct {
	Name           string `json:"name"`
	Category       string `json:"category"`
	SPDXLicenseKey string `json:"spdx_license_key"`
}

type scancodeSource struct {
	commit    string
	zipURL    string
	zipSHA256 string
}

var defaultScancodeSource = scancodeSource{
	commit:    scancodeCommit,
	zipURL:    defaultScancodeZipURL,
	zipSHA256: defaultScancodeZipSHA256,
}

type generationProvenance struct {
	GoSPDXVersion          string
	SPDXLicenseListVersion string
	ScancodeCommit         string
}

func main() {
	err := run(os.Args[1:], os.Stdout, os.Stderr, &http.Client{Timeout: defaultTimeout})
	if err != nil {
		fmt.Fprintf(os.Stderr, "license-categories failed: %v\n", err) //nolint:revive // error path output
		os.Exit(1)
	}
}

func run(args []string, stdout, stderr io.Writer, client *http.Client) error {
	fs := flag.NewFlagSet("license-categories", flag.ContinueOnError)
	fs.SetOutput(stderr)

	outPath := fs.String("out", "-", "Output file path, or '-' for stdout")
	err := fs.Parse(args)
	if err != nil {
		return err
	}

	spdxIDs := spdxlicenses.GetLicenses()
	records, err := loadScancodeBySpdx(context.Background(), client, defaultScancodeSource, maxZipBytes)
	if err != nil {
		return fmt.Errorf("load scancode data: %w", err)
	}

	licenses, warnings, err := buildLicenseData(spdxIDs, records, excludedSpdxIDs, spdxCategoryOverrides)
	if err != nil {
		return err
	}

	for _, warning := range warnings {
		_, _ = fmt.Fprintf(stderr, "warning: %s\n", warning) //nolint:revive // warning output
	}

	goSPDXVersion, err := moduleVersion()
	if err != nil {
		return err
	}

	output, err := renderTypeScript(licenses, generationProvenance{
		GoSPDXVersion:          goSPDXVersion,
		SPDXLicenseListVersion: spdxlicenses.LicenseListVersion(),
		ScancodeCommit:         defaultScancodeSource.commit,
	})
	if err != nil {
		return err
	}

	if *outPath == "-" {
		_, err = stdout.Write(output)
		return err
	}

	err = os.WriteFile(*outPath, output, outputFileMode)
	if err != nil {
		return fmt.Errorf("write output: %w", err)
	}
	return nil
}

func loadScancodeBySpdx(ctx context.Context, client *http.Client, source scancodeSource, maxZipBytes int64) (map[string]licenseMetadata, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, source.zipURL, http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("download zip: %w", err)
	}
	defer resp.Body.Close() //nolint:errcheck // best-effort close

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status downloading zip: %s", resp.Status)
	}

	content, err := io.ReadAll(io.LimitReader(resp.Body, maxZipBytes+1))
	if err != nil {
		return nil, fmt.Errorf("read zip response: %w", err)
	}

	if int64(len(content)) > maxZipBytes {
		return nil, fmt.Errorf("zip response exceeds maximum size of %d bytes", maxZipBytes)
	}
	if err := verifySHA256(content, source.zipSHA256); err != nil {
		return nil, err
	}

	reader, err := zip.NewReader(bytes.NewReader(content), int64(len(content)))
	if err != nil {
		return nil, fmt.Errorf("open zip: %w", err)
	}

	records := make(map[string]licenseMetadata)
	for _, file := range reader.File {
		if !isScancodeLicenseJSON(file.Name) {
			continue
		}

		if file.UncompressedSize64 > maxZipEntryBytes {
			return nil, fmt.Errorf("zip entry %s exceeds maximum uncompressed size (%d bytes)", file.Name, maxZipEntryBytes)
		}

		rc, err := file.Open()
		if err != nil {
			return nil, fmt.Errorf("open zip file %s: %w", file.Name, err)
		}

		var parsed scancodeLicense
		decodeErr := json.NewDecoder(io.LimitReader(rc, maxZipEntryBytes+1)).Decode(&parsed)
		closeErr := rc.Close()
		if decodeErr != nil {
			return nil, fmt.Errorf("decode %s: %w", file.Name, decodeErr)
		}
		if closeErr != nil {
			return nil, fmt.Errorf("close %s: %w", file.Name, closeErr)
		}

		spdxID := strings.TrimSpace(parsed.SPDXLicenseKey)
		if spdxID == "" {
			continue
		}

		if _, exists := records[spdxID]; !exists {
			records[spdxID] = licenseMetadata{
				Name:      strings.TrimSpace(parsed.Name),
				Category:  strings.TrimSpace(parsed.Category),
				SourceURL: scancodeGitHubURL(file.Name, source.commit),
			}
		}
	}

	if len(records) == 0 {
		return nil, fmt.Errorf("no scancode SPDX records found in zip %q", source.zipURL)
	}

	return records, nil
}

func verifySHA256(content []byte, expected string) error {
	digest := sha256.Sum256(content)
	actual := hex.EncodeToString(digest[:])
	if !strings.EqualFold(actual, expected) {
		return fmt.Errorf("scancode archive checksum mismatch: got %s, want %s", actual, expected)
	}
	return nil
}

func moduleVersion() (string, error) {
	buildInfo, ok := debug.ReadBuildInfo()
	if !ok {
		return "", errors.New("read Go build information")
	}
	return buildInfo.Main.Version, nil
}

func scancodeGitHubURL(zipPath, commit string) string {
	_, relativePath, found := strings.Cut(zipPath, "/")
	if !found {
		return zipPath
	}
	return "https://github.com/aboutcode-org/scancode-licensedb/blob/" + commit + "/" + relativePath
}

func isScancodeLicenseJSON(path string) bool {
	if !strings.HasSuffix(path, ".json") {
		return false
	}
	if strings.HasSuffix(path, "index.json") {
		return false
	}
	// GitHub zip adds a top-level "<repo>-<branch>/" directory.
	// The API JSON files currently live in docs/.
	return strings.Contains(path, "/docs/")
}

func buildLicenseData(
	spdxIDs []string,
	scancodeBySpdx map[string]licenseMetadata,
	excluded map[string]struct{},
	overrides map[string]string,
) (map[string]licenseMetadata, []string, error) {
	output := make(map[string]licenseMetadata, len(spdxIDs))
	warnings := make([]string, 0)

	for _, spdxID := range spdxIDs {
		if _, isExcluded := excluded[spdxID]; isExcluded {
			warnings = append(warnings, fmt.Sprintf("excluded SPDX license %q", spdxID))
			continue
		}

		// overrides work even if the license is missing from scancode
		if override, ok := overrides[spdxID]; ok {
			if _, ok := categoryDescriptions[override]; !ok {
				return nil, nil, fmt.Errorf("override category %q for %s has no description mapping", override, spdxID)
			}
			meta := licenseMetadata{
				Name:     spdxID,
				Category: override,
			}
			if record, found := scancodeBySpdx[spdxID]; found {
				meta.Name = record.Name
				meta.SourceURL = record.SourceURL
			}
			if meta.Name == "" {
				meta.Name = spdxID
			}
			output[spdxID] = meta
			continue
		}

		record, found := scancodeBySpdx[spdxID]
		if !found {
			output[spdxID] = licenseMetadata{
				Name:     spdxID,
				Category: "Unknown",
			}
			warnings = append(warnings, fmt.Sprintf("missing scancode metadata for %q, using Unknown category", spdxID))
			continue
		}

		mappedCategory := mapScancodeCategory(record.Category)

		if _, ok := categoryDescriptions[mappedCategory]; !ok {
			if record.SourceURL != "" {
				return nil, nil, fmt.Errorf("category %q for %s has no description mapping (source: %s)", mappedCategory, spdxID, record.SourceURL)
			}
			return nil, nil, fmt.Errorf("category %q for %s has no description mapping", mappedCategory, spdxID)
		}

		name := record.Name
		if name == "" {
			name = spdxID
		}

		output[spdxID] = licenseMetadata{
			Name:      name,
			Category:  mappedCategory,
			SourceURL: record.SourceURL,
		}
	}

	slices.Sort(warnings)
	return output, warnings, nil
}

func mapScancodeCategory(category string) string {
	switch category {
	case "Source-available":
		return "Source available"
	case "Free Restricted":
		return "Free restricted"
	case "Copyleft Limited":
		return "Copyleft limited"
	case "Public Domain":
		return "Public domain"
	case "Non-Commercial":
		return "Non-commercial"
	case "Proprietary Free":
		return "Proprietary free"
	case "":
		return "Unknown"
	default:
		return category
	}
}

func renderTypeScript(licensesByID map[string]licenseMetadata, provenance generationProvenance) ([]byte, error) {
	ids := make([]string, 0, len(licensesByID))
	categorySet := map[string]struct{}{}
	for id, meta := range licensesByID {
		ids = append(ids, id)
		categorySet[meta.Category] = struct{}{}
	}
	slices.Sort(ids)

	categories := make([]string, 0, len(categorySet))
	for category := range categorySet {
		categories = append(categories, category)
	}
	slices.Sort(categories)

	lines := []string{
		"// Code generated by cmd/license-categories in go-spdx. DO NOT EDIT.",
		"// go-spdx: " + provenance.GoSPDXVersion,
		"// spdx-license-list-version: " + provenance.SPDXLicenseListVersion,
		"// scancode-licensedb commit: " + provenance.ScancodeCommit,
		"",
		"export const categoryDescriptions: Record<string, string> = {",
	}
	for _, category := range categories {
		description, ok := categoryDescriptions[category]
		if !ok {
			return nil, fmt.Errorf("missing description for category %q", category)
		}
		lines = append(lines, "  "+singleQuote(category)+": "+singleQuote(description)+",")
	}
	lines = append(lines, "}", "", "export const licensesById: Record<string, {name: string; category: string}> = {")
	for _, id := range ids {
		meta := licensesByID[id]
		lines = append(lines, "  "+singleQuote(id)+": {name: "+singleQuote(meta.Name)+", category: "+singleQuote(meta.Category)+"},")
	}
	lines = append(lines, "}")

	return []byte(strings.Join(lines, "\n") + "\n"), nil
}

func singleQuote(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, "'", `\'`)
	return "'" + s + "'"
}
