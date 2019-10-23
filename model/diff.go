package model

// findDiffStart returns the first position where the two fragments have not
// the same content.
func findDiffStart(a, b *Fragment, pos int) *int {
	for i := 0; ; i++ {
		if i == a.ChildCount() || i == b.ChildCount() {
			if a.ChildCount() == b.ChildCount() {
				return nil
			}
			return &pos
		}

		childA := a.Child(i)
		childB := b.Child(i)
		if childA == childB {
			pos += childA.NodeSize()
			continue
		}

		if !childA.SameMarkup(childB) {
			return &pos
		}

		if childA.IsText() && childA.Text() != childB.Text() {
			for j := 0; childA.Text()[j] == childB.Text()[j]; j++ {
				pos++
			}
			return &pos
		}
		if childA.Content.Size > 0 || childB.Content.Size > 0 {
			inner := findDiffStart(childA.Content, childB.Content, pos+1)
			if inner != nil {
				return inner
			}
		}
		pos += childA.NodeSize()
	}
}

// diffEnd is the result of findDiffEnd with the positions in both a and b
// fragments.
type diffEnd struct {
	A int
	B int
}

// findDiffEnd returns the last position where the two fragments have not
// the same content.
func findDiffEnd(a, b *Fragment, posA, posB int) *diffEnd {
	ia := a.ChildCount()
	ib := b.ChildCount()
	for {
		if ia == 0 || ib == 0 {
			if ia == ib {
				return nil
			}
			return &diffEnd{A: posA, B: posB}
		}

		ia--
		ib--
		childA := a.Child(ia)
		childB := b.Child(ib)
		size := childA.NodeSize()
		if childA == childB {
			posA -= size
			posB -= size
			continue
		}

		if !childA.SameMarkup(childB) {
			return &diffEnd{A: posA, B: posB}
		}

		if childA.IsText() && childA.Text() != childB.Text() {
			same := 0
			la := len(childA.Text())
			lb := len(childB.Text())
			minSize := len(childA.Text())
			if lb < minSize {
				minSize = lb
			}
			for same < minSize && childA.Text()[la-same-1] == childB.Text()[lb-same-1] {
				same++
				posA--
				posB--
			}
			return &diffEnd{A: posA, B: posB}
		}
		if childA.Content.Size > 0 || childB.Content.Size > 0 {
			inner := findDiffEnd(childA.Content, childB.Content, posA-1, posB-1)
			if inner != nil {
				return inner
			}
		}
		posA -= size
		posB -= size
	}
}
