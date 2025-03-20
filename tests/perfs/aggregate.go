package main

import (
	"context"
	"encoding/csv"
	"fmt"
	"log"
	"math/big"
	"os"
	"time"

	"github.com/ethereum/go-ethereum/core/types"
)

func aggregateBlockData() {
	if fromBlock <= 0 {
		log.Fatalf("missing -from flag")
	}
	if toBlock <= 0 {
		log.Fatalf("missing -to flag")
	}

	if fromBlock > toBlock {
		log.Fatalf("-from must be < than -to")
	}

	client := getClient()
	blocks := []*types.Block{}
	for ; fromBlock <= toBlock; fromBlock++ {
		block, err := client.BlockByNumber(context.Background(), big.NewInt(int64(fromBlock)))
		if err != nil {
			log.Fatalf("Failed to get latest block: %v", err)
		}

		blocks = append(blocks, block)
	}

	records := [][]string{
		{"height", "timestamp", "timeTaken", "size (bytes)", "transactionsCount", "gasLimit", "gasUsed", "baseFeePerGas"},
	}

	for i, block := range blocks {
		var timeTaken time.Duration

		if i+1 < len(blocks) {
			this := time.Unix(int64(blocks[i].Time()), 0)
			next := time.Unix(int64(blocks[i+1].Time()), 0)
			timeTaken = next.Sub(this)
		}

		records = append(records,
			[]string{
				fmt.Sprintf("%v", block.NumberU64()),
				fmt.Sprintf("%v", block.Time()),
				timeTaken.String(),
				fmt.Sprintf("%v", block.Size()),
				fmt.Sprintf("%v", block.Transactions().Len()),
				fmt.Sprintf("%v", block.GasLimit()),
				fmt.Sprintf("%v", block.GasUsed()),
				block.BaseFee().String(),
			},
		)
	}

	f, err := os.Create(output)
	if err != nil {
		log.Fatalf("couldn't open output file: %v", err)
	}
	defer f.Close()

	w := csv.NewWriter(f)
	w.WriteAll(records)

	if err := w.Error(); err != nil {
		log.Fatalf("error writing csv: %v", err)
	}
}
