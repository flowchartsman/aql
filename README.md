# AQL

AQL is a Lucene-inspired query language for Go [currently targeted for arbitrary JSON].
It aims to provide the following features:

* Simple, recognizable syntax
* Expressive composition of boolean expressions
* Extensible user-defined operators
* A generic query front-end for multiple data sources (eventually)

It is currently in a POC state under active development.

## Usage

```go
import (
    "github.com/flowchartsman/aql"
)
//TODO: only the parser is complete.
```

## Contributing
PRs welcome. Please file issues if your PR addresses a bug.

Please make sure to update tests as appropriate.

## License
[MIT](https://choosealicense.com/licenses/mit/)

## TODO
* Typing/verifying of types in parser
* Operator-first definitions
* Pluggable string-based user-provided types
* Pluggable user-provided operators
* Flexible query backend with selectable language features (backend doesn't support clause)