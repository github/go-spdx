# spdx-expression

Golang implementation of a checker for determining if an SPDX ID satisfies an SPDX Expression.

## Public API

_NOTE: The public API is initially limited to the Satisfies method.  If there is interest in the
output of the parser or license checking being public, please submit an issue for consideration._

```go
Satisfies(testExpression string, allowedList []string)
```

### testExpression

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

### allowedList

allowedList is an array of single licenses describing what licenses can be used to satisfy the testExpression.

Example allowedList:

```go
[]string{"MIT"}
[]string{"MIT", "Apache-2.0"}
[]string{"MIT", "Apache-2.0", "ISC", "GPL-2.0"}
[]string{"MIT", "Apache-1.0+"}
[]string{"GPL-2.0-or-later"}
```

### Examples where Satisfies returns true

```go
Satisfies("MIT", []string{"MIT"})
Satisfies("MIT", []string{"MIT", "Apache-2.0"})
Satisfies("Apache-2.0", []string{"Apache-1.0+"})
Satisfies("MIT OR Apache-2.0", []string{"Apache-2.0"})
Satisfies("MIT OR Apache-2.0", []string{"MIT", "Apache-2.0"})
Satisfies("MIT AND Apache-2.0", []string{"MIT", "Apache-2.0"})
Satisfies("MIT AND Apache-2.0", []string{"MIT", "Apache-2.0", "GPL-2.0"})
```

### Examples where Satisfies returns false

```go
Satisfies("MIT", []string{"Apache-2.0"})
Satisfies("Apache-1.0", []string{"Apache-2.0+"})
Satisfies("MIT AND Apache-2.0", []string{"MIT"})
```

## Contributing


