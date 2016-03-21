package htmlcleaner

import "regexp"
import "golang.org/x/net/html/atom"

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

	// If true, HTML comments are turned into text.
	EscapeComments bool

	// Wrap text nodes in at least one tag.
	WrapText bool

	// Attributes with these names must have matching values.
	AttrMatch map[atom.Atom]*regexp.Regexp
}

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

		atom.Pre:  nil,
		atom.Code: nil,
		atom.Kbd:  nil,
		atom.Tt:   nil,

		atom.Details: nil,
		atom.Summary: nil,
	},

	Attr: map[atom.Atom]bool{
		atom.Title: true,
	},

	AllowJavascriptURL: false,

	EscapeComments: false,

	WrapText: false,
}
