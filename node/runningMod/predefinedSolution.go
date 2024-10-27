// Description: predefined solution of the node's runningMods
package runningMod

var (
	PBFTConsensusNode   = []string{PBFTMod}                          // a PBFT consensus node
	ClassicPBFTViewNode = []string{PBFTMod, ProposeBlockMod}         // a classic PBFT consensus node
	TBDViewNode         = []string{PBFTMod, ProposeBlock2ChannelMod} // a TBD consensus node
)

var (
	TestClientNode    = []string{StartSystemMod, TestMod}                             // a test client node
	ClassicClientNode = []string{StartSystemMod, MeasureMod}                          // a classic client node without txs injection
	TBDClientNode     = []string{StartSystemMod, MeasureMod, SendMimicContractTxsMod} // a TBD client node
)
