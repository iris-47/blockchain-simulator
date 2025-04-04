// the Dolev-Strong protocol module, only consider 1 round of protocol
package ds

import (
	"BlockChainSimulator/config"
	"BlockChainSimulator/message"
	"BlockChainSimulator/node/nodeattr"
	"BlockChainSimulator/node/p2p"
	"BlockChainSimulator/node/runningMod/runningModInterface"
	"BlockChainSimulator/signature"
	"BlockChainSimulator/utils"
	"context"
	"sync"
	"time"
)

type DSCosensusMod struct {
	// vars from the belonging node
	nodeAttr *nodeattr.NodeAttr // the attribute of the belonging node
	p2pMod   *p2p.P2PMod        // the p2p network module of the belonging node
	// consensus related
	view int // the nid of current view number
	// malicious     bool // whether this node is malicious, has not implemented this feature
	startTime *utils.AtomicValue[time.Time]

	// local sets
	ExtrSet     *utils.Set[string] // 'extracted set' in the paper
	CommitValue string             // the value to commit
}

// the content with a list of signatures, not aggregate signature, used in DS protocol
type SigListContent struct {
	Req      message.Request
	SigList  []*signature.Signature
	NodeList []int // indicate the nodes that have signed the request
}

func NewDSCosensusMod(attr *nodeattr.NodeAttr, p2p *p2p.P2PMod) runningModInterface.RunningMod {
	dsMod := new(DSCosensusMod)
	dsMod.nodeAttr = attr
	dsMod.p2pMod = p2p

	dsMod.view = 0 // set 0 as the default view

	dsMod.startTime = utils.NewAtomicValue(time.Now())

	dsMod.ExtrSet = utils.NewSet[string]()

	return dsMod
}

// Initially, every node i's extracted set extri=0.
func (dsMod *DSCosensusMod) HandleInitMsg(msg *message.Message) {
	utils.LoggerInstance.Info("Received an init message")

	startTime := time.Time{}
	err := utils.Decode(msg.Content, &startTime)
	if err != nil {
		utils.LoggerInstance.Error("Error decoding the init message, err: %v", err)
		return
	}

	// set the start time of the protocol
	dsMod.startTime.Set(startTime, nil)
	utils.LoggerInstance.Info("Set the start time of the protocol to %v", dsMod.startTime.Get())

	// clear the local sets
	dsMod.ExtrSet.Clear()

	// wait until f + 1 round to commit
	maliciousNodes := int64(config.MaliciousRatio * float64(config.NodeNum))
	utils.LoggerInstance.Debug("The number of malicious nodes is %d(%.2f)", maliciousNodes, config.MaliciousRatio)
	delayDuration := (maliciousNodes + 1) * config.TickInterval * int64(time.Millisecond)
	commitTime := dsMod.startTime.Get().Add(time.Duration(delayDuration))
	commitTimer := time.NewTimer(time.Until(commitTime))
	utils.LoggerInstance.Debug("Set the commit time to %v", commitTime)
	go func() {
		<-commitTimer.C
		utils.LoggerInstance.Info("Commit the request")

		// other commit operations

		for _, item := range dsMod.ExtrSet.GetItems() {
			utils.LoggerInstance.Info("The value in ExtrSet at %p are: %v", dsMod.ExtrSet, item)
		}

		if dsMod.ExtrSet.Size() == 1 {
			utils.LoggerInstance.Info("The value to commit is: %v", dsMod.ExtrSet.GetItems()[0])
			dsMod.CommitValue = dsMod.ExtrSet.GetItems()[0]
		} else {
			utils.LoggerInstance.Warn("The ExtrSet size is %d, The value to commit is 0", dsMod.ExtrSet.Size())
			dsMod.CommitValue = "0"
		}
	}()

	// if view node, set consensus done timer
	if dsMod.nodeAttr.Nid == dsMod.view {
		consensusDoneTimer := time.NewTimer(time.Duration(config.TickInterval*int64(config.NodeNum+7)) * time.Millisecond)
		go func() {
			<-consensusDoneTimer.C
			utils.LoggerInstance.Info("The consensus is done, start the next round")
			dsMod.p2pMod.MsgHandlerMap[message.MsgConsensusDone](nil)
		}()
	}
}

// handle the propose message, the message is sent by the view node
func (dsMod *DSCosensusMod) HandleProposeMsg(msg *message.Message) {
	utils.LoggerInstance.Info("Received a propose message")

	req := message.Request{}
	err := utils.Decode(msg.Content, &req)
	if err != nil {
		utils.LoggerInstance.Error("Error decoding the propose message")
		return
	}

	dsMod.ExtrSet.Add(string(req.Content))

	// view node does not need to forward the message
	if dsMod.nodeAttr.Nid != dsMod.view {
		// wait until round==1 and broadcast Forward message
		for {
			if time.Since(dsMod.startTime.Get()).Milliseconds()/config.TickInterval < int64(1) {
				break
			}
			time.Sleep(time.Millisecond * 100)
		}

		sigListContent := SigListContent{
			Req:      req,
			SigList:  []*signature.Signature{req.Sig, signature.Sign(dsMod.nodeAttr.SecKey, req.Content)},
			NodeList: []int{dsMod.view, dsMod.nodeAttr.Nid},
		}
		forwardMsg := message.Message{
			MsgType: message.MsgForward,
			Content: utils.Encode(sigListContent),
		}
		dsMod.p2pMod.ConnMananger.Broadcast(dsMod.nodeAttr.Ipaddr, utils.GetNeighbours(config.IPMap[0], dsMod.nodeAttr.Ipaddr), forwardMsg.JsonEncode())
		return
	}
}

