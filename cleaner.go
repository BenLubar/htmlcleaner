package htmlcleaner

import (
	"bytes"
	"fmt"
	"io"
	"net/url"
	"strings"

	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

// DefaultMaxDepth is the default maximum depth of the node trees returned by
// Parse.
const DefaultMaxDepth = 100

// Preprocess escapes disallowed tags in a cleaner way, but does not fix
// nesting problems. Use with Clean.
func Preprocess(config *Config, fragment string) (string, error) {
	if config == nil {
		config = DefaultConfig
	}

	var buf bytes.Buffer

	t := html.NewTokenizer(strings.NewReader(fragment))
	for {
		switch tok := t.Next(); tok {
		case html.ErrorToken:
			err := t.Err()
			if err == io.EOF {
				err = nil
			}
			return buf.String(), err
		case html.TextToken:
			if _, err := buf.Write(t.Raw()); err != nil {
				return "", err
			}
		case html.StartTagToken, html.EndTagToken, html.SelfClosingTagToken:
			raw := string(t.Raw())
			tagName, _ := t.TagName()
			tag := atom.Lookup(tagName)
			if _, ok := config.Elem[tag]; !ok {
				raw = html.EscapeString(raw)
			}
			if _, err := buf.WriteString(raw); err != nil {
				return "", err
			}
		case html.CommentToken:
			raw := string(t.Raw())
			if config.EscapeComments {
				raw = html.EscapeString(raw)
			}
			if _, err := buf.WriteString(raw); err != nil {
				return "", err
			}
		case html.DoctypeToken:
			if _, err := buf.WriteString(html.EscapeString(string(t.Raw()))); err != nil {
				return "", err
			}
		default:
			panic(fmt.Sprintf("htmlcleaner: unhandled token type: %d", tok))
		}
	}
}

// Parse is a convenience wrapper that calls ParseDepth with DefaultMaxDepth.
func Parse(fragment string) []*html.Node {
	return ParseDepth(fragment, DefaultMaxDepth)
}

// ParseDepth is a convenience function that wraps html.ParseFragment but takes
// a string instead of an io.Reader and omits deep trees.
func ParseDepth(fragment string, maxDepth int) []*html.Node {
	nodes, err := html.ParseFragment(strings.NewReader(fragment), &html.Node{
		Type:     html.ElementNode,
		Data:     "div",
		DataAtom: atom.Div,
	})
	if err != nil {
		// should never happen
		panic(err)
	}

	for _, n := range nodes {
		forceMaxDepth(n, maxDepth)
	}

	return nodes
}

// Render is a convenience function that wraps html.Render and renders to a
// string instead of an io.Writer.
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
	return Render(CleanNodes(c, Parse(fragment))...)
}

// CleanNodes calls CleanNode on each node, and additionally wraps inline
// elements in <p> tags and wraps dangling <li> tags in <ul> tags.
func CleanNodes(c *Config, nodes []*html.Node) []*html.Node {
	if c == nil {
		c = DefaultConfig
	}

	for i, n := range nodes {
		nodes[i] = CleanNode(c, n)
		if nodes[i].DataAtom == atom.Li {
			nodes[i] = wrapElement(atom.Ul, nodes[i])
		}
	}

	if c.WrapText {
		wrapped := make([]*html.Node, 0, len(nodes))
		var wrapper *html.Node
		for _, n := range nodes {
			if n.Type == html.ElementNode {
				switch n.DataAtom {
				case atom.Address, atom.Article, atom.Aside, atom.Blockquote, atom.Center, atom.Dd, atom.Details, atom.Dir, atom.Div, atom.Dl, atom.Dt, atom.Fieldset, atom.Figcaption, atom.Figure, atom.Footer, atom.Form, atom.H1, atom.H2, atom.H3, atom.H4, atom.H5, atom.H6, atom.Header, atom.Hgroup, atom.Hr, atom.Li, atom.Listing, atom.Menu, atom.Nav, atom.Ol, atom.P, atom.Plaintext, atom.Pre, atom.Section, atom.Summary, atom.Table, atom.Ul:
					if wrapper != nil {
						wrapped = append(wrapped, wrapper)
						wrapper = nil
					}
					wrapped = append(wrapped, n)
					continue
				}
			}
			if wrapper == nil && n.Type == html.TextNode && strings.TrimSpace(n.Data) == "" {
				wrapped = append(wrapped, n)
				continue
			}
			if wrapper == nil {
				wrapper = &html.Node{
					Type:     html.ElementNode,
					Data:     "p",
					DataAtom: atom.P,
				}
			}
			wrapper.AppendChild(n)
		}
		if wrapper != nil {
			wrapped = append(wrapped, wrapper)
		}
		nodes = wrapped
	}

	return nodes
}

