# Contributing to license categories

This guide covers generation and operation of `@github/license-categories`.
See [Contributing](../../../CONTRIBUTING.md) for repository-wide contribution
guidance and the [package README](../README.md) for installation, API usage,
compatibility, and licensing. See [Overrides and exclusions](./overrides-and-exclusions.md)
for correcting category assignments or omitting licenses from the npm data.

## Generate, build, and test

Use the Go version declared in [go.mod](../../../go.mod) and Node.js 20+ with npm.
The workflows use Node.js 24. Run from the repository root:

```sh
npm ci --prefix packages/license-categories
make license-categories
make verify-license-categories
go test ./cmd/license-categories/...
make test-license-categories
```

`make license-categories` runs `go run ./cmd/license-categories` and writes
`packages/license-categories/src/generated.ts`. Generation requires network
access to the pinned public ScanCode archive. `make verify-license-categories`
generates a temporary file and compares it byte for byte with the checked-in
file, including all provenance comments, without overwriting it.

For a build without tests, run:

```sh
npm run build --prefix packages/license-categories
```

The build removes `dist` before compiling TypeScript. `npm test` builds and runs
the API, tarball, and validation-workflow tests, including importing an extracted
package independently of the source tree.
`npm run test:tarball --prefix packages/license-categories` runs only the packaging
tests after building. The `prepack` hook also builds before packing or publishing.

## Maintain the data

SPDX IDs and the SPDX License List version come from this repository's
`spdxexp/spdxlicenses` package. Names and categories come from
[ScanCode LicenseDB](https://github.com/aboutcode-org/scancode-licensedb), pinned
by commit and verified by SHA-256 in
[the generator](../../../cmd/license-categories/main.go).
The generator keeps explicit, reviewed category overrides, normalizes category
names, and warns while assigning `Unknown` to IDs absent from ScanCode.
Unrecognized category mappings are errors; do not infer categories from SPDX IDs.

`src/generated.ts` records the Go main module version, SPDX list version, and
ScanCode commit. Running the generator from this checkout with `go run` records
`go-spdx: (devel)`, not the checkout's SHA or a release tag. No timestamps are
generated, and output is deterministic for the same inputs.

Review data changes and any warnings, and include regenerated output with its
source changes. Do not edit generated data by hand. When updating the ScanCode
pin, update both its commit and archive SHA-256, regenerate the data, and review
overrides, warnings, and attribution in the package's
[THIRD_PARTY_NOTICE.md](../THIRD_PARTY_NOTICE.md). Preserve upstream data
licensing and any other third-party notices.

[Validate License Categories](../../../.github/workflows/validate-license-categories.yaml)
verifies committed generated data and runs npm tests on relevant pull requests
and pushes to `main`. It has read-only permissions and cannot publish. The existing
[Build and test workflow](../../../.github/workflows/test.yaml) runs the Go generator
tests through `go test -race ./...`; validation does not duplicate them.

## Publication workflow

[Publish License Categories](../../../.github/workflows/publish-license-categories.yaml)
runs on pushes to `main` that change `packages/license-categories/**`, or through
`workflow_dispatch`. The workflow rejects refs other than the repository's
default branch. Changes only to the generator or Makefile do not trigger it;
regenerate and commit package output with relevant source changes, or dispatch
the workflow on the default branch.

The single publication job is disabled unless the repository variable
`LICENSE_CATEGORIES_NPM_PUBLISH_ENABLED` is exactly `true`. Publications are
serialized without canceling a running publication. The job verifies generated
data, installs npm dependencies with a lockfile-backed cache, and runs the API
and tarball tests before publishing with public access through npm trusted
publishing. It uses Node.js 24 and npm 11.20.0 for OIDC authentication, with no
long-lived npm token.

The workflow retains the original OLC publication behavior: npm provenance is
explicitly disabled, no publication header is stamped, and every enabled run
attempts to publish the version in the manifest. There is no registry comparison,
unchanged-data skip, version bump, or release-event trigger. npm rejects a version
that has already been published. Bump both `package.json` and `package-lock.json`
before subsequent publication attempts.

## Publishing permissions and versioning

Adding this package does not publish it or change registry settings. Do not
enable publication until permissions, version, public access, and trusted
publishing have been approved.

Before the first publication, a maintainer must:

1. Confirm npm organization permissions and package-name availability to publish
   public `@github/license-categories`.
2. Confirm that the retained initial version `0.1.2` is approved and unused under
   the new package name. Update both version files if necessary. npm versions
   cannot be republished.
3. Configure the new package's npm trusted publisher for GitHub Actions: owner
   `github`, repository `go-spdx`, workflow filename
   `publish-license-categories.yaml`, with no environment restriction (the
   workflow does not use an environment). If first-publication setup requires
   a separate bootstrap, arrange that with an authorized npm maintainer.
4. After publication approval, set the repository variable to `true` and push
   an approved package change to `main` or dispatch the workflow on the default
   branch. A workflow filename or environment change requires updating the
   npm trusted publisher.
5. Verify anonymous installation from npmjs, the MIT software license, and
   upstream data attribution.
