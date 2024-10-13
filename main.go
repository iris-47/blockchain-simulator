package main

import (
	"BlockChainSimulator/config"
	"BlockChainSimulator/node"
	"BlockChainSimulator/node/msgHandler"
	"BlockChainSimulator/utils"
	"flag"
	"fmt"
)

func runNode1() {
	utils.LoggerInstance, _ = utils.NewLogger("", utils.DEBUG, "S0N0", true)
	node1, err := node.NewNode(0, 0, nil, msgHandler.PBFT, "simple", msgHandler.ProposeTxsAuxiliary)
	if err != nil {
		utils.LoggerInstance.Error("Error creating node: %v", err)
		return
	}
	node1.Run()
}

func runNode2() {
	utils.LoggerInstance, _ = utils.NewLogger("", utils.DEBUG, "S0N1", true)
	node2, err := node.NewNode(0, 1, nil, msgHandler.PBFT, "simple", msgHandler.TestAuxiliary)
	if err != nil {
		utils.LoggerInstance.Error("Error creating node: %v", err)
		return
	}
	node2.Run()
}

func main() {
	if _, ok := config.IPMap[0]; !ok {
		config.IPMap[0] = make(map[int]string)
		config.IPMap[0][0] = "127.0.0.1:10001"
		config.IPMap[0][1] = "127.0.0.1:10002"
		config.IPMap[0][2] = "127.0.0.1:10003"
	}

	// 通过参数控制启动Node1或Node2
	task := flag.String("task", "node1", "task to run: node1 or node2")
	flag.Parse()

	switch *task {
	case "node1":
		runNode1()
	case "node2":
		runNode2()
	default:
		fmt.Println("Unknown task")
	}
}
