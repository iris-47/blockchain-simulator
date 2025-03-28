// some utility functions for tbb consensus module
package ds

import (
	"BlockChainSimulator/signature"
	"BlockChainSimulator/utils"
	"time"
)

func (dsMod *DSCosensusMod) setStartTime(startTime time.Time) {
	dsMod.startLock.Lock()
	defer dsMod.startLock.Unlock()
	if startTime.After(dsMod.startTime) {
		dsMod.startTime = startTime
	} else if startTime.Before(dsMod.startTime) {
		utils.LoggerInstance.Warn("The startTime is earlier than the current one, something wrong")
		return
	} else {
		return
	}
}

func (dsMod *DSCosensusMod) getStartTime() time.Time {
	dsMod.startLock.Lock()
	defer dsMod.startLock.Unlock()
	return dsMod.startTime
}

// check the signature of SigListContent
// TODO: implement this function
// 需要前置完成密钥广播模块，尚未完成
func (dsMod *DSCosensusMod) checkSigList([]*signature.Signature) bool {
	return true
}
