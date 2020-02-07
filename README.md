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

#### Warning

>*Version 3.2.0, 4.0.0, 4.0.5, 5.0.0 have been removed, please use corresponding version 1.3.2, 1.4.0, 1.4.5, 1.5.0.*


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
    - **Prefix** The prefix string for the segment (e.g. `"/"`)
    - **Suffix** The suffix string for the segment (e.g. `""`)
    - **Pattern** The RegExp used to match this token (`string`)
    - **Modifier** The modifier character used for the segment (e.g. `?`)
- **options**
  - **Sensitive** When `true` the regexp will be case sensitive. (default: `false`)
  - **Strict** When `true` the regexp allows an optional trailing delimiter to match. (default: `false`)
  - **End** When `true` the regexp will match to the end of the string. (default: `true`)
  - **Start** When `true` the regexp will match from the beginning of the string. (default: `true`)
  - **Validate** When `false` the function can produce an invalid (unmatched) path. (default: `true`)
  - **Delimiter** The default delimiter for segments, e.g. `[^/#?]` for `:named` patterns. (default: `'/#?'`)
  - **EndsWith** Optional character, or list of characters, to treat as "end" characters.
  - **prefixes** List of characters to automatically consider prefixes when parsing. (default: `./`)
  - **Encode** How to encode uri. (default: `func (uri string, token interface{}) string { return uri }`)
  - **Decode** How to decode uri. (default: `func (uri string, token interface{}) string { return uri }`)

```go
var tokens []pathToRegexp.Token
regexp := pathToRegexp.Must(pathToRegexp.PathToRegexp("/foo/:bar", &tokens, nil))
// regexp: ^\/foo(?:\/([^\/]+?))[\/]?(?=$)
// tokens: [{Name:"bar", Prefix:"/", Suffix:"", Pattern:"[^\\/]+?", Modifier:""}]
```

**Please note:** The `Regexp` returned by `path-to-regexp` is intended for ordered data (e.g. pathnames, hostnames). It can not handle arbitrarily ordered data (e.g. query strings, URL fragments, JSON, etc).

### Parameters

The path argument is used to define parameters and populate tokens.

#### Named Parameters

Named parameters are defined by prefixing a colon to the parameter name (`:foo`).

```go
var tokens []pathToRegexp.Token
regexp := pathToRegexp.Must(pathToRegexp.PathToRegexp("/:foo/:bar", &tokens, nil))
// tokens: [
//   {Name:"foo", Prefix:"/", Suffix:"", Pattern:"[^\\/]+?", Modifier:""},
//   {Name:"bar", Prefix:"/", Suffix:"", Pattern:"[^\\/]+?", Modifier:""}
// ]

match, _ := regexp.FindStringMatch("/test/route")
for _, g := range match.Groups() {
    fmt.Printf("%q ", g.String())
}
fmt.Printf("%d %q\n", match.Index, match)
//=> "/test/route" "test" "route" 0 "/test/route"
```

**Please note:** Parameter names must use "word characters" (`[A-Za-z0-9_]`).

##### Custom Matching Parameters

Parameters can have a custom regexp, which overrides the default match (`[^/]+`). For example, you can match digits or names in a path:

```go
var tokens []pathToRegexp.Token
regexpNumbers := pathToRegexp.Must(pathToRegexp.PathToRegexp("/icon-:foo(\\d+).png", &tokens, nil))
// tokens: [{Name:"foo", Prefix:"", Suffix:"", Pattern:"\\d+", Modifier:""}]

match, _ := regexpNumbers.FindStringMatch("/icon-123.png")
for _, g := range match.Groups() {
    fmt.Printf("%q ", g.String())
}
//=> "/icon-123.png" "123"

match, _ := regexpNumbers.FindStringMatch("/icon-abc.png")
fmt.Println(match)
//=> <nil>

tokens = make([]pathToRegexp.Token, 0)
regexpWord := pathToRegexp("/(user|u)", &tokens, nil)
// tokens: [{Name:0, Prefix:"/", Suffix:"", Pattern:"user|u", Modifier:""}]

match, _ = regexpWord.FindStringMatch("/u")
for _, g := range match.Groups() {
    fmt.Printf("%q ", g.String())
}
//=> "/u" "u"

match, _ = regexpWord.FindStringMatch("users")
fmt.Println(match)
//=> <nil>
```

**Tip:** Backslashes need to be escaped with another backslash in JavaScript strings.

##### Custom Prefix and Suffix

Parameters can be wrapped in `{}` to create custom prefixes or suffixes for your segment:

