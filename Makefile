.PHONY: all build clean config-port config-secret config-ssl

# 颜色定义
GREEN=\033[32m
YELLOW=\033[33m
BLUE=\033[34m
RED=\033[31m
CYAN=\033[36m
RESET=\033[0m

# 设置Go环境变量
GOOS=linux
GOARCH=amd64
BINARY_NAME=hexo-autocd

# 目录定义
INSTALL_DIR=/etc/hexo-autocd
SCRIPTS_DIR=$(INSTALL_DIR)/scripts
CERT_DIR=$(INSTALL_DIR)/cert
LOGS_DIR=$(INSTALL_DIR)/logs
BIN_DIR=/usr/local/bin
SYSTEMD_DIR=/etc/systemd/system

all: clean build

build:
	@echo "$(BLUE)开始构建...$(RESET)"
	@go mod tidy
	@CGO_ENABLED=0 GOOS=$(GOOS) GOARCH=$(GOARCH) go build -o $(BINARY_NAME) -ldflags="-s -w" main.go
	@echo "$(GREEN)✓ 构建完成: $(BINARY_NAME) (大小: $$(ls -lh $(BINARY_NAME) | awk '{print $$5}'))$(RESET)"

clean:
	@echo "$(BLUE)开始清理...$(RESET)"
	@if [ -f $(BINARY_NAME) ]; then \
		rm -f $(BINARY_NAME); \
		echo "$(GREEN)✓ 清理完成$(RESET)"; \
	else \
		echo "$(YELLOW)没有需要清理的文件$(RESET)"; \
	fi

install: build
	@echo "$(BLUE)开始安装服务...$(RESET)"
	
	# 创建目录
	@for dir in $(SCRIPTS_DIR) $(CERT_DIR) $(LOGS_DIR); do \
		mkdir -p $$dir; \
	done
	@echo "$(GREEN)✓ 目录创建完成$(RESET)"

	# 复制程序文件
	@cp $(BINARY_NAME) $(BIN_DIR)/ && \
		echo "$(GREEN)✓ 程序安装成功$(RESET)" || \
		echo "$(RED)✗ 程序安装失败$(RESET)"

	# 复制配置文件
	@if [ ! -f $(INSTALL_DIR)/config.yaml ]; then \
		cp config.yaml $(INSTALL_DIR)/ && echo "$(GREEN)✓ 配置文件复制成功$(RESET)"; \
	else \
		echo "$(YELLOW)! 配置文件已存在，跳过复制$(RESET)"; \
	fi

	# 复制部署脚本
	@if [ ! -f $(SCRIPTS_DIR)/deploy.sh ]; then \
		cp scripts/deploy.sh $(SCRIPTS_DIR)/ && \
		chmod +x $(SCRIPTS_DIR)/deploy.sh && \
		echo "$(GREEN)✓ 部署脚本复制成功$(RESET)"; \
	else \
		echo "$(YELLOW)! 部署脚本已存在，跳过复制$(RESET)"; \
	fi

	# 复制SSL证书
	@if [ ! -f $(CERT_DIR)/fullchain.pem ] && [ -f cert/fullchain.pem ]; then \
		cp cert/fullchain.pem $(CERT_DIR)/ && \
		echo "$(GREEN)✓ SSL证书复制成功$(RESET)"; \
	elif [ -f $(CERT_DIR)/fullchain.pem ]; then \
		echo "$(YELLOW)! SSL证书已存在，跳过复制$(RESET)"; \
	fi

	@if [ ! -f $(CERT_DIR)/privkey.pem ] && [ -f cert/privkey.pem ]; then \
		cp cert/privkey.pem $(CERT_DIR)/ && \
		echo "$(GREEN)✓ SSL私钥复制成功$(RESET)"; \
	elif [ -f $(CERT_DIR)/privkey.pem ]; then \
		echo "$(YELLOW)! SSL私钥已存在，跳过复制$(RESET)"; \
	fi

	# 安装hexo systemd服务
	@if [ -f systemd/hexo.service ]; then \
		cp systemd/hexo.service $(SYSTEMD_DIR)/ && \
		systemctl daemon-reload && \
		echo "$(GREEN)✓ hexo systemd服务安装成功$(RESET)"; \
	else \
		echo "$(RED)✗ hexo systemd服务文件不存在$(RESET)"; \
	fi

	# 安装hexo-autocd systemd服务
	@if [ -f systemd/hexo-autocd.service ]; then \
		cp systemd/hexo-autocd.service $(SYSTEMD_DIR)/ && \
		systemctl daemon-reload && \
		echo "$(GREEN)✓ hexo-autocd systemd服务安装成功$(RESET)"; \
	else \
		echo "$(RED)✗ hexo-autocd systemd服务文件不存在$(RESET)"; \
	fi
	@echo "$(YELLOW)提示: 执行以下命令启用并启动服务:$(RESET)"
	@echo "  sudo systemctl daemon-reload"
	@echo "  sudo systemctl enable hexo.service"
	@echo "  sudo systemctl start hexo.service"
	@echo "  sudo systemctl enable hexo-autocd.service"
	@echo "  sudo systemctl start hexo-autocd.service"

	@echo "\n$(GREEN)✓ 安装完成！$(RESET)"
	@echo "$(BLUE)═══════════════════════════════════════════$(RESET)"
	@echo "$(BLUE)  安装路径：$(RESET)"
	@echo "  $(CYAN)- 程序文件：$(BIN_DIR)/$(BINARY_NAME)$(RESET)"
	@echo "  $(CYAN)- 配置文件：$(INSTALL_DIR)/config.yaml$(RESET)"
	@echo "  $(CYAN)- 部署脚本：$(SCRIPTS_DIR)/deploy.sh$(RESET)"
	@echo "  $(CYAN)- SSL证书：$(CERT_DIR)/fullchain.pem, $(CERT_DIR)/privkey.pem$(RESET)"
	@echo "  $(CYAN)- 日志目录：$(LOGS_DIR)$(RESET)"
	@echo "  $(CYAN)- 服务文件：$(SYSTEMD_DIR)/hexo-autocd.service$(RESET)"
	@echo "$(BLUE)  后续配置：$(RESET)"
	@echo "  $(YELLOW)1. 编辑配置文件：$(INSTALL_DIR)/config.yaml$(RESET)"
	@echo "  $(YELLOW)2. 编辑部署脚本：$(SCRIPTS_DIR)/deploy.sh$(RESET)"
	@echo "  $(YELLOW)3. 编辑Hexo的路径：vim $(SYSTEMD_DIR)/hexo.service$(RESET)"
	@echo "  $(YELLOW)4. 重新加载systemd：sudo systemctl daemon-reload$(RESET)"
	@echo "  $(YELLOW)5. 启用开机启动并启动服务：systemctl enable --now hexo.service$(RESET)"
	@echo "  $(YELLOW)6. 启用开机启动并启动服务：systemctl enable --now hexo-autocd.service$(RESET)"
	@echo "$(BLUE)═══════════════════════════════════════════$(RESET)"

uninstall:
	@echo "$(BLUE)开始卸载服务...$(RESET)"
	@echo "$(YELLOW)! 警告：这将删除以下文件和目录：$(RESET)"
	@echo "$(CYAN)  - $(BIN_DIR)/$(BINARY_NAME)$(RESET)"
	@echo "$(CYAN)  - $(INSTALL_DIR)$(RESET)"
	@echo "$(CYAN)  - $(SYSTEMD_DIR)/hexo-autocd.service$(RESET)"
	@echo "$(CYAN)  - $(SYSTEMD_DIR)/hexo.service$(RESET)"
	@read -p "$(RED)确定要继续吗？[y/N] $(RESET)" confirm; \
	if [ "$$confirm" = "y" ] || [ "$$confirm" = "Y" ]; then \
		systemctl stop hexo-autocd.service 2>/dev/null || true; \
		systemctl disable hexo-autocd.service 2>/dev/null || true; \
		rm -f $(SYSTEMD_DIR)/hexo-autocd.service; \
		systemctl daemon-reload; \
		rm -f $(BIN_DIR)/$(BINARY_NAME); \
		rm -rf $(INSTALL_DIR); \
		echo "$(GREEN)✓ 卸载完成$(RESET)"; \
	else \
		echo "$(YELLOW)取消卸载$(RESET)"; \
	fi