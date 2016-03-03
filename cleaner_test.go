package htmlcleaner

import "testing"

var testTable = []struct {
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
	{`<img src href alt></img>`, `<img src="" alt=""/>`, nil},
	{`<p><p><p><p>`, `<p></p><p></p><p></p><p></p>`, nil},
	{`<script>foo.bar < baz</script>`, `&lt;script&gt;foo.bar &lt; baz&lt;/script&gt;`, nil},
	{`&`, `&amp;`, nil},
	{`&amp;`, `&amp;`, nil},
	{`<invalidtag>&#34;</invalidtag>`, `&lt;invalidtag&gt;&#34;&lt;/invalidtag&gt;`, nil},
}

func TestCleaner(t *testing.T) {
	for i, tt := range testTable {
		actual, expected := Clean(tt.Config, tt.Input), tt.Output

		if actual != expected {
			t.Logf("%d is %+v", i, tt)
			t.Logf("%d: expected %q", i, expected)
			t.Logf("%d: actual   %q", i, actual)
			t.Errorf("%d: expected != actual", i)
		}
	}
}