// check if the 'b' is already in the ExtrSet
// check if the length of the signature list is equal to the round number
// check if the signature list is correct
// forward the message to everyone
func (dsMod *DSCosensusMod) HandleForwardMsg(msg *message.Message) {
	utils.LoggerInstance.Debug("Received a forward message")

	sigListContent := SigListContent{}
	err := utils.Decode(msg.Content, &sigListContent)
	if err != nil {
		utils.LoggerInstance.Error("Error decoding the forward message")
		return
	}

	// get the round of the protocol
	round := int(time.Since(dsMod.startTime.Get()).Milliseconds() / config.TickInterval)
	sigLen := len(sigListContent.SigList)

	// If ~b not belongs to ExtrSet: add ~b to ExtrSet and send fowared to everyone
	if dsMod.ExtrSet.Contains(string(sigListContent.Req.Content)) {
		utils.LoggerInstance.Info("The request is already in the ExtrSet")
		return
	} else {
		if !dsMod.checkSigList(sigListContent.SigList) {
			utils.LoggerInstance.Warn("The signature list is not correct")
			return
		}

		if sigLen != round && sigLen != round+1 {
			utils.LoggerInstance.Warn("The length of the signature list %d is not equal to the round number %d", sigLen, round)
			// return
		}

		// add the request to the local set C
		utils.LoggerInstance.Info("Add the request to the ExtrSet")
		dsMod.ExtrSet.Add(string(sigListContent.Req.Content))

		// wait until round==sigLen and broadcast Forward message
		for {
			if time.Since(dsMod.startTime.Get()).Milliseconds()/config.TickInterval < int64(sigLen) {
				break
			}
			time.Sleep(time.Millisecond * 100)
		}

		sigListContent.SigList = append(sigListContent.SigList, signature.Sign(dsMod.nodeAttr.SecKey, sigListContent.Req.Content))
		sigListContent.NodeList = append(sigListContent.NodeList, dsMod.nodeAttr.Nid)

		forwardMsg := message.Message{
			MsgType: message.MsgForward,
			Content: utils.Encode(sigListContent),
		}

		dsMod.p2pMod.ConnMananger.Broadcast(dsMod.nodeAttr.Ipaddr, utils.GetNeighbours(config.IPMap[0], dsMod.nodeAttr.Ipaddr), forwardMsg.JsonEncode())
		utils.LoggerInstance.Info("Broadcast the forward message")
	}
}

func (dsMod *DSCosensusMod) handleQueryMsg(msg *message.Message) {
	ip := ""
	err := utils.Decode(msg.Content, &ip)
	if err != nil {
		utils.LoggerInstance.Error("Error decoding the query message")
		return
	}

	// send the commit value to the query node
	replyMsg := message.Message{
		MsgType: message.MsgReplyQuery,
		Content: utils.Encode(dsMod.CommitValue),
	}

	dsMod.p2pMod.ConnMananger.Send(ip, replyMsg.JsonEncode())
}

func (dsMod *DSCosensusMod) RegisterHandlers() {
	if config.IsMalicious && dsMod.nodeAttr.Nid != config.ViewNodeId { // malicious view node still needs to register the handlers, but sends wrong propose
		dsMod.p2pMod.RegisterHandler(message.MsgInit, dsMod.HandleInitMsg_m)
		dsMod.p2pMod.RegisterHandler(message.MsgPropose, dsMod.HandleProposeMsg_m)
		dsMod.p2pMod.RegisterHandler(message.MsgForward, dsMod.HandleForwardMsg_m)
		dsMod.p2pMod.RegisterHandler(message.MsgQuery, dsMod.handleQueryMsg_m)
	} else {
		dsMod.p2pMod.RegisterHandler(message.MsgInit, dsMod.HandleInitMsg)
		dsMod.p2pMod.RegisterHandler(message.MsgPropose, dsMod.HandleProposeMsg)
		dsMod.p2pMod.RegisterHandler(message.MsgForward, dsMod.HandleForwardMsg)
		dsMod.p2pMod.RegisterHandler(message.MsgQuery, dsMod.handleQueryMsg)
	}
}

func (dsMod *DSCosensusMod) Run(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()

	// do nothing, all the operations are triggered by the messages
}

// check the signature of SigListContent
// TODO: implement this function
// 需要前置完成密钥广播模块，尚未完成
func (dsMod *DSCosensusMod) checkSigList([]*signature.Signature) bool {
	return true
}
