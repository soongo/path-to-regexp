// Copyright 2019 Guoyao Wu. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package pathtoregexp

import (
	"errors"
	"fmt"
	"reflect"
	"testing"

	"github.com/dlclark/regexp2"
)

type a []interface{}
type m map[interface{}]interface{}

var (
	falseValue      = false
	prefixDollar    = "$"
	testErrorFormat = "got `%v`, expect `%v`"
)

var tests = []a{
	/**
	 * Simple paths.
	 */
	{
		"/",
		nil,
		a{
			"/",
		},
		a{
			a{"/", a{"/"}, &MatchResult{Path: "/", Index: 0, Params: m{}}},
			a{"/route", nil, nil},
		},
		a{
			a{nil, "/"},
			a{m{}, "/"},
			a{m{"id": 123}, "/"},
		},
	},
	{
		"/test",
		nil,
		a{
			"/test",
		},
		a{
			a{"/test", a{"/test"}, &MatchResult{Path: "/test", Index: 0, Params: m{}}},
			a{"/route", nil, nil},
			a{"/test/route", nil, nil},
			a{"/test/", a{"/test/"}, &MatchResult{Path: "/test/", Index: 0, Params: m{}}},
		},
		a{
			a{nil, "/test"},
			a{m{}, "/test"},
		},
	},
	{
		"/test/",
		nil,
		a{
			"/test/",
		},
		a{
			a{"/test", nil},
			a{"/test/", a{"/test/"}},
			a{"/test//", a{"/test//"}},
		},
		a{
			a{nil, "/test/"},
		},
	},

	/**
	 * Case-sensitive paths.
	 */
	{
		"/test",
		&Options{
			Sensitive: true,
		},
		a{
			"/test",
		},
		a{
			a{"/test", a{"/test"}},
			a{"/TEST", nil},
		},
		a{
			a{nil, "/test"},
		},
	},
	{
		"/TEST",
		&Options{
			Sensitive: true,
		},
		a{
			"/TEST",
		},
		a{
			a{"/test", nil},
			a{"/TEST", a{"/TEST"}},
		},
		a{
			a{nil, "/TEST"},
		},
	},

	/**
	 * Strict mode.
	 */
	{
		"/test",
		&Options{
			Strict: true,
		},
		a{
			"/test",
		},
		a{
			a{"/test", a{"/test"}},
			a{"/test/", nil}, a{"/TEST", a{"/TEST"}},
		},
		a{
			a{nil, "/test"},
		},
	},
	{
		"/test/",
		&Options{
			Strict: true,
		},
		a{
			"/test/",
		},
		a{
			a{"/test", nil},
			a{"/test/", a{"/test/"}},
			a{"/test//", nil},
		},
		a{
			a{nil, "/test/"},
		},
	},

	/**
	 * Non-ending mode.
	 */
	{
		"/test",
		&Options{
			End: &falseValue,
		},
		a{
			"/test",
		},
		a{
			a{"/test", a{"/test"}},
			a{"/test/", a{"/test/"}},
			a{"/test/route", a{"/test"}},
			a{"/route", nil},
		},
		a{
			a{nil, "/test"},
		},
	},
	{
		"/test/",
		&Options{
			End: &falseValue,
		},
		a{
			"/test/",
		},
		a{
			a{"/test", nil},
			a{"/test/route", a{"/test/"}},
			a{"/test//", a{"/test//"}},
			a{"/test//route", a{"/test/"}},
		},
		a{
			a{nil, "/test/"},
		},
	},
	{
		"/:test",
		&Options{
			End: &falseValue,
		},
		a{
			Token{
				Name:     "test",
				Prefix:   "/",
				Suffix:   "",
				Modifier: "",
				Pattern:  "[^\\/#\\?]+?",
			},
		},
		a{
			a{
				"/route",
				a{"/route", "route"},
				&MatchResult{Path: "/route", Index: 0, Params: m{"test": "route"}},
			},
			a{
				"/caf%C3%A9",
				a{"/caf%C3%A9", "caf%C3%A9"},
				&MatchResult{Path: "/caf%C3%A9", Index: 0, Params: m{"test": "caf%C3%A9"}},
			},
			a{
				"/caf%C3%A9",
				a{"/caf%C3%A9", "caf%C3%A9"},
				&MatchResult{Path: "/caf%C3%A9", Index: 0, Params: m{"test": "café"}},
				&Options{Decode: decodeURIComponent},
			},
		},
		a{
			a{m{}, nil},
			a{m{"test": "abc"}, "/abc"},
			a{m{"test": "a+b"}, "/a+b"},
			a{
				m{"test": "a+b"},
				"/test",
				&Options{
					Encode: func(uri string, token interface{}) string {
						if token, ok := token.(Token); ok {
							return fmt.Sprintf("%v", token.Name)
						}
						return ""
					},
				},
			},
			a{m{"test": "a+b"}, "/a%2Bb", &Options{Encode: encodeURIComponent}},
		},
	},
	{
		"/:test/",
		&Options{
			End: &falseValue,
		},
		a{
			Token{
				Name:     "test",
				Prefix:   "/",
				Suffix:   "",
				Modifier: "",
				Pattern:  "[^\\/#\\?]+?",
			},
			"/",
		},
		a{
			a{"/route", nil},
			a{"/route/", a{"/route/", "route"}},
		},
		a{
			a{m{"test": "abc"}, "/abc/"},
		},
	},
	{
		"",
		&Options{
			End: &falseValue,
		},
		a{},
		a{
			a{"", a{""}},
			a{"/", a{"/"}},
			a{"route", a{""}},
			a{"/route", a{""}},
			a{"/route/", a{""}},
		},
		a{
			a{nil, ""},
		},
	},

	/**
	 * Non-starting mode.
	 */
	{
		"/test",
		&Options{
			Start: &falseValue,
		},
		a{
			"/test",
		},
		a{
			a{"/test", a{"/test"}},
			a{"/test/", a{"/test/"}},
			a{"/route/test", a{"/test"}},
			a{"/test/route", nil},
			a{"/route/test/deep", nil},
			a{"/route", nil},
		},
		a{
			a{nil, "/test"},
		},
	},
	{
		"/test/",
		&Options{
			Start: &falseValue,
		},
		a{
			"/test/",
		},
		a{
			a{"/test", nil},
			a{"/test/route", nil},
			a{"/test//route", nil},
			a{"/test//", a{"/test//"}},
			a{"/route/test/", a{"/test/"}},
		},
		a{
			a{nil, "/test/"},
		},
	},
	{
		"/:test",
		&Options{
			Start: &falseValue,
		},
		a{
			Token{
				Name:     "test",
				Prefix:   "/",
				Suffix:   "",
				Modifier: "",
				Pattern:  "[^\\/#\\?]+?",
			},
		},
		a{
			a{"/route", a{"/route", "route"}},
		},
		a{
			a{m{}, nil},
			a{m{"test": "abc"}, "/abc"},
			a{m{"test": "a+b"}, "/a+b"},
			a{
				m{"test": "a+b"},
				"/test",
				&Options{
					Encode: func(uri string, token interface{}) string {
						if token, ok := token.(Token); ok {
							return fmt.Sprintf("%v", token.Name)
						}
						return ""
					},
				},
			},
			a{m{"test": "a+b"}, "/a%2Bb", &Options{Encode: encodeURIComponent}},
		},
	},
	{
		"/:test/",
		&Options{
			Start: &falseValue,
		},
		a{
			Token{
				Name:     "test",
				Prefix:   "/",
				Suffix:   "",
				Modifier: "",
				Pattern:  "[^\\/#\\?]+?",
			},
			"/",
		},
		a{
			a{"/route", nil},
			a{"/route/", a{"/route/", "route"}},
		},
		a{
			a{m{"test": "abc"}, "/abc/"},
		},
	},
	{
		"",
		&Options{
			Start: &falseValue,
		},
		a{},
		a{
			a{"", a{""}},
			a{"/", a{"/"}},
			a{"route", a{""}},
			a{"/route", a{""}},
			a{"/route/", a{"/"}},
		},
		a{
			a{nil, ""},
		},
	},

	/**
	 * Combine modes.
	 */
	{
		"/test",
		&Options{
			End:    &falseValue,
			Strict: true,
		},
		a{
			"/test",
		},
		a{
			a{"/test", a{"/test"}},
			a{"/test/", a{"/test"}},
			a{"/test/route", a{"/test"}},
		},
		a{
			a{nil, "/test"},
		},
	},
	{
		"/test/",
		&Options{
			End:    &falseValue,
			Strict: true,
		},
		a{
			"/test/",
		},
		a{
			a{"/test", nil},
			a{"/test/", a{"/test/"}},
			a{"/test//", a{"/test/"}},
			a{"/test/route", a{"/test/"}},
		},
		a{
			a{nil, "/test/"},
		},
	},
	{
		"/test.json",
		&Options{
			End:    &falseValue,
			Strict: true,
		},
		a{
			"/test.json",
		},
		a{
			a{"/test.json", a{"/test.json"}},
			a{"/test.json.hbs", nil},
			a{"/test.json/route", a{"/test.json"}},
		},
		a{
			a{nil, "/test.json"},
		},
	},
	{
		"/:test",
		&Options{
			End:    &falseValue,
			Strict: true,
		},
		a{
			Token{
				Name:     "test",
				Prefix:   "/",
				Suffix:   "",
				Modifier: "",
				Pattern:  "[^\\/#\\?]+?",
			},
		},
		a{
			a{"/route", a{"/route", "route"}},
			a{"/route/", a{"/route", "route"}},
		},
		a{
			a{m{}, nil},
			a{m{"test": "abc"}, "/abc"},
		},
	},
	{
		"/:test/",
		&Options{
			End:    &falseValue,
			Strict: true,
		},
		a{
			Token{
				Name:     "test",
				Prefix:   "/",
				Suffix:   "",
				Modifier: "",
				Pattern:  "[^\\/#\\?]+?",
			},
			"/",
		},
		a{
			a{"/route", nil},
			a{"/route/", a{"/route/", "route"}},
		},
		a{
			a{m{"test": "foobar"}, "/foobar/"},
		},
	},
	{
		"/test",
		&Options{
			Start: &falseValue,
			End:   &falseValue,
		},
		a{
			"/test",
		},
		a{
			a{"/test", a{"/test"}},
			a{"/test/", a{"/test/"}},
			a{"/test/route", a{"/test"}},
			a{"/route/test/deep", a{"/test"}},
		},
		a{
			a{nil, "/test"},
		},
	},
	{
		"/test/",
		&Options{
			Start: &falseValue,
			End:   &falseValue,
		},
		a{
			"/test/",
		},
		a{
			a{"/test", nil},
			a{"/test/", a{"/test/"}},
			a{"/test//", a{"/test//"}},
			a{"/test/route", a{"/test/"}},
			a{"/route/test/deep", a{"/test/"}},
		},
		a{
			a{nil, "/test/"},
		},
	},
	{
		"/test.json",
		&Options{
			Start: &falseValue,
			End:   &falseValue,
		},
		a{
			"/test.json",
		},
		a{
			a{"/test.json", a{"/test.json"}},
			a{"/test.json.hbs", nil},
			a{"/test.json/route", a{"/test.json"}},
			a{"/route/test.json/deep", a{"/test.json"}},
		},
		a{
			a{nil, "/test.json"},
		},
	},
	{
		"/:test",
		&Options{
			Start: &falseValue,
			End:   &falseValue,
		},
		a{
			Token{
				Name:     "test",
				Prefix:   "/",
				Suffix:   "",
				Modifier: "",
				Pattern:  "[^\\/#\\?]+?",
			},
		},
		a{
			a{"/route", a{"/route", "route"}},
			a{"/route/", a{"/route/", "route"}},
		},
		a{
			a{m{}, nil},
			a{m{"test": "abc"}, "/abc"},
		},
	},
	{
		"/:test/",
		&Options{
			Start: &falseValue,
			End:   &falseValue,
		},
		a{
			Token{
				Name:     "test",
				Prefix:   "/",
				Suffix:   "",
				Modifier: "",
				Pattern:  "[^\\/#\\?]+?",
			},
			"/",
		},
		a{
			a{"/route", nil},
			a{"/route/", a{"/route/", "route"}},
		},
		a{
			a{m{"test": "foobar"}, "/foobar/"},
		},
	},

	/**
	 * Arrays of simple paths.
	 */
	{
		a{"/one", "/two"},
		nil,
		a{},
		a{
			a{"/one", a{"/one"}},
			a{"/two", a{"/two"}}, a{"/three", nil},
			a{"/one/two", nil},
		},
		a{},
	},

	/**
	 * Non-ending simple path.
	 */
	{
		"/test",
		&Options{
			End: &falseValue,
		},
		a{
			"/test",
		},
		a{
			a{"/test/route", a{"/test"}},
		},
		a{
			a{nil, "/test"},
		},
	},

	/**
	 * Single named parameter.
	 */
	{
		"/:test",
		nil,
		a{
			Token{
				Name:     "test",
				Prefix:   "/",
				Suffix:   "",
				Modifier: "",
				Pattern:  "[^\\/#\\?]+?",
			},
		},
		a{
			a{"/route", a{"/route", "route"}},
			a{"/another", a{"/another", "another"}},
			a{"/something/else", nil},
			a{"/route.json", a{"/route.json", "route.json"}},
			a{"/something%2Felse", a{"/something%2Felse", "something%2Felse"}},
			a{"/something%2Felse%2Fmore", a{"/something%2Felse%2Fmore", "something%2Felse%2Fmore"}},
			a{"/;,:@&=+$-_.!~*()", a{"/;,:@&=+$-_.!~*()", ";,:@&=+$-_.!~*()"}},
		},
		a{
			a{m{"test": "route"}, "/route"},
			a{
				m{"test": "something/else"},
				"/something%2Felse",
				&Options{Encode: encodeURIComponent},
			},
			a{
				m{"test": "something/else/more"},
				"/something%2Felse%2Fmore",
				&Options{Encode: encodeURIComponent},
			},
		},
	},
	{
		"/:test",
		&Options{
			Strict: true,
		},
		a{
			Token{
				Name:     "test",
				Prefix:   "/",
				Suffix:   "",
				Modifier: "",
				Pattern:  "[^\\/#\\?]+?",
			},
		},
		a{
			a{"/route", a{"/route", "route"}},
			a{"/route/", nil},
		},
		a{
			a{m{"test": "route"}, "/route"},
		},
	},
	{
		"/:test/",
		&Options{
			Strict: true,
		},
		a{
			Token{
				Name:     "test",
				Prefix:   "/",
				Suffix:   "",
				Modifier: "",
				Pattern:  "[^\\/#\\?]+?",
			},
			"/",
		},
		a{
			a{"/route/", a{"/route/", "route"}},
			a{"/route//", nil},
		},
		a{
			a{m{"test": "route"}, "/route/"},
		},
	},
	{
		"/:test",
		&Options{
			End: &falseValue,
		},
		a{
			Token{
				Name:     "test",
				Prefix:   "/",
				Suffix:   "",
				Modifier: "",
				Pattern:  "[^\\/#\\?]+?",
			},
		},
		a{
			a{"/route.json", a{"/route.json", "route.json"}},
			a{"/route//", a{"/route", "route"}},
		},
		a{
			a{m{"test": "route"}, "/route"},
		},
	},

	/**
	 * Optional named parameter.
	 */
	{
		"/:test?",
		nil,
		a{
			Token{
				Name:     "test",
				Prefix:   "/",
				Suffix:   "",
				Modifier: "?",
				Pattern:  "[^\\/#\\?]+?",
			},
		},
		a{
			a{
				"/route",
				a{"/route", "route"},
				&MatchResult{Path: "/route", Index: 0, Params: m{"test": "route"}},
			},
			a{"/route/nested", nil, nil},
			a{"/", a{"/", ""}, &MatchResult{Path: "/", Index: 0, Params: m{}}},
			a{"//", nil},
		},
		a{
			a{nil, ""},
			a{m{"test": "foobar"}, "/foobar"},
		},
	},
	{
		"/:test?",
		&Options{
			Strict: true,
		},
		a{
			Token{
				Name:     "test",
				Prefix:   "/",
				Suffix:   "",
				Modifier: "?",
				Pattern:  "[^\\/#\\?]+?",
			},
		},
		a{
			a{"/route", a{"/route", "route"}},
			a{"/", nil}, a{"//", nil},
		},
		a{
			a{nil, ""},
			a{m{"test": "foobar"}, "/foobar"},
		},
	},
	{
		"/:test?/",
		&Options{
			Strict: true,
		},
		a{
			Token{
				Name:     "test",
				Prefix:   "/",
				Suffix:   "",
				Modifier: "?",
				Pattern:  "[^\\/#\\?]+?",
			},
			"/",
		},
		a{
			a{"/route", nil},
			a{"/route/", a{"/route/", "route"}},
			a{"/", a{"/", ""}},
			a{"//", nil},
		},
		a{
			a{nil, "/"},
			a{m{"test": "foobar"}, "/foobar/"},
		},
	},
	{
		"/:test?/bar",
		nil,
		a{
			Token{
				Name:     "test",
				Prefix:   "/",
				Suffix:   "",
				Modifier: "?",
				Pattern:  "[^\\/#\\?]+?",
			},
			"/bar",
		},
		a{
			a{"/bar", a{"/bar", ""}},
			a{"/foo/bar", a{"/foo/bar", "foo"}},
		},
		a{
			a{nil, "/bar"},
			a{m{"test": "foo"}, "/foo/bar"},
		},
	},
	{
		"/:test?-bar",
		nil,
		a{
			Token{
				Name:     "test",
				Prefix:   "/",
				Suffix:   "",
				Modifier: "?",
				Pattern:  "[^\\/#\\?]+?",
			},
			"-bar",
		},
		a{
			a{"-bar", a{"-bar", ""}},
			a{"/-bar", nil},
			a{"/foo-bar", a{"/foo-bar", "foo"}},
		},
		a{
			a{"", "-bar"},
			a{m{"test": "foo"}, "/foo-bar"},
		},
	},
	{
		"/:test*-bar",
		nil,
		a{
			Token{
				Name:     "test",
				Prefix:   "/",
				Suffix:   "",
				Modifier: "*",
				Pattern:  "[^\\/#\\?]+?",
			},
			"-bar",
		},
		a{
			a{"-bar", a{"-bar", ""}},
			a{"/-bar", nil},
			a{"/foo-bar", a{"/foo-bar", "foo"}},
			a{"/foo/baz-bar", a{"/foo/baz-bar", "foo/baz"}},
		},
		a{
			a{m{"test": "foo"}, "/foo-bar"},
		},
	},

	/**
	 * Repeated one or more times parameters.
	 */
	{
		"/:test+",
		nil,
		a{
			Token{
				Name:     "test",
				Prefix:   "/",
				Suffix:   "",
				Modifier: "+",
				Pattern:  "[^\\/#\\?]+?",
			},
		},
		a{
			a{"/", nil, nil},
			a{
				"/route",
				a{"/route", "route"},
				&MatchResult{Path: "/route", Index: 0, Params: m{"test": a{"route"}}},
			},
			a{
				"/some/basic/route",
				a{"/some/basic/route", "some/basic/route"},
				&MatchResult{
					Path:   "/some/basic/route",
					Index:  0,
					Params: m{"test": a{"some", "basic", "route"}},
				},
			},
			a{"//", nil, nil},
		},
		a{
			a{m{}, nil},
			a{m{"test": "foobar"}, "/foobar"},
			a{m{"test": a{"a", "b", "c"}}, "/a/b/c"},
		},
	},
	{
		"/:test(\\d+)+",
		nil,
		a{
			Token{
				Name:     "test",
				Prefix:   "/",
				Suffix:   "",
				Modifier: "+",
				Pattern:  "\\d+",
			},
		},
		a{
			a{"/abc/456/789", nil},
			a{"/123/456/789", a{"/123/456/789", "123/456/789"}},
		},
		a{
			a{m{"test": "abc"}, nil},
			a{m{"test": 123}, "/123"},
			a{m{"test": a{1, 2, 3}}, "/1/2/3"},
		},
	},
	{
		"/route.:ext(json|xml)+",
		nil,
		a{
			"/route",
			Token{
				Name:     "ext",
				Prefix:   ".",
				Suffix:   "",
				Modifier: "+",
				Pattern:  "json|xml",
			},
		},
		a{
			a{"/route", nil},
			a{"/route.json", a{"/route.json", "json"}},
			a{"/route.xml.json", a{"/route.xml.json", "xml.json"}},
			a{"/route.html", nil},
		},
		a{
			a{m{"ext": "foobar"}, nil},
			a{m{"ext": "xml"}, "/route.xml"},
			a{m{"ext": a{"xml", "json"}}, "/route.xml.json"},
		},
	},
	{
		"/route.:ext(\\w+)/test",
		nil,
		a{
			"/route",
			Token{
				Name:     "ext",
				Prefix:   ".",
				Suffix:   "",
				Modifier: "",
				Pattern:  "\\w+",
			},
			"/test",
		},
		a{
			a{"/route", nil},
			a{"/route.json", nil},
			a{"/route.xml/test", a{"/route.xml/test", "xml"}},
			a{"/route.json.gz/test", nil},
		},
		a{
			a{m{"ext": "xml"}, "/route.xml/test"},
		},
	},

	/**
	 * Repeated zero or more times parameters.
	 */
	{
		"/:test*",
		nil,
		a{
			Token{
				Name:     "test",
				Prefix:   "/",
				Suffix:   "",
				Modifier: "*",
				Pattern:  "[^\\/#\\?]+?",
			},
		},
		a{
			a{"/", a{"/", ""}, &MatchResult{Path: "/", Index: 0, Params: m{}}},
			a{"//", nil, nil},
			a{
				"/route",
				a{"/route", "route"},
				&MatchResult{Path: "/route", Index: 0, Params: m{"test": a{"route"}}},
			},
			a{
				"/some/basic/route",
				a{"/some/basic/route", "some/basic/route"},
				&MatchResult{
					Path:   "/some/basic/route",
					Index:  0,
					Params: m{"test": a{"some", "basic", "route"}},
				},
			},
		},
		a{
			a{m{}, ""},
			a{m{"test": a{}}, ""},
			a{m{"test": "foobar"}, "/foobar"},
			a{m{"test": a{"foo", "bar"}}, "/foo/bar"},
		},
	},
	{
		"/route.:ext([a-z]+)*",
		nil,
		a{
			"/route",
			Token{
				Name:     "ext",
				Prefix:   ".",
				Suffix:   "",
				Modifier: "*",
				Pattern:  "[a-z]+",
			},
		},
		a{
			a{"/route", a{"/route", ""}},
			a{"/route.json", a{"/route.json", "json"}},
			a{"/route.json.xml", a{"/route.json.xml", "json.xml"}},
			a{"/route.123", nil},
		},
		a{
			a{m{}, "/route"},
			a{m{"ext": a{}}, "/route"},
			a{m{"ext": "123"}, nil},
			a{m{"ext": "foobar"}, "/route.foobar"},
			a{m{"ext": a{"foo", "bar"}}, "/route.foo.bar"},
		},
	},

	/**
	 * Custom named parameters.
	 */
	{
		"/:test(\\d+)",
		nil,
		a{
			Token{
				Name:     "test",
				Prefix:   "/",
				Suffix:   "",
				Modifier: "",
				Pattern:  "\\d+",
			},
		},
		a{
			a{"/123", a{"/123", "123"}},
			a{"/abc", nil},
			a{"/123/abc", nil},
		},
		a{
			a{m{"test": "abc"}, nil},
			a{m{"test": "abc"}, "/abc", &Options{Validate: &falseValue}},
			a{m{"test": "123"}, "/123"},
		},
	},
	{
		"/:test(\\d+)",
		&Options{
			End: &falseValue,
		},
		a{
			Token{
				Name:     "test",
				Prefix:   "/",
				Suffix:   "",
				Modifier: "",
				Pattern:  "\\d+",
			},
		},
		a{
			a{"/123", a{"/123", "123"}},
			a{"/abc", nil},
			a{"/123/abc", a{"/123", "123"}},
		},
		a{
			a{m{"test": "123"}, "/123"},
		},
	},
	{
		"/:test(.*)",
		nil,
		a{
			Token{
				Name:     "test",
				Prefix:   "/",
				Suffix:   "",
				Modifier: "",
				Pattern:  ".*",
			},
		},
		a{
			a{"/anything/goes/here", a{"/anything/goes/here", "anything/goes/here"}},
			a{"/;,:@&=/+$-_.!/~*()", a{"/;,:@&=/+$-_.!/~*()", ";,:@&=/+$-_.!/~*()"}},
		},
		a{
			a{m{"test": ""}, "/"},
			a{m{"test": "abc"}, "/abc"},
			a{m{"test": "abc/123"}, "/abc%2F123", &Options{Encode: encodeURIComponent}},
			a{m{"test": "abc/123/456"}, "/abc%2F123%2F456", &Options{Encode: encodeURIComponent}},
		},
	},
	{
		"/:route([a-z]+)",
		nil,
		a{
			Token{
				Name:     "route",
				Prefix:   "/",
				Suffix:   "",
				Modifier: "",
				Pattern:  "[a-z]+",
			},
		},
		a{
			a{"/abcde", a{"/abcde", "abcde"}},
			a{"/12345", nil},
		},
		a{
			a{m{"route": ""}, nil},
			a{m{"route": ""}, "/", &Options{Validate: &falseValue}},
			a{m{"route": "123"}, nil},
			a{m{"route": "123"}, "/123", &Options{Validate: &falseValue}},
			a{m{"route": "abc"}, "/abc"},
		},
	},
	{
		"/:route(this|that)",
		nil,
		a{
			Token{
				Name:     "route",
				Prefix:   "/",
				Suffix:   "",
				Modifier: "",
				Pattern:  "this|that",
			},
		},
		a{
			a{"/this", a{"/this", "this"}},
			a{"/that", a{"/that", "that"}},
			a{"/foo", nil},
		},
		a{
			a{m{"route": "this"}, "/this"},
			a{m{"route": "foo"}, nil},
			a{m{"route": "foo"}, "/foo", &Options{Validate: &falseValue}},
			a{m{"route": "that"}, "/that"},
		},
	},
	{
		"/:path(abc|xyz)*",
		nil,
		a{
			Token{
				Name:     "path",
				Prefix:   "/",
				Suffix:   "",
				Modifier: "*",
				Pattern:  "abc|xyz",
			},
		},
		a{
			a{"/abc", a{"/abc", "abc"}},
			a{"/abc/abc", a{"/abc/abc", "abc/abc"}},
			a{"/xyz/xyz", a{"/xyz/xyz", "xyz/xyz"}},
			a{"/abc/xyz", a{"/abc/xyz", "abc/xyz"}},
			a{"/abc/xyz/abc/xyz", a{"/abc/xyz/abc/xyz", "abc/xyz/abc/xyz"}},
			a{"/xyzxyz", nil},
		},
		a{
			a{m{"path": "abc"}, "/abc"}, a{m{"path": a{"abc", "xyz"}}, "/abc/xyz"},
			a{m{"path": a{"xyz", "abc", "xyz"}}, "/xyz/abc/xyz"},
			a{m{"path": "abc123"}, nil},
			a{m{"path": "abc123"}, "/abc123", &Options{Validate: &falseValue}},
			a{m{"path": "abcxyz"}, nil},
			a{m{"path": "abcxyz"}, "/abcxyz", &Options{Validate: &falseValue}},
		},
	},

	/**
	 * Prefixed slashes could be omitted.
	 */
	{
		"test",
		nil,
		a{
			"test",
		},
		a{
			a{"test", a{"test"}},
			a{"/test", nil},
		},
		a{
			a{nil, "test"},
		},
	},
	{
		":test",
		nil,
		a{
			Token{
				Name:     "test",
				Prefix:   "",
				Suffix:   "",
				Modifier: "",
				Pattern:  "[^\\/#\\?]+?",
			},
		},
		a{
			a{"route", a{"route", "route"}},
			a{"/route", nil},
			a{"route/", a{"route/", "route"}},
		},
		a{
			a{m{"test": ""}, nil},
			a{m{}, nil},
			a{m{"test": nil}, nil},
			a{m{"test": "route"}, "route"},
		},
	},
	{
		":test",
		&Options{
			Strict: true,
		},
		a{
			Token{
				Name:     "test",
				Prefix:   "",
				Suffix:   "",
				Modifier: "",
				Pattern:  "[^\\/#\\?]+?",
			},
		},
		a{
			a{"route", a{"route", "route"}},
			a{"/route", nil}, a{"route/", nil},
		},
		a{
			a{m{"test": "route"}, "route"},
		},
	},
	{
		":test",
		&Options{
			End: &falseValue,
		},
		a{
			Token{
				Name:     "test",
				Prefix:   "",
				Suffix:   "",
				Modifier: "",
				Pattern:  "[^\\/#\\?]+?",
			},
		},
		a{
			a{"route", a{"route", "route"}},
			a{"/route", nil},
			a{"route/", a{"route/", "route"}},
			a{"route/foobar", a{"route", "route"}},
		},
		a{
			a{m{"test": "route"}, "route"},
		},
	},
	{
		":test?",
		nil,
		a{
			Token{
				Name:     "test",
				Prefix:   "",
				Suffix:   "",
				Modifier: "?",
				Pattern:  "[^\\/#\\?]+?",
			},
		},
		a{
			a{"route", a{"route", "route"}},
			a{"/route", nil}, a{"", a{"", ""}},
			a{"route/foobar", nil},
		},
		a{
			a{m{}, ""},
			a{m{"test": ""}, nil},
			a{m{"test": "route"}, "route"},
		},
	},
	{
		"{:test/}+",
		nil,
		a{
			Token{
				Name:     "test",
				Prefix:   "",
				Suffix:   "/",
				Modifier: "+",
				Pattern:  "[^\\/#\\?]+?",
			},
		},
		a{
			a{"route/", a{"route/", "route"}},
			a{"/route", nil},
			a{"", nil},
			a{"foo/bar/", a{"foo/bar/", "foo/bar"}},
		},
		a{
			a{m{}, nil},
			a{m{"test": ""}, nil},
			a{m{"test": a{"route"}}, "route/"},
			a{m{"test": a{"foo", "bar"}}, "foo/bar/"},
		},
	},

	/**
	 * Formats.
	 */
	{
		"/test.json",
		nil,
		a{
			"/test.json",
		},
		a{
			a{"/test.json", a{"/test.json"}},
			a{"/route.json", nil},
		},
		a{
			a{m{}, "/test.json"},
		},
	},
	{
		"/:test.json",
		nil,
		a{
			Token{
				Name:     "test",
				Prefix:   "/",
				Suffix:   "",
				Modifier: "",
				Pattern:  "[^\\/#\\?]+?",
			},
			".json",
		},
		a{
			a{"/.json", nil},
			a{"/test.json", a{"/test.json", "test"}},
			a{"/route.json", a{"/route.json", "route"}},
			a{"/route.json.json", a{"/route.json.json", "route.json"}},
		},
		a{
			a{m{"test": ""}, nil},
			a{m{"test": "foo"}, "/foo.json"},
		},
	},

	/**
	 * Format params.
	 */
	{
		"/test.:format(\\w+)",
		nil,
		a{
			"/test",
			Token{
				Name:     "format",
				Prefix:   ".",
				Suffix:   "",
				Modifier: "",
				Pattern:  "\\w+",
			},
		},
		a{
			a{"/test.html", a{"/test.html", "html"}},
			a{"/test.hbs.html", nil},
		},
		a{
			a{m{}, nil},
			a{m{"format": ""}, nil},
			a{m{"format": ""}, nil},
			a{m{"format": "foo"}, "/test.foo"},
		},
	},
	{
		"/test.:format(\\w+).:format(\\w+)",
		nil,
		a{
			"/test",
			Token{
				Name:     "format",
				Prefix:   ".",
				Suffix:   "",
				Modifier: "",
				Pattern:  "\\w+",
			},
			Token{
				Name:     "format",
				Prefix:   ".",
				Suffix:   "",
				Modifier: "",
				Pattern:  "\\w+",
			},
		},
		a{
			a{"/test.html", nil},
			a{"/test.hbs.html", a{"/test.hbs.html", "hbs", "html"}},
		},
		a{
			a{m{"format": "foo.bar"}, nil},
			a{m{"format": "foo"}, "/test.foo.foo"},
		},
	},
	{
		"/test{.:format}+",
		nil,
		a{
			"/test",
			Token{
				Name:     "format",
				Prefix:   ".",
				Suffix:   "",
				Modifier: "+",
				Pattern:  "[^\\/#\\?]+?",
			},
		},
		a{
			a{"/test.html", a{"/test.html", "html"}},
			a{"/test.hbs.html", a{"/test.hbs.html", "hbs.html"}},
		},
		a{
			a{m{"format": a{}}, nil},
			a{m{"format": "foo"}, "/test.foo"},
			a{m{"format": a{"foo", "bar"}}, "/test.foo.bar"},
		},
	},
	{
		"/test.:format(\\w+)",
		&Options{
			End: &falseValue,
		},
		a{
			"/test",
			Token{
				Name:     "format",
				Prefix:   ".",
				Suffix:   "",
				Modifier: "",
				Pattern:  "\\w+",
			},
		},
		a{
			a{"/test.html", a{"/test.html", "html"}},
			a{"/test.hbs.html", nil},
		},
		a{
			a{m{"format": "foo"}, "/test.foo"},
		},
	},
	{
		"/test.:format.",
		nil,
		a{
			"/test",
			Token{
				Name:     "format",
				Prefix:   ".",
				Suffix:   "",
				Modifier: "",
				Pattern:  "[^\\/#\\?]+?",
			},
			".",
		},
		a{
			a{"/test.html.", a{"/test.html.", "html"}},
			a{"/test.hbs.html", nil},
		},
		a{
			a{m{"format": ""}, nil},
			a{m{"format": "foo"}, "/test.foo."},
		},
	},

	/**
	 * Format and path params.
	 */
	{
		"/:test.:format",
		nil,
		a{
			Token{
				Name:     "test",
				Prefix:   "/",
				Suffix:   "",
				Modifier: "",
				Pattern:  "[^\\/#\\?]+?",
			},
			Token{
				Name:     "format",
				Prefix:   ".",
				Suffix:   "",
				Modifier: "",
				Pattern:  "[^\\/#\\?]+?",
			},
		},
		a{
			a{"/route.html", a{"/route.html", "route", "html"}},
			a{"route", nil},
			a{"/route.html.json", a{"/route.html.json", "route", "html.json"}},
		},
		a{
			a{m{}, nil},
			a{m{"test": "route", "format": "foo"}, "/route.foo"},
		},
	},
	{
		"/:test{.:format}?",
		nil,
		a{
			Token{
				Name:     "test",
				Prefix:   "/",
				Suffix:   "",
				Modifier: "",
				Pattern:  "[^\\/#\\?]+?",
			},
			Token{
				Name:     "format",
				Prefix:   ".",
				Suffix:   "",
				Modifier: "?",
				Pattern:  "[^\\/#\\?]+?",
			},
		},
		a{
			a{"/route", a{"/route", "route", ""}},
			a{"/route.json", a{"/route.json", "route", "json"}},
			a{"/route.json.html", a{"/route.json.html", "route", "json.html"}},
		},
		a{
			a{m{"test": "route"}, "/route"},
			a{m{"test": "route", "format": ""}, nil},
			a{m{"test": "route", "format": "foo"}, "/route.foo"},
		},
	},
	{
		"/:test.:format?",
		&Options{
			End: &falseValue,
		},
		a{
			Token{
				Name:     "test",
				Prefix:   "/",
				Suffix:   "",
				Modifier: "",
				Pattern:  "[^\\/#\\?]+?",
			},
			Token{
				Name:     "format",
				Prefix:   ".",
				Suffix:   "",
				Modifier: "?",
				Pattern:  "[^\\/#\\?]+?",
			},
		},
		a{
			a{"/route", a{"/route", "route", ""}},
			a{"/route.json", a{"/route.json", "route", "json"}},
			a{"/route.json.html", a{"/route.json.html", "route", "json.html"}},
		},
		a{
			a{m{"test": "route"}, "/route"},
			a{m{"test": "route", "format": nil}, "/route"},
			a{m{"test": "route", "format": ""}, nil},
			a{m{"test": "route", "format": "foo"}, "/route.foo"},
		},
	},
	{
		"/test.:format(.*)z",
		&Options{
			End: &falseValue,
		},
		a{
			"/test",
			Token{
				Name:     "format",
				Prefix:   ".",
				Suffix:   "",
				Modifier: "",
				Pattern:  ".*",
			},
			"z",
		},
		a{
			a{"/test.abc", nil},
			a{"/test.z", a{"/test.z", ""}},
			a{"/test.abcz", a{"/test.abcz", "abc"}},
		},
		a{
			a{m{}, nil},
			a{m{"format": ""}, "/test.z"},
			a{m{"format": "foo"}, "/test.fooz"},
		},
	},

	/**
	 * Unnamed params.
	 */
	{
		"/(\\d+)",
		nil,
		a{
			Token{
				Name:     0,
				Prefix:   "/",
				Suffix:   "",
				Modifier: "",
				Pattern:  "\\d+",
			},
		},
		a{
			a{"/123", a{"/123", "123"}},
			a{"/abc", nil},
			a{"/123/abc", nil},
		},
		a{
			a{m{}, nil},
			a{m{"0": "123"}, "/123"},
		},
	},
	{
		"/(\\d+)",
		&Options{
			End: &falseValue,
		},
		a{
			Token{
				Name:     0,
				Prefix:   "/",
				Suffix:   "",
				Modifier: "",
				Pattern:  "\\d+",
			},
		},
		a{
			a{"/123", a{"/123", "123"}},
			a{"/abc", nil},
			a{"/123/abc", a{"/123", "123"}},
			a{"/123/", a{"/123/", "123"}},
		},
		a{
			a{m{"0": "123"}, "/123"},
		},
	},
	{
		"/(\\d+)?",
		nil,
		a{
			Token{
				Name:     0,
				Prefix:   "/",
				Suffix:   "",
				Modifier: "?",
				Pattern:  "\\d+",
			},
		},
		a{
			a{"/", a{"/", ""}},
			a{"/123", a{"/123", "123"}},
		},
		a{
			a{m{}, ""},
			a{m{"0": "123"}, "/123"},
		},
	},
	{
		"/(.*)",
		nil,
		a{
			Token{
				Name:     0,
				Prefix:   "/",
				Suffix:   "",
				Modifier: "",
				Pattern:  ".*",
			},
		},
		a{
			a{"/", a{"/", ""}},
			a{"/route", a{"/route", "route"}},
			a{"/route/nested", a{"/route/nested", "route/nested"}},
		},
		a{
			a{m{"0": ""}, "/"},
			a{m{"0": "123"}, "/123"},
		},
	},
	{
		"/route\\(\\\\(\\d+\\\\)\\)",
		nil,
		a{
			"/route(\\",
			Token{
				Name:     0,
				Prefix:   "",
				Suffix:   "",
				Modifier: "",
				Pattern:  "\\d+\\\\",
			},
			")",
		},
		a{
			a{"/route(\\123\\)", a{"/route(\\123\\)", "123\\"}},
		},
		a{},
	},
	{
		"{/login}?",
		nil,
		a{
			Token{
				Name:     "",
				Prefix:   "/login",
				Suffix:   "",
				Modifier: "?",
				Pattern:  "",
			},
		},
		a{
			a{"/", a{"/"}},
			a{"/login", a{"/login"}},
		},
		a{
			a{nil, ""},
			a{m{"": ""}, "/login"},
		},
	},
	{
		"{/login}",
		nil,
		a{
			Token{
				Name:     "",
				Prefix:   "/login",
				Suffix:   "",
				Modifier: "",
				Pattern:  "",
			},
		},
		a{
			a{"/", nil},
			a{"/login", a{"/login"}},
		},
		a{
			a{m{"": ""}, "/login"},
		},
	},
	{
		"{/(.*)}",
		nil,
		a{
			Token{
				Name:     0,
				Prefix:   "/",
				Suffix:   "",
				Modifier: "",
				Pattern:  ".*",
			},
		},
		a{
			a{"/", a{"/", ""}},
			a{"/login", a{"/login", "login"}},
		},
		a{
			a{m{0: "test"}, "/test"},
		},
	},

	/**
	 * Regexps.
	 */
	{
		regexp2.MustCompile(".*", regexp2.None),
		nil,
		a{},
		a{
			a{"/match/anything", a{"/match/anything"}},
		},
		a{},
	},
	{
		regexp2.MustCompile("(.*)", regexp2.None),
		nil,
		a{
			Token{
				Name:     0,
				Prefix:   "",
				Suffix:   "",
				Modifier: "",
				Pattern:  "",
			},
		},
		a{
			a{"/match/anything", a{"/match/anything", "/match/anything"}},
		},
		a{},
	},
	{
		regexp2.MustCompile("\\/(\\d+)", regexp2.None),
		nil,
		a{
			Token{
				Name:     0,
				Prefix:   "",
				Suffix:   "",
				Modifier: "",
				Pattern:  "",
			},
		},
		a{
			a{"/abc/anything", nil},
			a{"/123", a{"/123", "123"}},
		},
		a{},
	},

	/**
	 * Mixed arrays.
	 */
	{
		a{"/test", regexp2.MustCompile("\\/(\\d+)", regexp2.None)},
		nil,
		a{
			Token{
				Name:     0,
				Prefix:   "",
				Suffix:   "",
				Modifier: "",
				Pattern:  "",
			},
		},
		a{
			a{"/test", a{"/test", ""}},
		},
		a{},
	},
	{
		a{
			"/:test(\\d+)",
			regexp2.MustCompile("(.*)", regexp2.None),
		},
		nil,
		a{
			Token{
				Name:     "test",
				Prefix:   "/",
				Suffix:   "",
				Modifier: "",
				Pattern:  "\\d+",
			},
			Token{
				Name:     0,
				Prefix:   "",
				Suffix:   "",
				Modifier: "",
				Pattern:  "",
			},
		},
		a{
			a{"/123", a{"/123", "123", ""}},
			a{"/abc", a{"/abc", "", "/abc"}},
		},
		a{},
	},

	/**
	 * Correct names and indexes.
	 */
	{
		a{
			"/:test",
			"/route/:test",
		},
		nil,
		a{
			Token{
				Name:     "test",
				Prefix:   "/",
				Suffix:   "",
				Modifier: "",
				Pattern:  "[^\\/#\\?]+?",
			},
			Token{
				Name:     "test",
				Prefix:   "/",
				Suffix:   "",
				Modifier: "",
				Pattern:  "[^\\/#\\?]+?",
			},
		},
		a{
			a{"/test", a{"/test", "test", ""}},
			a{"/route/test", a{"/route/test", "", "test"}},
		},
		a{},
	},
	{
		a{
			regexp2.MustCompile("^\\/([^\\/]+)$", regexp2.None),
			regexp2.MustCompile("^\\/route\\/([^\\/]+)$", regexp2.None),
		},
		nil,
		a{
			Token{
				Name:     0,
				Prefix:   "",
				Suffix:   "",
				Modifier: "",
				Pattern:  "",
			},
			Token{
				Name:     0,
				Prefix:   "",
				Suffix:   "",
				Modifier: "",
				Pattern:  "",
			},
		},
		a{
			a{"/test", a{"/test", "test", ""}},
			a{"/route/test", a{"/route/test", "", "test"}},
		},
		a{},
	},

	/**
	 * Ignore non-matching groups in regexps.
	 */
	{
		regexp2.MustCompile("(?:.*)", regexp2.None),
		nil,
		a{},
		a{
			a{"/anything/you/want", a{"/anything/you/want"}},
		},
		a{},
	},

	/**
	 * Respect escaped characters.
	 */
	{
		"/\\(testing\\)",
		nil,
		a{
			"/(testing)",
		},
		a{
			a{"/testing", nil},
			a{"/(testing)", a{"/(testing)"}},
		},
		a{
			a{nil, "/(testing)"},
		},
	},
	{
		"/.\\+\\*\\?\\{\\}=^!\\:$[]|",
		nil,
		a{
			"/.+*?{}=^!:$[]|",
		},
		a{
			a{"/.+*?{}=^!:$[]|", a{"/.+*?{}=^!:$[]|"}},
		},
		a{
			a{nil, "/.+*?{}=^!:$[]|"},
		},
	},
	{
		"/test\\/:uid(u\\d+)?:cid(c\\d+)?",
		nil,
		a{
			"/test/",
			Token{
				Name:     "uid",
				Prefix:   "",
				Suffix:   "",
				Modifier: "?",
				Pattern:  "u\\d+",
			},
			Token{
				Name:     "cid",
				Prefix:   "",
				Suffix:   "",
				Modifier: "?",
				Pattern:  "c\\d+",
			},
		},
		a{
			a{"/test", nil},
			a{"/test/", a{"/test/", "", ""}},
			a{"/test/u123", a{"/test/u123", "u123", ""}},
			a{"/test/c123", a{"/test/c123", "", "c123"}},
		},
		a{
			a{m{"uid": "u123"}, "/test/u123"},
			a{m{"cid": "c123"}, "/test/c123"},
			a{m{"cid": "u123"}, nil},
		},
	},

	/**
	 * Unnamed group prefix.
	 */
	{
		"/{apple-}?icon-:res(\\d+).png",
		nil,
		a{
			"/",
			Token{
				Name:     "",
				Prefix:   "apple-",
				Suffix:   "",
				Modifier: "?",
				Pattern:  "",
			},
			"icon-",
			Token{
				Name:     "res",
				Prefix:   "",
				Suffix:   "",
				Modifier: "",
				Pattern:  "\\d+",
			},
			".png",
		},
		a{
			a{"/icon-240.png", a{"/icon-240.png", "240"}},
			a{"/apple-icon-240.png", a{"/apple-icon-240.png", "240"}},
		},
		a{},
	},

	/**
	 * Random examples.
	 */
	{
		"/:foo/:bar",
		nil,
		a{
			Token{
				Name:     "foo",
				Prefix:   "/",
				Suffix:   "",
				Modifier: "",
				Pattern:  "[^\\/#\\?]+?",
			},
			Token{
				Name:     "bar",
				Prefix:   "/",
				Suffix:   "",
				Modifier: "",
				Pattern:  "[^\\/#\\?]+?",
			},
		},
		a{
			a{"/match/route", a{"/match/route", "match", "route"}},
		},
		a{
			a{m{"foo": "a", "bar": "b"}, "/a/b"},
		},
	},
	{
		"/:foo\\(test\\)/bar",
		nil,
		a{
			Token{
				Name:     "foo",
				Prefix:   "/",
				Suffix:   "",
				Modifier: "",
				Pattern:  "[^\\/#\\?]+?",
			},
			"(test)/bar",
		},
		a{},
		a{},
	},
	{
		"/:remote([\\w-.]+)/:user([\\w-]+)",
		nil,
		a{
			Token{
				Name:     "remote",
				Prefix:   "/",
				Suffix:   "",
				Modifier: "",
				Pattern:  "[\\w-.]+",
			},
			Token{
				Name:     "user",
				Prefix:   "/",
				Suffix:   "",
				Modifier: "",
				Pattern:  "[\\w-]+",
			},
		},
		a{
			a{"/endpoint/user", a{"/endpoint/user", "endpoint", "user"}},
			a{"/endpoint/user-name", a{"/endpoint/user-name", "endpoint", "user-name"}},
			a{"/foo.bar/user-name", a{"/foo.bar/user-name", "foo.bar", "user-name"}},
		},
		a{
			a{m{"remote": "foo", "user": "bar"}, "/foo/bar"},
			a{m{"remote": "foo.bar", "user": "uno"}, "/foo.bar/uno"},
		},
	},
	{
		"/:foo\\?",
		nil,
		a{
			Token{
				Name:     "foo",
				Prefix:   "/",
				Suffix:   "",
				Modifier: "",
				Pattern:  "[^\\/#\\?]+?",
			},
			"?",
		},
		a{
			a{"/route?", a{"/route?", "route"}},
		},
		a{
			a{m{"foo": "bar"}, "/bar?"},
		},
	},
	{
		"/:foo+baz",
		nil,
		a{
			Token{
				Name:     "foo",
				Prefix:   "/",
				Suffix:   "",
				Modifier: "+",
				Pattern:  "[^\\/#\\?]+?",
			},
			"baz",
		},
		a{
			a{"/foobaz", a{"/foobaz", "foo"}},
			a{"/foo/barbaz", a{"/foo/barbaz", "foo/bar"}},
			a{"/baz", nil},
		},
		a{
			a{m{"foo": "foo"}, "/foobaz"},
			a{m{"foo": "foo/bar"}, "/foo%2Fbarbaz", &Options{Encode: encodeURIComponent}},
			a{m{"foo": a{"foo", "bar"}}, "/foo/barbaz"},
		},
	},
	{
		"\\/:pre?baz",
		nil,
		a{
			"/",
			Token{
				Name:     "pre",
				Prefix:   "",
				Suffix:   "",
				Modifier: "?",
				Pattern:  "[^\\/#\\?]+?",
			},
			"baz",
		},
		a{
			a{"/foobaz", a{"/foobaz", "foo"}},
			a{"/baz", a{"/baz", ""}},
		},
		a{
			a{m{}, "/baz"},
			a{m{"pre": "foo"}, "/foobaz"},
		},
	},
	{
		"/:foo\\(:bar?\\)",
		nil,
		a{
			Token{
				Name:     "foo",
				Prefix:   "/",
				Suffix:   "",
				Modifier: "",
				Pattern:  "[^\\/#\\?]+?",
			},
			"(",
			Token{
				Name:     "bar",
				Prefix:   "",
				Suffix:   "",
				Modifier: "?",
				Pattern:  "[^\\/#\\?]+?",
			},
			")",
		},
		a{
			a{"/hello(world)", a{"/hello(world)", "hello", "world"}},
			a{"/hello()", a{"/hello()", "hello", ""}},
		},
		a{
			a{m{"foo": "hello", "bar": "world"}, "/hello(world)"},
			a{m{"foo": "hello"}, "/hello()"},
		},
	},
	{
		"/:postType(video|audio|text)(\\+.+)?",
		nil,
		a{
			Token{
				Name:     "postType",
				Prefix:   "/",
				Suffix:   "",
				Modifier: "",
				Pattern:  "video|audio|text",
			},
			Token{
				Name:     0,
				Prefix:   "",
				Suffix:   "",
				Modifier: "?",
				Pattern:  "\\+.+",
			},
		},
		a{
			a{"/video", a{"/video", "video", ""}},
			a{"/video+test", a{"/video+test", "video", "+test"}},
			a{"/video+", nil},
		},
		a{
			a{m{"postType": "video"}, "/video"},
			a{m{"postType": "random"}, nil},
		},
	},
	{
		"/:foo?/:bar?-ext",
		nil,
		a{
			Token{
				Name:     "foo",
				Prefix:   "/",
				Suffix:   "",
				Modifier: "?",
				Pattern:  "[^\\/#\\?]+?",
			},
			Token{
				Name:     "bar",
				Prefix:   "/",
				Suffix:   "",
				Modifier: "?",
				Pattern:  "[^\\/#\\?]+?",
			},
			"-ext",
		},
		a{
			a{"/-ext", nil},
			a{"-ext", a{"-ext", "", ""}},
			a{"/foo-ext", a{"/foo-ext", "foo", ""}},
			a{"/foo/bar-ext", a{"/foo/bar-ext", "foo", "bar"}},
			a{"/foo/-ext", nil},
		},
		a{
			a{m{}, "-ext"},
			a{m{"foo": "foo"}, "/foo-ext"},
			a{m{"bar": "bar"}, "/bar-ext"},
			a{m{"foo": "foo", "bar": "bar"}, "/foo/bar-ext"},
		},
	},
	{
		"/:required/:optional?-ext",
		nil,
		a{
			Token{
				Name:     "required",
				Prefix:   "/",
				Suffix:   "",
				Modifier: "",
				Pattern:  "[^\\/#\\?]+?",
			},
			Token{
				Name:     "optional",
				Prefix:   "/",
				Suffix:   "",
				Modifier: "?",
				Pattern:  "[^\\/#\\?]+?",
			},
			"-ext",
		},
		a{
			a{"/foo-ext", a{"/foo-ext", "foo", ""}},
			a{"/foo/bar-ext", a{"/foo/bar-ext", "foo", "bar"}},
			a{"/foo/-ext", nil},
		},
		a{
			a{m{"required": "foo"}, "/foo-ext"},
		},
	},

	/**
	 * Unicode characters.
	 */
	{
		"/:foo",
		nil,
		a{
			Token{
				Name:     "foo",
				Prefix:   "/",
				Suffix:   "",
				Modifier: "",
				Pattern:  "[^\\/#\\?]+?",
			},
		},
		a{
			a{"/café", a{"/café", "café"}},
		},
		a{
			a{m{"foo": "café"}, "/café"},
			a{m{"foo": "café"}, "/caf%C3%A9", &Options{Encode: encodeURIComponent}},
		},
	},
	{
		"/café",
		nil,
		a{
			"/café",
		},
		a{
			a{"/café", a{"/café"}},
		},
		a{
			a{nil, "/café"},
		},
	},
	{
		"/café",
		&Options{Encode: func(uri string, token interface{}) string {
			return encodeURI(uri)
		}},
		a{
			"/café",
		},
		a{
			a{"/caf%C3%A9", a{"/caf%C3%A9"}},
		},
		a{
			a{nil, "/café"},
		},
	},
	{
		"packages/",
		nil,
		a{
			"packages/",
		},
		a{
			a{"packages", nil},
			a{"packages/", a{"packages/"}},
		},
		a{
			a{nil, "packages/"},
		},
	},

	/**
	 * Hostnames.
	 */
	{
		":domain.com",
		&Options{
			Delimiter: ".",
		},
		a{
			Token{
				Name:     "domain",
				Prefix:   "",
				Suffix:   "",
				Modifier: "",
				Pattern:  "[^\\.]+?",
			},
			".com",
		},
		a{
			a{"example.com", a{"example.com", "example"}},
			a{"github.com", a{"github.com", "github"}},
		},
		a{
			a{m{"domain": "example"}, "example.com"},
			a{m{"domain": "github"}, "github.com"},
		},
	},
	{
		"mail.:domain.com",
		&Options{
			Delimiter: ".",
		},
		a{
			"mail",
			Token{
				Name:     "domain",
				Prefix:   ".",
				Suffix:   "",
				Modifier: "",
				Pattern:  "[^\\.]+?",
			},
			".com",
		},
		a{
			a{"mail.example.com", a{"mail.example.com", "example"}},
			a{"mail.github.com", a{"mail.github.com", "github"}},
		},
		a{
			a{m{"domain": "example"}, "mail.example.com"},
			a{m{"domain": "github"}, "mail.github.com"},
		},
	},
	{
		"example.:ext",
		&Options{},
		a{
			"example",
			Token{
				Name:     "ext",
				Prefix:   ".",
				Suffix:   "",
				Modifier: "",
				Pattern:  "[^\\/#\\?]+?",
			},
		},
		a{
			a{"example.com", a{"example.com", "com"}},
			a{"example.org", a{"example.org", "org"}},
		},
		a{
			a{m{"ext": "com"}, "example.com"},
			a{m{"ext": "org"}, "example.org"},
		},
	},
	{
		"this is",
		&Options{
			Delimiter: " ",
			End:       &falseValue,
		},
		a{
			"this is",
		},
		a{
			a{"this is a test", a{"this is"}},
			a{"this isn't", nil},
		},
		a{
			a{nil, "this is"},
		},
	},

	/**
	 * Ends with.
	 */
	{
		"/test",
		&Options{
			EndsWith: "?",
		},
		a{
			"/test",
		},
		a{
			a{"/test", a{"/test"}},
			a{"/test?query=string", a{"/test"}},
			a{"/test/?query=string", a{"/test/"}},
			a{"/testx", nil},
		},
		a{
			a{nil, "/test"},
		},
	},
	{
		"/test",
		&Options{
			EndsWith: "?",
			Strict:   true,
		},
		a{
			"/test",
		},
		a{
			a{"/test?query=string", a{"/test"}},
			a{"/test/?query=string", nil},
		},
		a{
			a{nil, "/test"},
		},
	},

	/**
	 * Custom prefixes.
	 */
	{
		"{$:foo}{$:bar}?",
		&Options{Prefixes: &prefixDollar},
		a{
			Token{
				Name:     "foo",
				Pattern:  "[^\\/#\\?]+?",
				Prefix:   "$",
				Suffix:   "",
				Modifier: "",
			},
			Token{
				Name:     "bar",
				Pattern:  "[^\\/#\\?]+?",
				Prefix:   "$",
				Suffix:   "",
				Modifier: "?",
			},
		},
		a{
			a{"$x", a{"$x", "x", ""}},
			a{"$x$y", a{"$x$y", "x", "y"}},
		},
		a{
			a{m{"foo": "foo"}, "$foo"},
			a{m{"foo": "foo", "bar": "bar"}, "$foo$bar"},
		},
	},
	{
		"name/:attr1?{-:attr2}?{-:attr3}?",
		&Options{},
		a{
			"name",
			Token{
				Name:     "attr1",
				Pattern:  "[^\\/#\\?]+?",
				Prefix:   "/",
				Suffix:   "",
				Modifier: "?",
			},
			Token{
				Name:     "attr2",
				Pattern:  "[^\\/#\\?]+?",
				Prefix:   "-",
				Suffix:   "",
				Modifier: "?",
			},
			Token{
				Name:     "attr3",
				Pattern:  "[^\\/#\\?]+?",
				Prefix:   "-",
				Suffix:   "",
				Modifier: "?",
			},
		},
		a{
			a{"name/test", a{"name/test", "test", "", ""}},
			a{"name/1", a{"name/1", "1", "", ""}},
			a{"name/1-2", a{"name/1-2", "1", "2", ""}},
			a{"name/1-2-3", a{"name/1-2-3", "1", "2", "3"}},
			a{"name/foo-bar/route", nil},
			a{"name/test/route", nil},
		},
		a{
			a{m{}, "name"},
			a{m{"attr1": "test"}, "name/test"},
			a{m{"attr2": "attr"}, "name-attr"},
		},
	},

	/**
	 * Case-sensitive compile tokensToFunction params.
	 */
	{
		"/:test(abc)",
		&Options{
			Sensitive: true,
		},
		a{
			Token{
				Name:     "test",
				Prefix:   "/",
				Suffix:   "",
				Modifier: "",
				Pattern:  "abc",
			},
		},
		a{
			a{"/abc", a{"/abc", "abc"}},
			a{"/ABC", nil},
		},
		a{
			a{m{"test": "abc"}, "/abc"},
			a{m{"test": "ABC"}, nil},
		},
	},
	{
		"/:test(abc)",
		&Options{},
		a{
			Token{
				Name:     "test",
				Prefix:   "/",
				Suffix:   "",
				Modifier: "",
				Pattern:  "abc",
			},
		},
		a{
			a{"/abc", a{"/abc", "abc"}},
			a{"/ABC", a{"/ABC", "ABC"}},
		},
		a{
			a{m{"test": "abc"}, "/abc"},
			a{m{"test": "ABC"}, "/ABC"},
		},
	},

	/**
	 * Nested parentheses.
	 */
	{
		"/:test(\\d+(?:\\.\\d+)?)",
		nil,
		a{
			Token{
				Name:     "test",
				Prefix:   "/",
				Suffix:   "",
				Modifier: "",
				Pattern:  "\\d+(?:\\.\\d+)?",
			},
		},
		a{
			a{"/123", a{"/123", "123"}},
			a{"/abc", nil},
			a{"/123/abc", nil},
			a{"/123.123", a{"/123.123", "123.123"}},
			a{"/123.abc", nil},
		},
		a{
			a{m{"test": 123}, "/123"},
			a{m{"test": 123.123}, "/123.123"},
			a{m{"test": "abc"}, nil},
			a{m{"test": "123"}, "/123"},
			a{m{"test": "123.123"}, "/123.123"},
			a{m{"test": "123.abc"}, nil},
		},
	},
	{
		"/:test((?!login)[^/]+)",
		nil,
		a{
			Token{
				Name:     "test",
				Prefix:   "/",
				Suffix:   "",
				Modifier: "",
				Pattern:  "(?!login)[^/]+",
			},
		},
		a{
			a{"/route", a{"/route", "route"}},
			a{"/login", nil},
		},
		a{
			a{m{"test": "route"}, "/route"},
			a{m{"test": "login"}, nil},
		},
	},
	{
		"/user(s)?/:user",
		nil,
		a{
			"/user",
			Token{
				Name:     0,
				Prefix:   "",
				Suffix:   "",
				Modifier: "?",
				Pattern:  "s",
			},
			Token{
				Name:     "user",
				Prefix:   "/",
				Suffix:   "",
				Modifier: "",
				Pattern:  "[^\\/#\\?]+?",
			},
		},
		a{
			a{"/user/123", a{"/user/123", "", "123"}},
			a{"/users/123", a{"/users/123", "s", "123"}},
		},
		a{
			a{m{"user": "123"}, "/user/123"},
		},
	},

	{
		"/whatever/:foo\\?query=str",
		nil,
		a{
			"/whatever",
			Token{
				Name:     "foo",
				Prefix:   "/",
				Suffix:   "",
				Modifier: "",
				Pattern:  "[^\\/#\\?]+?",
			},
			"?query=str",
		},
		a{a{"/whatever/123?query=str", a{"/whatever/123?query=str", "123"}}},
		a{a{m{"foo": "123"}, "/whatever/123?query=str"}},
	},
	{
		"/whatever/:foo",
		&Options{End: &falseValue},
		a{
			"/whatever",
			Token{
				Name:     "foo",
				Prefix:   "/",
				Suffix:   "",
				Modifier: "",
				Pattern:  "[^\\/#\\?]+?",
			},
		},
		a{
			a{"/whatever/123", a{"/whatever/123", "123"}},
			a{"/whatever/123/path", a{"/whatever/123", "123"}},
			a{"/whatever/123#fragment", a{"/whatever/123", "123"}},
			a{"/whatever/123?query=str", a{"/whatever/123", "123"}},
		},
		a{
			a{m{"foo": "123"}, "/whatever/123"},
			a{m{"foo": "#"}, nil},
		},
	},
}

