package main

import (
	"archive/zip"
	"bytes"
	"context"
	"crypto/sha256"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type scancodeTransport func(*http.Request) (*http.Response, error)

func (transport scancodeTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	return transport(req)
}

func scancodeArchive(t *testing.T, data string) []byte {
	t.Helper()
	var content bytes.Buffer
	writer := zip.NewWriter(&content)
	entry, err := writer.Create("licensedb-revision/docs/mit.json")
	require.NoError(t, err)
	_, err = io.WriteString(entry, data)
	require.NoError(t, err)
	require.NoError(t, writer.Close())
	return content.Bytes()
}

func TestLatestScancodeCommit(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name, body, wantError string
		status                int
	}{
		{"valid", `{"sha":"` + strings.Repeat("a", 40) + `"}`, "", 200},
		{"rate limit", `{}`, "unexpected status", 403},
		{"server failure", `{}`, "unexpected status", 500},
		{"malformed", `<html>`, "decode ScanCode", 200},
		{"missing", `{}`, "invalid full commit SHA", 200},
		{"short", `{"sha":"abc123"}`, "invalid full commit SHA", 200},
		{"unsafe", `{"sha":"refs/heads/main"}`, "invalid full commit SHA", 200},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			client := &http.Client{Transport: scancodeTransport(func(req *http.Request) (*http.Response, error) {
				assert.Equal(t, scancodeMainAPI, req.URL.String())
				return &http.Response{StatusCode: tt.status, Status: fmt.Sprint(tt.status), Body: io.NopCloser(strings.NewReader(tt.body))}, nil
			})}
			commit, err := latestScancodeCommit(context.Background(), client)
			if tt.wantError != "" {
				require.ErrorContains(t, err, tt.wantError)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, strings.Repeat("a", 40), commit)
		})
	}
}

func TestReplaceScancodePins(t *testing.T) {
	t.Parallel()
	input := []byte(`package main
// Keep this comment.
const (
	scancodeCommit = "old"
	scancodeZipURL = "https://codeload.github.com/aboutcode-org/scancode-licensedb/zip/" + scancodeCommit
	scancodeZipSHA256 = "old-checksum"
)
var spdxCategoryOverrides = map[string]string{"MIT": "Permissive"}
`)
	source := scancodeSource{commit: strings.Repeat("a", 40), zipSHA256: strings.Repeat("b", 64)}
	output, err := replaceScancodePins(input, source)
	require.NoError(t, err)
	assert.Contains(t, string(output), `"https://codeload.github.com/aboutcode-org/scancode-licensedb/zip/" + scancodeCommit`)
	assert.Contains(t, string(output), strconv.Quote(source.commit))
	assert.Contains(t, string(output), strconv.Quote(source.zipSHA256))
	assert.Contains(t, string(output), "// Keep this comment.")
	assert.Contains(t, string(output), `map[string]string{"MIT": "Permissive"}`)
	_, err = replaceScancodePins([]byte("package main\n"), source)
	require.ErrorContains(t, err, "missing scancode pin constants")
	_, err = replaceScancodePins([]byte("package main\nconst scancodeCommit = other\n"), source)
	require.ErrorContains(t, err, "must be a string literal")
}

