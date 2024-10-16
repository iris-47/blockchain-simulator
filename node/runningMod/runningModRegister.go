// Description: Contains the registry for the message handler module.
package runningMod

import (
	"BlockChainSimulator/node/nodeattr"
	"BlockChainSimulator/node/p2p"
	"BlockChainSimulator/node/runningMod/auxiliaryMod"
	"BlockChainSimulator/node/runningMod/consensusMod/pbft"
	"BlockChainSimulator/node/runningMod/runningModInterface"
)

// Running mod relates to consensus
const (
	PBFT     string = "pbft"
	HotStuff string = "hotstuff"
	// add more consensus type here
)

// Running mod does not relate to consensus
const (
	TestMod       string = "test"
	ProposeTxsMod string = "ProposeTxs"

	MeasureMod string = "measure" // used by the client to measure the performance of the system
)

var runnningModRegistry = make(map[string]func(attr *nodeattr.NodeAttr, p2p *p2p.P2PMod) runningModInterface.RunningMod)

func init() {
	// register more msgHandler type here

	// Consensus Running Mod
	runnningModRegistry[PBFT] = pbft.NewPbftCosensusMod

	// Auxiliary Running Mod
	runnningModRegistry[TestMod] = auxiliaryMod.NewTestAuxiliaryMod
	runnningModRegistry[ProposeTxsMod] = auxiliaryMod.NewProposeTxsAuxiliaryMod
}

// invoke by node.go to create a running mod
func NewRunningMod(runningModType string, attr *nodeattr.NodeAttr, p2p *p2p.P2PMod) runningModInterface.RunningMod {
	if constructor, exists := runnningModRegistry[runningModType]; exists {
		return constructor(attr, p2p)
	}
	return nil
}
