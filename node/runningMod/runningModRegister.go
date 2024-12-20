// Description: Contains the registry for the message handler module.
package runningMod

import (
	"BlockChainSimulator/node/nodeattr"
	"BlockChainSimulator/node/p2p"
	"BlockChainSimulator/node/runningMod/auxiliaryMod"
	"BlockChainSimulator/node/runningMod/clientMod"
	"BlockChainSimulator/node/runningMod/consensusMod/pbft"
	"BlockChainSimulator/node/runningMod/runningModInterface"
)

// Running mod relates to consensus
const (
	PBFTMod     string = "pbft"
	HotStuffMod string = "hotstuff"
	// add more consensus type here
)

// Running mod does not relate to consensus
const (
	ProposeTxsMod           string = "ProposeTxs"
	ProposeBlockMod         string = "ProposeBlock"
	ProposeBlock2ChannelMod string = "ProposeBlock2Channel"
)

// Running mod used by client
const (
	TestMod                 string = "test"
	StartSystemMod          string = "start"                // used by the client to start the system
	MeasureMod              string = "measure"              // used by the client to measure the performance of the system
	SendMimicContractTxsMod string = "sendMimicContractTxs" // used by the client to send mimic contract txs
)

var runnningModRegistry = make(map[string]func(attr *nodeattr.NodeAttr, p2p *p2p.P2PMod) runningModInterface.RunningMod)

func init() {
	// register more msgHandler type here

	// Consensus Running Mod
	runnningModRegistry[PBFTMod] = pbft.NewPbftCosensusMod

	// Auxiliary Running Mod
	runnningModRegistry[ProposeTxsMod] = auxiliaryMod.NewProposeTxsAuxiliaryMod
	runnningModRegistry[ProposeBlockMod] = auxiliaryMod.NewProposeBlockAuxiliaryMod
	runnningModRegistry[ProposeBlock2ChannelMod] = auxiliaryMod.NewProposeBlock2ChannelAuxiliaryMod

	// Client Running Mod
	runnningModRegistry[TestMod] = clientMod.NewTestAuxiliaryMod
	runnningModRegistry[MeasureMod] = clientMod.NewMeasureMod
	runnningModRegistry[StartSystemMod] = clientMod.NewStartSystemAuxiliaryMod
	runnningModRegistry[SendMimicContractTxsMod] = clientMod.NewSendMimicContractTxsMod
}

// invoke by node.go to create a running mod
func NewRunningMod(runningModType string, attr *nodeattr.NodeAttr, p2p *p2p.P2PMod) runningModInterface.RunningMod {
	if constructor, exists := runnningModRegistry[runningModType]; exists {
		return constructor(attr, p2p)
	}
	return nil
}
