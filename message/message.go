package message

import (
	"encoding/json"
	"fmt"
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
	MsgVerified  // send the verified requests back to the client

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
