// This file contains the handler functions for the pbft consensus module
// in this implementation, msg.Content is the request, req.Content is the block
package pbft

import (
	"BlockChainSimulator/config"
	"BlockChainSimulator/message"
	"BlockChainSimulator/node/nodeattr"
	"BlockChainSimulator/node/p2p"
	"BlockChainSimulator/node/runningMod/runningModInterface"
	"BlockChainSimulator/utils"
	"context"
	"log"
	"sync"
	"time"
)

// implement the ConsensusMod interface
var _ runningModInterface.RunningMod = &PbftCosensusMod{}

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

// Creates a new PbftCosensusMod with the config.ConsensusMethod, err is not nil if the config.ConsensusMethod is not supported
func NewPbftCosensusMod(attr *nodeattr.NodeAttr, p2p *p2p.P2PMod) runningModInterface.RunningMod {
	pbftMod := new(PbftCosensusMod)
	pbftMod.nodeAttr = attr
	pbftMod.p2pMod = p2p

	pbftMod.requestQueue = utils.NewQueue[message.Request]()
	pbftMod.requestPool = make(map[string]*RequestInfo)

	pbftMod.pbft_num = config.NodeNum
	pbftMod.malicious_num = (pbftMod.pbft_num - 1) / 3
	pbftMod.view = config.ViewNodeId
	pbftMod.consensusDone = make(chan struct{})

	addonMod, err := NewPbftAddon(config.ConsensusMethod, pbftMod)
	if err != nil {
		utils.LoggerInstance.Error("Error creating pbft addon module: %v", err)
		log.Panicf("Error creating pbft addon module: %v", err)
	}
	pbftMod.addonMod = addonMod

	return pbftMod
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
	flag := pbftmod.addonMod.HandleProposeAddon(&req)
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
	flag := pbftmod.addonMod.HandlePrePrepareAddon(&req)
	if !flag {
		utils.LoggerInstance.Warn("addon handle pre-prepare failed")
		return
	}
	pmsg := message.Message{
		MsgType: message.MsgPrepare,
		Content: msg.Content,
	}
	pbftmod.p2pMod.ConnMananger.Broadcast(pbftmod.nodeAttr.Ipaddr, pbftmod.getNeighbours(config.IPMap[req.ShardId]), pmsg.JsonEncode())
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

	pbftmod.addonMod.HandlePrepareAddon(&req)

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
		pbftmod.p2pMod.ConnMananger.Broadcast(pbftmod.nodeAttr.Ipaddr, pbftmod.getNeighbours(config.IPMap[req.ShardId]), cmsg.JsonEncode())
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

	pbftmod.requestPool[string(req.Digest[:])].IncCommitConfirm()

	if pbftmod.requestPool[string(req.Digest[:])].GetCommitConfirm() >= 2*pbftmod.malicious_num && !pbftmod.requestPool[string(req.Digest[:])].IsReplySent() {
		pbftmod.requestPool[string(req.Digest[:])].SetReplySent()
		utils.LoggerInstance.Info("Received enough commit messages")

		pbftmod.addonMod.HandleCommitAddon(&req)

		// if pbftmod.nodeAttr.Nid == pbftmod.view {
		// 	pbftmod.consensusDone <- struct{}{} // notify the consensus is done
		// }
	}

	// view node need to wait for all the nodes to confirm the commit message.
	// This is not the standard practice for production. It's solely to ensure all nodes are synchronized in the same PBFT round.
	// BUG: when Shard Num > 4, the view node will not receive enough commit message from the other nodes for no reason
	if pbftmod.nodeAttr.Nid == pbftmod.view && pbftmod.requestPool[string(req.Digest[:])].GetCommitConfirm() == config.ShardNum-1 {
		utils.LoggerInstance.Info("Consensus is done!")
		pbftmod.consensusDone <- struct{}{} // notify the consensus is done
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
func (pbftmod *PbftCosensusMod) getNeighbours(shardIPs map[int]string) []string {
	neighbours := make([]string, 0)
	for _, ip := range shardIPs {
		if ip == pbftmod.nodeAttr.Ipaddr {
			continue
		}
		neighbours = append(neighbours, ip)
	}
	return neighbours
}

// Run starts the intra-shard consensus.
// This function will work in the condition that the request is from other shards instead of the current shard
func (pbftmod *PbftCosensusMod) Run(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()
	if pbftmod.nodeAttr.Nid != pbftmod.view {
		utils.LoggerInstance.Info("This node is not the view node, do not need to run intra-shard consensus Mod")
		return
	}

	utils.LoggerInstance.Info("Start the intra-shard consensus Mod")
	for {
		select {
		case <-ctx.Done():
			utils.LoggerInstance.Info("Stop the intra-shard consensus Mod")
			return
		default:
			// get the request from the request queue and broadcast the pre-prepare message
			req, err := pbftmod.requestQueue.Dequeue()
			if err != nil {
				time.Sleep(100 * time.Millisecond)
				continue
			}

			// delay to mimic the network delay
			// time.Sleep(time.Millisecond * time.Duration(config.NodeNum*config.BlockSize/125))

			ppmsg := message.Message{
				MsgType: message.MsgPrePrepare,
				Content: utils.Encode(req),
			}
			pbftmod.p2pMod.ConnMananger.Broadcast(pbftmod.nodeAttr.Ipaddr, pbftmod.getNeighbours(config.IPMap[req.ShardId]), ppmsg.JsonEncode())
			utils.LoggerInstance.Info("Broadcast the propose message")
			pbftmod.p2pMod.MsgHandlerMap[message.MsgPrePrepare](&ppmsg)

			// wait for the consensus to be done
			select {
			case <-pbftmod.consensusDone:
				// consensus interval, waits for other nodes to complete the consensus
				// time.Sleep(time.Millisecond * time.Duration(config.ConsensusInterval))
				utils.LoggerInstance.Info("Consensus is done, go next round")
			case <-ctx.Done():
				utils.LoggerInstance.Info("Stop the intra-shard consensus Mod")
				return
			}
		}
	}
}
