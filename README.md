# Path-to-RegExp

> Turn a path string such as `/user/:name` into a regular expression.

Thanks to [path-to-regexp](https://github.com/pillarjs/path-to-regexp).

## Usage

```go
import pathToRegexp "github/soongo/path-to-regexp"

// pathToRegexp.PathToRegexp(path, keys, options) // keys and options can be nil
// pathToRegexp.Parse(path, options) // options can be nil
// pathToRegexp.Compile(path, options) // options can be nil
```

- **path** A string, array or slice of strings, or a regular expression with type *github.com/dlclark/regexp2.Regexp.
- **keys** An array to populate with keys found in the path.
  - key
    - **name** The name of the token (`string` for named or `number` for index)
    - **prefix** The prefix character for the segment (e.g. `/`)
    - **delimiter** The delimiter for the segment (same as prefix or default delimiter)
    - **optional** Indicates the token is optional (`boolean`)
    - **repeat** Indicates the token is repeated (`boolean`)
    - **pattern** The RegExp used to match this token (`string`)
- **options**
  - **sensitive** When `true` the regexp will be case sensitive. (default: `false`)
  - **strict** When `true` the regexp allows an optional trailing delimiter to match. (default: `false`)
  - **end** When `true` the regexp will match to the end of the string. (default: `true`)
  - **start** When `true` the regexp will match from the beginning of the string. (default: `true`)
  - **delimiter** The default delimiter for segments. (default: `'/'`)
  - **endsWith** Optional character, or list of characters, to treat as "end" characters.
  - **whitelist** List of characters to consider delimiters when parsing. (default: `nil`, any character)

```go
var keys []pathToRegexp.Key
regexp, err := pathToRegexp.PathToRegexp("/foo/:bar", &keys, nil)
// regexp: ^\/foo\/([^\/]+?)(?:\/)?$
// keys: [{name:"bar", prefix:"/", delimiter:"/", optional:false, repeat:false, pattern:"[^\\/]+?"}}]
```

**Please note:** The `Regexp` returned by `path-to-regexp` is intended for ordered data (e.g. pathnames, hostnames). It can not handle arbitrarily ordered data (e.g. query strings, URL fragments, JSON, etc).

### Parameters

The path argument is used to define parameters and populate the list of keys.

#### Named Parameters

Named parameters are defined by prefixing a colon to the parameter name (`:foo`). By default, the parameter will match until the next prefix (e.g. `[^/]+`).

```go
regexp, err := pathToRegexp.PathToRegexp("/:foo/:bar", nil, nil)
// keys: [
//   {name:"foo", prefix:"/", delimiter:"/", optional:false, repeat:false, pattern:"[^\\/]+?"},
//   {name:"bar", prefix:"/", delimiter:"/", optional:false, repeat:false, pattern:"[^\\/]+?"}
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
regexp, err := pathToRegexp.PathToRegexp("/:foo/:bar?", nil, nil)
// keys: [
//   {name:"foo", prefix:"/", delimiter:"/", optional:false, repeat:false, pattern:"[^\\/]+?"},
//   {name:"bar", prefix:"/", delimiter:"/", optional:true, repeat:false, pattern:"[^\\/]+?"}
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
regexp, err := pathToRegexp.PathToRegexp("/:foo*", nil, nil)
// keys: [{name:"foo", prefix:"/", delimiter:"/", optional:true, repeat:true, pattern:"[^\\/]+?"}]

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
regexp, err := pathToRegexp.PathToRegexp("/:foo+", nil, nil)
// keys: [{name:"foo", prefix:"/", delimiter:"/", optional:false, repeat:true, pattern:"[^\\/]+?"}]

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
regexp, err := pathToRegexp.PathToRegexp("/:foo/(.*)", nil, nil)
// keys: [
//   {name:"foo", prefix:"/", delimiter:"/", optional:false, repeat:false, pattern:"[^\\/]+?"},
//   {name:0, prefix:"/", delimiter:"/", optional:false, repeat:false, pattern:".*"}
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
regexpNumbers, err := pathToRegexp.PathToRegexp("/icon-:foo(\\d+).png", nil, nil)
// keys: {name:"foo", prefix:"-", delimiter:"-", optional:false, repeat:false, pattern:"\\d+"}

match, err := regexpNumbers.FindStringMatch("/icon-123.png")
for _, g := range match.Groups() {
    fmt.Printf("%q ", g.String())
}
//=> "/icon-123.png" "123"

match, err = regexpNumbers.FindStringMatch("/icon-abc.png")
fmt.Println(match)
//=> nil

regexpWord, err := pathToRegexp.PathToRegexp("/(user|u)", nil, nil)
// keys: {name:0, prefix:"/", delimiter:"/", optional:false, repeat:false, pattern:"user|u"}

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

### Parse

The parse function is exposed via `pathToRegexp.Parse`. This will return a slice of strings and keys.

```go
tokens := pathToRegexp.Parse("/route/:foo/(.*)", nil)

fmt.Printf("%#v\n", tokens[0])
//=> "/route"

fmt.Printf("%#v\n", tokens[1])
//=> pathToRegexp.Key{name:"foo", prefix:"/", delimiter:"/", optional:false, repeat:false, pattern:"[^\\/]+?"}

fmt.Printf("%#v\n", tokens[2])
//=> pathToRegexp.Key{name:0, prefix:"/", delimiter:"/", optional:false, repeat:false, pattern:".*"}
```

**Note:** This method only works with strings.

### Compile ("Reverse" Path-To-RegExp)

Path-To-RegExp exposes a compile function for transforming a string into a valid path.

```go
toPath, err := pathToRegexp.Compile("/user/:id", nil)

toPath(map[string]int{"id": 123}, nil) //=> "/user/123"
toPath(map[string]string{"id": "café"}, nil) //=> "/user/caf%C3%A9"
toPath(map[string]string{"id": "/"}, nil) //=> "/user/%2F"

toPath(map[string]string{"id": ":/"}, nil) //=> "/user/%3A%2F"
toPath(map[string]string{"id": ":&"}, &Options{encode: func(value string, token interface{}) string {
    return value
}}) //=> panic
toPath(map[string]string{"id": ":&"}, nil) //=> "/user/%3A%26"
toPath(map[string]string{"id": ":&"}, &Options{encode: func(value string, token interface{}) string {
    return value
}}) //=> /user/:&

toPathRepeated, err := pathToRegexp.Compile("/:segment+", nil)

toPathRepeated(map[string]string{"segment": "foo"}, nil) //=> "/foo"
toPathRepeated(map[string][]string{"segment": {"a", "b", "c"}}, nil) //=> "/a/b/c"

toPathRegexp, err := pathToRegexp.Compile("/user/:id(\\d+)", nil)

toPathRegexp(map[string]int{"id": 123}, nil) //=> "/user/123"
toPathRegexp(map[string]string{"id": "123"}, nil) //=> "/user/123"
toPathRegexp(map[string]string{"id": "abc"}, nil) //=> panic
t1 := true
toPathRegexp(map[string]string{"id": "abc"}, &Options{validate: &t1}) //=> panic
```

**Note:** The generated function will panic on invalid input. It will do all necessary checks to ensure the generated path is valid. This method only works with strings.