// Dynamically generate the entire test suite.
func TestPathToRegexp(t *testing.T) {
	testPath := "/user/:id"

	testParam := Token{
		Name:     "id",
		Prefix:   "/",
		Suffix:   "",
		Modifier: "",
		Pattern:  "[^\\/#\\?]+?",
	}

	t.Run("arguments", func(t *testing.T) {
		t.Run("should work without different call combinations", func(t *testing.T) {
			_, err := PathToRegexp("/test", nil, nil)
			if err != nil {
				t.Error(err)
			}
			_, err = PathToRegexp("/test", &[]Token{}, nil)
			if err != nil {
				t.Error(err)
			}
			_, err = PathToRegexp("/test", nil, &Options{})
			if err != nil {
				t.Error(err)
			}

			_, err = PathToRegexp(regexp2.MustCompile("^\\/test", regexp2.None), nil, nil)
			if err != nil {
				t.Error(err)
			}
			_, err = PathToRegexp(regexp2.MustCompile("^\\/test", regexp2.None), &[]Token{}, nil)
			if err != nil {
				t.Error(err)
			}
			_, err = PathToRegexp(regexp2.MustCompile("^\\/test", regexp2.None), nil, &Options{})
			if err != nil {
				t.Error(err)
			}

			_, err = PathToRegexp([]string{"/a", "/b"}, nil, nil)
			if err != nil {
				t.Error(err)
			}
			_, err = PathToRegexp([]string{"/a", "/b"}, &[]Token{}, nil)
			if err != nil {
				t.Error(err)
			}
			_, err = PathToRegexp([]string{"/a", "/b"}, nil, &Options{})
			if err != nil {
				t.Error(err)
			}
		})

		t.Run("should accept an array of tokens as the second argument", func(t *testing.T) {
			tokens := &[]Token{}
			r, err := PathToRegexp(testPath, tokens, &Options{End: &falseValue})
			if err != nil {
				t.Fatal(err)
			}
			var expect interface{}
			expect = &[]Token{testParam}

			if !reflect.DeepEqual(tokens, expect) {
				t.Errorf(testErrorFormat, tokens, expect)
			}

			expect = []string{"/user/123", "123"}
			if !reflect.DeepEqual(exec(r, "/user/123/show"), expect) {
				t.Errorf(testErrorFormat, tokens, expect)
			}
		})

		t.Run("should throw on non-capturing pattern", func(t *testing.T) {
			_, err := PathToRegexp("/:foo(?:\\d+(\\.\\d+)?)", nil, nil)
			expect := errors.New(`pattern cannot start with "?" at 6`)
			if !reflect.DeepEqual(err, expect) {
				t.Errorf(testErrorFormat, err, expect)
			}
		})

		t.Run("should throw on nested capturing group", func(t *testing.T) {
			_, err := PathToRegexp("/:foo(\\d+(\\.\\d+)?)", nil, nil)
			expect := errors.New("capturing groups are not allowed at 9")
			if !reflect.DeepEqual(err, expect) {
				t.Errorf(testErrorFormat, err, expect)
			}
		})

		t.Run("should throw on unbalanced pattern", func(t *testing.T) {
			_, err := PathToRegexp("/:foo(abc", nil, nil)
			expect := errors.New("unbalanced pattern at 5")
			if !reflect.DeepEqual(err, expect) {
				t.Errorf(testErrorFormat, err, expect)
			}
		})

		t.Run("should throw on missing pattern", func(t *testing.T) {
			_, err := PathToRegexp("/:foo()", nil, nil)
			expect := errors.New("missing pattern at 5")
			if !reflect.DeepEqual(err, expect) {
				t.Errorf(testErrorFormat, err, expect)
			}
		})

		t.Run("should throw on missing name", func(t *testing.T) {
			_, err := PathToRegexp("/:(test)", nil, nil)
			expect := errors.New("missing parameter name at 1")
			if !reflect.DeepEqual(err, expect) {
				t.Errorf(testErrorFormat, err, expect)
			}
		})

		t.Run("should throw on nested groups", func(t *testing.T) {
			_, err := PathToRegexp("/{a{b:foo}}", nil, nil)
			expect := fmt.Errorf("unexpected %d at 3, expected %d", modeOpen, modeClose)
			if !reflect.DeepEqual(err, expect) {
				t.Errorf(testErrorFormat, err, expect)
			}
		})

		t.Run("should throw on misplaced modifier", func(t *testing.T) {
			_, err := PathToRegexp("/foo?", nil, nil)
			expect := fmt.Errorf("unexpected %d at 4, expected %d", modeModifier, modeEnd)
			if !reflect.DeepEqual(err, expect) {
				t.Errorf(testErrorFormat, err, expect)
			}
		})
	})

	t.Run("tokens", func(t *testing.T) {
		tokens, err := Parse(testPath, nil)
		if err != nil {
			t.Fatal(err)
		}
		t.Run("should expose method to compile tokens to regexp", func(t *testing.T) {
			r, err := tokensToRegExp(tokens, nil, nil)
			if err != nil {
				t.Fatal(err)
			}
			expected := []string{"/user/123", "123"}
			result := exec(r, "/user/123")
			if !reflect.DeepEqual(result, expected) {
				t.Errorf(testErrorFormat, result, expected)
			}
		})
		t.Run("should expose method to compile tokens to a path function", func(t *testing.T) {
			fn, err := tokensToFunction(tokens, nil)
			if err != nil {
				t.Fatal(err)
			}
			expected := "/user/123"
			result, err := fn(map[interface{}]interface{}{"id": 123})
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(result, expected) {
				t.Errorf(testErrorFormat, result, expected)
			}
		})
	})

	t.Run("rules", func(t *testing.T) {
		for _, test := range tests {
			path, opts, rawTokens := test[0], test[1], test[2].(a)
			matchCases, compileCases := test[3].(a), test[4].(a)
			t.Run(inspect(path), func(t *testing.T) {
				tokens := &[]Token{}
				var o *Options
				if opts != nil {
					o = opts.(*Options)
				}
				r, err := PathToRegexp(path, tokens, o)
				if err != nil {
					t.Fatal(err)
				}
				// Parsing and compiling is only supported with string input.
				if path, ok := path.(string); ok {
					t.Run("should parse", func(t *testing.T) {
						parsedTokens, err := Parse(path, o)
						if err != nil {
							t.Fatal(err)
						}
						result := a(parsedTokens)
						if !reflect.DeepEqual(result, rawTokens) {
							t.Errorf(testErrorFormat, result, rawTokens)
						}
					})
					t.Run("compile", func(t *testing.T) {
						for _, v := range compileCases {
							io := v.(a)
							params, result := io[0], io[1]
							var o1 *Options
							if len(io) >= 3 && io[2] != nil {
								o1 = io[2].(*Options)
							}
							toPath, err := Compile(path, mergeOptions(o, o1))
							if err != nil {
								t.Fatal(err)
							}
							if result != nil {
								t.Run("should compile using "+inspect(params), func(t *testing.T) {
									r, err := toPath(params)
									if err != nil {
										t.Fatal(err)
									}
									if !reflect.DeepEqual(r, result) {
										t.Errorf(testErrorFormat, result, path)
									}
								})
							} else {
								t.Run("should not compile using "+inspect(params), func(t *testing.T) {
									_, err := toPath(params)
									if err == nil {
										t.Errorf(testErrorFormat, err, "error")
									}
								})
							}
						}
					})
				} else {
					t.Run("should parse tokens", func(t *testing.T) {
						tTokens := make([]interface{}, 0, len(rawTokens))
						for _, token := range rawTokens {
							if _, ok := token.(string); !ok {
								tTokens = append(tTokens, token)
							}
						}
						if !tokensDeepEqual(*tokens, tTokens) {
							t.Errorf(testErrorFormat, *tokens, tTokens)
						}
					})
				}

				name := ""
				if opts != nil {
					name = " using " + inspect(opts)
				}
				t.Run("match"+name, func(t *testing.T) {
					for _, v := range matchCases {
						io := v.(a)
						pathname, matches := io[0], io[1]
						message := " not "
						var o a
						if matches != nil {
							message = " "
							o = matches.(a)
						}
						message = "should" + message + "match " + inspect(pathname)
						t.Run(message, func(t *testing.T) {
							result := exec(r, pathname.(string))
							if !deepEqual(result, o) {
								t.Errorf(testErrorFormat, result, matches)
							}
						})

						var params *MatchResult
						if len(io) >= 3 && io[2] != nil {
							params = io[2].(*MatchResult)
						}

						var options *Options
						if len(io) >= 4 && io[3] != nil {
							options = io[3].(*Options)
						}

						if path, ok := path.(string); ok && params != nil {
							match := MustMatch(path, options)
							t.Run(message+" params", func(t *testing.T) {
								m, err := match(pathname.(string))
								if err != nil {
									t.Fatal(err)
								}
								if !params.equals(m) {
									t.Errorf(testErrorFormat, m, params)
								}
							})
						}
					}
				})
			})
		}
	})

	t.Run("compile errors", func(t *testing.T) {
		t.Run("should throw when a required param is undefined", func(t *testing.T) {
			toPath, err := Compile("/a/:b/c", nil)
			if err != nil {
				t.Fatal(err)
			}
			_, err = toPath(nil)
			expect := errors.New(`expected "b" to be a string`)
			if !reflect.DeepEqual(err, expect) {
				t.Errorf(testErrorFormat, err, expect)
			}
		})

		t.Run("should throw when it does not match the pattern", func(t *testing.T) {
			toPath, err := Compile("/:foo(\\d+)", nil)
			if err != nil {
				t.Fatal(err)
			}
			_, err = toPath(map[interface{}]interface{}{"foo": "abc"})
			expect := errors.New(`expected "foo" to match "\d+", but got "abc"`)
			if !reflect.DeepEqual(err, expect) {
				t.Errorf(testErrorFormat, err, expect)
			}
		})

		t.Run("should throw when expecting a repeated value", func(t *testing.T) {
			toPath, err := Compile("/:foo+", nil)
			if err != nil {
				t.Fatal(err)
			}
			_, err = toPath(map[interface{}]interface{}{"foo": []interface{}{}})
			expect := errors.New(`expected "foo" to not be empty`)
			if !reflect.DeepEqual(err, expect) {
				t.Errorf(testErrorFormat, err, expect)
			}
		})

		t.Run("should throw when not expecting a repeated value", func(t *testing.T) {
			toPath, err := Compile("/:foo", nil)
			if err != nil {
				t.Fatal(err)
			}
			_, err = toPath(map[interface{}]interface{}{"foo": []interface{}{}})
			expect := errors.New(`expected "foo" to not repeat, but got array`)
			if !reflect.DeepEqual(err, expect) {
				t.Errorf(testErrorFormat, err, expect)
			}
		})

		t.Run("should throw when repeated value does not match", func(t *testing.T) {
			toPath, err := Compile("/:foo(\\d+)+", nil)
			if err != nil {
				t.Fatal(err)
			}
			_, err = toPath(map[interface{}]interface{}{"foo": []interface{}{1, 2, 3, "a"}})
			expect := errors.New(`expected all "foo" to match "\d+"`)
			if !reflect.DeepEqual(err, expect) {
				t.Errorf(testErrorFormat, err, expect)
			}
		})
	})

	t.Run("path should be string, or strings, or a regular expression", func(t *testing.T) {
		_, err := PathToRegexp(123, nil, nil)
		if err == nil {
			t.Errorf(testErrorFormat, err, "none nil error")
		}
	})
}

