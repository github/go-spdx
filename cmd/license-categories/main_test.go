package main

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"runtime/debug"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestModuleVersion(t *testing.T) {
	t.Parallel()

	info, ok := debug.ReadBuildInfo()
	require.True(t, ok)
	got, err := moduleVersion()
	require.NoError(t, err)
	if info.Main.Version != "" && info.Main.Version != "(devel)" {
		assert.Equal(t, info.Main.Version, got)
	} else {
		assert.Regexp(t, `^[a-f0-9]{40} \(local\)$`, got)
		revision, err := exec.Command("git", "rev-parse", "HEAD").Output()
		require.NoError(t, err)
		assert.Equal(t, strings.TrimSpace(string(revision))+" (local)", got)
	}
}

func TestLocalCheckoutVersionAndVerificationAcrossCommits(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	source := filepath.Join(root, "cmd", "license-categories", "main.go")
	require.NoError(t, os.MkdirAll(filepath.Dir(source), 0o750))
	require.NoError(t, os.WriteFile(source, []byte("package main\n"), 0o600))
	git := func(args ...string) string {
		// #nosec G204 -- arguments are fixed test commands scoped to t.TempDir.
		command := exec.Command("git", append([]string{"-C", root}, args...)...)
		output, err := command.CombinedOutput()
		require.NoError(t, err, string(output))
		return strings.TrimSpace(string(output))
	}
	git("init", "--quiet")
	git("add", ".")
	commit := func() {
		git("-c", "user.name=Test", "-c", "user.email=test@example.com",
			"-c", "commit.gpgsign=false", "commit", "--quiet", "--allow-empty", "-m", "Test provenance")
	}
	commit()
	first, err := localCheckoutVersion(source)
	require.NoError(t, err)
	assert.Equal(t, git("rev-parse", "HEAD")+" (local)", first)
	commit()
	second, err := localCheckoutVersion(source)
	require.NoError(t, err)
	assert.NotEqual(t, first, second)
	assert.Regexp(t, `^[a-f0-9]{40} \(local\)$`, second)

	render := func(version string) []byte {
		output, err := renderTypeScript(map[string]licenseMetadata{"MIT": {Name: "MIT License", Category: "Permissive"}},
			generationProvenance{GoSPDXVersion: version, SPDXLicenseListVersion: "3.29.0", ScancodeCommit: "abc123"})
		require.NoError(t, err)
		return output
	}
	require.NoError(t, verifyGeneratedData(render(first), render(second)))
	require.NoError(t, verifyGeneratedData(render("(devel)"), render(second)))
}

func TestLocalCheckoutVersionErrors(t *testing.T) {
	t.Parallel()

	_, err := localCheckoutVersion("cmd/license-categories/main.go")
	require.ErrorContains(t, err, "absolute generator source path")
	_, err = localCheckoutVersion(filepath.Join(t.TempDir(), "main.go"))
	require.ErrorContains(t, err, "Git checkout")
}

func TestVerifyGeneratedDataIgnoresOnlyGoSPDXProvenance(t *testing.T) {
	t.Parallel()

	render := func(version string) []byte {
		output, err := renderTypeScript(map[string]licenseMetadata{"MIT": {Name: "MIT License", Category: "Permissive"}},
			generationProvenance{GoSPDXVersion: version, SPDXLicenseListVersion: "3.29.0", ScancodeCommit: "abc123"})
		require.NoError(t, err)
		return output
	}
	output := render(strings.Repeat("a", 40) + " (local)")
	for _, version := range []string{"(devel)", "v2.9.0 (release)", strings.Repeat("b", 40) + " (dispatch)"} {
		actual := render(version)
		require.NoError(t, verifyGeneratedData(output, actual))
		require.NoError(t, verifyGeneratedData(actual, output))
	}
}

