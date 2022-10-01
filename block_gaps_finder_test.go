package chaintoolkit

import (
	"math/rand"
	"testing"
	"time"
)

func TestBlockGapsAppend(t *testing.T) {
	firstChainPrevHash := "h-1"
	secondChainPrevHash := "b0"
	thirdChainPrevHash := "c0"

	blks := []*Block{
		&Block{BlockHash: "h1", PrevHash: "h0"},
		&Block{BlockHash: "h2", PrevHash: "h1"},
		&Block{BlockHash: "h3", PrevHash: "h2"},
		&Block{BlockHash: "h4", PrevHash: "h3"},
		&Block{BlockHash: "h5", PrevHash: "h4"},
		&Block{BlockHash: "b1", PrevHash: secondChainPrevHash},
		&Block{BlockHash: "b2", PrevHash: "b1"},
		&Block{BlockHash: "h7", PrevHash: "h6"},
		&Block{BlockHash: "h8", PrevHash: "h7"},
		&Block{BlockHash: "c1", PrevHash: thirdChainPrevHash},
	}

	blks2 := []*Block{
		&Block{BlockHash: "h6", PrevHash: "h5"},
		&Block{BlockHash: "h9", PrevHash: "h8"},
		&Block{BlockHash: "h10", PrevHash: "h9"},
		&Block{BlockHash: "h0", PrevHash: firstChainPrevHash},
		&Block{BlockHash: "c2", PrevHash: "c1"},
		&Block{BlockHash: "c3", PrevHash: "c2"},
		&Block{BlockHash: "h12", PrevHash: "h11"},
	}

	blks3 := []*Block{
		&Block{BlockHash: "h11", PrevHash: "h10"},
	}

	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(blks), func(i, j int) { blks[i], blks[j] = blks[j], blks[i] })
	rand.Shuffle(len(blks2), func(i, j int) { blks2[i], blks2[j] = blks2[j], blks2[i] })

	bg := NewBlockGapsFinder()

	if err := bg.Append(blks...); err != nil {
		t.Fatal(err)
	}

	if err := bg.Append(blks2...); err != nil {
		t.Fatal(err)
	}

	if err := bg.Append(blks3...); err != nil {
		t.Fatal(err)
	}

	res := bg.Result()

	chains := res.Chains()
	if len(chains) != 3 {
		t.Fatalf("want 3 chains, got %d", len(chains))
	}

	first := chains[firstChainPrevHash]

	if first.Length != 13 || first.Head.BlockHash != "h0" ||
		first.Tail.BlockHash != "h12" {
		t.Errorf(
			"wrong chain: wants len=13 head=h0 tail=h12, got len=%d head=%s tail=%s",
			first.Length, first.Head.BlockHash, first.Tail.BlockHash,
		)
	}

	second := chains[secondChainPrevHash]

	if second.Length != 2 || second.Head.BlockHash != "b1" ||
		second.Tail.BlockHash != "b2" {
		t.Errorf(
			"wrong chain: wants len=2 head=b1 tail=b2, got len=%d head=%s tail=%s",
			second.Length, second.Head.BlockHash, second.Tail.BlockHash,
		)
	}

	third := chains[thirdChainPrevHash]

	if third.Length != 3 || third.Head.BlockHash != "c1" || third.Tail.BlockHash != "c3" {
		t.Errorf(
			"wrong chain: wants len=3 head=c1 tail=c3, got len=%d head=%s tail=%s",
			third.Length, third.Head.BlockHash, third.Tail.BlockHash,
		)
	}
}

func TestBlockGapsAppendErrors(t *testing.T) {

	t.Run("cycle", func(t *testing.T) {
		blks := []*Block{
			&Block{BlockHash: "h1", PrevHash: "h4"}, // h1 refers to h4 (cycle)
			&Block{BlockHash: "h2", PrevHash: "h1"},
			&Block{BlockHash: "b1", PrevHash: "b0"},
		}

		blks2 := []*Block{
			&Block{BlockHash: "h3", PrevHash: "h2"},
			&Block{BlockHash: "h4", PrevHash: "h3"},
			&Block{BlockHash: "b2", PrevHash: "b1"},
		}

		cd := NewBlockGapsFinder()

		if err := cd.Append(blks...); err != nil {
			t.Fatalf("want no error, got %v", err)
		}

		if err := cd.Append(blks2...); err == nil {
			t.Fatal("want a cycle error, got nil")
		}
	})

	t.Run("fork_1", func(t *testing.T) {
		blks := []*Block{
			&Block{BlockHash: "h1", PrevHash: "h0"},
			&Block{BlockHash: "h2", PrevHash: "h0"},
		}

		cd := NewBlockGapsFinder()

		if err := cd.Append(blks...); err == nil {
			t.Fatal("want a fork error, got nil")
		}
	})

	t.Run("fork_2", func(t *testing.T) {
		blks := []*Block{
			&Block{BlockHash: "h1", PrevHash: "h0"},
			&Block{BlockHash: "h2", PrevHash: "h1"},
		}

		blks2 := []*Block{
			&Block{BlockHash: "b1", PrevHash: "h0"},
		}

		cd := NewBlockGapsFinder()

		if err := cd.Append(blks...); err != nil {
			t.Fatalf("want no error, got %v", err)
		}

		if err := cd.Append(blks2...); err == nil {
			t.Fatal("want a fork error, got nil")
		}
	})
}
