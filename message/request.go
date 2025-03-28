package message

import (
	"BlockChainSimulator/signature"
	"BlockChainSimulator/utils"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"time"
)

type RequestType int // what kind of request a consensus is proposing

const (
	ReqEmpty             RequestType = iota
	ReqVerifyString                  // the Content of the request is a string
	ReqVerifyBlock                   // the Content of the request is a block
	ReqVerifyBlockHeader             // the Content of the request is a block header
	ReqVerifyTxs                     // the Content of the request is a list of transactions
	ReqVerifyInputs                  // the Content of the request is a list of UTXOs
)

// always indicate which kind of request a consensus is proposing
type Request struct {
	ShardId int
	ReqType RequestType
	Content []byte    // the request body, e.g., the block to be verified
	ReqTime time.Time // the time when the request is created
	Digest  [32]byte  // hash of the request

	Sig *signature.Signature // the signature of the request, optional
}

func NewRequest(shardId int, reqType RequestType, content []byte) *Request {
	request := Request{
		ShardId: shardId,
		ReqType: reqType,
		Content: content,
		ReqTime: time.Now(),
	}
	request.CalDigest()
	return &request
}

// signature is added to the project midway, so we need a new constructor and maitain the old one
func NewRequestWithSignature(shardId int, reqType RequestType, content []byte, sig *signature.Signature) *Request {
	request := Request{
		ShardId: shardId,
		ReqType: reqType,
		Content: content,
		ReqTime: time.Now(),
		Sig:     sig,
	}
	request.CalDigest()
	return &request
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

func (req *Request) String() string {
	str := "{\n"
	str += fmt.Sprintf("\tShardId: %d\n", req.ShardId)
	str += fmt.Sprintf("\tReqType: %d\n", req.ReqType)
	// str += fmt.Sprintf("\tContent: %s\n", string(req.Content))
	str += fmt.Sprintf("\tReqTime: %s\n", req.ReqTime)
	str += fmt.Sprintf("\tDigest: %x\n", req.Digest)
	return str
}
