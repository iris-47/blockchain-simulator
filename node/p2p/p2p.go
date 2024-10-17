package p2p

import (
	"BlockChainSimulator/config"
	"BlockChainSimulator/message"
	"BlockChainSimulator/utils"
	"bufio"
	"io"
	"net"
	"sync"
)

type P2PMod struct {
	listenAddr    config.Address                                 // ip:port
	MsgHandlerMap map[message.MessageType]message.MessageHandler // message type -> handler
	ConnMananger  ConnMananger

	wg sync.WaitGroup
}

func NewP2PMod(listenAddr config.Address) *P2PMod {
	return &P2PMod{
		listenAddr:    listenAddr,
		ConnMananger:  ConnMananger{connPools: make(map[config.Address]*sync.Pool)},
		MsgHandlerMap: make(map[message.MessageType]message.MessageHandler),
	}
}

func (p2p *P2PMod) RegisterHandler(msgType message.MessageType, handler message.MessageHandler) {
	utils.LoggerInstance.Debug("Registering handler for message type: %v", msgType)
	p2p.MsgHandlerMap[msgType] = handler
}

// start listening on the p2p's listen address
func (p2p *P2PMod) StartListen() {
	utils.LoggerInstance.Info("Start listening on %v\n", p2p.listenAddr)
	ln, err := net.Listen("tcp", p2p.listenAddr)
	if err != nil {
		utils.LoggerInstance.Error("Error listening: %v", err)
		return
	}

	p2p.wg.Add(1)
	go func() {
		defer p2p.wg.Done()
		defer ln.Close()
		for {
			conn, err := ln.Accept()
			if err != nil {
				utils.LoggerInstance.Error("Error accepting: %v", err)
				return
			}
			p2p.wg.Add(1)
			go p2p.handleConnection(conn)
		}
	}()
}

func (p2p *P2PMod) handleConnection(conn net.Conn) {
	defer conn.Close()
	defer p2p.wg.Done()
	reader := bufio.NewReader(conn)
	for {

		content, err := reader.ReadBytes('\n')
		if err == io.EOF {
			utils.LoggerInstance.Warn("Connection closed by peer: %v\n", conn.RemoteAddr())
			return
		} else if err != nil {
			utils.LoggerInstance.Error("Error reading from connection: %v\n", err)
			return
		}

		msg := new(message.Message)
		message.JsonDecode(content, msg)
		utils.LoggerInstance.Debug("Received msg of type %v len %v", msg.MsgType, len(content))

		if handler, ok := p2p.MsgHandlerMap[msg.MsgType]; ok {
			handler(msg) // Q: why use/not use go here?
		} else {
			utils.LoggerInstance.Error("No handler for message type %v\n", msg.MsgType)
		}
	}
}
