package pbft

import (
	"BlockChainSimulator/message"
)

var _ PbftAddon = &PbftSimpleAddon{}

// a simple PBFT addon, which implemen
type PbftSimpleAddon struct {
	pbftMod *PbftCosensusMod // the belonging pbft module
}

func NewSimplePbftCosensusAddon(pbftMod *PbftCosensusMod) PbftAddon {
	return &PbftSimpleAddon{
		pbftMod: pbftMod,
	}
}

func (addon *PbftSimpleAddon) HandleProposeAddon(req *message.Request) bool {
	return true
}

func (addon *PbftSimpleAddon) HandlePrePrepareAddon(req *message.Request) bool {
	return true
}

func (addon *PbftSimpleAddon) HandlePrepareAddon(req *message.Request) bool {
	return true
}

func (addon *PbftSimpleAddon) HandleCommitAddon(req *message.Request) bool {
	return true
}
