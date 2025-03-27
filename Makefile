.PHONY: all build clean

# 设置Go环境变量
GOOS=linux
GOARCH=amd64
BINARY_NAME=hexo-webhooks-autocd

all: clean build

build:
	@echo "开始构建..."
	@go mod tidy
	@CGO_ENABLED=0 GOOS=$(GOOS) GOARCH=$(GOARCH) go build -o $(BINARY_NAME) -ldflags="-s -w" main.go
	@echo "构建完成: $(BINARY_NAME)"

clean:
	@echo "清理构建文件..."
	@rm -f $(BINARY_NAME)
	@echo "清理完成"

install: build
	@echo "安装服务..."
	@mkdir -p /etc/hexo-webhooks-autocd
	@cp $(BINARY_NAME) /usr/local/bin/
	@cp config.yaml /etc/hexo-webhooks-autocd/
	@cp deploy.sh /etc/hexo-webhooks-autocd/
	@echo "安装完成" 
	@echo "请编辑配置文件 /etc/hexo-webhooks-autocd/config.yaml"
	@echo "如果需要配置证书，请将证书文件放在 /etc/hexo-webhooks-autocd/cert/ 目录下"