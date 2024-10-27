package pbft

import (
	"BlockChainSimulator/message"
	"fmt"
)

type PbftAddon interface {
	// 4 kind  of PBFT intra message
	HandleProposeAddon(req *message.Request) bool    // usually used by view node to preliminary check the request
	HandlePrePrepareAddon(req *message.Request) bool // usually check if the request is legal
	HandlePrepareAddon(req *message.Request) bool    // usually nothing to do
	HandleCommitAddon(req *message.Request) bool     // usually used by some method to split the request and send to other shards
}

const (
	TestAddon   = "Test"
	SimpleAddon = "Simple"
	// add more addon type here
	TBDAddon = "TBD"
)

var addonRegistry = make(map[string]func(pbftMod *PbftCosensusMod) PbftAddon)

func init() {
	addonRegistry[TestAddon] = NewSimplePbftCosensusAddon
	addonRegistry[SimpleAddon] = NewSimplePbftCosensusAddon

	// init more addon type here
	addonRegistry[TBDAddon] = NewTBDPbftCosensusAddon
}

func NewPbftAddon(addonType string, pbftMod *PbftCosensusMod) (PbftAddon, error) {
	// check if the addon type exists, if so, return a new instance of the addon
	if constructor, exists := addonRegistry[addonType]; exists {
		return constructor(pbftMod), nil
	}
	return nil, fmt.Errorf("unknown addon type %s", addonType)
}