func wrapElement(a atom.Atom, node *html.Node) *html.Node {
	wrapper := &html.Node{
		Type:        html.ElementNode,
		Data:        a.String(),
		DataAtom:    a,
		PrevSibling: node.PrevSibling,
		NextSibling: node.NextSibling,
	}
	if wrapper.PrevSibling != nil {
		wrapper.PrevSibling.NextSibling = wrapper
	}
	if wrapper.NextSibling != nil {
		wrapper.NextSibling.PrevSibling = wrapper
	}
	node.Parent, node.PrevSibling, node.NextSibling = nil, nil, nil
	wrapper.AppendChild(node)
	return wrapper
}

func text(s string) *html.Node {
	return &html.Node{Type: html.TextNode, Data: s}
}

// CleanNode cleans an HTML node using the specified config. Doctype nodes and
// nodes that have a specified namespace are converted to text. Text nodes,
// document nodes, etc. are returned as-is. Element nodes are recursively
// checked for legality and have their attributes checked for legality as well.
// Elements with illegal attributes are copied and the problematic attributes
// are removed. Elements that are not in the set of legal elements are replaced
// with a textual version of their source code.
func CleanNode(c *Config, n *html.Node) *html.Node {
	return filterNode(c, n)
}

func filterNode(c *Config, n *html.Node) *html.Node {
	if c == nil {
		c = DefaultConfig
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
	return cleanNode(c, n)
}

func cleanNode(c *Config, n *html.Node) *html.Node {
	if allowedAttr, ok := c.Elem[n.DataAtom]; ok {
		// copy the node
		tmp := *n
		n = &tmp

		cleanChildren(c, n)

		haveSrc := false

		attrs := n.Attr
		n.Attr = make([]html.Attribute, 0, len(attrs))
		for _, attr := range attrs {
			a := atom.Lookup([]byte(attr.Key))
			if attr.Namespace != "" || (!allowedAttr[a] && !c.Attr[a]) {
				continue
			}

			if !c.AllowJavascriptURL && !cleanURL(c, a, &attr) {
				continue
			}

			if re, ok := c.AttrMatch[n.DataAtom][a]; ok && !re.MatchString(attr.Val) {
				continue
			}

			haveSrc = haveSrc || a == atom.Src

			n.Attr = append(n.Attr, attr)
		}

		if n.DataAtom == atom.Img && !haveSrc {
			// replace it with an empty text node
			return &html.Node{Type: html.TextNode}
		}

		return n
	}
	return text(html.UnescapeString(Render(n)))
}

var allowedURLSchemes = map[string]bool{
	"http":   true,
	"https":  true,
	"mailto": true,
	"data":   true,
	"":       true,
}

func cleanURL(c *Config, a atom.Atom, attr *html.Attribute) bool {
	if a != atom.Href && a != atom.Src && a != atom.Poster {
		return true
	}

	u, err := url.Parse(attr.Val)
	if err != nil {
		return false
	}
	if !allowedURLSchemes[u.Scheme] {
		return false
	}
	if c.ValidateURL != nil && !c.ValidateURL(u) {
		return false
	}
	attr.Val = u.String()
	return true
}

func cleanChildren(c *Config, parent *html.Node) {
	var children []*html.Node
	for child := parent.FirstChild; child != nil; child = child.NextSibling {
		children = append(children, filterNode(c, child))
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

func forceMaxDepth(n *html.Node, depth int) {
	if depth == 0 {
		n.Type = html.TextNode
		n.FirstChild, n.LastChild = nil, nil
		n.Attr = nil
		n.Data = "[omitted]"
		return
	}

	if n.Type != html.ElementNode {
		return
	}

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		forceMaxDepth(c, depth-1)
	}
}
