package tbb

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

// 1Δ-BB* (with BADs* embedded) protocol;
// Same as th DS protocol, its a part of the TBB protocol;
// It is not independently implemented in a package because it will not be used independently in the simulation
type _1Delta_BBConsensusMod struct {
	// vars from the belonging node
	nodeAttr *nodeattr.NodeAttr // the attribute of the belonging node
	p2pMod   *p2p.P2PMod        // the p2p network module of the belonging node
	// consensus related
	view int // the nid of current view number
	// malicious     bool // whether this node is malicious, does not implement this feature
	startTime *utils.AtomicValue[time.Time]

	voteMap map[string]*utils.Set[signature.Signature] // the map of votes, the key is the content of the vote, and the value is the signature of the vote

	// local sets
	QCMap      map[string]*signature.Signature // set '\phi' in the paper, used by 1Δ-BB* protocol
	proposeSet *utils.Set[string]              // the propose received
	BlckSet    *utils.Set[string]              // set 'blck' in the paper, used by 1Δ-BB* protocol

	// BADS* protocol
	BAproposeMap map[string]*utils.Set[signature.Signature] // the map used to generate aggregate signature
	BABlckSet    *utils.Set[string]                         // set 'A' in the paper, used by BADS* protocol embedded in 1Δ-BB*

	CommitPoint1Trigger chan struct{} // the trigger to commit the point 1
	isCommit            bool          // whether the node has committed the value
	CommitValue         string        // the value to commit
}

type QCContent struct {
	Content string // use string because []byte is not comparable
	Sig     *signature.Signature
}

type VoteContent struct {
	NodeId  int
	Content []byte
	Sig     *signature.Signature
}

func New_1Delta_BBConsensusMod(attr *nodeattr.NodeAttr, p2p *p2p.P2PMod) runningModInterface.RunningMod {
	_1dbbMod := new(_1Delta_BBConsensusMod)
	_1dbbMod.nodeAttr = attr
	_1dbbMod.p2pMod = p2p

	_1dbbMod.view = 0 // set 0 as the default view

	_1dbbMod.startTime = utils.NewAtomicValue(time.Now())

	_1dbbMod.voteMap = make(map[string]*utils.Set[signature.Signature])

	_1dbbMod.QCMap = make(map[string]*signature.Signature)
	_1dbbMod.proposeSet = utils.NewSet[string]()
	_1dbbMod.BlckSet = utils.NewSet[string]()

	_1dbbMod.BAproposeMap = make(map[string]*utils.Set[signature.Signature])
	_1dbbMod.BABlckSet = utils.NewSet[string]()

	_1dbbMod.CommitPoint1Trigger = make(chan struct{}, 1)
	_1dbbMod.isCommit = false
	_1dbbMod.CommitValue = ""

	return _1dbbMod
}

