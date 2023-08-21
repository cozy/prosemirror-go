package markdown

import (
	"fmt"
	"regexp"
	"strings"
	"unicode"

	"github.com/cozy/prosemirror-go/model"
)

// NodeSerializerFunc is the function to serialize a node.
type NodeSerializerFunc func(state *SerializerState, node, parent *model.Node, index int)

// MarkSerializerSpec is the serializer info for a mark.
type MarkSerializerSpec struct {
	Open                     interface{} // Can be a string or a func
	Close                    interface{} // Can be a string or a func
	Mixable                  bool
	ExpelEnclosingWhitespace bool
	NoEscape                 bool
}

// Serializer is a specification for serializing a ProseMirror document as
// Markdown/CommonMark text.
type Serializer struct {
	Nodes map[string]NodeSerializerFunc
	Marks map[string]MarkSerializerSpec
}

// NewSerializer constructs a serializer with the given configuration. The
// `nodes` object should map node names in a given schema to function that take
// a serializer state and such a node, and serialize the node.
//
// The `marks` object should hold objects with `open` and `close` properties,
// which hold the strings that should appear before and after a piece of text
// marked that way, either directly or as a function that takes a serializer
// state and a mark, and returns a string. `open` and `close` can also be
// functions, which will be called as
//
//	(state: MarkdownSerializerState, mark: Mark, parent: Fragment, index:
//	number) → string
//
// Where `parent` and `index` allow you to inspect the mark's context to see
// which nodes it applies to.
//
// Mark information objects can also have a `mixable` property which, when
// `true`, indicates that the order in which the mark's opening and closing
// syntax appears relative to other mixable marks can be varied. (For example,
// you can say `**a *b***` and `*a **b***`, but not “ `a *b*` “.)
//
// To disable character escaping in a mark, you can give it an `escape`
// property of `false`. Such a mark has to have the highest precedence (must
// always be the innermost mark).
//
// The `expelEnclosingWhitespace` mark property causes the serializer to move
// enclosing whitespace from inside the marks to outside the marks. This is
// necessary for emphasis marks as CommonMark does not permit enclosing
// whitespace inside emphasis marks, see:
// http://spec.commonmark.org/0.26/#example-330
func NewSerializer(nodes map[string]NodeSerializerFunc, marks map[string]MarkSerializerSpec) *Serializer {
	return &Serializer{
		Nodes: nodes,
		Marks: marks,
	}
}

// Serialize the content of the given node to
// [CommonMark](http://commonmark.org/).
func (s *Serializer) Serialize(content *model.Node, options ...map[string]interface{}) string {
	var opts map[string]interface{}
	if len(options) > 0 {
		opts = options[0]
	}
	state := NewSerializerState(s.Nodes, s.Marks, opts)
	state.RenderContent(content)
	return state.Out
}

func getAttrInt(attrs map[string]interface{}, name string, defaultValue int) int {
	value := defaultValue
	switch v := attrs[name].(type) {
	case int:
		value = v
	case float64:
		value = int(v)
	case int64:
		value = int(v)
	}
	return value
}

var backticksRegexp = regexp.MustCompile("`{3,}")

