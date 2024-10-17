// I create this package because package msgHandler and package pbft, auxiliaryHandler, etc. are all related to MsgHandlerMod interface.
// Yeah, it's a little bit weird, but I can't find a better way to organize the code.
package runningModInterface

import (
	"context"
	"sync"
)

// Q: Is there a better way to organize the code?
type RunningMod interface {
	RegisterHandlers()                           // register the message handlers to p2pMod
	Run(ctx context.Context, wg *sync.WaitGroup) // run the module, ctx is used if the mod needs to do something when the node is shutting down
}
