package pbft

import (
	"BlockChainSimulator/config"
	"BlockChainSimulator/message"
	"BlockChainSimulator/node/msgHandler/msgHandlerInterface"
	"BlockChainSimulator/node/nodeattr"
	"BlockChainSimulator/node/p2p"
	"BlockChainSimulator/utils"
	"time"
)

// implement the ConsensusMod interface
var _ msgHandlerInterface.MsgHandlerMod = &PbftCosensusMod{}

type PbftCosensusMod struct {
	// vars from the belonging node
	nodeAttr *nodeattr.NodeAttr // the attribute of the belonging node
	p2pMod   *p2p.P2PMod        // the p2p network module of the belonging node

	// pbft related
	view          int // the nid of current view number
	pbft_num      int // number of nodes in the pbft network
	malicious_num int // max number of malicious nodes in the pbft network
	// malicious     bool // whether this node is malicious, does not implement this feature

	// consensus related
	requestQueue  *utils.Queue[message.Request] // the queue of requests waiting for consensus
	requestPool   map[string]*RequestInfo       // the pool of requests that have been received
	consensusDone chan struct{}                 // the channel to notify a round of  consensus is done

	addonMod PbftAddon // PbftAddon is an pointer-type interface
}

// Creates a new PbftCosensusMod with the given addonModType, err is not nil if the addonModType is not supported
func NewPbftCosensusMod(addonModType string, attr *nodeattr.NodeAttr, p2p *p2p.P2PMod) (msgHandlerInterface.MsgHandlerMod, error) {
	pbftMod := new(PbftCosensusMod)
	pbftMod.nodeAttr = attr
	pbftMod.p2pMod = p2p

	pbftMod.requestQueue = utils.NewQueue[message.Request]()
	pbftMod.requestPool = make(map[string]*RequestInfo)

	pbftMod.pbft_num = config.NodeNum
	pbftMod.malicious_num = (pbftMod.pbft_num - 1) / 3
	pbftMod.view = config.ViewNodeId

	addonMod, err := NewPbftAddon(addonModType, attr)
	if err != nil {
		return nil, err
	}
	pbftMod.addonMod = addonMod

	return pbftMod, nil
}

// At present, only the view node will receive the Propose message
func (pbftmod *PbftCosensusMod) handlePropose(msg *message.Message) {
	utils.LoggerInstance.Debug("handle propose")

	// decode the request from the message
	req := message.Request{}
	err := utils.Decode(msg.Content, &req)
	if err != nil {
		utils.LoggerInstance.Error("Error decoding the request")
		return
	}

	pbftmod.requestPool[string(req.Digest[:])] = NewRequestInfo(req)
	// invoke the addon module to handle the propose message
	flag := pbftmod.addonMod.HandleProposeAddon(msg)
	if !flag {
		utils.LoggerInstance.Warn("addon handle propose failed")
		return
	}

	pbftmod.requestQueue.Enqueue(req)
}

// PrePrepare means the node has received the propose message and broadcast the request to other nodes, the nodes need to check the request legality
func (pbftmod *PbftCosensusMod) handlePrePrepare(msg *message.Message) {
	utils.LoggerInstance.Debug("handle pre-prepare")

	// decode the request from the message
	req := message.Request{}
	err := utils.Decode(msg.Content, &req)
	if err != nil {
		utils.LoggerInstance.Error("Error decoding the request")
		return
	}

	pbftmod.requestPool[string(req.Digest[:])] = NewRequestInfo(req)
	// invoke the addon module to handle the pre-prepare message
	flag := pbftmod.addonMod.HandlePrePrepareAddon(msg)
	if !flag {
		utils.LoggerInstance.Warn("addon handle propose failed")
		return
	}
	pmsg := message.Message{
		MsgType: message.MsgPrepare,
		Content: msg.Content,
	}
	pbftmod.p2pMod.ConnMananger.Broadcast(pbftmod.nodeAttr.Ipaddr, pbftmod.getNeighbours(), pmsg.JsonEncode())
	utils.LoggerInstance.Info("Broadcast the prepare message")
}