// DefaultSerializer is a serializer for the [basic schema](#schema).
var DefaultSerializer = NewSerializer(map[string]NodeSerializerFunc{
	"blockquote": func(state *SerializerState, node, _parent *model.Node, _index int) {
		state.WrapBlock("> ", nil, node, func() { state.RenderContent(node) })
	},
	"code_block": func(state *SerializerState, node, _parent *model.Node, _index int) {
		fence := "```"
		content := node.TextContent()
		matches := backticksRegexp.FindAllString(content, -1)
		for _, backticks := range matches {
			if len(backticks) >= len(fence) {
				fence = backticks + "`"
			}
		}

		params, _ := node.Attrs["params"].(string)
		state.Write(fence + params + "\n")
		state.Text(content, false)
		state.EnsureNewLine()
		state.Write(fence)
		state.CloseBlock(node)
	},
	"heading": func(state *SerializerState, node, _parent *model.Node, _index int) {
		level := getAttrInt(node.Attrs, "level", 1)
		state.Write(strings.Repeat("#", level) + " ")
		state.RenderInline(node)
		state.CloseBlock(node)
	},
	"horizontal_rule": func(state *SerializerState, node, _parent *model.Node, _index int) {
		markup := "---"
		if m, ok := node.Attrs["markup"].(string); ok {
			markup = m
		}
		state.Write(markup)
		state.CloseBlock(node)
	},
	"bullet_list": func(state *SerializerState, node, _parent *model.Node, _index int) {
		bullet := "*"
		if b, ok := node.Attrs["bullet"].(string); ok {
			bullet = b
		}
		state.RenderList(node, "  ", func(_ int) string { return bullet + " " })
	},
	"ordered_list": func(state *SerializerState, node, _parent *model.Node, _index int) {
		start := getAttrInt(node.Attrs, "order", 1)
		maxW := len(fmt.Sprintf("%d", start+node.ChildCount()-1))
		space := strings.Repeat(" ", maxW+2)
		state.RenderList(node, space, func(i int) string {
			nStr := fmt.Sprintf("%d", start+i)
			return strings.Repeat(" ", maxW-len(nStr)) + nStr + ". "
		})
	},
	"list_item": func(state *SerializerState, node, _parent *model.Node, _index int) {
		state.RenderContent(node)
	},
	"paragraph": func(state *SerializerState, node, _parent *model.Node, _index int) {
		state.RenderInline(node)
		state.CloseBlock(node)
	},
	"image": func(state *SerializerState, node, _parent *model.Node, _index int) {
		alt, _ := node.Attrs["alt"].(string)
		src, _ := node.Attrs["src"].(string)
		src = strings.ReplaceAll(src, "(", "\\(")
		src = strings.ReplaceAll(src, ")", "\\)")
		title := ""
		if t, ok := node.Attrs["title"].(string); ok {
			title = ` "` + strings.ReplaceAll(t, `"`, `\"`) + `"`
		}
		state.Write(fmt.Sprintf("![%s](%s)%s", state.Esc(alt), src, title))
	},
	"hard_break": func(state *SerializerState, node, parent *model.Node, index int) {
		for i := index; i < parent.ChildCount(); i++ {
			if child, err := parent.Child(i); err == nil {
				if child.Type != node.Type {
					state.Write("\\\n")
					return
				}
			}
		}
	},
	"text": func(state *SerializerState, node, _parent *model.Node, _index int) {
		state.Text(*node.Text, !state.InAutoLink)
	},
}, map[string]MarkSerializerSpec{
	"em":     {Open: "*", Close: "*", Mixable: true, ExpelEnclosingWhitespace: true},
	"strong": {Open: "**", Close: "**", Mixable: true, ExpelEnclosingWhitespace: true},
	"link": {
		Open: func(state *SerializerState, mark *model.Mark, parent *model.Node, index int) string {
			state.InAutoLink = isPlainURL(mark, parent, index)
			if state.InAutoLink {
				return "<"
			}
			return "["
		},
		Close: func(state *SerializerState, mark *model.Mark, parent *model.Node, index int) string {
			if state.InAutoLink {
				state.InAutoLink = false
				return ">"
			}
			href, _ := mark.Attrs["href"].(string)
			href = strings.ReplaceAll(href, "(", "\\(")
			href = strings.ReplaceAll(href, ")", "\\)")
			href = strings.ReplaceAll(href, `"`, `\"`)
			title, _ := mark.Attrs["title"].(string)
			if title != "" {
				title = ` "` + strings.ReplaceAll(title, `"`, `\"`) + `"`
			}
			return fmt.Sprintf("](%s%s)", href, title)
		},
		Mixable: true,
	},
	"code": {
		Open: func(_state *SerializerState, _mark *model.Mark, parent *model.Node, index int) string {
			child, err := parent.Child(index)
			if err != nil {
				return "`"
			}
			return backticksFor(child, -1)
		},
		Close: func(_state *SerializerState, _mark *model.Mark, parent *model.Node, index int) string {
			child, err := parent.Child(index - 1)
			if err != nil {
				return "`"
			}
			return backticksFor(child, 1)
		},
		NoEscape: true,
	},
})

