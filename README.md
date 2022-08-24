# spdx-expression

Golang implementation of a checker for determining if an SPDX ID satisfies an SPDX Expression.

Public API:

```go
Parse(expression string) (*Node, error)
Satisfies(spdxID string, expression string)
```

Example expressions:

```go
"MIT"
"MIT AND Apache-2.0"
"MIT OR Apache-2.0"
"MIT AND (Apache-1.0 OR Apache-2.0)"
"Apache-1.0+"
"DocumentRef-spdx-tool-1.2:LicenseRef-MIT-Style-2"
"GPL-2.0 WITH Bison-exception-2.2"
```
