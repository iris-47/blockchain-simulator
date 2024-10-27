package structs

import (
	"BlockChainSimulator/utils"
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"time"
)

var _ Transaction = &ContractTransaction{}

func init() {
	gob.Register(&ContractTransaction{})
}

type ContractTransaction struct {
	// Global
	Sender    Address
	Recipient Address // the address of the contract
	Nounce    int64
	TxHash    []byte

	Time time.Time

	Code            []byte    // the code of the contract
	RelatedContract []Address // the related contract address
	IsCrossShard    bool      // whether the contract will invoke cross-shard call
}

func NewContractTransaction(sender Address, recipient Address, nounce int64, time time.Time, code []byte, relatedContract []Address, isCrossShard bool) *ContractTransaction {
	tx := &ContractTransaction{
		Sender:          sender,
		Recipient:       recipient,
		Nounce:          nounce,
		Time:            time,
		Code:            code,
		RelatedContract: relatedContract,
		IsCrossShard:    isCrossShard,
	}
	tx.TxHash = utils.Hash(utils.Encode(tx))
	return tx
}

func (tx *ContractTransaction) Type() string {
	return ETHLikeContractTransactionType
}

func (tx *ContractTransaction) ID() []byte {
	return tx.TxHash
}

func (tx *ContractTransaction) From() []Address {
	return []Address{tx.Sender}
}

func (tx *ContractTransaction) To() []Address {
	return []Address{tx.Recipient}
}

func (tx *ContractTransaction) GetTime() time.Time {
	return tx.Time
}

func (tx *ContractTransaction) Hash() []byte {
	return tx.TxHash
}

func (tx *ContractTransaction) IsCoinBase() bool {
	return false
}

func (tx *ContractTransaction) GetNonce() int64 {
	return tx.Nounce
}

func (tx *ContractTransaction) SetTime(time time.Time) {
	tx.Time = time
}

func (tx ContractTransaction) String() string {
	str := "[\n"
	str += fmt.Sprintf("\tSender: %s\n", tx.Sender)
	str += fmt.Sprintf("\tRecipient: %s\n", tx.Recipient)
	str += fmt.Sprintf("\tNounce: %d\n", tx.Nounce)
	str += fmt.Sprintf("\tTime: %s\n", tx.Time)
	str += fmt.Sprintf("\tCode: %s\n", hex.EncodeToString(tx.Code))
	str += fmt.Sprintf("\tRelatedContract: %v\n", tx.RelatedContract)
	str += fmt.Sprintf("\tIsCrossShard: %t\n", tx.IsCrossShard)
	str += "]"
	return str
}
