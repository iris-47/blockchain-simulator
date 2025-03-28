// an interactive client that sends a string to the system, used by the TBB protocol and the DS protocol or for testing
package clientMod

import (
	"BlockChainSimulator/config"
	"BlockChainSimulator/message"
	"BlockChainSimulator/node/nodeattr"
	"BlockChainSimulator/node/p2p"
	"BlockChainSimulator/node/runningMod/runningModInterface"
	"BlockChainSimulator/utils"
	"bufio"
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"
)

var _ runningModInterface.RunningMod = &sendStringManualMod{}

type CommandHandler func(args []string) error

// just for test use, this mod sends Txs every 3 seconds
type sendStringManualMod struct {
	nodeAttr *nodeattr.NodeAttr
	p2pMod   *p2p.P2PMod

	commands     map[string]CommandHandler
	inputScanner *bufio.Scanner
}

// just for test use, this mod sends Txs every 3 seconds
func NewSendStringManualMod(attr *nodeattr.NodeAttr, p2p *p2p.P2PMod) runningModInterface.RunningMod {
	ssmm := new(sendStringManualMod)
	ssmm.nodeAttr = attr
	ssmm.p2pMod = p2p

	ssmm.commands = map[string]CommandHandler{
		"propose": ssmm.handleCmdPropose,
		"help":    ssmm.handleCmdHelp,
	}
	ssmm.inputScanner = bufio.NewScanner(os.Stdin)

	return ssmm
}

// ------------------------------- Interface Implementations -------------------------------

func (ssmm *sendStringManualMod) RegisterHandlers() {

}

func (ssmm *sendStringManualMod) Run(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()

	// wait for the system to start
	if !p2p.WaitForAllIPsReady(20 * time.Second) {
		utils.LoggerInstance.Error("Wait for all IPs ready timeout")
		return
	}
	utils.LoggerInstance.Info("All IPs are ready, start to send request")

	ssmm.printWelcome()

	inputChan := make(chan string)
	go ssmm.inputRoutine(inputChan)
	// read the input from stdin
	for {
		select {
		case <-ctx.Done():
			utils.LoggerInstance.Info("Stop the sendStringManualMod")
			return
		case input := <-inputChan:
			ssmm.processInput(input)
		}
	}
}

// ------------------------------- Module-specific Functions -------------------------------

func (ssmm *sendStringManualMod) printWelcome() {
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println("Interactive Blockchain Client")
	fmt.Printf("Connected to Node ID: %d | IP: %s\n", 0, config.IPMap[0][0])
	fmt.Println("Type 'help' for available commands")
	fmt.Println(strings.Repeat("=", 60))
}

func (ssmm *sendStringManualMod) printUsage() {
	fmt.Println(`
Available commands:
  propose <message>  - Submit a new string input to the blockchain
  help               - Show this help message

Examples:
  propose "Transfer 100 BTC to Alice"
  help

Press Ctrl+C to exit
  `)
}
func (ssmm *sendStringManualMod) inputRoutine(inputChan chan<- string) {
	fmt.Print("> ")
	for ssmm.inputScanner.Scan() {
		input := strings.TrimSpace(ssmm.inputScanner.Text())
		if input != "" {
			inputChan <- input
		}
		fmt.Print("> ")
	}
	close(inputChan)
}

func (ssmm *sendStringManualMod) processInput(input string) {
	parts := strings.SplitN(input, " ", 2)
	if len(parts) == 0 {
		fmt.Println("Please enter a command. Type 'help' for usage")
	}

	cmd := strings.ToLower(parts[0])
	handler, exists := ssmm.commands[cmd]
	if !exists {
		fmt.Printf("Unknown command: %v, type 'help' for usage\n", cmd)
		return
	}

	var args []string
	if len(parts) > 1 {
		args = []string{parts[1]}
	}

	handler(args)
}

func (ssmm *sendStringManualMod) handleCmdPropose(args []string) error {
	if len(args) == 0 || args[0] == "" {
		return errors.New("propose command requires a message")
	}

	proposeMsg := message.Message{
		MsgType: message.MsgInject,
		Content: utils.Encode(args[0]),
	}

	utils.LoggerInstance.Info("Inject value %v", args[0])
	ssmm.p2pMod.ConnMananger.Send(config.IPMap[0][0], proposeMsg.JsonEncode())
	return nil
}

func (ssmm *sendStringManualMod) handleCmdHelp([]string) error {
	ssmm.printUsage()
	return nil
}
