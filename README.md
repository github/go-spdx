# spdx-expression

Golang implementation of a checker for determining if a set of SPDX IDs satisfies an SPDX Expression.

## Public API

_NOTE: The public API is initially limited to the Satisfies and ValidateLicenses functions.  If
there is interest in the output of the parser or license checking being public, please submit an
issue for consideration._

### Function: Satisfies

```go
Satisfies(testExpression string, allowedList []string, options *Options)
```

**Parameter: testExpression**

testExpression is an [SPDX expression](https://spdx.github.io/spdx-spec/v2.3/SPDX-license-expressions/#d1-overview) describing the licensing terms of source code or a binary file.

Example expressions that can be used for testExpression:

```go
"MIT"
"MIT AND Apache-2.0"
"MIT OR Apache-2.0"
"MIT AND (Apache-1.0 OR Apache-2.0)"
"Apache-1.0+"
"DocumentRef-spdx-tool-1.2:LicenseRef-MIT-Style-2"
"GPL-2.0 WITH Bison-exception-2.2"
```

_See satisfies_test.go for more example expressions._

**Parameter: allowedList**

allowedList is an array of single licenses describing what licenses can be used to satisfy the testExpression.

Example allowedList:

```go
[]string{"MIT"}
[]string{"MIT", "Apache-2.0"}
[]string{"MIT", "Apache-2.0", "ISC", "GPL-2.0"}
[]string{"MIT", "Apache-1.0+"}
[]string{"GPL-2.0-or-later"}
```

**N.B.** If at least one of expressions from `allowedList` is not a valid SPDX expression, the call
to `Satisfies` will produce an error. Use [`ValidateLicenses`](###-ValidateLicenses) function
to first check if all of the expressions from `allowedList` are valid.

#### Examples: Satisfies returns true

```go
Satisfies("MIT", []string{"MIT"})
Satisfies("MIT", []string{"MIT", "Apache-2.0"})
Satisfies("Apache-2.0", []string{"Apache-1.0+"})
Satisfies("MIT OR Apache-2.0", []string{"Apache-2.0"})
Satisfies("MIT OR Apache-2.0", []string{"MIT", "Apache-2.0"})
Satisfies("MIT AND Apache-2.0", []string{"MIT", "Apache-2.0"})
Satisfies("MIT AND Apache-2.0", []string{"MIT", "Apache-2.0", "GPL-2.0"})
```

#### Examples: Satisfies returns false

```go
Satisfies("MIT", []string{"Apache-2.0"})
Satisfies("Apache-1.0", []string{"Apache-2.0+"})
Satisfies("MIT AND Apache-2.0", []string{"MIT"})
```

### ValidateLicenses

```go
func ValidateLicenses(licenses []string) (bool, []string)
```

Function `ValidateLicenses` is used to determine if any of the provided license expressions is
invalid.

**parameter: licenses**

Licenses is a slice of strings which must be validated as SPDX expressions.

**returns**

Function `ValidateLicenses` has 2 return values. First is `bool` which equals `true` if all of
the provided licenses provided are valid, and `false` otherwise.

The second parameter is a slice of all invalid licenses which were provided.

#### Examples: ValidateLicenses returns no invalid licenses

```go
valid, invalidLicenses := ValidateLicenses([]string{"Apache-2.0"})
assert.True(valid)
assert.Empty(invalidLicenses)
```

#### Examples: ValidateLicenses returns invalid licenses

```go
valid, invalidLicenses := ValidateLicenses([]string{"NON-EXISTENT-LICENSE", "MIT"})
assert.False(valid)
assert.Contains(invalidLicenses, "NON-EXISTENT-LICENSE")
assert.NotContains(invalidLicenses, "MIT")
```

## Background

This package was developed to support testing whether a repository's license requirements are met by an allowed-list of licenses.

Dependencies are defined in [go.mod](./go.mod).

Contributions and requests are welcome.  Refer to the [Contributing](#contributing) section for more information including how to set up a test environment and install dependencies.

## License

This project is licensed under the terms of the MIT open source license. Please refer to [MIT](./LICENSE.md) for the full terms.

## Maintainers

- @elrayle
- @ajhenry

## Support

You can expect the following support:

- bug fixes
- review of feature request issues
- review of questions in discussions

## Contributing

Contributions in the form of bug identification Issues, bug fix PRs, and feature requests are welcome.  See [CONTRIBUTING.md](./CONTRIBUTING.md) for more information on how to get involved and set up a testing environment.

_NOTE: The list of valid licenses is maintained manually.  If you notice a missing license, an excellent way to contribute to the long term viability of this package is to open an Issue or PR addressing the missing license._

## Acknowledgement

The process for parsing and evaluating expressions is a translation from JavaScript to Go based heavily on the JavaScript implementation defined across several repositories.

- [spdx-satisfies.js](https://github.com/clearlydefined/spdx-satisfies.js)
- [spdx-expression-parse.js](https://github.com/clearlydefined/spdx-expression-parse.js)
- [spdx-ranges](https://github.com/jslicense/spdx-ranges.js)
- [spdx-compare](https://github.com/jslicense/spdx-compare.js)
- [spdx-license-ids](https://github.com/jslicense/spdx-license-ids)
- [spdx-exceptions](https://github.com/jslicense/spdx-exceptions.json)