func TestMustCompile(t *testing.T) {
	r := MustCompile("/user/:id(\\d+)", nil)
	if r == nil {
		t.Errorf(testErrorFormat, r, "func")
	}
}

func TestDecodeURI(t *testing.T) {
	tests := map[string]string{
		"%3B%2F%3F%3A%40%26%3D%2B%24%2C%23": "%3B%2F%3F%3A%40%26%3D%2B%24%2C%23",
		"http%3A%2F%2Fwww.example.com%2Fstring%20with%20%2B%20and%20%3F%20and%20%26%20and%20spaces": "http%3A%2F%2Fwww.example.com%2Fstring with %2B and %3F and %26 and spaces",
		"https://developer.mozilla.org/ru/docs/JavaScript_%D1%88%D0%B5%D0%BB%D0%BB%D1%8B":           "https://developer.mozilla.org/ru/docs/JavaScript_шеллы",
	}
	for k, v := range tests {
		result, err := decodeURI(k)
		if err != nil {
			t.Error(err)
			continue
		}
		if result != v {
			t.Errorf(testErrorFormat, result, v)
		}
	}

	t.Run("malformed URI sequence", func(t *testing.T) {
		_, err := decodeURI("%E0%A4%A")
		if err == nil {
			t.Errorf(testErrorFormat, err, "error")
		}
	})
}

