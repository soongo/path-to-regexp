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
	"unicode"
	"unsafe"

	"golang.org/x/text/runes"

	"github.com/dlclark/regexp2"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

// Token is parsed from path. For example, using `/user/:id`, `tokens` will
// contain `[{Name:'id', Delimiter:'/', Optional:false, Repeat:false}]`
type Token struct {
	// The name of the token (string for named or number for index)
	Name interface{}

	// The prefix character for the segment (e.g. /)
	Prefix string

	// The delimiter for the segment (same as prefix or default delimiter)
	Delimiter string

	// Indicates the token is optional (boolean)
	Optional bool

	// Indicates the token is repeated (boolean)
	Repeat bool

	// The RegExp used to match this token (string)
	Pattern string
}

// Options contains some optional configs
type Options struct {
	// When true the regexp will be case sensitive. (default: false)
	Sensitive bool

	// When true the regexp allows an optional trailing delimiter to match. (default: false)
	Strict bool

	// When true the regexp will match to the end of the string. (default: true)
	End *bool

	// When true the regexp will match from the beginning of the string. (default: true)
	Start *bool

	Validate *bool

	// The default delimiter for segments. (default: '/')
	Delimiter string

	// Optional character, or list of characters, to treat as "end" characters.
	EndsWith interface{}

	// List of characters to consider delimiters when parsing. (default: nil, any character)
	Whitelist []string

	// how to encode uri
	Encode func(uri string, token interface{}) string

	// how to decode uri
	Decode func(str string, token interface{}) string
}

// MatchResult contains the result of match function
type MatchResult struct {
	// matched url path
	Path string

	// matched start index
	Index int

	// matched params in url
	Params map[interface{}]interface{}
}

// defaultDelimiter is the default delimiter of path.
const defaultDelimiter = "/"

var escapeRegexp = regexp2.MustCompile("([.+*?=^!:${}()[\\]|/\\\\])", regexp2.None)
var tokenRegexp = regexp2.MustCompile("\\((?!\\?)", regexp2.None)

// EncodeURIComponent encodes a text string as a valid component of a Uniform
// Resource Identifier (URI).
func EncodeURIComponent(str string) string {
	r := url.QueryEscape(str)
	r = strings.Replace(r, "+", "%20", -1)
	return r
}

// Gets the unencoded version of an encoded component of a Uniform Resource
// Identifier (URI).
func DecodeURIComponent(str string) string {
	r, err := url.QueryUnescape(str)
	if err != nil {
		panic(err)
	}
	return r
}

// Encodes a text string as a valid Uniform Resource Identifier (URI)
func encodeURI(str string) string {
	excludes := ";/?:@&=+$,#"
	arr := strings.Split(str, "")
	result := ""
	for _, v := range arr {
		if strings.Contains(excludes, v) {
			result += v
		} else {
			result += EncodeURIComponent(v)
		}
	}
	return result
}

// Gets the unencoded version of an encoded Uniform Resource Identifier (URI).
func decodeURI(str string) string {
	magicWords := "1@X#y!Z" // not a good idea
	excludes := []string{"%3B", "%2F", "%3F", "%3A", "%40", "%26", "%3D", "%2B", "%24", "%2C", "%23"}
	r := regexp2.MustCompile(strings.Join(excludes, "|"), regexp2.None)

	str, _ = r.ReplaceFunc(str, func(m regexp2.Match) string {
		return strings.Replace(m.String(), "%", magicWords, -1)
	}, -1, -1)

	str = decodeURIComponent(str, nil)

	for i, v := range excludes {
		excludes[i] = magicWords + strings.TrimPrefix(v, "%")
	}
	r = regexp2.MustCompile(strings.Join(excludes, "|"), regexp2.None)

	str, _ = r.ReplaceFunc(str, func(m regexp2.Match) string {
		return strings.Replace(m.String(), magicWords, "%", -1)
	}, -1, -1)

	return str
}

// Returns the String value result of normalizing the string into the normalization form
// named by form as specified in Unicode Standard Annex #15, Unicode Normalization Forms.
// param form Applicable values: "NFC", "NFD", "NFKC", or "NFKD", If not specified default
// is "NFC"
func normalize(str string, form ...norm.Form) string {
	f := norm.NFC
	if len(form) > 0 {
		f = form[0]
	}
	t := transform.Chain(f, runes.Remove(runes.In(unicode.Mn)), f)
	normStr, _, _ := transform.String(t, str)
	return normStr
}

