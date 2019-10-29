package model

// ContentMatch represents a match state of a node type's content expression,
// and can be used to find out whether further content matches here, and
// whether a given position is a valid end of the node.
type ContentMatch struct {
	// True when this match state represents a valid end of the node.
	ValidEnd  bool
	next      []interface{} // even indexes are *NodeType, odd are *ContentMatch
	wrapCache []interface{}
}

// NewContentMatch is the constructor for ContentMatch.
func NewContentMatch(validEnd bool) *ContentMatch {
	return &ContentMatch{
		ValidEnd:  validEnd,
		next:      nil,
		wrapCache: nil,
	}
}

func parseContentMatch(str string, nodeTypes map[string]*NodeType) (*ContentMatch, error) {
	return EmptyContentMatch, nil // TODO
}

// MatchType matches a node type, returning a match after that node if
// successful.
func (cm *ContentMatch) MatchType(typ *NodeType) *ContentMatch {
	for i := 0; i < len(cm.next); i += 2 {
		if cm.next[i] == typ {
			return cm.next[i+1].(*ContentMatch)
		}
	}
	return nil
}

// MatchFragment tries to match a fragment. Returns the resulting match when
// successful.
//
// :: (Fragment, ?number, ?number) → ?ContentMatch
func (cm *ContentMatch) MatchFragment(frag *Fragment, args ...int) *ContentMatch {
	cur := cm
	start := 0
	if len(args) > 0 {
		start = args[0]
	}
	end := 0
	if len(args) > 1 {
		end = args[1]
	} else {
		end = frag.ChildCount()
	}
	for i := start; cur != nil && i < end; i++ {
		child, err := frag.Child(i)
		if err != nil {
			return nil
		}
		cur = cur.MatchType(child.Type)
	}
	return cur
}

func (cm *ContentMatch) inlineContent() bool {
	if len(cm.next) == 0 {
		return false
	}
	return cm.next[0].(*NodeType).IsInline()
}

func (cm *ContentMatch) compatible(other *ContentMatch) bool {
	for i := 0; i < len(cm.next); i += 2 {
		for j := 0; j < len(other.next); j += 2 {
			if cm.next[i] == other.next[j] {
				return true
			}
		}
	}
	return false
}

// EmptyContentMatch is an empty ContentMatch.
var EmptyContentMatch = NewContentMatch(true)
