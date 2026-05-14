package main

import (
	"BlockChainSimulator/config"
	"BlockChainSimulator/message"
	"BlockChainSimulator/node/p2p"
	"BlockChainSimulator/structs"
	"BlockChainSimulator/utils"
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"math/rand"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync/atomic"
	"time"
)

type httpResult struct {
	OK    bool        `json:"ok"`
	Error string      `json:"error,omitempty"`
	Data  interface{} `json:"data,omitempty"`
}

var reqCounter uint64

func (c *STCClientPlugin) setupRoutes() {}

func (c *STCClientPlugin) routes() http.Handler {
	m := http.NewServeMux()
	m.HandleFunc("/", c.handleIndex)
	m.HandleFunc("/static/style.css", c.handleStyle)
	m.HandleFunc("/static/app.js", c.handleAppJS)

	m.HandleFunc("/api/status", c.handleStatus)
	m.HandleFunc("/api/metrics", c.handleMetrics)
	m.HandleFunc("/api/logs", c.handleNodeLogs)
	m.HandleFunc("/api/tx/send", c.handleSendTx)
	m.HandleFunc("/api/test/throughput", c.handleThroughputTest)
	m.HandleFunc("/api/control/start-node", c.handleStartNode)
	m.HandleFunc("/api/control/stop-node", c.handleStopNode)
	m.HandleFunc("/api/control/start-network", c.handleStartNetwork)
	m.HandleFunc("/api/control/stop-network", c.handleStopNetwork)
	m.HandleFunc("/api/control/forge-spacetime", c.handleForgeSpaceTime)
	m.HandleFunc("/api/control/byzantine", c.handleByzantine)
	m.HandleFunc("/api/query/block-hashes", c.handleBlockHashes)
	return m
}

func (c *STCClientPlugin) writeJSON(w http.ResponseWriter, code int, res httpResult) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(res)
}

func (c *STCClientPlugin) handleIndex(w http.ResponseWriter, _ *http.Request) {
	data, err := staticAssets.ReadFile("static/index.html")
	if err != nil {
		c.writeJSON(w, http.StatusInternalServerError, httpResult{OK: false, Error: err.Error()})
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write(data)
}

func (c *STCClientPlugin) handleStyle(w http.ResponseWriter, _ *http.Request) {
	data, err := staticAssets.ReadFile("static/style.css")
	if err != nil {
		c.writeJSON(w, http.StatusInternalServerError, httpResult{OK: false, Error: err.Error()})
		return
	}
	w.Header().Set("Content-Type", "text/css; charset=utf-8")
	_, _ = w.Write(data)
}

func (c *STCClientPlugin) handleAppJS(w http.ResponseWriter, _ *http.Request) {
	data, err := staticAssets.ReadFile("static/app.js")
	if err != nil {
		c.writeJSON(w, http.StatusInternalServerError, httpResult{OK: false, Error: err.Error()})
		return
	}
	w.Header().Set("Content-Type", "application/javascript; charset=utf-8")
	_, _ = w.Write(data)
}

func (c *STCClientPlugin) handleStatus(w http.ResponseWriter, _ *http.Request) {
	type nodeState struct {
		ShardID int    `json:"shardId"`
		NodeID  int    `json:"nodeId"`
		IP      string `json:"ip"`
		Online  bool   `json:"online"`
	}
	states := make([]nodeState, 0)
	for sid, nodes := range config.IPMap {
		if sid == config.ClientShard {
			continue
		}
		for nid, ip := range nodes {
			states = append(states, nodeState{ShardID: sid, NodeID: nid, IP: ip, Online: p2p.PortListening(ip)})
		}
	}
	c.writeJSON(w, http.StatusOK, httpResult{OK: true, Data: states})
}

func (c *STCClientPlugin) handleMetrics(w http.ResponseWriter, _ *http.Request) {
	c.cacheMu.Lock()
	defer c.cacheMu.Unlock()
	resp := make(map[string]STCQueryReply)
	for k, v := range c.latestStats {
		resp[k] = v
	}
	c.writeJSON(w, http.StatusOK, httpResult{OK: true, Data: resp})
}

func (c *STCClientPlugin) handleNodeLogs(w http.ResponseWriter, r *http.Request) {
	sid, _ := strconv.Atoi(r.URL.Query().Get("shardId"))
	nid, _ := strconv.Atoi(r.URL.Query().Get("nodeId"))
	reply, err := c.queryNode(sid, nid, STCQuery{Action: "node_overview"}, 2*time.Second)
	if err != nil {
		c.writeJSON(w, http.StatusBadGateway, httpResult{OK: false, Error: err.Error()})
		return
	}
	c.writeJSON(w, http.StatusOK, httpResult{OK: true, Data: reply.Events})
}

func (c *STCClientPlugin) handleSendTx(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		c.writeJSON(w, http.StatusMethodNotAllowed, httpResult{OK: false, Error: "method not allowed"})
		return
	}
	tx := makeSyntheticTx(time.Now().UnixNano(), "manual")
	msg := message.Message{Sender: c.nodeAttr.Ipaddr, MsgType: message.MsgInject, Content: utils.Encode([]structs.Transaction{tx})}
	for sid, nodes := range config.IPMap {
		if sid == config.ClientShard {
			continue
		}
		for _, ip := range nodes {
			_ = c.p2pMod.ConnManager.Send(ip, msg.JsonEncode())
		}
	}
	c.writeJSON(w, http.StatusOK, httpResult{OK: true})
}

