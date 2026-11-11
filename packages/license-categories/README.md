# @github/license-categories

SPDX license names and compliance risk categories, generated and published from
[go-spdx](https://github.com/github/go-spdx). The software is MIT-licensed;
the incorporated ScanCode metadata retains its upstream CC-BY-4.0 terms.

## Install and use

```sh
npm install @github/license-categories
```

The package targets Node.js 20+ and ESM consumers, with TypeScript declarations.

```js
import {
  categoryDescriptions,
  licensesById,
  categoryForLicense,
  categoriesWithLicenses,
} from '@github/license-categories'

categoryForLicense('MIT') // 'Permissive'
categoryForLicense('not-an-spdx-id') // undefined
licensesById.MIT // {name: 'MIT License', category: 'Permissive'}
categoryDescriptions.Permissive // 'Safe for most'
categoriesWithLicenses() // [{category, description, licenses: [{id, name}]}]
```

Lookups are case-sensitive SPDX IDs, not license expressions. Groups are sorted
by category, with licenses sorted by name and then ID.

## Data sources

SPDX IDs and the SPDX License List version come from go-spdx's
`spdxexp/spdxlicenses` package. License names and categories come from
[ScanCode LicenseDB](https://github.com/aboutcode-org/scancode-licensedb), pinned
by commit and verified by SHA-256 during generation. Category spellings are
normalized, with explicit, reviewed overrides; IDs absent from ScanCode are
assigned `Unknown`.

## Contributing

See [Operations](https://github.com/github/go-spdx/blob/main/packages/license-categories/docs/operations.md)
for generation, building, maintenance, publishing, and migration details, and
[Overrides and exclusions](https://github.com/github/go-spdx/blob/main/packages/license-categories/docs/overrides-and-exclusions.md)
for correcting category assignments or excluding licenses from the package data.

## License

The software is licensed under the [MIT License](./LICENSE). ScanCode LicenseDB
metadata is copyright nexB Inc. and others, under
[CC-BY-4.0](https://creativecommons.org/licenses/by/4.0/). The package's
[THIRD_PARTY_NOTICE.md](./THIRD_PARTY_NOTICE.md) preserves attribution and modifications,
and points to the generated data's exact source revision.
The data and categories are not legal advice.
