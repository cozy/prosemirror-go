package model_test

import (
	. "github.com/cozy/prosemirror-go/model"
	"github.com/cozy/prosemirror-go/test/builder"
)

var (
	schema     = builder.Schema
	doc        = builder.Doc
	blockquote = builder.Blockquote
	h1         = builder.H1
	h2         = builder.H2
	p          = builder.P
	em         = builder.Em
	strong     = builder.Strong
	ul         = builder.Ul
	li         = builder.Li
	img        = builder.Img
	br         = builder.Br
	code       = builder.Code

	strong2 = schema.Mark("strong")
	em2     = schema.Mark("em")
	code2   = schema.Mark("code")
	link    = func(href string, title ...string) *Mark {
		attrs := map[string]interface{}{"href": href}
		if len(title) > 0 {
			attrs["title"] = title[0]
		}
		return schema.Mark("link", attrs)
	}

	empty      = ""
	underscore = "_"
	falsy      = false
	emGroup    = "em-group"
	idAttrs    = map[string]*AttributeSpec{
		"id": {},
	}
)
