// Copyright 2019 Guoyao Wu. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package pathtoregexp

import (
	"errors"
	"fmt"
	"net/url"
	"reflect"
	"strconv"
	"strings"
	"unsafe"

	"github.com/dlclark/regexp2"
)

// Token is parsed from path. For example, using `/user/:id`, `tokens` will
// contain `[{name:'id', delimiter:'/', optional:false, repeat:false}]`
type Token struct {
	// The name of the token (string for named or number for index)
	name interface{}

	// The prefix character for the segment (e.g. /)
	prefix string

	// The delimiter for the segment (same as prefix or default delimiter)
	delimiter string

	// Indicates the token is optional (boolean)
	optional bool

	// Indicates the token is repeated (boolean)
	repeat bool

	// The RegExp used to match this token (string)
	pattern string
}

// Options contains some optional configs
type Options struct {
	// When true the regexp will be case sensitive. (default: false)
	sensitive bool

	// When true the regexp allows an optional trailing delimiter to match. (default: false)
	strict bool

	// When true the regexp will match to the end of the string. (default: true)
	end *bool

	// When true the regexp will match from the beginning of the string. (default: true)
	start *bool

	validate *bool

	// The default delimiter for segments. (default: '/')
	delimiter string

	// Optional character, or list of characters, to treat as "end" characters.
	endsWith interface{}

	// List of characters to consider delimiters when parsing. (default: nil, any character)
	whitelist []string

	// how to encode uri
	encode func(uri string, token interface{}) string
}

// defaultDelimiter is the default delimiter of path.
const defaultDelimiter = "/"

// pathRegexp is the main path matching regexp utility.
var pathRegexp = regexp2.MustCompile(strings.Join([]string{
	"(\\\\.)",
	"(?:\\:(\\w+)(?:\\(((?:\\\\.|[^\\\\()])+)\\))?|\\(((?:\\\\.|[^\\\\()])+)\\))([+*?])?",
}, "|"), regexp2.None)

var escapeRegexp = regexp2.MustCompile("([.+*?=^!:${}()[\\]|/\\\\])", regexp2.None)
var escapeGroupRegexp = regexp2.MustCompile("([=!:$/()])", regexp2.None)
var tokenRegexp = regexp2.MustCompile("\\((?!\\?)", regexp2.None)

// Parse a string for the raw tokens.
func Parse(str string, o *Options) []interface{} {
	tokens, tokenIndex, index, path, pathEscaped := make([]interface{}, 0), 0, 0, "", false
	if o == nil {
		o = &Options{}
	}
	defaultDelimiter := orString(o.delimiter, defaultDelimiter)
	whitelist := o.whitelist

	for matcher, _ := pathRegexp.FindStringMatch(str); matcher != nil; matcher,
		_ = pathRegexp.FindNextMatch(matcher) {
		groups := matcher.Groups()
		m := groups[0].String()
		escaped := groups[1].String()
		offset := matcher.Index
		path += str[index:offset]
		index = offset + len(m)

		// Ignore already escaped sequences.
		if escaped != "" {
			path += escaped[1:2]
			pathEscaped = true
			continue
		}

		prev, name, capture, group, modifier := "", groups[2].String(),
			groups[3].String(), groups[4].String(), groups[5].String()

		if !pathEscaped && len(path) > 0 {
			k := len(path) - 1
			c := path[k : k+1]
			matches := true
			if whitelist != nil {
				matches = indexOf(whitelist, c) > -1
			}

			if matches {
				prev = c
				path = path[0:k]
			}
		}

		// Push the current path onto the tokens.
		if path != "" {
			tokens = append(tokens, path)
			path = ""
			pathEscaped = false
		}

		repeat := modifier == "+" || modifier == "*"
		optional := modifier == "?" || modifier == "*"
		pattern := orString(capture, group)
		delimiter := orString(prev, defaultDelimiter)

		var tokenName interface{} = name
		if name == "" {
			tokenName = tokenIndex
			tokenIndex++
		}
		if pattern != "" {
			pattern = escapeGroup(pattern)
		} else {
			d := delimiter + defaultDelimiter
			if delimiter == defaultDelimiter {
				d = delimiter
			}
			pattern = "[^" + escapeString(d) + "]+?"
		}
		tokens = append(tokens, Token{
			name:      tokenName,
			prefix:    prev,
			delimiter: delimiter,
			optional:  optional,
			repeat:    repeat,
			pattern:   pattern,
		})
	}

	// Push any remaining characters.
	if path != "" || index < len(str) {
		tokens = append(tokens, path+str[index:])
	}

	return tokens
}

