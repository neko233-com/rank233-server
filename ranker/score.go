package ranker

import (
	"fmt"
	"math"
)

type Score struct {
	Primary   int64
	Secondary int64
	Arrival   int64
}

func (s Score) Compare(other Score) int {
	if s.Primary != other.Primary {
		if s.Primary > other.Primary {
			return -1
		}
		return 1
	}
	if s.Secondary != other.Secondary {
		if s.Secondary > other.Secondary {
			return -1
		}
		return 1
	}
	if s.Arrival != other.Arrival {
		if s.Arrival < other.Arrival {
			return -1
		}
		return 1
	}
	return 0
}

func (s Score) IsZero() bool {
	return s.Primary == 0 && s.Secondary == 0 && s.Arrival == 0
}

func (s Score) String() string {
	return fmt.Sprintf("{%d, %d, %d}", s.Primary, s.Secondary, s.Arrival)
}

func (s Score) Key() int64 {
	primaryNorm := float64(s.Primary+math.MaxInt64) / float64(math.MaxUint64)
	secondaryNorm := float64(s.Secondary+math.MaxInt64) / float64(math.MaxUint64)
	timeNorm := 1.0 - float64(s.Arrival+math.MaxInt64)/float64(math.MaxUint64)
	const scale = 1 << 30
	return int64(primaryNorm*scale) + int64(secondaryNorm*scale/1000) + int64(timeNorm*scale/1000000)
}

type ScoredEntry struct {
	PlayerID int64
	Score    Score
	Rank     int32
}

func (e ScoredEntry) String() string {
	return fmt.Sprintf("rank=%d player=%d score=%s", e.Rank, e.PlayerID, e.Score)
}
