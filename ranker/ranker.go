package ranker

import (
	"fmt"
	"sync"
)

type Ranker struct {
	mu    sync.RWMutex
	lists map[string]*RankList
}

func NewRanker() *Ranker {
	return &Ranker{
		lists: make(map[string]*RankList),
	}
}

func (r *Ranker) Create(name string, capacity int32) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.lists[name]; exists {
		return fmt.Errorf("ranklist %q already exists", name)
	}
	r.lists[name] = NewRankList(name, capacity)
	return nil
}

func (r *Ranker) Get(name string) (*RankList, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	rl, exists := r.lists[name]
	return rl, exists
}

func (r *Ranker) Exists(name string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	_, exists := r.lists[name]
	return exists
}

func (r *Ranker) Delete(name string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.lists[name]; !exists {
		return false
	}
	delete(r.lists, name)
	return true
}

func (r *Ranker) Put(name string, playerID int64, score Score) (updated, accepted bool, err error) {
	r.mu.RLock()
	rl, exists := r.lists[name]
	r.mu.RUnlock()
	if !exists {
		return false, false, fmt.Errorf("ranklist %q not found", name)
	}
	u, a := rl.Put(playerID, score)
	return u, a, nil
}

func (r *Ranker) Remove(name string, playerID int64) (bool, error) {
	r.mu.RLock()
	rl, exists := r.lists[name]
	r.mu.RUnlock()
	if !exists {
		return false, fmt.Errorf("ranklist %q not found", name)
	}
	return rl.Remove(playerID), nil
}

func (r *Ranker) Snapshot(name string) (*RankList, error) {
	r.mu.RLock()
	rl, exists := r.lists[name]
	r.mu.RUnlock()
	if !exists {
		return nil, fmt.Errorf("ranklist %q not found", name)
	}
	return rl.Snapshot(), nil
}

func (r *Ranker) SnapshotAll() map[string]*RankList {
	r.mu.RLock()
	defer r.mu.RUnlock()
	snapshots := make(map[string]*RankList, len(r.lists))
	for name, rl := range r.lists {
		snapshots[name] = rl.Snapshot()
	}
	return snapshots
}

func (r *Ranker) ListNames() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	names := make([]string, 0, len(r.lists))
	for name := range r.lists {
		names = append(names, name)
	}
	return names
}

func (r *Ranker) Len(name string) (int32, error) {
	r.mu.RLock()
	rl, exists := r.lists[name]
	r.mu.RUnlock()
	if !exists {
		return 0, fmt.Errorf("ranklist %q not found", name)
	}
	return rl.Len(), nil
}

func (r *Ranker) Version(name string) (int64, error) {
	r.mu.RLock()
	rl, exists := r.lists[name]
	r.mu.RUnlock()
	if !exists {
		return 0, fmt.Errorf("ranklist %q not found", name)
	}
	return rl.Version(), nil
}
