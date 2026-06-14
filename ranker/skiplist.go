package ranker

import (
	"math/rand"
)

const (
	maxLevel = 32
	pValue   = 0.25
)

type skipNode struct {
	score Score
	key   int64
	next  [maxLevel]*skipNode
}

type skipList struct {
	head  *skipNode
	level int
	size  int
}

func newSkipList() *skipList {
	return &skipList{
		head:  &skipNode{},
		level: 1,
	}
}

func randomLevel() int {
	level := 1
	for level < maxLevel && rand.Float64() < pValue {
		level++
	}
	return level
}

func skipLess(scoreA, keyA int64, scoreB, keyB int64) bool {
	if scoreA != scoreB {
		return scoreA > scoreB
	}
	return keyA < keyB
}

func skipLessScore(scoreA Score, keyA int64, scoreB Score, keyB int64) bool {
	cmp := scoreA.Compare(scoreB)
	if cmp != 0 {
		return cmp < 0
	}
	return keyA < keyB
}

func (sl *skipList) Insert(score Score, key int64) {
	update := [maxLevel]*skipNode{}
	x := sl.head

	for i := sl.level - 1; i >= 0; i-- {
		for x.next[i] != nil && skipLessScore(x.next[i].score, x.next[i].key, score, key) {
			x = x.next[i]
		}
		update[i] = x
	}

	next := x.next[0]
	if next != nil && next.score.Compare(score) == 0 && next.key == key {
		for i := 0; i < sl.level; i++ {
			if update[i].next[i] != next {
				break
			}
			update[i].next[i] = next.next[i]
		}
		sl.size--
		for sl.level > 1 && sl.head.next[sl.level-1] == nil {
			sl.level--
		}

		x = sl.head
		for i := sl.level - 1; i >= 0; i-- {
			for x.next[i] != nil && skipLessScore(x.next[i].score, x.next[i].key, score, key) {
				x = x.next[i]
			}
			update[i] = x
		}
	}

	lvl := randomLevel()
	if lvl > sl.level {
		for i := sl.level; i < lvl; i++ {
			update[i] = sl.head
		}
		sl.level = lvl
	}

	node := &skipNode{score: score, key: key}
	for i := 0; i < lvl; i++ {
		node.next[i] = update[i].next[i]
		update[i].next[i] = node
	}
	sl.size++
}

func (sl *skipList) Delete(score Score, key int64) bool {
	update := [maxLevel]*skipNode{}
	x := sl.head

	for i := sl.level - 1; i >= 0; i-- {
		for x.next[i] != nil && skipLessScore(x.next[i].score, x.next[i].key, score, key) {
			x = x.next[i]
		}
		update[i] = x
	}

	target := x.next[0]
	if target == nil || target.score.Compare(score) != 0 || target.key != key {
		return false
	}

	for i := 0; i < sl.level; i++ {
		if update[i].next[i] != target {
			break
		}
		update[i].next[i] = target.next[i]
	}

	sl.size--
	for sl.level > 1 && sl.head.next[sl.level-1] == nil {
		sl.level--
	}
	return true
}

func (sl *skipList) Find(score Score, key int64) bool {
	x := sl.head
	for i := sl.level - 1; i >= 0; i-- {
		for x.next[i] != nil && skipLessScore(x.next[i].score, x.next[i].key, score, key) {
			x = x.next[i]
		}
	}
	n := x.next[0]
	return n != nil && n.score.Compare(score) == 0 && n.key == key
}

func (sl *skipList) Rank(score Score) int {
	x := sl.head
	for i := sl.level - 1; i >= 0; i-- {
		for x.next[i] != nil && scoreLess(x.next[i].score, score) {
			x = x.next[i]
		}
	}
	rank := 0
	cur := sl.head.next[0]
	for cur != nil && cur != x.next[0] {
		if scoreLess(cur.score, score) {
			rank++
		}
		cur = cur.next[0]
	}
	if x.next[0] != nil && x.next[0].score.Compare(score) == 0 {
		rank++
	}
	return rank
}

func (sl *skipList) TopN(limit int) []*skipNode {
	if limit <= 0 || sl.size == 0 {
		return nil
	}
	if limit > sl.size {
		limit = sl.size
	}
	nodes := make([]*skipNode, 0, limit)
	x := sl.head.next[0]
	for x != nil && len(nodes) < limit {
		nodes = append(nodes, x)
		x = x.next[0]
	}
	return nodes
}

func (sl *skipList) Range(startRank, endRank int) []*skipNode {
	if startRank <= 0 {
		startRank = 1
	}
	if endRank > sl.size {
		endRank = sl.size
	}
	if startRank > endRank || sl.size == 0 {
		return nil
	}

	nodes := make([]*skipNode, 0, endRank-startRank+1)
	rank := 0
	x := sl.head

	for i := sl.level - 1; i >= 0; i-- {
		for x.next[i] != nil {
			nextRank := rank + sl.countBetween(x, x.next[i])
			if nextRank < startRank {
				rank = nextRank
				x = x.next[i]
			} else {
				break
			}
		}
	}

	cur := x.next[0]
	for cur != nil && rank < endRank {
		rank++
		if rank >= startRank {
			nodes = append(nodes, cur)
		}
		cur = cur.next[0]
	}
	return nodes
}

func (sl *skipList) countBetween(from, to *skipNode) int {
	count := 0
	cur := from.next[0]
	for cur != nil && cur != to {
		count++
		cur = cur.next[0]
	}
	return count + 1
}

func (sl *skipList) Len() int {
	return sl.size
}

func (sl *skipList) clone() *skipList {
	newSl := newSkipList()
	newSl.level = sl.level
	newSl.size = sl.size

	x := sl.head.next[0]
	var prev *skipNode
	for x != nil {
		node := &skipNode{score: x.score, key: x.key}
		if prev == nil {
			newSl.head.next[0] = node
		} else {
			prev.next[0] = node
		}
		prev = node
		x = x.next[0]
	}

	for i := 1; i < sl.level; i++ {
		src := sl.head.next[i]
		var dstPrev *skipNode
		for src != nil {
			n := newSl.findNode(src.key)
			if n != nil {
				if dstPrev == nil {
					newSl.head.next[i] = n
				} else {
					dstPrev.next[i] = n
				}
				dstPrev = n
			}
			src = src.next[i]
		}
	}

	return newSl
}

func (sl *skipList) findNode(key int64) *skipNode {
	x := sl.head.next[0]
	for x != nil {
		if x.key == key {
			return x
		}
		x = x.next[0]
	}
	return nil
}

func scoreLess(a, b Score) bool {
	return a.Compare(b) < 0
}
