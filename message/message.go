package message

import (
	"encoding/json"
	"fmt"
	"time"
)

type MessageType int
type MessageHandler func(msg *Message)

type Message struct {
	MsgType MessageType
	Content []byte // the message body, for example, the Request, etc.
}

const (
	MsgEmpty MessageType = iota
	// Global
	MsgNodeReady // a node is ready to receive messages
	MsgStop      // stop the node
	MsgInject    // inject the Txs data(always from the client) to the system
	MsgReply     // send the verified requests back to the client

	// PBFT
	MsgPropose
	MsgPrePrepare
	MsgPrepare
	MsgCommit

	// Request for old sequence
	MsgRequestSeq
	MsgSeq

	// CShard
	MsgInputVerifyResult // L sends the result of input verification to LL
	MsgPreInject         // used to pre-inject the data to the system
	MsgBlockLegal        // a legal block, used to store the block
)

// the basic info of the shard to send back to the client
type Reply struct {
	Req  *Request  // verified request
	Time time.Time // the time when creating the reply

	Sid         int // the shard id
	ReqQueueLen int // the length of the request queue of the shard
}

func (rep *Reply) String() string {
	result := "["
	result += fmt.Sprintf("Time: %s, ", rep.Time.String())
	result += fmt.Sprintf("Sid: %d, ", rep.Sid)
	result += fmt.Sprintf("ReqQueueLen: %d, ", rep.ReqQueueLen)
	result += "], "
	result += fmt.Sprintf("Req:%v", rep.Req)
	return result
}

// Encode the message into a byte array
func (msg *Message) JsonEncode() []byte {
	msgbytes, err := json.Marshal(msg)
	if err != nil {
		return nil
	}
	return msgbytes
}

// Decode the byte array into a message
func JsonDecode(data []byte, msg *Message) error {
	return json.Unmarshal(data, msg)
}

func (msg *Message) String() string {
	result := "Message{"
	result += fmt.Sprintf("MsgType: %d, ", msg.MsgType)
	result += "Content len: " + fmt.Sprintf("%d", len(msg.Content))
	result += "}"
	return result
}
