package pluginloader

import (
	"encoding/json"
	"fmt"
	"os"
)

type PluginConfig struct {
	Plugin  string `json:"plugin"`  // 插件名
	Package string `json:"package"` // 所属包名
}

type ProtocolConfig struct {
	Description string         `json:"description"`
	Client      []PluginConfig `json:"client"`
	Normal      []PluginConfig `json:"normal"`
	View        []PluginConfig `json:"view"`
}

type ProtocolsConfig struct {
	Protocols map[string]ProtocolConfig `json:"protocols"`
}

// LoadProtocolsConfig 加载协议配置文件
func LoadProtocolsConfig(configPath string) (*ProtocolsConfig, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %v", err)
	}

	var config ProtocolsConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config: %v", err)
	}

	return &config, nil
}

// GetRequiredPackages 获取协议需要的所有包
func (pc *ProtocolsConfig) GetRequiredPackages(protocolName string) ([]string, error) {
	protocol, exists := pc.Protocols[protocolName]
	if !exists {
		return nil, fmt.Errorf("protocol %s not found", protocolName)
	}

	packageSet := make(map[string]bool)

	for _, plugin := range protocol.Client {
		packageSet[plugin.Package] = true
	}
	for _, plugin := range protocol.Normal {
		packageSet[plugin.Package] = true
	}
	for _, plugin := range protocol.View {
		packageSet[plugin.Package] = true
	}

	packages := make([]string, 0, len(packageSet))
	for pkg := range packageSet {
		packages = append(packages, pkg)
	}

	return packages, nil
}

// GetPluginsForNode 根据节点类型获取插件列表
func (pc *ProtocolsConfig) GetPluginsForNode(protocolName string, nodeType string) ([]PluginConfig, error) {
	protocol, exists := pc.Protocols[protocolName]
	if !exists {
		return nil, fmt.Errorf("protocol %s not found", protocolName)
	}

	switch nodeType {
	case "client":
		return protocol.Client, nil
	case "normal":
		return protocol.Normal, nil
	case "view":
		return protocol.View, nil
	default:
		return nil, fmt.Errorf("unknown node type: %s", nodeType)
	}
}
