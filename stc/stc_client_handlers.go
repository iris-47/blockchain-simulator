package main

import (
	"BlockChainSimulator/config"
	"BlockChainSimulator/message"
	"BlockChainSimulator/utils"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

func (c *STCClientPlugin) registerHTTPRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/", c.serveFrontend)
	mux.HandleFunc("/api/status", c.apiStatus)
	mux.HandleFunc("/api/metrics", c.apiMetrics)
	mux.HandleFunc("/api/network/start", c.apiStartNetwork)
	mux.HandleFunc("/api/network/stop", c.apiStopNetwork)
	mux.HandleFunc("/api/node/start", c.apiStartNode)
	mux.HandleFunc("/api/node/stop", c.apiStopNode)
	mux.HandleFunc("/api/tx/send", c.apiSendTx)
	mux.HandleFunc("/api/tx/throughput", c.apiThroughputTest)
	mux.HandleFunc("/api/node/forge-spacetime", c.apiForgeSpaceTime)
	mux.HandleFunc("/api/node/byzantine", c.apiSetByzantine)
	mux.HandleFunc("/api/node/blocks", c.apiQueryBlocks)
	mux.HandleFunc("/api/node/info", c.apiNodeInfo)
	mux.HandleFunc("/static/", c.serveFrontend)
}

func (c *STCClientPlugin) serveJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func (c *STCClientPlugin) apiStatus(w http.ResponseWriter, r *http.Request) {
	c.mu.RLock()
	list := make([]*STCNodeStatus, 0, len(c.nodes))
	for _, s := range c.nodes {
		list = append(list, s)
	}
	c.mu.RUnlock()
	c.serveJSON(w, http.StatusOK, map[string]interface{}{"nodes": list})
}

func (c *STCClientPlugin) apiMetrics(w http.ResponseWriter, r *http.Request) {
	c.mu.RLock()
	res := make(map[int]STCMetricsSnapshot, len(c.metrics))
	for sid, m := range c.metrics {
		res[sid] = m
	}
	c.mu.RUnlock()
	c.serveJSON(w, http.StatusOK, map[string]interface{}{"shards": res})
}

func (c *STCClientPlugin) apiStartNetwork(w http.ResponseWriter, r *http.Request) {
	for sid := 0; sid < config.ShardNum; sid++ {
		for nid := 0; nid < config.NodeNum; nid++ {
			c.startNodeBySSH(sid, nid)
		}
	}
	c.serveJSON(w, http.StatusOK, map[string]interface{}{"ok": true})
}

func (c *STCClientPlugin) apiStopNetwork(w http.ResponseWriter, r *http.Request) {
	for sid := 0; sid < config.ShardNum; sid++ {
		for nid := 0; nid < config.NodeNum; nid++ {
			msg := message.Message{MsgType: message.MsgStop}
			_ = c.p2pMod.ConnManager.Send(config.IPMap[sid][nid], msg.JsonEncode())
		}
	}
	c.serveJSON(w, http.StatusOK, map[string]interface{}{"ok": true})
}

func (c *STCClientPlugin) apiStartNode(w http.ResponseWriter, r *http.Request) {
	targetShard, targetNode, ok := c.parseShardNode(r)
	if !ok {
		c.serveJSON(w, http.StatusBadRequest, map[string]interface{}{"error": "invalid shard/node"})
		return
	}
	c.startNodeBySSH(targetShard, targetNode)
	c.serveJSON(w, http.StatusOK, map[string]interface{}{"ok": true})
}

func (c *STCClientPlugin) apiStopNode(w http.ResponseWriter, r *http.Request) {
	targetShard, targetNode, ok := c.parseShardNode(r)
	if !ok {
		c.serveJSON(w, http.StatusBadRequest, map[string]interface{}{"error": "invalid shard/node"})
		return
	}
	msg := message.Message{MsgType: message.MsgStop}
	_ = c.p2pMod.ConnManager.Send(config.IPMap[targetShard][targetNode], msg.JsonEncode())
	c.serveJSON(w, http.StatusOK, map[string]interface{}{"ok": true})
}

