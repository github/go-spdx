# Contributing to license categories

This guide covers generation and operation of `@github/license-categories`.
See [Contributing](../../../CONTRIBUTING.md) for repository-wide contribution guidance
and the [package README](../README.md) for consumer installation, API usage,
compatibility, and licensing. See [Overrides and exclusions](./overrides-and-exclusions.md)
for correcting individual category assignments or omitting licenses from the npm data.

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
file without overwriting it. Only the leading `// go-spdx:` provenance value
is ignored, avoiding a self-reference when committing the generated checkout SHA.
The generator header and provenance labels/order must match, with a nonempty
value on each provenance line. The SPDX list version, ScanCode commit, actual data,
and all other content must match exactly.

For a build without tests, run:

```sh
npm run build --prefix packages/license-categories
```

The build removes `dist` before compiling TypeScript. `npm test` builds and runs
API, tarball, publication decision, and provenance tests, including importing an
extracted package independently of the source tree.
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
source changes. Do not edit generated data by hand. When updating the ScanCode pin,
update both its commit and archive SHA-256, regenerate the data, and review
overrides, warnings, and attribution in the package's
[THIRD_PARTY_NOTICE.md](../THIRD_PARTY_NOTICE.md). Preserve upstream data licensing
and any other third-party notices.

## Validation and publication workflows

[Validate License Categories](../../../.github/workflows/validate-license-categories.yaml)
verifies committed generated data and runs npm tests on relevant pull requests
and pushes to `main`. It has read-only permissions and cannot publish. The existing
[Build and test workflow](../../../.github/workflows/test.yaml) runs the Go generator
tests through `go test -race ./...`; validation does not duplicate them.

[Publish License Categories](../../../.github/workflows/publish-license-categories.yaml)
is publication-only. It runs when a GitHub release is published, checking out that
release's exact tag, or through `workflow_dispatch`, using the selected ref.
Draft releases do not trigger publication. Releases marked **pre-release** in GitHub
skip the publication job. This checks GitHub's `release.prerelease` flag, not the
tag's spelling or suffix.

The single publication job is disabled unless the repository variable
`LICENSE_CATEGORIES_NPM_PUBLISH_ENABLED` is exactly `true`. Publications are
serialized across all refs, without canceling a running publication. Dispatch
bypasses only the Go-release event restriction: it does not bypass the enablement
gate, generated-data verification, package-content comparison, or version checks.
Only dispatch trusted, reviewed refs. npm semantic prerelease versions follow the
same ordering rules; the GitHub release flag does not change those rules.
Backports on older release tracks would require a future version-policy change.

Publication independently runs `make verify-license-categories` using the Go
version from `go.mod`, downloading the pinned ScanCode archive and comparing fresh
output. Only the leading go-spdx provenance value is ignored. It then installs
lockfile-backed dependencies with npm caching and stamps only the `// go-spdx:`
header in the Actions checkout's `src/generated.ts`:

- **Release:** Actual release tag plus ` (release)`, after confirming the tag
  identifies the checked-out commit.
- **Dispatch:** Full checked-out `git rev-parse HEAD` plus ` (dispatch)`,
  not a moving branch tip.

Stamping follows verification and does not regenerate assignments, write back
to the repository, or create a commit. Publication runs `npm test`, which cleanly
builds the verified source with its publication header and runs all package tests.
The job uses Node.js 24 and npm 11.20.0 for OIDC authentication. It publishes the
exact compared tarball with public access and `--provenance`, without rerunning
lifecycle scripts or retaining Git write credentials.

### Package comparison and version checks

- **Compare:** The [publication helper](../script/prepare-publication.mjs) compares
  every packaged file against the most recent npm publication by timestamp,
  including prereleases. The baseline is not a Go tag, repository diff, or npm's
  `latest` tag. Tarball entry order and archive metadata do not affect comparison.
- **Ignored:** Manifest version, npm bookkeeping fields (leading `_` and `gitHead`),
  and manifest JSON key order. Unpackaged maintainer docs do not affect contents.
- **Significant:** API code, generated data, comments and provenance, declarations,
  consumer README, license/attribution notices, consumer metadata, and file additions
  or removals. **At this intermediate stack layer, a different publication stamp
  alone can count as a package change and require a new version.**
- **Unchanged:** Succeed with an explicit skip, including reruns and version-only
  changes. Manual dispatch uses the same content gate; it does not force publication.
- **Changed:** Require a valid version and matching names/versions in `package.json`
  and `package-lock.json`. The version must be unused and semantically newer than
  every published version. **No automatic bumps.**
- **First publication:** Only confirmed npm package absence counts; version and
  lockfile checks still apply.
- **Errors:** Registry/authentication, network, malformed-response, archive,
  download, and integrity failures stop publication; they never mean "first publication."
- **Retry:** Fix the reported problem and release or dispatch an approved corrected
  ref. Every attempt checks npm again, so skipped or failed Go releases do not change
  the baseline. A repeated attempt cannot reuse a published version for changed contents.

Ignoring all leading provenance during source verification and comparing only
literal license data for automatic releases, with forced manual dispatch, are
reserved for the next stack layer. Scheduled ScanCode updates and npm regeneration
in the existing fetch workflow are also separate; this layer does not change that workflow.

## Publishing permissions and versioning

Adding this package does not publish it or change registry settings. Do not enable
publication until permissions, version, public access, and trusted publishing are approved.

Before the first release, a maintainer must:

1. Confirm npm organization permissions and package-name availability to publish
   public `@github/license-categories`.
2. Select and approve the initial version separately. The manifest and lockfile
   retain `0.1.2`; the package rename did not select a new initial version.
   Confirm it is approved and unused under the new package name, and update both
   files if necessary. npm versions cannot be republished.
3. Configure the new package's npm trusted publisher for GitHub Actions: owner
   `github`, repository `go-spdx`, workflow filename `publish-license-categories.yaml`,
   with no environment restriction (the workflow does not use an environment).
   No long-lived npm token is required. If first-publication setup requires a separate
   bootstrap, arrange that with an authorized npm maintainer before enabling this workflow.
4. After release approval, enable the repository variable and publish the Go release,
   or dispatch the publication workflow with an explicitly selected reviewed ref.
   Both paths require changed, version-validated package contents with public access
   and provenance. A workflow filename or environment change requires updating
   the npm trusted publisher.
5. Verify anonymous installation from npmjs, the MIT software license, upstream data
   attribution, and provenance.

## Migration and cutover

`@github/license-categories` is a new package identity, distinct from the historical
`@github-ui/license-categories` package. Consumers must update their dependency name
and imports to adopt it; the exported API and literal data are unchanged. Publishing
the new package does not rename, change the visibility of, or replace the trusted-publisher
settings of the historical package. Its publisher binding is independent.

Keep historical generation and publication intact until the go-spdx addition is
reviewed and ready. Do not change the historical publisher as part of this addition.
After verifying the new public package, coordinate consumer migration and any
separately approved old-generation cleanup.