// Initially, every node pi starts the protocol at the same time T= 0, initializes the evidenceset i 0and bick_.
func (_1dbbMod *_1Delta_BBConsensusMod) HandleInitMsg(msg *message.Message) {
	utils.LoggerInstance.Info("Received an init message")

	startTime := time.Time{}
	err := utils.Decode(msg.Content, &startTime)
	if err != nil {
		utils.LoggerInstance.Error("Error decoding the init message, err: %v", err)
		return
	}

	// set the start time of the protocol
	_1dbbMod.startTime.Set(startTime, nil)
	utils.LoggerInstance.Info("Set the start time of the protocol to %v", _1dbbMod.startTime.Get())

	// clear the local sets
	_1dbbMod.voteMap = make(map[string]*utils.Set[signature.Signature])
	_1dbbMod.QCMap = make(map[string]*signature.Signature)
	_1dbbMod.proposeSet.Clear()
	_1dbbMod.BlckSet.Clear()

	_1dbbMod.BAproposeMap = make(map[string]*utils.Set[signature.Signature])
	_1dbbMod.BABlckSet.Clear()

	_1dbbMod.CommitPoint1Trigger = make(chan struct{}, 1)
	_1dbbMod.isCommit = false
	_1dbbMod.CommitValue = ""

	// wait until 4 delta to go BA protocol
	BAStartTime := _1dbbMod.startTime.Get().Add(time.Duration(4*config.TickInterval) * time.Millisecond)
	BAStartTimer := time.NewTimer(time.Until(BAStartTime))
	utils.LoggerInstance.Debug("Set the BA start time to %v", BAStartTime)
	go func() {
		<-BAStartTimer.C

		var input []byte
		if _1dbbMod.BlckSet.Size() == 1 {
			input = []byte(_1dbbMod.BlckSet.GetItems()[0])
		} else {
			input = []byte("0")
		}

		// broadcast the SigListContent(type of Forward2)
		sigListContent := SigListContent{
			Input:    input,
			SigList:  []*signature.Signature{signature.Sign(_1dbbMod.nodeAttr.SecKey, input)},
			NodeList: []int{_1dbbMod.nodeAttr.Nid},
		}

		forwardMsg := message.Message{
			MsgType: message.MsgForward2,
			Content: utils.Encode(sigListContent),
		}
		if _1dbbMod.BAproposeMap[string(sigListContent.Input)] == nil {
			_1dbbMod.BAproposeMap[string(sigListContent.Input)] = utils.NewSet[signature.Signature]()
		}
		_1dbbMod.BAproposeMap[string(sigListContent.Input)].Add(*sigListContent.SigList[0])
		utils.LoggerInstance.Info("Broadcast the start forward message of BADS*")
		_1dbbMod.p2pMod.ConnMananger.Broadcast(_1dbbMod.nodeAttr.Ipaddr, utils.GetNeighbours(config.IPMap[0], _1dbbMod.nodeAttr.Ipaddr), forwardMsg.JsonEncode())
	}()

	// wait until f + 6 round to final commit
	maliciousNodes := int64(config.MaliciousRatio * float64(config.NodeNum))
	delayDuration := (maliciousNodes + 6) * config.TickInterval * int64(time.Millisecond)
	commitTime := _1dbbMod.startTime.Get().Add(time.Duration(delayDuration))
	commitTimer := time.NewTimer(time.Until(commitTime))
	utils.LoggerInstance.Debug("Set final the commit time to %v", commitTime)
	go func() {
		<-commitTimer.C
		if _1dbbMod.isCommit {
			utils.LoggerInstance.Info("The final commit time f+6 is up, but the node has already committed the value")
		} else {
			utils.LoggerInstance.Info("Commit the request")

			// other commit operations

			for _, item := range _1dbbMod.BABlckSet.GetItems() {
				utils.LoggerInstance.Info("The value in BABlckSet are: %v", item)
			}

			if _1dbbMod.BABlckSet.Size() == 1 {
				utils.LoggerInstance.Info("The value to commit is: %v", _1dbbMod.BABlckSet.GetItems()[0])
				_1dbbMod.CommitValue = _1dbbMod.BABlckSet.GetItems()[0]
			} else {
				utils.LoggerInstance.Warn("The BABlckSet size is %d, The value to commit is 0", _1dbbMod.BABlckSet.Size())
				_1dbbMod.CommitValue = "0"
			}
		}
	}()

	// if view node, wait n+7 round to start the next consensus
	if _1dbbMod.nodeAttr.Nid == _1dbbMod.view {
		consensusDoneTimer := time.NewTimer(time.Duration(config.TickInterval*int64(config.NodeNum+7)) * time.Millisecond)
		go func() {
			<-consensusDoneTimer.C
			utils.LoggerInstance.Info("The consensus is done, start the next round")
			_1dbbMod.p2pMod.MsgHandlerMap[message.MsgConsensusDone](nil)
		}()
	}
}

