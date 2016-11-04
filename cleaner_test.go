package htmlcleaner

import (
	"io"
	"strings"
	"testing"

	"golang.org/x/net/html/atom"
)

type testTable struct {
	Name   string
	Input  string
	Output string
	Config *Config
}

func doTableTest(f func(*Config, string) string, t *testing.T, table []testTable) {
	for _, tt := range table {
		t.Run(tt.Name, func(t *testing.T) {
			actual, expected := f(tt.Config, tt.Input), tt.Output

			if actual != expected {
				t.Logf("expected %q", expected)
				t.Logf("actual   %q", actual)
				t.Fatal("expected != actual")
			}
		})
	}
}

var wrapConfig = func() *Config {
	c := *DefaultConfig

	c.WrapText = true

	return &c
}()

var testTableClean = []testTable{
	{"Empty", ``, ``, nil},
	{"PlainText", `a`, `a`, nil},
	{"UnterminatedOpenTag", `<a`, ``, nil},
	{"LessThanAtEnd", `a<`, `a&lt;`, nil},
	{"LinkMissingPunctuation", `<a href http://golang.org>`, `<a href=""></a>`, nil},
	{"LinkMissingClosingTag", `<a href="http://golang.org">Go`, `<a href="http://golang.org">Go</a>`, nil},
	{"LinkTwoClosingTags", `<a href="http://golang.org">Go</a></a>`, `<a href="http://golang.org">Go</a>`, nil},
	{"LinkJavaScript", `<a href="javascript:malicious()">`, `<a></a>`, nil},
	{"InvalidNesting", `<b><i>hello</b></i>`, `<b><i>hello</i></b>`, nil},
	{"InvalidNestingUnclosed", `<b><i>hello</b></i> <u>there`, `<b><i>hello</i></b> <u>there</u>`, nil},
	{"ImageInvalid", `<img href alt></img>`, ``, nil},
	{"FourParagraphs", `<p><p><p><p>`, `<p></p><p></p><p></p><p></p>`, nil},
	{"ScriptLessThan", `<script>foo.bar < baz</script>`, `&lt;script&gt;foo.bar &lt; baz&lt;/script&gt;`, nil},
	{"Ampersand", `&`, `&amp;`, nil},
	{"AmpersandEntity", `&amp;`, `&amp;`, nil},
	{"InvalidTagEntity", `<invalidtag>&#34;</invalidtag>`, `&lt;invalidtag&gt;&#34;&lt;/invalidtag&gt;`, nil},
	{"StrayListItem", `<li>`, `<ul><li></li></ul>`, (&Config{}).ElemAtom(atom.Ul, atom.Li)},
	{"LinkPercent", `<a href="https://www.%google.com">google</a>`, `<a>google</a>`, nil},
	{"LinkPercentWrap", `<a href="https://www.%google.com">google</a>`, `<p><a>google</a></p>`, wrapConfig},
	{"GreaterThanInfix", `foo>bar`, `foo&gt;bar`, nil},
	{"GreaterThanPrefix", `>bar`, `&gt;bar`, nil},
	{"GreaterThanSuffix", `foo>`, `foo&gt;`, nil},
	{"Comment", `<!--comment-->`, `<!--comment-->`, nil},
	{"CommentEscaped", `<!--comment-->`, `&lt;!--comment--&gt;`, &Config{EscapeComments: true}},
	{"CDATA", `<![CDATA[ foo ]]>`, `<!--[CDATA[ foo ]]-->`, nil},
	{"CDATAEscaped", `<![CDATA[ foo ]]>`, `&lt;!--[CDATA[ foo ]]--&gt;`, &Config{EscapeComments: true}},
	{"XML", `<?xml version="1.0"?>`, `<!--?xml version="1.0"?-->`, nil},
	{"XMLEscaped", `<?xml version="1.0"?>`, `&lt;!--?xml version=&#34;1.0&#34;?--&gt;`, &Config{EscapeComments: true}},
	{"Doctype", `<!DOCTYPE html>`, ``, nil},
	{"DoctypeEscaped", `<!DOCTYPE html>`, ``, &Config{EscapeComments: true}},
	{"PHP", `<?php echo mysql_real_escape_string('foo'); ?>`, `<!--?php echo mysql_real_escape_string('foo'); ?-->`, nil},
	{"PHPEscaped", `<?php echo mysql_real_escape_string('foo'); ?>`, `&lt;!--?php echo mysql_real_escape_string(&#39;foo&#39;); ?--&gt;`, &Config{EscapeComments: true}},
	{"Small250", strings.Repeat(`<small>a `, 250), strings.Repeat(`<small>a `, 99) + "<small>[omitted]" + strings.Repeat(`</small>`, 100), nil},
	{"WrapUnclosed", `hello <em>world`, `<p>hello <em>world</em></p>`, wrapConfig},
	{"WrapStraySpace", `<p>hello</p> <p>world</p>`, `<p>hello</p> <p>world</p>`, wrapConfig},
	{"WrapInvalidNesting", `<em>hello <p>world</p>`, `<p><em>hello </em></p><p><em>world</em></p><p></p>`, wrapConfig},
}

