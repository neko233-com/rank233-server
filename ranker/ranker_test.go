package ranker

import (
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"
)

func sc(primary, secondary, arrival int64) Score {
	return Score{Primary: primary, Secondary: secondary, Arrival: arrival}
}

func TestRankList_Basic(t *testing.T) {
	rl := NewRankList("test", 100)
	if !rl.IsEmpty() {
		t.Fatal("should be empty")
	}
	rl.Put(1, sc(100, 0, 1))
	rl.Put(2, sc(200, 0, 1))
	rl.Put(3, sc(150, 0, 1))
	if rl.Len() != 3 {
		t.Fatalf("expected 3, got %d", rl.Len())
	}
	top := rl.GetTopN(10)
	if len(top) != 3 {
		t.Fatalf("expected 3 entries, got %d", len(top))
	}
	if top[0].PlayerID != 2 {
		t.Fatalf("rank 1 should be 2, got %d", top[0].PlayerID)
	}
	if top[1].PlayerID != 3 {
		t.Fatalf("rank 2 should be 3, got %d", top[1].PlayerID)
	}
	if top[2].PlayerID != 1 {
		t.Fatalf("rank 3 should be 1, got %d", top[2].PlayerID)
	}
}

func TestRankList_UpdateScore(t *testing.T) {
	rl := NewRankList("test", 100)
	rl.Put(1, sc(100, 0, 1))
	rl.Put(2, sc(200, 0, 1))
	rank, _ := rl.GetRank(1)
	if rank != 2 {
		t.Fatalf("player 1 should be rank 2, got %d", rank)
	}
	rl.Put(1, sc(300, 0, 1))
	rank, _ = rl.GetRank(1)
	if rank != 1 {
		t.Fatalf("player 1 should be rank 1 after update, got %d", rank)
	}
	if rl.Len() != 2 {
		t.Fatalf("should still have 2 members, got %d", rl.Len())
	}
}

func TestRankList_Remove(t *testing.T) {
	rl := NewRankList("test", 100)
	rl.Put(1, sc(100, 0, 1))
	rl.Put(2, sc(200, 0, 1))
	if !rl.Remove(1) {
		t.Fatal("should remove player 1")
	}
	if rl.Remove(1) {
		t.Fatal("should not remove player 1 again")
	}
	if rl.Has(1) {
		t.Fatal("player 1 should not exist")
	}
	if rl.Len() != 1 {
		t.Fatalf("expected 1, got %d", rl.Len())
	}
}

func TestRankList_TimeTiebreak(t *testing.T) {
	rl := NewRankList("test", 100)
	rl.Put(1, sc(100, 0, 1000))
	rl.Put(2, sc(100, 0, 2000))
	top := rl.GetTopN(2)
	if top[0].PlayerID != 1 {
		t.Fatalf("earlier arrival should rank first, got %d", top[0].PlayerID)
	}
}

func TestRankList_Capacity(t *testing.T) {
	rl := NewRankList("test", 2)
	rl.Put(1, sc(300, 0, 1))
	rl.Put(2, sc(200, 0, 1))
	if !rl.IsFull() {
		t.Fatal("should be full")
	}
	_, accepted := rl.Put(3, sc(100, 0, 1))
	if accepted {
		t.Fatal("should reject lower score when full")
	}
	updated, accepted := rl.Put(3, sc(400, 0, 1))
	if !accepted || !updated {
		t.Fatal("should accept higher score when full")
	}
	if rl.Len() != 2 {
		t.Fatalf("should have 2 entries, got %d", rl.Len())
	}
}

func TestRankList_GetRange(t *testing.T) {
	rl := NewRankList("test", 100)
	for i := 0; i < 10; i++ {
		rl.Put(int64(i), sc(int64(100-i*10), 0, int64(i)))
	}
	entries := rl.GetRange(3, 7)
	if len(entries) != 5 {
		t.Fatalf("expected 5 entries, got %d", len(entries))
	}
	for i, e := range entries {
		if e.Rank != int32(3+i) {
			t.Fatalf("expected rank %d, got %d", 3+i, e.Rank)
		}
	}
}

