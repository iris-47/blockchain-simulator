package consensusMod

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

var _ runningModInterface.RunningMod = &ProposeStringAuxiliaryMod{}

// this mod will receive the string from client
type ProposeStringAuxiliaryMod struct {
	nodeAttr *nodeattr.NodeAttr // the attribute of the belonging node
	p2pMod   *p2p.P2PMod        // the p2p network module of the belonging node

	// for multi-round consensus case
	stringQueue   *utils.Queue[string] // the queue of requests waiting for consensus
	consensusDone chan struct{}
}

// this mod will receive the txs from client and propose them to the shard
func NewProposeStringAuxiliaryMod(attr *nodeattr.NodeAttr, p2p *p2p.P2PMod) runningModInterface.RunningMod {
	sam := new(ProposeStringAuxiliaryMod)
	sam.nodeAttr = attr
	sam.p2pMod = p2p

	sam.stringQueue = utils.NewQueue[string]()
	sam.consensusDone = make(chan struct{}, 1)

	return sam
}

// receive the request from the client, and add it to the request queue
func (sam *ProposeStringAuxiliaryMod) handleInject(msg *message.Message) {
	utils.LoggerInstance.Debug("handle inject")

	str := ""
	err := utils.Decode(msg.Content, &str)

	if err != nil {
		utils.LoggerInstance.Error("error decoding the inject message")
		return
	}

	sam.stringQueue.Enqueue(str)
}

func (sam *ProposeStringAuxiliaryMod) handleConsensusDone(_ *message.Message) {
	// non-blocking send to the channel
	select {
	case sam.consensusDone <- struct{}{}:
	default:
	}
}

func (sam *ProposeStringAuxiliaryMod) RegisterHandlers() {
	sam.p2pMod.RegisterHandler(message.MsgInject, sam.handleInject)
	sam.p2pMod.RegisterHandler(message.MsgConsensusDone, sam.handleConsensusDone)
}

// get the txs from the txPool and propose them to the shard
func (sam *ProposeStringAuxiliaryMod) Run(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()

	for {
		select {
		case <-ctx.Done():
			utils.LoggerInstance.Info("Stop the consensus Mod")
			return
		default:
			// get the request from the request queue every 200ms, the request is from client by InjectMsg
			str, err := sam.stringQueue.Dequeue()
			if err != nil {
				time.Sleep(200 * time.Millisecond)
				continue
			}

			// set the start time
			startTime := time.Now().Add(time.Duration(config.StartTimeWait) * time.Millisecond)

			initMsg := message.Message{
				MsgType: message.MsgInit,
				Content: utils.Encode(startTime),
			}
			utils.LoggerInstance.Info("Broadcast the init message")
			sam.p2pMod.ConnMananger.Broadcast(sam.nodeAttr.Ipaddr, utils.GetNeighbours(config.IPMap[0], sam.nodeAttr.Ipaddr), initMsg.JsonEncode())
			// alse send to client to sync time
			sam.p2pMod.ConnMananger.Send(config.ClientAddr, initMsg.JsonEncode())

			sam.p2pMod.MsgHandlerMap[message.MsgInit](&initMsg)

			// wait to start the protocol at the startTime
			startTimer := time.NewTimer(time.Until(startTime))
			<-startTimer.C

			sig := signature.Sign(sam.nodeAttr.SecKey, []byte(str))
			req := message.NewRequestWithSignature(sam.nodeAttr.Sid, message.ReqVerifyString, []byte(str), sig)

			proposeMsg := message.Message{
				MsgType: message.MsgPropose,
				Content: utils.Encode(req),
			}

			badViewNode := true

			if badViewNode {
				badproposeMsg := message.Message{
					MsgType: message.MsgPropose,
					Content: utils.Encode(message.NewRequestWithSignature(sam.nodeAttr.Sid, message.ReqVerifyString, []byte("bad"), sig)),
				}
				utils.LoggerInstance.Info("Broadcast the propose message")
				// sam.p2pMod.ConnMananger.Broadcast(sam.nodeAttr.Ipaddr, utils.GetNeighbours(config.IPMap[0], sam.nodeAttr.Ipaddr), proposeMsg.JsonEncode())
				sam.p2pMod.MsgHandlerMap[message.MsgPropose](&proposeMsg)

				sam.p2pMod.ConnMananger.Send(config.IPMap[0][1], badproposeMsg.JsonEncode())
				sam.p2pMod.ConnMananger.Send(config.IPMap[0][2], badproposeMsg.JsonEncode())
			} else {
				utils.LoggerInstance.Info("Broadcast the propose message")
				sam.p2pMod.ConnMananger.Broadcast(sam.nodeAttr.Ipaddr, utils.GetNeighbours(config.IPMap[0], sam.nodeAttr.Ipaddr), proposeMsg.JsonEncode())
				sam.p2pMod.MsgHandlerMap[message.MsgPropose](&proposeMsg)
			}
			// wait for the consensus to be done
			select {
			case <-sam.consensusDone:
				// consensus interval, waits for other nodes to complete the consensus
				// time.Sleep(time.Millisecond * time.Duration(config.ConsensusInterval))
				utils.LoggerInstance.Info("Consensus is done, go next round")
			case <-ctx.Done():
				utils.LoggerInstance.Info("Stop the consensus Mod")
				return
			}
		}
	}
}
