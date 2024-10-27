package pbft

import (
	"BlockChainSimulator/message"
	"sync"
)

// mutex-protected, stores the information of a request,
type RequestInfo struct {
	sync.Mutex
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

// Increase the number of prepare messages received and return the new count
func (ri *RequestInfo) IncPrepareConfirm() int {
	ri.Lock()
	defer ri.Unlock()
	ri.cntPrepareConfirm++
	return ri.cntPrepareConfirm
}

func (ri *RequestInfo) GetPrepareConfirm() int {
	ri.Lock()
	defer ri.Unlock()
	return ri.cntPrepareConfirm
}

// Increase the number of commit messages received and return the new count
func (ri *RequestInfo) IncCommitConfirm() int {
	ri.Lock()
	defer ri.Unlock()
	ri.cntCommitConfirm++
	return ri.cntCommitConfirm
}

func (ri *RequestInfo) GetCommitConfirm() int {
	ri.Lock()
	defer ri.Unlock()
	return ri.cntCommitConfirm
}

// Set isCommitBroadcast to true
func (ri *RequestInfo) SetCommitBroadcasted() {
	ri.Lock()
	defer ri.Unlock()
	ri.isCommitBroadcast = true
}

func (ri *RequestInfo) IsCommitBroadcasted() bool {
	ri.Lock()
	defer ri.Unlock()
	return ri.isCommitBroadcast
}

// Set isReply to true
func (ri *RequestInfo) SetReplySent() {
	ri.Lock()
	defer ri.Unlock()
	ri.isReply = true
}

func (ri *RequestInfo) IsReplySent() bool {
	ri.Lock()
	defer ri.Unlock()
	return ri.isReply
}
