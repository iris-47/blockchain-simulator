// External configuration initialized from static settings.
// Contains deployment-specific parameters (e.g., network addresses, storage paths) that vary across environments (local, distributed, demo).
// Separated to decouple environment setup from runtime logic, enabling easy adjustments for deployments without code changes.
package config

import (
	"encoding/json"
	"fmt"
	"os"
)

var (
	ViewNodeId  = 0         // the nodeID of the initial view nodes
	ClientShard = 0xfffffff // the shardID of the client

	StoragePath = "./blockchain_data/" // the path to store the blockchain data
	ResultPath  = "./result/"          // measurement data result output path
	LogPath     = "./log/"             // log output path

	StartPort  = 28800                                                           // the start port of the IPnodeTable, in local environment
	ClientAddr = "127.0.0.1:23333"                                               // client ip address
	FileInput  = `/home/pjj/Desktop/BlockChain/dataset/0to99999_Transaction.csv` // the BlockTransaction data path

	DemoServerURL = "192.168.80.1:23333" // to send the log to the demo server, empty means not to send
)

// config of the distributed environment
var (
	NodePerServer = 40        // the number of nodes in a server
	ServerAddrs   = []string{ // for distribute experiment
		"192.168.0.1", "192.168.0.245", "192.168.0.251", "192.168.0.243",
		"192.168.0.246", "192.168.0.8", "192.168.0.250", "192.168.0.252",
		"192.168.0.249", "192.168.0.5", "192.168.0.244", "192.168.0.11",
		"192.168.0.247", "192.168.0.4", "192.168.0.6", "192.168.0.14",
		"192.168.0.2", "192.168.0.10", "192.168.0.12", "192.168.0.248",
		"192.168.0.15", "192.168.0.7", "192.168.0.3", "192.168.0.9",
		"192.168.0.13",
		// "192.168.0.1", "192.168.0.2", "192.168.0.3", "192.168.0.4",
		// "192.168.0.5", "192.168.0.6", "192.168.0.7", "192.168.0.8",
		// "192.168.0.9", "192.168.0.10", "192.168.0.11", "192.168.0.12",
		// "192.168.0.13", "192.168.0.14", "192.168.0.15", "192.168.0.16",
		// "192.168.0.17", "192.168.0.18", "192.168.0.19", "192.168.0.20",
		// "192.168.0.21", "192.168.0.22", "192.168.0.23", "192.168.0.24",
		// "192.168.0.25", "192.168.0.26", "192.168.0.27", "192.168.0.28",
		// "192.168.0.29", "192.168.0.30", "192.168.0.31", "192.168.0.32",
		// "192.168.0.33", "192.168.0.34", "192.168.0.35", "192.168.0.36",
		// "192.168.0.37", "192.168.0.38", "192.168.0.39", "192.168.0.40",
		// "192.168.0.41", "192.168.0.42", "192.168.0.43", "192.168.0.44",
		// "192.168.0.45", "192.168.0.46", "192.168.0.47", "192.168.0.48",
		// "192.168.0.49", "192.168.0.50", "192.168.0.51", "192.168.0.52",
		// "192.168.0.53", "192.168.0.54", "192.168.0.55", "192.168.0.56",
		// "192.168.0.57", "192.168.0.58", "192.168.0.59", "192.168.0.60",
		// "192.168.0.61", "192.168.0.62", "192.168.0.63", "192.168.0.64",
		// "192.168.0.65", "192.168.0.66", "192.168.0.67", "192.168.0.68",
		// "192.168.0.69", "192.168.0.70", "192.168.0.71", "192.168.0.72",
		// "192.168.0.73", "192.168.0.74", "192.168.0.75", "192.168.0.76",
		// "192.168.0.77", "192.168.0.78", "192.168.0.79", "192.168.0.80",
		// "127.0.0.1",
	}
)

var (
	// used to synchronize the start time of the protocol in some protocol using synchronous network model like TBB and DS
	StartTimeWait int64 = 1000 // the start time of the protocol(ms)
	TickInterval  int64 = 1000 // the interval between each clock(ms)
)

// LoadConfig loads configuration from config file if it exists
func LoadConfig() {
	// Check for config files in different locations
	configPaths := []string{
		"./config.json",
		"../config.json",
	}

	var configFile string
	for _, path := range configPaths {
		if _, err := os.Stat(path); err == nil {
			configFile = path
			break
		}
	}

	// If no config file found, use defaults
	if configFile == "" {
		fmt.Println("No configuration file found, using default values")
		return
	}

	// Read the config file
	data, err := os.ReadFile(configFile)
	if err != nil {
		fmt.Printf("Error reading config file: %s, using default values\n", err)
		return
	}

	// Parse the config
	var config ExtConfig
	if err := json.Unmarshal(data, &config); err != nil {
		fmt.Printf("Error parsing config file: %s, using default values\n", err)
		return
	}

	fmt.Printf("Using configuration file: %s\n", configFile)

	// Apply the configuration, only for values that are set
	if config.ViewNodeId != nil {
		ViewNodeId = *config.ViewNodeId
	}
	if config.ClientShard != nil {
		ClientShard = *config.ClientShard
	}
	if config.StoragePath != nil {
		StoragePath = *config.StoragePath
	}
	if config.ResultPath != nil {
		ResultPath = *config.ResultPath
	}
	if config.LogPath != nil {
		LogPath = *config.LogPath
	}
	if config.StartPort != nil {
		StartPort = *config.StartPort
	}
	if config.ClientAddr != nil {
		ClientAddr = *config.ClientAddr
	}
	if config.FileInput != nil {
		FileInput = *config.FileInput
	}
	if config.DemoServerURL != nil {
		DemoServerURL = *config.DemoServerURL
	}
	if config.NodePerServer != nil {
		NodePerServer = *config.NodePerServer
	}
	if config.ServerAddrs != nil && len(*config.ServerAddrs) > 0 {
		ServerAddrs = *config.ServerAddrs
	}
	if config.StartTimeWait != nil {
		StartTimeWait = *config.StartTimeWait
	}
	if config.TickInterval != nil {
		TickInterval = *config.TickInterval
	}
}