// Compile a string to a template function for the path.
func Compile(str string, o *Options) (func(interface{}, *Options) string, error) {
	return tokensToFunction(Parse(str, o), o)
}

// MustCompile is like Compile but panics if the expression cannot be compiled.
// It simplifies safe initialization of global variables.
func MustCompile(str string, o *Options) func(interface{}, *Options) string {
	f, err := tokensToFunction(Parse(str, o), o)
	if err != nil {
		panic(`pathtoregexp: Compile(` + quote(str) + `): ` + err.Error())
	}
	return f
}

// Expose a method for transforming tokens into the path function.
func tokensToFunction(tokens []interface{}, o *Options) (
	func(interface{}, *Options) string, error) {
	// Compile all the tokens into regexps.
	matches := make([]*regexp2.Regexp, len(tokens))

	// Compile all the patterns before compilation.
	for i, token := range tokens {
		if token, ok := token.(Token); ok {
			m, err := regexp2.Compile("^(?:"+token.pattern+")$", flags(o))
			if err != nil {
				return nil, err
			}
			matches[i] = m
		}
	}

	return func(data interface{}, o *Options) string {
		t := true
		path, validate, encode := "", &t, encodeURIComponent
		if o != nil {
			if o.encode != nil {
				encode = o.encode
			}
			if o.validate != nil {
				validate = o.validate
			}
		}

		for i, token := range tokens {
			if token, ok := token.(string); ok {
				path += token
				continue
			}

			if token, ok := token.(Token); ok {
				if data != nil && reflect.TypeOf(data).Kind() == reflect.Map {
					data := toMap(data)
					value := data[token.name]
					if value == nil {
						if intValue, ok := token.name.(int); ok {
							value = data[strconv.Itoa(intValue)]
						}
					}

					var segment string

					if value != nil {
						if k := reflect.TypeOf(value).Kind(); k == reflect.Slice || k == reflect.Array {
							value := toSlice(value)
							if !token.repeat {
								panic(fmt.Sprintf("Expected \"%v\" to not repeat, but got array",
									token.name))
							}

							if len(value) == 0 {
								if token.optional {
									continue
								}
								panic(fmt.Sprintf("Expected \"%v\" to not be empty", token.name))
							}

							for j, v := range value {
								segment = encode(fmt.Sprintf("%v", v), token)

								if *validate {
									if ok, err := matches[i].MatchString(segment); err != nil || !ok {
										panic(fmt.Sprintf("Expected all \"%v\" to match \"%v\"",
											token.name, token.pattern))
									}
								}

								if j == 0 {
									path += token.prefix
								} else {
									path += token.delimiter
								}
								path += segment
							}

							continue
						}
					}

					vString, isString := value.(string)
					vInt, isInt := value.(int)
					vBool, isBool := value.(bool)
					if isString || isInt || isBool {
						var v string
						if isString {
							v = vString
						} else if isInt {
							v = strconv.Itoa(vInt)
						} else if isBool {
							v = strconv.FormatBool(vBool)
						}
						segment = encode(v, token)

						if *validate {
							if ok, err := matches[i].MatchString(segment); err != nil || !ok {
								panic(fmt.Sprintf("Expected \"%v\" to match \"%v\", but got \"%v\"",
									token.name, token.pattern, segment))
							}
						}

						path += token.prefix + segment
						continue
					}
				}

				if token.optional {
					continue
				}

				s := "a string"
				if token.repeat {
					s = "an array"
				}
				panic(fmt.Sprintf("Expected \"%v\" to be %v", token.name, s))
			}
		}

		return path
	}, nil
}

