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
	"strings"
	"sync"
	"time"
)

// used by client node to run the whole blockchain system
type StartDistributedSystemAuxiliaryMod struct {
	nodeAttr *nodeattr.NodeAttr // the attribute of the belonging node
	p2pMod   *p2p.P2PMod        // the p2p network module of the belonging node
}

func NewStartDistributedSystemAuxiliaryMod(attr *nodeattr.NodeAttr, p2p *p2p.P2PMod) runningModInterface.RunningMod {
	sdm := new(StartDistributedSystemAuxiliaryMod)
	sdm.nodeAttr = attr
	sdm.p2pMod = p2p

	return sdm
}

func (sdm *StartDistributedSystemAuxiliaryMod) RegisterHandlers() {

}

func (sdm *StartDistributedSystemAuxiliaryMod) Run(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()

	// always start the main node of each shard
	// start the main node of each shard
	for i := 0; i < config.ShardNum; i++ {
		sdm.startNode(i, 0)
	}

	// sleep for a while to ensure the main nodes are up
	time.Sleep(200 * time.Millisecond)

	// start the system in distributed mode via SSH
	for i := 0; i < config.ShardNum; i++ {
		for j := 1; j < config.NodeNum; j++ {
			sdm.startNode(i, j)
		}
	}
}

func (sdm *StartDistributedSystemAuxiliaryMod) startNode(ShardID int, NodeID int) {
	// get IP address from IPMap
	ipPort, ok := config.IPMap[ShardID][NodeID]
	if !ok {
		utils.LoggerInstance.Error("IP not found for shard %d node %d", ShardID, NodeID)
		return
	}

	// extract IP from "IP:Port" format
	ip := strings.Split(ipPort, ":")[0]

	// build the command to run on remote node
	processName := config.ProcessName
	remoteCmd := "cd " + config.RemoteWorkDir + " && nohup ./" + processName +
		" -b " + strconv.Itoa(config.BlockSize) +
		" -S " + strconv.Itoa(config.ShardNum) + " -N " + strconv.Itoa(config.NodeNum) +
		" -s " + strconv.Itoa(ShardID) + " -n " + strconv.Itoa(NodeID) +
		" -m " + config.ConsensusMethod +
		" -l " + config.LogLevel + " -t " + config.TxType +
		" -r " + strconv.FormatFloat(config.MaliciousRatio, 'f', -3, 64) +
		" -R " + strconv.FormatFloat(config.ResilientRatio, 'f', -3, 64)
	// malicious node
	if float64(ShardID*config.NodeNum+NodeID) > (1-config.MaliciousRatio)*float64(config.ShardNum*config.NodeNum) {
		remoteCmd += " -M "
	}
	if config.ConnectRemoteDemo {
		remoteCmd += " -C "
	}

	// redirect output and run in background
	remoteCmd += " > /dev/null 2>&1 &"
	// build SSH command with key authentication
	sshCmd := "ssh -i " + config.SSHKeyPath +
		" -o StrictHostKeyChecking=no " +
		config.SSHUser + "@" + ip +
		" \"" + remoteCmd + "\""
	// for example: ssh -i /home/pjj/.ssh/id_ed25519 -o StrictHostKeyChecking=no pjj@127.0.0.1 "cd /home/pjj/Desktop/BlockChain/blockchain-simulator && nohup ./DroneSystem -b 500 -S 2 -N 4 -s 0 -n 1 -m RBE -l INFO -t UTXO -r 0 -R 0.5 > /dev/null 2>&1 &"
	utils.LoggerInstance.Debug("run ssh cmd: %s", sshCmd)
	cmd := exec.Command("bash", "-c", sshCmd)
	// start the SSH command without waiting for it to complete
	err := cmd.Start()
	if err != nil {
		utils.LoggerInstance.Error("Error starting node %d in shard %d on %s: %v", NodeID, ShardID, ip, err)
	} else {
		utils.LoggerInstance.Info("Node %d in shard %d started on %s", NodeID, ShardID, ip)
	}
}
