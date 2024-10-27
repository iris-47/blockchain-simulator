package nodeattr

import (
	"BlockChainSimulator/blockchain"
	"BlockChainSimulator/config"
	"BlockChainSimulator/utils"
	"log"
	"strconv"
)

type NodeAttr struct {
	Sid      int
	Nid      int
	Ipaddr   string
	CurChain *blockchain.BlockChain
}

// Opt: why not move CurChain to the PBFT running mod?
func NewNodeAttr(sid int, nid int, pcc *config.ChainConfig) *NodeAttr {
	nodeAttr := new(NodeAttr)
	nodeAttr.Sid = sid
	nodeAttr.Nid = nid
	nodeAttr.Ipaddr = config.IPMap[sid][nid]
	if sid == config.ClientShard {
		nodeAttr.Ipaddr = config.ClientAddr
		return nodeAttr
	}

	var err error
	nodeAttr.CurChain, err = blockchain.NewBlockChain(pcc)
	if err != nil {
		utils.LoggerInstance.Error("Failed to create the blockchain")
		log.Panic(err)
	}

	return nodeAttr
}

// the hash of sid, nid and ipaddr
func (n *NodeAttr) GetIdentifier() string {
	return strconv.Itoa(n.Sid) + "-" + strconv.Itoa(n.Nid) + "-" + n.Ipaddr
}