func (c *STCClientPlugin) handleThroughputTest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		c.writeJSON(w, http.StatusMethodNotAllowed, httpResult{OK: false, Error: "method not allowed"})
		return
	}
	var req struct {
		TPS      int    `json:"tps"`
		Duration int    `json:"durationSec"`
		CSVPath  string `json:"csvPath"`
	}
	_ = decodeBody(r, &req)
	if req.TPS <= 0 {
		req.TPS = 100000
	}
	if req.Duration <= 0 {
		req.Duration = 180
	}
	if req.CSVPath == "" {
		req.CSVPath = config.FileInput
	}

	go c.runThroughputTest(req.TPS, req.Duration, req.CSVPath)
	c.writeJSON(w, http.StatusOK, httpResult{OK: true, Data: req})
}

func (c *STCClientPlugin) runThroughputTest(tps int, durationSec int, csvPath string) {
	rows := c.readCSVRows(csvPath)

	ctl := STCControlPayload{Action: "set_target_tps", TargetTPS: int64(tps), DurationSec: int64(durationSec)}
	for sid, nodes := range config.IPMap {
		if sid == config.ClientShard {
			continue
		}
		for _, ip := range nodes {
			m := message.Message{Sender: c.nodeAttr.Ipaddr, MsgType: MsgSTCControl, Content: utils.Encode(ctl)}
			_ = c.p2pMod.ConnManager.Send(ip, m.JsonEncode())
		}
	}

	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()
	deadline := time.Now().Add(time.Duration(durationSec) * time.Second)
	batch := tps / 10
	if batch < 1 {
		batch = 1
	}

	idx := int64(0)
	for time.Now().Before(deadline) {
		<-ticker.C
		txs := make([]structs.Transaction, 0, batch)
		for i := 0; i < batch; i++ {
			seed := atomic.AddUint64(&reqCounter, 1)
			payload := "synthetic"
			if len(rows) > 0 {
				payload = rows[idx%int64(len(rows))]
			}
			txs = append(txs, makeSyntheticTx(int64(seed), payload))
			idx++
		}
		msg := message.Message{Sender: c.nodeAttr.Ipaddr, MsgType: message.MsgInject, Content: utils.Encode(txs)}
		for sid, nodes := range config.IPMap {
			if sid == config.ClientShard {
				continue
			}
			for _, ip := range nodes {
				_ = c.p2pMod.ConnManager.Send(ip, msg.JsonEncode())
			}
		}
	}
}

