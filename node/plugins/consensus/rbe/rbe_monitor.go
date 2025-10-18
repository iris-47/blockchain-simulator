// rbe_monitor.go
package main

import (
	"BlockChainSimulator/config"
	"BlockChainSimulator/message"
	"BlockChainSimulator/node/nodeattr"
	"BlockChainSimulator/node/p2p"
	"BlockChainSimulator/node/plugins/plugininterface"
	"BlockChainSimulator/utils"
	"context"
	"fmt"
	"sync"
	"time"
)

var _ plugininterface.Plugin = &RBEMonitorPlugin{}

// RBEMonitorPlugin 监控节点模块
type RBEMonitorPlugin struct {
	nodeAttr *nodeattr.NodeAttr
	p2pMod   *p2p.P2PMod

	// 统计数据收集
	reportsLock sync.Mutex
	reports     map[string]*MonitorReport // key: "shard_node"

	// 汇总统计
	totalVerifyCount int64
	avgTPS           float64
	peakTPS          float64
	nodeCount        int

	startTime time.Time
}

func NewRBEMonitorPlugin(attr *nodeattr.NodeAttr, p2p *p2p.P2PMod) plugininterface.Plugin {
	return &RBEMonitorPlugin{
		nodeAttr:  attr,
		p2pMod:    p2p,
		reports:   make(map[string]*MonitorReport),
		startTime: time.Now(),
	}
}

func (mon *RBEMonitorPlugin) Initialize() {
	mon.p2pMod.RegisterHandler(MsgRBEMonitorReport, mon.handleMonitorReportMsg)
}

func (mon *RBEMonitorPlugin) Run(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()

	utils.LoggerInstance.Info("=== RBE性能监控节点启动 ===")
	utils.LoggerInstance.Info("监控节点 [分片%d, 节点%d] 开始运行",
		mon.nodeAttr.Sid, mon.nodeAttr.Nid)

	// 等待系统稳定
	time.Sleep(2 * time.Second)

	// 定期查询所有节点的统计数据
	queryTicker := time.NewTicker(10 * time.Second)
	defer queryTicker.Stop()

	// 定期输出汇总报告
	reportTicker := time.NewTicker(30 * time.Second)
	defer reportTicker.Stop()

	for {
		select {
		case <-ctx.Done():
			mon.printFinalReport()
			utils.LoggerInstance.Info("监控节点 [分片%d, 节点%d] 停止",
				mon.nodeAttr.Sid, mon.nodeAttr.Nid)
			return

		case <-queryTicker.C:
			mon.queryAllNodes()

		case <-reportTicker.C:
			mon.printSummaryReport()
		}
	}
}

func (mon *RBEMonitorPlugin) Cleanup() {

}

// queryAllNodes 查询所有节点的统计数据
func (mon *RBEMonitorPlugin) queryAllNodes() {
	// 遍历所有分片的所有节点
	for sid, shardNodes := range config.IPMap {
		for nid, ip := range shardNodes {
			// 跳过KC节点和自己
			if nid == 0 || (sid == mon.nodeAttr.Sid && nid == mon.nodeAttr.Nid) {
				continue
			}

			// 发送查询消息
			queryMsg := message.Message{
				MsgType: MsgRBEMonitorQuery,
				Content: utils.Encode(mon.nodeAttr.Ipaddr),
			}

			mon.p2pMod.ConnMananger.Send(ip, queryMsg.JsonEncode())
		}
	}

	utils.LoggerInstance.Debug("监控节点发送查询请求到所有节点")
}

// handleMonitorReportMsg 处理监控报告消息
func (mon *RBEMonitorPlugin) handleMonitorReportMsg(msg *message.Message) {
	var report MonitorReport
	err := utils.Decode(msg.Content, &report)
	if err != nil {
		utils.LoggerInstance.Error("监控节点解码报告失败: %v", err)
		return
	}
	// 存储报告
	key := fmt.Sprintf("%d_%d", report.Sid, report.Nid)
	mon.reportsLock.Lock()
	mon.reports[key] = &report
	mon.reportsLock.Unlock()
	utils.LoggerInstance.Debug("监控节点收到报告 [分片%d, 节点%d] - 累计验证: %d, 平均TPS: %.2f, 瞬时TPS: %.2f",
		report.Sid, report.Nid, report.TotalVerifyCount, report.AvgTPS, report.InstantTPS)
}