func (c *STCClientPlugin) apiSendTx(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Shard   int    `json:"shard"`
		Payload string `json:"payload"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		body.Shard = 0
		body.Payload = "manual"
	}
	if body.Shard < 0 || body.Shard >= config.ShardNum {
		body.Shard = 0
	}
	tx := STCTransaction{ID: fmt.Sprintf("manual-%d", time.Now().UnixNano()), Payload: body.Payload, SubmitUnixNano: time.Now().UnixNano(), FromShard: body.Shard}
	m := message.Message{MsgType: MsgSTCForwardTx, Content: utils.Encode(STCInjectRequest{Txs: []STCTransaction{tx}})}
	_ = c.p2pMod.ConnManager.Send(config.IPMap[body.Shard][config.ViewNodeId], m.JsonEncode())
	c.serveJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "txId": tx.ID})
}

func (c *STCClientPlugin) apiThroughputTest(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Rate            int    `json:"rate"`
		DurationSeconds int    `json:"durationSeconds"`
		CSVPath         string `json:"csvPath"`
		Shard           int    `json:"shard"`
	}
	_ = json.NewDecoder(r.Body).Decode(&body)
	if body.Rate <= 0 {
		body.Rate = 100000
	}
	if body.DurationSeconds <= 0 {
		body.DurationSeconds = 120
	}
	if body.Shard < 0 || body.Shard >= config.ShardNum {
		body.Shard = 0
	}
	go c.runThroughputTest(body.Rate, body.DurationSeconds, body.CSVPath, body.Shard)
	c.serveJSON(w, http.StatusOK, map[string]interface{}{"ok": true})
}

func (c *STCClientPlugin) runThroughputTest(rate, durationSeconds int, csvPath string, shard int) {
	seedPayloads := c.loadPayloads(csvPath)
	if len(seedPayloads) == 0 {
		seedPayloads = []string{"simulated-tx"}
	}
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()
	deadline := time.Now().Add(time.Duration(durationSeconds) * time.Second)
	perTick := rate / 10
	if perTick <= 0 {
		perTick = 1
	}
	index := 0
	for time.Now().Before(deadline) {
		<-ticker.C
		batch := make([]STCTransaction, 0, perTick)
		now := time.Now().UnixNano()
		for i := 0; i < perTick; i++ {
			payload := seedPayloads[index%len(seedPayloads)]
			index++
			batch = append(batch, STCTransaction{ID: fmt.Sprintf("stress-%d-%d", now, i), Payload: payload, SubmitUnixNano: now, FromShard: shard})
		}
		msg := message.Message{MsgType: MsgSTCForwardTx, Content: utils.Encode(STCInjectRequest{Txs: batch})}
		_ = c.p2pMod.ConnManager.Send(config.IPMap[shard][config.ViewNodeId], msg.JsonEncode())
	}
}

func (c *STCClientPlugin) loadPayloads(path string) []string {
	if strings.TrimSpace(path) == "" {
		return []string{"generated"}
	}
	f, err := os.Open(path)
	if err != nil {
		utils.LoggerInstance.Warn("load csv failed, fallback to generated txs: %v", err)
		return []string{"generated"}
	}
	defer f.Close()
	rd := csv.NewReader(f)
	records, err := rd.ReadAll()
	if err != nil {
		utils.LoggerInstance.Warn("read csv failed, fallback to generated txs: %v", err)
		return []string{"generated"}
	}
	res := make([]string, 0, len(records))
	for _, row := range records {
		if len(row) == 0 {
			continue
		}
		res = append(res, row[0])
	}
	if len(res) == 0 {
		res = append(res, "generated")
	}
	return res
}

func (c *STCClientPlugin) apiForgeSpaceTime(w http.ResponseWriter, r *http.Request) {
	targetShard, targetNode, ok := c.parseShardNode(r)
	if !ok {
		c.serveJSON(w, http.StatusBadRequest, map[string]interface{}{"error": "invalid shard/node"})
		return
	}
	fakeLoc := r.URL.Query().Get("location")
	req := STCControlRequest{Action: "forge_spacetime_once", TargetShard: targetShard, TargetNode: targetNode, FakeLocation: fakeLoc}
	c.sendControl(targetShard, targetNode, req)
	c.serveJSON(w, http.StatusOK, map[string]interface{}{"ok": true})
}

func (c *STCClientPlugin) apiSetByzantine(w http.ResponseWriter, r *http.Request) {
	targetShard, targetNode, ok := c.parseShardNode(r)
	if !ok {
		c.serveJSON(w, http.StatusBadRequest, map[string]interface{}{"error": "invalid shard/node"})
		return
	}
	enabled := r.URL.Query().Get("enabled") == "true"
	behavior := r.URL.Query().Get("behavior")
	if behavior == "" {
		behavior = "broadcast_bad_block"
	}
	req := STCControlRequest{Action: "set_byzantine", TargetShard: targetShard, TargetNode: targetNode, Enabled: enabled, Behavior: behavior}
	c.sendControl(targetShard, targetNode, req)
	c.serveJSON(w, http.StatusOK, map[string]interface{}{"ok": true})
}

func (c *STCClientPlugin) apiQueryBlocks(w http.ResponseWriter, r *http.Request) {
	targetShard, targetNode, ok := c.parseShardNode(r)
	if !ok {
		c.serveJSON(w, http.StatusBadRequest, map[string]interface{}{"error": "invalid shard/node"})
		return
	}
	start, _ := strconv.ParseInt(r.URL.Query().Get("start"), 10, 64)
	end, _ := strconv.ParseInt(r.URL.Query().Get("end"), 10, 64)
	reply, err := c.queryNode(targetShard, targetNode, STCQueryRequest{Action: "blocks", StartHeight: start, EndHeight: end, ReplyTo: c.nodeAttr.Ipaddr})
	if err != nil {
		c.serveJSON(w, http.StatusGatewayTimeout, map[string]interface{}{"error": err.Error()})
		return
	}
	c.serveJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "blocks": reply.Blocks})
}

func (c *STCClientPlugin) apiNodeInfo(w http.ResponseWriter, r *http.Request) {
	targetShard, targetNode, ok := c.parseShardNode(r)
	if !ok {
		c.serveJSON(w, http.StatusBadRequest, map[string]interface{}{"error": "invalid shard/node"})
		return
	}
	reply, err := c.queryNode(targetShard, targetNode, STCQueryRequest{Action: "node_info", ReplyTo: c.nodeAttr.Ipaddr})
	if err != nil {
		c.serveJSON(w, http.StatusGatewayTimeout, map[string]interface{}{"error": err.Error()})
		return
	}
	c.serveJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "status": reply.Status})
}

func (c *STCClientPlugin) parseShardNode(r *http.Request) (int, int, bool) {
	sid, err1 := strconv.Atoi(r.URL.Query().Get("shard"))
	nid, err2 := strconv.Atoi(r.URL.Query().Get("node"))
	if err1 != nil || err2 != nil || sid < 0 || sid >= config.ShardNum || nid < 0 || nid >= config.NodeNum {
		return 0, 0, false
	}
	return sid, nid, true
}

func (c *STCClientPlugin) sendControl(shardID, nodeID int, req STCControlRequest) {
	msg := message.Message{MsgType: MsgSTCControl, Content: utils.Encode(req)}
	_ = c.p2pMod.ConnManager.Send(config.IPMap[shardID][nodeID], msg.JsonEncode())
}

func (c *STCClientPlugin) queryNode(shardID, nodeID int, req STCQueryRequest) (STCQueryReply, error) {
	req.RequestID = fmt.Sprintf("req-%d", time.Now().UnixNano())
	ch := make(chan STCQueryReply, 1)
	c.mu.Lock()
	c.pendingResp[req.RequestID] = ch
	c.mu.Unlock()
	defer func() {
		c.mu.Lock()
		delete(c.pendingResp, req.RequestID)
		c.mu.Unlock()
	}()
	msg := message.Message{MsgType: MsgSTCQuery, Content: utils.Encode(req)}
	if err := c.p2pMod.ConnManager.Send(config.IPMap[shardID][nodeID], msg.JsonEncode()); err != nil {
		return STCQueryReply{}, err
	}
	select {
	case resp := <-ch:
		if !resp.OK {
			return resp, fmt.Errorf(resp.Error)
		}
		return resp, nil
	case <-time.After(2 * time.Second):
		return STCQueryReply{}, fmt.Errorf("query timeout")
	}
}

func (c *STCClientPlugin) startNodeBySSH(shardID int, nodeID int) {
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
