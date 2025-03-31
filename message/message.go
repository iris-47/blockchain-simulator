package message

import (
	"encoding/json"
	"fmt"
	"time"
)

type MessageType int
type MessageHandler func(msg *Message)

// TODO: Add a field to indicate the sender of the message, and the signature
type Message struct {
	MsgType MessageType
	Content []byte // the message body, for example, the Request, etc.
}

const (
	MsgEmpty MessageType = iota
	// Client-related
	MsgInit       // to inform a node to get ready
	MsgNodeReady  // to notify a node is ready
	MsgStop       // to stop a node
	MsgInject     // to inject the transactions data(or request) to the consensus system
	MsgReply      // to active send something to the client
	MsgQuery      // to query the status/value of the blockchian system
	MsgReplyQuery // to reply the query message

	// Protocol-related
	MsgPropose
	MsgPrePrepare
	MsgPrepare
	MsgCommit
	MsgForward // to forward the message to another node, has different meaning in different protocols
	MsgForward1
	MsgForward2
	MsgVote          // to vote
	MsgQC            // to send the quorum certificate
	MsgConsensusDone // to notify the consensus is done

	// Sync-related
	MsgRequestSeq
	MsgSeq

	// CShard protocol
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
