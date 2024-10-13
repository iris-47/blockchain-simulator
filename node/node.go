package node

import (
	"BlockChainSimulator/config"
	"BlockChainSimulator/message"
	"BlockChainSimulator/node/msgHandler"
	"BlockChainSimulator/node/msgHandler/msgHandlerInterface"
	"BlockChainSimulator/node/nodeattr"
	"BlockChainSimulator/node/p2p"
	"BlockChainSimulator/utils"
	"context"
	"os"
	"os/signal"
	"syscall"
)

type Node struct {
	Attr   *nodeattr.NodeAttr // the base attribute of the node
	P2PMod *p2p.P2PMod        // the p2p network module

	// message handler modules
	ConsensusMod msgHandlerInterface.MsgHandlerMod // message handler related to consensus, for example, pbft, hotstuff, etc.
	AuxiliaryMod msgHandlerInterface.MsgHandlerMod // message handler does not related to consensus, for example, read the Txs dataset, etc.
}

func NewNode(sid int, nid int, pcc *config.ChainConfig, consensusType msgHandler.ConsensusHandlerType, consensusAddonType string, auxiliaryType msgHandler.AuxiliaryHandlerType) (*Node, error) {
	var err error
	node := new(Node)
	node.Attr = nodeattr.NewNodeAttr(sid, nid, pcc)
	node.P2PMod = p2p.NewP2PMod(node.Attr.Ipaddr)

	node.ConsensusMod, err = msgHandler.NewConsensusMod(consensusType, consensusAddonType, node.Attr, node.P2PMod)
	if err != nil {
		utils.LoggerInstance.Error("Error creating consensus module: %v", err)
		return nil, err
	}

	node.AuxiliaryMod, err = msgHandler.NewAuxiliaryMod(auxiliaryType, node.Attr, node.P2PMod)
	if err != nil {
		utils.LoggerInstance.Error("Error creating auxiliary module: %v", err)
		return nil, err
	}

	// register the messgae process handlers to the p2p module
	node.ConsensusMod.RegisterHandlers()
	node.AuxiliaryMod.RegisterHandlers()

	return node, nil
}

func (n *Node) Run() {
	ctx, cancel := context.WithCancel(context.Background())

	n.P2PMod.MsgHandlerMap[message.MsgStop] = func(msg *message.Message) {
		utils.LoggerInstance.Info("Received stop message. Stopping the node...")
		cancel()
	}

	// start to receive the message
	n.P2PMod.StartListen()

	// start to run custom modules
	// no more goroutine inside the MsgHandlerMod Run()
	go n.ConsensusMod.Run()
	go n.AuxiliaryMod.Run()

	// ctrl+c to stop all the goroutines
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-ctx.Done():
		utils.LoggerInstance.Info("Node stopped") // invoke by other nodes, for example, the client
	case <-sigChan:
		utils.LoggerInstance.Info("Received system interrupt. Shutting down the system...")
	}
}
