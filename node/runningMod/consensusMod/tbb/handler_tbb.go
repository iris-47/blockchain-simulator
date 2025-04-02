// the TBB protocol module
package tbb

import (
	"BlockChainSimulator/config"
	"BlockChainSimulator/message"
	"BlockChainSimulator/node/nodeattr"
	"BlockChainSimulator/node/p2p"
	"BlockChainSimulator/node/runningMod/consensusMod/ds"
	"BlockChainSimulator/node/runningMod/runningModInterface"
	"BlockChainSimulator/signature"
	"BlockChainSimulator/utils"
	"context"
	"sync"
	"time"
)

type TBBCosensusMod struct {
	// vars from the belonging node
	nodeAttr *nodeattr.NodeAttr // the attribute of the belonging node
	p2pMod   *p2p.P2PMod        // the p2p network module of the belonging node
	// consensus related
	view int // the nid of current view number
	// malicious     bool // whether this node is malicious, does not implement this feature
	startTime *utils.AtomicValue[time.Time]

	DSMod  *ds.DSCosensusMod       // the Delev-Strong consensus module
	DBBMod *_1Delta_BBConsensusMod // the DBB consensus module

	// local sets
	localSetA *utils.Set[string] // set 'A' in the paper, used by 1Δ-BB* protocol
	localSetC *utils.Set[string] // set 'C' in the paper, used by DS protocol
	localSetD []string           // set 'D' in the paper, to store the 3 commit point values

	referenceValue1 string // the reference value for the next step
	referenceValue2 string // the reference value for the next step

	isCommit bool // whether any commit point is reached
}

// the content with a list of signatures, not aggregate signature, used in DS protocol
type SigListContent struct {
	Input    []byte
	SigList  []*signature.Signature
	NodeList []int // indicate the nodes that have signed the request
}

type ReplyValue struct {
	Value string
	QCMap map[string]*signature.Signature
}

func NewTBBCosensusMod(attr *nodeattr.NodeAttr, p2p *p2p.P2PMod) runningModInterface.RunningMod {
	tbbMod := new(TBBCosensusMod)
	tbbMod.nodeAttr = attr
	tbbMod.p2pMod = p2p

	tbbMod.view = 0 // set 0 as the default view

	tbbMod.startTime = utils.NewAtomicValue[time.Time](time.Now())

	tbbMod.DSMod = ds.NewDSCosensusMod(attr, p2p).(*ds.DSCosensusMod)
	tbbMod.DBBMod = New_1Delta_BBConsensusMod(attr, p2p).(*_1Delta_BBConsensusMod)

	tbbMod.localSetA = tbbMod.DBBMod.BlckSet
	tbbMod.localSetC = tbbMod.DSMod.ExtrSet
	tbbMod.localSetD = make([]string, 3) // set D to store the 3 commit point values

	tbbMod.referenceValue1 = ""
	tbbMod.referenceValue2 = ""

	tbbMod.isCommit = false

	return tbbMod
}

// Initially, every node starts 1Δ-BB* (with BADs* embedded) protocol and DS protocol at time r = 0
// set all sets -> 0.
func (tbbMod *TBBCosensusMod) handleInitMsg(msg *message.Message) {
	utils.LoggerInstance.Info("Received an init message")

	startTime := time.Time{}
	err := utils.Decode(msg.Content, &startTime)
	if err != nil {
		utils.LoggerInstance.Error("Error decoding the init message, err: %v", err)
		return
	}

	// set the start time of the protocol
	tbbMod.startTime.Set(startTime, nil)
	utils.LoggerInstance.Info("Set the start time of the protocol to %v", tbbMod.startTime.Get())

	// clear the local sets
	tbbMod.localSetD = make([]string, 3)
	tbbMod.referenceValue1 = ""
	tbbMod.referenceValue2 = ""
	tbbMod.isCommit = false

	// init two sub-modules
	tbbMod.DSMod.HandleInitMsg(msg)
	tbbMod.DBBMod.HandleInitMsg(msg)

	t1 := int(float64(config.NodeNum)*config.ResilientRatio) - 1
	t2 := config.NodeNum - 1

	checkPoint1Timer := time.NewTimer(time.Until(startTime.Add(time.Duration(int64(t1+1)*config.TickInterval) * time.Millisecond)))
	checkPoint2Timer := time.NewTimer(time.Until(startTime.Add(time.Duration(int64(t2+1)*config.TickInterval) * time.Millisecond)))
	commitPoint2Timer := time.NewTimer(time.Until(startTime.Add(time.Duration(int64(t1+6)*config.TickInterval) * time.Millisecond)))
	commitPoint3Timer := time.NewTimer(time.Until(startTime.Add(time.Duration(int64(t2+6)*config.TickInterval) * time.Millisecond)))

	go tbbMod.referencePoint1Handler(*checkPoint1Timer)
	go tbbMod.referencePoint2Handler(*checkPoint2Timer)
	go tbbMod.commitPoint1Handler()
	go tbbMod.commitPoint2Handler(*commitPoint2Timer)
	go tbbMod.commitPoint3Handler(*commitPoint3Timer)
}

