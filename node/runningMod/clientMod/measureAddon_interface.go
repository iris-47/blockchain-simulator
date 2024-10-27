package clientMod

import (
	"BlockChainSimulator/message"
	"fmt"
)

type MeasureAddon interface {
	UpdateRecord(*message.Reply) // update the record of the addon
	WriteResult()                // write the result to file and close the file
}

const (
	TCL     = "TCL"     // Transaction confirmation latency
	TPS     = "TPS"     // Transaction per second
	WaitLen = "WaitLen" // Wait Request Length
)

var addonRegistry = make(map[string]func(measureMod *measureMod) MeasureAddon)

func init() {
	addonRegistry[TCL] = NewMeasureAddonTCL
	addonRegistry[TPS] = NewMeasureAddonTPS
	addonRegistry[WaitLen] = NewMeasureAddonWaitLen
}

func NewMeasureAddon(addonType string, measureMod *measureMod) (MeasureAddon, error) {
	// check if the addon type exists, if so, return a new instance of the addon
	if constructor, exists := addonRegistry[addonType]; exists {
		return constructor(measureMod), nil
	}
	return nil, fmt.Errorf("unknown addon type %s", addonType)
}
