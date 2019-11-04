ProseMirror in Go
=================

[![GoDoc](https://godoc.org/github.com/cozy/prosemirror-go?status.svg)](https://godoc.org/github.com/cozy/prosemirror-go)
[![Build Status](https://github.com/cozy/prosemirror-go/workflows/CI/badge.svg)](https://github.com/cozy/prosemirror-go/actions)
[![Go Report Card](https://goreportcard.com/badge/github.com/cozy/prosemirror-go)](https://goreportcard.com/report/github.com/cozy/prosemirror-go)

## Introduction

[ProseMirror](http://prosemirror.net/) is a well-behaved rich semantic content
editor based on contentEditable, with support for collaborative editing and
custom document schemas. This repository contains a port in Go of
[prosemirror-model](https://github.com/ProseMirror/prosemirror-model) and
[prosemirror-transform](https://github.com/ProseMirror/prosemirror-transform/)
in order to have the server part of the collaborative editing in Go.

## Notes

1. Only the code necessary for writing a server for collaborative editing will
   be ported, not things like translating a document to/from a DOM
   representation which are only useful on the clients.

2. In go, the `map`s don't preserve the order of the key (and even the JSON
   spec doesn't say that there is an order for object fields). `OrderedMap` in
   the JS `schema` needs to be serialized in JSON to an array of tuples
   `[key, value]` to keep the order:

```json
[
  ["em", { "inclusive": true, "group": "fontStyle" }],
  ["strong", { "inclusive": true, "group": "fontStyle" }]
]
```

3. Go doesn't support optional arguments like JS does. We are emulating this
   with variadic arguments:

```js
  // :: (number, number, (node: Node, pos: number, parent: Node, index: number) → ?bool, ?number)
  nodesBetween(from, to, f, startPos = 0) {
    this.content.nodesBetween(from, to, f, startPos, this)
  }
```

```go
func (n *Node) NodesBetween(from, to int, fn NBCallback, startPos ...int) {
	s := 0
	if len(startPos) > 0 {
		s = startPos[0]
	}
	n.Content.NodesBetween(from, to, fn, s, n)
}
```

4. Exceptions in JS can be manager in Go by returning an error, or with a
   panic. We have tried to panic only for logic bugs and out of bounds access,
   and returning an error everywhere else.

## License

The port in Go of ProseMirror has been developed by Cozy Cloud and is
distributed under the AGPL v3 license.