// NormalizePathname normalizes a pathname for matching, replaces multiple slashes
// with a single slash and normalizes unicode characters to "NFC". When using this method,
// `decode` should be an identity function so you don't decode strings twice.
func NormalizePathname(pathname string) string {
	pathname = decodeURI(pathname)
	r := regexp2.MustCompile("\\/+", regexp2.None)
	pathname, err := r.Replace(pathname, "/", -1, -1)
	if err != nil {
		panic(err)
	}
	return pathname
}

// Balanced bracket helper function.
func balanced(open string, close string, str string, index int) int {
	count, i, arr := 0, index, strings.Split(str, "")

	for i < len(arr) {
		if arr[i] == "\\" {
			i += 2
			continue
		}

		if arr[i] == close {
			count--

			if count == 0 {
				return i + 1
			}
		}

		if arr[i] == open {
			count++
		}

		i++
	}

	return -1
}

// Parse a string for the raw tokens.
func Parse(str string, o *Options) []interface{} {
	tokens, tokenIndex, index, path, isEscaped := make([]interface{}, 0), 0, 0, "", false
	if o == nil {
		o = &Options{}
	}
	defaultDelimiter := anyString(o.Delimiter, defaultDelimiter)
	whitelist := o.Whitelist

	// use list to deal with unicode in str
	arr := strings.Split(str, "")

	// Ignore escaped sequences.
	length := len(arr)
	for index < length {
		prefix, name, pattern := "", "", ""

		if arr[index] == "\\" {
			index++
			path += arr[index]
			index++
			isEscaped = true
			continue
		}

		if arr[index] == ":" {
			for index++; index < length; index++ {
				if len(arr[index]) == 1 {
					code := arr[index][0]
					isNumber := code >= 48 && code <= 57 // `0-9`
					isUpper := code >= 65 && code <= 90  // `A-Z`
					isLower := code >= 97 && code <= 122 // `a-z`
					isUnderscore := code == 95           // `_`
					if isNumber || isUpper || isLower || isUnderscore {
						name += string(code)
						continue
					}
				}
				break
			}

			// False positive on param name.
			if name == "" {
				index--
			}
		}

		if index < length && arr[index] == "(" {
			end := balanced("(", ")", str, index)

			// False positive on matching brackets.
			if end > -1 {
				pattern = strings.Join(arr[index+1:end-1], "")
				index = end
				if pattern[0] == '?' {
					panic("Path pattern must be a capturing group")
				}

				r := regexp2.MustCompile("\\((?=[^?])", regexp2.None)
				if ok, _ := r.MatchString(pattern); ok {
					validPattern, err := r.Replace(pattern, "(?:", -1, -1)
					if err != nil {
						panic(err)
					}

					panic(fmt.Sprintf("Capturing groups are not allowed in pattern, "+
						"use a non-capturing group: (%s)", validPattern))
				}
			}
		}

		// Add regular characters to the path string.
		if name == "" && pattern == "" {
			path += string(arr[index])
			index++
			isEscaped = false
			continue
		}

		// Extract the final character from `path` for the prefix.
		if len(path) > 0 && !isEscaped {
			s := string(path[len(path)-1])
			matches := true
			if whitelist != nil {
				matches = stringIndexOf(whitelist, s) > -1
			}
			if matches {
				prefix = s
				path = path[0 : len(path)-1]
			}
		}

		// Push the current path onto the list of tokens.
		if len(path) > 0 {
			tokens = append(tokens, path)
			path = ""
		}

		repeat := index < length && (arr[index] == "+" || arr[index] == "*")
		optional := index < length && (arr[index] == "?" || arr[index] == "*")
		delimiter := anyString(prefix, defaultDelimiter)

		// Increment `i` past modifier token.
		if repeat || optional {
			index++
		}

		var tokenName interface{} = name
		if name == "" {
			tokenName = tokenIndex
			tokenIndex++
		}

		p := pattern
		if pattern == "" {
			d := delimiter + defaultDelimiter
			if delimiter == defaultDelimiter {
				d = delimiter
			}
			p = "[^" + escapeString(d) + "]+?"
		}
		tokens = append(tokens, Token{
			Name:      tokenName,
			Prefix:    prefix,
			Delimiter: delimiter,
			Optional:  optional,
			Repeat:    repeat,
			Pattern:   p,
		})
	}

	if len(path) > 0 {
		tokens = append(tokens, path)
	}

	return tokens
}

