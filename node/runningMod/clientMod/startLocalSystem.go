package clientMod

import (
	"BlockChainSimulator/config"
	"BlockChainSimulator/node/nodeattr"
	"BlockChainSimulator/node/p2p"
	"BlockChainSimulator/node/runningMod/runningModInterface"
	"BlockChainSimulator/utils"
	"context"
	"os/exec"
	"strconv"
	"sync"
	"syscall"
)

// used by client node to run the whole blockchain system
type StartLocalSystemAuxiliaryMod struct {
	nodeAttr *nodeattr.NodeAttr // the attribute of the belonging node
	p2pMod   *p2p.P2PMod        // the p2p network module of the belonging node
}

func NewStartLocalSystemAuxiliaryMod(attr *nodeattr.NodeAttr, p2p *p2p.P2PMod) runningModInterface.RunningMod {
	sam := new(StartLocalSystemAuxiliaryMod)
	sam.nodeAttr = attr
	sam.p2pMod = p2p

	return sam
}

func (sam *StartLocalSystemAuxiliaryMod) RegisterHandlers() {

}

func (sam *StartLocalSystemAuxiliaryMod) Run(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()
	// start the system
	if config.IsDistributed {
		utils.LoggerInstance.Error("Auto start in distributed mode is not supported yet, maybe use some manual script to start the system :)")
		return
	}
	for i := 0; i < config.ShardNum; i++ {
		// start the shard
		for j := 0; j < config.NodeNum; j++ {
			// start the consensus node, no more need the client-related parameters
			cmdstr := "./BlockChainSimulator" +
				" -b " + strconv.Itoa(config.BlockSize) +
				" -S " + strconv.Itoa(config.ShardNum) + " -N " + strconv.Itoa(config.NodeNum) +
				" -s " + strconv.Itoa(i) + " -n " + strconv.Itoa(j) +
				" -m " + config.ConsensusMethod +
				" -l " + config.LogLevel + " -t " + config.TxType

			if config.IsDistributed {
				cmdstr += " -d "
			}

			utils.LoggerInstance.Debug("run cmd: %s", cmdstr)

			cmd := exec.Command("bash", "-c", cmdstr)

			// detach the process
			cmd.SysProcAttr = &syscall.SysProcAttr{
				Setsid: true,
			}

			err := cmd.Start()
			if err != nil {
				utils.LoggerInstance.Error("Error starting node %d in shard %d: %v", j, i, err)
			}

			utils.LoggerInstance.Info("Node %d in shard %d started", j, i)
		}
	}
}
