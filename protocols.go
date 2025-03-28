// this file indicate some predefined protocols in the system
package main

import "BlockChainSimulator/node/runningMod"

type ProtocolMods struct {
	clientMods   []string
	nodeMods     []string
	viewNodeMods []string
}

var (
	// PredefinedProtocolMods is a map that contains the predefined protocol mods
	PredefinedProtocolMods = map[string]ProtocolMods{
		"ClassicPBFT": {
			clientMods:   []string{runningMod.StartSystemMod, runningMod.MeasureMod, runningMod.TestMod},
			nodeMods:     []string{runningMod.PBFTMod},
			viewNodeMods: []string{runningMod.PBFTMod, runningMod.ProposeBlockMod},
		},
		"TBD": {
			clientMods:   []string{runningMod.StartSystemMod, runningMod.MeasureMod, runningMod.SendMimicContractTxsMod},
			nodeMods:     []string{runningMod.PBFTMod},
			viewNodeMods: []string{runningMod.PBFTMod, runningMod.ProposeBlockMod},
		},
		"TBB": {
			clientMods:   []string{runningMod.StartSystemMod, runningMod.SendStringManualMod, runningMod.QueryMod},
			nodeMods:     []string{runningMod.TBBMod},
			viewNodeMods: []string{runningMod.TBBMod, runningMod.ProposeStringMod},
		},
		"DS": {
			clientMods:   []string{runningMod.StartSystemMod, runningMod.SendStringManualMod, runningMod.QueryMod},
			nodeMods:     []string{runningMod.DSMod},
			viewNodeMods: []string{runningMod.DSMod, runningMod.ProposeStringMod},
		},
	}
)
