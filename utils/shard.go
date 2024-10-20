package utils

import "BlockChainSimulator/config"

func Addr2Shard(addr config.Address) int {
	subaddr := addr[len(addr)-8:]
	bytes := []byte(subaddr)
	var num int
	for _, b := range bytes {
		num = (num << 8) | int(b)
	}
	if config.ConsensusMethod == "CShard" {
		return num % config.K // Only origin shards has Txs
	} else {
		return num % config.ShardNum
	}
}
