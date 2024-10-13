// Description: Contains the registry for the message handler module.
package msgHandler

import (
	auxiliaryhandler "BlockChainSimulator/node/msgHandler/auxiliaryHandler"
	"BlockChainSimulator/node/msgHandler/consensusHandler/pbft"
	"BlockChainSimulator/node/msgHandler/msgHandlerInterface"
	"BlockChainSimulator/node/nodeattr"
	"BlockChainSimulator/node/p2p"
	"fmt"
)

type ConsensusHandlerType = string // consensus handler process the message related to consensus, for example, pbft, hotstuff, etc.
type AuxiliaryHandlerType = string // auxiliary handler process the message does not related to consensus, for example, read the Txs dataset, etc.

const (
	PBFT     ConsensusHandlerType = "pbft"
	HotStuff ConsensusHandlerType = "hotstuff"
	// add more intra consensus type here
)

const (
	TestAuxiliary       AuxiliaryHandlerType = "test"
	ProposeTxsAuxiliary AuxiliaryHandlerType = "ProposeTxs"
)

var consensusRegistry = make(map[string]func(addonModType string, attr *nodeattr.NodeAttr, p2p *p2p.P2PMod) (msgHandlerInterface.MsgHandlerMod, error))
var auxiliaryRegistry = make(map[string]func(attr *nodeattr.NodeAttr, p2p *p2p.P2PMod) msgHandlerInterface.MsgHandlerMod)

func init() {
	// register more msgHandler type here

	// Consensus Message Handler
	consensusRegistry[PBFT] = pbft.NewPbftCosensusMod

	// Auxiliary Message Handler
	auxiliaryRegistry[TestAuxiliary] = auxiliaryhandler.NewTestAuxiliaryMod
	auxiliaryRegistry[ProposeTxsAuxiliary] = auxiliaryhandler.NewProposeTxsAuxiliaryMod
}

// invoke by node.go to create a consensus-related msgHandler
func NewConsensusMod(consensusType string, addonModType string, attr *nodeattr.NodeAttr, p2p *p2p.P2PMod) (msgHandlerInterface.MsgHandlerMod, error) {
	// check if the consensus type exists, if so, return a new instance of the consensus
	if constructor, exists := consensusRegistry[consensusType]; exists {
		return constructor(addonModType, attr, p2p)
	}
	return nil, fmt.Errorf("unknown consensus type %s", consensusType)
}

// invoke by node.go to create an auxiliary msgHandler
func NewAuxiliaryMod(auxiliaryType string, attr *nodeattr.NodeAttr, p2p *p2p.P2PMod) (msgHandlerInterface.MsgHandlerMod, error) {
	// check if the auxiliary type exists, if so, return a new instance of the auxiliary
	if constructor, exists := auxiliaryRegistry[auxiliaryType]; exists {
		return constructor(attr, p2p), nil
	}
	return nil, fmt.Errorf("unknown auxiliary type %s", auxiliaryType)
}