func orString(str ...string) string {
	for _, v := range str {
		if v != "" {
			return v
		}
	}
	return ""
}

func indexOf(in interface{}, elem interface{}) int {
	inValue := reflect.ValueOf(in)
	elemValue := reflect.ValueOf(elem)
	inType := inValue.Type()

	if inType.Kind() == reflect.String {
		return strings.Index(inValue.String(), elemValue.String())
	}

	if inType.Kind() == reflect.Slice {
		for i := 0; i < inValue.Len(); i++ {
			if reflect.DeepEqual(inValue.Index(i).Interface(), elem) {
				return i
			}
		}
	}

	return -1
}

func toSlice(data interface{}) []interface{} {
	v := reflect.ValueOf(data)
	length := v.Len()
	arr := make([]interface{}, length, v.Cap())
	for i := 0; i < length; i++ {
		arr[i] = v.Index(i).Interface()
	}
	return arr
}

func toMap(data interface{}) map[interface{}]interface{} {
	v, m := reflect.ValueOf(data), make(map[interface{}]interface{})
	for _, k := range v.MapKeys() {
		value := v.MapIndex(k)
		m[k.Interface()] = value.Interface()
	}
	return m
}

func encodeURIComponent(str string, token interface{}) string {
	r := url.QueryEscape(str)
	r = strings.Replace(r, "+", "%20", -1)
	return r
}

// Escape a regular expression string.
func escapeString(str string) string {
	str, err := escapeRegexp.Replace(str, "\\$1", -1, -1)
	if err != nil {
		panic(err)
	}
	return str
}

// Escape the capturing group by escaping special characters and meaning.
func escapeGroup(group string) string {
	group, err := escapeGroupRegexp.Replace(group, "\\$1", -1, -1)
	if err != nil {
		panic(err)
	}
	return group
}

func quote(s string) string {
	if strconv.CanBackquote(s) {
		return "`" + s + "`"
	}
	return strconv.Quote(s)
}

// Get the flags for a regexp from the options.
func flags(o *Options) regexp2.RegexOptions {
	if o != nil && o.sensitive {
		return regexp2.None
	}
	return regexp2.IgnoreCase
}

// Must is a helper that wraps a call to a function returning (*regexp2.Regexp, error)
// and panics if the error is non-nil. It is intended for use in variable initializations
// such as
//	var r = pathtoregexp.Must(pathtoregexp.PathToRegexp("/", nil, nil))
func Must(r *regexp2.Regexp, err error) *regexp2.Regexp {
	if err != nil {
		panic(err)
	}
	return r
}

// Pull out tokens from a regexp.
func regexpToRegexp(path *regexp2.Regexp, tokens *[]Token) *regexp2.Regexp {
	if tokens != nil {
		m, _ := tokenRegexp.FindStringMatch(path.String())
		if m != nil && m.GroupCount() > 0 {
			newTokens := make([]Token, 0, len(*tokens)+m.GroupCount())
			newTokens = append(newTokens, *tokens...)
			for i := 0; i < m.GroupCount(); i++ {
				newTokens = append(newTokens, Token{
					name:      i,
					prefix:    "",
					delimiter: "",
					optional:  false,
					repeat:    false,
					pattern:   "",
				})
			}
			hdr := (*reflect.SliceHeader)(unsafe.Pointer(tokens))
			*hdr = *(*reflect.SliceHeader)(unsafe.Pointer(&newTokens))
		}
	}

	return path
}

// Transform an array into a regexp.
func arrayToRegexp(path []interface{}, tokens *[]Token, o *Options) (*regexp2.Regexp, error) {
	var parts []string

	for i := 0; i < len(path); i++ {
		r, err := PathToRegexp(path[i], tokens, o)
		if err != nil {
			return nil, err
		}
		parts = append(parts, r.String())
	}

	return regexp2.Compile("(?:"+strings.Join(parts, "|")+")", flags(o))
}