func TestAnyString(t *testing.T) {
	tests := map[string][]string{
		"foo": {"", "", "foo", ""},
		"bar": {"bar", "", "foo", ""},
		"":    {"", "", "", ""},
	}
	for k, v := range tests {
		result := anyString(v...)
		if result != k {
			t.Errorf(testErrorFormat, result, k)
		}
	}
}

func TestQuote(t *testing.T) {
	tests := map[string]string{
		"foo":   "`foo`",
		"`bar`": "\"`bar`\"",
	}
	for k, v := range tests {
		result := quote(k)
		if result != v {
			t.Errorf(testErrorFormat, result, v)
		}
	}
}

func TestMust(t *testing.T) {
	r := regexp2.MustCompile("^\\/([^\\/]+)$", regexp2.None)
	result := Must(r, nil)
	if result != r {
		t.Errorf(testErrorFormat, result, r)
	}

	t.Run("non nil error", func(t *testing.T) {
		defer func() {
			if err := recover(); err == nil {
				t.Errorf(testErrorFormat, err, "error")
			}
		}()
		result = Must(r, errors.New("error"))
	})
}

func BenchmarkPathToRegexp(b *testing.B) {
	b.Run("string", func(b *testing.B) {
		b.Run("no parameters", func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				PathToRegexp("/foo", &[]Token{}, nil)
			}
		})
		b.Run("named parameters", func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				PathToRegexp("/foo/:bar", &[]Token{}, nil)
			}
		})
		b.Run("unnamed parameters", func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				PathToRegexp("/:foo/(.*)", &[]Token{}, nil)
			}
		})
		b.Run("optional parameters", func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				PathToRegexp("/:foo/:bar?", &[]Token{}, nil)
			}
		})
	})

	b.Run("regexp", func(b *testing.B) {
		b.Run("simple", func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				PathToRegexp(regexp2.MustCompile("^/foo/\\d+", regexp2.None), &[]Token{}, nil)
			}
		})
		b.Run("complex", func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				PathToRegexp(regexp2.MustCompile("^/foo/\\d+(?:\\.\\d+)?", regexp2.None), &[]Token{}, nil)
			}
		})
	})

	b.Run("array", func(b *testing.B) {
		b.Run("simple", func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				PathToRegexp([]string{"/foo", "/foo/bar"}, &[]Token{}, nil)
			}
		})
		b.Run("complex", func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				PathToRegexp([]string{"/foo/:bar", "/:foo/:bar?"}, &[]Token{}, nil)
			}
		})
	})

	b.Run("with end false", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			PathToRegexp("/foo/:bar", &[]Token{}, &Options{End: &falseValue})
		}
	})
}

