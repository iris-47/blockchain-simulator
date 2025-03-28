// used to query the result of the consensus, the client node will send a query message to the view nodes
// in some discription of the protocol, the client node need to actively query the result of the consensus instead of waiting for the result
package clientMod

import (
	"BlockChainSimulator/config"
	"BlockChainSimulator/message"
	"BlockChainSimulator/node/nodeattr"
	"BlockChainSimulator/node/p2p"
	"BlockChainSimulator/node/runningMod/runningModInterface"
	"BlockChainSimulator/utils"
	"context"
	"sync"
	"time"
)

type queryMod struct {
	nodeAttr *nodeattr.NodeAttr
	p2pMod   *p2p.P2PMod
}

func NewQueryMod(attr *nodeattr.NodeAttr, p2p *p2p.P2PMod) runningModInterface.RunningMod {
	qm := new(queryMod)
	qm.nodeAttr = attr
	qm.p2pMod = p2p

	return qm
}

func (qm *queryMod) handleInitMsg(msg *message.Message) {
	utils.LoggerInstance.Info("Received an init message")

	startTime := time.Time{}
	err := utils.Decode(msg.Content, &startTime)
	if err != nil {
		utils.LoggerInstance.Error("Error decoding the init message, err: %v", err)
		return
	}

	// wait until f + 1 round, the node commit, client query the result
	maliciousNodes := int64(config.MaliciousRatio * float64(config.NodeNum))
	delayDuration := (maliciousNodes + 1) * config.TickInterval * int64(time.Millisecond)
	commitTime := startTime.Add(time.Duration(delayDuration))
	commitTimer := time.NewTimer(time.Until(commitTime))

	go func() {
		<-commitTimer.C

		// sleep for a while to wait for the consensus to really finish
		time.Sleep(time.Duration(config.TickInterval/4) * time.Millisecond)
		queryMsg := message.Message{
			MsgType: message.MsgQuery,
			Content: utils.Encode(qm.nodeAttr.Ipaddr),
		}
		// send to the view node
		utils.LoggerInstance.Info("Send the query message to the view node")
		qm.p2pMod.ConnMananger.Send(config.IPMap[0][0], queryMsg.JsonEncode())
	}()
}

func (qm *queryMod) handleReplyQueryMsg(msg *message.Message) {
	utils.LoggerInstance.Info("Received a query reply message")

	result := ""
	err := utils.Decode(msg.Content, &result)
	if err != nil {
		utils.LoggerInstance.Error("Error decoding the query reply message, err: %v", err)
		return
	}

	utils.LoggerInstance.Info("The result of the query is: %v", result)
}

func (qm *queryMod) RegisterHandlers() {
	qm.p2pMod.RegisterHandler(message.MsgInit, qm.handleInitMsg)
	qm.p2pMod.RegisterHandler(message.MsgReplyQuery, qm.handleReplyQueryMsg)
}

func (qm *queryMod) Run(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()

	// do nothing
}
