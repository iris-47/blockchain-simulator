//go:build ignore
// +build ignore

package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"plugin"
	"sort"
	"strings"

	"BlockChainSimulator/node/plugins/plugininterface"
)

const (
	ColorReset  = "\033[0m"
	ColorRed    = "\033[31m"
	ColorGreen  = "\033[32m"
	ColorYellow = "\033[33m"
	ColorBlue   = "\033[34m"
	ColorPurple = "\033[35m"
	ColorCyan   = "\033[36m"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		return
	}

	command := os.Args[1]
	switch command {
	case "list":
		listPlugins()
	case "info":
		if len(os.Args) > 2 {
			showPluginInfo(os.Args[2])
		} else {
			showAllPluginsInfo()
		}
	case "json":
		exportJSON()
	default:
		printUsage()
	}
}

func printUsage() {
	fmt.Println("Plugin Information Tool")
	fmt.Println("\nUsage:")
	fmt.Println("  go run tools/plugin_info.go <command> [args]")
	fmt.Println("\nCommands:")
	fmt.Println("  list              - List all available plugins")
	fmt.Println("  info [plugin]     - Show detailed information for a plugin (or all if not specified)")
	fmt.Println("  json              - Export all plugin information as JSON")
}

func listPlugins() {
	plugins := loadAllPlugins()
	if len(plugins) == 0 {
		fmt.Println(ColorRed + "No plugins found in plugins_bin/" + ColorReset)
		return
	}

	// 按类别分组
	categories := make(map[string][]PluginInfo)
	for _, p := range plugins {
		categories[p.Category] = append(categories[p.Category], p)
	}

	fmt.Println(ColorBlue + "=== Available Plugins ===" + ColorReset)
	fmt.Println()

	categoryOrder := []string{"consensus", "client", "auxiliary", "global"}
	categoryNames := map[string]string{
		"consensus": "Consensus Plugins",
		"client":    "Client Plugins",
		"auxiliary": "Auxiliary Plugins",
		"global":    "Global Plugins",
	}

	for _, cat := range categoryOrder {
		if pluginList, exists := categories[cat]; exists {
			fmt.Printf("%s[%s]%s\n", ColorGreen, categoryNames[cat], ColorReset)
			sort.Slice(pluginList, func(i, j int) bool {
				return pluginList[i].Name < pluginList[j].Name
			})
			for _, p := range pluginList {
				fmt.Printf("  %s%-20s%s - %s (from %s.so)\n",
					ColorCyan, p.Name, ColorReset, p.Description, p.Package)
			}
			fmt.Println()
		}
	}

	fmt.Printf("Total: %s%d%s plugins from %s%d%s packages\n",
		ColorYellow, len(plugins), ColorReset,
		ColorYellow, countPackages(plugins), ColorReset)
}

func showPluginInfo(pluginName string) {
	plugins := loadAllPlugins()
	for _, p := range plugins {
		if strings.EqualFold(p.Name, pluginName) {
			printPluginDetail(p)
			return
		}
	}
	fmt.Printf("%sPlugin '%s' not found%s\n", ColorRed, pluginName, ColorReset)
}

func showAllPluginsInfo() {
	plugins := loadAllPlugins()
	if len(plugins) == 0 {
		fmt.Println(ColorRed + "No plugins found" + ColorReset)
		return
	}

	for i, p := range plugins {
		printPluginDetail(p)
		if i < len(plugins)-1 {
			fmt.Println(strings.Repeat("-", 80))
		}
	}
}

func printPluginDetail(p PluginInfo) {
	fmt.Printf("%s=== Plugin: %s ===%s\n", ColorBlue, p.Name, ColorReset)
	fmt.Printf("Description:  %s\n", p.Description)
	fmt.Printf("Version:      %s\n", p.Version)
	fmt.Printf("Category:     %s%s%s\n", ColorYellow, p.Category, ColorReset)
	fmt.Printf("Package:      %s.so\n", p.Package)
	if len(p.Dependencies) > 0 {
		fmt.Printf("Dependencies: %s\n", strings.Join(p.Dependencies, ", "))
	} else {
		fmt.Printf("Dependencies: %sNone%s\n", ColorGreen, ColorReset)
	}
	fmt.Println()
}

func exportJSON() {
	plugins := loadAllPlugins()
	data, err := json.MarshalIndent(plugins, "", "  ")
	if err != nil {
		fmt.Printf("%sError exporting JSON: %v%s\n", ColorRed, err, ColorReset)
		return
	}
	fmt.Println(string(data))
}

type PluginInfo struct {
	Name         string   `json:"name"`
	Description  string   `json:"description"`
	Version      string   `json:"version"`
	Category     string   `json:"category"`
	Package      string   `json:"package"`
	Dependencies []string `json:"dependencies"`
}

func loadAllPlugins() []PluginInfo {
	var allPlugins []PluginInfo

	pluginDir := "plugins_bin"
	entries, err := os.ReadDir(pluginDir)
	if err != nil {
		return allPlugins
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".so") {
			continue
		}

		soPath := filepath.Join(pluginDir, entry.Name())
		p, err := plugin.Open(soPath)
		if err != nil {
			continue
		}

		symPackage, err := p.Lookup("PluginPackage")
		if err != nil {
			continue
		}

		pluginPkg, ok := symPackage.(*plugininterface.PluginPackage)
		if !ok {
			continue
		}

		packageName := strings.TrimSuffix(entry.Name(), ".so")
		for name, meta := range pluginPkg.Metadata {
			allPlugins = append(allPlugins, PluginInfo{
				Name:         name,
				Description:  meta.Description,
				Version:      meta.Version,
				Category:     meta.Category,
				Package:      packageName,
				Dependencies: meta.Dependencies,
			})
		}
	}

	sort.Slice(allPlugins, func(i, j int) bool {
		if allPlugins[i].Category != allPlugins[j].Category {
			return allPlugins[i].Category < allPlugins[j].Category
		}
		return allPlugins[i].Name < allPlugins[j].Name
	})

	return allPlugins
}

func countPackages(plugins []PluginInfo) int {
	packages := make(map[string]bool)
	for _, p := range plugins {
		packages[p.Package] = true
	}
	return len(packages)
}
