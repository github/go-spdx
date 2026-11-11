# Category overrides and license exclusions

ScanCode can miscategorize a license or lack useful category metadata. The
license categories generator supports explicit overrides to correct an
assignment without waiting for an upstream data update. It also supports
excluding individual licenses from the generated npm data.

Both mechanisms are defined in
[`cmd/license-categories/main.go`](../../../cmd/license-categories/main.go).
For general build, generation, and release instructions, see
[Operations](./operations.md).

## Choose the exact SPDX ID

The generator at [cmd/license-categories/main.go](../../../cmd/license-categories/main.go) iterates the active IDs returned by
`spdxlicenses.GetLicenses()` in
[`spdxexp/spdxlicenses/get_licenses.go`](../../../spdxexp/spdxlicenses/get_licenses.go).
Use the exact, case-sensitive SPDX ID as the map key, not the ScanCode key,
display name, or an SPDX expression.

Overrides and exclusions only affect IDs in that active list. Unmatched keys
are not applied and are not reported as errors, so verify spelling and generated
output. These maps do not add new IDs to go-spdx or change Go license validation.

## Override a category

Edit `spdxCategoryOverrides` in the generator.
Each entry maps an SPDX ID to the desired category. For example, this override assigns `ESA-PL-strong-copyleft-2.4`
to `Copyleft`:

```go
"ESA-PL-strong-copyleft-2.4": "Copyleft", // incorrectly categorized as Permissive https://github.com/aboutcode-org/scancode-licensedb/issues/85
```

> [!IMPORTANT]
> Review scancode periodically to determine if a correction has been made.
> Report incorrect categorizations to https://github.com/aboutcode-org/scancode-licensedb/issues

Keep the list explicit and reviewed. Record why the assignment needs correction
in a concise comment or the change description, with a relevant upstream record
or license reference where available. Review the license's actual terms rather
than inferring a category from its SPDX ID. This guide does not establish a new
classification or approval policy.

Override values must exactly match a key in `categoryDescriptions` defined in the generator:

- `Commercial`
- `Copyleft`
- `Non-commercial`
- `Copyleft limited`
- `Free restricted`
- `Proprietary free`
- `Permissive`
- `Public domain`
- `Source available`
- `Unknown`

Overrides are validated directly, **without** ScanCode category normalization.
For example, use `Copyleft limited`, not `Copyleft Limited`. An invalid category
for an active, non-excluded ID stops generation with a
`has no description mapping` error.

An override changes the category while preserving the ScanCode name and source
URL when present. If the record is missing or its name is empty, the display
name falls back to the SPDX ID. An override works even when ScanCode lacks that
ID; it does not emit the missing-metadata warning for the overridden ID.
Source URLs are retained internally but are not exported in `licensesById`.

## Exclude a license

Edit `excludedSpdxIDs` in the same generator. It is currently an empty
`map[string]struct{}`. To omit an ID, add it as a set entry:

```go
var excludedSpdxIDs = map[string]struct{}{
    "EXACT-SPDX-ID": {}, // Replace with an active SPDX ID and explain the exclusion.
}
```

The placeholder above is illustrative, not a proposed exclusion. Document the
reason for any actual exclusion in the change. Exclusion removes the license
entirely from the npm data rather than assigning `Unknown`, and emits
`excluded SPDX license "..."` on stderr. It does not remove the license from
go-spdx's active SPDX list.

Consequently, an excluded ID is absent from `licensesById` and
`categoriesWithLicenses()`, and `categoryForLicense(id)` returns `undefined`.
The generated `categoryDescriptions` contains only categories still represented
by at least one included license.

## Precedence and fallback behavior

For each active SPDX ID, `buildLicenseData` applies these rules in order:

1. If excluded, omit it and emit an exclusion warning. Exclusion wins over an
   override for the same ID; the override is not validated or applied.
2. Otherwise, if overridden, validate and apply the explicit category as
   described above.
3. Otherwise, use ScanCode metadata. If the record is missing, use the SPDX ID
   as the name and `Unknown` as the category, with a missing-metadata warning.
   If a record exists but its name is empty, use the SPDX ID as the name.

For non-overridden ScanCode records, `mapScancodeCategory` normalizes
`Source-available`, `Free Restricted`, `Copyleft Limited`, `Public Domain`,
`Non-Commercial`, and `Proprietary Free` to the corresponding spellings in the
list above. An empty category becomes `Unknown` without a missing-record
warning. Other category strings are unchanged and must match
`categoryDescriptions`, or generation fails. Warnings are sorted before output.

## Regenerate and review

Run from the repository root after changing either map:

```sh
make license-categories
make verify-license-categories
go test ./cmd/license-categories/...
make test-license-categories
```

Generation downloads the pinned ScanCode archive and verifies its SHA-256.
Install npm dependencies with
`npm ci --prefix packages/license-categories` first if needed.
Review the warnings and the diff in
`packages/license-categories/src/generated.ts`. Include regenerated data with
the map change; do not edit the generated assignments manually.

The existing tests in
[`main_test.go`](../../../cmd/license-categories/main_test.go) cover exclusions
and missing-metadata fallback, overrides with and without ScanCode records, and
invalid override categories. Add or adjust focused cases for the intended
correction or exclusion. The
rendering golden test and npm API/tarball tests provide additional regression
coverage; do not update a golden fixture merely to hide an unexpected change.

When updating the upstream ScanCode pin later, review whether each override is
still necessary. Removing an override restores ScanCode or fallback behavior;
removing an exclusion restores the license using the same precedence rules.