func (c *STCClientPlugin) readCSVRows(path string) []string {
	f, err := os.Open(path)
	if err != nil {
		return nil
	}
	defer f.Close()

	rows := make([]string, 0, 1000)
	s := bufio.NewScanner(f)
	for s.Scan() {
		line := strings.TrimSpace(s.Text())
		if line != "" {
			rows = append(rows, line)
		}
	}
	if len(rows) == 0 {
		return nil
	}
	return rows
}

func makeSyntheticTx(seed int64, payload string) structs.Transaction {
	if config.TxType == structs.UTXOTransactionType {
		tx := &structs.UTXOTransaction{
			TxId:       []byte(fmt.Sprintf("%d-%s", seed, payload)),
			Vin:        []structs.TxIn{{Addr: fmt.Sprintf("from-%d", seed), Value: *big.NewFloat(1)}},
			Vout:       []structs.TxOut{{Addr: fmt.Sprintf("to-%d", seed), Value: *big.NewFloat(1)}},
			Nonce:      seed,
			IsCoinbase: false,
			Time:       time.Now(),
		}
		return tx
	}
	return structs.NewAccountTransaction(
		fmt.Sprintf("from-%d", seed),
		fmt.Sprintf("to-%d", seed),
		seed,
		big.NewInt(int64(rand.Intn(1000)+1)),
	)
}

func (c *STCClientPlugin) handleStartNode(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		c.writeJSON(w, http.StatusMethodNotAllowed, httpResult{OK: false, Error: "method not allowed"})
		return
	}
	var req struct {
		ShardID int `json:"shardId"`
		NodeID  int `json:"nodeId"`
	}
	_ = decodeBody(r, &req)
	c.startNode(req.ShardID, req.NodeID)
	c.writeJSON(w, http.StatusOK, httpResult{OK: true})
}

func (c *STCClientPlugin) handleStopNode(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		c.writeJSON(w, http.StatusMethodNotAllowed, httpResult{OK: false, Error: "method not allowed"})
		return
	}
	var req struct {
		ShardID int `json:"shardId"`
		NodeID  int `json:"nodeId"`
	}
	_ = decodeBody(r, &req)
	ip, ok := config.IPMap[req.ShardID][req.NodeID]
	if !ok {
		c.writeJSON(w, http.StatusBadRequest, httpResult{OK: false, Error: "node not found"})
		return
	}
	m := message.Message{Sender: c.nodeAttr.Ipaddr, MsgType: message.MsgStop}
	_ = c.p2pMod.ConnManager.Send(ip, m.JsonEncode())
	c.writeJSON(w, http.StatusOK, httpResult{OK: true})
}

func (c *STCClientPlugin) handleStartNetwork(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		c.writeJSON(w, http.StatusMethodNotAllowed, httpResult{OK: false, Error: "method not allowed"})
		return
	}
	for sid := 0; sid < config.ShardNum; sid++ {
		for nid := 0; nid < config.NodeNum; nid++ {
			c.startNode(sid, nid)
		}
	}
	c.writeJSON(w, http.StatusOK, httpResult{OK: true})
}

func (c *STCClientPlugin) handleStopNetwork(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		c.writeJSON(w, http.StatusMethodNotAllowed, httpResult{OK: false, Error: "method not allowed"})
		return
	}
	for sid := 0; sid < config.ShardNum; sid++ {
		for nid := 0; nid < config.NodeNum; nid++ {
			m := message.Message{Sender: c.nodeAttr.Ipaddr, MsgType: message.MsgStop}
			_ = c.p2pMod.ConnManager.Send(config.IPMap[sid][nid], m.JsonEncode())
		}
	}
	c.writeJSON(w, http.StatusOK, httpResult{OK: true})
}