// Simultaneously follow the remaining rules of both protocols
func (tbbMod *TBBCosensusMod) handleProposeMsg(msg *message.Message) {
	tbbMod.DBBMod.HandleProposeMsg(msg)
	tbbMod.DSMod.HandleProposeMsg(msg)
}

func (tbbMod *TBBCosensusMod) handleForwardMsg(msg *message.Message) {
	tbbMod.DSMod.HandleForwardMsg(msg)
}

func (tbbMod *TBBCosensusMod) handleForward1Msg(msg *message.Message) {
	tbbMod.DBBMod.HandleForward1Msg(msg)
}

func (tbbMod *TBBCosensusMod) handleForward2Msg(msg *message.Message) {
	tbbMod.DBBMod.HandleForward2Msg(msg)
}

func (tbbMod *TBBCosensusMod) handleVoteMsg(msg *message.Message) {
	tbbMod.DBBMod.HandleVoteMsg(msg)
}

func (tbbMod *TBBCosensusMod) handleQCMsg(msg *message.Message) {
	tbbMod.DBBMod.HandleQCMsg(msg)
}

func (tbbMod *TBBCosensusMod) handleQueryMsg(msg *message.Message) {
	utils.LoggerInstance.Info("Received a query message")
	ipaddr := ""
	err := utils.Decode(msg.Content, &ipaddr)
	if err != nil {
		utils.LoggerInstance.Error("Error decoding the query message, err: %v", err)
		return
	}

	value := ""
	for _, item := range tbbMod.localSetD {
		if item != "" {
			value = item
			break
		}
	}

	replyValue := ReplyValue{
		Value: value,
		QCMap: tbbMod.DBBMod.QCMap,
	}

	repMsg := message.Message{
		MsgType: message.MsgReplyQuery,
		Content: utils.Encode(replyValue),
	}

	// send the reply message to the client
	tbbMod.p2pMod.ConnMananger.Send(ipaddr, repMsg.JsonEncode())
}

func (tbbMod *TBBCosensusMod) RegisterHandlers() {
	if config.IsMalicious && tbbMod.nodeAttr.Nid != config.ViewNodeId { // malicious view node still needs to register the handlers, but sends wrong propose
		tbbMod.p2pMod.RegisterHandler(message.MsgInit, tbbMod.handleInitMsg_m)
		tbbMod.p2pMod.RegisterHandler(message.MsgPropose, tbbMod.handleProposeMsg_m)
		tbbMod.p2pMod.RegisterHandler(message.MsgForward, tbbMod.handleForwardMsg_m)
		tbbMod.p2pMod.RegisterHandler(message.MsgForward1, tbbMod.handleForward1Msg_m)
		tbbMod.p2pMod.RegisterHandler(message.MsgForward2, tbbMod.handleForward2Msg_m)
		tbbMod.p2pMod.RegisterHandler(message.MsgVote, tbbMod.handleVoteMsg_m)
		tbbMod.p2pMod.RegisterHandler(message.MsgQC, tbbMod.handleQCMsg_m)
		tbbMod.p2pMod.RegisterHandler(message.MsgQuery, tbbMod.handleQueryMsg_m)

	} else {
		tbbMod.p2pMod.RegisterHandler(message.MsgInit, tbbMod.handleInitMsg)
		tbbMod.p2pMod.RegisterHandler(message.MsgPropose, tbbMod.handleProposeMsg)
		tbbMod.p2pMod.RegisterHandler(message.MsgForward, tbbMod.handleForwardMsg)
		tbbMod.p2pMod.RegisterHandler(message.MsgForward1, tbbMod.handleForward1Msg)
		tbbMod.p2pMod.RegisterHandler(message.MsgForward2, tbbMod.handleForward2Msg)
		tbbMod.p2pMod.RegisterHandler(message.MsgVote, tbbMod.handleVoteMsg)
		tbbMod.p2pMod.RegisterHandler(message.MsgQC, tbbMod.handleQCMsg)
		tbbMod.p2pMod.RegisterHandler(message.MsgQuery, tbbMod.handleQueryMsg)
	}
}

