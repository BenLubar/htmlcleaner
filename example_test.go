package htmlcleaner

import "fmt"

func ExampleClean() {
	fmt.Println(Clean(nil, `<a href="http://golang.org/" onclick="malicious()" title="Go">hello</a> <script>malicious()</script>`))

	// Output:
	// <a href="http://golang.org/" title="Go">hello</a> &lt;script&gt;malicious()&lt;/script&gt;
}