// Compile a string to a template function for the path.
func Compile(str string, o *Options) (func(interface{}) string, error) {
	return tokensToFunction(Parse(str, o), o)
}

// MustCompile is like Compile but panics if the expression cannot be compiled.
// It simplifies safe initialization of global variables.
func MustCompile(str string, o *Options) func(interface{}) string {
	f, err := Compile(str, o)
	if err != nil {
		panic(`pathtoregexp: Compile(` + quote(str) + `): ` + err.Error())
	}
	return f
}

// Match creates path match function from `path-to-regexp` spec.
func Match(path interface{}, o *Options) (func(string) *MatchResult, error) {
	var tokens []Token
	re, err := PathToRegexp(path, &tokens, o)
	if err != nil {
		return nil, err
	}

	return regexpToFunction(re, tokens, o), nil
}

// MustMatch is like Match but panics if err occur in match function.
func MustMatch(path interface{}, o *Options) func(string) *MatchResult {
	f, err := Match(path, o)
	if err != nil {
		panic(err)
	}
	return f
}

// Create a path match function from `path-to-regexp` output.
func regexpToFunction(re *regexp2.Regexp, tokens []Token, o *Options) func(string) *MatchResult {
	decode := func(str string, token interface{}) string {
		return str
	}
	if o != nil && o.Decode != nil {
		decode = o.Decode
	}

	return func(pathname string) *MatchResult {
		m, err := re.FindStringMatch(pathname)
		if m == nil || m.GroupCount() == 0 || err != nil {
			return nil
		}

		path := m.Groups()[0].String()
		index := m.Index
		params := make(map[interface{}]interface{})

		for i := 1; i < m.GroupCount(); i++ {
			group := m.Groups()[i]
			if len(group.Captures) == 0 {
				continue
			}

			token := tokens[i-1]
			matchedStr := group.String()

			if token.Repeat {
				arr := strings.Split(matchedStr, token.Delimiter)
				length := len(arr)
				if length > 0 {
					for i, str := range arr {
						arr[i] = decode(str, token)
					}
					params[token.Name] = arr
				}
			} else {
				params[token.Name] = decode(matchedStr, token)
			}
		}

		return &MatchResult{Path: path, Index: index, Params: params}
	}
}

// Expose a method for transforming tokens into the path function.
func tokensToFunction(tokens []interface{}, o *Options) (
	func(interface{}) string, error) {
	if o == nil {
		o = &Options{}
	}
	reFlags := flags(o)
	encode, validate := encodeURIComponent, true
	if o.Encode != nil {
		encode = o.Encode
	}
	if o.Validate != nil {
		validate = *o.Validate
	}

	// Compile all the tokens into regexps.
	matches := make([]*regexp2.Regexp, len(tokens))
	for i, token := range tokens {
		if token, ok := token.(Token); ok {
			m, err := regexp2.Compile("^(?:"+token.Pattern+")$", reFlags)
			if err != nil {
				return nil, err
			}
			matches[i] = m
		}
	}

	return func(data interface{}) string {
		path := ""

		for i, token := range tokens {
			if token, ok := token.(string); ok {
				path += token
				continue
			}

			if token, ok := token.(Token); ok {
				if data != nil && reflect.TypeOf(data).Kind() == reflect.Map {
					data := toMap(data)
					value := data[token.Name]
					if value == nil {
						if intValue, ok := token.Name.(int); ok {
							value = data[strconv.Itoa(intValue)]
						}
					}

					if value != nil {
						if k := reflect.TypeOf(value).Kind(); k == reflect.Slice || k == reflect.Array {
							value := toSlice(value)
							if !token.Repeat {
								panic(fmt.Sprintf("Expected \"%v\" to not repeat, but got array",
									token.Name))
							}

							if len(value) == 0 {
								if token.Optional {
									continue
								}
								panic(fmt.Sprintf("Expected \"%v\" to not be empty", token.Name))
							}

							for j, v := range value {
								segment := encode(fmt.Sprintf("%v", v), token)

								if validate {
									if ok, err := matches[i].MatchString(segment); err != nil || !ok {
										panic(fmt.Sprintf("Expected all \"%v\" to match \"%v\"",
											token.Name, token.Pattern))
									}
								}

								if j == 0 {
									path += token.Prefix
								} else {
									path += token.Delimiter
								}
								path += segment
							}

							continue
						}
					}

					vString, isString := value.(string)
					vInt, isInt := value.(int)
					vFloat, isFloat := value.(float64)
					if isString || isInt || isFloat {
						var v string
						if isString {
							v = vString
						} else if isInt {
							v = strconv.Itoa(vInt)
						} else if isFloat {
							v = strconv.FormatFloat(vFloat, 'f', -1, 64)
						}
						segment := encode(v, token)

						if validate {
							if ok, err := matches[i].MatchString(segment); err != nil || !ok {
								panic(fmt.Sprintf("Expected \"%v\" to match \"%v\", but got \"%v\"",
									token.Name, token.Pattern, segment))
							}
						}

						path += token.Prefix + segment
						continue
					}
				}

				if token.Optional {
					continue
				}

				s := "a string"
				if token.Repeat {
					s = "an array"
				}
				panic(fmt.Sprintf("Expected \"%v\" to be %v", token.Name, s))
			}
		}

		return path
	}, nil
}