func (tbbMod *TBBCosensusMod) Run(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()

}

// When r = (t1 + 1) and (t2 + 1), check set C in DS, if |C| = 1, use the value i from C as areference for the next step.
func (tbbMod *TBBCosensusMod) referencePoint1Handler(refPoint1 time.Timer) {
	<-refPoint1.C
	if tbbMod.localSetC.Size() == 1 && tbbMod.DSMod.CommitValue == tbbMod.localSetC.GetItems()[0] {
		tbbMod.referenceValue1 = tbbMod.localSetC.GetItems()[0]
		utils.LoggerInstance.Info("The reference value is %v", tbbMod.referenceValue1)
	} else {
		utils.LoggerInstance.Info("The size of localSetC at %p is %d, the reference value is empty", tbbMod.localSetC, tbbMod.localSetC.Size())
	}
}

func (tbbMod *TBBCosensusMod) referencePoint2Handler(refPoint1 time.Timer) {
	<-refPoint1.C
	if tbbMod.localSetC.Size() == 1 && tbbMod.DSMod.CommitValue == tbbMod.localSetC.GetItems()[0] {
		tbbMod.referenceValue2 = tbbMod.localSetC.GetItems()[0]
		utils.LoggerInstance.Info("The reference value is %v", tbbMod.referenceValue2)
	} else {
		utils.LoggerInstance.Info("The size of localSetC at %p is %d, the reference value is empty", tbbMod.localSetC, tbbMod.localSetC.Size())
	}
}

// If Event i. in Step 2 of protocol 1A-BB* is triggered, commit the corresponding value. Then, add the value and the corresponding Commit point to the Di.
func (tbbMod *TBBCosensusMod) commitPoint1Handler() {
	<-tbbMod.DBBMod.CommitPoint1Trigger
	tbbMod.localSetD[0] = tbbMod.DBBMod.CommitValue
	tbbMod.isCommit = true
	utils.LoggerInstance.Info("The value commited at commit point 1 is %v", tbbMod.localSetD[0])
}

func (tbbMod *TBBCosensusMod) commitPoint2Handler(commitPoint2 time.Timer) {
	<-commitPoint2.C
	if tbbMod.isCommit {
		utils.LoggerInstance.Info("already commit, no need to commit again in commitPoint 2")
		return
	}

	if tbbMod.localSetA.Size() == 1 {
		tbbMod.localSetD[1] = tbbMod.localSetA.GetItems()[0]
		utils.LoggerInstance.Info("The value commited at  commit point 2 is %v, based on BADS* protocol", tbbMod.localSetD[1])
	} else if tbbMod.referenceValue1 != "" {
		tbbMod.localSetD[1] = tbbMod.referenceValue1
		utils.LoggerInstance.Info("The value commited at  commit point 2 is %v, based on the reference value", tbbMod.localSetD[1])
	} else {
		utils.LoggerInstance.Info("The value commited at  commit point 2 is empty")
	}
}

func (tbbMod *TBBCosensusMod) commitPoint3Handler(commitPoint3 time.Timer) {
	<-commitPoint3.C

	if tbbMod.localSetA.Size() == 1 {
		tbbMod.localSetD[2] = tbbMod.localSetA.GetItems()[0]
		utils.LoggerInstance.Info("The value commited at  commit point 3 is %v, based on BADS* protocol", tbbMod.localSetD[2])
	} else if tbbMod.referenceValue2 != "" {
		tbbMod.localSetD[2] = tbbMod.referenceValue2
		utils.LoggerInstance.Info("The value commited at  commit point 3 is %v, based on the reference value", tbbMod.localSetD[2])
	} else {
		utils.LoggerInstance.Info("The value commited at  commit point 3 is empty")
	}
}