func TestRankList_AroundMe(t *testing.T) {
	rl := NewRankList("test", 100)
	for i := 0; i < 10; i++ {
		rl.Put(int64(i), sc(int64(100-i*10), 0, int64(i)))
	}
	entries := rl.Around(5, 2, 2)
	if len(entries) != 5 {
		t.Fatalf("expected 5 entries, got %d", len(entries))
	}
	if entries[2].PlayerID != 5 {
		t.Fatalf("center should be 5, got %d", entries[2].PlayerID)
	}
}

func TestRankList_SnapshotIsolation(t *testing.T) {
	rl := NewRankList("test", 100)
	rl.Put(1, sc(100, 0, 1))
	rl.Put(2, sc(200, 0, 1))
	snap := rl.Snapshot()
	rl.Put(1, sc(300, 0, 1))
	rank, _ := snap.GetRank(1)
	if rank != 2 {
		t.Fatalf("snapshot should show player 1 at rank 2, got %d", rank)
	}
	rank, _ = rl.GetRank(1)
	if rank != 1 {
		t.Fatalf("live should show player 1 at rank 1, got %d", rank)
	}
}

func TestRankList_CompositeScore(t *testing.T) {
	rl := NewRankList("test", 100)
	rl.Put(1, sc(100, 50, 1000))
	rl.Put(2, sc(100, 50, 2000))
	rl.Put(3, sc(100, 40, 500))
	rl.Put(4, sc(200, 10, 3000))
	top := rl.GetTopN(4)
	if top[0].PlayerID != 4 {
		t.Fatalf("rank 1 should be 4, got %d", top[0].PlayerID)
	}
	if top[1].PlayerID != 1 {
		t.Fatalf("rank 2 should be 1, got %d", top[1].PlayerID)
	}
	if top[2].PlayerID != 2 {
		t.Fatalf("rank 3 should be 2, got %d", top[2].PlayerID)
	}
	if top[3].PlayerID != 3 {
		t.Fatalf("rank 4 should be 3, got %d", top[3].PlayerID)
	}
}

func TestRankList_LargeScale(t *testing.T) {
	const n = 10000
	rl := NewRankList("test", n)
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := 0; i < n; i++ {
		rl.Put(int64(i), sc(rng.Int63n(100000), rng.Int63n(1000), rng.Int63n(100000)))
	}
	if rl.Len() != n {
		t.Fatalf("expected %d entries, got %d", n, rl.Len())
	}
	top := rl.GetTopN(10)
	for i := 1; i < len(top); i++ {
		if top[i].Score.Compare(top[i-1].Score) < 0 {
			t.Fatalf("top N not sorted at position %d", i)
		}
	}
}

func TestRankList_ConcurrentSnapshot(t *testing.T) {
	rl := NewRankList("test", 5000)
	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < 50; j++ {
				key := int64(id*50 + j)
				rl.Put(key, sc(key, 0, int64(j)))
			}
		}(i)
	}
	wg.Wait()
	snap := rl.Snapshot()
	if snap.Len() != rl.Len() {
		t.Fatalf("snapshot len mismatch: %d vs %d", snap.Len(), rl.Len())
	}
	snapTop := snap.GetTopN(10)
	liveTop := rl.GetTopN(10)
	for i := range snapTop {
		if snapTop[i].PlayerID != liveTop[i].PlayerID {
			t.Fatalf("snapshot/live mismatch at rank %d", i+1)
		}
	}
}

func TestRankList_Stress(t *testing.T) {
	rl := NewRankList("stress", 1000)
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	for round := 0; round < 100; round++ {
		for i := 0; i < 100; i++ {
			key := int64(rng.Intn(5000))
			rl.Put(key, sc(rng.Int63n(10000), rng.Int63n(1000), rng.Int63n(100000)))
		}
		for i := 0; i < 10; i++ {
			key := int64(rng.Intn(5000))
			rl.Remove(key)
		}
		top := rl.GetTopN(10)
		for i := 1; i < len(top); i++ {
			if top[i].Score.Compare(top[i-1].Score) < 0 {
				t.Fatalf("not sorted at round %d position %d", round, i)
			}
		}
	}
}

