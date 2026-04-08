package spdxexp

import (
	"context"
	"os"
	"os/exec"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

const kernelHeadersLicense = `(GPL-2.0-only WITH Linux-syscall-note OR BSD-2-Clause) AND (GPL-2.0-only WITH Linux-syscall-note OR BSD-3-Clause) AND (GPL-2.0-only WITH Linux-syscall-note OR CDDL-1.0) AND (GPL-2.0-only WITH Linux-syscall-note OR Linux-OpenIB) AND (GPL-2.0-only WITH Linux-syscall-note OR MIT) AND (GPL-2.0-or-later WITH Linux-syscall-note OR BSD-3-Clause) AND (GPL-2.0-or-later WITH Linux-syscall-note OR MIT) AND Apache-2.0 AND BSD-2-Clause AND BSD-3-Clause AND BSD-3-Clause-Clear AND GFDL-1.1-no-invariants-or-later AND GPL-1.0-or-later AND (GPL-1.0-or-later OR BSD-3-Clause) AND GPL-1.0-or-later WITH Linux-syscall-note AND GPL-2.0-only AND (GPL-2.0-only OR Apache-2.0) AND (GPL-2.0-only OR BSD-2-Clause) AND (GPL-2.0-only OR BSD-3-Clause) AND (GPL-2.0-only OR CDDL-1.0) AND (GPL-2.0-only OR GFDL-1.1-no-invariants-or-later) AND (GPL-2.0-only OR GFDL-1.2-no-invariants-only) AND GPL-2.0-only WITH Linux-syscall-note AND GPL-2.0-or-later AND (GPL-2.0-or-later OR BSD-2-Clause) AND (GPL-2.0-or-later OR BSD-3-Clause) AND (GPL-2.0-or-later OR CC-BY-4.0) AND GPL-2.0-or-later WITH GCC-exception-2.0 AND GPL-2.0-or-later WITH Linux-syscall-note AND ISC AND LGPL-2.0-or-later AND (LGPL-2.0-or-later OR BSD-2-Clause) AND LGPL-2.0-or-later WITH Linux-syscall-note AND LGPL-2.1-only AND (LGPL-2.1-only OR BSD-2-Clause) AND LGPL-2.1-only WITH Linux-syscall-note AND LGPL-2.1-or-later AND LGPL-2.1-or-later WITH Linux-syscall-note AND (Linux-OpenIB OR GPL-2.0-only) AND (Linux-OpenIB OR GPL-2.0-only OR BSD-2-Clause) AND Linux-man-pages-copyleft AND MIT AND (MIT OR GPL-2.0-only) AND (MIT OR GPL-2.0-or-later) AND (MIT OR LGPL-2.1-only) AND (MPL-1.1 OR GPL-2.0-only) AND (X11 OR GPL-2.0-only) AND (X11 OR GPL-2.0-or-later) AND Zlib AND (copyleft-next-0.3.1 OR GPL-2.0-or-later)`

var expectedKernelHeadersLicenses = []string{
	"GPL-2.0-only WITH Linux-syscall-note",
	"BSD-2-Clause",
	"BSD-3-Clause",
	"CDDL-1.0",
	"Linux-OpenIB",
	"MIT",
	"GPL-2.0-or-later WITH Linux-syscall-note",
	"Apache-2.0",
	"BSD-3-Clause-Clear",
	"GFDL-1.1-no-invariants-or-later",
	"GPL-1.0-or-later",
	"GPL-1.0-or-later WITH Linux-syscall-note",
	"GPL-2.0-only",
	"GFDL-1.2-no-invariants-only",
	"GPL-2.0-or-later",
	"CC-BY-4.0",
	"GPL-2.0-or-later WITH GCC-exception-2.0",
	"ISC",
	"LGPL-2.0-or-later",
	"LGPL-2.0-or-later WITH Linux-syscall-note",
	"LGPL-2.1-only",
	"LGPL-2.1-only WITH Linux-syscall-note",
	"LGPL-2.1-or-later",
	"LGPL-2.1-or-later WITH Linux-syscall-note",
	"Linux-man-pages-copyleft",
	"MPL-1.1",
	"X11",
	"Zlib",
	"copyleft-next-0.3.1",
}

func TestExtractLicenses(t *testing.T) {
	tests := []struct {
		name              string
		inputExpression   string
		extractedLicenses []string
	}{
		{"Single license", "MIT", []string{"MIT"}},
		{"AND'ed licenses", "MIT AND Apache-2.0", []string{"MIT", "Apache-2.0"}},
		{"AND'ed & OR'ed licenses", "(MIT AND Apache-2.0) OR GPL-3.0", []string{"GPL-3.0", "MIT", "Apache-2.0"}},
		{"ONLY modifiers", "LGPL-2.1-only OR MIT OR BSD-3-Clause", []string{"MIT", "BSD-3-Clause", "LGPL-2.1-only"}},
		{"WITH modifiers", "GPL-2.0-or-later WITH Bison-exception-2.2", []string{"GPL-2.0-or-later WITH Bison-exception-2.2"}},
		{"Invalid SPDX expression", "MIT OR INVALID", nil},
		{"-or-later suffix with mixed case", "GPL-2.0-Or-later", []string{"GPL-2.0-or-later"}},
		{"+ operator in Apache", "APACHE-2.0+", []string{"Apache-2.0+"}},
		{"+ operator in GPL", "GPL-2.0+", []string{"GPL-2.0-or-later"}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			licenses, err := ExtractLicenses(test.inputExpression)
			assert.ElementsMatch(t, test.extractedLicenses, licenses)
			if test.extractedLicenses == nil {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestExtractLicensesLicenseRefAndDedup(t *testing.T) {
	licenses, err := ExtractLicenses("(LicenseRef-custom OR LicenseRef-custom) AND (DocumentRef-spdx-tool-1.2:LicenseRef-custom OR MIT)")
	assert.NoError(t, err)
	assert.ElementsMatch(t, []string{"LicenseRef-custom", "DocumentRef-spdx-tool-1.2:LicenseRef-custom", "MIT"}, licenses)
}

func TestExtractLicensesLongExpressionDoesNotHang(t *testing.T) {
	if os.Getenv("GO_SPDX_EXTRACT_LICENSES_LONG_CHILD") == "1" {
		licenses, err := ExtractLicenses(kernelHeadersLicense)
		assert.NoError(t, err)
		assert.ElementsMatch(t, expectedKernelHeadersLicenses, licenses)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, os.Args[0], "-test.run", "^TestExtractLicensesLongExpressionDoesNotHang$")
	cmd.Env = append(os.Environ(), "GO_SPDX_EXTRACT_LICENSES_LONG_CHILD=1")
	output, err := cmd.CombinedOutput()
	if ctx.Err() == context.DeadlineExceeded {
		t.Fatalf("ExtractLicenses timed out on long expression: %s", output)
	}
	if err != nil {
		t.Fatalf("child process failed: %v\n%s", err, output)
	}
}
