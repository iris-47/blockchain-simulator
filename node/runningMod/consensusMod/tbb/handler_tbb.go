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

	referenceValue string // the reference value for the next step

	isCommit bool // whether any commit point is reached
}

// the content with a list of signatures, not aggregate signature, used in DS protocol
type SigListContent struct {
	Input    []byte
	SigList  []*signature.Signature
	NodeList []int // indicate the nodes that have signed the request
}

type ReplyValue struct {
	SetD  []string
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

	tbbMod.referenceValue = ""

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
	tbbMod.localSetA = utils.NewSet[string]()
	tbbMod.localSetC = utils.NewSet[string]()
	tbbMod.localSetD = make([]string, 3)
	tbbMod.referenceValue = ""
	tbbMod.isCommit = false

	// init two sub-modules
	tbbMod.DSMod.HandleInitMsg(msg)
	tbbMod.DBBMod.HandleInitMsg(msg)

	t1 := int(float64(config.NodeNum)*config.ResilientRatio) - 1
	t2 := config.NodeNum - 1

	checkPoint1Timer := time.NewTimer(time.Duration(int64(t1+1)*config.TickInterval) * time.Millisecond)
	checkPoint2Timer := time.NewTimer(time.Duration(int64(t2+1)*config.TickInterval) * time.Millisecond)
	commitPoint2Timer := time.NewTimer(time.Duration(int64(t1+6)*config.TickInterval) * time.Millisecond)
	commitPoint3Timer := time.NewTimer(time.Duration(int64(t2+6)*config.TickInterval) * time.Millisecond)

	go tbbMod.referencePointHandler(*checkPoint1Timer)
	go tbbMod.referencePointHandler(*checkPoint2Timer)
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

	replyValue := ReplyValue{
		SetD:  tbbMod.localSetD,
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
	tbbMod.p2pMod.RegisterHandler(message.MsgInit, tbbMod.handleInitMsg)
	tbbMod.p2pMod.RegisterHandler(message.MsgPropose, tbbMod.handleProposeMsg)
	tbbMod.p2pMod.RegisterHandler(message.MsgForward, tbbMod.handleForwardMsg)
	tbbMod.p2pMod.RegisterHandler(message.MsgForward1, tbbMod.handleForward1Msg)
	tbbMod.p2pMod.RegisterHandler(message.MsgForward2, tbbMod.handleForward2Msg)
	tbbMod.p2pMod.RegisterHandler(message.MsgVote, tbbMod.handleVoteMsg)
	tbbMod.p2pMod.RegisterHandler(message.MsgQC, tbbMod.handleQCMsg)
	tbbMod.p2pMod.RegisterHandler(message.MsgQuery, tbbMod.handleQueryMsg)
}

func (tbbMod *TBBCosensusMod) Run(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()

}

// When r = (t1 + 1) and (t2 + 1), check set C in DS, if |C| = 1, use the value i from C as areference for the next step.
func (tbbMod *TBBCosensusMod) referencePointHandler(refPoint1 time.Timer) {
	<-refPoint1.C
	if tbbMod.localSetC.Size() == 1 {
		tbbMod.referenceValue = tbbMod.localSetC.GetItems()[0]
		utils.LoggerInstance.Info("The reference value is %v", tbbMod.referenceValue)
	} else {
		utils.LoggerInstance.Info("The reference value is empty")
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
		utils.LoggerInstance.Info("The value commited at  commit point 2 is %v, based on 1Δ-BB* protocol", tbbMod.localSetD[1])
	} else if tbbMod.referenceValue != "" {
		tbbMod.localSetD[1] = tbbMod.referenceValue
		utils.LoggerInstance.Info("The value commited at  commit point 2 is %v, based on the reference value", tbbMod.localSetD[1])
	} else {
		utils.LoggerInstance.Info("The value commited at  commit point 2 is empty")
	}
}

func (tbbMod *TBBCosensusMod) commitPoint3Handler(commitPoint3 time.Timer) {
	<-commitPoint3.C

	if tbbMod.localSetC.Size() == 1 {
		tbbMod.localSetD[2] = tbbMod.localSetC.GetItems()[0]
		utils.LoggerInstance.Info("The value commited at  commit point 3 is %v, based on DS protocol", tbbMod.localSetD[2])
	} else if tbbMod.referenceValue != "" {
		tbbMod.localSetD[2] = tbbMod.referenceValue
		utils.LoggerInstance.Info("The value commited at  commit point 3 is %v, based on the reference value", tbbMod.localSetD[2])
	} else {
		utils.LoggerInstance.Info("The value commited at  commit point 3 is empty")
	}
}