func TestRankList_Version(t *testing.T) {
	rl := NewRankList("test", 100)
	if rl.Version() != 0 {
		t.Fatal("initial version should be 0")
	}
	rl.Put(1, sc(100, 0, 1))
	if rl.Version() != 1 {
		t.Fatal("version should be 1 after put")
	}
	rl.Put(2, sc(200, 0, 1))
	if rl.Version() != 2 {
		t.Fatal("version should be 2 after second put")
	}
	rl.Remove(1)
	if rl.Version() != 3 {
		t.Fatal("version should be 3 after remove")
	}
}

func TestRanker_Basic(t *testing.T) {
	r := NewRanker()
	r.Create("global", 1000)
	r.Create("server1", 100)
	if !r.Exists("global") {
		t.Fatal("global should exist")
	}
	if r.Exists("nonexistent") {
		t.Fatal("nonexistent should not exist")
	}
}

func TestRanker_DuplicateCreate(t *testing.T) {
	r := NewRanker()
	r.Create("test", 100)
	err := r.Create("test", 200)
	if err == nil {
		t.Fatal("should error on duplicate create")
	}
}

func TestRanker_PutGet(t *testing.T) {
	r := NewRanker()
	r.Create("global", 1000)
	r.Put("global", 1, sc(100, 0, 1))
	r.Put("global", 2, sc(200, 0, 1))
	snap, _ := r.Snapshot("global")
	top := snap.GetTopN(10)
	if len(top) != 2 {
		t.Fatalf("expected 2, got %d", len(top))
	}
	if top[0].PlayerID != 2 {
		t.Fatalf("rank 1 should be 2, got %d", top[0].PlayerID)
	}
}

func TestRanker_MultiList(t *testing.T) {
	r := NewRanker()
	r.Create("global", 1000)
	r.Create("s1", 100)
	r.Create("s2", 100)
	r.Put("global", 1, sc(100, 0, 1))
	r.Put("s1", 1, sc(200, 0, 1))
	r.Put("s2", 1, sc(150, 0, 1))
	gSnap, _ := r.Snapshot("global")
	s1Snap, _ := r.Snapshot("s1")
	s2Snap, _ := r.Snapshot("s2")
	gRank, _ := gSnap.GetRank(1)
	s1Rank, _ := s1Snap.GetRank(1)
	s2Rank, _ := s2Snap.GetRank(1)
	if gRank != 1 || s1Rank != 1 || s2Rank != 1 {
		t.Fatalf("player 1 should be rank 1 everywhere: global=%d s1=%d s2=%d", gRank, s1Rank, s2Rank)
	}
	gScore, _ := gSnap.GetScore(1)
	s1Score, _ := s1Snap.GetScore(1)
	if gScore.Primary != 100 || s1Score.Primary != 200 {
		t.Fatal("scores should differ per list")
	}
}

func TestRanker_Concurrent(t *testing.T) {
	r := NewRanker()
	r.Create("global", 10000)
	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				key := int64(id*100 + j)
				r.Put("global", key, sc(key, 0, int64(j)))
			}
		}(i)
	}
	wg.Wait()
	snap, _ := r.Snapshot("global")
	if snap.Len() != 10000 {
		t.Fatalf("expected 10000 entries, got %d", snap.Len())
	}
}

func TestRanker_GlobalVsLocal(t *testing.T) {
	r := NewRanker()
	r.Create("global", 10000)
	r.Create("s1", 1000)
	r.Create("s2", 1000)
	type player struct {
		id                            int64
		gScore, s1Score, s2Score Score
	}
	players := []player{
		{1, sc(1000, 0, 1), sc(2000, 0, 1), sc(500, 0, 1)},
		{2, sc(800, 0, 2), sc(1500, 0, 2), sc(900, 0, 2)},
		{3, sc(1200, 0, 3), sc(300, 0, 3), sc(1100, 0, 3)},
	}
	for _, p := range players {
		r.Put("global", p.id, p.gScore)
		r.Put("s1", p.id, p.s1Score)
		r.Put("s2", p.id, p.s2Score)
	}
	gSnap, _ := r.Snapshot("global")
	s1Snap, _ := r.Snapshot("s1")
	s2Snap, _ := r.Snapshot("s2")
	gTop := gSnap.GetTopN(3)
	s1Top := s1Snap.GetTopN(3)
	s2Top := s2Snap.GetTopN(3)
	if gTop[0].PlayerID != 3 {
		t.Fatalf("global rank 1 should be 3, got %d", gTop[0].PlayerID)
	}
	if s1Top[0].PlayerID != 1 {
		t.Fatalf("s1 rank 1 should be 1, got %d", s1Top[0].PlayerID)
	}
	if s2Top[0].PlayerID != 3 {
		t.Fatalf("s2 rank 1 should be 3, got %d", s2Top[0].PlayerID)
	}
}

