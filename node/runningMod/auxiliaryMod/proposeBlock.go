package auxiliaryMod

// Q: there will be a lot of duplicated code, how to avoid it?

// var _ msgHandlerInterface.MsgHandlerMod = &ProposeBlockAuxiliaryMod{}

// // this mod will receive the txs from client and package them into a block, then propose the block to the shard
// type ProposeBlockAuxiliaryMod struct {
// nodeAttr *nodeattr.NodeAttr // the attribute of the belonging node
// p2pMod   *p2p.P2PMod        // the p2p network module of the belonging node

// 	txPool structs.TxPool
// }