// Prepare means some node think the request is legal and broadcast the prepare message to other nodess
func (pbftmod *PbftCosensusMod) handlePrepare(msg *message.Message) {
	utils.LoggerInstance.Debug("handle prepare")

	// decode the request from the message
	req := message.Request{}
	err := utils.Decode(msg.Content, &req)
	if err != nil {
		utils.LoggerInstance.Error("Error decoding the request")
		return
	}

	pbftmod.addonMod.HandlePrepareAddon(msg)

	// Seems the node received the PrepareMsg before any of the PrePrepareMsg
	if pbftmod.requestPool[string(req.Digest[:])] == nil {
		utils.LoggerInstance.Warn("Received prepare message before pre-prepare message")
		pbftmod.requestPool[string(req.Digest[:])] = NewRequestInfo(req)
	}

	prepareCnt := pbftmod.requestPool[string(req.Digest[:])].IncPrepareConfirm()
	needCnt := 0
	if pbftmod.nodeAttr.Nid == pbftmod.view {
		needCnt = 2 * pbftmod.malicious_num
	} else {
		needCnt = 2*pbftmod.malicious_num - 1 // in the current implementation, the view node will not send the prepare message
	}

	// Received enough prepare messages, broadcast the commit message
	if prepareCnt >= needCnt && !pbftmod.requestPool[string(req.Digest[:])].IsCommitBroadcasted() {
		pbftmod.requestPool[string(req.Digest[:])].SetCommitBroadcasted()
		utils.LoggerInstance.Info("Received enough prepare messages, broadcast the commit message")

		cmsg := message.Message{
			MsgType: message.MsgCommit,
			Content: msg.Content,
		}
		pbftmod.p2pMod.ConnMananger.Broadcast(pbftmod.nodeAttr.Ipaddr, pbftmod.getNeighbours(), cmsg.JsonEncode())
	}
}

// Commit means a node received enough prepare messages (usually 2/3) and broadcast the commit message to other nodes
func (pbftmod *PbftCosensusMod) handleCommit(msg *message.Message) {
	utils.LoggerInstance.Debug("handle commit")

	// decode the request from the message
	req := message.Request{}
	err := utils.Decode(msg.Content, &req)
	if err != nil {
		utils.LoggerInstance.Error("Error decoding the request")
		return
	}

	// Seems the node received the CommitMsg before any of the PrePrepareMsg and PrepareMsg
	if pbftmod.requestPool[string(req.Digest[:])] == nil {
		utils.LoggerInstance.Warn("Received commit message before pre-prepare message and prepare message")
		pbftmod.requestPool[string(req.Digest[:])] = NewRequestInfo(req)
	}

	if pbftmod.requestPool[string(req.Digest[:])].IncCommitConfirm() >= 2*pbftmod.malicious_num && !pbftmod.requestPool[string(req.Digest[:])].IsReplySent() {
		pbftmod.requestPool[string(req.Digest[:])].SetReplySent()
		utils.LoggerInstance.Info("Received enough commit messages")

		pbftmod.addonMod.HandleCommitAddon(msg)

		if pbftmod.nodeAttr.Nid == pbftmod.view {
			// time.Sleep(time.Millisecond * time.Duration(config.NodeNum*config.BlockSize/125)) // delay to mimic the network delay
			pbftmod.consensusDone <- struct{}{} // notify the consensus is done
		}
	}
}

// call in node.go according to the current implementation, you can also call this function in the New() function
func (pbftmod *PbftCosensusMod) RegisterHandlers() {
	pbftmod.p2pMod.MsgHandlerMap[message.MsgPropose] = pbftmod.handlePropose
	pbftmod.p2pMod.MsgHandlerMap[message.MsgPrePrepare] = pbftmod.handlePrePrepare
	pbftmod.p2pMod.MsgHandlerMap[message.MsgPrepare] = pbftmod.handlePrepare
	pbftmod.p2pMod.MsgHandlerMap[message.MsgCommit] = pbftmod.handleCommit
}

// get the ip addresses of the nodes in the same shard
func (pbftmod *PbftCosensusMod) getNeighbours() []string {
	neighbours := make([]string, 0)
	for _, ip := range config.IPMap[pbftmod.nodeAttr.Sid] {
		if ip == pbftmod.nodeAttr.Ipaddr {
			continue
		}
		neighbours = append(neighbours, ip)
	}
	return neighbours
}

// Run starts the intra-shard consensus.
// This function will work in the condition that the request is from other shards instead of the current shard
func (pbftmod *PbftCosensusMod) Run() {
	if pbftmod.nodeAttr.Nid != pbftmod.view {
		utils.LoggerInstance.Info("This node is not the view node, do not need to run intra-shard consensus")
		return
	}
	utils.LoggerInstance.Info("Start intra-shard consensus")
	for {
		// get the request from the request queue and broadcast the pre-prepare message
		req := pbftmod.requestQueue.Dequeue()
		// delay to mimic the network delay
		// time.Sleep(time.Millisecond * time.Duration(config.NodeNum*config.BlockSize/125))
		ppmsg := message.Message{
			MsgType: message.MsgPropose,
			Content: utils.Encode(req),
		}
		pbftmod.p2pMod.ConnMananger.Broadcast(pbftmod.nodeAttr.Ipaddr, pbftmod.getNeighbours(), ppmsg.JsonEncode())
		utils.LoggerInstance.Info("Broadcast the propose message")

		// wait for the consensus to be done
		<-pbftmod.consensusDone

		// consensus interval, waits for other nodes to complete the consensus
		time.Sleep(time.Millisecond * time.Duration(config.ConsensusInterval))
	}
}
