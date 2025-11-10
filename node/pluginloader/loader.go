package pluginloader

import (
	"BlockChainSimulator/node/plugins/plugininterface"
	"BlockChainSimulator/utils"
	"fmt"
	"path/filepath"
	"plugin"
	"sync"
)

type PluginLoader struct {
	loadedPackages map[string]*plugin.Plugin                    // 已加载的 .so 文件
	registry       map[string]plugininterface.PluginConstructor // 插件名 -> 构造函数
	metadata       map[string]*plugininterface.PluginMetadata   // 插件名 -> 元数据
	packagePlugins map[string][]string                          // 包名 -> 插件列表
	mu             sync.RWMutex
}

var globalLoader *PluginLoader
var loaderOnce sync.Once

// GetLoader 获取全局插件加载器单例
func GetLoader() *PluginLoader {
	loaderOnce.Do(func() {
		globalLoader = &PluginLoader{
			loadedPackages: make(map[string]*plugin.Plugin),
			registry:       make(map[string]plugininterface.PluginConstructor),
			metadata:       make(map[string]*plugininterface.PluginMetadata),
			packagePlugins: make(map[string][]string),
		}
	})
	return globalLoader
}

// LoadPackage 加载一个插件包（.so 文件）
func (pl *PluginLoader) LoadPackage(packageName string, pluginDir string) error {
	pl.mu.Lock()
	defer pl.mu.Unlock()

	// 检查是否已加载
	if _, exists := pl.loadedPackages[packageName]; exists {
		utils.LoggerInstance.Debug("Package %s already loaded", packageName)
		return nil
	}

	// 构造 .so 文件路径
	soPath := filepath.Join(pluginDir, packageName+".so")

	// 加载插件
	p, err := plugin.Open(soPath)
	if err != nil {
		return fmt.Errorf("failed to open plugin %s: %v", soPath, err)
	}

	// 查找导出的 PluginPackage
	symPackage, err := p.Lookup("PluginPackage")
	if err != nil {
		return fmt.Errorf("plugin %s does not export PluginPackage: %v", packageName, err)
	}

	// 类型断言
	pluginPkg, ok := symPackage.(*plugininterface.PluginPackage)
	if !ok {
		return fmt.Errorf("plugin %s: PluginPackage has wrong type", packageName)
	}

	// 注册插件
	var pluginNames []string
	for name, constructor := range pluginPkg.Plugins {
		pl.registry[name] = constructor
		if meta, exists := pluginPkg.Metadata[name]; exists {
			pl.metadata[name] = meta
		}
		pluginNames = append(pluginNames, name)
		utils.LoggerInstance.Info("Registered plugin: %s from package %s", name, packageName)
	}

	pl.loadedPackages[packageName] = p
	pl.packagePlugins[packageName] = pluginNames

	utils.LoggerInstance.Info("Successfully loaded plugin package: %s (contains %d plugins)",
		packageName, len(pluginNames))

	return nil
}

// LoadPlugins 批量加载插件包
func (pl *PluginLoader) LoadPackages(packageNames []string, pluginDir string) error {
	for _, pkgName := range packageNames {
		if err := pl.LoadPackage(pkgName, pluginDir); err != nil {
			return err
		}
	}
	return nil
}

// GetPlugin 获取插件构造函数
func (pl *PluginLoader) GetPlugin(pluginName string) (plugininterface.PluginConstructor, error) {
	pl.mu.RLock()
	defer pl.mu.RUnlock()

	constructor, exists := pl.registry[pluginName]
	if !exists {
		return nil, fmt.Errorf("plugin %s not found", pluginName)
	}
	return constructor, nil
}

// GetMetadata 获取插件元数据
func (pl *PluginLoader) GetMetadata(pluginName string) (*plugininterface.PluginMetadata, error) {
	pl.mu.RLock()
	defer pl.mu.RUnlock()

	meta, exists := pl.metadata[pluginName]
	if !exists {
		return nil, fmt.Errorf("metadata for plugin %s not found", pluginName)
	}
	return meta, nil
}

// GetAllMetadata 获取所有插件元数据
func (pl *PluginLoader) GetAllMetadata() map[string]*plugininterface.PluginMetadata {
	pl.mu.RLock()
	defer pl.mu.RUnlock()

	result := make(map[string]*plugininterface.PluginMetadata)
	for k, v := range pl.metadata {
		result[k] = v
	}
	return result
}

// IsLoaded 检查插件是否已加载
func (pl *PluginLoader) IsLoaded(pluginName string) bool {
	pl.mu.RLock()
	defer pl.mu.RUnlock()

	_, exists := pl.registry[pluginName]
	return exists
}
