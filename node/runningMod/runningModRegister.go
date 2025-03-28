// Description: Contains the registry for the message handler module.
package runningMod

import (
	"BlockChainSimulator/node/nodeattr"
	"BlockChainSimulator/node/p2p"
	"BlockChainSimulator/node/runningMod/clientMod"
	"BlockChainSimulator/node/runningMod/consensusMod"
	"BlockChainSimulator/node/runningMod/consensusMod/ds"
	"BlockChainSimulator/node/runningMod/consensusMod/pbft"
	"BlockChainSimulator/node/runningMod/consensusMod/tbb"
	"BlockChainSimulator/node/runningMod/runningModInterface"
)

// Running mod relates to consensus
const (
	PBFTMod     string = "pbft"
	HotStuffMod string = "hotstuff"

	// add more consensus type here
	TBBMod string = "tbb"
	DSMod  string = "dolev-strong"
)

// Running mod does not relate to consensus
const (
	ProposeTxsMod    string = "ProposeTxs"
	ProposeBlockMod  string = "ProposeBlock"
	ProposeStringMod string = "ProposeString"
)

// Running mod used by client
const (
	TestMod        string = "test"
	StartSystemMod string = "start"   // used by the client to start the system
	MeasureMod     string = "measure" // used by the client to measure the performance of the system
	QueryMod       string = "query"   // used by the client to query the consensus result

	// add more send type here
	SendMimicContractTxsMod string = "sendMimicContractTxs" // used by the client to send mimic contract txs
	SendStringManualMod     string = "sendStringManual"     // used by the client to send a string manually
)

var runningModRegistry = make(map[string]func(attr *nodeattr.NodeAttr, p2p *p2p.P2PMod) runningModInterface.RunningMod)

func init() {
	// register more msgHandler type here

	// Consensus Running Mod
	runningModRegistry[PBFTMod] = pbft.NewPbftCosensusMod
	runningModRegistry[TBBMod] = tbb.NewTBBCosensusMod
	runningModRegistry[DSMod] = ds.NewDSCosensusMod

	runningModRegistry[ProposeTxsMod] = consensusMod.NewProposeTxsAuxiliaryMod
	runningModRegistry[ProposeBlockMod] = consensusMod.NewProposeBlockAuxiliaryMod
	runningModRegistry[ProposeStringMod] = consensusMod.NewProposeStringAuxiliaryMod

	// Auxiliary Running Mod

	// Client Running Mod
	runningModRegistry[TestMod] = clientMod.NewTestAuxiliaryMod
	runningModRegistry[MeasureMod] = clientMod.NewMeasureMod
	runningModRegistry[QueryMod] = clientMod.NewQueryMod
	runningModRegistry[StartSystemMod] = clientMod.NewStartSystemAuxiliaryMod
	runningModRegistry[SendMimicContractTxsMod] = clientMod.NewSendMimicContractTxsMod
	runningModRegistry[SendStringManualMod] = clientMod.NewSendStringManualMod
}

// invoke by node.go to create a running mod
func NewRunningMod(runningModType string, attr *nodeattr.NodeAttr, p2p *p2p.P2PMod) runningModInterface.RunningMod {
	if constructor, exists := runningModRegistry[runningModType]; exists {
		return constructor(attr, p2p)
	}
	return nil
}
