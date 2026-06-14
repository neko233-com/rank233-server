package ranker

const (
	RED   = true
	BLACK = false
)

type rbNode[Key comparable] struct {
	score Score
	value Key
	color bool
	left  *rbNode[Key]
	right *rbNode[Key]
	p     *rbNode[Key]
}

type rbTree[Key comparable] struct {
	root *rbNode[Key]
	size int
}

func (t *rbTree[Key]) rotateLeft(x *rbNode[Key]) {
	y := x.right
	x.right = y.left
	if y.left != nil {
		y.left.p = x
	}
	y.p = x.p
	if x.p == nil {
		t.root = y
	} else if x == x.p.left {
		x.p.left = y
	} else {
		x.p.right = y
	}
	y.left = x
	x.p = y
}

func (t *rbTree[Key]) rotateRight(x *rbNode[Key]) {
	y := x.left
	x.left = y.right
	if y.right != nil {
		y.right.p = x
	}
	y.p = x.p
	if x.p == nil {
		t.root = y
	} else if x == x.p.right {
		x.p.right = y
	} else {
		x.p.left = y
	}
	y.right = x
	x.p = y
}

func (t *rbTree[Key]) insert(z *rbNode[Key]) {
	var y *rbNode[Key]
	x := t.root
	for x != nil {
		y = x
		cmp := z.score.Compare(x.score)
		if cmp <= 0 {
			x = x.left
		} else {
			x = x.right
		}
	}
	z.p = y
	if y == nil {
		t.root = z
	} else if z.score.Compare(y.score) <= 0 {
		y.left = z
	} else {
		y.right = z
	}
	z.color = RED
	t.insertFixup(z)
	t.size++
}

func (t *rbTree[Key]) insertFixup(z *rbNode[Key]) {
	for z.p != nil && z.p.color == RED {
		if z.p == z.p.p.left {
			y := z.p.p.right
			if y != nil && y.color == RED {
				z.p.color = BLACK
				y.color = BLACK
				z.p.p.color = RED
				z = z.p.p
			} else {
				if z == z.p.right {
					z = z.p
					t.rotateLeft(z)
				}
				z.p.color = BLACK
				z.p.p.color = RED
				t.rotateRight(z.p.p)
			}
		} else {
			y := z.p.p.left
			if y != nil && y.color == RED {
				z.p.color = BLACK
				y.color = BLACK
				z.p.p.color = RED
				z = z.p.p
			} else {
				if z == z.p.left {
					z = z.p
					t.rotateRight(z)
				}
				z.p.color = BLACK
				z.p.p.color = RED
				t.rotateLeft(z.p.p)
			}
		}
	}
	t.root.color = BLACK
}

func (t *rbTree[Key]) transplant(u, v *rbNode[Key]) {
	if u.p == nil {
		t.root = v
	} else if u == u.p.left {
		u.p.left = v
	} else {
		u.p.right = v
	}
	if v != nil {
		v.p = u.p
	}
}

func (t *rbTree[Key]) minimum(x *rbNode[Key]) *rbNode[Key] {
	for x.left != nil {
		x = x.left
	}
	return x
}

func (t *rbTree[Key]) maximum(x *rbNode[Key]) *rbNode[Key] {
	for x.right != nil {
		x = x.right
	}
	return x
}

func (t *rbTree[Key]) delete(z *rbNode[Key]) {
	if z.left != nil && z.right != nil {
		y := t.minimum(z.right)
		z.score = y.score
		z.value = y.value
		t.deleteNode(y)
	} else {
		t.deleteNode(z)
	}
	t.size--
}

func (t *rbTree[Key]) deleteNode(z *rbNode[Key]) {
	var child *rbNode[Key]
	if z.left != nil {
		child = z.left
	} else {
		child = z.right
	}

	if child != nil {
		child.p = z.p
	}

	if z.p == nil {
		t.root = child
	} else if z == z.p.left {
		z.p.left = child
	} else {
		z.p.right = child
	}

	if z.color == BLACK && child != nil {
		t.deleteFixup(child)
	} else if z.color == BLACK && child == nil {
		if z.p != nil {
			t.deleteFixupFromParent(z.p, z == z.p.left)
		}
	}
}