// handle the propose message, the message is sent by the view node
func (_1dbbMod *_1Delta_BBConsensusMod) HandleProposeMsg(msg *message.Message) {
	utils.LoggerInstance.Info("Received a propose message")

	req := message.Request{}
	err := utils.Decode(msg.Content, &req)
	if err != nil {
		utils.LoggerInstance.Error("Error decoding the propose message")
		return
	}

	if _1dbbMod.checkSig(req.Sig) {
		_1dbbMod.proposeSet.Add(string(req.Content))

		// view node does not need to forward the message
		if _1dbbMod.nodeAttr.Nid != _1dbbMod.view {
			forwardMsg := message.Message{
				MsgType: message.MsgForward1,
				Content: utils.Encode(req),
			}
			utils.LoggerInstance.Info("Broadcast the forward message of 1Δ-BB*")
			_1dbbMod.p2pMod.ConnMananger.Broadcast(_1dbbMod.nodeAttr.Ipaddr, utils.GetNeighbours(config.IPMap[0], _1dbbMod.nodeAttr.Ipaddr), forwardMsg.JsonEncode())
		}

		voteTimer := time.NewTimer(time.Duration(config.TickInterval) * time.Millisecond)
		go func() {
			<-voteTimer.C
			utils.LoggerInstance.Info("The vote time is up, try to vote")
			if _1dbbMod.proposeSet.Size() == 1 {
				sig := signature.Sign(_1dbbMod.nodeAttr.SecKey, []byte(_1dbbMod.proposeSet.GetItems()[0]))
				voteContent := VoteContent{
					NodeId:  _1dbbMod.nodeAttr.Nid,
					Content: []byte(_1dbbMod.proposeSet.GetItems()[0]),
					Sig:     sig,
				}
				voteMsg := message.Message{
					MsgType: message.MsgVote,
					Content: utils.Encode(voteContent),
				}

				if _1dbbMod.voteMap[string(voteContent.Content)] == nil {
					_1dbbMod.voteMap[string(voteContent.Content)] = utils.NewSet[signature.Signature]()
				}
				_1dbbMod.voteMap[string(voteContent.Content)].Add(*sig)
				utils.LoggerInstance.Info("Broadcast the vote message")
				_1dbbMod.p2pMod.ConnMananger.Broadcast(_1dbbMod.nodeAttr.Ipaddr, utils.GetNeighbours(config.IPMap[0], _1dbbMod.nodeAttr.Ipaddr), voteMsg.JsonEncode())
			}
		}()
	} else {
		utils.LoggerInstance.Warn("The signature of the propose message is not valid")
		return
	}
}

// similar to handleProposeMsg, but the message is forwarded by the nodes
func (_1dbbMod *_1Delta_BBConsensusMod) HandleForward1Msg(msg *message.Message) {
	utils.LoggerInstance.Debug("Received a forward message")

	req := message.Request{}
	err := utils.Decode(msg.Content, &req)
	if err != nil {
		utils.LoggerInstance.Error("Error decoding the forward message")
		return
	}

	if _1dbbMod.checkSig(req.Sig) {
		if _1dbbMod.proposeSet.Contains(string(req.Content)) {
			utils.LoggerInstance.Info("The content of the forward message is already in the propose set")
			return
		}
		_1dbbMod.proposeSet.Add(string(req.Content))
		utils.LoggerInstance.Info("The content of the forward message is added to the propose set")
	} else {
		utils.LoggerInstance.Warn("The signature of the forward message is not valid")
		return
	}
}

