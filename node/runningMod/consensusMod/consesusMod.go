// Abandoned
package consensusmod

// Just a rant: in some ways, Go really doesn't measure up to C++. The lack of inheritance and polymorphism

// All consensusMods share some common variables.
// So we could define a structure as follows and specify each consensus as a modular interface.
// However, this adds an overly complex layer of hierarchy, so I opted to discard this approach.
// Maybe one day, when the codebase becomes too large, we can reconsider this solution.
// type ConsensusMod struct {
// 	// vars from the belonging node
// 	nodeAttr *nodeattr.NodeAttr // the attribute of the belonging node
// 	p2pMod   *p2p.P2PMod        // the p2p network module

// 	// consensus related
// 	requestQueue  *utils.Queue[message.Request] // the queue of requests waiting for consensus
// 	requestPool   map[string]*pbft.RequestInfo  // the pool of requests that have been received
// 	consensusDone chan struct{}                 // the channel to notify a round of  consensus is done

// 	// PBFT or HotStuff  or other consensus related module etc.
// 	// ...
// }
