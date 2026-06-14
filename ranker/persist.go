package ranker

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type PersistEntry struct {
	PlayerID  int64 `json:"player_id"`
	Primary   int64 `json:"primary"`
	Secondary int64 `json:"secondary"`
	Arrival   int64 `json:"arrival"`
}

type PersistRankList struct {
	Name     string         `json:"name"`
	Capacity int32          `json:"capacity"`
	Version  int64          `json:"version"`
	Entries  []PersistEntry `json:"entries"`
}

type Persister struct {
	mu       sync.Mutex
	dir      string
	interval time.Duration
	ranker   *Ranker
	stopCh   chan struct{}
}

func NewPersister(ranker *Ranker, dir string, interval time.Duration) *Persister {
	if interval <= 0 {
		interval = 60 * time.Second
	}
	return &Persister{
		dir:      dir,
		interval: interval,
		ranker:   ranker,
		stopCh:   make(chan struct{}),
	}
}

func (p *Persister) Start() {
	os.MkdirAll(p.dir, 0755)
	go func() {
		ticker := time.NewTicker(p.interval)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				p.SaveAll()
			case <-p.stopCh:
				return
			}
		}
	}()
}

func (p *Persister) Stop() {
	close(p.stopCh)
}

func (p *Persister) SaveAll() error {
	p.mu.Lock()
	defer p.mu.Unlock()
	snapshots := p.ranker.SnapshotAll()
	for name, rl := range snapshots {
		if err := p.saveOne(name, rl); err != nil {
			return fmt.Errorf("save %q: %w", name, err)
		}
	}
	return nil
}

func (p *Persister) Save(name string) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	snap, err := p.ranker.Snapshot(name)
	if err != nil {
		return err
	}
	return p.saveOne(name, snap)
}

func (p *Persister) saveOne(name string, rl *RankList) error {
	entries := rl.Entries()
	persist := PersistRankList{
		Name:     name,
		Capacity: rl.Capacity(),
		Version:  rl.Version(),
		Entries:  make([]PersistEntry, len(entries)),
	}
	for i, e := range entries {
		persist.Entries[i] = PersistEntry{
			PlayerID:  e.PlayerID,
			Primary:   e.Score.Primary,
			Secondary: e.Score.Secondary,
			Arrival:   e.Score.Arrival,
		}
	}
	data, err := json.Marshal(persist)
	if err != nil {
		return err
	}
	path := filepath.Join(p.dir, name+".json")
	return os.WriteFile(path, data, 0644)
}

func (p *Persister) LoadAll() (int, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	entries, err := os.ReadDir(p.dir)
	if err != nil {
		if os.IsNotExist(err) {
			return 0, nil
		}
		return 0, err
	}
	loaded := 0
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
			continue
		}
		name := entry.Name()[:len(entry.Name())-5]
		if err := p.loadOne(name); err != nil {
			return loaded, fmt.Errorf("load %q: %w", name, err)
		}
		loaded++
	}
	return loaded, nil
}

func (p *Persister) loadOne(name string) error {
	path := filepath.Join(p.dir, name+".json")
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	var persist PersistRankList
	if err := json.Unmarshal(data, &persist); err != nil {
		return err
	}
	if p.ranker.Exists(name) {
		return nil
	}
	rl := NewRankList(persist.Name, persist.Capacity)
	for _, e := range persist.Entries {
		rl.Put(e.PlayerID, Score{
			Primary:   e.Primary,
			Secondary: e.Secondary,
			Arrival:   e.Arrival,
		})
	}
	p.ranker.mu.Lock()
	p.ranker.lists[name] = rl
	p.ranker.mu.Unlock()
	return nil
}

func (p *Persister) CleanBefore(name string, beforeVersion int64) (int, error) {
	rl, ok := p.ranker.Get(name)
	if !ok {
		return 0, fmt.Errorf("ranklist %q not found", name)
	}
	removed := 0
	entries := rl.Entries()
	for _, e := range entries {
		if e.Score.Arrival < beforeVersion {
			rl.Remove(e.PlayerID)
			removed++
		}
	}
	return removed, nil
}

func (p *Persister) CleanFile(name string) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	path := filepath.Join(p.dir, name+".json")
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}
