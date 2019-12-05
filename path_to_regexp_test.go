// Copyright 2019 Guoyao Wu. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package pathtoregexp

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/dlclark/regexp2"
)

type a []interface{}
type m map[interface{}]interface{}

var falseValue = false

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
			a{m{}, "/"}, a{m{"id": 123}, "/"},
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
				Name:      "test",
				Prefix:    "/",
				Delimiter: "/",
				Optional:  false,
				Repeat:    false,
				Pattern:   "[^\\/]+?",
			},
		},
		a{
			a{
				"/route",
				a{"/route", "route"},
				&MatchResult{Path: "/route", Index: 0, Params: m{"test": "route"}},
			},
		},
		a{
			a{m{}, nil},
			a{m{"test": "abc"}, "/abc"},
			a{
				m{"test": "a+b"},
				"/a+b",
				&Options{
					Encode: func(uri string, token interface{}) string {
						return uri
					},
				},
			},
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
			a{m{"test": "a+b"}, "/a%2Bb"},
		},
	},
	{
		"/:test/",
		&Options{
			End: &falseValue,
		},
		a{
			Token{
				Name:      "test",
				Prefix:    "/",
				Delimiter: "/",
				Optional:  false,
				Repeat:    false,
				Pattern:   "[^\\/]+?",
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
				Name:      "test",
				Prefix:    "/",
				Delimiter: "/",
				Optional:  false,
				Repeat:    false,
				Pattern:   "[^\\/]+?",
			},
		},
		a{
			a{"/route", a{"/route", "route"}},
		},
		a{
			a{m{}, nil},
			a{m{"test": "abc"}, "/abc"},
			a{
				m{"test": "a+b"},
				"/a+b",
				&Options{
					Encode: func(uri string, token interface{}) string {
						return uri
					},
				},
			},
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
			a{m{"test": "a+b"}, "/a%2Bb"},
		},
	},
	{
		"/:test/",
		&Options{
			Start: &falseValue,
		},
		a{
			Token{
				Name:      "test",
				Prefix:    "/",
				Delimiter: "/",
				Optional:  false,
				Repeat:    false,
				Pattern:   "[^\\/]+?",
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
				Name:      "test",
				Prefix:    "/",
				Delimiter: "/",
				Optional:  false,
				Repeat:    false,
				Pattern:   "[^\\/]+?",
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
				Name:      "test",
				Prefix:    "/",
				Delimiter: "/",
				Optional:  false,
				Repeat:    false,
				Pattern:   "[^\\/]+?",
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
				Name:      "test",
				Prefix:    "/",
				Delimiter: "/",
				Optional:  false,
				Repeat:    false,
				Pattern:   "[^\\/]+?",
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
				Name:      "test",
				Prefix:    "/",
				Delimiter: "/",
				Optional:  false,
				Repeat:    false,
				Pattern:   "[^\\/]+?",
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
				Name:      "test",
				Prefix:    "/",
				Delimiter: "/",
				Optional:  false,
				Repeat:    false,
				Pattern:   "[^\\/]+?",
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
			a{m{"test": "something/else"}, "/something%2Felse"},
			a{m{"test": "something/else/more"}, "/something%2Felse%2Fmore"},
		},
	},
	{
		"/:test",
		&Options{
			Strict: true,
		},
		a{
			Token{
				Name:      "test",
				Prefix:    "/",
				Delimiter: "/",
				Optional:  false,
				Repeat:    false,
				Pattern:   "[^\\/]+?",
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
				Name:      "test",
				Prefix:    "/",
				Delimiter: "/",
				Optional:  false,
				Repeat:    false,
				Pattern:   "[^\\/]+?",
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
				Name:      "test",
				Prefix:    "/",
				Delimiter: "/",
				Optional:  false,
				Repeat:    false,
				Pattern:   "[^\\/]+?",
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
				Name:      "test",
				Prefix:    "/",
				Delimiter: "/",
				Optional:  true,
				Repeat:    false,
				Pattern:   "[^\\/]+?",
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
				Name:      "test",
				Prefix:    "/",
				Delimiter: "/",
				Optional:  true,
				Repeat:    false,
				Pattern:   "[^\\/]+?",
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
				Name:      "test",
				Prefix:    "/",
				Delimiter: "/",
				Optional:  true,
				Repeat:    false,
				Pattern:   "[^\\/]+?",
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
				Name:      "test",
				Prefix:    "/",
				Delimiter: "/",
				Optional:  true,
				Repeat:    false,
				Pattern:   "[^\\/]+?",
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
				Name:      "test",
				Prefix:    "/",
				Delimiter: "/",
				Optional:  true,
				Repeat:    false,
				Pattern:   "[^\\/]+?",
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
				Name:      "test",
				Prefix:    "/",
				Delimiter: "/",
				Optional:  true,
				Repeat:    true,
				Pattern:   "[^\\/]+?",
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
				Name:      "test",
				Prefix:    "/",
				Delimiter: "/",
				Optional:  false,
				Repeat:    true,
				Pattern:   "[^\\/]+?",
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
				Name:      "test",
				Prefix:    "/",
				Delimiter: "/",
				Optional:  false,
				Repeat:    true,
				Pattern:   "\\d+",
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
				Name:      "ext",
				Prefix:    ".",
				Delimiter: ".",
				Optional:  false,
				Repeat:    true,
				Pattern:   "json|xml",
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
		"/route.:ext/test",
		nil,
		a{
			"/route",
			Token{
				Name:      "ext",
				Prefix:    ".",
				Delimiter: ".",
				Optional:  false,
				Repeat:    false,
				Pattern:   "[^\\.\\/]+?",
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
				Name:      "test",
				Prefix:    "/",
				Delimiter: "/",
				Optional:  true,
				Repeat:    true,
				Pattern:   "[^\\/]+?",
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
				Name:      "ext",
				Prefix:    ".",
				Delimiter: ".",
				Optional:  true,
				Repeat:    true,
				Pattern:   "[a-z]+",
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
				Name:      "test",
				Prefix:    "/",
				Delimiter: "/",
				Optional:  false,
				Repeat:    false,
				Pattern:   "\\d+",
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
				Name:      "test",
				Prefix:    "/",
				Delimiter: "/",
				Optional:  false,
				Repeat:    false,
				Pattern:   "\\d+",
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
				Name:      "test",
				Prefix:    "/",
				Delimiter: "/",
				Optional:  false,
				Repeat:    false,
				Pattern:   ".*",
			},
		},
		a{
			a{"/anything/goes/here", a{"/anything/goes/here", "anything/goes/here"}},
			a{"/;,:@&=/+$-_.!/~*()", a{"/;,:@&=/+$-_.!/~*()", ";,:@&=/+$-_.!/~*()"}},
		},
		a{
			a{m{"test": ""}, "/"},
			a{m{"test": "abc"}, "/abc"},
			a{m{"test": "abc/123"}, "/abc%2F123"},
			a{m{"test": "abc/123/456"}, "/abc%2F123%2F456"},
		},
	},
	{
		"/:route([a-z]+)",
		nil,
		a{
			Token{
				Name:      "route",
				Prefix:    "/",
				Delimiter: "/",
				Optional:  false,
				Repeat:    false,
				Pattern:   "[a-z]+",
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
				Name:      "route",
				Prefix:    "/",
				Delimiter: "/",
				Optional:  false,
				Repeat:    false,
				Pattern:   "this|that",
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
				Name:      "path",
				Prefix:    "/",
				Delimiter: "/",
				Optional:  true,
				Repeat:    true,
				Pattern:   "abc|xyz",
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
				Name:      "test",
				Prefix:    "",
				Delimiter: "/",
				Optional:  false,
				Repeat:    false,
				Pattern:   "[^\\/]+?",
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
				Name:      "test",
				Prefix:    "",
				Delimiter: "/",
				Optional:  false,
				Repeat:    false,
				Pattern:   "[^\\/]+?",
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
				Name:      "test",
				Prefix:    "",
				Delimiter: "/",
				Optional:  false,
				Repeat:    false,
				Pattern:   "[^\\/]+?",
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
				Name:      "test",
				Prefix:    "",
				Delimiter: "/",
				Optional:  true,
				Repeat:    false,
				Pattern:   "[^\\/]+?",
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
		":test+",
		nil,
		a{
			Token{
				Name:      "test",
				Prefix:    "",
				Delimiter: "/",
				Optional:  false,
				Repeat:    true,
				Pattern:   "[^\\/]+?",
			},
		},
		a{
			a{"route", a{"route", "route"}},
			a{"/route", nil}, a{"", nil},
			a{"foo/bar", a{"foo/bar", "foo/bar"}},
		},
		a{
			a{m{}, nil},
			a{m{"test": ""}, nil},
			a{m{"test": a{"route"}}, "route"},
			a{m{"test": a{"foo", "bar"}}, "foo/bar"},
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
				Name:      "test",
				Prefix:    "/",
				Delimiter: "/",
				Optional:  false,
				Repeat:    false,
				Pattern:   "[^\\/]+?",
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
		"/test.:format",
		nil,
		a{
			"/test",
			Token{
				Name:      "format",
				Prefix:    ".",
				Delimiter: ".",
				Optional:  false,
				Repeat:    false,
				Pattern:   "[^\\.\\/]+?",
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
		"/test.:format.:format",
		nil,
		a{
			"/test",
			Token{
				Name:      "format",
				Prefix:    ".",
				Delimiter: ".",
				Optional:  false,
				Repeat:    false,
				Pattern:   "[^\\.\\/]+?",
			},
			Token{
				Name:      "format",
				Prefix:    ".",
				Delimiter: ".",
				Optional:  false,
				Repeat:    false,
				Pattern:   "[^\\.\\/]+?",
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
		"/test.:format+",
		nil,
		a{
			"/test",
			Token{
				Name:      "format",
				Prefix:    ".",
				Delimiter: ".",
				Optional:  false,
				Repeat:    true,
				Pattern:   "[^\\.\\/]+?",
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
		"/test.:format",
		&Options{
			End: &falseValue,
		},
		a{
			"/test",
			Token{
				Name:      "format",
				Prefix:    ".",
				Delimiter: ".",
				Optional:  false,
				Repeat:    false,
				Pattern:   "[^\\.\\/]+?",
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
				Name:      "format",
				Prefix:    ".",
				Delimiter: ".",
				Optional:  false,
				Repeat:    false,
				Pattern:   "[^\\.\\/]+?",
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
				Name:      "test",
				Prefix:    "/",
				Delimiter: "/",
				Optional:  false,
				Repeat:    false,
				Pattern:   "[^\\/]+?",
			},
			Token{
				Name:      "format",
				Prefix:    ".",
				Delimiter: ".",
				Optional:  false,
				Repeat:    false,
				Pattern:   "[^\\.\\/]+?",
			},
		},
		a{
			a{"/route.html", a{"/route.html", "route", "html"}},
			a{"route", nil},
			a{"/route.html.json", a{"/route.html.json", "route.html", "json"}},
		},
		a{
			a{m{}, nil},
			a{m{"test": "route", "format": "foo"}, "/route.foo"},
		},
	},
	{
		"/:test.:format?",
		nil,
		a{
			Token{
				Name:      "test",
				Prefix:    "/",
				Delimiter: "/",
				Optional:  false,
				Repeat:    false,
				Pattern:   "[^\\/]+?",
			},
			Token{
				Name:      "format",
				Prefix:    ".",
				Delimiter: ".",
				Optional:  true,
				Repeat:    false,
				Pattern:   "[^\\.\\/]+?",
			},
		},
		a{
			a{"/route", a{"/route", "route", ""}},
			a{"/route.json", a{"/route.json", "route", "json"}},
			a{"/route.json.html", a{"/route.json.html", "route.json", "html"}},
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
				Name:      "test",
				Prefix:    "/",
				Delimiter: "/",
				Optional:  false,
				Repeat:    false,
				Pattern:   "[^\\/]+?",
			},
			Token{
				Name:      "format",
				Prefix:    ".",
				Delimiter: ".",
				Optional:  true,
				Repeat:    false,
				Pattern:   "[^\\.\\/]+?",
			},
		},
		a{
			a{"/route", a{"/route", "route", ""}},
			a{"/route.json", a{"/route.json", "route", "json"}},
			a{"/route.json.html", a{"/route.json.html", "route.json", "html"}},
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
				Name:      "format",
				Prefix:    ".",
				Delimiter: ".",
				Optional:  false,
				Repeat:    false,
				Pattern:   ".*",
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
				Name:      0,
				Prefix:    "/",
				Delimiter: "/",
				Optional:  false,
				Repeat:    false,
				Pattern:   "\\d+",
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
				Name:      0,
				Prefix:    "/",
				Delimiter: "/",
				Optional:  false,
				Repeat:    false,
				Pattern:   "\\d+",
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
				Name:      0,
				Prefix:    "/",
				Delimiter: "/",
				Optional:  true,
				Repeat:    false,
				Pattern:   "\\d+",
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
				Name:      0,
				Prefix:    "/",
				Delimiter: "/",
				Optional:  false,
				Repeat:    false,
				Pattern:   ".*",
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
				Name:      0,
				Prefix:    "",
				Delimiter: "/",
				Optional:  false,
				Repeat:    false,
				Pattern:   "\\d+\\\\",
			},
			")",
		},
		a{
			a{"/route(\\123\\)", a{"/route(\\123\\)", "123\\"}},
		},
		a{},
	},
	{
		"/(login)?",
		nil,
		a{
			Token{
				Name:      0,
				Prefix:    "/",
				Delimiter: "/",
				Optional:  true,
				Repeat:    false,
				Pattern:   "login",
			},
		},
		a{
			a{"/", a{"/", ""}},
			a{"/login", a{"/login", "login"}},
		},
		a{},
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
				Name:      0,
				Prefix:    "",
				Delimiter: "",
				Optional:  false,
				Repeat:    false,
				Pattern:   "",
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
				Name:      0,
				Prefix:    "",
				Delimiter: "",
				Optional:  false,
				Repeat:    false,
				Pattern:   "",
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
				Name:      0,
				Prefix:    "",
				Delimiter: "",
				Optional:  false,
				Repeat:    false,
				Pattern:   "",
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
				Name:      "test",
				Prefix:    "/",
				Delimiter: "/",
				Optional:  false,
				Repeat:    false,
				Pattern:   "\\d+",
			},
			Token{
				Name:      0,
				Prefix:    "",
				Delimiter: "",
				Optional:  false,
				Repeat:    false,
				Pattern:   "",
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
				Name:      "test",
				Prefix:    "/",
				Delimiter: "/",
				Optional:  false,
				Repeat:    false,
				Pattern:   "[^\\/]+?",
			},
			Token{
				Name:      "test",
				Prefix:    "/",
				Delimiter: "/",
				Optional:  false,
				Repeat:    false,
				Pattern:   "[^\\/]+?",
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
				Name:      0,
				Prefix:    "",
				Delimiter: "",
				Optional:  false,
				Repeat:    false,
				Pattern:   "",
			},
			Token{
				Name:      0,
				Prefix:    "",
				Delimiter: "",
				Optional:  false,
				Repeat:    false,
				Pattern:   "",
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
		"/.+*?=^!:${}[]|",
		nil,
		a{
			"/.+*?=^!:${}[]|",
		},
		a{
			a{"/.+*?=^!:${}[]|", a{"/.+*?=^!:${}[]|"}},
		},
		a{
			a{nil, "/.+*?=^!:${}[]|"},
		},
	},
	{
		"/test\\/:uid(u\\d+)?:cid(c\\d+)?",
		nil,
		a{
			"/test/",
			Token{
				Name:      "uid",
				Prefix:    "",
				Delimiter: "/",
				Optional:  true,
				Repeat:    false,
				Pattern:   "u\\d+",
			},
			Token{
				Name:      "cid",
				Prefix:    "",
				Delimiter: "/",
				Optional:  true,
				Repeat:    false,
				Pattern:   "c\\d+",
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
		"\\/(apple-)?icon-:res(\\d+).png",
		nil,
		a{
			"/",
			Token{
				Name:      0,
				Prefix:    "",
				Delimiter: "/",
				Optional:  true,
				Repeat:    false,
				Pattern:   "apple-",
			},
			"icon",
			Token{
				Name:      "res",
				Prefix:    "-",
				Delimiter: "-",
				Optional:  false,
				Repeat:    false,
				Pattern:   "\\d+",
			},
			".png",
		},
		a{
			a{"/icon-240.png", a{"/icon-240.png", "", "240"}},
			a{"/apple-icon-240.png", a{"/apple-icon-240.png", "apple-", "240"}},
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
				Name:      "foo",
				Prefix:    "/",
				Delimiter: "/",
				Optional:  false,
				Repeat:    false,
				Pattern:   "[^\\/]+?",
			},
			Token{
				Name:      "bar",
				Prefix:    "/",
				Delimiter: "/",
				Optional:  false,
				Repeat:    false,
				Pattern:   "[^\\/]+?",
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
		"/:foo(test\\)/bar",
		nil,
		a{
			Token{
				Name:      "foo",
				Prefix:    "/",
				Delimiter: "/",
				Optional:  false,
				Repeat:    false,
				Pattern:   "[^\\/]+?",
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
				Name:      "remote",
				Prefix:    "/",
				Delimiter: "/",
				Optional:  false,
				Repeat:    false,
				Pattern:   "[\\w-.]+",
			},
			Token{
				Name:      "user",
				Prefix:    "/",
				Delimiter: "/",
				Optional:  false,
				Repeat:    false,
				Pattern:   "[\\w-]+",
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
				Name:      "foo",
				Prefix:    "/",
				Delimiter: "/",
				Optional:  false,
				Repeat:    false,
				Pattern:   "[^\\/]+?",
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
				Name:      "foo",
				Prefix:    "/",
				Delimiter: "/",
				Optional:  false,
				Repeat:    true,
				Pattern:   "[^\\/]+?",
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
			a{m{"foo": "foo/bar"}, "/foo%2Fbarbaz"},
			a{m{"foo": a{"foo", "bar"}}, "/foo/barbaz"},
		},
	},
	{
		"\\/:pre?baz",
		nil,
		a{
			"/",
			Token{
				Name:      "pre",
				Prefix:    "",
				Delimiter: "/",
				Optional:  true,
				Repeat:    false,
				Pattern:   "[^\\/]+?",
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
				Name:      "foo",
				Prefix:    "/",
				Delimiter: "/",
				Optional:  false,
				Repeat:    false,
				Pattern:   "[^\\/]+?",
			},
			"(",
			Token{
				Name:      "bar",
				Prefix:    "",
				Delimiter: "/",
				Optional:  true,
				Repeat:    false,
				Pattern:   "[^\\/]+?",
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
				Name:      "postType",
				Prefix:    "/",
				Delimiter: "/",
				Optional:  false,
				Repeat:    false,
				Pattern:   "video|audio|text",
			},
			Token{
				Name:      0,
				Prefix:    "",
				Delimiter: "/",
				Optional:  true,
				Repeat:    false,
				Pattern:   "\\+.+",
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
				Name:      "foo",
				Prefix:    "/",
				Delimiter: "/",
				Optional:  true,
				Repeat:    false,
				Pattern:   "[^\\/]+?",
			},
			Token{
				Name:      "bar",
				Prefix:    "/",
				Delimiter: "/",
				Optional:  true,
				Repeat:    false,
				Pattern:   "[^\\/]+?",
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
				Name:      "required",
				Prefix:    "/",
				Delimiter: "/",
				Optional:  false,
				Repeat:    false,
				Pattern:   "[^\\/]+?",
			},
			Token{
				Name:      "optional",
				Prefix:    "/",
				Delimiter: "/",
				Optional:  true,
				Repeat:    false,
				Pattern:   "[^\\/]+?",
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
			a{m{"required": "foo", "optional": "bar"}, "/foo/bar-ext"},
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
				Name:      "foo",
				Prefix:    "/",
				Delimiter: "/",
				Optional:  false,
				Repeat:    false,
				Pattern:   "[^\\/]+?",
			},
		},
		a{
			a{"/café", a{"/café", "café"}},
		},
		a{
			a{m{"foo": "café"}, "/caf%C3%A9"},
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
				Name:      "domain",
				Prefix:    "",
				Delimiter: ".",
				Optional:  false,
				Repeat:    false,
				Pattern:   "[^\\.]+?",
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
				Name:      "domain",
				Prefix:    ".",
				Delimiter: ".",
				Optional:  false,
				Repeat:    false,
				Pattern:   "[^\\.]+?",
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
		&Options{
			Delimiter: ".",
		},
		a{
			"example",
			Token{
				Name:      "ext",
				Prefix:    ".",
				Delimiter: ".",
				Optional:  false,
				Repeat:    false,
				Pattern:   "[^\\.]+?",
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
	 * Custom delimiters.
	 */
	{
		"$:foo$:bar?",
		&Options{
			Delimiter: "$",
		},
		a{
			Token{
				Delimiter: "$",
				Name:      "foo",
				Optional:  false,
				Pattern:   "[^\\$]+?",
				Prefix:    "$",
				Repeat:    false,
			},
			Token{
				Delimiter: "$",
				Name:      "bar",
				Optional:  true,
				Pattern:   "[^\\$]+?",
				Prefix:    "$",
				Repeat:    false,
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
		":test+",
		&Options{
			Delimiter: " ",
		},
		a{
			Token{
				Name:      "test",
				Prefix:    "",
				Delimiter: " ",
				Optional:  false,
				Repeat:    true,
				Pattern:   "[^ ]+?",
			},
		},
		a{
			a{"hello", a{"hello", "hello"}},
			a{" hello", nil},
			a{"", nil},
			a{"hello world", a{"hello world", "hello world"}},
		},
		a{
			a{m{}, nil},
			a{m{"test": ""}, nil},
			a{m{"test": a{"hello"}}, "hello"},
			a{m{"test": a{"hello", "world"}}, "hello world"},
		},
	},
	{
		"name/:attr1?-:attr2?-:attr3?",
		&Options{},
		a{
			"name",
			Token{
				Delimiter: "/",
				Name:      "attr1",
				Optional:  true,
				Pattern:   "[^\\/]+?",
				Prefix:    "/",
				Repeat:    false,
			},
			Token{
				Delimiter: "-",
				Name:      "attr2",
				Optional:  true,
				Pattern:   "[^-\\/]+?",
				Prefix:    "-",
				Repeat:    false,
			},
			Token{
				Delimiter: "-",
				Name:      "attr3",
				Optional:  true,
				Pattern:   "[^-\\/]+?",
				Prefix:    "-",
				Repeat:    false,
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
	{
		"name/:attr1?-:attr2?",
		&Options{
			Whitelist: []string{"/"},
		},
		a{
			"name",
			Token{
				Delimiter: "/",
				Name:      "attr1",
				Optional:  true,
				Pattern:   "[^\\/]+?",
				Prefix:    "/",
				Repeat:    false,
			},
			"-",
			Token{
				Delimiter: "/",
				Name:      "attr2",
				Optional:  true,
				Pattern:   "[^\\/]+?",
				Prefix:    "",
				Repeat:    false,
			},
		},
		a{
			a{"name/1", nil},
			a{"name/1-", a{"name/1-", "1", ""}},
			a{"name/1-2", a{"name/1-2", "1", "2"}},
			a{"name/1-2-3", a{"name/1-2-3", "1", "2-3"}},
			a{"name/foo-bar/route", nil},
			a{"name/test/route", nil},
		},
		a{
			a{m{}, "name-"},
			a{m{"attr1": "test"}, "name/test-"},
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
				Name:      "test",
				Prefix:    "/",
				Delimiter: "/",
				Optional:  false,
				Repeat:    false,
				Pattern:   "abc",
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
				Name:      "test",
				Prefix:    "/",
				Delimiter: "/",
				Optional:  false,
				Repeat:    false,
				Pattern:   "abc",
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
	 * Nested parenthesis.
	 */
	{
		"/:foo(\\d+(?:\\.\\d+)?)",
		&Options{},
		a{
			Token{
				Name:      "foo",
				Prefix:    "/",
				Delimiter: "/",
				Optional:  false,
				Repeat:    false,
				Pattern:   "\\d+(?:\\.\\d+)?",
			},
		},
		a{
			a{"/123", a{"/123", "123"}},
			a{"/123.123", a{"/123.123", "123.123"}},
		},
		a{
			a{m{"foo": 123}, "/123"},
			a{m{"foo": 123.123}, "/123.123"},
			a{m{"foo": "123"}, "/123"},
		},
	},
}

// Dynamically generate the entire test suite.
func TestPathToRegexp(t *testing.T) {
	testPath := "/user/:id"

	testParam := Token{
		Name:      "id",
		Prefix:    "/",
		Delimiter: "/",
		Optional:  false,
		Repeat:    false,
		Pattern:   "[^\\/]+?",
	}

	t.Run("arguments", func(t *testing.T) {
		t.Run("should work without different call combinations", func(t *testing.T) {
			_, err := PathToRegexp("/test", nil, nil)
			if err != nil {
				t.Error(err.Error())
			}
			_, err = PathToRegexp("/test", &[]Token{}, nil)
			if err != nil {
				t.Error(err.Error())
			}
			_, err = PathToRegexp("/test", nil, &Options{})
			if err != nil {
				t.Error(err.Error())
			}

			_, err = PathToRegexp(regexp2.MustCompile("^\\/test", regexp2.None), nil, nil)
			if err != nil {
				t.Error(err.Error())
			}
			_, err = PathToRegexp(regexp2.MustCompile("^\\/test", regexp2.None), &[]Token{}, nil)
			if err != nil {
				t.Error(err.Error())
			}
			_, err = PathToRegexp(regexp2.MustCompile("^\\/test", regexp2.None), nil, &Options{})
			if err != nil {
				t.Error(err.Error())
			}

			_, err = PathToRegexp([]string{"/a", "/b"}, nil, nil)
			if err != nil {
				t.Error(err.Error())
			}
			_, err = PathToRegexp([]string{"/a", "/b"}, &[]Token{}, nil)
			if err != nil {
				t.Error(err.Error())
			}
			_, err = PathToRegexp([]string{"/a", "/b"}, nil, &Options{})
			if err != nil {
				t.Error(err.Error())
			}
		})

		t.Run("should accept an array of tokens as the second argument", func(t *testing.T) {
			tokens := &[]Token{}
			r, err := PathToRegexp(testPath, tokens, &Options{End: &falseValue})
			if err != nil {
				t.Error(err.Error())
				return
			}
			var want interface{}
			want = &[]Token{testParam}

			if !reflect.DeepEqual(tokens, want) {
				t.Errorf("got %v want %v", tokens, want)
			}

			want = []string{"/user/123", "123"}
			if !reflect.DeepEqual(exec(r, "/user/123/show"), want) {
				t.Errorf("got %v want %v", tokens, want)
			}
		})

		t.Run("should throw on non-capturing pattern group", func(t *testing.T) {
			defer func() {
				want := `Path pattern must be a capturing group`
				if err := recover(); !reflect.DeepEqual(err, want) {
					t.Errorf("got panic(%v) want panic(%v)", err, want)
				}
			}()
			_, err := PathToRegexp("/:foo(?:\\d+(\\.\\d+)?)", nil, nil)
			if err != nil {
				t.Error(err)
			}
		})

		t.Run("should throw on nested capturing regexp groups", func(t *testing.T) {
			defer func() {
				want := "Capturing groups are not allowed in pattern, use a non-capturing group: (\\d+(?:\\.\\d+)?)"
				if err := recover(); !reflect.DeepEqual(err, want) {
					t.Errorf("got panic(%v) want panic(%v)", err, want)
				}
			}()
			_, err := PathToRegexp("/:foo(\\d+(\\.\\d+)?)", nil, nil)
			if err != nil {
				t.Error(err)
			}
		})
	})

	t.Run("tokens", func(t *testing.T) {
		tokens := Parse(testPath, nil)
		t.Run("should expose method to compile tokens to regexp", func(t *testing.T) {
			r, err := tokensToRegExp(tokens, nil, nil)
			if err != nil {
				t.Error(err.Error())
				return
			}
			want := []string{"/user/123", "123"}
			result := exec(r, "/user/123")
			if !reflect.DeepEqual(result, want) {
				t.Errorf("got %v want %v", result, want)
			}
		})
		t.Run("should expose method to compile tokens to a path function", func(t *testing.T) {
			fn, err := tokensToFunction(tokens, nil)
			if err != nil {
				t.Error(err.Error())
				return
			}
			want := "/user/123"
			result := fn(map[interface{}]interface{}{"id": 123})
			if !reflect.DeepEqual(result, want) {
				t.Errorf("got %v want %v", result, want)
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
					t.Error(err.Error())
					return
				}
				// Parsing and compiling is only supported with string input.
				if path, ok := path.(string); ok {
					t.Run("should parse", func(t *testing.T) {
						result := a(Parse(path, o))
						if !reflect.DeepEqual(result, rawTokens) {
							t.Errorf("got %v want %v", result, rawTokens)
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
								t.Error(err.Error())
								return
							}
							if result != nil {
								t.Run("should compile using "+inspect(params), func(t *testing.T) {
									r := toPath(params)
									if !reflect.DeepEqual(r, result) {
										t.Errorf("got %v want %v", result, path)
									}
								})
							} else {
								t.Run("should not compile using "+inspect(params), func(t *testing.T) {
									defer func() {
										if err := recover(); err == nil {
											t.Errorf("got %v want panic", err)
										}
									}()
									toPath(params)
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
							t.Errorf("got %v want %v", *tokens, tTokens)
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
								t.Errorf("got %v want %v", result, matches)
							}
						})

						var params *MatchResult
						if len(io) >= 3 && io[2] != nil {
							params = io[2].(*MatchResult)
						}

						if path, ok := path.(string); ok && params != nil {
							match := MustMatch(path, nil)
							t.Run(message+" params", func(t *testing.T) {
								m := match(pathname.(string))
								if !params.equals(m) {
									t.Errorf("got %v want %v", m, params)
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
			defer func() {
				want := `Expected "b" to be a string`
				if err := recover(); !reflect.DeepEqual(err, want) {
					t.Errorf("got panic(%v) want panic(%v)", err, want)
				}
			}()
			toPath, err := Compile("/a/:b/c", nil)
			if err != nil {
				t.Error(err.Error())
				return
			}
			toPath(nil)
		})

		t.Run("should throw when it does not match the pattern", func(t *testing.T) {
			defer func() {
				want := `Expected "foo" to match "\d+", but got "abc"`
				if err := recover(); !reflect.DeepEqual(err, want) {
					t.Errorf("got panic(%v) want panic(%v)", err, want)
				}
			}()
			toPath, err := Compile("/:foo(\\d+)", nil)
			if err != nil {
				t.Error(err.Error())
				return
			}
			toPath(map[interface{}]interface{}{"foo": "abc"})
		})

		t.Run("should throw when expecting a repeated value", func(t *testing.T) {
			defer func() {
				want := `Expected "foo" to not be empty`
				if err := recover(); !reflect.DeepEqual(err, want) {
					t.Errorf("got panic(%v) want panic(%v)", err, want)
				}
			}()
			toPath, err := Compile("/:foo+", nil)
			if err != nil {
				t.Error(err.Error())
				return
			}
			toPath(map[interface{}]interface{}{"foo": []interface{}{}})
		})

		t.Run("should throw when not expecting a repeated value", func(t *testing.T) {
			defer func() {
				want := `Expected "foo" to not repeat, but got array`
				if err := recover(); !reflect.DeepEqual(err, want) {
					t.Errorf("got panic(%v) want panic(%v)", err, want)
				}
			}()
			toPath, err := Compile("/:foo", nil)
			if err != nil {
				t.Error(err.Error())
				return
			}
			toPath(map[interface{}]interface{}{"foo": []interface{}{}})
		})

		t.Run("should throw when repeated value does not match", func(t *testing.T) {
			defer func() {
				want := `Expected all "foo" to match "\d+"`
				if err := recover(); !reflect.DeepEqual(err, want) {
					t.Errorf("got panic(%v) want panic(%v)", err, want)
				}
			}()
			toPath, err := Compile("/:foo(\\d+)+", nil)
			if err != nil {
				t.Error(err.Error())
				return
			}
			toPath(map[interface{}]interface{}{"foo": []interface{}{1, 2, 3, "a"}})
		})
	})

	t.Run("path should be string, or strings, or a regular expression", func(t *testing.T) {
		_, err := PathToRegexp(123, nil, nil)
		if err == nil {
			t.Errorf("should got non nil error")
		}
	})
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
	if o2.EndsWith != nil {
		endsWith = o2.EndsWith
	}

	whitelist := o1.Whitelist
	if o2.Whitelist != nil {
		whitelist = o2.Whitelist
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
		Whitelist: whitelist,
		Encode:    encode,
		Decode:    decode,
	}
}