func TestClean(t *testing.T) {
	doTableTest(Clean, t, testTableClean)
}

var testTablePreprocess = []testTable{
	{"Empty", ``, ``, nil},
	{"NoMarkup", `a`, `a`, nil},
	{"NonHTML", `<insert text here>`, `&lt;insert text here&gt;`, nil},
	{"LessThanInfix", `foo<bar`, `foo&lt;bar`, nil},
	{"LessThanPrefix", `<bar`, `&lt;bar`, nil},
	{"LessThanSuffix", `foo<`, `foo<`, nil},
	{"GreaterThanInfix", `foo>bar`, `foo>bar`, nil},
	{"GreaterThanPrefix", `>bar`, `>bar`, nil},
	{"GreaterThanSuffix", `foo>`, `foo>`, nil},
	{"Comment", `<!--comment-->`, `<!--comment-->`, nil},
	{"CommentEscape", `<!--comment-->`, `&lt;!--comment--&gt;`, &Config{EscapeComments: true}},
	{"CDATA", `<![CDATA[ foo ]]>`, `&lt;![CDATA[ foo ]]&gt;`, nil},
	{"CDATAEscape", `<![CDATA[ foo ]]>`, `&lt;![CDATA[ foo ]]&gt;`, &Config{EscapeComments: true}},
	{"XML", `<?xml version="1.0"?>`, `&lt;?xml version=&#34;1.0&#34;?&gt;`, nil},
	{"XMLEscape", `<?xml version="1.0"?>`, `&lt;?xml version=&#34;1.0&#34;?&gt;`, &Config{EscapeComments: true}},
	{"Doctype", `<!DOCTYPE html>`, `&lt;!DOCTYPE html&gt;`, nil},
	{"DoctypeEscape", `<!DOCTYPE html>`, `&lt;!DOCTYPE html&gt;`, &Config{EscapeComments: true}},
	{"PHP", `<?php echo mysql_real_escape_string('foo'); ?>`, `&lt;?php echo mysql_real_escape_string(&#39;foo&#39;); ?&gt;`, nil},
	{"PHPEscape", `<?php echo mysql_real_escape_string('foo'); ?>`, `&lt;?php echo mysql_real_escape_string(&#39;foo&#39;); ?&gt;`, &Config{EscapeComments: true}},
}

func TestPreprocess(t *testing.T) {
	doTableTest(Preprocess, t, testTablePreprocess)
}

func TestExpectError(t *testing.T) {
	defer func() {
		if r := recover(); r != "htmlcleaner: unexpected error: EOF" {
			t.Errorf("expectError paniced with %v", r)
		}
	}()

	expectError(io.EOF, nil)
}
