# Path-to-RegExp

[![Build Status](https://travis-ci.org/soongo/path-to-regexp.svg)](https://travis-ci.org/soongo/path-to-regexp)
[![codecov](https://codecov.io/gh/soongo/path-to-regexp/branch/master/graph/badge.svg)](https://codecov.io/gh/soongo/path-to-regexp)
[![Go Report Card](https://goreportcard.com/badge/github.com/soongo/path-to-regexp)](https://goreportcard.com/report/github.com/soongo/path-to-regexp)
[![GoDoc](https://godoc.org/github.com/soongo/path-to-regexp?status.svg)](https://godoc.org/github.com/soongo/path-to-regexp)
[![License](https://img.shields.io/badge/MIT-green.svg)](https://opensource.org/licenses/MIT)

> Turn a path string such as `/user/:name` into a regular expression.

Thanks to [path-to-regexp](https://github.com/pillarjs/path-to-regexp) which is the original version written in javascript.

## Installation

To install `Path-to-RegExp` package, you need to install Go and set your Go workspace first.

The first need [Go](https://golang.org/) installed (**version 1.11+ is required**), then you can use the below Go command to install `Path-to-RegExp`.

```sh
$ go get -u github.com/soongo/path-to-regexp
```

## Usage

```go
import pathToRegexp "github.com/soongo/path-to-regexp"

// pathToRegexp.PathToRegexp(path, tokens, options) // tokens and options can be nil
// pathToRegexp.Parse(path, options) // options can be nil
// pathToRegexp.Compile(path, options) // options can be nil
// pathToRegexp.MustCompile(path, options) // like Compile but panics if the error is non-nil
// pathToRegexp.Match(path, options) // options can be nil
// pathToRegexp.MustMatch(path, options) // like Match but panics if the error is non-nil
// pathToRegexp.Must(regexp, err) // wraps a call to a function returning (*regexp2.Regexp, error) and panics if the error is non-nil
// pathToRegexp.EncodeURI(str) // encodes characters in URI except `;/?:@&=+$,#`, like javascript's encodeURI
// pathToRegexp.EncodeURIComponent(str) // encodes characters in URI, like javascript's encodeURIComponent
```

- **path** A string, array or slice of strings, or a regular expression with type *github.com/dlclark/regexp2.Regexp.
- **tokens** An array to populate with tokens found in the path.
  - token
    - **Name** The name of the token (`string` for named or `number` for index)
    - **Prefix** The prefix character for the segment (e.g. `/`)
    - **Delimiter** The delimiter for the segment (same as prefix or default delimiter)
    - **Optional** Indicates the token is optional (`boolean`)
    - **Repeat** Indicates the token is repeated (`boolean`)
    - **Pattern** The RegExp used to match this token (`string`)
- **options**
  - **Sensitive** When `true` the regexp will be case sensitive. (default: `false`)
  - **Strict** When `true` the regexp allows an optional trailing delimiter to match. (default: `false`)
  - **End** When `true` the regexp will match to the end of the string. (default: `true`)
  - **Start** When `true` the regexp will match from the beginning of the string. (default: `true`)
  - **Delimiter** The default delimiter for segments. (default: `'/'`)
  - **EndsWith** Optional character, or list of characters, to treat as "end" characters.
  - **Whitelist** List of characters to consider delimiters when parsing. (default: `nil`, any character)
  - **Encode** How to encode uri. (default: `func (uri string, token interface{}) string { return uri }`)
  - **Decode** How to decode uri. (default: `func (uri string, token interface{}) string { return uri }`)

```go
var tokens []pathToRegexp.Token
regexp := pathToRegexp.Must(pathToRegexp.PathToRegexp("/foo/:bar", &tokens, nil))
// regexp: ^\/foo\/([^\/]+?)(?:\/)?$
// tokens: [{Name:"bar", Prefix:"/", Delimiter:"/", Optional:false, Repeat:false, Pattern:"[^\\/]+?"}}]
```

**Please note:** The `Regexp` returned by `path-to-regexp` is intended for ordered data (e.g. pathnames, hostnames). It can not handle arbitrarily ordered data (e.g. query strings, URL fragments, JSON, etc).

### Parameters

The path argument is used to define parameters and populate the list of tokens.

#### Named Parameters

Named parameters are defined by prefixing a colon to the parameter name (`:foo`). By default, the parameter will match until the next prefix (e.g. `[^/]+`).

```go
regexp := pathToRegexp.Must(pathToRegexp.PathToRegexp("/:foo/:bar", nil, nil))
// tokens: [
//   {Name:"foo", Prefix:"/", Delimiter:"/", Optional:false, Repeat:false, Pattern:"[^\\/]+?"},
//   {Name:"bar", Prefix:"/", Delimiter:"/", Optional:false, Repeat:false, Pattern:"[^\\/]+?"}
// ]

match, err := regexp.FindStringMatch("/test/route")
for _, g := range match.Groups() {
    fmt.Printf("%q ", g.String())
}
fmt.Printf("%d, %q\n", match.Index, match)
//=> "/test/route" "test" "route" 0, "/test/route"
```

**Please note:** Parameter names must use "word characters" (`[A-Za-z0-9_]`).

#### Parameter Modifiers

##### Optional

Parameters can be suffixed with a question mark (`?`) to make the parameter optional.

```go
regexp := pathToRegexp.Must(pathToRegexp.PathToRegexp("/:foo/:bar?", nil, nil))
// tokens: [
//   {Name:"foo", Prefix:"/", Delimiter:"/", Optional:false, Repeat:false, Pattern:"[^\\/]+?"},
//   {Name:"bar", Prefix:"/", Delimiter:"/", Optional:true, Repeat:false, Pattern:"[^\\/]+?"}
// ]

match, err := regexp.FindStringMatch("/test")
for _, g := range match.Groups() {
    fmt.Printf("%q ", g.String())
}
fmt.Printf("%d, %q\n", match.Index, match)
//=> "/test" "test" "" 0, "/test"

match, err = regexp.FindStringMatch("/test/route")
for _, g := range match.Groups() {
    fmt.Printf("%q ", g.String())
}
fmt.Printf("%d, %q\n", match.Index, match)
//=> "/test/route" "test" "route" 0, "/test/route"
```

**Tip:** The prefix is also optional, escape the prefix `\/` to make it required.

##### Zero or more

Parameters can be suffixed with an asterisk (`*`) to denote a zero or more parameter matches. The prefix is used for each match.

```go
regexp := pathToRegexp.Must(pathToRegexp.PathToRegexp("/:foo*", nil, nil))
// tokens: [{Name:"foo", Prefix:"/", Delimiter:"/", Optional:true, Repeat:true, Pattern:"[^\\/]+?"}]

match, err := regexp.FindStringMatch("/")
for _, g := range match.Groups() {
    fmt.Printf("%q ", g.String())
}
fmt.Printf("%d, %q\n", match.Index, match)
//=> "/" "" 0, "/"

match, err = regexp.FindStringMatch("/bar/baz")
for _, g := range match.Groups() {
    fmt.Printf("%q ", g.String())
}
fmt.Printf("%d, %q\n", match.Index, match)
//=> "/bar/baz" "bar/baz" 0, "/bar/baz"
```

##### One or more

Parameters can be suffixed with a plus sign (`+`) to denote a one or more parameter matches. The prefix is used for each match.

```go
regexp := pathToRegexp.Must(pathToRegexp.PathToRegexp("/:foo+", nil, nil))
// tokens: [{Name:"foo", Prefix:"/", Delimiter:"/", Optional:false, Repeat:true, Pattern:"[^\\/]+?"}]

match, err := regexp.FindStringMatch("/")
fmt.Println(match)
//=> nil

match, err = regexp.FindStringMatch("/bar/baz")
for _, g := range match.Groups() {
    fmt.Printf("%q ", g.String())
}
fmt.Printf("%d, %q\n", match.Index, match)
//=> "/bar/baz" "bar/baz" 0, "/bar/baz"
```

#### Unnamed Parameters

It is possible to write an unnamed parameter that only consists of a matching group. It works the same as a named parameter, except it will be numerically indexed.

```go
regexp := pathToRegexp.Must(pathToRegexp.PathToRegexp("/:foo/(.*)", nil, nil))
// tokens: [
//   {Name:"foo", Prefix:"/", Delimiter:"/", Optional:false, Repeat:false, Pattern:"[^\\/]+?"},
//   {Name:0, Prefix:"/", Delimiter:"/", Optional:false, Repeat:false, Pattern:".*"}
// ]

match, err := regexp.FindStringMatch("/test/route")
for _, g := range match.Groups() {
    fmt.Printf("%q ", g.String())
}
fmt.Printf("%d, %q\n", match.Index, match)
//=> "/test/route" "test" "route" 0, "/test/route"
```

#### Custom Matching Parameters

All parameters can have a custom regexp, which overrides the default match (`[^/]+`). For example, you can match digits or names in a path:

```go
regexpNumbers := pathToRegexp.Must(pathToRegexp.PathToRegexp("/icon-:foo(\\d+).png", nil, nil))
// tokens: {Name:"foo", Prefix:"-", Delimiter:"-", Optional:false, Repeat:false, Pattern:"\\d+"}

match, err := regexpNumbers.FindStringMatch("/icon-123.png")
for _, g := range match.Groups() {
    fmt.Printf("%q ", g.String())
}
//=> "/icon-123.png" "123"

match, err = regexpNumbers.FindStringMatch("/icon-abc.png")
fmt.Println(match)
//=> nil

regexpWord := pathToRegexp.Must(pathToRegexp.PathToRegexp("/(user|u)", nil, nil))
// tokens: {Name:0, Prefix:"/", Delimiter:"/", Optional:false, Repeat:false, Pattern:"user|u"}

match, err = regexpWord.FindStringMatch("/u")
for _, g := range match.Groups() {
    fmt.Printf("%q ", g.String())
}
//=> "/u" "u"

match, err = regexpWord.FindStringMatch("/users")
fmt.Println(match)
//=> nil
```

**Tip:** Backslashes need to be escaped with another backslash in Go strings.

### Match

The `match` function will return a function for transforming paths into parameters:

```go
match := pathToRegexp.MustMatch("/user/:id", &pathToRegexp.Options{Decode: func(str string, token interface{}) string {
    return pathToRegexp.DecodeURIComponent(str)
}})

match("/user/123")
//=> &pathtoregexp.MatchResult{Path:"/user/123", Index:0, Params:map[interface {}]interface {}{"id":"123"}}

match("/invalid") //=> nil

match("/user/caf%C3%A9")
//=> &pathtoregexp.MatchResult{Path:"/user/caf%C3%A9", Index:0, Params:map[interface {}]interface {}{"id":"café"}}
```

### Parse

The `Parse` function will return a list of strings and tokens from a path string:

```go
tokens := pathToRegexp.Parse("/route/:foo/(.*)", nil)

fmt.Printf("%#v\n", tokens[0])
//=> "/route"

fmt.Printf("%#v\n", tokens[1])
//=> pathToRegexp.Token{Name:"foo", Prefix:"/", Delimiter:"/", Optional:false, Repeat:false, Pattern:"[^\\/]+?"}

fmt.Printf("%#v\n", tokens[2])
//=> pathToRegexp.Token{Name:0, Prefix:"/", Delimiter:"/", Optional:false, Repeat:false, Pattern:".*"}
```

**Note:** This method only works with strings.

### Compile ("Reverse" Path-To-RegExp)

The `Compile` function will return a function for transforming parameters into a valid path:

```go
toPath := pathToRegexp.MustCompile("/user/:id", &pathToRegexp.Options{Encode: func(str string, token interface{}) string {
    return pathToRegexp.EncodeURIComponent(str)
}})

toPath(map[string]int{"id": 123}) //=> "/user/123"
toPath(map[string]string{"id": "café"}) //=> "/user/caf%C3%A9"
toPath(map[string]string{"id": "/"}) //=> "/user/%2F"

toPath(map[string]string{"id": ":/"}) //=> "/user/%3A%2F"

// Without `encode`, you need to make sure inputs are encoded correctly.
falseValue := false
toPathRaw := pathToRegexp.MustCompile("/user/:id", &pathToRegexp.Options{Validate: &falseValue})
toPathRaw(map[string]string{"id": "%3A%2F"}); //=> "/user/%3A%2F"
toPathRaw(map[string]string{"id": ":/"}); //=> "/user/:/"

toPathRepeated := pathToRegexp.MustCompile("/:segment+", nil)

toPathRepeated(map[string]string{"segment": "foo"}) //=> "/foo"
toPathRepeated(map[string][]string{"segment": {"a", "b", "c"}}) //=> "/a/b/c"

toPathRegexp := pathToRegexp.MustCompile("/user/:id(\\d+)", &pathToRegexp.Options{Validate: &falseValue})

toPathRegexp(map[string]int{"id": 123}) //=> "/user/123"
toPathRegexp(map[string]string{"id": "123"}) //=> "/user/123"
toPathRegexp(map[string]string{"id": "abc"}) //=> "/user/abc"

toPathRegexp = pathToRegexp.MustCompile("/user/:id(\\d+)", nil)
toPathRegexp(map[string]string{"id": "abc"}) //=> panic
```

**Note:** The generated function will panic on invalid input. It will do all necessary checks to ensure the generated path is valid. This method only works with strings.
