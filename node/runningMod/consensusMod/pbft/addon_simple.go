package pbft

import (
	"BlockChainSimulator/message"
	"BlockChainSimulator/node/nodeattr"
)

// a simple PBFT addon, which implemen
type PbftSimpleAddon struct {
	nodeAttr *nodeattr.NodeAttr
}

func NewSimplePbftCosensusAddon(attr *nodeattr.NodeAttr) PbftAddon {
	return &PbftSimpleAddon{nodeAttr: attr}
}

func (addon *PbftSimpleAddon) HandleProposeAddon(msg *message.Message) bool {
	return true
}

func (addon *PbftSimpleAddon) HandlePrePrepareAddon(msg *message.Message) bool {
	return true
}

func (addon *PbftSimpleAddon) HandlePrepareAddon(msg *message.Message) bool {
	return true
}

func (addon *PbftSimpleAddon) HandleCommitAddon(msg *message.Message) bool {
	return true
}
