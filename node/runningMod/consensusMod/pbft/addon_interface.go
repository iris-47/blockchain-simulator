package pbft

import (
	"BlockChainSimulator/message"
	"BlockChainSimulator/node/nodeattr"
	"fmt"
)

type PbftAddon interface {
	// 4 kind  of PBFT intra message
	HandleProposeAddon(*message.Message) bool    // usually used by view node to preliminary check the request
	HandlePrePrepareAddon(*message.Message) bool // usually check if the request is legal
	HandlePrepareAddon(*message.Message) bool    // usually nothing to do
	HandleCommitAddon(*message.Message) bool     // usually used by some method to split the request and send to other shards
}

const (
	SimpleAddon = "simple"
	// add more addon type here
)

var addonRegistry = make(map[string]func(attr *nodeattr.NodeAttr) PbftAddon)

func init() {
	addonRegistry[SimpleAddon] = NewSimplePbftCosensusAddon
	// init more addon type here
}

func NewPbftAddon(addonType string, attr *nodeattr.NodeAttr) (PbftAddon, error) {
	// check if the addon type exists, if so, return a new instance of the addon
	if constructor, exists := addonRegistry[addonType]; exists {
		return constructor(attr), nil
	}
	return nil, fmt.Errorf("unknown addon type %s", addonType)
}
