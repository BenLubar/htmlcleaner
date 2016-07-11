package htmlcleaner_test

import (
	"fmt"
	"net/url"
	"regexp"

	"golang.org/x/net/html/atom"

	"github.com/BenLubar/htmlcleaner"
)

func ExampleClean() {
	config := &htmlcleaner.Config{
		Elem: map[atom.Atom]map[atom.Atom]bool{
			atom.Span: {
				atom.Class: true,
			},
			atom.A: {
				atom.Href: true,
			},
		},
		AttrMatch: map[atom.Atom]map[atom.Atom]*regexp.Regexp{
			atom.Span: {
				atom.Class: regexp.MustCompile(`\Afa-spin\z`),
			},
		},
		ValidateURL: func(u *url.URL) bool {
			return u.Scheme != "http"
		},
	}

	fmt.Println(htmlcleaner.Clean(config, htmlcleaner.Preprocess(config, `<span class="fa-spin">[whee]</span>
<span class="hello">[aww]</span>
<a href="https://www.google.com">Google</a>
<a href="http://www.google.com">Google</a>
<some tag that doesn't exist>`)))

	// Output:
	// <span class="fa-spin">[whee]</span>
	// <span>[aww]</span>
	// <a href="https://www.google.com">Google</a>
	// <a>Google</a>
	// &lt;some tag that doesn&#39;t exist&gt;
}

func ExampleCleanNode() {
	var config *htmlcleaner.Config = nil

	nodes := htmlcleaner.Parse(`<a href="http://golang.org/" onclick="malicious()" title="Go">hello</a>
<script>malicious()</script>`)

	for i, n := range nodes {
		nodes[i] = htmlcleaner.CleanNode(config, n)
	}

	fmt.Println(htmlcleaner.Render(nodes...))

	// Output:
	// <a href="http://golang.org/" title="Go">hello</a>
	// &lt;script&gt;malicious()&lt;/script&gt;
}
