package p2p

import (
	"BlockChainSimulator/config"
	"net"
	"time"
)

// Wait for all the nodes in the shard to be ready, usually invoked by the view node
func WaitForShardReady(sid int, timeout time.Duration) bool {
	start := time.Now()

	for i, ip := range config.IPMap[sid] {
		// skip the view node
		if i == config.ViewNodeId {
			continue
		}
		for !PortListening(ip) {
			if time.Since(start) > timeout {
				return false
			}
			time.Sleep(time.Second)
		}
	}
	return true
}

// wait for all the nodes in the network to be ready, usually invoked by client
func WaitForAllIPsReady(timeout time.Duration) bool {
	start := time.Now()

	for _, shardips := range config.IPMap {
		for _, ip := range shardips {
			for !PortListening(ip) {
				if time.Since(start) > timeout {
					return false
				}
				time.Sleep(time.Second)
			}
		}
	}

	return true
}

// check if the port is listening
func PortListening(address string) bool {
	conn, err := net.DialTimeout("tcp", address, 100*time.Millisecond)
	if err != nil {
		return false
	}
	conn.Close()
	return true
}
