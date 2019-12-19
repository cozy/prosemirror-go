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

		childA, err := a.Child(i)
		if err != nil {
			panic(err)
		}
		childB, err := b.Child(i)
		if err != nil {
			panic(err)
		}
		if childA == childB {
			pos += childA.NodeSize()
			continue
		}

		if !childA.SameMarkup(childB) {
			return &pos
		}

		if childA.IsText() && (*childA.Text) != (*childB.Text) {
			for j := 0; j < childA.NodeSize() && j < childB.NodeSize(); j++ {
				if childA.UnitCodeAt(j) != childB.UnitCodeAt(j) {
					break
				}
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

// DiffEnd is the result of findDiffEnd with the positions in both a and b
// fragments.
type DiffEnd struct {
	A int
	B int
}

// findDiffEnd returns the last position where the two fragments have not
// the same content.
func findDiffEnd(a, b *Fragment, posA, posB int) *DiffEnd {
	ia := a.ChildCount()
	ib := b.ChildCount()
	for {
		if ia == 0 || ib == 0 {
			if ia == ib {
				return nil
			}
			return &DiffEnd{A: posA, B: posB}
		}

		ia--
		ib--
		childA, err := a.Child(ia)
		if err != nil {
			panic(err)
		}
		childB, err := b.Child(ib)
		if err != nil {
			panic(err)
		}
		size := childA.NodeSize()
		if childA == childB {
			posA -= size
			posB -= size
			continue
		}

		if !childA.SameMarkup(childB) {
			return &DiffEnd{A: posA, B: posB}
		}

		if childA.IsText() && *childA.Text != *childB.Text {
			same := 0
			la := childA.NodeSize()
			lb := childB.NodeSize()
			minSize := la
			if lb < minSize {
				minSize = lb
			}
			for same < minSize && la-same > 0 && lb-same > 0 {
				if childA.UnitCodeAt(la-same-1) != childB.UnitCodeAt(lb-same-1) {
					break
				}
				same++
				posA--
				posB--
			}
			return &DiffEnd{A: posA, B: posB}
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
