package config

import (
	"encoding/json"
	"os"
	"reflect"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	origViewNodeId := ViewNodeId
	origClientShard := ClientShard
	origStoragePath := StoragePath
	origResultPath := ResultPath
	origLogPath := LogPath
	origStartPort := StartPort
	origClientAddr := ClientAddr
	origFileInput := FileInput
	origDemoServerURL := DemoServerURL
	origNodePerServer := NodePerServer
	origServerAddrs := make([]string, len(ServerAddrs))
	copy(origServerAddrs, ServerAddrs)
	origStartTimeWait := StartTimeWait
	origTickInterval := TickInterval

	defer func() {
		ViewNodeId = origViewNodeId
		ClientShard = origClientShard
		StoragePath = origStoragePath
		ResultPath = origResultPath
		LogPath = origLogPath
		StartPort = origStartPort
		ClientAddr = origClientAddr
		FileInput = origFileInput
		DemoServerURL = origDemoServerURL
		NodePerServer = origNodePerServer
		ServerAddrs = origServerAddrs
		StartTimeWait = origStartTimeWait
		TickInterval = origTickInterval

		os.Remove("./config.json")
	}()

	// Test1: test when no config file is present
	t.Run("UseDefaultValuesWhenNoConfigFile", func(t *testing.T) {
		// make sure the config file does not exist
		os.Remove("./config.json")
		os.Remove("../config.json")

		ViewNodeId = origViewNodeId
		ClientShard = origClientShard

		LoadConfig()

		if ViewNodeId != origViewNodeId {
			t.Errorf("Default ViewNodeId changed: got %d, want %d", ViewNodeId, origViewNodeId)
		}
		if ClientShard != origClientShard {
			t.Errorf("Default ClientShard changed: got %d, want %d", ClientShard, origClientShard)
		}
	})

	// Test2: test when config file is present
	t.Run("LoadConfigFromFile", func(t *testing.T) {
		testConfig := ExtConfig{
			ViewNodeId:    intPtr(10),
			ClientShard:   intPtr(20),
			StoragePath:   strPtr("./test_storage/"),
			ResultPath:    strPtr("./test_result/"),
			LogPath:       strPtr("./test_log/"),
			StartPort:     intPtr(9000),
			ClientAddr:    strPtr("192.168.1.1:12345"),
			FileInput:     strPtr("/test/data.csv"),
			DemoServerURL: strPtr("192.168.1.2:12345"),
			NodePerServer: intPtr(5),
			ServerAddrs:   &[]string{"192.168.1.3", "192.168.1.4"},
			StartTimeWait: int64Ptr(500),
			TickInterval:  int64Ptr(200),
		}

		configBytes, err := json.Marshal(testConfig)
		if err != nil {
			t.Fatalf("Failed to marshal test config: %v", err)
		}
		err = os.WriteFile("./config.json", configBytes, 0644)
		if err != nil {
			t.Fatalf("Failed to write test config file: %v", err)
		}

		LoadConfig()

		if ViewNodeId != *testConfig.ViewNodeId {
			t.Errorf("ViewNodeId not loaded correctly: got %d, want %d", ViewNodeId, *testConfig.ViewNodeId)
		}
		if ClientShard != *testConfig.ClientShard {
			t.Errorf("ClientShard not loaded correctly: got %d, want %d", ClientShard, *testConfig.ClientShard)
		}
		if StoragePath != *testConfig.StoragePath {
			t.Errorf("StoragePath not loaded correctly: got %s, want %s", StoragePath, *testConfig.StoragePath)
		}
		if ResultPath != *testConfig.ResultPath {
			t.Errorf("ResultPath not loaded correctly: got %s, want %s", ResultPath, *testConfig.ResultPath)
		}
		if LogPath != *testConfig.LogPath {
			t.Errorf("LogPath not loaded correctly: got %s, want %s", LogPath, *testConfig.LogPath)
		}
		if StartPort != *testConfig.StartPort {
			t.Errorf("StartPort not loaded correctly: got %d, want %d", StartPort, *testConfig.StartPort)
		}
		if ClientAddr != *testConfig.ClientAddr {
			t.Errorf("ClientAddr not loaded correctly: got %s, want %s", ClientAddr, *testConfig.ClientAddr)
		}
		if FileInput != *testConfig.FileInput {
			t.Errorf("FileInput not loaded correctly: got %s, want %s", FileInput, *testConfig.FileInput)
		}
		if DemoServerURL != *testConfig.DemoServerURL {
			t.Errorf("DemoServerURL not loaded correctly: got %s, want %s", DemoServerURL, *testConfig.DemoServerURL)
		}
		if NodePerServer != *testConfig.NodePerServer {
			t.Errorf("NodePerServer not loaded correctly: got %d, want %d", NodePerServer, *testConfig.NodePerServer)
		}
		if !reflect.DeepEqual(ServerAddrs, *testConfig.ServerAddrs) {
			t.Errorf("ServerAddrs not loaded correctly: got %v, want %v", ServerAddrs, *testConfig.ServerAddrs)
		}
		if StartTimeWait != *testConfig.StartTimeWait {
			t.Errorf("StartTimeWait not loaded correctly: got %d, want %d", StartTimeWait, *testConfig.StartTimeWait)
		}
		if TickInterval != *testConfig.TickInterval {
			t.Errorf("TickInterval not loaded correctly: got %d, want %d", TickInterval, *testConfig.TickInterval)
		}
	})

	// Test3: Partial configuration
	t.Run("PartialConfiguration", func(t *testing.T) {
		ViewNodeId = origViewNodeId
		ClientShard = origClientShard
		StoragePath = origStoragePath

		partialConfig := ExtConfig{
			ViewNodeId: intPtr(99),
			// other fields omitted
		}

		configBytes, err := json.Marshal(partialConfig)
		if err != nil {
			t.Fatalf("Failed to marshal partial test config: %v", err)
		}
		err = os.WriteFile("./config.json", configBytes, 0644)
		if err != nil {
			t.Fatalf("Failed to write partial test config file: %v", err)
		}

		LoadConfig()

		if ViewNodeId != *partialConfig.ViewNodeId {
			t.Errorf("ViewNodeId not loaded correctly: got %d, want %d", ViewNodeId, *partialConfig.ViewNodeId)
		}

		if ClientShard != origClientShard {
			t.Errorf("Unconfigured ClientShard changed: got %d, want %d", ClientShard, origClientShard)
		}
		if StoragePath != origStoragePath {
			t.Errorf("Unconfigured StoragePath changed: got %s, want %s", StoragePath, origStoragePath)
		}
	})

	// Test4: Invalid configuration
	t.Run("InvalidJSONConfig", func(t *testing.T) {
		ViewNodeId = origViewNodeId
		ClientShard = origClientShard

		invalidJSON := []byte(`{"ViewNodeId": 5, "ClientShard": invalid}`)
		err := os.WriteFile("./config.json", invalidJSON, 0644)
		if err != nil {
			t.Fatalf("Failed to write invalid test config file: %v", err)
		}

		LoadConfig()

		if ViewNodeId != origViewNodeId {
			t.Errorf("ViewNodeId changed with invalid JSON: got %d, want %d", ViewNodeId, origViewNodeId)
		}
		if ClientShard != origClientShard {
			t.Errorf("ClientShard changed with invalid JSON: got %d, want %d", ClientShard, origClientShard)
		}
	})
}

func intPtr(i int) *int {
	return &i
}

func strPtr(s string) *string {
	return &s
}

func int64Ptr(i int64) *int64 {
	return &i
}