// printSummaryReport 打印汇总报告
func (mon *RBEMonitorPlugin) printSummaryReport() {
	mon.reportsLock.Lock()
	defer mon.reportsLock.Unlock()
	if len(mon.reports) == 0 {
		utils.LoggerInstance.Warn("监控节点：暂无统计数据")
		return
	}
	var totalVerify int64
	var systemAvgTPS float64             // 系统总平均TPS（所有节点累加）
	var systemInstantTPS float64         // 系统总瞬时TPS（所有节点累加）
	var maxNodeAvgTPS float64            // 单节点最高平均TPS
	var minNodeAvgTPS float64 = 999999.0 // 单节点最低平均TPS
	nodeCount := len(mon.reports)
	// 按分片汇总
	shardStats := make(map[int]struct {
		totalVerifyCount int64
		avgTPS           float64 // 分片总平均TPS（累加）
		instantTPS       float64 // 分片总瞬时TPS（累加）
		nodeCount        int
	})
	utils.LoggerInstance.Info("========== RBE性能监控汇总报告 ==========")
	utils.LoggerInstance.Info("报告时间: %s", time.Now().Format("2006-01-02 15:04:05"))
	utils.LoggerInstance.Info("统计节点数: %d", nodeCount)
	utils.LoggerInstance.Info("")
	for _, report := range mon.reports {
		totalVerify += report.TotalVerifyCount

		// TPS使用累加而非平均
		systemAvgTPS += report.AvgTPS
		systemInstantTPS += report.InstantTPS

		// 记录单节点的最大最小TPS
		if report.AvgTPS > maxNodeAvgTPS {
			maxNodeAvgTPS = report.AvgTPS
		}
		if report.AvgTPS < minNodeAvgTPS && report.AvgTPS > 0 {
			minNodeAvgTPS = report.AvgTPS
		}
		// 分片统计（也使用累加）
		stat := shardStats[report.Sid]
		stat.totalVerifyCount += report.TotalVerifyCount
		stat.avgTPS += report.AvgTPS
		stat.instantTPS += report.InstantTPS
		stat.nodeCount++
		shardStats[report.Sid] = stat
		utils.LoggerInstance.Info("  [分片%d-节点%d] 累计验证: %d次, 平均TPS: %.2f, 瞬时TPS: %.2f, 时长: %.2fs",
			report.Sid, report.Nid, report.TotalVerifyCount, report.AvgTPS, report.InstantTPS, report.Duration)
	}
	// 更新历史峰值TPS
	if systemInstantTPS > mon.peakTPS {
		mon.peakTPS = systemInstantTPS
	}
	utils.LoggerInstance.Info("")
	utils.LoggerInstance.Info("---------- 分片统计 ----------")
	for sid, stat := range shardStats {
		utils.LoggerInstance.Info("  分片%d: 节点数=%d, 累计验证=%d, 分片总平均TPS=%.2f, 分片总瞬时TPS=%.2f",
			sid, stat.nodeCount, stat.totalVerifyCount, stat.avgTPS, stat.instantTPS)
	}
	systemDuration := time.Since(mon.startTime).Seconds()
	utils.LoggerInstance.Info("")
	utils.LoggerInstance.Info("---------- 系统总体性能 ----------")
	utils.LoggerInstance.Info("  累计验证总次数: %d", totalVerify)
	utils.LoggerInstance.Info("  系统平均TPS: %.2f ", systemAvgTPS)
	utils.LoggerInstance.Info("  系统当前TPS: %.2f ", systemInstantTPS)
	utils.LoggerInstance.Info("  系统历史峰值TPS: %.2f", mon.peakTPS)
	utils.LoggerInstance.Info("  单节点最高平均TPS: %.2f", maxNodeAvgTPS)
	utils.LoggerInstance.Info("  单节点最低平均TPS: %.2f", minNodeAvgTPS)
	// utils.LoggerInstance.Info("  系统总吞吐量(基于累计): %.2f 次/秒", float64(totalVerify)/systemDuration)
	utils.LoggerInstance.Info("  系统运行时长: %.2f秒", systemDuration)
	utils.LoggerInstance.Info("==========================================")
}

// printFinalReport 打印最终报告
func (mon *RBEMonitorPlugin) printFinalReport() {
	utils.LoggerInstance.Info("")
	utils.LoggerInstance.Info("========== RBE最终性能报告 ==========")
	mon.printSummaryReport()
	utils.LoggerInstance.Info("监控结束时间: %s", time.Now().Format("2006-01-02 15:04:05"))
	utils.LoggerInstance.Info("=====================================")
}
