package p2p

import (
	"BlockChainSimulator/config"
	"BlockChainSimulator/utils"
	"net"
	"sync"
)

type ConnMananger struct {
	connPools   map[config.Address]*sync.Pool // address -> connection pool
	connPoolock sync.Mutex                    // mutex for connPools
}

// get the connection pool for the target address, if not exist, create a new one
func (conm *ConnMananger) getPool(addr string) *sync.Pool {
	conm.connPoolock.Lock()
	defer conm.connPoolock.Unlock()

	// check if already have a connection pool for this address
	pool, ok := conm.connPools[addr]
	if !ok {
		// create a new connection pool
		pool = &sync.Pool{
			New: func() interface{} {
				conn, err := net.Dial("tcp", addr)
				if err != nil {
					utils.LoggerInstance.Error("Failed to connect to %s, err: %v", addr, err)
					return nil
				}
				return conn
			},
		}
		conm.connPools[addr] = pool
	}

	return pool
}

// send a message to the target address
func (conm *ConnMananger) Send(addr string, context []byte) {
	pool := conm.getPool(addr)

	connInterface := pool.Get()
	if connInterface == nil {
		utils.LoggerInstance.Error("Failed to get connection from pool")
		return
	}

	conn := pool.Get().(net.Conn)
	defer pool.Put(conn)

	_, err := conn.Write(append(context, '\n'))
	if err != nil {
		conn.Close()
		utils.LoggerInstance.Error("Failed to send message to %s, err: %v", addr, err)

		return
	}
}

func (conm *ConnMananger) Broadcast(sender string, addrs []string, context []byte) {
	for _, addr := range addrs {
		if addr != sender {
			go conm.Send(addr, context)
		}
	}
}