func (t *rbTree[Key]) deleteFixupFromParent(xP *rbNode[Key], isLeftChild bool) {
	for xP != t.root && xP.color == BLACK {
		if isLeftChild {
			w := xP.right
			if w != nil && w.color == RED {
				w.color = BLACK
				xP.color = RED
				t.rotateLeft(xP)
				w = xP.right
			}
			if w == nil {
				xP.color = RED
				break
			}
			lb := w.left == nil || w.left.color == BLACK
			rb := w.right == nil || w.right.color == BLACK
			if lb && rb {
				w.color = RED
				if xP.p == nil {
					break
				}
				isLeftChild = xP == xP.p.left
				xP = xP.p
			} else {
				if w.right == nil || w.right.color == BLACK {
					if w.left != nil {
						w.left.color = BLACK
					}
					w.color = RED
					t.rotateRight(w)
					w = xP.right
				}
				if w != nil {
					w.color = xP.color
					if w.right != nil {
						w.right.color = BLACK
					}
				}
				xP.color = BLACK
				t.rotateLeft(xP)
				break
			}
		} else {
			w := xP.left
			if w != nil && w.color == RED {
				w.color = BLACK
				xP.color = RED
				t.rotateRight(xP)
				w = xP.left
			}
			if w == nil {
				xP.color = RED
				break
			}
			lb := w.left == nil || w.left.color == BLACK
			rb := w.right == nil || w.right.color == BLACK
			if lb && rb {
				w.color = RED
				if xP.p == nil {
					break
				}
				isLeftChild = xP == xP.p.left
				xP = xP.p
			} else {
				if w.left == nil || w.left.color == BLACK {
					if w.right != nil {
						w.right.color = BLACK
					}
					w.color = RED
					t.rotateLeft(w)
					w = xP.left
				}
				if w != nil {
					w.color = xP.color
					if w.left != nil {
						w.left.color = BLACK
					}
				}
				xP.color = BLACK
				t.rotateRight(xP)
				break
			}
		}
	}
	if xP != nil {
		xP.color = BLACK
	}
}

func (t *rbTree[Key]) deleteFixup(x *rbNode[Key]) {
	for x != t.root && x.color == BLACK {
		if x == x.p.left {
			w := x.p.right
			if w != nil && w.color == RED {
				w.color = BLACK
				x.p.color = RED
				t.rotateLeft(x.p)
				w = x.p.right
			}
			if w == nil {
				break
			}
			lb := w.left == nil || w.left.color == BLACK
			rb := w.right == nil || w.right.color == BLACK
			if lb && rb {
				w.color = RED
				x = x.p
			} else {
				if w.right == nil || w.right.color == BLACK {
					if w.left != nil {
						w.left.color = BLACK
					}
					w.color = RED
					t.rotateRight(w)
					w = x.p.right
				}
				if w != nil {
					w.color = x.p.color
					if w.right != nil {
						w.right.color = BLACK
					}
				}
				x.p.color = BLACK
				t.rotateLeft(x.p)
				x = t.root
			}
		} else {
			w := x.p.left
			if w != nil && w.color == RED {
				w.color = BLACK
				x.p.color = RED
				t.rotateRight(x.p)
				w = x.p.left
			}
			if w == nil {
				break
			}
			lb := w.left == nil || w.left.color == BLACK
			rb := w.right == nil || w.right.color == BLACK
			if lb && rb {
				w.color = RED
				x = x.p
			} else {
				if w.left == nil || w.left.color == BLACK {
					if w.right != nil {
						w.right.color = BLACK
					}
					w.color = RED
					t.rotateLeft(w)
					w = x.p.left
				}
				if w != nil {
					w.color = x.p.color
					if w.left != nil {
						w.left.color = BLACK
					}
				}
				x.p.color = BLACK
				t.rotateRight(x.p)
				x = t.root
			}
		}
	}
	x.color = BLACK
}

func (t *rbTree[Key]) inorder(fn func(*rbNode[Key]) bool) {
	t.inorderNode(t.root, fn)
}

func (t *rbTree[Key]) inorderNode(n *rbNode[Key], fn func(*rbNode[Key]) bool) bool {
	if n == nil {
		return true
	}
	if !t.inorderNode(n.left, fn) {
		return false
	}
	if !fn(n) {
		return false
	}
	return t.inorderNode(n.right, fn)
}

func (t *rbTree[Key]) clone() *rbTree[Key] {
	newT := &rbTree[Key]{}
	if t.root != nil {
		newT.root = cloneNode(t.root, nil)
	}
	newT.size = t.size
	return newT
}

func cloneNode[Key comparable](n *rbNode[Key], p *rbNode[Key]) *rbNode[Key] {
	if n == nil {
		return nil
	}
	c := &rbNode[Key]{
		score: n.score,
		value: n.value,
		color: n.color,
		p:     p,
	}
	c.left = cloneNode(n.left, c)
	c.right = cloneNode(n.right, c)
	return c
}

func (t *rbTree[Key]) find(score Score) *rbNode[Key] {
	x := t.root
	for x != nil {
		cmp := score.Compare(x.score)
		if cmp == 0 {
			return x
		} else if cmp > 0 {
			x = x.right
		} else {
			x = x.left
		}
	}
	return nil
}

func (t *rbTree[Key]) rank(score Score) int {
	rank := 0
	x := t.root
	for x != nil {
		cmp := score.Compare(x.score)
		if cmp == 0 {
			if x.left != nil {
				rank += countNodes(x.left)
			}
			rank++
			break
		} else if cmp < 0 {
			x = x.left
		} else {
			if x.left != nil {
				rank += countNodes(x.left)
			}
			rank++
			x = x.right
		}
	}
	return rank
}

func countNodes[Key comparable](n *rbNode[Key]) int {
	if n == nil {
		return 0
	}
	return 1 + countNodes(n.left) + countNodes(n.right)
}
