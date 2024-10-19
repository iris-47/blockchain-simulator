package nodeattr

import (
	"BlockChainSimulator/blockchain"
	"BlockChainSimulator/config"
	"strconv"
)

type NodeAttr struct {
	Sid      int
	Nid      int
	Ipaddr   string
	CurChain *blockchain.BlockChain
}

// URGENT: fullfill this function
func NewNodeAttr(sid int, nid int, pcc *config.ChainConfig) *NodeAttr {
	nodeAttr := new(NodeAttr)
	nodeAttr.Sid = sid
	nodeAttr.Nid = nid
	nodeAttr.Ipaddr = config.IPMap[sid][nid]
	if sid == config.ClientShard {
		nodeAttr.Ipaddr = config.ClientAddr
		return nodeAttr
	}
	// nodeAttr.CurChain = blockchain.NewBlockChain(pcc, nodeAttr.DB)
	return nodeAttr
}

// the hash of sid, nid and ipaddr
func (n *NodeAttr) GetIdentifier() string {
	return strconv.Itoa(n.Sid) + "-" + strconv.Itoa(n.Nid) + "-" + n.Ipaddr
}
