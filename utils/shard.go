package utils

import (
	"BlockChainSimulator/config"
	"strings"
)

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

func GetNeighbours(IPs map[int]string, SelfIP string) []string {
	neighbours := make([]string, 0)
	for _, ip := range IPs {
		if ip == SelfIP {
			continue
		}
		neighbours = append(neighbours, ip)
	}
	return neighbours
}

// It assumes the input string is in the format "IP:Port" and returns the IP part.
func GetIP(IP string) string {
	parts := strings.Split(IP, ":")
	if len(parts) > 0 {
		return parts[0]
	}
	return "" // Return an empty string if no IP part is found
}