func (c *STCClientPlugin) startNode(shardID int, nodeID int) {
	ipPort, ok := config.IPMap[shardID][nodeID]
	if !ok {
		utils.LoggerInstance.Error("IP not found for shard %d node %d", shardID, nodeID)
		return
	}
	ip := strings.Split(ipPort, ":")[0]
	processName := config.ProcessName
	remoteCmd := "cd " + config.RemoteWorkDir + " && nohup ./" + processName +
		" -b " + strconv.Itoa(config.BlockSize) +
		" -S " + strconv.Itoa(config.ShardNum) + " -N " + strconv.Itoa(config.NodeNum) +
		" -s " + strconv.Itoa(shardID) + " -n " + strconv.Itoa(nodeID) +
		" -m " + config.ConsensusMethod +
		" -l " + config.LogLevel + " -t " + config.TxType +
		" -r " + strconv.FormatFloat(config.MaliciousRatio, 'f', -3, 64) +
		" -R " + strconv.FormatFloat(config.ResilientRatio, 'f', -3, 64)
	if config.IsDistributed {
		remoteCmd += " -d "
	}
	if float64(shardID*config.NodeNum+nodeID) > (1-config.MaliciousRatio)*float64(config.ShardNum*config.NodeNum) {
		remoteCmd += " -M "
	}
	if config.ConnectRemoteDemo {
		remoteCmd += " -C "
	}
	remoteCmd += " > /dev/null 2>&1 &"
	sshCmd := "ssh -i " + config.SSHKeyPath + " -o StrictHostKeyChecking=no " + config.SSHUser + "@" + ip + " \"" + remoteCmd + "\""
	cmd := exec.Command("bash", "-c", sshCmd)
	if err := cmd.Start(); err != nil {
		utils.LoggerInstance.Error("Error starting node %d in shard %d on %s: %v", nodeID, shardID, ip, err)
	}
}

func (c *STCClientPlugin) handleForgeSpaceTime(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		c.writeJSON(w, http.StatusMethodNotAllowed, httpResult{OK: false, Error: "method not allowed"})
		return
	}
	var req struct {
		ShardID       int    `json:"shardId"`
		NodeID        int    `json:"nodeId"`
		TimeOffsetSec int64  `json:"timeOffsetSec"`
		FakeLocation  string `json:"fakeLocation"`
	}
	_ = decodeBody(r, &req)
	ctl := STCControlPayload{Action: "forge_next_spacetime", TimeOffsetSec: req.TimeOffsetSec, FakeLocation: req.FakeLocation}
	if ctl.TimeOffsetSec == 0 {
		ctl.TimeOffsetSec = 600
	}
	if ctl.FakeLocation == "" {
		ctl.FakeLocation = "forged-location"
	}
	if err := c.sendControl(req.ShardID, req.NodeID, ctl); err != nil {
		c.writeJSON(w, http.StatusBadGateway, httpResult{OK: false, Error: err.Error()})
		return
	}
	c.writeJSON(w, http.StatusOK, httpResult{OK: true})
}

func (c *STCClientPlugin) handleByzantine(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		c.writeJSON(w, http.StatusMethodNotAllowed, httpResult{OK: false, Error: "method not allowed"})
		return
	}
	var req struct {
		ShardID           int  `json:"shardId"`
		NodeID            int  `json:"nodeId"`
		MaliciousDrop     bool `json:"maliciousDrop"`
		BroadcastMalBlock bool `json:"broadcastMaliciousBlock"`
	}
	_ = decodeBody(r, &req)
	if err := c.sendControl(req.ShardID, req.NodeID, STCControlPayload{Action: "set_malicious_drop", Enabled: req.MaliciousDrop}); err != nil {
		c.writeJSON(w, http.StatusBadGateway, httpResult{OK: false, Error: err.Error()})
		return
	}
	if err := c.sendControl(req.ShardID, req.NodeID, STCControlPayload{Action: "broadcast_malicious_block", Enabled: req.BroadcastMalBlock}); err != nil {
		c.writeJSON(w, http.StatusBadGateway, httpResult{OK: false, Error: err.Error()})
		return
	}
	c.writeJSON(w, http.StatusOK, httpResult{OK: true})
}

