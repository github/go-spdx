package spdxexp

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSatisfies(t *testing.T) {
	tests := []struct {
		name      string
		firstExp  string
		secondExp string
		satisfied bool
		err       error
	}{
		// regression tests from spdx-satisfies.js - comments for satisfies function
		// TODO: Commented out tests are not yet supported.
		{"MIT satisfies MIT is true", "MIT", "MIT", true, nil},

		// {"MIT satisfies MIT OR Apache-2.0 is true", "MIT", "MIT OR Apache-2.0", true, nil},
		// {"MIT OR Apache-2.0 satisfies MIT is true", "MIT OR Apache-2.0", "MIT", true, nil},
		// {"GPL satisfies MIT OR Apache-2.0 is false", "GPL", "MIT OR Apache-2.0", false, nil},
		// {"MIT OR Apache-2.0 satisfies GPL is false", "MIT OR Apache-2.0", "GPL", false, nil},

		// {"Apache-2.0 AND MIT satisfies MIT AND Apache-2.0 is true", "Apache-2.0 AND MIT", "MIT AND Apache-2.0", true, nil},
		// {"MIT AND Apache-2.0 satisfies MIT AND Apache-2.0 is true", "MIT AND Apache-2.0", "MIT AND Apache-2.0", true, nil},
		// {"MIT satisfies MIT AND Apache-2.0 is false", "MIT", "MIT AND Apache-2.0", false, nil},
		// {"MIT AND Apache-2.0 satisfies MIT is false", "MIT AND Apache-2.0", "MIT", false, nil},
		// {"GPL satisfies MIT AND Apache-2.0 is false", "GPL", "MIT AND Apache-2.0", false, nil},

		// {"MIT AND Apache-2.0 satisfies MIT AND (Apache-1.0 OR Apache-2.0)", "MIT AND Apache-2.0", "MIT AND (Apache-1.0 OR Apache-2.0)", true, nil},

		// {"Apache-1.0+ satisfies Apache-2.0+ is true", "Apache-1.0+", "Apache-2.0+", true, nil}, // TODO: why does this fail here but passes js?
		{"Apache-1.0 satisfies Apache-2.0+ is false", "Apache-1.0", "Apache-2.0+", false, nil},
		{"Apache-2.0 satisfies Apache-2.0+ is true", "Apache-2.0", "Apache-2.0+", true, nil},
		// {"Apache-3.0 satisfies Apache-2.0+ is true", "Apache-3.0", "Apache-2.0+", true, nil}, // TODO: gets error b/c Apache-3.0 doesn't exist -- need better error message

		{"Apache-1.0 satisfies Apache-2.0-or-later is false", "Apache-1.0", "Apache-2.0-or-later", false, nil},
		{"Apache-2.0 satisfies Apache-2.0-or-later is true", "Apache-2.0", "Apache-2.0-or-later", true, nil},
		// {"Apache-3.0 satisfies Apache-2.0-or-later is true", "Apache-3.0", "Apache-2.0-or-later", true, nil},

		{"Apache-1.0 satisfies Apache-2.0-only is false", "Apache-1.0", "Apache-2.0-only", false, nil},
		{"Apache-2.0 satisfies Apache-2.0-only is true", "Apache-2.0", "Apache-2.0-only", true, nil},
		// {"Apache-3.0 satisfies Apache-2.0-only is false", "Apache-3.0", "Apache-2.0-only", false, nil},

		// regression tests from spdx-satisfies.js - assert statements in README
		// TODO: Commented out tests are not yet supported.
		{"MIT satisfies MIT", "MIT", "MIT", true, nil},

		// {"MIT satisfies (ISC OR MIT)", "MIT", "(ISC OR MIT)", true, nil},
		// {"Zlib satisfies (ISC OR (MIT OR Zlib))", "Zlib", "(ISC OR (MIT OR Zlib))", true, nil},
		// {"GPL-3.0 !satisfies (ISC OR MIT)", "GPL-3.0", "(ISC OR MIT)", false, nil},

		// {"GPL-2.0 satisfies GPL-2.0+", "GPL-2.0", "GPL-2.0+", true, nil}, // TODO: why does this fail here but passes js?
		// {"GPL-2.0 satisfies GPL-2.0-or-later", "GPL-2.0", "GPL-2.0-or-later", true, nil}, // TODO: why does this fail here but passes js?
		{"GPL-3.0 satisfies GPL-2.0+", "GPL-3.0", "GPL-2.0+", true, nil},
		{"GPL-1.0-or-later satisfies GPL-2.0-or-later", "GPL-1.0-or-later", "GPL-2.0-or-later", true, nil},
		{"GPL-1.0+ satisfies GPL-2.0+", "GPL-1.0+", "GPL-2.0+", true, nil},
		{"GPL-1.0 !satisfies GPL-2.0+", "GPL-1.0", "GPL-2.0+", false, nil},
		{"GPL-2.0-only satisfies GPL-2.0-only", "GPL-2.0-only", "GPL-2.0-only", true, nil},
		{"GPL-3.0-only satisfies GPL-2.0+", "GPL-3.0-only", "GPL-2.0+", true, nil},

		// {"GPL-2.0 !satisfies GPL-2.0+ WITH Bison-exception-2.2",
		// 	"GPL-2.0", "GPL-2.0+ WITH Bison-exception-2.2", false, nil},
		// {"GPL-3.0 WITH Bison-exception-2.2 satisfies GPL-2.0+ WITH Bison-exception-2.2",
		// 	"GPL-3.0 WITH Bison-exception-2.2", "GPL-2.0+ WITH Bison-exception-2.2", true, nil},

		// {"(MIT OR GPL-2.0) satisfies (ISC OR MIT)", "(MIT OR GPL-2.0)", "(ISC OR MIT)", true, nil},
		// {"(MIT AND GPL-2.0) satisfies (MIT AND GPL-2.0)", "(MIT AND GPL-2.0)", "(MIT AND GPL-2.0)", true, nil},
		// {"MIT AND GPL-2.0 AND ISC satisfies MIT AND GPL-2.0 AND ISC",
		// 	"MIT AND GPL-2.0 AND ISC", "MIT AND GPL-2.0 AND ISC", true, nil},
		// {"MIT AND GPL-2.0 AND ISC satisfies ISC AND GPL-2.0 AND MIT",
		// 	"MIT AND GPL-2.0 AND ISC", "ISC AND GPL-2.0 AND MIT", true, nil},
		// {"(MIT OR GPL-2.0) AND ISC satisfies MIT AND ISC",
		// 	"(MIT OR GPL-2.0) AND ISC", "MIT AND ISC", true, nil},
		// {"MIT AND ISC satisfies (MIT OR GPL-2.0) AND ISC",
		// 	"MIT AND ISC", "(MIT OR GPL-2.0) AND ISC", true, nil},
		// {"MIT AND ISC satisfies (MIT AND GPL-2.0) OR ISC",
		// 	"MIT AND ISC", "(MIT AND GPL-2.0) OR ISC", true, nil},
		// {"(MIT OR Apache-2.0) AND (ISC OR GPL-2.0) satisfies Apache-2.0 AND ISC",
		// 	"(MIT OR Apache-2.0) AND (ISC OR GPL-2.0)", "Apache-2.0 AND ISC", true, nil},
		// {"(MIT OR Apache-2.0) AND (ISC OR GPL-2.0) satisfies Apache-2.0 OR ISC",
		// 	"(MIT OR Apache-2.0) AND (ISC OR GPL-2.0)", "Apache-2.0 OR ISC", true, nil},
		// {"(MIT AND GPL-2.0) satisfies (MIT OR GPL-2.0)",
		// 	"(MIT AND GPL-2.0)", "(MIT OR GPL-2.0)", true, nil},
		// {"(MIT AND GPL-2.0) satisfies (GPL-2.0 AND MIT)",
		// 	"(MIT AND GPL-2.0)", "(GPL-2.0 AND MIT)", true, nil},
		// {"MIT satisfies (GPL-2.0 OR MIT) AND (MIT OR ISC)",
		// 	"MIT", "(GPL-2.0 OR MIT) AND (MIT OR ISC)", true, nil},
		// {"MIT AND ICU satisfies (MIT AND GPL-2.0) OR (ISC AND (Apache-2.0 OR ICU))",
		// 	"MIT AND ICU", "(MIT AND GPL-2.0) OR (ISC AND (Apache-2.0 OR ICU))", true, nil},
		// {"(MIT AND GPL-2.0) !satisfies (ISC OR GPL-2.0)",
		// 	"(MIT AND GPL-2.0)", "(ISC OR GPL-2.0)", false, nil},
		// {"MIT AND (GPL-2.0 OR ISC) !satisfies MIT",
		// 	"MIT AND (GPL-2.0 OR ISC)", "MIT", false, nil},
		// {"(MIT OR Apache-2.0) AND (ISC OR GPL-2.0) !satisfies MIT",
		// 	"(MIT OR Apache-2.0) AND (ISC OR GPL-2.0)", "MIT", false, nil},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			satisfied, err := satisfies(test.firstExp, test.secondExp)
			assert.Equal(t, test.err, err)
			assert.Equal(t, test.satisfied, satisfied)
		})
	}
}
