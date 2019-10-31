// Copyright 2019 Guoyao Wu. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package path_to_regexp

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
			a{"/", a{"/"}},
			a{"/route", nil},
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
			a{"/test", a{"/test"}},
			a{"/route", nil},
			a{"/test/route", nil},
			a{"/test/", a{"/test/"}},
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
			sensitive: true,
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
			sensitive: true,
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
			strict: true,
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
			strict: true,
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
			end: &falseValue,
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
			end: &falseValue,
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
			end: &falseValue,
		},
		a{
			Key{
				name:      "test",
				prefix:    "/",
				delimiter: "/",
				optional:  false,
				repeat:    false,
				pattern:   "[^\\/]+?",
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
					encode: func(uri string, token interface{}) string {
						return uri
					},
				},
			},
			a{
				m{"test": "a+b"},
				"/test",
				&Options{
					encode: func(uri string, token interface{}) string {
						if key, ok := token.(Key); ok {
							return fmt.Sprintf("%v", key.name)
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
			end: &falseValue,
		},
		a{
			Key{
				name:      "test",
				prefix:    "/",
				delimiter: "/",
				optional:  false,
				repeat:    false,
				pattern:   "[^\\/]+?",
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
			end: &falseValue,
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
			start: &falseValue,
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
			start: &falseValue,
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
			start: &falseValue,
		},
		a{
			Key{
				name:      "test",
				prefix:    "/",
				delimiter: "/",
				optional:  false,
				repeat:    false,
				pattern:   "[^\\/]+?",
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
					encode: func(uri string, token interface{}) string {
						return uri
					},
				},
			},
			a{
				m{"test": "a+b"},
				"/test",
				&Options{
					encode: func(uri string, token interface{}) string {
						if key, ok := token.(Key); ok {
							return fmt.Sprintf("%v", key.name)
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
			start: &falseValue,
		},
		a{
			Key{
				name:      "test",
				prefix:    "/",
				delimiter: "/",
				optional:  false,
				repeat:    false,
				pattern:   "[^\\/]+?",
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
			start: &falseValue,
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
			end:    &falseValue,
			strict: true,
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
			end:    &falseValue,
			strict: true,
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
			end:    &falseValue,
			strict: true,
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
			end:    &falseValue,
			strict: true,
		},
		a{
			Key{
				name:      "test",
				prefix:    "/",
				delimiter: "/",
				optional:  false,
				repeat:    false,
				pattern:   "[^\\/]+?",
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
			end:    &falseValue,
			strict: true,
		},
		a{
			Key{
				name:      "test",
				prefix:    "/",
				delimiter: "/",
				optional:  false,
				repeat:    false,
				pattern:   "[^\\/]+?",
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
			start: &falseValue,
			end:   &falseValue,
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
			start: &falseValue,
			end:   &falseValue,
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
			start: &falseValue,
			end:   &falseValue,
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
			start: &falseValue,
			end:   &falseValue,
		},
		a{
			Key{
				name:      "test",
				prefix:    "/",
				delimiter: "/",
				optional:  false,
				repeat:    false,
				pattern:   "[^\\/]+?",
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
			start: &falseValue,
			end:   &falseValue,
		},
		a{
			Key{
				name:      "test",
				prefix:    "/",
				delimiter: "/",
				optional:  false,
				repeat:    false,
				pattern:   "[^\\/]+?",
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
			end: &falseValue,
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
			Key{
				name:      "test",
				prefix:    "/",
				delimiter: "/",
				optional:  false,
				repeat:    false,
				pattern:   "[^\\/]+?",
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
			strict: true,
		},
		a{
			Key{
				name:      "test",
				prefix:    "/",
				delimiter: "/",
				optional:  false,
				repeat:    false,
				pattern:   "[^\\/]+?",
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
			strict: true,
		},
		a{
			Key{
				name:      "test",
				prefix:    "/",
				delimiter: "/",
				optional:  false,
				repeat:    false,
				pattern:   "[^\\/]+?",
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
			end: &falseValue,
		},
		a{
			Key{
				name:      "test",
				prefix:    "/",
				delimiter: "/",
				optional:  false,
				repeat:    false,
				pattern:   "[^\\/]+?",
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
			Key{
				name:      "test",
				prefix:    "/",
				delimiter: "/",
				optional:  true,
				repeat:    false,
				pattern:   "[^\\/]+?",
			},
		},
		a{
			a{"/route", a{"/route", "route"}},
			a{"/route/nested", nil},
			a{"/", a{"/", ""}},
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
			strict: true,
		},
		a{
			Key{
				name:      "test",
				prefix:    "/",
				delimiter: "/",
				optional:  true,
				repeat:    false,
				pattern:   "[^\\/]+?",
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
			strict: true,
		},
		a{
			Key{
				name:      "test",
				prefix:    "/",
				delimiter: "/",
				optional:  true,
				repeat:    false,
				pattern:   "[^\\/]+?",
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
			Key{
				name:      "test",
				prefix:    "/",
				delimiter: "/",
				optional:  true,
				repeat:    false,
				pattern:   "[^\\/]+?",
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
			Key{
				name:      "test",
				prefix:    "/",
				delimiter: "/",
				optional:  true,
				repeat:    false,
				pattern:   "[^\\/]+?",
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
			Key{
				name:      "test",
				prefix:    "/",
				delimiter: "/",
				optional:  true,
				repeat:    true,
				pattern:   "[^\\/]+?",
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
			Key{
				name:      "test",
				prefix:    "/",
				delimiter: "/",
				optional:  false,
				repeat:    true,
				pattern:   "[^\\/]+?",
			},
		},
		a{
			a{"/", nil},
			a{"/route", a{"/route", "route"}},
			a{"/some/basic/route", a{"/some/basic/route", "some/basic/route"}},
			a{"//", nil},
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
			Key{
				name:      "test",
				prefix:    "/",
				delimiter: "/",
				optional:  false,
				repeat:    true,
				pattern:   "\\d+",
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
			Key{
				name:      "ext",
				prefix:    ".",
				delimiter: ".",
				optional:  false,
				repeat:    true,
				pattern:   "json|xml",
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
			Key{
				name:      "ext",
				prefix:    ".",
				delimiter: ".",
				optional:  false,
				repeat:    false,
				pattern:   "[^\\.\\/]+?",
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
			Key{
				name:      "test",
				prefix:    "/",
				delimiter: "/",
				optional:  true,
				repeat:    true,
				pattern:   "[^\\/]+?",
			},
		},
		a{
			a{"/", a{"/", ""}},
			a{"//", nil},
			a{"/route", a{"/route", "route"}},
			a{"/some/basic/route", a{"/some/basic/route", "some/basic/route"}},
		},
		a{
			a{m{}, ""},
			a{m{"test": "foobar"}, "/foobar"},
			a{m{"test": a{"foo", "bar"}}, "/foo/bar"},
		},
	},
	{
		"/route.:ext([a-z]+)*",
		nil,
		a{
			"/route",
			Key{
				name:      "ext",
				prefix:    ".",
				delimiter: ".",
				optional:  true,
				repeat:    true,
				pattern:   "[a-z]+",
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
			Key{
				name:      "test",
				prefix:    "/",
				delimiter: "/",
				optional:  false,
				repeat:    false,
				pattern:   "\\d+",
			},
		},
		a{
			a{"/123", a{"/123", "123"}},
			a{"/abc", nil},
			a{"/123/abc", nil},
		},
		a{
			a{m{"test": "abc"}, nil},
			a{m{"test": "abc"}, "/abc", &Options{validate: &falseValue}},
			a{m{"test": "123"}, "/123"},
		},
	},
	{
		"/:test(\\d+)",
		&Options{
			end: &falseValue,
		},
		a{
			Key{
				name:      "test",
				prefix:    "/",
				delimiter: "/",
				optional:  false,
				repeat:    false,
				pattern:   "\\d+",
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
			Key{
				name:      "test",
				prefix:    "/",
				delimiter: "/",
				optional:  false,
				repeat:    false,
				pattern:   ".*",
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
			Key{
				name:      "route",
				prefix:    "/",
				delimiter: "/",
				optional:  false,
				repeat:    false,
				pattern:   "[a-z]+",
			},
		},
		a{
			a{"/abcde", a{"/abcde", "abcde"}},
			a{"/12345", nil},
		},
		a{
			a{m{"route": ""}, nil},
			a{m{"route": ""}, "/", &Options{validate: &falseValue}},
			a{m{"route": "123"}, nil},
			a{m{"route": "123"}, "/123", &Options{validate: &falseValue}},
			a{m{"route": "abc"}, "/abc"},
		},
	},
	{
		"/:route(this|that)",
		nil,
		a{
			Key{
				name:      "route",
				prefix:    "/",
				delimiter: "/",
				optional:  false,
				repeat:    false,
				pattern:   "this|that",
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
			a{m{"route": "foo"}, "/foo", &Options{validate: &falseValue}},
			a{m{"route": "that"}, "/that"},
		},
	},
	{
		"/:path(abc|xyz)*",
		nil,
		a{
			Key{
				name:      "path",
				prefix:    "/",
				delimiter: "/",
				optional:  true,
				repeat:    true,
				pattern:   "abc|xyz",
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
			a{m{"path": "abc123"}, "/abc123", &Options{validate: &falseValue}},
			a{m{"path": "abcxyz"}, nil},
			a{m{"path": "abcxyz"}, "/abcxyz", &Options{validate: &falseValue}},
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
			Key{
				name:      "test",
				prefix:    "",
				delimiter: "/",
				optional:  false,
				repeat:    false,
				pattern:   "[^\\/]+?",
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
			strict: true,
		},
		a{
			Key{
				name:      "test",
				prefix:    "",
				delimiter: "/",
				optional:  false,
				repeat:    false,
				pattern:   "[^\\/]+?",
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
			end: &falseValue,
		},
		a{
			Key{
				name:      "test",
				prefix:    "",
				delimiter: "/",
				optional:  false,
				repeat:    false,
				pattern:   "[^\\/]+?",
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
			Key{
				name:      "test",
				prefix:    "",
				delimiter: "/",
				optional:  true,
				repeat:    false,
				pattern:   "[^\\/]+?",
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
			Key{
				name:      "test",
				prefix:    "",
				delimiter: "/",
				optional:  false,
				repeat:    true,
				pattern:   "[^\\/]+?",
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
			Key{
				name:      "test",
				prefix:    "/",
				delimiter: "/",
				optional:  false,
				repeat:    false,
				pattern:   "[^\\/]+?",
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
			Key{
				name:      "format",
				prefix:    ".",
				delimiter: ".",
				optional:  false,
				repeat:    false,
				pattern:   "[^\\.\\/]+?",
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
			Key{
				name:      "format",
				prefix:    ".",
				delimiter: ".",
				optional:  false,
				repeat:    false,
				pattern:   "[^\\.\\/]+?",
			},
			Key{
				name:      "format",
				prefix:    ".",
				delimiter: ".",
				optional:  false,
				repeat:    false,
				pattern:   "[^\\.\\/]+?",
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
			Key{
				name:      "format",
				prefix:    ".",
				delimiter: ".",
				optional:  false,
				repeat:    true,
				pattern:   "[^\\.\\/]+?",
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
			end: &falseValue,
		},
		a{
			"/test",
			Key{
				name:      "format",
				prefix:    ".",
				delimiter: ".",
				optional:  false,
				repeat:    false,
				pattern:   "[^\\.\\/]+?",
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
			Key{
				name:      "format",
				prefix:    ".",
				delimiter: ".",
				optional:  false,
				repeat:    false,
				pattern:   "[^\\.\\/]+?",
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
			Key{
				name:      "test",
				prefix:    "/",
				delimiter: "/",
				optional:  false,
				repeat:    false,
				pattern:   "[^\\/]+?",
			},
			Key{
				name:      "format",
				prefix:    ".",
				delimiter: ".",
				optional:  false,
				repeat:    false,
				pattern:   "[^\\.\\/]+?",
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
			Key{
				name:      "test",
				prefix:    "/",
				delimiter: "/",
				optional:  false,
				repeat:    false,
				pattern:   "[^\\/]+?",
			},
			Key{
				name:      "format",
				prefix:    ".",
				delimiter: ".",
				optional:  true,
				repeat:    false,
				pattern:   "[^\\.\\/]+?",
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
			end: &falseValue,
		},
		a{
			Key{
				name:      "test",
				prefix:    "/",
				delimiter: "/",
				optional:  false,
				repeat:    false,
				pattern:   "[^\\/]+?",
			},
			Key{
				name:      "format",
				prefix:    ".",
				delimiter: ".",
				optional:  true,
				repeat:    false,
				pattern:   "[^\\.\\/]+?",
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
			end: &falseValue,
		},
		a{
			"/test",
			Key{
				name:      "format",
				prefix:    ".",
				delimiter: ".",
				optional:  false,
				repeat:    false,
				pattern:   ".*",
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
			Key{
				name:      0,
				prefix:    "/",
				delimiter: "/",
				optional:  false,
				repeat:    false,
				pattern:   "\\d+",
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
			end: &falseValue,
		},
		a{
			Key{
				name:      0,
				prefix:    "/",
				delimiter: "/",
				optional:  false,
				repeat:    false,
				pattern:   "\\d+",
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
			Key{
				name:      0,
				prefix:    "/",
				delimiter: "/",
				optional:  true,
				repeat:    false,
				pattern:   "\\d+",
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
			Key{
				name:      0,
				prefix:    "/",
				delimiter: "/",
				optional:  false,
				repeat:    false,
				pattern:   ".*",
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
			Key{
				name:      0,
				prefix:    "",
				delimiter: "/",
				optional:  false,
				repeat:    false,
				pattern:   "\\d+\\\\",
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
			Key{
				name:      0,
				prefix:    "/",
				delimiter: "/",
				optional:  true,
				repeat:    false,
				pattern:   "login",
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
			Key{
				name:      0,
				prefix:    "",
				delimiter: "",
				optional:  false,
				repeat:    false,
				pattern:   "",
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
			Key{
				name:      0,
				prefix:    "",
				delimiter: "",
				optional:  false,
				repeat:    false,
				pattern:   "",
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
			Key{
				name:      0,
				prefix:    "",
				delimiter: "",
				optional:  false,
				repeat:    false,
				pattern:   "",
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
			Key{
				name:      "test",
				prefix:    "/",
				delimiter: "/",
				optional:  false,
				repeat:    false,
				pattern:   "\\d+",
			},
			Key{
				name:      0,
				prefix:    "",
				delimiter: "",
				optional:  false,
				repeat:    false,
				pattern:   "",
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
			Key{
				name:      "test",
				prefix:    "/",
				delimiter: "/",
				optional:  false,
				repeat:    false,
				pattern:   "[^\\/]+?",
			},
			Key{
				name:      "test",
				prefix:    "/",
				delimiter: "/",
				optional:  false,
				repeat:    false,
				pattern:   "[^\\/]+?",
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
			Key{
				name:      0,
				prefix:    "",
				delimiter: "",
				optional:  false,
				repeat:    false,
				pattern:   "",
			},
			Key{
				name:      0,
				prefix:    "",
				delimiter: "",
				optional:  false,
				repeat:    false,
				pattern:   "",
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
			Key{
				name:      "uid",
				prefix:    "",
				delimiter: "/",
				optional:  true,
				repeat:    false,
				pattern:   "u\\d+",
			},
			Key{
				name:      "cid",
				prefix:    "",
				delimiter: "/",
				optional:  true,
				repeat:    false,
				pattern:   "c\\d+",
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
			Key{
				name:      0,
				prefix:    "",
				delimiter: "/",
				optional:  true,
				repeat:    false,
				pattern:   "apple-",
			},
			"icon",
			Key{
				name:      "res",
				prefix:    "-",
				delimiter: "-",
				optional:  false,
				repeat:    false,
				pattern:   "\\d+",
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
			Key{
				name:      "foo",
				prefix:    "/",
				delimiter: "/",
				optional:  false,
				repeat:    false,
				pattern:   "[^\\/]+?",
			},
			Key{
				name:      "bar",
				prefix:    "/",
				delimiter: "/",
				optional:  false,
				repeat:    false,
				pattern:   "[^\\/]+?",
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
			Key{
				name:      "foo",
				prefix:    "/",
				delimiter: "/",
				optional:  false,
				repeat:    false,
				pattern:   "[^\\/]+?",
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
			Key{
				name:      "remote",
				prefix:    "/",
				delimiter: "/",
				optional:  false,
				repeat:    false,
				pattern:   "[\\w-.]+",
			},
			Key{
				name:      "user",
				prefix:    "/",
				delimiter: "/",
				optional:  false,
				repeat:    false,
				pattern:   "[\\w-]+",
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
			Key{
				name:      "foo",
				prefix:    "/",
				delimiter: "/",
				optional:  false,
				repeat:    false,
				pattern:   "[^\\/]+?",
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
			Key{
				name:      "foo",
				prefix:    "/",
				delimiter: "/",
				optional:  false,
				repeat:    true,
				pattern:   "[^\\/]+?",
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
			Key{
				name:      "pre",
				prefix:    "",
				delimiter: "/",
				optional:  true,
				repeat:    false,
				pattern:   "[^\\/]+?",
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
			Key{
				name:      "foo",
				prefix:    "/",
				delimiter: "/",
				optional:  false,
				repeat:    false,
				pattern:   "[^\\/]+?",
			},
			"(",
			Key{
				name:      "bar",
				prefix:    "",
				delimiter: "/",
				optional:  true,
				repeat:    false,
				pattern:   "[^\\/]+?",
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
			Key{
				name:      "postType",
				prefix:    "/",
				delimiter: "/",
				optional:  false,
				repeat:    false,
				pattern:   "video|audio|text",
			},
			Key{
				name:      0,
				prefix:    "",
				delimiter: "/",
				optional:  true,
				repeat:    false,
				pattern:   "\\+.+",
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
			Key{
				name:      "foo",
				prefix:    "/",
				delimiter: "/",
				optional:  true,
				repeat:    false,
				pattern:   "[^\\/]+?",
			},
			Key{
				name:      "bar",
				prefix:    "/",
				delimiter: "/",
				optional:  true,
				repeat:    false,
				pattern:   "[^\\/]+?",
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
			Key{
				name:      "required",
				prefix:    "/",
				delimiter: "/",
				optional:  false,
				repeat:    false,
				pattern:   "[^\\/]+?",
			},
			Key{
				name:      "optional",
				prefix:    "/",
				delimiter: "/",
				optional:  true,
				repeat:    false,
				pattern:   "[^\\/]+?",
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
			Key{
				name:      "foo",
				prefix:    "/",
				delimiter: "/",
				optional:  false,
				repeat:    false,
				pattern:   "[^\\/]+?",
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
			delimiter: ".",
		},
		a{
			Key{
				name:      "domain",
				prefix:    "",
				delimiter: ".",
				optional:  false,
				repeat:    false,
				pattern:   "[^\\.]+?",
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
			delimiter: ".",
		},
		a{
			"mail",
			Key{
				name:      "domain",
				prefix:    ".",
				delimiter: ".",
				optional:  false,
				repeat:    false,
				pattern:   "[^\\.]+?",
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
			delimiter: ".",
		},
		a{
			"example",
			Key{
				name:      "ext",
				prefix:    ".",
				delimiter: ".",
				optional:  false,
				repeat:    false,
				pattern:   "[^\\.]+?",
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
			delimiter: " ",
			end:       &falseValue,
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
			endsWith: "?",
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
			endsWith: "?",
			strict:   true,
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
			delimiter: "$",
		},
		a{
			Key{
				delimiter: "$",
				name:      "foo",
				optional:  false,
				pattern:   "[^\\$]+?",
				prefix:    "$",
				repeat:    false,
			},
			Key{
				delimiter: "$",
				name:      "bar",
				optional:  true,
				pattern:   "[^\\$]+?",
				prefix:    "$",
				repeat:    false,
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
			delimiter: " ",
		},
		a{
			Key{
				name:      "test",
				prefix:    "",
				delimiter: " ",
				optional:  false,
				repeat:    true,
				pattern:   "[^ ]+?",
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
			Key{
				delimiter: "/",
				name:      "attr1",
				optional:  true,
				pattern:   "[^\\/]+?",
				prefix:    "/",
				repeat:    false,
			},
			Key{
				delimiter: "-",
				name:      "attr2",
				optional:  true,
				pattern:   "[^-\\/]+?",
				prefix:    "-",
				repeat:    false,
			},
			Key{
				delimiter: "-",
				name:      "attr3",
				optional:  true,
				pattern:   "[^-\\/]+?",
				prefix:    "-",
				repeat:    false,
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
			whitelist: []string{"/"},
		},
		a{
			"name",
			Key{
				delimiter: "/",
				name:      "attr1",
				optional:  true,
				pattern:   "[^\\/]+?",
				prefix:    "/",
				repeat:    false,
			},
			"-",
			Key{
				delimiter: "/",
				name:      "attr2",
				optional:  true,
				pattern:   "[^\\/]+?",
				prefix:    "",
				repeat:    false,
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
			sensitive: true,
		},
		a{
			Key{
				name:      "test",
				prefix:    "/",
				delimiter: "/",
				optional:  false,
				repeat:    false,
				pattern:   "abc",
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
			Key{
				name:      "test",
				prefix:    "/",
				delimiter: "/",
				optional:  false,
				repeat:    false,
				pattern:   "abc",
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
}

func TestPathToRegexp(t *testing.T) {
	testPath := "/user/:id"

	testParam := Key{
		name:      "id",
		prefix:    "/",
		delimiter: "/",
		optional:  false,
		repeat:    false,
		pattern:   "[^\\/]+?",
	}

	t.Run("arguments", func(t *testing.T) {
		t.Run("should work without different call combinations", func(t *testing.T) {
			_, err := PathToRegexp("/test", nil, nil)
			if err != nil {
				t.Error(err.Error())
			}
			_, err = PathToRegexp("/test", &[]Key{}, nil)
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
			_, err = PathToRegexp(regexp2.MustCompile("^\\/test", regexp2.None), &[]Key{}, nil)
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
			_, err = PathToRegexp([]string{"/a", "/b"}, &[]Key{}, nil)
			if err != nil {
				t.Error(err.Error())
			}
			_, err = PathToRegexp([]string{"/a", "/b"}, nil, &Options{})
			if err != nil {
				t.Error(err.Error())
			}
		})

		t.Run("should accept an array of keys as the second argument", func(t *testing.T) {
			keys := &[]Key{}
			r, err := PathToRegexp(testPath, keys, &Options{end: &falseValue})
			if err != nil {
				t.Error(err.Error())
				return
			}
			var want interface{}
			want = &[]Key{testParam}

			if !reflect.DeepEqual(keys, want) {
				t.Errorf("got %v want %v", keys, want)
			}

			want = []string{"/user/123", "123"}
			if !reflect.DeepEqual(exec(r, "/user/123/show"), want) {
				t.Errorf("got %v want %v", keys, want)
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
			result := fn(map[interface{}]interface{}{"id": 123}, nil)
			if !reflect.DeepEqual(result, want) {
				t.Errorf("got %v want %v", result, want)
			}
		})
	})

	t.Run("rules", func(t *testing.T) {
		for _, test := range tests {
			path, opts, tokens := test[0], test[1], test[2].(a)
			matchCases, compileCases := test[3].(a), test[4].(a)
			t.Run(inspect(path), func(t *testing.T) {
				keys := &[]Key{}
				var o *Options
				if opts != nil {
					o = opts.(*Options)
				}
				r, err := PathToRegexp(path, keys, o)
				if err != nil {
					t.Error(err.Error())
					return
				}
				if path, ok := path.(string); ok {
					t.Run("should parse", func(t *testing.T) {
						result := a(Parse(path, o))
						if !reflect.DeepEqual(result, tokens) {
							t.Errorf("got %v want %v", result, tokens)
						}
					})
					t.Run("compile", func(t *testing.T) {
						toPath, err := Compile(path, o)
						if err != nil {
							t.Error(err.Error())
							return
						}
						for _, v := range compileCases {
							io := v.(a)
							input, output := io[0], io[1]
							var o1 *Options
							if len(io) >= 3 && io[2] != nil {
								o1 = io[2].(*Options)
							}
							if output != nil {
								t.Run("should compile using "+inspect(input), func(t *testing.T) {
									result := toPath(input, o1)
									if !reflect.DeepEqual(result, output) {
										t.Errorf("got %v want %v", result, output)
									}
								})
							} else {
								t.Run("should not compile using "+inspect(input), func(t *testing.T) {
									defer func() {
										if err := recover(); err == nil {
											t.Errorf("got %v want panic", err)
										}
									}()
									toPath(input, o1)
								})
							}
						}
					})
				} else {
					t.Run("should parse keys", func(t *testing.T) {
						tTokens := make([]interface{}, 0, len(tokens))
						for _, token := range tokens {
							if _, ok := token.(string); !ok {
								tTokens = append(tTokens, token)
							}
						}
						if !keysAndTokensDeepEqual(*keys, tTokens) {
							t.Errorf("got %v want %v", *keys, tTokens)
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
						input, output := io[0], io[1]
						message := " not "
						var o a
						if output != nil {
							message = " "
							o = output.(a)
						}
						message = "should" + message + "match " + inspect(input)
						t.Run(message, func(t *testing.T) {
							result := exec(r, input.(string))
							if !deepEqual(result, o) {
								t.Errorf("got %v want %v", result, output)
							}
						})
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
			toPath(nil, nil)
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
			toPath(map[interface{}]interface{}{"foo": "abc"}, nil)
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
			toPath(map[interface{}]interface{}{"foo": []interface{}{}}, nil)
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
			toPath(map[interface{}]interface{}{"foo": []interface{}{}}, nil)
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
			toPath(map[interface{}]interface{}{"foo": []interface{}{1, 2, 3, "a"}}, nil)
		})
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

func keysAndTokensDeepEqual(keys []Key, tokens []interface{}) bool {
	if len(keys) != len(tokens) {
		return false
	}

	if len(keys) == 0 && len(tokens) == 0 {
		return true
	}

	if keys == nil || tokens == nil {
		return false
	}

	for i, v := range keys {
		if !reflect.DeepEqual(v, tokens[i]) {
			return false
		}
	}

	return true
}
