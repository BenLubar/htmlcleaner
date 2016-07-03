package htmlcleaner

import "fmt"

func ExampleClean() {
	fragment, err := Preprocess(nil, `<a href="http://golang.org/" onclick="malicious()" title="Go">hello</a>
<some tag that doesn't exist>
<script>malicious()</script>`)
	if err != nil {
		panic(err)
	}
	fmt.Println(Clean(nil, fragment))

	// Output:
	// <a href="http://golang.org/" title="Go">hello</a>
	// &lt;some tag that doesn&#39;t exist&gt;
	// &lt;script&gt;malicious()&lt;/script&gt;
}