func TestRunUpdateScancode(t *testing.T) {
	// Chdir and the synthetic current source are process-wide; do not run in parallel.
	original, err := os.ReadFile("main.go")
	require.NoError(t, err)
	const newCommit = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	valid := scancodeArchive(t, `{"name":"MIT License","category":"Permissive","spdx_license_key":"MIT"}`)
	for _, tt := range []struct {
		name, commit, wantError string
		archive                 []byte
		matchingChecksum        bool
	}{
		{"new revision", newCommit, "", valid, false},
		{"unchanged revision", currentScancodeSource.commit, "", valid, true},
		{"unchanged checksum mismatch", currentScancodeSource.commit, "checksum mismatch", valid, false},
		{"bad zip", newCommit, "open zip", []byte("not a zip"), false},
		{"bad json", newCommit, "decode", scancodeArchive(t, "{"), false},
		{"empty records", newCommit, "no scancode SPDX records", scancodeArchive(t, `{}`), false},
		{"unrecognized category", newCommit, "has no description mapping", scancodeArchive(t, `{"name":"MIT","category":"New risk","spdx_license_key":"MIT"}`), false},
	} {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			require.NoError(t, os.MkdirAll(filepath.Join(dir, "cmd/license-categories"), 0o750))
			// #nosec G703 -- t.TempDir and the constant fixture path confine this write to the test directory.
			require.NoError(t, os.WriteFile(filepath.Join(dir, scancodePinPath), original, 0o600))
			t.Chdir(dir)
			if tt.matchingChecksum {
				previous := currentScancodeSource
				t.Cleanup(func() { currentScancodeSource = previous })
				currentScancodeSource.zipSHA256 = fmt.Sprintf("%x", sha256.Sum256(tt.archive))
			}
			var requested []string
			client := &http.Client{Transport: scancodeTransport(func(req *http.Request) (*http.Response, error) {
				requested = append(requested, req.URL.String())
				body := tt.archive
				if req.URL.String() == scancodeMainAPI {
					body = []byte(`{"sha":"` + tt.commit + `"}`)
				}
				return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(bytes.NewReader(body))}, nil
			})}
			var warnings bytes.Buffer
			err := run([]string{"-update-scancode", "-out", "generated.ts"}, io.Discard, &warnings, client)
			pins, readErr := os.ReadFile(scancodePinPath)
			require.NoError(t, readErr)
			if tt.wantError != "" {
				require.ErrorContains(t, err, tt.wantError)
				assert.Equal(t, original, pins)
				_, err = os.Stat("generated.ts")
				require.True(t, os.IsNotExist(err))
				return
			}
			require.NoError(t, err)
			assert.Equal(t, []string{scancodeMainAPI, "https://codeload.github.com/aboutcode-org/scancode-licensedb/zip/" + tt.commit}, requested)
			if tt.commit == currentScancodeSource.commit {
				assert.Equal(t, original, pins)
				assert.Contains(t, warnings.String(), "already current")
			} else {
				digest := sha256.Sum256(tt.archive)
				assert.Contains(t, string(pins), fmt.Sprintf("%x", digest))
				assert.Contains(t, string(pins), newCommit)
			}
			output, err := os.ReadFile("generated.ts")
			require.NoError(t, err)
			assert.Contains(t, string(output), "// scancode-licensedb commit: "+tt.commit)
			assert.Contains(t, string(output), "'MIT': {name: 'MIT License', category: 'Permissive'}")
			assert.Contains(t, string(output), "'any-OSI': {name: 'any-OSI', category: 'Permissive'}")
			assert.Contains(t, warnings.String(), "missing scancode metadata")
		})
	}
}

func TestPrepareScancodeUpdate_Unchanged(t *testing.T) {
	t.Parallel()
	requests := 0
	client := &http.Client{Transport: scancodeTransport(func(req *http.Request) (*http.Response, error) {
		requests++
		assert.Equal(t, scancodeMainAPI, req.URL.String())
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(`{"sha":"` + currentScancodeSource.commit + `"}`)),
		}, nil
	})}
	source, records, pins, err := prepareScancodeUpdate(context.Background(), client, io.Discard)
	require.NoError(t, err)
	assert.Equal(t, currentScancodeSource, source)
	assert.Nil(t, records)
	assert.Nil(t, pins)
	assert.Equal(t, 1, requests)
}

func TestRunRejectsUpdatingDuringVerification(t *testing.T) {
	t.Parallel()
	err := run([]string{"-verify", "generated.ts", "-update-scancode"}, io.Discard, io.Discard, nil)
	require.ErrorContains(t, err, "-verify cannot be combined with -update-scancode")
}
