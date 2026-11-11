package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"io"
	"net/http"
	"os"
	"regexp"
	"strconv"
)

const (
	scancodePinPath = "cmd/license-categories/main.go"
	scancodeMainAPI = "https://api.github.com/repos/aboutcode-org/scancode-licensedb/commits/main"
)

var fullCommitSHA = regexp.MustCompile(`^[0-9a-f]{40}$`)

func prepareScancodeUpdate(ctx context.Context, client *http.Client, stderr io.Writer) (scancodeSource, map[string]licenseMetadata, []byte, error) {
	commit, err := latestScancodeCommit(ctx, client)
	if err != nil {
		return scancodeSource{}, nil, nil, err
	}
	if commit == currentScancodeSource.commit {
		fmt.Fprintln(stderr, "ScanCode pin is already current.") //nolint:errcheck,revive // status output
		return currentScancodeSource, nil, nil, nil
	}

	source := scancodeSource{
		commit: commit,
		zipURL: "https://codeload.github.com/aboutcode-org/scancode-licensedb/zip/" + commit,
	}
	content, err := downloadScancodeArchive(ctx, client, source.zipURL, maxZipBytes)
	if err != nil {
		return scancodeSource{}, nil, nil, err
	}
	digest := sha256.Sum256(content)
	source.zipSHA256 = hex.EncodeToString(digest[:])
	records, err := parseScancodeArchive(content, source)
	if err != nil {
		return scancodeSource{}, nil, nil, err
	}
	pins, err := os.ReadFile(scancodePinPath)
	if err != nil {
		return scancodeSource{}, nil, nil, fmt.Errorf("read scancode pin: %w", err)
	}
	pins, err = replaceScancodePins(pins, source)
	if err != nil {
		return scancodeSource{}, nil, nil, err
	}
	fmt.Fprintf(stderr, "Updating ScanCode pin: %s -> %s\n", currentScancodeSource.commit, commit) //nolint:errcheck,revive // status output
	return source, records, pins, nil
}

func latestScancodeCommit(ctx context.Context, client *http.Client) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, scancodeMainAPI, http.NoBody)
	if err != nil {
		return "", fmt.Errorf("create ScanCode main request: %w", err)
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("resolve ScanCode main: %w", err)
	}
	defer resp.Body.Close() //nolint:errcheck // best-effort close
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("resolve ScanCode main: unexpected status %s", resp.Status)
	}
	var response struct {
		SHA string `json:"sha"`
	}
	if err := json.NewDecoder(io.LimitReader(resp.Body, maxZipEntryBytes)).Decode(&response); err != nil {
		return "", fmt.Errorf("decode ScanCode main: %w", err)
	}
	if !fullCommitSHA.MatchString(response.SHA) {
		return "", fmt.Errorf("ScanCode main returned invalid full commit SHA %q", response.SHA)
	}
	return response.SHA, nil
}

func replaceScancodePins(contents []byte, source scancodeSource) ([]byte, error) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, scancodePinPath, contents, parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("parse scancode pin: %w", err)
	}
	values := map[string]string{
		"scancodeCommit":    source.commit,
		"scancodeZipSHA256": source.zipSHA256,
	}
	type replacement struct {
		start, end int
		text       string
	}
	var replacements []replacement
	for _, decl := range file.Decls {
		gen, ok := decl.(*ast.GenDecl)
		if !ok || gen.Tok != token.CONST {
			continue
		}
		for _, spec := range gen.Specs {
			value, ok := spec.(*ast.ValueSpec)
			if !ok || len(value.Names) != 1 || len(value.Values) != 1 {
				continue
			}
			name := value.Names[0].Name
			text, ok := values[name]
			if !ok {
				continue
			}
			literal, ok := value.Values[0].(*ast.BasicLit)
			if !ok || literal.Kind != token.STRING {
				return nil, fmt.Errorf("scancode pin %s must be a string literal", name)
			}
			start, end := fset.Position(literal.Pos()).Offset, fset.Position(literal.End()).Offset
			replacements = append(replacements, replacement{start, end, strconv.Quote(text)})
			delete(values, name)
		}
	}
	if len(values) != 0 {
		return nil, fmt.Errorf("missing scancode pin constants: %v", values)
	}
	// Work backwards so changed literal lengths cannot shift later offsets.
	for i := len(replacements) - 1; i >= 0; i-- {
		edit := replacements[i]
		updated := append([]byte{}, contents[:edit.start]...)
		updated = append(updated, edit.text...)
		contents = append(updated, contents[edit.end:]...)
	}
	return format.Source(contents)
}
