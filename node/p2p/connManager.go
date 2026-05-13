// node/p2p/connManager.go
package p2p

import (
	"BlockChainSimulator/utils"
	"net"
	"sync"
	"time"
)

type ConnManager struct {
	conns     map[string]net.Conn // address -> persistent connection
	connMutex sync.RWMutex
}

func NewConnManager() *ConnManager {
	return &ConnManager{
		conns: make(map[string]net.Conn),
	}
}

// getOrCreateConn gets existing connection or creates a new one
func (cm *ConnManager) getOrCreateConn(addr string) (net.Conn, error) {
	// try to get existing connection
	cm.connMutex.RLock()
	conn, exists := cm.conns[addr]
	cm.connMutex.RUnlock()

	if exists && conn != nil {
		// verify connection is still alive by checking if we can set deadline
		if err := conn.SetWriteDeadline(time.Now().Add(time.Second)); err == nil {
			conn.SetWriteDeadline(time.Time{}) // reset deadline
			return conn, nil
		}
		// connection is dead, remove it
		cm.removeConn(addr)
	}

	// create new connection
	return cm.createConn(addr)
}

// createConn establishes a new TCP connection with keepalive
func (cm *ConnManager) createConn(addr string) (net.Conn, error) {
	cm.connMutex.Lock()
	defer cm.connMutex.Unlock()

	// double-check: another goroutine might have created the connection
	if conn, exists := cm.conns[addr]; exists && conn != nil {
		return conn, nil
	}

	conn, err := net.DialTimeout("tcp", addr, 5*time.Second)
	if err != nil {
		return nil, err
	}

	// enable TCP keepalive
	if tcpConn, ok := conn.(*net.TCPConn); ok {
		tcpConn.SetKeepAlive(true)
		tcpConn.SetKeepAlivePeriod(30 * time.Second)
	}

	cm.conns[addr] = conn
	utils.LoggerInstance.Debug("Created new connection to %s", addr)
	return conn, nil
}

// storeConn stores an incoming connection (from remote peer)
func (cm *ConnManager) storeConn(addr string, conn net.Conn) {
	cm.connMutex.Lock()
	defer cm.connMutex.Unlock()

	// close old connection if exists
	if oldConn, exists := cm.conns[addr]; exists && oldConn != nil {
		oldConn.Close()
	}

	// enable TCP keepalive
	if tcpConn, ok := conn.(*net.TCPConn); ok {
		tcpConn.SetKeepAlive(true)
		tcpConn.SetKeepAlivePeriod(30 * time.Second)
	}

	cm.conns[addr] = conn
	utils.LoggerInstance.Debug("Stored incoming connection from %s", addr)
}

// removeConn removes and closes a connection
func (cm *ConnManager) removeConn(addr string) {
	cm.connMutex.Lock()
	defer cm.connMutex.Unlock()

	if conn, exists := cm.conns[addr]; exists {
		conn.Close()
		delete(cm.conns, addr)
		utils.LoggerInstance.Debug("Removed connection to %s", addr)
	}
}

// Send sends a message to the target address
func (cm *ConnManager) Send(addr string, context []byte) error {
	conn, err := cm.getOrCreateConn(addr)
	if err != nil {
		utils.LoggerInstance.Error("Failed to get connection to %s: %v", addr, err)
		return err
	}

	// set write deadline to detect broken connections
	conn.SetWriteDeadline(time.Now().Add(5 * time.Second))
	defer conn.SetWriteDeadline(time.Time{}) // reset deadline

	_, err = conn.Write(append(context, '\n'))
	if err != nil {
		utils.LoggerInstance.Error("Failed to send message to %s: %v", addr, err)
		cm.removeConn(addr) // remove broken connection
		return err
	}

	return nil
}

// Broadcast sends a message to multiple addresses
func (cm *ConnManager) Broadcast(sender string, addrs []string, context []byte) {
	var wg sync.WaitGroup
	for _, addr := range addrs {
		if addr != sender {
			wg.Add(1)
			go func(targetAddr string) {
				defer wg.Done()
				cm.Send(targetAddr, context)
			}(addr)
		}
	}
	wg.Wait()
}

// Close closes all connections
func (cm *ConnManager) Close() {
	cm.connMutex.Lock()
	defer cm.connMutex.Unlock()

	for addr, conn := range cm.conns {
		conn.Close()
		utils.LoggerInstance.Debug("Closed connection to %s", addr)
	}
	cm.conns = make(map[string]net.Conn)
}