func (c *STCClientPlugin) sendControl(sid, nid int, ctl STCControlPayload) error {
	ip, ok := config.IPMap[sid][nid]
	if !ok {
		return fmt.Errorf("target node not found")
	}
	m := message.Message{Sender: c.nodeAttr.Ipaddr, MsgType: MsgSTCControl, Content: utils.Encode(ctl)}
	return c.p2pMod.ConnManager.Send(ip, m.JsonEncode())
}

func (c *STCClientPlugin) handleBlockHashes(w http.ResponseWriter, r *http.Request) {
	sid, _ := strconv.Atoi(r.URL.Query().Get("shardId"))
	nid, _ := strconv.Atoi(r.URL.Query().Get("nodeId"))
	start, _ := strconv.ParseInt(r.URL.Query().Get("start"), 10, 64)
	end, _ := strconv.ParseInt(r.URL.Query().Get("end"), 10, 64)

	reply, err := c.queryNode(sid, nid, STCQuery{Action: "block_hashes", StartHeight: start, EndHeight: end}, 3*time.Second)
	if err != nil {
		c.writeJSON(w, http.StatusBadGateway, httpResult{OK: false, Error: err.Error()})
		return
	}
	c.writeJSON(w, http.StatusOK, httpResult{OK: true, Data: reply.BlockHashes})
}

func (c *STCClientPlugin) queryNode(sid, nid int, query STCQuery, timeout time.Duration) (STCQueryReply, error) {
	ip, ok := config.IPMap[sid][nid]
	if !ok {
		return STCQueryReply{}, fmt.Errorf("target node not found")
	}
	requestID := fmt.Sprintf("stc-%d-%d-%d", sid, nid, atomic.AddUint64(&reqCounter, 1))
	query.RequestID = requestID

	ch := make(chan STCQueryReply, 1)
	c.pendingMu.Lock()
	c.pending[requestID] = ch
	c.pendingMu.Unlock()

	defer func() {
		c.pendingMu.Lock()
		delete(c.pending, requestID)
		c.pendingMu.Unlock()
	}()

	msgOut := message.Message{Sender: c.nodeAttr.Ipaddr, MsgType: message.MsgQuery, Content: utils.Encode(query)}
	if err := c.p2pMod.ConnManager.Send(ip, msgOut.JsonEncode()); err != nil {
		return STCQueryReply{}, err
	}

	select {
	case rep := <-ch:
		return rep, nil
	case <-time.After(timeout):
		return STCQueryReply{}, fmt.Errorf("query timeout")
	}
}

func (c *STCClientPlugin) handleReplyQuery(msg *message.Message) {
	var rep STCQueryReply
	if err := utils.Decode(msg.Content, &rep); err != nil {
		return
	}
	key := fmt.Sprintf("%d-%d", rep.ShardID, rep.NodeID)
	c.cacheMu.Lock()
	c.latestStats[key] = rep
	c.cacheMu.Unlock()

	c.pendingMu.Lock()
	ch, ok := c.pending[rep.RequestID]
	c.pendingMu.Unlock()
	if ok {
		select {
		case ch <- rep:
		default:
		}
	}
}

func (c *STCClientPlugin) pollMetricsAllNodes() {
	for sid := 0; sid < config.ShardNum; sid++ {
		for nid := 0; nid < config.NodeNum; nid++ {
			go func(s, n int) {
				rep, err := c.queryNode(s, n, STCQuery{Action: "metrics"}, 1200*time.Millisecond)
				if err == nil {
					key := fmt.Sprintf("%d-%d", rep.ShardID, rep.NodeID)
					c.cacheMu.Lock()
					c.latestStats[key] = rep
					c.cacheMu.Unlock()
				}
			}(sid, nid)
		}
	}
}

func decodeBody(r *http.Request, v interface{}) error {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return err
	}
	if len(body) == 0 {
		return nil
	}
	return json.Unmarshal(body, v)
}
