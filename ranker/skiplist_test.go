package ranker

import (
	"fmt"
	"math/rand"
	"sync"
	"testing"
	"time"
)

func TestSkipList_Basic(t *testing.T) {
	sl := newSkipList()
	sl.Insert(sc(100, 0, 1), 1)
	sl.Insert(sc(200, 0, 1), 2)
	sl.Insert(sc(150, 0, 1), 3)

	if sl.Len() != 3 {
		t.Fatalf("expected 3, got %d", sl.Len())
	}

	top := sl.TopN(10)
	if len(top) != 3 {
		t.Fatalf("expected 3, got %d", len(top))
	}
	if top[0].key != 2 {
		t.Fatalf("rank 1 should be 2, got %d", top[0].key)
	}
	if top[1].key != 3 {
		t.Fatalf("rank 2 should be 3, got %d", top[1].key)
	}
	if top[2].key != 1 {
		t.Fatalf("rank 3 should be 1, got %d", top[2].key)
	}
}

func TestSkipList_Update(t *testing.T) {
	sl := newSkipList()
	sl.Insert(sc(100, 0, 1), 1)
	sl.Insert(sc(200, 0, 1), 2)

	rank := sl.Rank(sc(100, 0, 1))
	if rank != 2 {
		t.Fatalf("player 1 should be rank 2, got %d", rank)
	}

	sl.Insert(sc(300, 0, 1), 1)
	rank = sl.Rank(sc(300, 0, 1))
	if rank != 1 {
		t.Fatalf("player 1 should be rank 1 after update, got %d", rank)
	}
	if sl.Len() != 2 {
		t.Fatalf("should still have 2 entries, got %d", sl.Len())
	}
}

func TestSkipList_Delete(t *testing.T) {
	sl := newSkipList()
	sl.Insert(sc(100, 0, 1), 1)
	sl.Insert(sc(200, 0, 1), 2)

	if !sl.Delete(sc(100, 0, 1), 1) {
		t.Fatal("should delete")
	}
	if sl.Delete(sc(100, 0, 1), 1) {
		t.Fatal("should not delete again")
	}
	if sl.Len() != 1 {
		t.Fatalf("expected 1, got %d", sl.Len())
	}
}

func TestSkipList_Capacity(t *testing.T) {
	sl := newSkipList()
	sl.Insert(sc(300, 0, 1), 1)
	sl.Insert(sc(200, 0, 1), 2)

	top := sl.TopN(sl.Len())
	worst := top[len(top)-1]
	if worst.key != 2 || worst.score.Primary != 200 {
		t.Fatalf("worst should be 2 with score 200, got %d/%d", worst.key, worst.score.Primary)
	}

	if sc(100, 0, 1).Compare(worst.score) >= 0 {
		t.Fatal("should reject lower score")
	}
	if sc(400, 0, 1).Compare(worst.score) < 0 {
		t.Fatal("should accept higher score")
	}
}

func TestSkipList_Range(t *testing.T) {
	sl := newSkipList()
	for i := 0; i < 10; i++ {
		sl.Insert(sc(int64(100-i*10), 0, int64(i)), int64(i))
	}

	nodes := sl.Range(3, 7)
	if len(nodes) != 5 {
		t.Fatalf("expected 5, got %d", len(nodes))
	}
}

func TestSkipList_TopN(t *testing.T) {
	sl := newSkipList()
	for i := 0; i < 100; i++ {
		sl.Insert(sc(int64(i), 0, int64(i)), int64(i))
	}

	top := sl.TopN(10)
	if len(top) != 10 {
		t.Fatalf("expected 10, got %d", len(top))
	}
	for i := 1; i < len(top); i++ {
		if top[i].score.Compare(top[i-1].score) >= 0 {
			t.Fatalf("not sorted at %d", i)
		}
	}
}

func TestSkipList_LargeScale(t *testing.T) {
	sl := newSkipList()
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := 0; i < 10000; i++ {
		sl.Insert(sc(rng.Int63n(100000), rng.Int63n(1000), rng.Int63n(100000)), int64(i))
	}
	if sl.Len() != 10000 {
		t.Fatalf("expected 10000, got %d", sl.Len())
	}
	top := sl.TopN(10)
	for i := 1; i < len(top); i++ {
		if top[i].score.Compare(top[i-1].score) >= 0 {
			t.Fatalf("not sorted at %d", i)
		}
	}
}