// Create a path regexp from string input.
func stringToRegexp(path string, tokens *[]Token, o *Options) (*regexp2.Regexp, error) {
	return tokensToRegExp(Parse(path, o), tokens, o)
}

// Expose a function for taking tokens and returning a RegExp.
func tokensToRegExp(rawTokens []interface{}, tokens *[]Token, o *Options) (*regexp2.Regexp, error) {
	if o == nil {
		o = &Options{}
	}

	strict, start, end, route := o.strict, true, true, ""
	if o.start != nil {
		start = *o.start
	}
	if o.end != nil {
		end = *o.end
	}

	var ends []string
	if o.endsWith != nil {
		if str, ok := o.endsWith.(string); ok {
			ends = []string{str}
		} else if arr, ok := o.endsWith.([]string); ok {
			ends = arr
		} else {
			return nil, errors.New("endsWith should be string or []string")
		}
	}

	delimiter := orString(o.delimiter, defaultDelimiter)
	arr := make([]string, len(ends)+1)
	for i, v := range ends {
		v = escapeString(v)
		arr[i] = v
	}
	arr[len(ends)] = "$"
	endsWith := strings.Join(arr, "|")

	if start {
		route = "^"
	}

	var newTokens []Token
	if tokens != nil {
		newTokens = make([]Token, 0, len(*tokens)+len(rawTokens))
		newTokens = append(newTokens, *tokens...)
	}

	// Iterate over the tokens and create our regexp string.
	for _, token := range rawTokens {
		if str, ok := token.(string); ok {
			route += escapeString(str)
		} else if token, ok := token.(Token); ok {
			capture := token.pattern
			if token.repeat {
				capture = "(?:" + token.pattern + ")(?:" + escapeString(token.delimiter) +
					"(?:" + token.pattern + "))*"
			}

			if tokens != nil {
				newTokens = append(newTokens, token)
			}

			if token.optional {
				if token.prefix == "" {
					route += "(" + capture + ")?"
				} else {
					route += "(?:" + escapeString(token.prefix) + "(" + capture + "))?"
				}
			} else {
				route += escapeString(token.prefix) + "(" + capture + ")"
			}
		}
	}

	if tokens != nil {
		hdr := (*reflect.SliceHeader)(unsafe.Pointer(tokens))
		*hdr = *(*reflect.SliceHeader)(unsafe.Pointer(&newTokens))
	}

	if end {
		if !strict {
			route += "(?:" + escapeString(delimiter) + ")?"
		}

		s := "(?=" + endsWith + ")"
		if endsWith == "$" {
			s = "$"
		}
		route += s
	} else {
		isEndDelimited := false
		if len(rawTokens) == 0 {
			isEndDelimited = true
		} else {
			endToken := rawTokens[len(rawTokens)-1]
			if endToken == nil {
				isEndDelimited = true
			} else if str, ok := endToken.(string); ok {
				isEndDelimited = str[len(str)-1:] == delimiter
			}
		}

		if !strict {
			route += "(?:" + escapeString(delimiter) + "(?=" + endsWith + "))?"
		}
		if !isEndDelimited {
			route += "(?=" + escapeString(delimiter) + "|" + endsWith + ")"
		}
	}

	return regexp2.Compile(route, flags(o))
}

// PathToRegexp normalizes the given path string, returning a regular expression.
// An empty array can be passed in for the tokens, which will hold the
// placeholder token descriptions. For example, using `/user/:id`, `tokens` will
// contain `[{name: 'id', delimiter: '/', optional: false, repeat: false}]`.
func PathToRegexp(path interface{}, tokens *[]Token, options *Options) (*regexp2.Regexp, error) {
	switch path := path.(type) {
	case *regexp2.Regexp:
		return regexpToRegexp(path, tokens), nil
	case string:
		return stringToRegexp(path, tokens, options)
	}

	switch reflect.TypeOf(path).Kind() {
	case reflect.Slice, reflect.Array:
		return arrayToRegexp(toSlice(path), tokens, options)
	}

	return nil, errors.New(`path should be string, array or slice of strings, 
or a regular expression with type *github.com/dlclark/regexp2.Regexp`)
}
