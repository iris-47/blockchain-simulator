package config

import (
	"encoding/json"
	"fmt"
	"os"
)

type ProtocolConfig struct {
	Description string   `json:"description"`
	Client      []string `json:"client"`
	Normal      []string `json:"normal"`
	View        []string `json:"view"`
}

type ProtocolsConfig struct {
	Protocols map[string]ProtocolConfig `json:"protocols"`
}

var GlobalProtocolsConfig *ProtocolsConfig

// LoadProtocolsConfig 加载协议配置
func LoadProtocolsConfig(filePath string) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read protocols config: %w", err)
	}

	config := &ProtocolsConfig{}
	if err := json.Unmarshal(data, config); err != nil {
		return fmt.Errorf("failed to parse protocols config: %w", err)
	}

	GlobalProtocolsConfig = config
	return nil
}

// GetProtocolPlugins 获取指定协议和节点类型的插件列表
func GetProtocolPlugins(protocol string, nodeType string) ([]string, error) {
	if GlobalProtocolsConfig == nil {
		return nil, fmt.Errorf("protocols config not loaded")
	}

	protocolConfig, exists := GlobalProtocolsConfig.Protocols[protocol]
	if !exists {
		return nil, fmt.Errorf("protocol %s not found", protocol)
	}

	switch nodeType {
	case "client":
		return protocolConfig.Client, nil
	case "normal":
		return protocolConfig.Normal, nil
	case "view":
		return protocolConfig.View, nil
	default:
		return nil, fmt.Errorf("invalid node type: %s", nodeType)
	}
}