func (_1dbbMod *_1Delta_BBConsensusMod) HandleVoteMsg(msg *message.Message) {
	utils.LoggerInstance.Debug("Received a vote message")

	voteContent := VoteContent{}
	err := utils.Decode(msg.Content, &voteContent)
	if err != nil {
		utils.LoggerInstance.Error("Error decoding the vote message")
		return
	}
	if _1dbbMod.checkSig(voteContent.Sig) {
		if _1dbbMod.voteMap[string(voteContent.Content)] == nil {
			_1dbbMod.voteMap[string(voteContent.Content)] = utils.NewSet[signature.Signature]()
		}
		_1dbbMod.voteMap[string(voteContent.Content)].Add(*voteContent.Sig)
		maliciousNum := int(config.MaliciousRatio * float64(config.NodeNum))
		if _1dbbMod.voteMap[string(voteContent.Content)].Size() == config.NodeNum-maliciousNum {
			if _1dbbMod.QCMap[string(voteContent.Content)] != nil { // already have the qc
				return
			}
			utils.LoggerInstance.Info("Get enough votes of %v to make QC", string(voteContent.Content))
			aggSig, err := signature.AggregateSignatures(_1dbbMod.voteMap[string(voteContent.Content)].GetItemRefs())
			if err != nil {
				utils.LoggerInstance.Error("Error aggregating the signatures, err: %v", err)
				return
			}
			qc := QCContent{
				Content: string(voteContent.Content),
				Sig:     aggSig,
			}
			_1dbbMod.QCMap[qc.Content] = qc.Sig
			_1dbbMod.BlckSet.Add(qc.Content)

			timeFromStart := time.Since(_1dbbMod.startTime.Get())
			if _1dbbMod.BlckSet.Size() == 1 && timeFromStart < time.Duration(3*config.TickInterval)*time.Millisecond {
				utils.LoggerInstance.Info("Commit the value %v at commit point 1 in 1Δ-BB* protocol", string(voteContent.Content))
				_1dbbMod.CommitValue = string(voteContent.Content)
				_1dbbMod.isCommit = true
				_1dbbMod.CommitPoint1Trigger <- struct{}{}

				// broadcast the qc
				qcMsg := message.Message{
					MsgType: message.MsgQC,
					Content: utils.Encode(qc),
				}
				utils.LoggerInstance.Info("Broadcast the qc message")
				_1dbbMod.p2pMod.ConnMananger.Broadcast(_1dbbMod.nodeAttr.Ipaddr, utils.GetNeighbours(config.IPMap[0], _1dbbMod.nodeAttr.Ipaddr), qcMsg.JsonEncode())
			}
		}
	} else {
		utils.LoggerInstance.Warn("The signature of the vote message is not valid")
		return
	}
}

func (_1dbbMod *_1Delta_BBConsensusMod) HandleQCMsg(msg *message.Message) {
	utils.LoggerInstance.Debug("Received a qc message")
	qc := QCContent{}
	err := utils.Decode(msg.Content, &qc)
	if err != nil {
		utils.LoggerInstance.Error("Error decoding the qc message")
		return
	}
	if _1dbbMod.checkSig(qc.Sig) {
		// Non-Honest Maiority Detection in 1Δ-BB* protocol
		// not implemented yet
	}
}

