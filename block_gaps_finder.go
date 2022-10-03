package chaintoolkit

import (
	"fmt"
	"io"
	"sort"
	"strings"
	"text/tabwriter"
	"time"
)

// BlockGapsFinder searches for all possible chains that blocks may form,
// a head, tail and length are found for each chain. It doesn't track all blocks
// along a chain, but keeps only head and tail, therefore not all forks and cycles
// can be spotted, also the current algorithm doesn't find duplicates.
type BlockGapsFinder struct {
	// chains is a map of "PrevHash" to a chain, in other words a previous hash
	// that the chain's head points to (chain.head.PrevHash).
	chains    map[string]*Chain
	cntBlocks int64

	took time.Duration
}

// Chain represents an info of a single chain.
type Chain struct {
	Head   *Block
	Tail   *Block
	Length int
}

// Block represents a link in the chain with a block hash and previous hash.
type Block struct {
	// BlockHash is a current block hash.
	BlockHash string

	// PrevHash is a hash that points to the previouse block.
	PrevHash string

	// Time is time of a block.
	Time time.Time
}

// BlogGapsResult is a result returned after inspecting blocks,
// it contains info about chains that blocks organized.
type BlockGapsResult struct {
	bg *BlockGapsFinder
}

// NewBlockGapsFinder returns an instance of *BlockGapsFinder.
func NewBlockGapsFinder() *BlockGapsFinder {
	return &BlockGapsFinder{chains: map[string]*Chain{}}
}

// Append inpects given blocks and appends info to already inspected.
// It helps to reduce resource usage during debugging large amount of blocks.
func (bg *BlockGapsFinder) Append(blks ...*Block) error {
	if len(blks) == 0 {
		return nil
	}

	bg.cntBlocks += int64(len(blks))
	start := time.Now()
	defer func() { bg.took += time.Since(start) }()

	for i := 0; i < len(blks); i++ {
		b := blks[i]

		chained := false

		// try to chain a block if there are already chains
		if ch, ok := bg.chains[b.BlockHash]; ok {

			if b.BlockHash == ch.Tail.BlockHash {
				return fmt.Errorf(
					"a cycled chain found: head: %v, tail %v", b, ch.Tail,
				)
			}

			ch.Head = b
			ch.Length++
			chained = true
			bg.chains[b.PrevHash] = ch
			delete(bg.chains, b.BlockHash)
		} else {
			for _, ch := range bg.chains {
				if ch.Tail.BlockHash == b.PrevHash {

					if b.BlockHash == ch.Head.PrevHash {
						return fmt.Errorf(
							"a cycled chain found: head: %v, tail %v", ch.Head, b,
						)
					}

					ch.Tail = b
					ch.Length++
					chained = true
					break
				}
			}
		}

		if chained {
			continue
		}

		// otherwise try create a new chain from the block
		if ch, ok := bg.chains[b.PrevHash]; ok {
			if ch.Head.BlockHash != b.BlockHash {
				return fmt.Errorf("found a fork block %s", b.BlockHash)
			}
		} else {
			bg.chains[b.PrevHash] = &Chain{Length: 1, Head: b, Tail: b}
		}
	}

	// try merge chains
	for prevHash, ch := range bg.chains {
		for {

			if prevHash == ch.Tail.BlockHash {
				return fmt.Errorf(
					"a cycled chain found: head: %v, tail %v", ch.Head, ch.Tail,
				)
			}

			ch2, ok := bg.chains[ch.Tail.BlockHash]

			if !ok {
				break
			}

			delete(bg.chains, ch.Tail.BlockHash)

			ch.Tail = ch2.Tail
			ch.Length += ch2.Length
		}
	}

	return nil
}

func (bg *BlockGapsFinder) Result() *BlockGapsResult {
	return &BlockGapsResult{bg: bg}
}

// Chains returns all found chains.
func (bgr *BlockGapsResult) Chains() map[string]*Chain {
	return bgr.bg.chains
}

// ChainsAsArray returns all found chains, it converts a map of chains to an array.
func (bgr *BlockGapsResult) ChainsAsArray() []*Chain {
	ret := []*Chain{}
	for _, c := range bgr.bg.chains {
		ret = append(ret, c)
	}
	return ret
}

// Print prints a debug info to w.
func (bgr *BlockGapsResult) Print(w io.Writer) {

	tw := tabwriter.NewWriter(w, 1, 1, 1, ' ', 0)
	var longest *Chain
	timeLayout := "2006-01-02 15:04:05"
	fmt.Fprintln(
		tw, "chain\tblocks\thead\thead time\ttail\ttail time\t",
	)

	chains := []*Chain{}
	for _, c := range bgr.bg.chains {
		chains = append(chains, c)
	}

	sort.Slice(chains, func(i, j int) bool {
		return chains[i].Head.Time.Before(chains[j].Head.Time)
	})

	i := 0
	for _, c := range chains {
		if longest == nil || longest.Length < c.Length {
			longest = c
		}

		fmt.Fprintf(
			tw, "%d\t%d\t%s\t%s\t%s\t%s\t\n",
			i, c.Length, c.Head.BlockHash, c.Head.Time.Format(timeLayout),
			c.Tail.BlockHash, c.Tail.Time.Format(timeLayout),
		)
		i++
	}

	s := fmt.Sprintf(
		"Longest:\t%d blocks\thead %s\ttail %s\t",
		longest.Length, longest.Head.BlockHash, longest.Tail.BlockHash,
	)
	fmt.Fprintln(tw, strings.Repeat("-", len(s)))
	fmt.Fprintln(tw, s)
	fmt.Fprintf(
		tw, " \t \thead time %s\ttail time %s\t\n",
		longest.Head.Time.Format(timeLayout), longest.Tail.Time.Format(timeLayout),
	)
	fmt.Fprintf(tw, "Total:\t%d blocks\t\n", bgr.bg.cntBlocks)
	fmt.Fprintf(tw, "Found:\t%d chains\t\n", len(bgr.bg.chains))
	fmt.Fprintf(tw, "Took:\t%s\t\t\n", bgr.bg.took)

	tw.Flush()
}
