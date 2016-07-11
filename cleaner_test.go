package htmlcleaner

import (
	"io"
	"strings"
	"testing"

	"golang.org/x/net/html/atom"
)

var testTableClean = []struct {
	Input  string
	Output string
	Config *Config
}{
	{``, ``, nil},
	{`a`, `a`, nil},
	{`<a`, ``, nil},
	{`a<`, `a&lt;`, nil},
	{`<a href http://golang.org>`, `<a href=""></a>`, nil},
	{`<a href="http://golang.org">Go`, `<a href="http://golang.org">Go</a>`, nil},
	{`<a href="http://golang.org">Go</a></a>`, `<a href="http://golang.org">Go</a>`, nil},
	{`<a href="javascript:malicious()">`, `<a></a>`, nil},
	{`<b><i>hello</b></i>`, `<b><i>hello</i></b>`, nil},
	{`<b><i>hello</b></i> <u>there`, `<b><i>hello</i></b> <u>there</u>`, nil},
	{`<img href alt></img>`, ``, nil},
	{`<p><p><p><p>`, `<p></p><p></p><p></p><p></p>`, nil},
	{`<script>foo.bar < baz</script>`, `&lt;script&gt;foo.bar &lt; baz&lt;/script&gt;`, nil},
	{`&`, `&amp;`, nil},
	{`&amp;`, `&amp;`, nil},
	{`<invalidtag>&#34;</invalidtag>`, `&lt;invalidtag&gt;&#34;&lt;/invalidtag&gt;`, nil},
	{`<li>`, `<ul><li></li></ul>`, &Config{Elem: map[atom.Atom]map[atom.Atom]bool{atom.Ul: nil, atom.Li: nil}}},
	{`<a href="https://www.%google.com">google</a>`, `<a>google</a>`, nil},
	{`<a href="https://www.%google.com">google</a>`, `<p><a>google</a></p>`, &Config{Elem: DefaultConfig.Elem, WrapText: true}},
	{`foo>bar`, `foo&gt;bar`, nil},
	{`>bar`, `&gt;bar`, nil},
	{`foo>`, `foo&gt;`, nil},
	{`<!--comment-->`, `<!--comment-->`, nil},
	{`<!--comment-->`, `&lt;!--comment--&gt;`, &Config{EscapeComments: true}},
	{`<![CDATA[ foo ]]>`, `<!--[CDATA[ foo ]]-->`, nil},
	{`<![CDATA[ foo ]]>`, `&lt;!--[CDATA[ foo ]]--&gt;`, &Config{EscapeComments: true}},
	{`<?xml version="1.0"?>`, `<!--?xml version="1.0"?-->`, nil},
	{`<?xml version="1.0"?>`, `&lt;!--?xml version=&#34;1.0&#34;?--&gt;`, &Config{EscapeComments: true}},
	{`<!DOCTYPE html>`, ``, nil},
	{`<!DOCTYPE html>`, ``, &Config{EscapeComments: true}},
	{`<?php echo mysql_real_escape_string('foo'); ?>`, `<!--?php echo mysql_real_escape_string('foo'); ?-->`, nil},
	{`<?php echo mysql_real_escape_string('foo'); ?>`, `&lt;!--?php echo mysql_real_escape_string(&#39;foo&#39;); ?--&gt;`, &Config{EscapeComments: true}},
	{strings.Repeat(`<small>a `, 250), strings.Repeat(`<small>a `, 99) + "<small>[omitted]" + strings.Repeat(`</small>`, 100), nil},
	{`hello <em>world`, `<p>hello <em>world</em></p>`, &Config{Elem: DefaultConfig.Elem, WrapText: true}},
	{`<p>hello</p> <p>world</p>`, `<p>hello</p> <p>world</p>`, &Config{Elem: DefaultConfig.Elem, WrapText: true}},
	{`<em>hello <p>world</p>`, `<p><em>hello </em></p><p><em>world</em></p><p></p>`, &Config{Elem: DefaultConfig.Elem, WrapText: true}},
}

func TestClean(t *testing.T) {
	for i, tt := range testTableClean {
		actual, expected := Clean(tt.Config, tt.Input), tt.Output

		if actual != expected {
			t.Logf("%d is %+v", i, tt)
			t.Logf("%d: expected %q", i, expected)
			t.Logf("%d: actual   %q", i, actual)
			t.Errorf("%d: expected != actual", i)
		}
	}
}

var testTablePreprocess = []struct {
	Input  string
	Output string
	Config *Config
}{
	{``, ``, nil},
	{`a`, `a`, nil},
	{`<insert text here>`, `&lt;insert text here&gt;`, nil},
	{`foo<bar`, `foo&lt;bar`, nil},
	{`<bar`, `&lt;bar`, nil},
	{`foo<`, `foo<`, nil},
	{`foo>bar`, `foo>bar`, nil},
	{`>bar`, `>bar`, nil},
	{`foo>`, `foo>`, nil},
	{`<!--comment-->`, `<!--comment-->`, nil},
	{`<!--comment-->`, `&lt;!--comment--&gt;`, &Config{EscapeComments: true}},
	{`<![CDATA[ foo ]]>`, `&lt;![CDATA[ foo ]]&gt;`, nil},
	{`<![CDATA[ foo ]]>`, `&lt;![CDATA[ foo ]]&gt;`, &Config{EscapeComments: true}},
	{`<?xml version="1.0"?>`, `&lt;?xml version=&#34;1.0&#34;?&gt;`, nil},
	{`<?xml version="1.0"?>`, `&lt;?xml version=&#34;1.0&#34;?&gt;`, &Config{EscapeComments: true}},
	{`<!DOCTYPE html>`, `&lt;!DOCTYPE html&gt;`, nil},
	{`<!DOCTYPE html>`, `&lt;!DOCTYPE html&gt;`, &Config{EscapeComments: true}},
	{`<?php echo mysql_real_escape_string('foo'); ?>`, `&lt;?php echo mysql_real_escape_string(&#39;foo&#39;); ?&gt;`, nil},
	{`<?php echo mysql_real_escape_string('foo'); ?>`, `&lt;?php echo mysql_real_escape_string(&#39;foo&#39;); ?&gt;`, &Config{EscapeComments: true}},
}

func TestPreprocess(t *testing.T) {
	for i, tt := range testTablePreprocess {
		actual, expected := Preprocess(tt.Config, tt.Input), tt.Output

		if actual != expected {
			t.Logf("%d is %+v", i, tt)
			t.Logf("%d: expected %q", i, expected)
			t.Logf("%d: actual   %q", i, actual)
			t.Errorf("%d: expected != actual", i)
		}
	}
}

func TestExpectError(t *testing.T) {
	defer func() {
		if r := recover(); r != "htmlcleaner: unexpected error: EOF" {
			t.Errorf("expectError paniced with %v", r)
		}
	}()

	expectError(io.EOF, nil)
}
