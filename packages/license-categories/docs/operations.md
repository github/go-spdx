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
the API, tarball, workflow-contract, and publication-preparation tests, including
importing an extracted package independently of the source tree.
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
is publication-only. It runs when a GitHub release is published, checking out that
release's exact tag, or through `workflow_dispatch`, using the selected ref.
Draft releases do not trigger publication. Releases marked **pre-release** in GitHub
skip the publication job. This checks GitHub's `release.prerelease` flag, not the
tag's spelling or suffix. Pushes do not trigger publication.

The single publication job is disabled unless the repository variable
`LICENSE_CATEGORIES_NPM_PUBLISH_ENABLED` is exactly `true`. Publications are
serialized across all refs without canceling a running publication. There is no
default-branch-only restriction. Dispatch bypasses only the release event
restriction, not enablement, generated-data verification, package comparison,
or version checks. Only release or dispatch trusted, reviewed refs. npm semantic
prerelease versions follow the same ordering rules; the GitHub release flag does
not change those rules. Backports on older release tracks would require a future
version-policy change.

The job verifies committed generated data byte for byte with
`make verify-license-categories`, installs npm dependencies with a lockfile-backed
cache, and runs `npm test`, including API and tarball tests. It uses Node.js 24 and
npm 11.20.0 for OIDC authentication, with no long-lived npm token. Workflow default
permissions are read-only; only the publication job receives OIDC permission.
Checkout does not retain Git write credentials.

## Publication artifact preparation and version checks

The [publication helper](../script/prepare-publication.mjs) prepares and compares
the already-built package after tests. Its `changed` output gates publication:
unchanged contents skip publishing, while eligible changed contents are published
from the exact compared tarball using `npm publish --ignore-scripts` with public
access. Lifecycle scripts do not rebuild the compared artifact. npm provenance
remains explicitly disabled, and no publication header is stamped.

To run the helper separately after building:

```sh
npm run build --prefix packages/license-categories
node packages/license-categories/script/prepare-publication.mjs /absolute/path/candidate.tgz
```

The helper uses `npm pack --ignore-scripts` and retains the exact candidate
tarball at the requested destination only when publication is eligible. It does
not publish, regenerate data, stamp provenance, or bump versions. Its behavior is:

- **Compare:** Compare every packaged file against the most recent npm publication
  by timestamp, including prereleases, not a Go tag or npm's `latest` tag. Archive
  metadata and entry order do not affect comparison.
- **Ignore:** Manifest version, npm bookkeeping fields (leading `_` and `gitHead`),
  and top-level manifest JSON key order. Unpackaged maintainer docs do not affect contents.
- **Preserve:** API code, generated data, comments and provenance, declarations,
  consumer README, license/attribution notices, consumer metadata, and file additions
  or removals all remain significant. Nested manifest key order is preserved because
  conditional `exports`/`imports` and overlapping `typesVersions` ranges can depend
  on it. A provenance-only change can require a new version.
- **Unchanged:** Succeed with an explicit skip, including reruns and version-only
  changes. No candidate is written for that attempt. Version checks apply only
  after finding changed contents or confirming first publication.
- **Changed:** Require matching names and versions in `package.json` and
  `package-lock.json`, with a valid, unused semantic version newer than every
  published version. Prerelease versions follow semantic version ordering.
  Versions recorded in npm's publication history remain used even after unpublishing.
  Update both package files explicitly; no automatic bump is performed.
- **First publication:** Only an explicit npm package-not-found response counts.
  Manifest and lockfile validation still applies.
- **Errors:** Registry/authentication, network, malformed metadata, archive,
  download, and integrity failures stop preparation instead of assuming first
  publication. Downloads and unpacked contents are limited to 100 MiB; archive
  listings are limited to 100 entries and 1 MiB of cumulative UTF-8 path bytes,
  including directory entries. File entries must be regular files under `package/`
  with safe, unique paths. Invalid archives reject preparation through its normal
  error handling, removing temporary files without writing a candidate tarball.

When run in GitHub Actions, the helper accepts only a published release tag or a
`workflow_dispatch` selected ref and checks that the resolved ref matches checkout
`HEAD`.
Both Git checks run in the package directory containing the helper, so they
validate the checkout being packaged regardless of the caller's working directory.
The helper writes `changed=true` or `changed=false` to `GITHUB_OUTPUT` when set.
Retries recheck npm rather than using a prior failed or skipped run as a baseline.
Fix the reported problem and release or dispatch an approved corrected ref. A
repeated attempt cannot reuse a published version for changed contents.

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
4. After publication approval, set the repository variable to `true` and publish
   an approved Go release, or dispatch the workflow with an explicitly selected
   reviewed ref. Both paths require changed, version-validated package contents.
   A workflow filename or environment change requires updating the npm trusted
   publisher.
5. Verify anonymous installation from npmjs, the MIT software license, and
   upstream data attribution.