func TestVerifyGeneratedDataRejectsOtherDifferences(t *testing.T) {
	t.Parallel()

	provenance := generationProvenance{GoSPDXVersion: strings.Repeat("a", 40) + " (local)",
		SPDXLicenseListVersion: "3.29.0", ScancodeCommit: "abc123"}
	output, err := renderTypeScript(map[string]licenseMetadata{"MIT": {Name: "MIT License", Category: "Permissive"}}, provenance)
	require.NoError(t, err)
	for _, change := range [][2]string{
		{"MIT License", "changed-name"},
		{"'MIT':", "'changed-ID':"},
		{"Permissive", "Copyleft"},
		{"Safe for most", "changed-description"},
		{"// Code generated", "// Different owner"},
		{"// go-spdx: ", "// wrong-header: "},
		{"// spdx-license-list-version: ", "// wrong-header: "},
		{"// scancode-licensedb commit: ", "// wrong-header: "},
		{provenance.GoSPDXVersion, ""},
		{"3.29.0", ""},
		{"3.29.0", "3.30.0"},
		{"abc123", "def456"},
		{"abc123", "   "},
		{"// go-spdx: ", "// Extra comment\n// go-spdx: "},
		{"// spdx-license-list-version: ", "// Extra comment\n// spdx-license-list-version: "},
		{"// scancode-licensedb commit: ", "// Extra comment\n// scancode-licensedb commit: "},
		{"// scancode-licensedb commit: abc123\n", ""},
		{"export const", "// Extra comment\nexport const"},
		{"export const", "// go-spdx: not leading provenance\nexport const"},
	} {
		changed := []byte(strings.Replace(string(output), change[0], change[1], 1))
		require.NotEqual(t, output, changed, change)
		require.Error(t, verifyGeneratedData(output, changed), change)
		require.Error(t, verifyGeneratedData(changed, output), change)
	}
	require.Error(t, verifyGeneratedData([]byte("no headers"), output))
}

func TestRunRejectsConflictingOutputAndVerification(t *testing.T) {
	t.Parallel()

	err := run([]string{"-verify", "generated.ts", "-out", "other.ts"}, io.Discard, io.Discard, nil)
	require.ErrorContains(t, err, "-verify cannot be combined with -out")
}

func TestCategoryOverrides(t *testing.T) {
	t.Parallel()

	expected := map[string]string{
		"any-OSI":                    "Permissive",
		"any-OSI-perl-modules":       "Permissive",
		"ESA-PL-strong-copyleft-2.4": "Copyleft",
		"ESA-PL-weak-copyleft-2.4":   "Copyleft limited",
		"MakeIndex":                  "Copyleft",
	}
	assert.Equal(t, expected, spdxCategoryOverrides)
}

func TestBuildLicenseData_ExclusionsAndUnknown(t *testing.T) {
	t.Parallel()

	spdxIDs := []string{"MIT", "GPL-2.0-only", "Apache-2.0"}
	scancode := map[string]licenseMetadata{
		"MIT": {Name: "MIT License", Category: "Permissive", SourceURL: "https://github.com/aboutcode-org/scancode-licensedb/blob/main/docs/mit.json"},
	}
	excluded := map[string]struct{}{
		"GPL-2.0-only": {},
	}

	overrides := map[string]string{}

	got, warnings, err := buildLicenseData(spdxIDs, scancode, excluded, overrides)
	require.NoError(t, err)

	require.Len(t, got, 2)
	assert.Equal(t, licenseMetadata{Name: "MIT License", Category: "Permissive", SourceURL: "https://github.com/aboutcode-org/scancode-licensedb/blob/main/docs/mit.json"}, got["MIT"])
	assert.Equal(t, licenseMetadata{Name: "Apache-2.0", Category: "Unknown"}, got["Apache-2.0"])
	_, hasExcluded := got["GPL-2.0-only"]
	assert.False(t, hasExcluded)

	assert.Len(t, warnings, 2)
	assert.Contains(t, warnings[0], "excluded SPDX license")
	assert.Contains(t, strings.Join(warnings, "\n"), "missing scancode metadata for \"Apache-2.0\"")
}

func TestBuildLicenseData_AppliesSpdxCategoryOverrides(t *testing.T) {
	t.Parallel()

	const licenseNoScancode = "License-No-Scancode-Record"

	spdxIDs := []string{"MIT", licenseNoScancode}
	scancode := map[string]licenseMetadata{
		"MIT": {Name: "MIT License", Category: "Permissive", SourceURL: "https://github.com/aboutcode-org/scancode-licensedb/blob/main/docs/mit.json"},
	}
	excluded := map[string]struct{}{}
	overrides := map[string]string{
		"MIT":             "Copyleft",
		licenseNoScancode: "Permissive",
	}

	got, warnings, err := buildLicenseData(spdxIDs, scancode, excluded, overrides)
	require.NoError(t, err)
	assert.Empty(t, warnings)

	// Override with scancode data: preserves scancode Name and SourceURL
	assert.Equal(t, "Copyleft", got["MIT"].Category)
	assert.Equal(t, "MIT License", got["MIT"].Name)
	assert.Equal(t, "https://github.com/aboutcode-org/scancode-licensedb/blob/main/docs/mit.json", got["MIT"].SourceURL)

	// Override without scancode data: falls back to spdxID as Name, empty SourceURL
	assert.Equal(t, "Permissive", got[licenseNoScancode].Category)
	assert.Equal(t, licenseNoScancode, got[licenseNoScancode].Name)
	assert.Equal(t, "", got[licenseNoScancode].SourceURL)
}

