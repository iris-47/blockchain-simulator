package structs

// State is the interface to define the state of the blockchain. including the account state, contract state, etc.
// For broader applicability, UTXOSet is also considered a special type of blockchain state.s
type State interface {
	GetKey() []byte             // get the key of the state to store in the trie, for example, the address of the account
	Update(tx Transaction) bool // verify the tx and update the state
	Commit()                    // commit the state
	Rollback()                  // rollback the state to the previous state
}
