package main

import (
	"flag"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var update = flag.Bool("update", false, "update golden files")

// TestRenderTypeScript_Golden compares renderTypeScript output against a checked-in fixture.
// Run with -update to regenerate the golden file after intentional output changes:
//
//	go test ./cmd/license-categories/... -run TestRenderTypeScript_Golden -update
func TestRenderTypeScript_Golden(t *testing.T) {
	t.Parallel()

	input := map[string]licenseMetadata{
		"Apache-2.0":   {Name: "Apache License 2.0", Category: "Permissive"},
		"MIT":          {Name: "MIT License", Category: "Permissive"},
		"GPL-2.0-only": {Name: "GNU General Public License v2.0 only", Category: "Copyleft"},
		"0BSD":         {Name: "BSD Zero Clause License", Category: "Permissive"},
		"CC-BY-4.0":    {Name: "Creative Commons Attribution 4.0 International", Category: "Free restricted"},
	}

	got, err := renderTypeScript(input, generationProvenance{
		GoSPDXVersion:          "v2.7.1",
		SPDXLicenseListVersion: "3.29.0",
		ScancodeCommit:         "abc123",
	})
	require.NoError(t, err)

	const golden = "testdata/render_typescript.golden"

	if *update {
		err = os.WriteFile(golden, got, 0o600)
		require.NoError(t, err)
		t.Log("updated golden file")
		return
	}

	expected, err := os.ReadFile(golden)
	require.NoError(t, err, "golden file missing; run with -update to create")

	assert.Equal(t, string(expected), string(got))
}
