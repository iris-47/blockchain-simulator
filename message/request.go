package message

import (
	"BlockChainSimulator/utils"
	"crypto/sha256"
	"encoding/json"
	"time"
)

type RequestType int // what kind of request a consensus is proposing

const (
	ReqEmpty       RequestType = iota
	ReqVerifyBlock             // the Content of the request is a block
	ReqVerifyTxs               // the Content of the request is a list of transactions

	// CShard
	ReqVerifyInputs // the Content of the request is a list of UTXOs
)

// always indicate which kind of request a consensus is proposing
type Request struct {
	ReqType RequestType
	Content []byte    // the request body, e.g., the block to be verified
	ReqTime time.Time // the time when the request is created
	Digest  [32]byte  // hash of the request
}

// fill the Digest field of the request
func (req *Request) CalDigest() {
	req.Digest = [32]byte{}
	b, err := json.Marshal(req)
	if err != nil {
		utils.LoggerInstance.Error("Error in encoding the request")
	}
	req.Digest = sha256.Sum256(b)
}

// used in Propose to indicate the sender of the request
// not in use now
type RequestWithSender struct {
	Sender  string
	Request Request
}