// Returns the first non empty string
func anyString(str ...string) string {
	for _, v := range str {
		if v != "" {
			return v
		}
	}
	return ""
}

// Returns the index of str in string slice
func stringIndexOf(arr []string, str string) int {
	for i, v := range arr {
		if v == str {
			return i
		}
	}
	return -1
}

// Transform data which is reflect.Slice, reflect.Array to slice
func toSlice(data interface{}) []interface{} {
	v := reflect.ValueOf(data)
	length := v.Len()
	arr := make([]interface{}, length, v.Cap())
	for i := 0; i < length; i++ {
		arr[i] = v.Index(i).Interface()
	}
	return arr
}

// Transform data which is reflect.Map to map
func toMap(data interface{}) map[interface{}]interface{} {
	v, m := reflect.ValueOf(data), make(map[interface{}]interface{})
	for _, k := range v.MapKeys() {
		value := v.MapIndex(k)
		m[k.Interface()] = value.Interface()
	}
	return m
}

func encodeURIComponent(str string, token interface{}) string {
	return EncodeURIComponent(str)
}

func decodeURIComponent(str string, token interface{}) string {
	return DecodeURIComponent(str)
}

// Escape a regular expression string.
func escapeString(str string) string {
	str, err := escapeRegexp.Replace(str, "\\$1", -1, -1)
	if err != nil {
		panic(err)
	}
	return str
}

func quote(s string) string {
	if strconv.CanBackquote(s) {
		return "`" + s + "`"
	}
	return strconv.Quote(s)
}

// Get the flags for a regexp from the options.
func flags(o *Options) regexp2.RegexOptions {
	if o != nil && o.Sensitive {
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
		totalGroupCount := 0
		for m, _ := tokenRegexp.FindStringMatch(path.String()); m != nil; m,
			_ = tokenRegexp.FindNextMatch(m) {
			totalGroupCount += m.GroupCount()
		}

		if totalGroupCount > 0 {
			newTokens := append(make([]Token, 0), *tokens...)

			for i := 0; i < totalGroupCount; i++ {
				newTokens = append(newTokens, Token{
					Name:      i,
					Prefix:    "",
					Delimiter: "",
					Optional:  false,
					Repeat:    false,
					Pattern:   "",
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

	strict, start, end, route := o.Strict, true, true, ""
	if o.Start != nil {
		start = *o.Start
	}
	if o.End != nil {
		end = *o.End
	}

	var ends []string
	if o.EndsWith != nil {
		if str, ok := o.EndsWith.(string); ok {
			ends = []string{str}
		} else if arr, ok := o.EndsWith.([]string); ok {
			ends = arr
		} else {
			return nil, errors.New("endsWith should be string or []string")
		}
	}

	delimiter := anyString(o.Delimiter, defaultDelimiter)
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
			capture := token.Pattern
			if token.Repeat {
				capture = "(?:" + token.Pattern + ")(?:" + escapeString(token.Delimiter) +
					"(?:" + token.Pattern + "))*"
			}

			if tokens != nil {
				newTokens = append(newTokens, token)
			}

			if token.Optional {
				if token.Prefix == "" {
					route += "(" + capture + ")?"
				} else {
					route += "(?:" + escapeString(token.Prefix) + "(" + capture + "))?"
				}
			} else {
				route += escapeString(token.Prefix) + "(" + capture + ")"
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
// contain `[{Name: 'id', Delimiter: '/', Optional: false, Repeat: false}]`.
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
