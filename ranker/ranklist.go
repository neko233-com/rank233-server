package ranker

import "sync"

type RankList struct {
	mu       sync.RWMutex
	name     string
	version  int64
	capacity int32
	tree     *rbTree[int64]
	scores   map[int64]Score
}

func NewRankList(name string, capacity int32) *RankList {
	if capacity <= 0 {
		capacity = 10000
	}
	return &RankList{
		name:     name,
		capacity: capacity,
		tree:     &rbTree[int64]{},
		scores:   make(map[int64]Score),
	}
}

func (rl *RankList) clone() *RankList {
	newScores := make(map[int64]Score, len(rl.scores))
	for k, v := range rl.scores {
		newScores[k] = v
	}
	return &RankList{
		name:     rl.name,
		version:  rl.version,
		capacity: rl.capacity,
		tree:     rl.tree.clone(),
		scores:   newScores,
	}
}

func (rl *RankList) Snapshot() *RankList {
	rl.mu.RLock()
	defer rl.mu.RUnlock()
	return rl.clone()
}

func (rl *RankList) Name() string    { return rl.name }
func (rl *RankList) Capacity() int32 { return rl.capacity }

func (rl *RankList) Version() int64 {
	rl.mu.RLock()
	defer rl.mu.RUnlock()
	return rl.version
}

func (rl *RankList) IsEmpty() bool {
	rl.mu.RLock()
	defer rl.mu.RUnlock()
	return rl.tree.size == 0
}

func (rl *RankList) IsFull() bool {
	rl.mu.RLock()
	defer rl.mu.RUnlock()
	return int32(rl.tree.size) >= rl.capacity
}

func (rl *RankList) Len() int32 {
	rl.mu.RLock()
	defer rl.mu.RUnlock()
	return int32(rl.tree.size)
}

func (rl *RankList) Put(playerID int64, score Score) (updated, accepted bool) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	old, exists := rl.scores[playerID]
	if exists {
		if score.Compare(old) == 0 {
			return false, true
		}
		node := rl.tree.find(old)
		if node != nil {
			rl.tree.delete(node)
		}
		delete(rl.scores, playerID)
	} else if int32(rl.tree.size) >= rl.capacity {
		worst := rl.tree.maximum(rl.tree.root)
		if worst != nil && score.Compare(worst.score) >= 0 {
			return false, false
		}
		rl.tree.delete(worst)
		delete(rl.scores, worst.value)
	}

	rl.tree.insert(&rbNode[int64]{score: score, value: playerID})
	rl.scores[playerID] = score
	rl.version++
	return true, true
}

func (rl *RankList) Remove(playerID int64) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	score, exists := rl.scores[playerID]
	if !exists {
		return false
	}
	node := rl.tree.find(score)
	if node != nil {
		rl.tree.delete(node)
	}
	delete(rl.scores, playerID)
	rl.version++
	return true
}

func (rl *RankList) Has(playerID int64) bool {
	rl.mu.RLock()
	defer rl.mu.RUnlock()
	_, exists := rl.scores[playerID]
	return exists
}

func (rl *RankList) GetScore(playerID int64) (Score, bool) {
	rl.mu.RLock()
	defer rl.mu.RUnlock()
	score, exists := rl.scores[playerID]
	return score, exists
}

func (rl *RankList) GetRank(playerID int64) (int32, bool) {
	rl.mu.RLock()
	defer rl.mu.RUnlock()
	score, exists := rl.scores[playerID]
	if !exists {
		return 0, false
	}
	return int32(rl.tree.rank(score)), true
}

func (rl *RankList) GetTopN(limit int32) []ScoredEntry {
	rl.mu.RLock()
	defer rl.mu.RUnlock()
	if limit <= 0 || rl.tree.size == 0 {
		return nil
	}
	if limit > int32(rl.tree.size) {
		limit = int32(rl.tree.size)
	}
	entries := make([]ScoredEntry, 0, limit)
	var rank int32
	rl.tree.inorder(func(n *rbNode[int64]) bool {
		rank++
		entries = append(entries, ScoredEntry{
			PlayerID: n.value,
			Score:    n.score,
			Rank:     rank,
		})
		return rank < limit
	})
	return entries
}

func (rl *RankList) GetRange(startRank, endRank int32) []ScoredEntry {
	rl.mu.RLock()
	defer rl.mu.RUnlock()
	return rl.getRangeLocked(startRank, endRank)
}

func (rl *RankList) getRangeLocked(startRank, endRank int32) []ScoredEntry {
	if startRank <= 0 {
		startRank = 1
	}
	if endRank > int32(rl.tree.size) {
		endRank = int32(rl.tree.size)
	}
	if startRank > endRank || rl.tree.size == 0 {
		return nil
	}
	entries := make([]ScoredEntry, 0, endRank-startRank+1)
	var rank int32
	rl.tree.inorder(func(n *rbNode[int64]) bool {
		rank++
		if rank >= startRank && rank <= endRank {
			entries = append(entries, ScoredEntry{
				PlayerID: n.value,
				Score:    n.score,
				Rank:     rank,
			})
		}
		return rank <= endRank
	})
	return entries
}

func (rl *RankList) Around(playerID int64, beforeCount, afterCount int32) []ScoredEntry {
	rl.mu.RLock()
	defer rl.mu.RUnlock()
	score, exists := rl.scores[playerID]
	if !exists {
		return nil
	}
	rank := int32(rl.tree.rank(score))
	startRank := rank - beforeCount
	if startRank < 1 {
		startRank = 1
	}
	endRank := rank + afterCount
	total := int32(rl.tree.size)
	if endRank > total {
		endRank = total
	}
	return rl.getRangeLocked(startRank, endRank)
}

func (rl *RankList) Entries() []ScoredEntry {
	rl.mu.RLock()
	defer rl.mu.RUnlock()
	entries := make([]ScoredEntry, 0, rl.tree.size)
	var rank int32
	rl.tree.inorder(func(n *rbNode[int64]) bool {
		rank++
		entries = append(entries, ScoredEntry{
			PlayerID: n.value,
			Score:    n.score,
			Rank:     rank,
		})
		return true
	})
	return entries
}