func BenchmarkParse(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Parse("/foo/:bar/(.*)", nil)
	}
}

func BenchmarkCompile(b *testing.B) {
	b.Run("simple", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			Compile("/foo/:bar", nil)
		}
	})
	b.Run("complex", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			Compile("/foo/:bar(\\d+)", nil)
		}
	})
}

func BenchmarkMatch(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Match("/foo/:bar", nil)
	}
}

func exec(r *regexp2.Regexp, str string) []string {
	m, _ := r.FindStringMatch(str)
	if m == nil {
		return nil
	}

	result := make([]string, m.GroupCount())
	for i, g := range m.Groups() {
		result[i] = g.String()
	}
	return result
}

func inspect(v interface{}) string {
	return fmt.Sprintf("%v", v)
}

func deepEqual(p1 []string, p2 a) bool {
	if p1 == nil && p2 == nil {
		return true
	}
	if p1 == nil || p2 == nil {
		return false
	}

	for i, v := range p1 {
		if !reflect.DeepEqual(v, p2[i]) {
			return false
		}
	}

	return true
}

func tokensDeepEqual(t1 []Token, t2 []interface{}) bool {
	if len(t1) != len(t2) {
		return false
	}

	if len(t1) == 0 && len(t2) == 0 {
		return true
	}

	if t1 == nil || t2 == nil {
		return false
	}

	for i, v := range t1 {
		if !reflect.DeepEqual(v, t2[i]) {
			return false
		}
	}

	return true
}