func TestBuildLicenseData_InvalidOverrideCategoryReturnsError(t *testing.T) {
	t.Parallel()

	spdxIDs := []string{"MIT"}
	scancode := map[string]licenseMetadata{
		"MIT": {Name: "MIT License", Category: "Permissive", SourceURL: "https://github.com/aboutcode-org/scancode-licensedb/blob/main/docs/mit.json"},
	}
	excluded := map[string]struct{}{}
	overrides := map[string]string{"MIT": "Not a real category"}

	_, _, err := buildLicenseData(spdxIDs, scancode, excluded, overrides)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "has no description mapping")
}

func TestRenderTypeScript_Deterministic(t *testing.T) {
	t.Parallel()

	input := map[string]licenseMetadata{
		"Apache-2.0": {Name: "Apache License 2.0", Category: "Permissive"},
		"MIT":        {Name: "MIT License", Category: "Permissive"},
		"Zlib":       {Name: "zlib License", Category: "Permissive"},
	}

	provenance := generationProvenance{
		GoSPDXVersion:          "v2.7.1",
		SPDXLicenseListVersion: "3.29.0",
		ScancodeCommit:         "abc123",
	}
	out1, err := renderTypeScript(input, provenance)
	require.NoError(t, err)

	out2, err := renderTypeScript(input, provenance)
	require.NoError(t, err)

	assert.Equal(t, string(out1), string(out2))
	assert.Contains(t, string(out1), "export const categoryDescriptions")
	assert.Contains(t, string(out1), "export const licensesById")
	assert.Contains(t, string(out1), "// go-spdx: v2.7.1")
	assert.Contains(t, string(out1), "// spdx-license-list-version: 3.29.0")
	assert.Contains(t, string(out1), "// scancode-licensedb commit: abc123")
	assert.NotContains(t, string(out1), "export function")
}

func TestVerifySHA256(t *testing.T) {
	t.Parallel()

	err := verifySHA256([]byte("abc"), "ba7816bf8f01cfea414140de5dae2223b00361a396177a9cb410ff61f20015ad")
	require.NoError(t, err)

	err = verifySHA256([]byte("abc"), strings.Repeat("0", 64))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "checksum mismatch")
}

func TestScancodeGitHubURL(t *testing.T) {
	t.Parallel()
	const commit = "test-commit"

	assert.Equal(t,
		"https://github.com/aboutcode-org/scancode-licensedb/blob/"+commit+"/docs/mit.json",
		scancodeGitHubURL("scancode-licensedb-main/docs/mit.json", commit),
	)
	assert.Equal(t,
		"https://github.com/aboutcode-org/scancode-licensedb/blob/"+commit+"/docs/mit.json",
		scancodeGitHubURL("scancode-licensedb-v3.0/docs/mit.json", commit),
	)
	assert.Equal(t,
		"https://github.com/aboutcode-org/scancode-licensedb/blob/"+commit+"/path.json",
		scancodeGitHubURL("unexpected/path.json", commit),
	)
	assert.Equal(t, "noprefix.json", scancodeGitHubURL("noprefix.json", commit))
}

func TestIsScancodeLicenseJSON(t *testing.T) {
	t.Parallel()

	assert.True(t, isScancodeLicenseJSON("scancode-licensedb-main/docs/mit.json"))
	assert.False(t, isScancodeLicenseJSON("scancode-licensedb-main/docs/index.json"))
	assert.False(t, isScancodeLicenseJSON("scancode-licensedb-main/something/mit.json"))
	assert.False(t, isScancodeLicenseJSON("scancode-licensedb-main/docs/mit.yml"))
}

func TestLoadScancodeBySpdx_RejectsOversizedResponse(t *testing.T) {
	t.Parallel()

	const testMaxSize int64 = 1024
	body := make([]byte, testMaxSize+1)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(body)
	}))
	defer srv.Close()

	_, err := loadScancodeBySpdx(context.Background(), srv.Client(), scancodeSource{
		commit:    "test-commit",
		zipURL:    srv.URL,
		zipSHA256: strings.Repeat("0", 64),
	}, testMaxSize)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "exceeds")
}