func backticksFor(node *model.Node, side int) string {
	length := 0
	if node.IsText() {
		ticks := strings.FieldsFunc(*node.Text, func(r rune) bool { return r != '`' })
		for _, t := range ticks {
			if l := len(t); l > length {
				length = l
			}
		}
	}
	result := "`"
	if length > 0 && side > 0 {
		result = " `"
	}
	for i := 0; i < length; i++ {
		result += "`"
	}
	if length > 0 && side < 0 {
		result += " "
	}
	return result
}

func isPlainURL(link *model.Mark, parent *model.Node, index int) bool {
	if _, ok := link.Attrs["title"].(string); ok {
		return false
	}
	href, _ := link.Attrs["href"].(string)
	if !strings.Contains(href, ":") {
		return false
	}
	content, err := parent.Child(index)
	if err != nil {
		return true
	}
	if !content.IsText() || *content.Text != href || content.Marks[len(content.Marks)-1] != link {
		return false
	}
	if index == parent.ChildCount()-1 {
		return true
	}
	next, err := parent.Child(index + 1)
	if err != nil {
		return true
	}
	return !link.IsInSet(next.Marks)
}

// SerializerState is an object used to track state and expose methods related
// to markdown serialization. Instances are passed to node and mark
// serialization methods (see `toMarkdown`).
type SerializerState struct {
	Nodes        map[string]NodeSerializerFunc
	Marks        map[string]MarkSerializerSpec
	Delim        string
	Out          string
	Closed       *model.Node
	InAutoLink   bool
	AtBlockStart bool
	InTightList  bool
	tightLists   bool
}

// NewSerializerState is the constructor for NewSerializerState.
//
// Options are the options passed to the serializer.
//
//	tightLists:: ?bool
//	Whether to render lists in a tight style. This can be overridden
//	on a node level by specifying a tight attribute on the node.
//	Defaults to false.
func NewSerializerState(
	nodes map[string]NodeSerializerFunc,
	marks map[string]MarkSerializerSpec,
	options map[string]interface{},
) *SerializerState {
	tight := false
	if t, ok := options["tightLists"].(bool); ok {
		tight = t
	}
	return &SerializerState{
		Nodes:       nodes,
		Marks:       marks,
		Delim:       "",
		Out:         "",
		Closed:      nil,
		InTightList: false,
		tightLists:  tight,
	}
}

func (s *SerializerState) flushClose(size ...int) {
	if s.Closed == nil {
		return
	}
	s.EnsureNewLine()
	siz := 2
	if len(size) > 0 {
		siz = size[0]
	}
	if siz > 1 {
		delimMin := strings.TrimRightFunc(s.Delim, unicode.IsSpace)
		for i := 1; i < siz; i++ {
			s.Out += delimMin + "\n"
		}
	}
	s.Closed = nil
}

// WrapBlock renders a block, prefixing each line with `delim`, and the first
// line in `firstDelim`. `node` should be the node that is closed at the end of
// the block, and `f` is a function that renders the content of the block.
func (s *SerializerState) WrapBlock(delim string, firstDelim *string, node *model.Node, f func()) {
	old := s.Delim
	d := delim
	if firstDelim != nil {
		d = *firstDelim
	}
	s.Write(d)
	s.Delim += delim
	f()
	s.Delim = old
	s.CloseBlock(node)
}

func (s *SerializerState) atBlank() bool {
	if len(s.Out) == 0 {
		return true
	}
	return s.Out[len(s.Out)-1] == '\n'
}

// EnsureNewLine ensures the current content ends with a newline.
func (s *SerializerState) EnsureNewLine() {
	if !s.atBlank() {
		s.Out += "\n"
	}
}

// Write prepares the state for writing output (closing closed paragraphs,
// adding delimiters, and so on), and then optionally add content
// (unescaped) to the output.
func (s *SerializerState) Write(content ...string) {
	s.flushClose()
	if s.Delim != "" && s.atBlank() {
		s.Out += s.Delim
	}
	if len(content) > 0 {
		s.Out += content[0]
	}
}

// CloseBlock closes the block for the given node.
func (s *SerializerState) CloseBlock(node *model.Node) {
	s.Closed = node
}

var textRegexp1 = regexp.MustCompile(`(^|[^\\])\!$`)

