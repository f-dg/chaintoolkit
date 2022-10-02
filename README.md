# chaintoolkit

a simple toolkit to instrument blockchain blocks (essentially linked lists).

### BlockGapsFinder
The finder allows to spot gaps in blocks, it finds all possible chains that blocks may form and returns result that can be inspected by code or printed in a human readable table format. The algorithm only tracks
`head` and `tail` of a chain but not blocks in between. Therefore it is possible to feed blocks by batches and save resource usage. It also has drawbacks such as not all possible fork blocks nor cycled chains can be spotted, the algorithm is sensitive to duplicates, the calling code must make sure to not provide them. For instance:
```
SELECT COUNT(*), block_hash FROM blocks GROUP BY block_hash HAVING COUNT(*) > 1;
SELECT COUNT(*), prev_hash FROM blocks GROUP BY prev_hash HAVING COUNT(*) > 1;
```
Keep in mind that `LIMIT OFFSET` pagination consumes a lot of memory if there huge amounts of blocks, such a pagination also gives duplicates without sorting. The better way to paginate blocks is by `height`, since most of blockchains use height of a block as its id that increments sequentially.
#### code example
```golang
blks := []*chaintoolkit.Block{
  &chaintoolkit.Block{BlockHash: "h1", PrevHash: "h0", Time: time.Now()},
  &chaintoolkit.Block{BlockHash: "h2", PrevHash: "h1", Time: time.Now()},
}

blks2 := []*chaintoolkit.Block{
  &chaintoolkit.Block{BlockHash: "b1", PrevHash: "b0", Time: time.Now()},
  &chaintoolkit.Block{BlockHash: "h3", PrevHash: "h2", Time: time.Now()},
}

bg := chaintoolkit.NewBlockGapsFinder()

if err := bg.Append(blks...); err != nil {
	log.Fatal(err)
}

if err := bg.Append(blks2...); err != nil {
	log.Fatal(err)
}

res := bg.Result()

chains := res.Chains()

for _, c := range chains {
  fmt.Println(c.Length, c.Head, c.Tail)
}

res.Print(os.Stdout)
```

### ForkBlocksFinder
not implemented


### CycledChainsFinder
not implemented
