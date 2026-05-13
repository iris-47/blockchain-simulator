// node/p2p/p2p.go
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
	listenAddr    config.Address
	MsgHandlerMap map[message.MessageType]message.MessageHandler
	ConnManager   *ConnManager

	listener net.Listener
	wg       sync.WaitGroup
	stopChan chan struct{}
}

func NewP2PMod(listenAddr config.Address) *P2PMod {
	return &P2PMod{
		listenAddr:    listenAddr,
		ConnManager:   NewConnManager(),
		MsgHandlerMap: make(map[message.MessageType]message.MessageHandler),
		stopChan:      make(chan struct{}),
	}
}

func (p2p *P2PMod) RegisterHandler(msgType message.MessageType, handler message.MessageHandler) {
	utils.LoggerInstance.Debug("Registering handler for message type: %v", msgType)
	p2p.MsgHandlerMap[msgType] = handler
}

// StartListen starts listening on the p2p's listen address
func (p2p *P2PMod) StartListen() error {
	_, port, err := net.SplitHostPort(p2p.listenAddr)
	if err != nil {
		utils.LoggerInstance.Error("Error parsing listen address: %v", err)
		return err
	}

	if port == "" {
		utils.LoggerInstance.Error("Port is required in listen address: %v", p2p.listenAddr)
		return err
	}

	utils.LoggerInstance.Info("Start listening on port %v", port)
	listenAddr := net.JoinHostPort("0.0.0.0", port)

	ln, err := net.Listen("tcp", listenAddr)
	if err != nil {
		utils.LoggerInstance.Error("Error listening: %v", err)
		return err
	}

	p2p.listener = ln

	p2p.wg.Add(1)
	go func() {
		defer p2p.wg.Done()
		defer ln.Close()

		for {
			conn, err := ln.Accept()
			if err != nil {
				select {
				case <-p2p.stopChan:
					return
				default:
					utils.LoggerInstance.Error("Error accepting: %v", err)
					return
				}
			}

			p2p.wg.Add(1)
			go p2p.handleConnection(conn)
		}
	}()

	return nil
}

func (p2p *P2PMod) handleConnection(conn net.Conn) {
	defer p2p.wg.Done()

	reader := bufio.NewReader(conn)
	var senderAddr string
	connStored := false

	defer func() {
		if !connStored {
			conn.Close()
		}
	}()

	for {
		content, err := reader.ReadBytes('\n')
		if err == io.EOF {
			return
		} else if err != nil {
			utils.LoggerInstance.Error("Error reading from connection: %v", err)
			return
		}

		msg := new(message.Message)
		if err := message.JsonDecode(content, msg); err != nil {
			utils.LoggerInstance.Error("Error decoding message: %v", err)
			continue
		}

		// store the connection using sender's address for bidirectional communication
		if !connStored && msg.Sender != "" {
			senderAddr = msg.Sender
			p2p.ConnManager.storeConn(senderAddr, conn)
			connStored = true
		}

		if handler, ok := p2p.MsgHandlerMap[msg.MsgType]; ok {
			// handle message in a new goroutine to avoid blocking the read loop
			go handler(msg)
		} else {
			utils.LoggerInstance.Error("No handler for message type %v", msg.MsgType)
		}
	}
}

// Stop gracefully stops the P2P module
func (p2p *P2PMod) Stop() {
	close(p2p.stopChan)

	if p2p.listener != nil {
		p2p.listener.Close()
	}

	p2p.ConnManager.Close()
	p2p.wg.Wait()
	utils.LoggerInstance.Info("P2P module stopped")
}
