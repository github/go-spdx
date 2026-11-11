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
uses the generator's `-verify` mode to compare fresh output to the checked-in
file without overwriting it. The three leading provenance values (`go-spdx`,
`spdx-license-list-version`, and `scancode-licensedb commit`) are ignored,
including the self-reference when committing the generated checkout SHA.
The generator header and provenance labels/order must match, with a nonempty
value on each provenance line. All remaining content, including non-provenance
comments, must match exactly. Provenance header values cannot contain
CR, LF, or Unicode line/paragraph separators (U+2028/U+2029), which terminate
TypeScript line comments.

For a build without tests, run:

```sh
npm run build --prefix packages/license-categories
```

The build removes `dist` before compiling TypeScript. `npm test` builds and runs
the API, tarball, workflow-contract, publication-preparation, and provenance tests, including
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

`src/generated.ts` records the SPDX list version and ScanCode commit. The
go-spdx provenance is recorded according to the generation or publication context:

- Local checkout: `// go-spdx: <full checkout HEAD SHA> (local)`.
- Manual publication: `// go-spdx: <full checked-out commit SHA> (dispatch)`.
- Release publication: `// go-spdx: v2.9.0 (release)`, using the actual release tag.

Versioned Go installations retain their module version. The local SHA identifies
the source checkout's base commit, not an assertion that uncommitted files match
that revision. It is resolved from the generator's source directory, not an
unrelated working directory. A local build without an accessible source checkout
(including a trimmed source path) fails explicitly; use the checkout or a versioned
installation rather than recording a fabricated revision. No timestamps are generated.

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
is publication-only. It listens for `release: released`, which runs when a stable
GitHub release is first published or an already published prerelease is promoted
to a stable release, checking out that release's exact tag. It also runs through
`workflow_dispatch`, using the selected ref.
Draft releases and prereleases do not trigger publication. The publication job
also checks GitHub's `release.prerelease` flag, not the tag's spelling or suffix.
A promoted release can retain its original tag without being marked as a
prerelease. Pushes do not trigger publication.

The single publication job is disabled unless the repository variable
`LICENSE_CATEGORIES_NPM_PUBLISH_ENABLED` is exactly `true`. Publications are
serialized across all refs without canceling a running publication. There is no
default-branch-only restriction. Dispatch bypasses the Go-release event
requirement and unchanged-data skip, not the
enablement gate, generated-data validation, or version checks. Dispatch is the
explicit way to publish non-data fixes; only release or dispatch trusted, reviewed refs.
npm semantic prerelease versions follow the same ordering rules; the GitHub
release flag does not change those rules.
Backports on older release tracks would require a future version-policy change.

Publication independently runs `make verify-license-categories` using the Go
version from `go.mod`, downloading the pinned ScanCode archive and comparing fresh
output. Only the three leading provenance values are ignored. It then installs
lockfile-backed dependencies with npm caching and stamps only the `// go-spdx:`
header in the Actions checkout's `src/generated.ts`:

- **Release:** Actual release tag plus ` (release)`, after confirming the tag
  identifies the checked-out commit.
- **Dispatch:** Full checked-out `git rev-parse HEAD` plus ` (dispatch)`,
  not a moving branch tip.

Before stamping, both paths require checkout `HEAD` to match the run's full
`GITHUB_SHA`, which npm uses for source attestation. A missing, malformed, or
mismatched event SHA stops the job before the source is stamped or built.
A queued or rerun dispatch whose selected branch has advanced must use a new
dispatch run for that commit rather than publishing with mismatched provenance.

Release tags and stamped provenance values must not contain CR, LF, U+2028, or
U+2029; these are rejected rather than allowing a header to introduce source code.

Stamping follows verification and does not regenerate assignments, write back
to the repository, or create a commit. Publication runs `npm test`, which cleanly
builds the verified source with its publication header and runs all package tests.
The job uses Node.js 24 and npm 11.20.0 for OIDC authentication, with no long-lived
npm token. Workflow default permissions are read-only; only the publication job
receives OIDC permission. Checkout does not retain Git write credentials.

## Publication artifact preparation and version checks

