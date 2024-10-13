package message

// // Description: a simple version of connection pool for reusing connections.
// // Opt: consider using sync.Pool instead
// package networks

// import (
// 	"net"
// 	"sync"
// )

// type ConnectionPool struct {
// 	pool map[string][]net.Conn
// 	lock sync.Mutex // mutex for []net.Conn
// }

// func NewConnectionPool() *ConnectionPool {
// 	return &ConnectionPool{
// 		pool: make(map[string][]net.Conn),
// 	}
// }

// // Get connection from the pool, or create a new one.
// func (cp *ConnectionPool) GetConnection(addr string) (net.Conn, error) {
// 	cp.lock.Lock()
// 	defer cp.lock.Unlock()

// 	// Check if we already have a connection for this address
// 	if connections, ok := cp.pool[addr]; ok && len(connections) > 0 {
// 		// Return the last connection and remove it from the list
// 		c := connections[len(connections)-1]
// 		cp.pool[addr] = connections[:len(connections)-1]
// 		return c, nil
// 	}

// 	// Create a new connection if none are available
// 	newConn, err := net.Dial("tcp", addr)
// 	if err != nil {
// 		return nil, err
// 	}
// 	return newConn, nil
// }

// // Return connection back to the pool.
// func (cp *ConnectionPool) PutConnection(addr string, conn net.Conn) {
// 	cp.lock.Lock()
// 	defer cp.lock.Unlock()

// 	// Add the connection back to the pool
// 	cp.pool[addr] = append(cp.pool[addr], conn)
// }
