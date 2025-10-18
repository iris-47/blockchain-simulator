# Makefile for BlockChainSimulator Plugin System

# 目录配置
PLUGIN_SRC_DIR := node/plugins
PLUGIN_BIN_DIR := plugins_bin
CONSENSUS_DIR := $(PLUGIN_SRC_DIR)/consensus
CLIENT_DIR := $(PLUGIN_SRC_DIR)/client
AUXILIARY_DIR := $(PLUGIN_SRC_DIR)/auxiliary
GLOBAL_DIR := $(PLUGIN_SRC_DIR)/global

# Go 编译器配置
GO := go
GOBUILD := $(GO) build
GOCLEAN := $(GO) clean
GOTEST := $(GO) test
PLUGIN_FLAGS := -buildmode=plugin

# 查找所有插件包
CONSENSUS_PACKAGES := $(wildcard $(CONSENSUS_DIR)/*)
CLIENT_PACKAGES := $(wildcard $(CLIENT_DIR)/*)
AUXILIARY_PACKAGES := $(wildcard $(AUXILIARY_DIR)/*)
GLOBAL_PACKAGES := $(wildcard $(GLOBAL_DIR)/*)

ALL_PACKAGES := $(CONSENSUS_PACKAGES) $(CLIENT_PACKAGES) $(AUXILIARY_PACKAGES) $(GLOBAL_PACKAGES)

# 生成 .so 文件路径
CONSENSUS_SOS := $(patsubst $(CONSENSUS_DIR)/%,$(PLUGIN_BIN_DIR)/%.so,$(CONSENSUS_PACKAGES))
CLIENT_SOS := $(patsubst $(CLIENT_DIR)/%,$(PLUGIN_BIN_DIR)/%.so,$(CLIENT_PACKAGES))
AUXILIARY_SOS := $(patsubst $(AUXILIARY_DIR)/%,$(PLUGIN_BIN_DIR)/%.so,$(AUXILIARY_PACKAGES))
GLOBAL_SOS := $(patsubst $(GLOBAL_DIR)/%,$(PLUGIN_BIN_DIR)/%.so,$(GLOBAL_PACKAGES))

ALL_SOS := $(CONSENSUS_SOS) $(CLIENT_SOS) $(AUXILIARY_SOS) $(GLOBAL_SOS)

# 颜色输出
COLOR_RESET := \033[0m
COLOR_GREEN := \033[32m
COLOR_YELLOW := \033[33m
COLOR_BLUE := \033[34m

.PHONY: all clean plugins main help list protocol test

# 默认目标：编译主程序和所有插件
all: main plugins
	@echo "$(COLOR_GREEN)✓ Build completed successfully$(COLOR_RESET)"

# 编译主程序
main:
	@echo "$(COLOR_BLUE)Building main program...$(COLOR_RESET)"
	@$(GOBUILD) -o BlockChainSimulator main.go
	@echo "$(COLOR_GREEN)✓ Main program built$(COLOR_RESET)"

# 编译所有插件
plugins: $(PLUGIN_BIN_DIR) $(ALL_SOS)
	@echo "$(COLOR_GREEN)✓ All plugins built$(COLOR_RESET)"

# 创建插件输出目录
$(PLUGIN_BIN_DIR):
	@mkdir -p $(PLUGIN_BIN_DIR)

# 编译单个插件的规则
$(PLUGIN_BIN_DIR)/%.so: $(PLUGIN_SRC_DIR)/*/%
	@echo "$(COLOR_YELLOW)Building plugin: $*...$(COLOR_RESET)"
	@$(GOBUILD) $(PLUGIN_FLAGS) -o $@ $</*.go
	@echo "$(COLOR_GREEN)✓ Built $*.so$(COLOR_RESET)"

# 按协议编译（从 protocols.json 读取）
protocol:
	@if [ -z "$(PROTOCOL)" ]; then \
		echo "$(COLOR_YELLOW)Usage: make protocol PROTOCOL=<protocol_name>$(COLOR_RESET)"; \
		echo "Example: make protocol PROTOCOL=RBE"; \
		exit 1; \
	fi
	@echo "$(COLOR_BLUE)Building plugins for protocol: $(PROTOCOL)...$(COLOR_RESET)"
	@$(GO) run tools/build_protocol.go $(PROTOCOL)

# 编译单个插件包
plugin:
	@if [ -z "$(PKG)" ]; then \
		echo "$(COLOR_YELLOW)Usage: make plugin PKG=<package_name>$(COLOR_RESET)"; \
		echo "Example: make plugin PKG=pbft"; \
		exit 1; \
	fi
	@echo "$(COLOR_BLUE)Building plugin package: $(PKG)...$(COLOR_RESET)"
	@mkdir -p $(PLUGIN_BIN_DIR)
	@if [ -d "$(CONSENSUS_DIR)/$(PKG)" ]; then \
		$(GOBUILD) $(PLUGIN_FLAGS) -o $(PLUGIN_BIN_DIR)/$(PKG).so $(CONSENSUS_DIR)/$(PKG)/*.go; \
	elif [ -d "$(CLIENT_DIR)/$(PKG)" ]; then \
		$(GOBUILD) $(PLUGIN_FLAGS) -o $(PLUGIN_BIN_DIR)/$(PKG).so $(CLIENT_DIR)/$(PKG)/*.go; \
	elif [ -d "$(AUXILIARY_DIR)/$(PKG)" ]; then \
		$(GOBUILD) $(PLUGIN_FLAGS) -o $(PLUGIN_BIN_DIR)/$(PKG).so $(AUXILIARY_DIR)/$(PKG)/*.go; \
	elif [ -d "$(GLOBAL_DIR)/$(PKG)" ]; then \
		$(GOBUILD) $(PLUGIN_FLAGS) -o $(PLUGIN_BIN_DIR)/$(PKG).so $(GLOBAL_DIR)/$(PKG)/*.go; \
	else \
		echo "$(COLOR_YELLOW)Plugin package $(PKG) not found$(COLOR_RESET)"; \
		exit 1; \
	fi
	@echo "$(COLOR_GREEN)✓ Plugin $(PKG).so built$(COLOR_RESET)"

# 列出所有可用插件
list:
	@echo "$(COLOR_BLUE)Available plugins:$(COLOR_RESET)"
	@$(GO) run tools/plugin_info.go list

# 显示插件信息
info:
	@if [ -z "$(PLUGIN)" ]; then \
		$(GO) run tools/plugin_info.go info; \
	else \
		$(GO) run tools/plugin_info.go info $(PLUGIN); \
	fi

# 测试
test:
	@echo "$(COLOR_BLUE)Running tests...$(COLOR_RESET)"
	@$(GOTEST) -v ./...

# 清理
clean:
	@echo "$(COLOR_YELLOW)Cleaning...$(COLOR_RESET)"
	@$(GOCLEAN)
	@rm -rf $(PLUGIN_BIN_DIR)
	@rm -f BlockChainSimulator
	@echo "$(COLOR_GREEN)✓ Clean completed$(COLOR_RESET)"

# 增量编译（只编译修改过的插件）
incremental:
	@echo "$(COLOR_BLUE)Incremental build...$(COLOR_RESET)"
	@for pkg_path in $(ALL_PACKAGES); do \
		pkg_name=$$(basename $$pkg_path); \
		so_file=$(PLUGIN_BIN_DIR)/$$pkg_name.so; \
		if [ ! -f $$so_file ] || [ $$pkg_path -nt $$so_file ]; then \
			echo "$(COLOR_YELLOW)Rebuilding $$pkg_name...$(COLOR_RESET)"; \
			$(GOBUILD) $(PLUGIN_FLAGS) -o $$so_file $$pkg_path/*.go; \
			echo "$(COLOR_GREEN)✓ Built $$pkg_name.so$(COLOR_RESET)"; \
		fi; \
	done
	@echo "$(COLOR_GREEN)✓ Incremental build completed$(COLOR_RESET)"

# 帮助信息
help:
	@echo "$(COLOR_BLUE)BlockChainSimulator Plugin Build System$(COLOR_RESET)"
	@echo ""
	@echo "$(COLOR_GREEN)Targets:$(COLOR_RESET)"
	@echo "  all         - Build main program and all plugins (default)"
	@echo "  main        - Build only the main program"
	@echo "  plugins     - Build all plugins"
	@echo "  plugin      - Build a single plugin (Usage: make plugin PKG=<name>)"
	@echo "  protocol    - Build plugins for a protocol (Usage: make protocol PROTOCOL=<name>)"
	@echo "  incremental - Incremental build (only modified plugins)"
	@echo "  list        - List all available plugins"
	@echo "  info        - Show plugin information (Usage: make info [PLUGIN=<name>])"
	@echo "  test        - Run tests"
	@echo "  clean       - Remove build artifacts"
	@echo "  help        - Show this help message"
	@echo ""
	@echo "$(COLOR_GREEN)Examples:$(COLOR_RESET)"
	@echo "  make                          # Full build"
	@echo "  make plugin PKG=pbft          # Build only pbft plugin"
	@echo "  make protocol PROTOCOL=RBE    # Build all plugins for RBE protocol"
	@echo "  make incremental              # Incremental build"
	@echo "  make list                     # List all plugins"
	@echo "  make info PLUGIN=PBFT         # Show info for PBFT plugin"
