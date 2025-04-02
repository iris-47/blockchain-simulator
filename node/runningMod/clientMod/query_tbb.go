// used to query the result of the consensus, the client node will send a query message to the view nodes
// in some discription of the protocol, the client node need to actively query the result of the consensus instead of waiting for the result
package clientMod

import (
	"BlockChainSimulator/config"
	"BlockChainSimulator/message"
	"BlockChainSimulator/node/nodeattr"
	"BlockChainSimulator/node/p2p"
	"BlockChainSimulator/node/runningMod/runningModInterface"
	"BlockChainSimulator/signature"
	"BlockChainSimulator/utils"
	"context"
	"sync"
	"time"
)

type queryTBBMod struct {
	nodeAttr  *nodeattr.NodeAttr
	p2pMod    *p2p.P2PMod
	startTime *utils.AtomicValue[time.Time]

	latencys []time.Duration
}

type ReplyValue struct {
	Value string
	QCMap map[string]*signature.Signature
}

func NewQueryTBBMod(attr *nodeattr.NodeAttr, p2p *p2p.P2PMod) runningModInterface.RunningMod {
	qtm := new(queryTBBMod)
	qtm.nodeAttr = attr
	qtm.p2pMod = p2p
	qtm.startTime = utils.NewAtomicValue[time.Time](time.Now())

	return qtm
}

func (qtm *queryTBBMod) handleInitMsg(msg *message.Message) {
	utils.LoggerInstance.Info("Received an init message")

	startTime := time.Time{}
	err := utils.Decode(msg.Content, &startTime)
	if err != nil {
		utils.LoggerInstance.Error("Error decoding the init message, err: %v", err)
		return
	}

	qtm.startTime.Set(startTime, nil)

	t2 := config.NodeNum - 1

	aggresiveTimer := time.NewTimer(time.Until(startTime.Add(time.Duration(3*config.TickInterval) * time.Millisecond)))
	conservitiveTimer := time.NewTimer(time.Until(startTime.Add(time.Duration(int64(t2+6)*config.TickInterval) * time.Millisecond)))

	utils.LoggerInstance.Debug("Set the aggresive timer to %v", aggresiveTimer.C)
	utils.LoggerInstance.Debug("Set the conservitive timer to %v", conservitiveTimer.C)

	go qtm.sendQueryOnTimeout(aggresiveTimer)
	go qtm.sendQueryOnTimeout(conservitiveTimer)
}

func (qtm *queryTBBMod) handleReplyQueryMsg(msg *message.Message) {
	utils.LoggerInstance.Info("Received a query reply message")

	replyValue := ReplyValue{}
	err := utils.Decode(msg.Content, &replyValue)
	if err != nil {
		utils.LoggerInstance.Error("Error decoding the query reply message, err: %v", err)
		return
	}

	t1 := int(float64(config.NodeNum)*config.ResilientRatio) - 1
	t2 := config.NodeNum - 1

	timeNow := time.Since(qtm.startTime.Get())

	// point1
	if timeNow < time.Duration(int64(t1+6)*config.TickInterval)*time.Millisecond {
		utils.LoggerInstance.Info("Point1: aggresive client receive the reply message of %v", replyValue)
		if len(replyValue.QCMap) <= 1 {
			if replyValue.Value != "" {
				utils.LoggerInstance.Info("Point1: aggresive client confirm: %v", replyValue.Value)
				qtm.latencys = append(qtm.latencys, timeNow)
				utils.LoggerInstance.Info("Latency1: %vs", timeNow.Seconds())
			} else { // if D¡ = 0
				sendQueryTimer := time.NewTimer(time.Until(qtm.startTime.Get().Add(time.Duration(int64(t1+6)*config.TickInterval) * time.Millisecond)))
				go qtm.sendQueryOnTimeout(sendQueryTimer)
			}
		} else {
			utils.LoggerInstance.Info("aggresive client switch to conservitive mode")
		}
		return
	}

	// point2
	if timeNow < time.Duration(int64(t2+6)*config.TickInterval)*time.Millisecond {
		utils.LoggerInstance.Info("Point2: aggresive client receive the reply message of %v", replyValue)
		utils.LoggerInstance.Info("Point2: aggresive client confirm: %v", replyValue.Value)
		if replyValue.Value != "" {
			qtm.latencys = append(qtm.latencys, timeNow)
			utils.LoggerInstance.Info("Latency2: %vs", timeNow.Seconds())
		}
		return
	}

	// point3
	if len(qtm.latencys) == 0 {
		qtm.latencys = append(qtm.latencys, timeNow)
	}
	utils.LoggerInstance.Info("Point3: conservitive client receive the reply message of %v", replyValue)
	utils.LoggerInstance.Info("Point3: conservitive client confirm: %v", replyValue.Value)

	utils.LoggerInstance.Info("Latency3: %vs", timeNow.Seconds())
}

func (qtm *queryTBBMod) RegisterHandlers() {
	qtm.p2pMod.RegisterHandler(message.MsgInit, qtm.handleInitMsg)
	qtm.p2pMod.RegisterHandler(message.MsgReplyQuery, qtm.handleReplyQueryMsg)
}

func (qtm *queryTBBMod) Run(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()

	<-ctx.Done()
	// utils.LoggerInstance.Info("The latencys are:")
	// for _, latency := range qtm.latencys {
	// 	utils.LoggerInstance.Info("%v", latency)
	// }
	// do nothing
}

func (qtm *queryTBBMod) sendQueryOnTimeout(queryTimer *time.Timer) {
	<-queryTimer.C

	// wait another half config.TickInterval to make sure the consensus is finished
	time.Sleep(time.Duration(config.TickInterval/2) * time.Millisecond)

	queryMsg := message.Message{
		MsgType: message.MsgQuery,
		Content: utils.Encode(qtm.nodeAttr.Ipaddr),
	}
	utils.LoggerInstance.Info("Send the query message to the view node")
	qtm.p2pMod.ConnMananger.Send(config.IPMap[0][0], queryMsg.JsonEncode())
}
