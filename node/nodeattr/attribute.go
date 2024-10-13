package nodeattr

import (
	"BlockChainSimulator/blockchain"
	"BlockChainSimulator/config"
	"strconv"

	"github.com/ethereum/go-ethereum/ethdb"
)

type NodeAttr struct {
	Sid      int
	Nid      int
	Ipaddr   string
	CurChain *blockchain.BlockChain
	DB       ethdb.Database
}

// URGENT: fullfill this function
func NewNodeAttr(sid int, nid int, pcc *config.ChainConfig) *NodeAttr {
	return &NodeAttr{
		Sid:    sid,
		Nid:    nid,
		Ipaddr: config.IPMap[sid][nid],
	}
}

// the hash of sid, nid and ipaddr
func (n *NodeAttr) GetIdentifier() string {
	return strconv.Itoa(n.Sid) + "-" + strconv.Itoa(n.Nid) + "-" + n.Ipaddr
}
