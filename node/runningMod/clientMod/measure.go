package clientMod

import (
	"BlockChainSimulator/config"
	"BlockChainSimulator/message"
	"BlockChainSimulator/node/nodeattr"
	"BlockChainSimulator/node/p2p"
	"BlockChainSimulator/node/runningMod/runningModInterface"
	"BlockChainSimulator/utils"
	"context"
	"os"
	"sync"
)

var _ runningModInterface.RunningMod = &measureMod{}

// just for test use, this mod sends Txs every 3 seconds
type measureMod struct {
	nodeAttr *nodeattr.NodeAttr
	p2pMod   *p2p.P2PMod

	measureAddons []MeasureAddon
}

// just for test use, this mod sends Txs every 3 seconds
func NewMeasureMod(attr *nodeattr.NodeAttr, p2p *p2p.P2PMod) runningModInterface.RunningMod {
	mm := new(measureMod)
	mm.nodeAttr = attr
	mm.p2pMod = p2p
	mm.measureAddons = make([]MeasureAddon, 0)

	return mm
}

func (mm *measureMod) RegisterHandlers() {
	mm.p2pMod.RegisterHandler(message.MsgReply, mm.HandleReply)
}

func (mm *measureMod) Run(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()
	// check if config.ResultPath exists
	if _, err := os.Stat(config.ResultPath); os.IsNotExist(err) {
		err := os.MkdirAll(config.ResultPath, os.ModePerm)
		if err != nil {
			utils.LoggerInstance.Error("Failed to create the result dir:%v", err)
			return
		}
	}

	// init the measureAddons
	for _, addonType := range config.MeasureMethod {
		measureMod, err := NewMeasureAddon(addonType, mm)
		if err != nil {
			utils.LoggerInstance.Error("Failed to create the measureMod:%v", err)
			return
		}
		utils.LoggerInstance.Info("Added the measureMod:%v", addonType)
		mm.measureAddons = append(mm.measureAddons, measureMod)
	}

	// the system running done, write the result to file
	<-ctx.Done()
	utils.LoggerInstance.Info("Stop the measureMod, writing the result")

	for _, addon := range mm.measureAddons {
		addon.WriteResult() // write the result to file and close the file
	}
}

func (mm *measureMod) HandleReply(msg *message.Message) {
	rep := &message.Reply{}
	err := utils.Decode(msg.Content, rep)
	if err != nil {
		utils.LoggerInstance.Error("Failed to decode the request")
		return
	}

	utils.LoggerInstance.Info("Received the reply from the shard %d", rep.Sid)
	for _, addon := range mm.measureAddons {
		addon.UpdateRecord(rep)
	}
}
