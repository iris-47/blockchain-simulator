package node

import (
	"BlockChainSimulator/config"
	"BlockChainSimulator/message"
	"BlockChainSimulator/node/nodeattr"
	"BlockChainSimulator/node/p2p"
	"BlockChainSimulator/node/runningMod"
	"BlockChainSimulator/node/runningMod/runningModInterface"
	"BlockChainSimulator/utils"
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

type Node struct {
	Attr   *nodeattr.NodeAttr // the base attribute of the node
	P2PMod *p2p.P2PMod        // the p2p network module

	// running modules
	RunningMods []runningModInterface.RunningMod
}

func NewNode(sid int, nid int, pcc *config.ChainConfig, runningModTypes []string) (*Node, error) {
	var err error
	node := new(Node)
	node.Attr = nodeattr.NewNodeAttr(sid, nid, pcc)
	node.P2PMod = p2p.NewP2PMod(node.Attr.Ipaddr)

	for _, runningModType := range runningModTypes {
		runningMod := runningMod.NewRunningMod(runningModType, node.Attr, node.P2PMod)
		if runningMod == nil {
			utils.LoggerInstance.Error("Error creating running module: %v", runningModType)
			return nil, err
		}
		node.RunningMods = append(node.RunningMods, runningMod)
	}

	// register the messgae process handlers to the p2p module
	for _, runningMod := range node.RunningMods {
		runningMod.RegisterHandlers()
	}

	return node, nil
}

func (n *Node) Run() {
	wg := sync.WaitGroup{}
	ctx, cancel := context.WithCancel(context.Background())

	n.P2PMod.MsgHandlerMap[message.MsgStop] = func(msg *message.Message) {
		utils.LoggerInstance.Info("Received stop message...now close the running Mod")
		cancel()
	}

	// start to receive the message
	n.P2PMod.StartListen()

	// start to run custom modules
	for _, runningMod := range n.RunningMods {
		wg.Add(1)
		go runningMod.Run(ctx, &wg)
	}

	// ctrl+c to stop all the goroutines
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-ctx.Done(): // invoke by other nodes, for example, the client

	case <-sigChan:
		utils.LoggerInstance.Info("Node stopped by system interrupt...now close the running Mod")
		cancel()
	}

	// wait for all the runningMods to stop
	wg.Wait()
	utils.LoggerInstance.Info("Node stopped...")
}
