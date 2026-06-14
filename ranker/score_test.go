package ranker

import (
	"testing"
)

func TestScoreCompare_PrimaryScore(t *testing.T) {
	high := Score{Primary: 100}
	low := Score{Primary: 50}
	if high.Compare(low) != -1 {
		t.Fatal("higher primary should rank first")
	}
	if low.Compare(high) != 1 {
		t.Fatal("lower primary should rank second")
	}
}

func TestScoreCompare_SecondaryScore(t *testing.T) {
	a := Score{Primary: 100, Secondary: 200}
	b := Score{Primary: 100, Secondary: 100}
	if a.Compare(b) != -1 {
		t.Fatal("higher secondary should rank first when primary equal")
	}
}

func TestScoreCompare_TimeTiebreak(t *testing.T) {
	early := Score{Primary: 100, Secondary: 100, Arrival: 1000}
	late := Score{Primary: 100, Secondary: 100, Arrival: 2000}
	if early.Compare(late) != -1 {
		t.Fatal("earlier arrival should rank first when both scores equal")
	}
}

func TestScoreCompare_Equal(t *testing.T) {
	a := Score{Primary: 100, Secondary: 50, Arrival: 1000}
	b := Score{Primary: 100, Secondary: 50, Arrival: 1000}
	if a.Compare(b) != 0 {
		t.Fatal("identical scores should be equal")
	}
}