func TestSkipList_Concurrent(t *testing.T) {
	sl := newSkipList()
	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				key := int64(id*100 + j)
				sl.Insert(sc(key, 0, int64(j)), key)
			}
		}(i)
	}
	wg.Wait()
	if sl.Len() != 10000 {
		t.Fatalf("expected 10000, got %d", sl.Len())
	}
}

func TestSkipList_Clone(t *testing.T) {
	sl := newSkipList()
	sl.Insert(sc(100, 0, 1), 1)
	sl.Insert(sc(200, 0, 1), 2)

	clone := sl.clone()
	sl.Insert(sc(300, 0, 1), 3)

	if clone.Len() != 2 {
		t.Fatalf("clone should have 2, got %d", clone.Len())
	}
	if sl.Len() != 3 {
		t.Fatalf("original should have 3, got %d", sl.Len())
	}
}

func TestSkipList_RankStress(t *testing.T) {
	sl := newSkipList()
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	for round := 0; round < 100; round++ {
		for i := 0; i < 100; i++ {
			key := int64(rng.Intn(5000))
			sl.Insert(sc(rng.Int63n(10000), rng.Int63n(1000), rng.Int63n(100000)), key)
		}
		for i := 0; i < 10; i++ {
			key := int64(rng.Intn(5000))
			sl.Delete(sc(0, 0, 0), key)
		}
		top := sl.TopN(10)
		for i := 1; i < len(top); i++ {
			if top[i].score.Compare(top[i-1].score) >= 0 {
				t.Fatalf("not sorted at round %d pos %d", round, i)
			}
		}
	}
}

func TestRankList_Pagination(t *testing.T) {
	rl := NewRankList("test", 1000)
	for i := 0; i < 50; i++ {
		rl.Put(int64(i), sc(int64(100-i), 0, int64(i)))
	}

	entries, totalPages := rl.Page(1, 10)
	if totalPages != 5 {
		t.Fatalf("expected 5 pages, got %d", totalPages)
	}
	if len(entries) != 10 {
		t.Fatalf("expected 10 entries, got %d", len(entries))
	}
	if entries[0].Rank != 1 {
		t.Fatalf("first entry rank should be 1, got %d", entries[0].Rank)
	}
	if entries[9].Rank != 10 {
		t.Fatalf("last entry rank should be 10, got %d", entries[9].Rank)
	}

	entries2, _ := rl.Page(5, 10)
	if len(entries2) != 10 {
		t.Fatalf("page 5 should have 10 entries, got %d", len(entries2))
	}
	if entries2[0].Rank != 41 {
		t.Fatalf("page 5 first rank should be 41, got %d", entries2[0].Rank)
	}

	entries3, _ := rl.Page(6, 10)
	if len(entries3) != 0 {
		t.Fatalf("page 6 should be empty, got %d", len(entries3))
	}

	entries4, totalPages := rl.Page(1, 20)
	if totalPages != 3 {
		t.Fatalf("page_size=20 should have 3 pages, got %d", totalPages)
	}
	if len(entries4) != 20 {
		t.Fatalf("page 1 size=20 should have 20 entries, got %d", len(entries4))
	}
}

func TestRankList_PaginationEdge(t *testing.T) {
	rl := NewRankList("test", 1000)
	entries, totalPages := rl.Page(1, 10)
	if totalPages != 0 {
		t.Fatalf("empty list should have 0 pages, got %d", totalPages)
	}
	if len(entries) != 0 {
		t.Fatalf("empty list should have 0 entries, got %d", len(entries))
	}

	rl.Put(1, sc(100, 0, 1))
	entries, totalPages = rl.Page(0, 0)
	if totalPages != 1 {
		t.Fatalf("1 entry should have 1 page, got %d", totalPages)
	}
	if len(entries) != 1 {
		t.Fatalf("page(0,0) should default to page 1 size 10, got %d entries", len(entries))
	}

	fmt.Println("All pagination tests passed")
}
