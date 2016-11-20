package htmlcleaner

import (
	"net/url"
	"regexp"

	"golang.org/x/net/html/atom"
)

// Config holds the settings for htmlcleaner.
type Config struct {
	elem       map[atom.Atom]map[atom.Atom]*regexp.Regexp
	attr       map[atom.Atom]struct{}
	elemCustom map[string]map[string]*regexp.Regexp
	attrCustom map[string]struct{}
	wrap       map[atom.Atom]struct{}
	wrapCustom map[string]struct{}

	// A custom URL validation function. If it is set and returns false,
	// the attribute will be removed. Called for attributes such as src
	// and href.
	ValidateURL func(*url.URL) bool

	// If true, HTML comments are turned into text.
	EscapeComments bool

	// Wrap text nodes in at least one tag.
	WrapText bool
}

// Elem ensures an element name is allowed. The receiver is returned to
// allow call chaining.
func (c *Config) Elem(names ...string) *Config {
	for _, name := range names {
		if a := atom.Lookup([]byte(name)); a != 0 {
			c.ElemAtom(a)
			continue
		}

		if c.elemCustom == nil {
			c.elemCustom = make(map[string]map[string]*regexp.Regexp)
		}

		if _, ok := c.elemCustom[name]; !ok {
			c.elemCustom[name] = nil
		}
	}

	return c
}

// ElemAtom ensures an element name is allowed. The receiver is returned to
// allow call chaining.
func (c *Config) ElemAtom(elem ...atom.Atom) *Config {
	if c.elem == nil {
		c.elem = make(map[atom.Atom]map[atom.Atom]*regexp.Regexp)
	}

	for _, a := range elem {
		if _, ok := c.elem[a]; !ok {
			c.elem[a] = nil
		}
	}

	return c
}

// GlobalAttr allows an attribute name on all allowed elements. The
// receiver is returned to allow call chaining.
func (c *Config) GlobalAttr(names ...string) *Config {
	for _, name := range names {
		if a := atom.Lookup([]byte(name)); a != 0 {
			c.GlobalAttrAtom(a)
			continue
		}

		if c.attrCustom == nil {
			c.attrCustom = make(map[string]struct{})
		}

		c.attrCustom[name] = struct{}{}
	}

	return c
}

// GlobalAttrAtom allows an attribute name on all allowed elements. The
// receiver is returned to allow call chaining.
func (c *Config) GlobalAttrAtom(a atom.Atom) *Config {
	if c.attr == nil {
		c.attr = make(map[atom.Atom]struct{})
	}

	c.attr[a] = struct{}{}

	return c
}

// ElemAttr allows an attribute name on the specified element. The
// receiver is returned to allow call chaining.
func (c *Config) ElemAttr(elem string, attr ...string) *Config {
	for _, a := range attr {
		c.ElemAttrMatch(elem, a, nil)
	}
	return c
}

// ElemAttrAtom allows an attribute name on the specified element. The
// receiver is returned to allow call chaining.
func (c *Config) ElemAttrAtom(elem atom.Atom, attr ...atom.Atom) *Config {
	for _, a := range attr {
		c.ElemAttrAtomMatch(elem, a, nil)
	}
	return c
}

// ElemAttrMatch allows an attribute name on the specified element, but
// only if the value matches a regular expression. The receiver is returned to
// allow call chaining.
func (c *Config) ElemAttrMatch(elem, attr string, match *regexp.Regexp) *Config {
	if e, a := atom.Lookup([]byte(elem)), atom.Lookup([]byte(attr)); e != 0 && a != 0 {
		return c.ElemAttrAtomMatch(e, a, match)
	}

	if c.elemCustom == nil {
		c.elemCustom = make(map[string]map[string]*regexp.Regexp)
	}

	attrs := c.elemCustom[elem]
	if attrs == nil {
		attrs = make(map[string]*regexp.Regexp)
		c.elemCustom[elem] = attrs
	}

	attrs[attr] = match

	return c
}

// ElemAttrAtomMatch allows an attribute name on the specified element,
// but only if the value matches a regular expression. The receiver is returned
// to allow call chaining.
func (c *Config) ElemAttrAtomMatch(elem, attr atom.Atom, match *regexp.Regexp) *Config {
	if c.elem == nil {
		c.elem = make(map[atom.Atom]map[atom.Atom]*regexp.Regexp)
	}

	attrs := c.elem[elem]
	if attrs == nil {
		attrs = make(map[atom.Atom]*regexp.Regexp)
		c.elem[elem] = attrs
	}

	attrs[attr] = match

	return c
}

// WrapTextInside makes an element's children behave as if they are root nodes
// in the context of WrapText. The receiver is returned to allow call chaining.
func (c *Config) WrapTextInside(names ...string) *Config {
	if c.wrapCustom == nil {
		c.wrapCustom = make(map[string]struct{})
	}

	for _, name := range names {
		if a := atom.Lookup([]byte(name)); a != 0 {
			c.WrapTextInsideAtom(a)
			continue
		}

		c.wrapCustom[name] = struct{}{}
	}

	return c
}

// WrapTextInsideAtom makes an element's children behave as if they are root
// nodes in the context of WrapText. The receiver is returned to allow call
// chaining.
func (c *Config) WrapTextInsideAtom(elem ...atom.Atom) *Config {
	if c.wrap == nil {
		c.wrap = make(map[atom.Atom]struct{})
	}

	for _, a := range elem {
		c.wrap[a] = struct{}{}
	}

	return c
}

// DefaultConfig is the default settings for htmlcleaner.
var DefaultConfig = (&Config{
	ValidateURL: SafeURLScheme,
}).GlobalAttrAtom(atom.Title).
	ElemAttrAtom(atom.A, atom.Href).
	ElemAttrAtom(atom.Img, atom.Src, atom.Alt).
	ElemAttrAtom(atom.Video, atom.Src, atom.Poster, atom.Controls).
	ElemAttrAtom(atom.Audio, atom.Src, atom.Controls).
	ElemAtom(atom.B, atom.I, atom.U, atom.S).
	ElemAtom(atom.Em, atom.Strong, atom.Strike).
	ElemAtom(atom.Big, atom.Small, atom.Sup, atom.Sub).
	ElemAtom(atom.Ins, atom.Del).
	ElemAtom(atom.Abbr, atom.Address, atom.Cite, atom.Q).
	ElemAtom(atom.P, atom.Blockquote, atom.Pre).
	ElemAtom(atom.Code, atom.Kbd, atom.Tt).
	ElemAttrAtom(atom.Details, atom.Open).
	ElemAtom(atom.Summary)
