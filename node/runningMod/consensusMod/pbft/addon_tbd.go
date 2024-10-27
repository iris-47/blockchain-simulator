package pbft

import (
	"BlockChainSimulator/blockchain"
	"BlockChainSimulator/config"
	"BlockChainSimulator/message"
	"BlockChainSimulator/structs"
	"BlockChainSimulator/utils"
	"sync"
	"time"
)

var ChannelId = 23333

var _ PbftAddon = &PbftTBDAddon{}

// implement TBD method
type PbftTBDAddon struct {
	pbftMod *PbftCosensusMod // the belonging pbft module

	stateUpdateDone bool
	stateLock       sync.Mutex
}

func NewTBDPbftCosensusAddon(pbftMod *PbftCosensusMod) PbftAddon {
	return &PbftTBDAddon{
		pbftMod:         pbftMod,
		stateUpdateDone: false,
	}
}

// no more things to do
func (addon *PbftTBDAddon) HandleProposeAddon(req *message.Request) bool {
	return true
}

// verify the request
func (addon *PbftTBDAddon) HandlePrePrepareAddon(req *message.Request) bool {
	b := &structs.Block{}
	err := utils.Decode(req.Content, b)
	if err != nil {
		utils.LoggerInstance.Error("Error decoding the block")
		return false
	}

	bc := addon.pbftMod.nodeAttr.CurChain
	if b.Header.Height != bc.CurrentBlock.Header.Height+1 {
		utils.LoggerInstance.Warn("the block height is not equal to the current block height")
	} else if string(blockchain.GetTxTreeRoot(b.Transactions)) != string(b.Header.TxRoot) {
		utils.LoggerInstance.Warn("the transaction root is wrong")
	}

	txs := b.Transactions
	utils.LoggerInstance.Debug("The block is legal, try update the state")
	bc.StateManager.UpdateStates(txs, bc.CurrentBlock.Header.StateRoot)

	addon.SetStateUpdateDone(true) // indicate the state update is done

	return true
}

// Wait until the state update is done
func (addon *PbftTBDAddon) HandlePrepareAddon(req *message.Request) bool {
	// This is not the standard practice for production.
	// It's solely to ensure all nodes are synchronized in the same PBFT round.
	for !addon.GetStateUpdateDone() {
		time.Sleep(10 * time.Millisecond) // 休眠10毫秒
	}
	return true
}

func (addon *PbftTBDAddon) HandleCommitAddon(req *message.Request) bool {
	// TODO: implement the commit logic
	for !addon.GetStateUpdateDone() {
		time.Sleep(10 * time.Millisecond) // 休眠10毫秒
	}

	b := &structs.Block{}
	err := utils.Decode(req.Content, b)
	if err != nil {
		utils.LoggerInstance.Error("Error decoding the block")
		return false
	}

	addon.pbftMod.nodeAttr.CurChain.CommitBlock(b)
	addon.SetStateUpdateDone(false)
	// send the verified message back to the client
	if addon.pbftMod.nodeAttr.Nid == config.ViewNodeId {
		reply := &message.Reply{
			Req:  req,
			Time: time.Now(),

			Sid:         addon.pbftMod.nodeAttr.Sid,
			ReqQueueLen: addon.pbftMod.requestQueue.Size(),
		}

		replaymsg := message.Message{
			MsgType: message.MsgReply,
			Content: utils.Encode(reply),
		}

		utils.LoggerInstance.Info("Send the reply message back to the client")
		addon.pbftMod.p2pMod.ConnMananger.Send(config.ClientAddr, replaymsg.JsonEncode())
	}

	return true
}

func (addon *PbftTBDAddon) SetStateUpdateDone(flag bool) {
	addon.stateLock.Lock()
	defer addon.stateLock.Unlock()
	addon.stateUpdateDone = flag
}

func (addon *PbftTBDAddon) GetStateUpdateDone() bool {
	addon.stateLock.Lock()
	defer addon.stateLock.Unlock()
	return addon.stateUpdateDone
}