The [publication helper](../script/prepare-publication.mjs) prepares and compares
the already-built package after tests. Its `changed` output gates publication:
unchanged data skips automatic publishing, while eligible releases or dispatches publish
from the exact compared tarball using `npm publish --ignore-scripts --provenance`
with public access. Lifecycle scripts do not rebuild the compared artifact.

To run the helper separately after building:

```sh
npm run build --prefix packages/license-categories
node packages/license-categories/script/prepare-publication.mjs /absolute/path/candidate.tgz
```

The helper uses `npm pack --ignore-scripts` and retains the exact candidate
tarball at the requested destination only when publication is eligible. It does
not publish, regenerate data, stamp provenance, or bump versions. Its behavior is:

- **Compare:** Compare literal `categoryDescriptions` and `licensesById` from packed `dist/generated.js`
  against the most recent npm publication by timestamp, including prereleases.
  The baseline is not a Go tag, repository diff, or npm's `latest` tag.
- **Counts as a change:** License IDs, membership, names, categories, or category descriptions.
- **Ignored:** The entire generated header, all comments, formatting, quoting,
  and object ordering. Other package files, metadata, versions, and unpackaged
  maintainer docs do not trigger automatic publication.
- **Unchanged release:** Succeed with an explicit skip, including reruns, even
  after a version bump. Non-data fixes ship with the next data update unless
  published sooner by manual dispatch. No candidate is written for a skipped attempt.
- **Manual dispatch:** Force publication even without data changes, only after
  verifying the Actions event and selected ref; all data, version, and enablement
  checks still apply.
- **Changed data or dispatch:** Require a valid version and matching names/versions in
  `package.json` and `package-lock.json`. The version must be unused and
  semantically newer than every published version. **No automatic bumps.**
  Versions recorded in npm's publication history remain used even after unpublishing.
- **First publication:** Only confirmed npm package absence counts; valid
  generated data, version, and lockfile checks are still required.
- **Safe parsing:** Use the existing TypeScript parser to read literal data,
  never import or execute downloaded JavaScript. Missing, malformed, duplicate,
  non-literal, or unsupported generated data fails explicitly, even on dispatch.
  Packaged manifests must parse and identify the expected package before comparison.
- **Verification is separate:** Committed-source verification ignores only the
  three leading provenance values. Header structure, actual data, and all other
  content, including non-provenance comments, must still match.
- **Errors:** Registry/authentication, network, malformed-response, archive,
  download, and integrity failures stop publication; they never mean "first publication."
  Downloads and unpacked contents are limited to 100 MiB; archive listings are
  limited to 100 entries and 1 MiB of cumulative UTF-8 path bytes, including directories.
  Files must be regular entries under `package/` with safe, unique paths.
  Invalid archives reject normally, removing temporary files without writing a candidate.
- **Retry:** Fix the reported problem and release or dispatch an approved corrected
  ref. Every attempt checks npm again, so skipped or failed Go releases do not change
  the baseline. A repeated dispatch cannot reuse a published version;
  bump both version files before dispatching again.

When run in GitHub Actions, the helper accepts only a `released` event's tag or a
`workflow_dispatch` selected ref and checks that the resolved ref matches checkout
`HEAD`.
Both Git checks run in the package directory containing the helper, so they
validate the checkout being packaged regardless of the caller's working directory.
The helper writes `changed=true` or `changed=false` to `GITHUB_OUTPUT` when set.
Retries recheck npm rather than using a prior failed or skipped run as a baseline.
Fix the reported problem and release or dispatch an approved corrected ref. A
repeated attempt cannot reuse a published version for changed contents.

Scheduled ScanCode updates and npm regeneration in the existing fetch workflow
are separate; this layer does not change that workflow. The existing SPDX list
identifier behavior is unchanged; expanding a no-version identifier to a full SHA
is deferred.

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
   reviewed ref. Releases require changed data; dispatch may force publication.
   Both paths require version-validated contents with public access and provenance.
   A workflow filename or environment change requires updating the npm trusted
   publisher.
5. Verify anonymous installation from npmjs, the MIT software license,
   upstream data attribution, and provenance.
