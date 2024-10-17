// Description: predefined solution of the node's runningMods
package runningMod

var (
	ClassicPBFTConsensusNode = []string{PBFTMod, ProposeTxsMod} // a classic PBFT consensus node
)

var (
	TestClientNode = []string{TestMod, StartSystemMod} // a test client node
)
