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

	"github.com/dlclark/regexp2"
)

// Token is parsed from path. For example, using `/user/:id`, `tokens` will
// contain `[{Name:'id', Delimiter:'/', Optional:false, Repeat:false}]`
type Token struct {
	// The name of the token (string for named or number for index)
	Name interface{}

	// The prefix character for the segment (e.g. /)
	Prefix string

	// The suffix string for the segment (e.g. `""`)
	Suffix string

	// The RegExp used to match this token (string)
	Pattern string

	// The modifier character used for the segment (e.g. `?`)
	Modifier string
}

// Options contains some optional configs
type Options struct {
	// When true the regexp will be case sensitive. (default: false)
	Sensitive bool

	// When true the regexp won't allow an optional trailing delimiter to match. (default: false)
	Strict bool

	// When true the regexp will match to the end of the string. (default: true)
	End *bool

	// When true the regexp will match from the beginning of the string. (default: true)
	Start *bool

	// When `false` the function can produce an invalid (unmatched) path. (default: `true`)
	Validate *bool

	// Sets the final character for non-ending optimistic matches. (default: `/`)
	Delimiter string

	// Optional character to treat as "end" characters.
	EndsWith string

	// List of characters to automatically consider prefixes when parsing. (default: `./`)
	Prefixes *string

	// how to encode uri
	Encode func(uri string, token interface{}) string

	// how to decode uri
	Decode func(str string, token interface{}) (string, error)
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

type lexTokenMode uint8

const (
	modeOpen lexTokenMode = iota
	modeClose
	modePattern
	modeName
	modeChar
	modeEscapedChar
	modeModifier
	modeEnd
)

type lexToken struct {
	mode  lexTokenMode
	index int
	value string
}

var escapeRegexp = regexp2.MustCompile("([.+*?=^!:${}()[\\]|/\\\\])", regexp2.None)
var tokenRegexp = regexp2.MustCompile("\\((?!\\?)", regexp2.None)

func identity(uri string, token interface{}) string {
	return uri
}

// EncodeURIComponent encodes a text string as a valid component of a Uniform
// Resource Identifier (URI).
func EncodeURIComponent(str string) string {
	return strings.Replace(url.QueryEscape(str), "+", "%20", -1)
}

// Gets the unencoded version of an encoded component of a Uniform Resource
// Identifier (URI).
func DecodeURIComponent(str string) (string, error) {
	return url.QueryUnescape(str)
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
func decodeURI(str string) (string, error) {
	magicWords := "1@X#y!Z" // not a good idea
	excludes := []string{"%3B", "%2F", "%3F", "%3A", "%40", "%26", "%3D", "%2B", "%24", "%2C", "%23"}
	r := regexp2.MustCompile(strings.Join(excludes, "|"), regexp2.None)

	str, _ = r.ReplaceFunc(str, func(m regexp2.Match) string {
		return strings.Replace(m.String(), "%", magicWords, -1)
	}, -1, -1)

	str, err := decodeURIComponent(str, nil)
	if err != nil {
		return "", err
	}

	for i, v := range excludes {
		excludes[i] = magicWords + strings.TrimPrefix(v, "%")
	}
	r = regexp2.MustCompile(strings.Join(excludes, "|"), regexp2.None)

	str, _ = r.ReplaceFunc(str, func(m regexp2.Match) string {
		return strings.Replace(m.String(), magicWords, "%", -1)
	}, -1, -1)

	return str, nil
}

// Tokenize input string.
func lexer(str string) ([]lexToken, error) {
	tokens, i := make([]lexToken, 0), 0

	// use list to deal with unicode in str
	arr := strings.Split(str, "")

	length := len(arr)
	for i < length {
		char := arr[i]
		if char == "*" || char == "+" || char == "?" {
			tokens = append(tokens, lexToken{mode: modeModifier, index: i, value: arr[i]})
			i++
			continue
		}

		if char == "\\" {
			tokens = append(tokens, lexToken{mode: modeEscapedChar, index: i, value: arr[i+1]})
			i += 2
			continue
		}

		if char == "{" {
			tokens = append(tokens, lexToken{mode: modeOpen, index: i, value: arr[i]})
			i++
			continue
		}

		if char == "}" {
			tokens = append(tokens, lexToken{mode: modeClose, index: i, value: arr[i]})
			i++
			continue
		}

		if char == ":" {
			name, j := "", i+1

			for j < length {
				if len(arr[j]) == 1 {
					code := arr[j][0]
					isNumber := code >= 48 && code <= 57 // `0-9`
					isUpper := code >= 65 && code <= 90  // `A-Z`
					isLower := code >= 97 && code <= 122 // `a-z`
					isUnderscore := code == 95           // `_`
					if isNumber || isUpper || isLower || isUnderscore {
						name += arr[j]
						j++
						continue
					}
				}

				break
			}

			if name == "" {
				return nil, fmt.Errorf("missing parameter name at %d", i)
			}

			tokens = append(tokens, lexToken{mode: modeName, index: i, value: name})
			i = j
			continue
		}

		if char == "(" {
			count, pattern, j := 1, "", i+1

			if arr[j] == "?" {
				return nil, fmt.Errorf("pattern cannot start with \"?\" at %d", j)
			}

			for j < length {
				if arr[j] == "\\" {
					pattern += arr[j] + arr[j+1]
					j += 2
					continue
				}

				if arr[j] == ")" {
					count--
					if count == 0 {
						j++
						break
					}
				} else if arr[j] == "(" {
					count++
					if arr[j+1] != "?" {
						return nil, fmt.Errorf("capturing groups are not allowed at %d", j)
					}
				}

				pattern += arr[j]
				j++
			}

			if count != 0 {
				return nil, fmt.Errorf("unbalanced pattern at %d", i)
			}
			if pattern == "" {
				return nil, fmt.Errorf("missing pattern at %d", i)
			}

			tokens = append(tokens, lexToken{mode: modePattern, index: i, value: pattern})
			i = j
			continue
		}

		tokens = append(tokens, lexToken{mode: modeChar, index: i, value: arr[i]})
		i++
	}

	tokens = append(tokens, lexToken{mode: modeEnd, index: i, value: ""})

	return tokens, nil
}

// Parse a string for the raw tokens.
func Parse(str string, options *Options) ([]interface{}, error) {
	if options == nil {
		options = &Options{}
	}
	tokens, err := lexer(str)
	if err != nil {
		return nil, err
	}
	prefixes := "./"
	if options.Prefixes != nil {
		prefixes = *options.Prefixes
	}
	delimiter, err := escapeString(anyString(options.Delimiter, "/#?"))
	if err != nil {
		return nil, err
	}
	defaultPattern := "[^" + delimiter + "]+?"
	result, key, i, path := make([]interface{}, 0), 0, 0, ""

	tryConsume := func(mode lexTokenMode) *string {
		if i < len(tokens) && tokens[i].mode == mode {
			result := tokens[i].value
			i++
			return &result
		}
		return nil
	}

	mustConsume := func(mode lexTokenMode) error {
		value := tryConsume(mode)
		if value != nil {
			return nil
		}
		nextMode, index := tokens[i].mode, tokens[i].index
		return fmt.Errorf("unexpected %d at %d, expected %d", nextMode, index, mode)
	}

	consumeText := func() string {
		result, value := "", tryConsume(modeChar)
		if value == nil || *value == "" {
			value = tryConsume(modeEscapedChar)
		}
		for value != nil && *value != "" {
			result += *value
			value = tryConsume(modeChar)
			if value == nil || *value == "" {
				value = tryConsume(modeEscapedChar)
			}
		}
		return result
	}

	for i < len(tokens) {
		char, name, pattern := tryConsume(modeChar), tryConsume(modeName), tryConsume(modePattern)

		if (name != nil && *name != "") || (pattern != nil && *pattern != "") {
			prefix := ""
			if char != nil && *char != "" {
				prefix = *char
			}

			if strings.Index(prefixes, prefix) == -1 {
				path += prefix
				prefix = ""
			}

			if path != "" {
				result = append(result, path)
				path = ""
			}

			result = append(result, Token{
				Name: func() interface{} {
					if name != nil && *name != "" {
						return *name
					}
					result := key
					key++
					return result
				}(),
				Prefix: prefix,
				Suffix: "",
				Pattern: func() string {
					if pattern != nil && *pattern != "" {
						return *pattern
					}
					return defaultPattern
				}(),
				Modifier: func() string {
					result := tryConsume(modeModifier)
					if result != nil && *result != "" {
						return *result
					}
					return ""
				}(),
			})
			continue
		}

		var value *string
		if char != nil && *char != "" {
			value = char
		} else {
			value = tryConsume(modeEscapedChar)
		}
		if value != nil && *value != "" {
			path += *value
			continue
		}

		if path != "" {
			result = append(result, path)
			path = ""
		}

		open := tryConsume(modeOpen)
		if open != nil && *open != "" {
			prefix, name, pattern := consumeText(), tryConsume(modeName), tryConsume(modePattern)
			suffix := consumeText()
			err := mustConsume(modeClose)
			if err != nil {
				return nil, err
			}

			result = append(result, Token{
				Name: func() interface{} {
					if name != nil && *name != "" {
						return *name
					}
					if pattern != nil && *pattern != "" {
						result := key
						key++
						return result
					}
					return ""
				}(),
				Prefix: prefix,
				Suffix: suffix,
				Pattern: func() string {
					if (name != nil && *name != "") && (pattern == nil || *pattern == "") {
						return defaultPattern
					}
					if pattern == nil {
						return ""
					}
					return *pattern
				}(),
				Modifier: func() string {
					result := tryConsume(modeModifier)
					if result != nil && *result != "" {
						return *result
					}
					return ""
				}(),
			})

			continue
		}

		err := mustConsume(modeEnd)
		if err != nil {
			return nil, err
		}
	}

	return result, nil
}

// Compile a string to a template function for the path.
func Compile(str string, options *Options) (func(interface{}) (string, error), error) {
	tokens, err := Parse(str, options)
	if err != nil {
		return nil, err
	}
	return tokensToFunction(tokens, options)
}

// MustCompile is like Compile but panics if the expression cannot be compiled.
// It simplifies safe initialization of global variables.
func MustCompile(str string, options *Options) func(interface{}) (string, error) {
	f, err := Compile(str, options)
	if err != nil {
		panic(`pathtoregexp: Compile(` + quote(str) + `): ` + err.Error())
	}
	return f
}

// Match creates path match function from `path-to-regexp` spec.
func Match(path interface{}, options *Options) (func(string) (*MatchResult, error), error) {
	var tokens []Token
	re, err := PathToRegexp(path, &tokens, options)
	if err != nil {
		return nil, err
	}

	return regexpToFunction(re, tokens, options), nil
}

// MustMatch is like Match but panics if err occur in match function.
func MustMatch(path interface{}, options *Options) func(string) (*MatchResult, error) {
	f, err := Match(path, options)
	if err != nil {
		panic(err)
	}
	return f
}

// Create a path match function from `path-to-regexp` output.
func regexpToFunction(re *regexp2.Regexp, tokens []Token, options *Options) func(string) (*MatchResult, error) {
	decode := func(str string, token interface{}) (string, error) {
		return str, nil
	}
	if options != nil && options.Decode != nil {
		decode = options.Decode
	}

	return func(pathname string) (*MatchResult, error) {
		m, err := re.FindStringMatch(pathname)
		if m == nil || m.GroupCount() == 0 || err != nil {
			return nil, err
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

			if token.Modifier == "*" || token.Modifier == "+" {
				arr := strings.Split(matchedStr, token.Prefix+token.Suffix)
				length := len(arr)
				if length > 0 {
					for i, str := range arr {
						arr[i], err = decode(str, token)
						if err != nil {
							return nil, err
						}
					}
					params[token.Name] = arr
				}
			} else {
				params[token.Name], err = decode(matchedStr, token)
				if err != nil {
					return nil, err
				}
			}
		}

		return &MatchResult{Path: path, Index: index, Params: params}, nil
	}
}

// Expose a method for transforming tokens into the path function.
func tokensToFunction(tokens []interface{}, options *Options) (
	func(interface{}) (string, error), error) {
	if options == nil {
		options = &Options{}
	}
	reFlags := flags(options)
	encode, validate := identity, true
	if options.Encode != nil {
		encode = options.Encode
	}
	if options.Validate != nil {
		validate = *options.Validate
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

	return func(data interface{}) (string, error) {
		path := ""

		for i, token := range tokens {
			if token, ok := token.(string); ok {
				path += token
				continue
			}

			if token, ok := token.(Token); ok {
				optional := token.Modifier == "?" || token.Modifier == "*"
				repeat := token.Modifier == "*" || token.Modifier == "+"
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
							if !repeat {
								return "", fmt.Errorf("expected \"%v\" to not repeat, "+
									"but got array", token.Name)
							}

							if len(value) == 0 {
								if optional {
									continue
								}
								return "", fmt.Errorf("expected \"%v\" to not be empty", token.Name)
							}

							for _, v := range value {
								segment := encode(fmt.Sprintf("%v", v), token)

								if validate {
									if ok, err := matches[i].MatchString(segment); err != nil || !ok {
										return "", fmt.Errorf("expected all \"%v\" to match \"%v\"",
											token.Name, token.Pattern)
									}
								}

								path += token.Prefix + segment + token.Suffix
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
								return "", fmt.Errorf("expected \"%v\" to match \"%v\", "+
									"but got \"%v\"", token.Name, token.Pattern, segment)
							}
						}

						path += token.Prefix + segment + token.Suffix
						continue
					}
				}

				if optional {
					continue
				}

				s := "a string"
				if repeat {
					s = "an array"
				}
				return "", fmt.Errorf("expected \"%v\" to be %v", token.Name, s)
			}
		}

		return path, nil
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

func decodeURIComponent(str string, token interface{}) (string, error) {
	return DecodeURIComponent(str)
}

// Escape a regular expression string.
func escapeString(str string) (string, error) {
	return escapeRegexp.Replace(str, "\\$1", -1, -1)
}

func quote(s string) string {
	if strconv.CanBackquote(s) {
		return "`" + s + "`"
	}
	return strconv.Quote(s)
}

// Get the flags for a regexp from the options.
func flags(options *Options) regexp2.RegexOptions {
	if options != nil && options.Sensitive {
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
			for i := 0; i < totalGroupCount; i++ {
				*tokens = append(*tokens, Token{
					Name:     i,
					Prefix:   "",
					Suffix:   "",
					Modifier: "",
					Pattern:  "",
				})
			}
		}
	}

	return path
}

// Transform an array into a regexp.
func arrayToRegexp(path []interface{}, tokens *[]Token, options *Options) (*regexp2.Regexp, error) {
	var parts []string

	for i := 0; i < len(path); i++ {
		r, err := PathToRegexp(path[i], tokens, options)
		if err != nil {
			return nil, err
		}
		parts = append(parts, r.String())
	}

	return regexp2.Compile("(?:"+strings.Join(parts, "|")+")", flags(options))
}

// Create a path regexp from string input.
func stringToRegexp(path string, tokens *[]Token, options *Options) (*regexp2.Regexp, error) {
	parsedTokens, err := Parse(path, options)
	if err != nil {
		return nil, err
	}
	return tokensToRegExp(parsedTokens, tokens, options)
}

// Expose a function for taking tokens and returning a RegExp.
func tokensToRegExp(rawTokens []interface{}, tokens *[]Token, options *Options) (*regexp2.Regexp, error) {
	if options == nil {
		options = &Options{}
	}

	strict, start, end, route, encode := options.Strict, true, true, "", identity
	if options.Start != nil {
		start = *options.Start
	}
	if options.End != nil {
		end = *options.End
	}
	if options.Encode != nil {
		encode = options.Encode
	}

	endsWith := "$"
	// avoid syntax.ErrUnterminatedBracket `unterminated [] set`
	// empty [] is not allowed in regexp2
	if options.EndsWith != "" {
		t, err := escapeString(options.EndsWith)
		if err != nil {
			return nil, err
		}
		endsWith = "[" + t + "]|$"
	}
	t, err := escapeString(anyString(options.Delimiter, "/#?"))
	if err != nil {
		return nil, err
	}
	delimiter := "[" + t + "]"
	if start {
		route = "^"
	}

	// Iterate over the tokens and create our regexp string.
	for _, token := range rawTokens {
		if str, ok := token.(string); ok {
			t, err := escapeString(encode(str, nil))
			if err != nil {
				return nil, err
			}
			route += t
		} else if token, ok := token.(Token); ok {
			t, err := escapeString(encode(token.Prefix, nil))
			if err != nil {
				return nil, err
			}
			prefix := t
			t, err = escapeString(encode(token.Suffix, nil))
			if err != nil {
				return nil, err
			}
			suffix := t

			if token.Pattern != "" {
				if tokens != nil {
					*tokens = append(*tokens, token)
				}
				if prefix != "" || suffix != "" {
					if token.Modifier == "+" || token.Modifier == "*" {
						mod := ""
						if token.Modifier == "*" {
							mod = "?"
						}
						route += "(?:" + prefix + "((?:" + token.Pattern + ")" +
							"(?:" + suffix + prefix + "(?:" + token.Pattern + "))" +
							"*)" + suffix + ")" + mod
					} else {
						route += "(?:" + prefix + "(" + token.Pattern + ")" +
							"" + suffix + ")" + token.Modifier
					}
				} else {
					route += "(" + token.Pattern + ")" + token.Modifier
				}
			} else {
				route += "(?:" + prefix + suffix + ")" + token.Modifier
			}
		}
	}

	if end {
		if !strict {
			route += delimiter + "?"
		}

		s := "(?=" + endsWith + ")"
		if options.EndsWith == "" {
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
				isEndDelimited = strings.Index(delimiter, str[len(str)-1:]) > -1
			}
		}

		if !strict {
			route += "(?:" + delimiter + "(?=" + endsWith + "))?"
		}
		if !isEndDelimited {
			route += "(?=" + delimiter + "|" + endsWith + ")"
		}
	}

	return regexp2.Compile(route, flags(options))
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
