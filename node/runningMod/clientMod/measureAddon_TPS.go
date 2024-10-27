package clientMod

import (
	"BlockChainSimulator/config"
	"BlockChainSimulator/message"
	"BlockChainSimulator/structs"
	"BlockChainSimulator/utils"
	"encoding/csv"
	"fmt"
	"os"
	"sync"
	"time"
)

var _ MeasureAddon = &measureAddonTPS{}

type measureAddonTPS struct {
	TPS []int // TPS per second

	startTime time.Time

	fp *os.File
	mu sync.Mutex
}

func NewMeasureAddonTPS(measureMod *measureMod) MeasureAddon {
	fp, err := os.Create(config.ResultPath + "TPS.csv")
	if err != nil {
		utils.LoggerInstance.Error("Failed to create the file:%v", err)
		return nil
	}

	return &measureAddonTPS{
		TPS:       make([]int, 0),
		startTime: time.Now(),
		fp:        fp,
	}
}

func (mat *measureAddonTPS) UpdateRecord(rep *message.Reply) {
	req := rep.Req

	b := &structs.Block{}
	err := utils.Decode(req.Content, &b)
	if err != nil {
		utils.LoggerInstance.Error("Error in decoding the request content")
		return
	}

	txs := b.Transactions

	mat.mu.Lock()
	defer mat.mu.Unlock()

	// calculate the TPS
	elapsed := int(time.Since(mat.startTime).Seconds())

	// if the elapsed time is larger than the length of the TPS, extend the TPS
	if elapsed >= len(mat.TPS) {
		newRecords := make([]int, elapsed+1)
		copy(newRecords, mat.TPS)
		mat.TPS = newRecords
	}

	mat.TPS[elapsed] += len(txs)
}

func (mat *measureAddonTPS) WriteResult() {
	mat.mu.Lock()
	defer mat.mu.Unlock()

	timeElapsed := 0
	totalTxs := 0

	writer := csv.NewWriter(mat.fp)
	writer.Write([]string{"Time", "TPS", "TotalTxs", "AvgTPS"})
	for i, tps := range mat.TPS {
		if tps != 0 {
			totalTxs += tps
			timeElapsed += 1
		}
		avgTPS := float64(totalTxs) / float64(timeElapsed)
		writer.Write([]string{fmt.Sprintf("%d", i), fmt.Sprintf("%d", tps), fmt.Sprintf("%d", totalTxs), fmt.Sprintf("%.2f", avgTPS)})
	}
	writer.Flush()

	mat.fp.Close()
}