// Text adds the given text to the document. When escape is not `false`, it
// will be escaped.
func (s *SerializerState) Text(text string, escape ...bool) {
	lines := strings.Split(text, "\n")
	esc := true
	if len(escape) > 0 {
		esc = escape[0]
	}
	for i, line := range lines {
		s.Write()
		// Escape exclamation marks in front of links
		if !esc && line[0] == '[' && textRegexp1.MatchString(s.Out) {
			s.Out = s.Out[:len(s.Out)-1] + "\\!"
		}
		if esc {
			s.Out += s.Esc(line, s.AtBlockStart)
		} else {
			s.Out += line
		}
		if i != len(lines)-1 {
			s.Out += "\n"
		}
	}
}

// Render the given node as a block.
func (s *SerializerState) Render(node, parent *model.Node, index int) {
	if fn, ok := s.Nodes[node.Type.Name]; ok {
		fn(s, node, parent, index)
	}
}

// RenderContent renders the contents of `parent` as block nodes.
func (s *SerializerState) RenderContent(parent *model.Node) {
	parent.ForEach(func(node *model.Node, _ int, i int) {
		s.Render(node, parent, i)
	})
}

var inlineRegexp = regexp.MustCompile(`^(\s*)(.*?)(\s*)$`)

// RenderInline renders the contents of `parent` as inline content.
func (s *SerializerState) RenderInline(parent *model.Node) {
	s.AtBlockStart = true
	var active []*model.Mark
	var trailing string

	progress := func(node *model.Node, _offset, index int) {
		var marks []*model.Mark
		if node != nil {
			marks = node.Marks
		}

		// Remove marks from `hard_break` that are the last node inside
		// that mark to prevent parser edge cases with new lines just
		// before closing marks.
		// (FIXME it'd be nice if we had a schema-agnostic way to
		// identify nodes that serialize as hard breaks)
		if node != nil && node.Type.Name == "hard_break" {
			var filtered []*model.Mark
			for _, m := range marks {
				if index+1 == parent.ChildCount() {
					continue
				}
				next, err := parent.Child(index + 1)
				if err != nil {
					continue
				}
				if !m.IsInSet(next.Marks) {
					continue
				}
				if !next.IsText() || strings.TrimSpace(*next.Text) != "" {
					filtered = append(filtered, m)
				}
			}
			marks = filtered
		}

		leading := trailing
		trailing = ""
		// If whitespace has to be expelled from the node, adjust
		// leading and trailing accordingly.
		if node != nil && node.IsText() {
			expel := false
			for _, mark := range marks {
				if info, ok := s.Marks[mark.Type.Name]; ok && info.ExpelEnclosingWhitespace {
					if mark.IsInSet(active) {
						continue
					}
					if index >= parent.ChildCount()-1 {
						expel = true
						break
					}
					other, err := parent.Child(index + 1)
					if err == nil && !mark.IsInSet(other.Marks) {
						expel = true
						break
					}
				}
			}
			if expel {
				parts := inlineRegexp.FindStringSubmatch(*node.Text)
				if len(parts) == 4 {
					leading += parts[1]
					trailing = parts[3]
					if parts[1] != "" || parts[3] != "" {
						if inner := parts[2]; inner != "" {
							node = node.WithText(inner)
						} else {
							node = nil
						}
						if node == nil {
							marks = active
						}
					}
				}
			}
		}

		var inner *model.Mark
		if len(marks) > 0 {
			inner = marks[len(marks)-1]
		}
		noEsc := false
		if inner != nil {
			noEsc = s.Marks[inner.Type.Name].NoEscape
		}
		length := len(marks)
		if noEsc {
			length--
		}

		// Try to reorder 'mixable' marks, such as em and strong, which
		// in Markdown may be opened and closed in different order, so
		// that order of the marks for the token matches the order in
		// active.
		for i, mark := range marks {
			if !s.Marks[mark.Type.Name].Mixable {
				break
			}
			for j, other := range active {
				if !s.Marks[other.Type.Name].Mixable {
					break
				}
				if mark.Eq(other) {
					mixed := make([]*model.Mark, 0, len(marks))
					if i > j {
						mixed = append(mixed, marks[:j]...)
						mixed = append(mixed, mark)
						mixed = append(mixed, marks[j:i]...)
						mixed = append(mixed, marks[i+1:]...)
					} else {
						mixed = append(mixed, marks[:i]...)
						if i != j {
							mixed = append(mixed, marks[i+1:j]...)
						}
						mixed = append(mixed, mark)
						mixed = append(mixed, marks[j:]...)
					}
					marks = mixed
					break
				}
			}
		}

		// Find the prefix of the mark set that didn't change
		min := len(marks)
		if l := len(active); l < min {
			min = l
		}
		keep := 0
		for keep < min && marks[keep].Eq(active[keep]) {
			keep++
		}

		// Close the marks that need to be closed
		for keep < len(active) {
			s.Text(s.MarkString(active[len(active)-1], false, parent, index), false)
			active = active[:len(active)-1]
		}

		// Output any previously expelled trailing whitespace outside the marks
		if leading != "" {
			s.Text(leading)
		}

		// Open the marks that need to be opened
		if node != nil {
			for len(active) < length {
				add := marks[len(active)]
				active = append(active, add)
				s.Text(s.MarkString(add, true, parent, index), false)
			}

			// Render the node. Special case code marks, since their content
			// may not be escaped.
			if noEsc && node.IsText() {
				s.Text(s.MarkString(inner, true, parent, index)+*node.Text+
					s.MarkString(inner, false, parent, index+1), false)
			} else {
				s.Render(node, parent, index)
			}
		}
	}

	parent.ForEach(progress)
	progress(nil, 0, parent.ChildCount())
	s.AtBlockStart = false
}

