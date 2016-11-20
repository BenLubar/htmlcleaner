package htmlcleaner_test

import (
	"regexp"
	"testing"

	"github.com/BenLubar/htmlcleaner"
)

func TestConfig(t *testing.T) {
	run := func(name, input, expected1, expected2 string, c *htmlcleaner.Config, f func(*htmlcleaner.Config)) {
		if c == nil {
			c = &htmlcleaner.Config{}
		}
		t.Run(name, func(t *testing.T) {
			actual1 := htmlcleaner.Clean(c, htmlcleaner.Preprocess(c, input))
			if expected1 != actual1 {
				t.Errorf("BEFORE")
				t.Errorf("input:    %q", input)
				t.Errorf("expected: %q", expected1)
				t.Errorf("actual:   %q", actual1)
			}

			f(c)
			actual2 := htmlcleaner.Clean(c, htmlcleaner.Preprocess(c, input))
			if expected2 != actual2 {
				t.Errorf("AFTER")
				t.Errorf("input:    %q", input)
				t.Errorf("expected: %q", expected2)
				t.Errorf("actual:   %q", actual2)
			}
		})
	}

	run("Elem", `<p>Hello</p>`, `&lt;p&gt;Hello&lt;/p&gt;`, `<p>Hello</p>`, nil, func(c *htmlcleaner.Config) { c.Elem("p") })
	run("CustomElem", `<custom-element>Hello</custom-element>`, `&lt;custom-element&gt;Hello&lt;/custom-element&gt;`, `<custom-element>Hello</custom-element>`, nil, func(c *htmlcleaner.Config) { c.Elem("custom-element") })
	run("ElemAttr", `<p title="World">Hello</p>`, `<p>Hello</p>`, `<p title="World">Hello</p>`, (&htmlcleaner.Config{}).Elem("p"), func(c *htmlcleaner.Config) { c.ElemAttr("p", "title") })
	run("CustomElemAttr", `<custom-element title="World">Hello</custom-element>`, `<custom-element>Hello</custom-element>`, `<custom-element title="World">Hello</custom-element>`, (&htmlcleaner.Config{}).Elem("custom-element"), func(c *htmlcleaner.Config) { c.ElemAttr("custom-element", "title") })
	run("ElemCustomAttr", `<p data-original-title="World">Hello</p>`, `<p>Hello</p>`, `<p data-original-title="World">Hello</p>`, (&htmlcleaner.Config{}).Elem("p"), func(c *htmlcleaner.Config) { c.ElemAttr("p", "data-original-title") })
	run("CustomElemCustomAttr", `<custom-element data-original-title="World">Hello</custom-element>`, `<custom-element>Hello</custom-element>`, `<custom-element data-original-title="World">Hello</custom-element>`, (&htmlcleaner.Config{}).Elem("custom-element"), func(c *htmlcleaner.Config) { c.ElemAttr("custom-element", "data-original-title") })
	run("GlobalAttr", `<p title="World">Hello</p><custom-element title="Hello">World</custom-element>`, `<p>Hello</p><custom-element>World</custom-element>`, `<p title="World">Hello</p><custom-element title="Hello">World</custom-element>`, (&htmlcleaner.Config{}).Elem("p", "custom-element"), func(c *htmlcleaner.Config) { c.GlobalAttr("title") })
	run("AttrMatch", `<p title="Hello"></p><p title="World"></p>`, `<p title="Hello"></p><p title="World"></p>`, `<p></p><p title="World"></p>`, (&htmlcleaner.Config{}).ElemAttr("p", "title"), func(c *htmlcleaner.Config) { c.ElemAttrMatch("p", "title", regexp.MustCompile(`or`)) })
	run("CustomAttrMatch", `<p data-original-title="Hello"></p><p data-original-title="World"></p>`, `<p data-original-title="Hello"></p><p data-original-title="World"></p>`, `<p></p><p data-original-title="World"></p>`, (&htmlcleaner.Config{}).ElemAttr("p", "data-original-title"), func(c *htmlcleaner.Config) { c.ElemAttrMatch("p", "data-original-title", regexp.MustCompile(`or`)) })
	run("GlobalCustomAttr", `<p data-original-title="World">Hello</p><custom-element data-original-title="Hello">World</custom-element>`, `<p>Hello</p><custom-element>World</custom-element>`, `<p data-original-title="World">Hello</p><custom-element data-original-title="Hello">World</custom-element>`, (&htmlcleaner.Config{}).Elem("p", "custom-element"), func(c *htmlcleaner.Config) { c.GlobalAttr("data-original-title") })
	run("WrapText", `a<blockquote>b</blockquote>c<custom-element>d</custom-element>e`, `<p>a</p><blockquote>b</blockquote><p>c</p><custom-element>d</custom-element><p>e</p>`, `<p>a</p><blockquote><p>b</p></blockquote><p>c</p><custom-element><p>d</p></custom-element><p>e</p>`, (&htmlcleaner.Config{WrapText: true}).Elem("p", "blockquote", "custom-element"), func(c *htmlcleaner.Config) { c.WrapTextInside("blockquote", "custom-element") })
}
