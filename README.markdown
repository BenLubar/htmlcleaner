# htmlcleaner
--
    import "github.com/BenLubar/htmlcleaner"


## Usage

```go
var DefaultConfig = &Config{
	Elem: map[atom.Atom]map[atom.Atom]bool{
		atom.A: {
			atom.Href: true,
		},
		atom.Img: {
			atom.Src: true,
			atom.Alt: true,
		},
		atom.Video: {
			atom.Src:      true,
			atom.Poster:   true,
			atom.Controls: true,
		},
		atom.Audio: {
			atom.Src:      true,
			atom.Controls: true,
		},

		atom.B: nil,
		atom.I: nil,
		atom.U: nil,
		atom.S: nil,

		atom.Em:     nil,
		atom.Strong: nil,
		atom.Strike: nil,

		atom.Big:   nil,
		atom.Small: nil,
		atom.Sup:   nil,
		atom.Sub:   nil,

		atom.Ins: nil,
		atom.Del: nil,

		atom.Abbr:    nil,
		atom.Address: nil,
		atom.Cite:    nil,
		atom.Q:       nil,

		atom.P:          nil,
		atom.Blockquote: nil,
	},

	Attr: map[atom.Atom]bool{
		atom.Title: true,
	},

	AllowJavascriptURL: false,
}
```

#### func  Clean

```go
func Clean(c *Config, fragment string) string
```
Clean a fragment of HTML using the specified Config, or the DefaultConfig if it
is nil.

#### func  CleanNode

```go
func CleanNode(c *Config, n *html.Node) *html.Node
```
Clean an HTML node using the specified config. Doctype nodes and nodes that have
a specified namespace are converted to text. Text nodes, document nodes, etc.
are returned as-is. Element nodes are recursively checked for legality and have
their attributes checked for legality as well. Elements with illegal attributes
are copied and the problematic attributes are removed. Elements that are not in
the set of legal elements are replaced with a textual version of their source
code.

#### func  Parse

```go
func Parse(fragment string) []*html.Node
```
Convenience function that takes a string instead of an io.Reader.

#### func  Render

```go
func Render(nodes ...*html.Node) string
```
Convenience function that writes a string instead of an io.Writer.

#### type Config

```go
type Config struct {
	// element => attribute => allowed
	// if an element's attribute map exists, even if it is nil,
	// the element is legal.
	Elem map[atom.Atom]map[atom.Atom]bool

	// attribute => allowed (for all legal elements)
	Attr map[atom.Atom]bool

	// If true, URLs starting with javascript: are allowed in <a href>
	// <img src> <video src> <audio src>, etc. If false, attributes with
	// JavaScript URLs are removed.
	AllowJavascriptURL bool
}
```

--
**godocdown** http://github.com/robertkrimen/godocdown
