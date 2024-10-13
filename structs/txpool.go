// Description: Define the struct of TxPool, Node use TxPool to store the transaction in the memory
package structs

import (
	"sync"
	"time"
)

type TxPool struct {
	Txs []Transaction

	batchSize       int
	lock            sync.Mutex
	batchSizeEnough *sync.Cond
}

func NewTxPool(batchSize int) *TxPool {
	txPool := &TxPool{
		Txs: make([]Transaction, 0),

		batchSize: batchSize,
	}
	txPool.batchSizeEnough = sync.NewCond(&txPool.lock)
	return txPool
}

// AddTx adds a transaction to last of the TxPool
func (txPool *TxPool) AddTx(tx Transaction) {
	txPool.lock.Lock()
	defer txPool.lock.Unlock()
	if tx.GetTime().IsZero() {
		tx.SetTime(time.Now())
	}
	txPool.Txs = append(txPool.Txs, tx)

	// notify that there is enough transactions in the pool
	if len(txPool.Txs) >= txPool.batchSize {
		txPool.batchSizeEnough.Signal()
	}
}

// AddTxs adds a list of transactions to last of the TxPool
func (txPool *TxPool) AddTxs(txs []Transaction) {
	txPool.lock.Lock()
	defer txPool.lock.Unlock()
	for _, tx := range txs {
		if tx.GetTime().IsZero() {
			tx.SetTime(time.Now())
		}
		txPool.Txs = append(txPool.Txs, tx)
	}

	// notify that there is enough transactions in the pool
	if len(txPool.Txs) >= txPool.batchSize {
		txPool.batchSizeEnough.Signal()
	}
}

// GetTxs() returns the first `count` transactions from the TxPool.
// If `count` is greater than the number of transactions in the pool, it returns all available transactions.
func (txPool *TxPool) GetTxs(count int) []Transaction {
	txPool.lock.Lock()
	defer txPool.lock.Unlock()

	if count > len(txPool.Txs) {
		count = len(txPool.Txs)
	}

	txs := txPool.Txs[:count]
	txPool.Txs = txPool.Txs[count:]
	return txs
}

// WaitTxs() returns the first `batchSize` transactions from the TxPool.
// If insufficient transactions available, it will block until there are enough transactions.
func (txPool *TxPool) WaitTxs() []Transaction {
	txPool.lock.Lock()
	defer txPool.lock.Unlock()

	for len(txPool.Txs) < txPool.batchSize {
		txPool.batchSizeEnough.Wait()
	}

	txs := txPool.Txs[:txPool.batchSize]
	txPool.Txs = txPool.Txs[txPool.batchSize:]
	return txs
}

func (txPool *TxPool) Size() int {
	txPool.lock.Lock()
	defer txPool.lock.Unlock()
	return len(txPool.Txs)
}

// Notify() is used to stop the waiting goroutine
func (txPool *TxPool) Notify() {
	txPool.lock.Lock()
	defer txPool.lock.Unlock()
	txPool.batchSizeEnough.Signal()
}