// RenderList renders a node's content as a list. `delim` should be the extra
// indentation added to all lines except the first in an item, `firstDelim` is
// a function going from an item index to a delimiter for the first line of the
// item.
func (s *SerializerState) RenderList(node *model.Node, delim string, firstDelim func(i int) string) {
	if s.Closed != nil && s.Closed.Type == node.Type {
		s.flushClose(3)
	} else if s.InTightList {
		s.flushClose(1)
	}

	isTight := s.tightLists
	if t, ok := node.Attrs["tight"].(bool); ok {
		isTight = t
	}
	prevTight := s.InTightList
	s.InTightList = isTight
	node.ForEach(func(child *model.Node, _, i int) {
		if i > 0 && isTight {
			s.flushClose(1)
		}
		first := firstDelim(i)
		s.WrapBlock(delim, &first, node, func() { s.Render(child, node, i) })
	})
	s.InTightList = prevTight
}

var (
	escRegexp1 = regexp.MustCompile("([`*\\\\~\\[\\]])")
	escRegexp2 = regexp.MustCompile(`(\b_)|(_\b)`)
	escRegexp3 = regexp.MustCompile(`^([#\-*+>])`)
	escRegexp4 = regexp.MustCompile(`(\s*\d+)\.`)
)

// Esc escapes the given string so that it can safely appear in Markdown
// content. If `startOfLine` is true, also escape characters that have special
// meaning only at the start of the line.
func (s *SerializerState) Esc(str string, startOfLine ...bool) string {
	start := false
	if len(startOfLine) > 0 {
		start = startOfLine[0]
	}
	str = escRegexp1.ReplaceAllString(str, "\\$1")
	str = escRegexp2.ReplaceAllString(str, "\\_")
	if start {
		str = escRegexp3.ReplaceAllString(str, "\\$1")
		str = escRegexp4.ReplaceAllString(str, "$1\\.")
	}
	return str
}

// Quote wraps the string as a quote.
func (s *SerializerState) Quote(str string) string {
	wrap := `()`
	if !strings.Contains(str, `"`) {
		wrap = `""`
	} else if !strings.Contains(str, "'") {
		wrap = "''"
	}
	return wrap[:1] + str + wrap[1:]
}

// MarkString gets the markdown string for a given opening or closing mark.
func (s *SerializerState) MarkString(mark *model.Mark, open bool, parent *model.Node, index int) string {
	info := s.Marks[mark.Type.Name]
	value := info.Open
	if !open {
		value = info.Close
	}
	switch value := value.(type) {
	case string:
		return value
	case func(state *SerializerState, mark *model.Mark, parent *model.Node, index int) string:
		return value(s, mark, parent, index)
	}
	return ""
}