func (m *MatchResult) equals(o *MatchResult) bool {
	result := m.Path == o.Path && m.Index == m.Index

	if result {
		for k, v := range m.Params {
			if reflect.TypeOf(v).Kind() == reflect.Slice {
				arr1 := toSlice(v)
				arr2 := toSlice(o.Params[k])
				if !reflect.DeepEqual(arr1, arr2) {
					return false
				}
			} else if !reflect.DeepEqual(v, o.Params[k]) {
				return false
			}
		}
	}

	return result
}

func mergeOptions(o1 *Options, o2 *Options) *Options {
	if o1 == nil {
		return o2
	}

	if o2 == nil {
		return o1
	}

	end := o1.End
	if o2.End != nil {
		end = o2.End
	}

	start := o1.Start
	if o2.Start != nil {
		start = o2.Start
	}

	validate := o1.Validate
	if o2.Validate != nil {
		validate = o2.Validate
	}

	endsWith := o1.EndsWith
	if o2.EndsWith != "" {
		endsWith = o2.EndsWith
	}

	encode := o1.Encode
	if o2.Encode != nil {
		encode = o2.Encode
	}

	decode := o1.Decode
	if o2.Decode != nil {
		decode = o2.Decode
	}

	return &Options{
		Sensitive: o2.Sensitive,
		Strict:    o2.Strict,
		End:       end,
		Start:     start,
		Validate:  validate,
		Delimiter: o2.Delimiter,
		EndsWith:  endsWith,
		Encode:    encode,
		Decode:    decode,
	}
}
