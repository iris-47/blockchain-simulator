package clientMod

import (
	"BlockChainSimulator/config"
	"BlockChainSimulator/message"
	"BlockChainSimulator/utils"
	"encoding/csv"
	"fmt"
	"os"
	"sync"
)

var _ MeasureAddon = &measureAddonTCL{}

type measureAddonTCL struct {
	latencys []float64 // TCL of each Request

	mu sync.Mutex
	fp *os.File
}

func NewMeasureAddonTCL(measureMod *measureMod) MeasureAddon {
	fp, err := os.Create(config.ResultPath + "TCL.csv")
	if err != nil {
		utils.LoggerInstance.Error("Failed to create the file:%v", err)
		return nil
	}
	return &measureAddonTCL{
		latencys: make([]float64, 0),
		fp:       fp,
	}
}

func (mat *measureAddonTCL) UpdateRecord(rep *message.Reply) {
	req := rep.Req

	mat.mu.Lock()
	defer mat.mu.Unlock()

	// rep.Time - req.ReqTime
	latency := rep.Time.Sub(req.ReqTime).Seconds()
	mat.latencys = append(mat.latencys, latency)
}

func (mat *measureAddonTCL) WriteResult() {
	mat.mu.Lock()
	defer mat.mu.Unlock()

	avgLatency := 0.0
	for _, latency := range mat.latencys {
		avgLatency += latency / float64(len(mat.latencys))
	}
	// write the result to file
	writer := csv.NewWriter(mat.fp)
	writer.Write([]string{"TCL", "AvgTCL"})
	for _, record := range mat.latencys {
		writer.Write([]string{fmt.Sprintf("%.2f", record), fmt.Sprintf("%.2f", avgLatency)})
	}
	writer.Flush()

	// close the file pointer
	mat.fp.Close()
}
