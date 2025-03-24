package main

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/big"
	"net/http"
	"os"
	"time"

	"github.com/ethereum/go-ethereum/core/types"
)

var (
	tendermintEndpoint = func(baseURL string) string {
		return fmt.Sprintf("%v/num_unconfirmed_txs", baseURL)
	}

	totalTxsSentAtBlock = []int64{}
)

type UnconfirmedTxs struct {
	Result struct {
		NTxs  string `json:"n_txs"`
		Total string `json:"total"`
	} `json:"result"`
}

func getMempoolSize() *UnconfirmedTxs {
	resp, err := http.Get(tendermintEndpoint(tendermintRPC))
	if err != nil {
		log.Printf("couldn't reach tendermint RPC: %v", err)
		return nil
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("couldn't read tendermint RPC response body: %v", err)
	}

	txs := UnconfirmedTxs{}
	err = json.Unmarshal(body, &txs)
	if err != nil {
		log.Printf("couldn't unmarshal tendermint RPC response: %v", err)
	}

	return &txs
}

func aggregateBlockDataParallel() {
	client := getClient()

	latestBlock, err := client.BlockByNumber(context.Background(), nil)
	if err != nil {
		log.Fatalf("Failed to get latest block: %v", err)
	}
	totalTxsSentAtBlock = append(totalTxsSentAtBlock, 0)

	backlog := UnconfirmedTxs{}
	backlog.Result.NTxs = "0"
	backlog.Result.Total = "0"
	blocks := []BlockAndBacklog{{Block: latestBlock, Backlog: &backlog}}
	for {
		// wait a little
		time.Sleep(1 * time.Second)
		newBlock, err := client.BlockByNumber(context.Background(), big.NewInt(0).Add(latestBlock.Number(), big.NewInt(1)))
		if err != nil {
			log.Printf("no new block yet")
			continue
		}

		backlog := getMempoolSize()
		if backlog == nil {
			backlog = &UnconfirmedTxs{}
		}

		log.Printf("new block %v found", newBlock.Number().String())
		log.Printf("current backlog: unconfirmed(%v), total(%v)", backlog.Result.NTxs, backlog.Result.Total)

		latestBlock = newBlock
		blocks = append(blocks, BlockAndBacklog{newBlock, backlog})
		writeBlocks(blocks)
	}
}

type BlockAndBacklog struct {
	Block   *types.Block
	Backlog *UnconfirmedTxs
}

func writeBlocks(blocks []BlockAndBacklog) {
	records := [][]string{
		{"blockIndex", "height", "timestamp", "timeTaken", "size (bytes)", "transactionsCount", "transactionBacklog", "gasLimit", "gasUsed", "baseFeePerGas", "totalTxSent", "totalConfirmedTx"},
	}

	totalTxsSentAtBlock = append(totalTxsSentAtBlock, ongoingTotalTxs.Load())

	totalInBlocks := []int{}
	ongoing := 0
	for _, b := range blocks {
		ongoing += b.Block.Transactions().Len()
		totalInBlocks = append(totalInBlocks, ongoing)
	}

	for i, block := range blocks {
		var timeTaken time.Duration

		if i+1 < len(blocks) {
			this := time.Unix(int64(blocks[i].Block.Time()), 0)
			next := time.Unix(int64(blocks[i+1].Block.Time()), 0)
			timeTaken = next.Sub(this)
		}

		records = append(records,
			[]string{
				fmt.Sprintf("%v", i),
				fmt.Sprintf("%v", block.Block.NumberU64()),
				fmt.Sprintf("%v", block.Block.Time()),
				timeTaken.String(),
				fmt.Sprintf("%v", block.Block.Size()),
				fmt.Sprintf("%v", block.Block.Transactions().Len()),
				fmt.Sprintf("%v", block.Backlog.Result.NTxs),
				fmt.Sprintf("%v", block.Block.GasLimit()),
				fmt.Sprintf("%v", block.Block.GasUsed()),
				block.Block.BaseFee().String(),
				fmt.Sprintf("%v", totalTxsSentAtBlock[i]),
				fmt.Sprintf("%v", totalInBlocks[i]),
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
