package chaintoolkit

import (
	"fmt"
	"io"
	"strings"
	"text/tabwriter"
)

// BlockGapsFinder searches for all possible chains that blocks may form,
// a head, tail and length are found for each chain. It doesn't track all blocks
// along a chain, but keeps only head and tail, therefore not all forks and cycles
// can be spotted, also the current algorithm doesn't find duplications.
type BlockGapsFinder struct {
	// chains is a map of "PrevHash" to a chain, in other words a previous hash
	// that the chain's head points to (chain.head.PrevHash).
	chains    map[string]*Chain
	cntBlocks int64
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

	/*
		fmt.Println("\n\n>>> append")
		fmt.Println("---- input")
		for _, b := range blks {
			fmt.Printf("prev=%s hash=%s\n", b.PrevHash, b.BlockHash)
		}

		fmt.Println("\tstage 1")
		bg.Result().Print(os.Stdout)
	*/

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

	/*
		fmt.Println("\tstage 2")
		bg.Result().Print(os.Stdout)
	*/

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

	/*
		fmt.Println("\tstage 3")
		bg.Result().Print(os.Stdout)
	*/

	return nil
}

func (bg *BlockGapsFinder) Result() *BlockGapsResult {
	return &BlockGapsResult{bg: bg}
}

// Chains returns all found chains.
func (bgr *BlockGapsResult) Chains() map[string]*Chain {
	return bgr.bg.chains
}

// Print prints a debug info to w.
func (bgr *BlockGapsResult) Print(w io.Writer) {
	s := fmt.Sprintf("Counted\t%d\tblocks\t", bgr.bg.cntBlocks)

	tw := tabwriter.NewWriter(w, 1, 1, 1, ' ', 0)
	fmt.Fprintln(tw, s)
	fmt.Fprintf(tw, "Found\t%d\tchains\t\n", len(bgr.bg.chains))
	fmt.Fprintln(tw, strings.Repeat("-", len(s)-1))

	i := 0
	for _, c := range bgr.bg.chains {
		fmt.Fprintf(
			tw, "chain %d:\tlen %d\thead %s\ttail %s\t\n",
			i, c.Length, c.Head.BlockHash, c.Tail.BlockHash,
		)
		i++
	}

	tw.Flush()
}
