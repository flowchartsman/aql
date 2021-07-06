# AQL

AQL is a Lucene-inspired query language for Go [currently targeted for arbitrary JSON].
It aims to provide the following features:

* Simple, recognizable syntax
* Expressive composition of boolean expressions
* Extensible user-defined operators
* A generic query front-end for multiple data sources (eventually)

Currently it is used for searching arbitrary JSON data.

# Usage

```go
import (
	"strings"

	"github.com/flowchartsman/aql/jsonquery"
)

const json = `{
    "date": "1970-01-02",
    "number": 2,
    "name": "Andy",
    "description": "大懒虫"
}`

func main() {
	js := strings.NewReader(json)
	q, err := jsonquery.NewQuerier(`date:><[1970-01-01,1970-01-03]`)
	if err != nil {
		panic(err.Error())
	}
	result, err := q.Match(js)
	if err != nil {
		panic(err.Error())
	}
	println(result) // true
}
```
# AQL Syntax

In its simplest form, an AQL query is just a field identifier followed by a search term describing a value to match:

`<field>:<value>`

For example, given the following JSON:

```json
{
    "text":"hello"
}
```

The following query returns true:

`text: "hello"`==true

## Paths

Nested values are specified by a path, using dot-separated fields:

```json
{
    "level1": {
        "level2": {
            "level3": "here!"
        }
    }
}
```
`level1.level2.level3:"here!"` == true

## Array Introspection

If the targetted field is an array, all values will be tested, and the query will return true if it finds one that matches:

```json
{
    "outer":{
        "inner": [
            "hello",
            "world",
            "from AQL"
        ]
    }
}
```
`outer.inner:"world"`==true

However, intervening arrays will need to use the `*` path wildcard to ensure inspection:

```json
{
    "outer": [
        {
            "inner": "hello"
        },
        {
            "inner": "world"
        },
        {
            "inner": "from AQL"
        }
    ]
}
```
`outer.*.inner:"world"`==true

(this is a limitation of the pathfinding in the underlying `gabs` library, and may be resolved in a future version)

## Boolean Logic

Queries can also be combined using `AND` and `OR` or negated with `NOT`/`!`:

```json
{
    "text":"AQL",
    "number":1
}
```
`text:"AQL" AND number:1` == true

`text:"Nope" OR number:1` == true

`text:"AQL" AND !number:1` == false

Parenthetical grouping is also supported:

`(text:"Nope" OR text:"AQL") AND number:1`==true

## Types
AQL recognizes several different types of terms:

| Type | Examples | Description | Notes |
|------|----------|-------------|-------|
|string|`"hello"`|a literal string|supports regular and unicode escaping|
|integer|`1`|an integer number| |
|floating point|`1.0`|a floating point number| |
|timestamp|`1970-01-02`<br/><br/>`1970-01-02T00:00:00Z`|A string representing a moment in time, following the [RFC3339](https://datatracker.ietf.org/doc/html/rfc3339) standard format|[**DateTime** or **FullDate** values](https://datatracker.ietf.org/doc/html/rfc3339#section-5.6) are supported|
|CIDR|`192.168.0.0/16`|a network block| |
|boolean|`true`<br /><br/>`false`|a boolean literal value| |
|regex|`/^hello to \d{2} people$/`|a regular expression for advanced string matching|uses [Go regex syntax](https://golang.org/pkg/regexp/syntax/)|
|Exists|`exists`|a special value to check that a field exists|cannot be combined with other terms in the same clause|

**Note**: Not all terms work with all operators, see the next section for details

## Operators
AQL can also perform many different types of checks, depending on the type of data.

### Equality
`field:value`

This is the basic equality check we've seen so far.

|Supported Types|Examples|Notes|
|---------------|--------|-----|
|string|`field:"value"`|searches for the exact string value provided|
|integer|`field:1`|searches for a numeric value of the exact value provided|
|float|`field:1.0`|searches for a numeric value of the exact value provided|
|timestamp|`field:1970-01-01`<br/><br />`field:1970-01-02T15:53:33+00:00`|searches for a string whith represnts this date. AQL attempts to detect a number of different possible time representations to make this check. For details, see [here](https://github.com/araddon/dateparse#extended-example). Note that this check is currently for the exact timestamp specified, and other operations may be more useful for working with timestamps.
|boolean|`field:true`<br/><br/>`field:false`|searches for a JSON boolean of the exact value provided|
|exists|`field:exists`|matches if the field exists in the document in any surveyed location. This is a special value and only works with an equality check

### Equality set
`field:(value1, value2, ...)`

Analagous to a SQL IN query, this is basically a shorthand for `field:value1 OR field:value2 OR ...`

|Supported Types|Examples|
|---------------|--------|
|string|`field:("value1", "value2", "value3")`|
|integer|`field:(1,2)`|
|float|`field:(1.1, 2.2)`|
|timestamp|`field:(1970-01-01, 1970-01-02T15:53:33Z)`|


### Numeric Comparison
`field:>value`

`field:>=value`

`field:<value`

`field:<=value`

These operations operate on numbers or those values which can be meaningfully compared numerically.

|Supported Types|Examples|Notes|
|---------------|--------|-----|
|integer|`field:>1`|searches for a numeric value greater than 1|
|float|`field:<2.5`|searches for a numeric value less than 2.5 |
|timestamp|`field:<=1970-01-01`|Attempts to match a timestamp in one of the  [recognized formats](https://github.com/araddon/dateparse#extended-example) that occurs on or before midnight, UTC, January 1, 1970|
||`field:>=1970-01-02T15:53:33−05:00`|Attempts to match a timestamp in one of the  [recognized formats](https://github.com/araddon/dateparse#extended-example) that occurs on or after 3:53 PM, EST, January 1, 1970|

### Between
`field:><(value1, value2)`

This operation attempts to search for a value which falls between the two provided terms. **Note**: only supports two terms.

|Supported Types|Examples|Notes|
|---------------|--------|-----|
|integer|`field:><(1, 2)`|searches for a numeric value greater than 1 and less than 2|
|float|`field:><(2.1, 2.2)`|searches for a numeric value less than 2.5 |
|timestamp|`field:><(1970-01-01, 1970-01-02)`|Attempts to match a timestamp in one of the  [recognized formats](https://github.com/araddon/dateparse#extended-example) that occurs between midnight, January 1, 1970 and midnight, January 2, 1970|

### Similarity
`field:~value`

This operation allows searching on values that are similar to the provided value.

|Supported Types|Examples|Notes|
|---------------|--------|-----|
|string|`field:~"wildcar? *"`|wildcard match. `*` is any number of characters, while `?` is any one character|
|regex|`field:~/attack of the \d+ foot (?:cat\|dog)/`|matches a string where a dog or cat of any height attacks (uses [Go regex syntax](https://golang.org/pkg/regexp/syntax/))|
|boolean|`field:~true`<br/><br/>`field:~false`|boolean similarity will search for "truthy" things, which are: boolean `true`, string "true", numeric 1, string "1", while boolean false will search for "falsy" things, which are boolean `false`, string "false", numeric 0, and string "0"|
|CIDR|`field:~192.168.0.0/24`|Attempts to match a string representing an IP address between 192.168.0.0 and 192.168.0.255|


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