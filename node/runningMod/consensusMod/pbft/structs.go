package pbft

import (
	"BlockChainSimulator/message"
	"BlockChainSimulator/utils"
)

type PbftMessage struct {
	Request message.Request
	Round   int // the round of the consensus
}

// stores the information of a request,
type RequestInfo struct {
	Req               message.Request
	cntPrepareConfirm int  // the number of prepare messages received
	cntCommitConfirm  int  // the number of commit messages received
	isCommitBroadcast bool // whether the commit message has been broadcasted
	isReply           bool // whether the reply message has been sent
}

func NewRequestInfo(req message.Request) *RequestInfo {
	return &RequestInfo{
		Req:               req,
		cntPrepareConfirm: 0,
		cntCommitConfirm:  0,
		isCommitBroadcast: false,
		isReply:           false,
	}
}

func (pbftmod *PbftCosensusMod) getCurrentRound() int {
	pbftmod.roundLock.RLock()
	defer pbftmod.roundLock.RUnlock()
	return pbftmod.currentRound
}
func (pbftmod *PbftCosensusMod) IncCurrentRound() {
	pbftmod.roundLock.Lock()
	defer pbftmod.roundLock.Unlock()
	pbftmod.currentRound++
}

func (pbftmod *PbftCosensusMod) isCommitBroadcasted(digestStr string) bool {
	pbftmod.requestPoolLock.RLock()
	defer pbftmod.requestPoolLock.RUnlock()

	if pbftmod.requestPool[digestStr] == nil {
		return false
	}

	return pbftmod.requestPool[digestStr].isCommitBroadcast
}

func (pbftmod *PbftCosensusMod) setCommitBroadcasted(digestStr string) bool {
	pbftmod.requestPoolLock.Lock()
	defer pbftmod.requestPoolLock.Unlock()

	if pbftmod.requestPool[digestStr] == nil {
		utils.LoggerInstance.Error("Request not found in the request pool")
		return false
	}

	pbftmod.requestPool[digestStr].isCommitBroadcast = true
	return true
}

func (pbftmod *PbftCosensusMod) isReplySent(digestStr string) bool {
	pbftmod.requestPoolLock.RLock()
	defer pbftmod.requestPoolLock.RUnlock()

	if pbftmod.requestPool[digestStr] == nil {
		return false
	}

	return pbftmod.requestPool[digestStr].isReply
}

func (pbftmod *PbftCosensusMod) setReplySent(digestStr string) bool {
	pbftmod.requestPoolLock.Lock()
	defer pbftmod.requestPoolLock.Unlock()

	if pbftmod.requestPool[digestStr] == nil {
		utils.LoggerInstance.Error("Request not found in the request pool")
		return false
	}

	pbftmod.requestPool[digestStr].isReply = true
	return true
}

// // stores the information of a request,
// type RequestInfo struct {
// 	Req               message.Request
// 	cntPrepareConfirm int  // the number of prepare messages received
// 	cntCommitConfirm  int  // the number of commit messages received
// 	isCommitBroadcast bool // whether the commit message has been broadcasted
// 	isReply           bool // whether the reply message has been sent
// }

// func NewRequestInfo(req message.Request) *RequestInfo {
// 	return &RequestInfo{
// 		Req:               req,
// 		cntPrepareConfirm: 0,
// 		cntCommitConfirm:  0,
// 		isCommitBroadcast: false,
// 		isReply:           false,
// 	}
// }

// // Increase the number of prepare messages received and return the new count
// func (ri *RequestInfo) IncPrepareConfirm() int {
// 	ri.Lock()
// 	defer ri.Unlock()
// 	ri.cntPrepareConfirm++
// 	return ri.cntPrepareConfirm
// }

// func (ri *RequestInfo) GetPrepareConfirm() int {
// 	ri.Lock()
// 	defer ri.Unlock()
// 	return ri.cntPrepareConfirm
// }

// // Increase the number of commit messages received and return the new count
// func (ri *RequestInfo) IncCommitConfirm() int {
// 	ri.Lock()
// 	defer ri.Unlock()
// 	ri.cntCommitConfirm++
// 	return ri.cntCommitConfirm
// }

// func (ri *RequestInfo) GetCommitConfirm() int {
// 	ri.Lock()
// 	defer ri.Unlock()
// 	return ri.cntCommitConfirm
// }

// // Set isCommitBroadcast to true
// func (ri *RequestInfo) SetCommitBroadcasted() {
// 	ri.Lock()
// 	defer ri.Unlock()
// 	ri.isCommitBroadcast = true
// }

// func (ri *RequestInfo) IsCommitBroadcasted() bool {
// 	ri.Lock()
// 	defer ri.Unlock()
// 	return ri.isCommitBroadcast
// }

// // Set isReply to true
// func (ri *RequestInfo) SetReplySent() {
// 	ri.Lock()
// 	defer ri.Unlock()
// 	ri.isReply = true
// }

// func (ri *RequestInfo) IsReplySent() bool {
// 	ri.Lock()
// 	defer ri.Unlock()
// 	return ri.isReply
// }
