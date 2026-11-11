# Contributing to license categories

This guide covers generation and operation of `@github/license-categories`.
See [Contributing](../../../CONTRIBUTING.md) for repository-wide contribution guidance
and the [package README](../README.md) for consumer
installation, API usage, compatibility, and licensing.
See [Overrides and exclusions](./overrides-and-exclusions.md) for correcting
individual category assignments or omitting licenses from the npm data.

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
generates a temporary file and compares it to the checked-in output without
overwriting it.

For a build without tests, run:

```sh
npm run build --prefix packages/license-categories
```

The build removes `dist` before compiling TypeScript. `npm test` builds and runs
the API and tarball tests, including importing an extracted package independently
of the source tree. `npm run test:tarball --prefix packages/license-categories`
runs only the packaging tests after building. The `prepack` hook also builds
before packing or publishing.

## Maintain the data

SPDX IDs and the SPDX License List version come from this repository's
`spdxexp/spdxlicenses` package. Names and categories come from
[ScanCode LicenseDB](https://github.com/aboutcode-org/scancode-licensedb), pinned
by commit and verified by SHA-256 in
[the generator](../../../cmd/license-categories/main.go).
The generator keeps explicit, reviewed category overrides, normalizes category
names, and warns while assigning `Unknown` to IDs absent from ScanCode.
Unrecognized category mappings are errors; do not infer categories from SPDX IDs.

`src/generated.ts` records the SPDX list version and ScanCode commit. The
go-spdx version is `(devel)` when using the local repository, or the main module
version when running a versioned installation. Local generation intentionally
does not embed a changing Git revision or timestamp, so verification is
deterministic. Review data changes and any warnings, and include regenerated
output with its source changes. Do not edit generated data by hand.

When updating the ScanCode pin, update both its commit and archive SHA-256,
regenerate the data, and review overrides, warnings, and attribution in the
package's [THIRD_PARTY_NOTICE.md](../THIRD_PARTY_NOTICE.md). Preserve upstream data
licensing and any other third-party notices.

The scheduled [Fetch Licenses workflow](../../../.github/workflows/fetch-licenses.yaml)
regenerates npm data alongside SPDX Go data in its update pull requests.

## Validation and publication workflows

[Validate License Categories](../../../.github/workflows/validate-license-categories.yaml)
verifies committed generated data and runs npm API, tarball, and publication
decision tests on relevant pull requests and pushes to `main`. It has read-only
permissions and cannot publish. The existing
[Build and test workflow](../../../.github/workflows/test.yaml) runs the Go
generator tests through `go test -race ./...`; validation does not duplicate them.

[Publish License Categories](../../../.github/workflows/publish-license-categories.yaml)
is publication-only. It runs when a GitHub release is published, checking out
that release's tag, or through `workflow_dispatch`, using the ref selected for
the dispatch. Draft releases do not trigger publication. Releases marked
**pre-release** in GitHub skip the publication job. This checks GitHub's
`release.prerelease` flag, not the tag's spelling or suffix.

The single publication job is disabled unless the repository variable
`LICENSE_CATEGORIES_NPM_PUBLISH_ENABLED` is exactly `true`. Publications are
serialized across all refs, without canceling a running publication. Dispatch
bypasses the Go-release event requirement, not the enablement gate, payload
comparison, or version checks. Only dispatch trusted, reviewed refs.
Npm semantic prerelease versions still follow the existing version-ordering
rules; the GitHub release flag does not change those rules. Dispatch does not
bypass the requirement to be newer than every published npm version. Supporting
backports on older release tracks would require a future version-policy change.

Publication independently runs `make verify-license-categories` using the Go
version from `go.mod`. This downloads the pinned ScanCode archive and generates
a temporary comparison file; a mismatch fails without replacing committed data.
It then installs npm dependencies using the lockfile-backed npm cache and runs
`npm test`, which cleanly builds the **committed** TypeScript source and runs API,
tarball, and publication decision tests before packing it. Generation is verified
both in PR validation and before publication.

The exact compared tarball is published using npm trusted publishing, public
access, and provenance, without rerunning lifecycle scripts.

### Package comparison and version checks

- **Compare:** The [publication helper](../script/prepare-publication.mjs) compares
  packed paths and contents against the most recent npm publication by timestamp,
  including prereleases, not a Go tag, repository diff, or npm's `latest` tag.
- **Counts as a change:** JavaScript, declarations, consumer README, licenses,
  notices, and authored package metadata.
- **Ignored:** Archive metadata, manifest formatting, `version`, `gitHead`, npm
  bookkeeping (`_*`), and unpackaged files such as maintainer docs. A version
  bump alone does not publish.
- **Unchanged:** Succeed with an explicit skip, including reruns after publication.
- **Changed:** Require a valid version and matching names/versions in
  `package.json` and `package-lock.json`. The version must be unused and
  semantically newer than every published version. **No automatic bumps.**
- **First publication:** Only confirmed npm package absence counts; a changed
  payload must still pass version and lockfile checks.
- **Errors:** Registry/authentication, network, malformed-response, download, and
  integrity failures stop publication; they never mean "first publication."
- **Retry:** Fix the reported problem and release or dispatch an approved corrected
  ref. Every attempt compares against npm again, so skipped or failed Go releases
  do not change the baseline.

## Publishing permissions and versioning

Adding this package does not publish it or change registry settings. Do not
enable publication until permissions, version, public access, and trusted
publishing have been approved.

Before the first release, a maintainer must:

1. Confirm npm organization permissions and package-name availability to publish
   public `@github/license-categories`.
2. Select and approve the initial version separately. The current manifest and
   lockfile retain `0.1.2`; the package rename did not select a new initial
   version. Confirm the approved version is unused under the new package name
   and update both files if necessary. npm versions cannot be republished.
   Bump the version for every subsequent release.
3. Configure the new package's npm trusted publisher for GitHub Actions: owner
   `github`, repository `go-spdx`, workflow filename
   `publish-license-categories.yaml`, with no environment restriction (the
   workflow does not use an environment). No long-lived npm token is required.
   If first-publication setup requires a separate bootstrap, arrange that with
   an authorized npm maintainer before enabling this workflow.
4. After release approval, enable the go-spdx repository variable and publish
   the Go release, or dispatch the publication workflow with an explicitly
   selected reviewed ref for an exception. It publishes only changed,
   version-validated contents with public access and provenance. A workflow
   filename or environment change requires updating the npm trusted publisher.
5. Verify anonymous installation of `@github/license-categories` from npmjs and
   the published package's MIT software license, upstream data attribution,
   and provenance.

## Migration and cutover

`@github/license-categories` is a new package identity, distinct from the
historical `@github-ui/license-categories` package. Consumers must update their
dependency name and imports to adopt it; the exported API and data are unchanged.
Publishing the new package does not rename, change the visibility of, or replace
the trusted-publisher settings of the historical package. Its publisher binding
is independent of the new package's configuration.

Keep historical generation and publication intact until the go-spdx addition
is reviewed and ready. Do not change the historical publisher as part of this
addition. After verifying the new public package, coordinate consumer migration
and any separately approved old-generation cleanup.