func TestRanker_Delete(t *testing.T) {
	r := NewRanker()
	r.Create("test", 100)
	r.Put("test", 1, sc(100, 0, 1))
	if !r.Delete("test") {
		t.Fatal("should delete")
	}
	if r.Exists("test") {
		t.Fatal("should not exist after delete")
	}
}

func TestRanker_ListNames(t *testing.T) {
	r := NewRanker()
	r.Create("a", 100)
	r.Create("b", 100)
	r.Create("c", 100)
	names := r.ListNames()
	if len(names) != 3 {
		t.Fatalf("expected 3, got %d", len(names))
	}
}

func TestRanker_NotFound(t *testing.T) {
	r := NewRanker()
	_, _, err := r.Put("nonexistent", 1, sc(100, 0, 1))
	if err == nil {
		t.Fatal("should error on nonexistent list")
	}
	_, err = r.Remove("nonexistent", 1)
	if err == nil {
		t.Fatal("should error on nonexistent list")
	}
	_, err = r.Snapshot("nonexistent")
	if err == nil {
		t.Fatal("should error on nonexistent list")
	}
}

func TestPersist_SaveLoad(t *testing.T) {
	dir := filepath.Join(os.TempDir(), fmt.Sprintf("rank233-test-%d", time.Now().UnixNano()))
	os.MkdirAll(dir, 0755)
	defer os.RemoveAll(dir)

	r := NewRanker()
	r.Create("test", 100)
	r.Put("test", 1, sc(100, 0, 1000))
	r.Put("test", 2, sc(200, 0, 2000))
	r.Put("test", 3, sc(150, 0, 3000))

	p := NewPersister(r, dir, time.Hour)
	if err := p.SaveAll(); err != nil {
		t.Fatalf("save failed: %v", err)
	}

	r2 := NewRanker()
	p2 := NewPersister(r2, dir, time.Hour)
	loaded, err := p2.LoadAll()
	if err != nil {
		t.Fatalf("load failed: %v", err)
	}
	if loaded != 1 {
		t.Fatalf("expected 1 loaded, got %d", loaded)
	}

	snap, _ := r2.Snapshot("test")
	if snap.Len() != 3 {
		t.Fatalf("expected 3 entries, got %d", snap.Len())
	}
	top := snap.GetTopN(3)
	if top[0].PlayerID != 2 {
		t.Fatalf("rank 1 should be 2, got %d", top[0].PlayerID)
	}
	if top[0].Score.Primary != 200 {
		t.Fatalf("score should be 200, got %d", top[0].Score.Primary)
	}
}

func TestPersist_CleanBefore(t *testing.T) {
	dir := filepath.Join(os.TempDir(), fmt.Sprintf("rank233-clean-%d", time.Now().UnixNano()))
	os.MkdirAll(dir, 0755)
	defer os.RemoveAll(dir)

	r := NewRanker()
	r.Create("test", 100)
	r.Put("test", 1, sc(100, 0, 1000))
	r.Put("test", 2, sc(200, 0, 2000))
	r.Put("test", 3, sc(150, 0, 3000))

	p := NewPersister(r, dir, time.Hour)
	removed, err := p.CleanBefore("test", 2500)
	if err != nil {
		t.Fatalf("clean failed: %v", err)
	}
	if removed != 2 {
		t.Fatalf("expected 2 removed, got %d", removed)
	}
	if r.lists["test"].Len() != 1 {
		t.Fatalf("expected 1 remaining, got %d", r.lists["test"].Len())
	}
}
