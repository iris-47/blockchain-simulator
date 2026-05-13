//go:build ignore
// +build ignore

package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type PluginConfig struct {
	Plugin  string `json:"plugin"`
	Package string `json:"package"`
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

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run tools/build_protocol.go <protocol_name>")
		os.Exit(1)
	}

	protocolName := os.Args[1]

	// 读取配置
	data, err := os.ReadFile("protocols.json")
	if err != nil {
		fmt.Printf("Error reading protocols.json: %v\n", err)
		os.Exit(1)
	}

	var config ProtocolsConfig
	if err := json.Unmarshal(data, &config); err != nil {
		fmt.Printf("Error parsing protocols.json: %v\n", err)
		os.Exit(1)
	}

	protocol, exists := config.Protocols[protocolName]
	if !exists {
		fmt.Printf("Protocol '%s' not found in protocols.json\n", protocolName)
		os.Exit(1)
	}

	// 收集所有需要的包
	packages := make(map[string]bool)
	for _, p := range protocol.Client {
		packages[p.Package] = true
	}
	for _, p := range protocol.Normal {
		packages[p.Package] = true
	}
	for _, p := range protocol.View {
		packages[p.Package] = true
	}

	fmt.Printf("Building plugins for protocol: %s\n", protocolName)
	fmt.Printf("Description: %s\n", protocol.Description)
	fmt.Printf("Required packages: %v\n", getKeys(packages))
	fmt.Println()

	// 编译每个包
	for pkg := range packages {
		if err := buildPackage(pkg); err != nil {
			fmt.Printf("Error building package %s: %v\n", pkg, err)
			os.Exit(1)
		}
	}

	fmt.Println("\n✓ All plugins for protocol", protocolName, "built successfully")
}

func buildPackage(pkgName string) error {
	// 查找包的路径
	searchPaths := []string{
		filepath.Join("node", "plugins", "consensus", pkgName),
		filepath.Join("node", "plugins", "client", pkgName),
		filepath.Join("node", "plugins", "auxiliary", pkgName),
		filepath.Join("node", "plugins", "global", pkgName),
	}

	var pkgPath string
	for _, path := range searchPaths {
		if _, err := os.Stat(path); err == nil {
			pkgPath = path
			break
		}
	}

	if pkgPath == "" {
		return fmt.Errorf("package %s not found", pkgName)
	}

	// 编译
	outputPath := filepath.Join("plugins_bin", pkgName+".so")
	fmt.Printf("Building %s from %s...\n", outputPath, pkgPath)

	// TODO: 解决项目路径带来的强制查找mod文件的问题，目前可以通过编译/*.go或者修改为相对路径两种方式来规避，但很不优雅。
	if !filepath.IsAbs(outputPath) && !strings.HasPrefix(outputPath, "./") {
		outputPath = "./" + outputPath
	}

	if !filepath.IsAbs(pkgPath) && !strings.HasPrefix(pkgPath, "./") {
		pkgPath = "./" + pkgPath
	}

	cmd := exec.Command("go", "build", "-buildmode=plugin", "-o", outputPath, pkgPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

func getKeys(m map[string]bool) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}