// BADS* protocol
func (_1dbbMod *_1Delta_BBConsensusMod) HandleForward2Msg(msg *message.Message) {
	utils.LoggerInstance.Debug("Received a forward message of BADS*")

	sigListContent := SigListContent{}
	err := utils.Decode(msg.Content, &sigListContent)
	if err != nil {
		utils.LoggerInstance.Error("Error decoding the forward message of BADS*")
		return
	}

	if _1dbbMod.checkSigList(sigListContent.SigList) {
		// get the round of the protocol
		timeFromStart := time.Since(_1dbbMod.startTime.Get())

		// time < 5Δ, try to generate the aggregate signature
		if timeFromStart < time.Duration(5*config.TickInterval)*time.Millisecond {
			if _1dbbMod.BAproposeMap[string(sigListContent.Input)] == nil {
				_1dbbMod.BAproposeMap[string(sigListContent.Input)] = utils.NewSet[signature.Signature]()
			}
			_1dbbMod.BAproposeMap[string(sigListContent.Input)].Add(*sigListContent.SigList[0])
			maliciousNum := int(config.MaliciousRatio * float64(config.NodeNum))
			if _1dbbMod.BAproposeMap[string(sigListContent.Input)].Size() == config.NodeNum-maliciousNum {
				utils.LoggerInstance.Info("Get enough \"propose\" of BADS* to make aggSig on %v", string(sigListContent.Input))
				aggSig, err := signature.AggregateSignatures(_1dbbMod.BAproposeMap[string(sigListContent.Input)].GetItemRefs())
				if err != nil {
					utils.LoggerInstance.Error("Error aggregating the signatures, err: %v", err)
					return
				}

				// wait until 6Δ to broadcast the aggregate signature
				forwardTimer := time.NewTimer(time.Until(_1dbbMod.startTime.Get().Add(time.Duration(6*config.TickInterval) * time.Millisecond)))
				go func() {
					<-forwardTimer.C
					BAForwardMsg := message.Message{
						MsgType: message.MsgForward2,
						Content: utils.Encode(SigListContent{
							Input:    sigListContent.Input,
							SigList:  []*signature.Signature{aggSig},
							NodeList: []int{-1}, // -1 means the signature is aggregate signature
						}),
					}
					utils.LoggerInstance.Info("Broadcast the forward message of BADS* with aggSig")
					_1dbbMod.p2pMod.ConnMananger.Broadcast(_1dbbMod.nodeAttr.Ipaddr, utils.GetNeighbours(config.IPMap[0], _1dbbMod.nodeAttr.Ipaddr), BAForwardMsg.JsonEncode())
				}()
			}
		} else {
			// time > 5Δ, same to DS protocol
			if _1dbbMod.BABlckSet.Contains(string(sigListContent.Input)) {
				return
			} else {
				sigLen := len(sigListContent.SigList)
				round := int(timeFromStart.Milliseconds()/config.TickInterval) - 5

				// check if the length of the signature list is equal to the round number
				if sigLen != round && sigLen != round+1 {
					utils.LoggerInstance.Warn("The length of the signature list %d is not equal to the round number %d", sigLen, round)
					// return
				}

				_1dbbMod.BABlckSet.Add(string(sigListContent.Input))
				utils.LoggerInstance.Info("The content of the forward message is added to the blck set")

				// wait until round==sigLen and broadcast Forward message
				for {
					if time.Since(_1dbbMod.startTime.Get()).Milliseconds()/config.TickInterval-5 < int64(sigLen) {
						break
					}
					time.Sleep(time.Millisecond * 100)
				}

				// add signature and forward the message
				sigListContent.SigList = append(sigListContent.SigList, signature.Sign(_1dbbMod.nodeAttr.SecKey, sigListContent.Input))
				sigListContent.NodeList = append(sigListContent.NodeList, _1dbbMod.nodeAttr.Nid)

				BAForwardMsg := message.Message{
					MsgType: message.MsgForward2,
					Content: utils.Encode(sigListContent),
				}
				_1dbbMod.p2pMod.ConnMananger.Broadcast(_1dbbMod.nodeAttr.Ipaddr, utils.GetNeighbours(config.IPMap[0], _1dbbMod.nodeAttr.Ipaddr), BAForwardMsg.JsonEncode())
				utils.LoggerInstance.Info("Broadcast the forward message of BADS* with own signature")
			}
		}
	} else {
		utils.LoggerInstance.Warn("The signature of the forward message is not valid")
		return
	}
}

func (_1dbbMod *_1Delta_BBConsensusMod) RegisterHandlers() {
	_1dbbMod.p2pMod.RegisterHandler(message.MsgInit, _1dbbMod.HandleInitMsg)
	_1dbbMod.p2pMod.RegisterHandler(message.MsgPropose, _1dbbMod.HandleProposeMsg)
	_1dbbMod.p2pMod.RegisterHandler(message.MsgForward1, _1dbbMod.HandleForward1Msg)
	_1dbbMod.p2pMod.RegisterHandler(message.MsgForward2, _1dbbMod.HandleForward2Msg)
	_1dbbMod.p2pMod.RegisterHandler(message.MsgVote, _1dbbMod.HandleVoteMsg)
	_1dbbMod.p2pMod.RegisterHandler(message.MsgQC, _1dbbMod.HandleQCMsg)
}

func (_1dbbMod *_1Delta_BBConsensusMod) Run(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()

	// do nothing, all the operations are triggered by the messages
}

// check the signature of SigListContent
// TODO: implement this function
// 需要前置完成密钥广播模块，尚未完成
func (_1dbbMod *_1Delta_BBConsensusMod) checkSigList([]*signature.Signature) bool {
	return true
}

func (_1dbbMod *_1Delta_BBConsensusMod) checkSig(*signature.Signature) bool {
	return true
}
