package ranker

import "sync"

type RankList struct {
	mu       sync.RWMutex
	name     string
	version  int64
	capacity int32
	sl       *skipList
	scores   map[int64]Score
}

func NewRankList(name string, capacity int32) *RankList {
	if capacity <= 0 {
		capacity = 10000
	}
	return &RankList{
		name:     name,
		capacity: capacity,
		sl:       newSkipList(),
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
		sl:       rl.sl.clone(),
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
	return rl.sl.Len() == 0
}

func (rl *RankList) IsFull() bool {
	rl.mu.RLock()
	defer rl.mu.RUnlock()
	return int32(rl.sl.Len()) >= rl.capacity
}

func (rl *RankList) Len() int32 {
	rl.mu.RLock()
	defer rl.mu.RUnlock()
	return int32(rl.sl.Len())
}

func (rl *RankList) Put(playerID int64, score Score) (updated, accepted bool) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	old, exists := rl.scores[playerID]
	if exists {
		if score.Compare(old) == 0 {
			return false, true
		}
		rl.sl.Delete(old, playerID)
		delete(rl.scores, playerID)
	} else if int32(rl.sl.Len()) >= rl.capacity {
		top := rl.sl.TopN(int(rl.sl.Len()))
		if len(top) > 0 {
			worst := top[len(top)-1]
			if score.Compare(worst.score) >= 0 {
				return false, false
			}
			rl.sl.Delete(worst.score, worst.key)
			delete(rl.scores, worst.key)
		}
	}

	rl.sl.Insert(score, playerID)
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
	rl.sl.Delete(score, playerID)
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
	return int32(rl.sl.Rank(score)), true
}

func (rl *RankList) GetTopN(limit int32) []ScoredEntry {
	rl.mu.RLock()
	defer rl.mu.RUnlock()
	if limit <= 0 || rl.sl.Len() == 0 {
		return nil
	}
	nodes := rl.sl.TopN(int(limit))
	entries := make([]ScoredEntry, len(nodes))
	for i, n := range nodes {
		entries[i] = ScoredEntry{
			PlayerID: n.key,
			Score:    n.score,
			Rank:     int32(i + 1),
		}
	}
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
	if endRank > int32(rl.sl.Len()) {
		endRank = int32(rl.sl.Len())
	}
	if startRank > endRank || rl.sl.Len() == 0 {
		return nil
	}
	nodes := rl.sl.Range(int(startRank), int(endRank))
	entries := make([]ScoredEntry, len(nodes))
	for i, n := range nodes {
		entries[i] = ScoredEntry{
			PlayerID: n.key,
			Score:    n.score,
			Rank:     startRank + int32(i),
		}
	}
	return entries
}

func (rl *RankList) Page(page, pageSize int32) ([]ScoredEntry, int32) {
	rl.mu.RLock()
	defer rl.mu.RUnlock()
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	total := int32(rl.sl.Len())
	totalPages := (total + pageSize - 1) / pageSize
	if page > totalPages {
		return nil, totalPages
	}
	start := (page-1)*pageSize + 1
	end := page * pageSize
	if end > total {
		end = total
	}
	nodes := rl.sl.Range(int(start), int(end))
	entries := make([]ScoredEntry, len(nodes))
	for i, n := range nodes {
		entries[i] = ScoredEntry{
			PlayerID: n.key,
			Score:    n.score,
			Rank:     start + int32(i),
		}
	}
	return entries, totalPages
}

func (rl *RankList) Around(playerID int64, beforeCount, afterCount int32) []ScoredEntry {
	rl.mu.RLock()
	defer rl.mu.RUnlock()
	score, exists := rl.scores[playerID]
	if !exists {
		return nil
	}
	rank := int32(rl.sl.Rank(score))
	startRank := rank - beforeCount
	if startRank < 1 {
		startRank = 1
	}
	endRank := rank + afterCount
	total := int32(rl.sl.Len())
	if endRank > total {
		endRank = total
	}
	return rl.getRangeLocked(startRank, endRank)
}

func (rl *RankList) Entries() []ScoredEntry {
	rl.mu.RLock()
	defer rl.mu.RUnlock()
	nodes := rl.sl.TopN(rl.sl.Len())
	entries := make([]ScoredEntry, len(nodes))
	for i, n := range nodes {
		entries[i] = ScoredEntry{
			PlayerID: n.key,
			Score:    n.score,
			Rank:     int32(i + 1),
		}
	}
	return entries
}
