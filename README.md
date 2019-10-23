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

## Limitations

1. Only the code necessary for writing a server for collaborative editing will
   be ported, not things like translating a document to/from a DOM
   representation which are only useful on the clients.

2. In go, the `map`s don't preserve the order of the key. `OrderedMap` in the
   JS `schema` needs to be serialized in JSON to an array of tuples `[key,
   value]` to keep the order:

```json
[
  ["em", { "inclusive": true, "group": "fontStyle" }],
  ["strong", { "inclusive": true, "group": "fontStyle" }]
]
```

## License

The port in Go of ProseMirror has been developed by Cozy Cloud and is
distributed under the AGPL v3 license.
