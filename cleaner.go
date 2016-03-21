package htmlcleaner

import (
	"bytes"
	"strings"

	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

// Convenience function that takes a string instead of an io.Reader.
func Parse(fragment string) []*html.Node {
	nodes, err := html.ParseFragment(strings.NewReader(fragment), &html.Node{
		Type:     html.ElementNode,
		Data:     "div",
		DataAtom: atom.Div,
	})
	if err != nil {
		// should never happen
		panic(err)
	}

	return nodes
}

// Convenience function that writes a string instead of an io.Writer.
func Render(nodes ...*html.Node) string {
	var buf bytes.Buffer

	for _, n := range nodes {
		if err := html.Render(&buf, n); err != nil {
			// should never happen
			panic(err)
		}
	}

	return string(buf.Bytes())
}

// Clean a fragment of HTML using the specified Config, or the DefaultConfig
// if it is nil.
func Clean(c *Config, fragment string) string {
	nodes := Parse(fragment)

	if c == nil {
		c = DefaultConfig
	}

	for i, n := range nodes {
		nodes[i] = CleanNode(c, n)
		if c.WrapText && nodes[i].Type == html.TextNode {
			n := &html.Node{
				Type:     html.ElementNode,
				Data:     "span",
				DataAtom: atom.Span,
			}
			n.AppendChild(nodes[i])
			nodes[i] = n
		}
	}

	return Render(nodes...)
}

func text(s string) *html.Node {
	return &html.Node{Type: html.TextNode, Data: s}
}

// Clean an HTML node using the specified config. Doctype nodes and nodes that
// have a specified namespace are converted to text. Text nodes, document nodes,
// etc. are returned as-is. Element nodes are recursively checked for legality
// and have their attributes checked for legality as well. Elements with illegal
// attributes are copied and the problematic attributes are removed. Elements
// that are not in the set of legal elements are replaced with a textual
// version of their source code.
func CleanNode(c *Config, n *html.Node) *html.Node {
	return cleanNodeMax(c, n, 100)
}

func cleanNodeMax(c *Config, n *html.Node, depth int) *html.Node {
	if c == nil {
		c = DefaultConfig
	}
	if depth == 0 {
		return text("[omitted]")
	}
	if n.Type == html.DoctypeNode {
		return text(Render(n))
	}
	if n.Type == html.CommentNode && c.EscapeComments {
		return text(Render(n))
	}
	if n.Type != html.ElementNode {
		return n
	}
	if n.Namespace != "" {
		return text(Render(n))
	}
	if allowedAttr, ok := c.Elem[n.DataAtom]; ok {
		// copy the node
		tmp := *n
		n = &tmp

		cleanChildren(c, n, depth)

		attr := n.Attr
		n.Attr = make([]html.Attribute, 0, len(attr))
		for _, a := range attr {
			aatom := atom.Lookup([]byte(a.Key))
			if a.Namespace != "" || (!allowedAttr[aatom] && !c.Attr[aatom]) {
				continue
			}

			if !c.AllowJavascriptURL && (aatom == atom.Href || aatom == atom.Src || aatom == atom.Poster) {
				if i := strings.IndexRune(a.Val, ':'); i >= 0 && strings.IndexRune(a.Val[:i], '/') < 0 {
					protocol := strings.ToLower(a.Val[:i])
					if protocol != "http" && protocol != "https" && protocol != "mailto" {
						continue
					}
				}
			}

			if re, ok := c.AttrMatch[aatom]; ok && !re.MatchString(a.Val) {
				continue
			}

			n.Attr = append(n.Attr, a)
		}

		return n
	}
	return text(html.UnescapeString(Render(n)))
}

func cleanChildren(c *Config, parent *html.Node, depth int) {
	var children []*html.Node
	for child := parent.FirstChild; child != nil; child = child.NextSibling {
		children = append(children, cleanNodeMax(c, child, depth-1))
	}

	for i, child := range children {
		child.Parent = parent
		if i == 0 {
			parent.FirstChild = child
		} else {
			child.PrevSibling = children[i-1]
		}
		if i == len(children)-1 {
			parent.LastChild = child
		} else {
			child.NextSibling = children[i+1]
		}
	}
}
