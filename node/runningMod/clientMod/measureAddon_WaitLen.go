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

var _ MeasureAddon = &measureAddonWaitLen{}

type measureAddonWaitLen struct {
	WaitLen map[int]int // shardID -> Wait Request Len

	fp *os.File
	mu sync.Mutex
}

func NewMeasureAddonWaitLen(measureMod *measureMod) MeasureAddon {
	fp, err := os.Create(config.ResultPath + "WaitLen.csv")
	if err != nil {
		utils.LoggerInstance.Error("Failed to create the file:%v", err)
		return nil
	}

	return &measureAddonWaitLen{
		WaitLen: make(map[int]int),
		fp:      fp,
	}
}

func (mawl *measureAddonWaitLen) UpdateRecord(rep *message.Reply) {
	mawl.mu.Lock()
	defer mawl.mu.Unlock()

	mawl.WaitLen[rep.Sid] = rep.ReqQueueLen
}

func (mawl *measureAddonWaitLen) WriteResult() {
	mawl.mu.Lock()
	defer mawl.mu.Unlock()

	writer := csv.NewWriter(mawl.fp)
	title := []string{}
	result := []string{}
	for sid, waitlen := range mawl.WaitLen {
		title = append(title, fmt.Sprintf("Shard%d", sid))
		result = append(result, fmt.Sprintf("%d", waitlen))
	}

	writer.Write(title)
	writer.Flush()
	writer.Write(result)
	writer.Flush()

	mawl.fp.Close()
}