```go
regexp := pathToRegexp.Must(pathToRegexp.PathToRegexp("/:attr1?{-:attr2}?{-:attr3}?", nil, nil))

match, _ := regexp.FindStringMatch("/test")
for _, g := range match.Groups() {
    fmt.Printf("%q ", g.String())
}
//=> "/test" "test" "" ""

match, _ = regexp.FindStringMatch("/test-test")
for _, g := range match.Groups() {
    fmt.Printf("%q ", g.String())
}
//=> "/test-test" "test" "test" ""
```

#### Unnamed Parameters

It is possible to write an unnamed parameter that only consists of a regexp. It works the same the named parameter, except it will be numerically indexed:

```go
var tokens []pathToRegexp.Token
regexp := pathToRegexp.Must(pathToRegexp.PathToRegexp("/:foo/(.*)", &tokens, nil))
// tokens: [
//   {Name:"foo", Prefix:"/", Suffix:"", Pattern:"[^\\/]+?", Modifier:""}
//   {Name:0, Prefix:"/", Suffix:"", Pattern:".*", Modifier:""}
// ]

match, _ := regexp.FindStringMatch("/test/route")
for _, g := range match.Groups() {
    fmt.Printf("%q ", g.String())
}
fmt.Printf("%d %q\n", match.Index, match)
//=> "/test/route" "test" "route" 0 "/test/route"
```

#### Modifiers

Modifiers must be placed after the parameter (e.g. `/:foo?`, `/(test)?`, or `/:foo(test)?`).

##### Optional

Parameters can be suffixed with a question mark (`?`) to make the parameter optional.

```go
regexp := pathToRegexp.Must(pathToRegexp.PathToRegexp("/:foo/:bar?", nil, nil))
// tokens: [
//   {Name:"foo", Prefix:"/", Suffix:"", Pattern:"[^\\/]+?", Modifier:""}
//   {Name:"bar", Prefix:"/", Suffix:"", Pattern:"[^\\/]+?", Modifier:"?"}
// ]

match, err := regexp.FindStringMatch("/test")
for _, g := range match.Groups() {
    fmt.Printf("%q ", g.String())
}
fmt.Printf("%d %q\n", match.Index, match)
//=> "/test" "test" "" 0 "/test"

match, err = regexp.FindStringMatch("/test/route")
for _, g := range match.Groups() {
    fmt.Printf("%q ", g.String())
}
fmt.Printf("%d %q\n", match.Index, match)
//=> "/test/route" "test" "route" 0 "/test/route"
```

**Tip:** The prefix is also optional, escape the prefix `\/` to make it required.

##### Zero or more

Parameters can be suffixed with an asterisk (`*`) to denote a zero or more parameter matches.

```go
regexp := pathToRegexp.Must(pathToRegexp.PathToRegexp("/:foo*", nil, nil))
// tokens: [{Name:"foo", Prefix:"/", Suffix:"", Pattern:"[^\\/]+?", Modifier:"*"}]

match, err := regexp.FindStringMatch("/")
for _, g := range match.Groups() {
    fmt.Printf("%q ", g.String())
}
fmt.Printf("%d %q\n", match.Index, match)
//=> "/" "" 0 "/"

match, err = regexp.FindStringMatch("/bar/baz")
for _, g := range match.Groups() {
    fmt.Printf("%q ", g.String())
}
fmt.Printf("%d %q\n", match.Index, match)
//=> "/bar/baz" "bar/baz" 0 "/bar/baz"
```

##### One or more

Parameters can be suffixed with a plus sign (`+`) to denote a one or more parameter matches.

```go
regexp := pathToRegexp.Must(pathToRegexp.PathToRegexp("/:foo+", nil, nil))
// tokens: [{Name:"foo", Prefix:"/", Suffix:"", Pattern:"[^\\/]+?", Modifier:"+"}]

match, err := regexp.FindStringMatch("/")
fmt.Println(match)
//=> nil

match, err = regexp.FindStringMatch("/bar/baz")
for _, g := range match.Groups() {
    fmt.Printf("%q ", g.String())
}
fmt.Printf("%d %q\n", match.Index, match)
//=> "/bar/baz" "bar/baz" 0 "/bar/baz"
```

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
//=> pathToRegexp.Token{Name:"foo", Prefix:"/", Suffix:"", Pattern:"[^\\/]+?", Modifier:""}

fmt.Printf("%#v\n", tokens[2])
//=> pathToRegexp.Token{Name:0, Prefix:"/", Suffix:"", Pattern:".*", Modifier:""}
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

**Note:** The generated function will panic on invalid input.
