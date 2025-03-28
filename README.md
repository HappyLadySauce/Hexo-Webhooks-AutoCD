# Hexo Webhooks AutoCD

<div align="center">

![GitHub](https://img.shields.io/github/license/ssddffaa/Hexo-Webhooks-AutoCD)
![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/ssddffaa/Hexo-Webhooks-AutoCD)
![GitHub last commit](https://img.shields.io/github/last-commit/ssddffaa/Hexo-Webhooks-AutoCD)

</div>

这是一个用于自动部署Hexo博客的Webhook服务器。当GitHub仓库收到push事件时，会自动触发博客的更新和部署。

## 功能特点

- 支持GitHub Webhooks
- 自动同步markdown文件
- 自动处理文章的front-matter
- 支持文章分类和标签
- 自动部署Hexo博客
- 支持HTTPS
- 系统服务自动管理

## 安装要求

- Go 1.16+
- Linux系统
- systemd
- Git
- Hexo
- SSL证书（用于HTTPS）

## 安装步骤

1. 克隆仓库：
```bash
git clone https://github.com/yourusername/Hexo-Webhooks-AutoCD.git
cd Hexo-Webhooks-AutoCD
```

2. 准备SSL证书：
```bash
# 创建证书目录
mkdir -p cert

# 复制SSL证书（替换为你的证书路径）
cp /path/to/fullchain.pem cert/
cp /path/to/privkey.pem cert/

# 设置证书权限
chmod 600 cert/*.pem
```

3. 配置：
```bash
# 复制示例配置文件
cp config_example.yaml config.yaml

# 编辑配置文件
vim config.yaml

# 配置文件内容示例：
webhook:
    port: 8080            # 服务监听端口
    path: /webhook        # Webhook路径
    secret: your_secret   # GitHub Webhook密钥

logs:
    path: /etc/hexo-autocd/logs/webhooks.log
    level: info           # 日志级别，可选：trace、debug、info、warn、error、fatal、panic
    format: text          # 日志格式，可选：json、text
    max_size: 100         # 日志文件最大大小（MB）
    max_backups: 5        # 日志文件最大备份数
    max_age: 30           # 日志文件最大保存时间（天）

scripts:
    path: /etc/hexo-autocd/scripts
    push: deploy.sh       # 部署脚本
    timeout: 5m           # 脚本执行超时时间
    max_concurrent: 5     # 最大并发执行数

ssl:
    enabled: true
    cert_file: /etc/hexo-autocd/cert/fullchain.pem
    key_file: /etc/hexo-autocd/cert/privkey.pem
```

4. 编译安装：
```bash
# 编译并安装（需要root权限）
sudo make install
```

安装过程将自动完成以下工作：
- 编译生成可执行文件
- 创建必要的目录结构
- 复制配置文件和脚本
- 安装systemd服务
- 配置SSL证书

5. 启动服务：
```bash
# 重载systemd
sudo systemctl daemon-reload

# 启用并启动服务
sudo systemctl enable --now hexo-autocd.service

# 检查服务状态
sudo systemctl status hexo-autocd.service
```

## 目录结构

安装完成后的目录结构：
```
/etc/hexo-autocd/
├── config.yaml          # 配置文件
├── scripts/
│   └── deploy.sh        # 部署脚本
├── cert/
│   ├── fullchain.pem    # SSL证书
│   └── privkey.pem      # SSL私钥
└── logs/                # 日志目录

/usr/local/bin/
└── hexo-autocd          # 可执行文件

/etc/systemd/system/
└── hexo-autocd.service  # hexo-autocd systemd服务文件
└── hexo.service  # hexo systemd服务文件
```

## 部署脚本

默认的部署脚本位于 `/etc/hexo-autocd/scripts/deploy.sh`，你可以根据自己的博客部署需求修改此脚本，也可以直接使用：

```shell
#!/bin/bash

# 日志函数
log() {
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] $1"
}

# 配置Git的函数
setup_git() {
    log "配置Git环境..."
    # 配置Git安全目录
    git config --global --add safe.directory "$POSTS_DIR"
    # 配置Git用户信息（如果需要）
    git config --global user.name "HappyLadySauce"
    git config --global user.email "13452552349@163.com"
}

# 停止hexo进程的函数
stop_hexo() {
    log "正在停止hexo服务..."
    if systemctl is-active hexo >/dev/null 2>&1; then
        if ! sudo systemctl stop hexo; then
            log "警告：停止hexo服务失败"
            return 1
        fi
        # 等待服务完全停止
        local count=0
        while systemctl is-active hexo >/dev/null 2>&1; do
            sleep 1
            count=$((count + 1))
            if [ $count -ge 10 ]; then
                log "错误：hexo服务停止超时"
                return 1
            fi
        done
        log "hexo服务已停止"
    else
        log "hexo服务未运行"
    fi
    # 等待3秒
    sleep 3
    return 0
}

# 启动hexo服务的函数
start_hexo() {
    log "正在启动hexo服务..."
    if ! sudo systemctl start hexo; then
        log "错误：启动hexo服务失败"
        return 1
    fi
    
    # 等待服务启动并检查状态
    local count=0
    while ! systemctl is-active hexo >/dev/null 2>&1; do
        sleep 1
        count=$((count + 1))
        if [ $count -ge 10 ]; then
            log "错误：hexo服务启动超时"
            return 1
        fi
    done
    
    log "hexo服务已成功启动"
    return 0
}

# 工作目录
BLOG_DIR="/home/hexo/blog"
POSTS_DIR="/home/hexo/markdown"
BLOG_POSTS_DIR="$BLOG_DIR/source/_posts"

# 处理单个文件的函数
process_file() {
    local src_file="$1"
    local rel_path="${src_file#$POSTS_DIR/}"  # 获取相对路径
    local filename=$(basename "$src_file")
    local dir_path=$(dirname "$rel_path")
    local dest_file="$BLOG_POSTS_DIR/$(basename "$src_file")"
    
    # 获取分类（目录路径）
    local categories=""
    if [ "$dir_path" != "." ]; then
        categories=$(echo "$dir_path" | tr '/' '\n' | sed 's/^/    - /')
    fi
    
    # 获取标签（从文件名中提取）
    local tags=""
    if [[ "$filename" == *"&"* ]]; then
        # 提取&后面的部分（不包括.md）
        local tag_part="${filename#*&}"
        tag_part="${tag_part%.md}"
        # 将&分隔的标签转换为yaml格式
        tags=$(echo "$tag_part" | tr '&' '\n' | sed 's/^/- /')
    fi
    
    # 生成随机封面URL
    cover_base_url="https://lsky.happyladysauce.cn/i/1/"
    cover_url="$cover_base_url$(($RANDOM % 10)).webp"

    # 创建临时文件
    local temp_file=$(mktemp)
    
    # 先写入新的front-matter
    {
        echo "---"
        echo "title: ${filename%%&*}"  # 使用&之前的部分作为标题
        echo "date: $(date '+%Y-%m-%d %H:%M:%S')"
        echo "categories:"
        if [ -n "$categories" ]; then
            echo "$categories"
        fi
        echo "tags:"
        if [ -n "$tags" ]; then
            # 确保tags缩进正确
            echo "$tags" | sed 's/^-/  -/'
        fi
        echo "cover: $cover_url"
        echo "ai: true"
        echo "---"
        echo ""  # 添加空行
    } > "$temp_file"
    
    # 检查文件是否已存在front-matter，如果存在则跳过它
    if head -n 1 "$src_file" | grep -q "^---$"; then
        # 找到第二个 "---" 的行号
        local end_line=$(awk '/^---$/ {count++; if (count==2) {print NR; exit}}' "$src_file")
        if [ -n "$end_line" ]; then
            # 只添加front-matter之后的内容
            tail -n +$((end_line + 1)) "$src_file" >> "$temp_file"
        else
            # 如果没找到第二个 "---"，添加整个文件内容
            cat "$src_file" >> "$temp_file"
        fi
    else
        # 如果不存在front-matter，直接添加文件内容
        cat "$src_file" >> "$temp_file"
    fi
    
    # 移动临时文件到目标位置
    mv "$temp_file" "$src_file"
    cp "$src_file" "$dest_file"
    
    log "处理文件: $filename"
}

# 输出提交信息
log "收到新的提交："
log "提交ID: $COMMIT_ID"
log "提交信息: $COMMIT_MESSAGE"
log "提交时间: $COMMIT_TIMESTAMP"
log "新增文件: $COMMIT_ADDED"
log "删除文件: $COMMIT_REMOVED"
log "修改文件: $COMMIT_MODIFIED"

# 更新文章仓库
cd "$POSTS_DIR" || exit 1
log "拉取最新文章..."

# 配置Git环境
setup_git

# 重置所有本地更改
git reset --hard HEAD
git clean -fd  # 删除未跟踪的文件和目录

# 拉取最新更改
git pull origin main

# 检查是否有错误
if [ $? -ne 0 ]; then
    log "错误：拉取文章失败"
    exit 1
fi

# 处理新增和修改的文件
IFS=',' read -ra ADDED_FILES <<< "$COMMIT_ADDED"
IFS=',' read -ra MODIFIED_FILES <<< "$COMMIT_MODIFIED"

for file in "${ADDED_FILES[@]}" "${MODIFIED_FILES[@]}"; do
    if [ -n "$file" ] && [ -f "$POSTS_DIR/$file" ]; then
        process_file "$POSTS_DIR/$file"
    fi
done

# 处理删除的文件
IFS=',' read -ra REMOVED_FILES <<< "$COMMIT_REMOVED"
for file in "${REMOVED_FILES[@]}"; do
    if [ -n "$file" ]; then
        rm -f "$BLOG_POSTS_DIR/$(basename "$file")"
        log "删除文件: $(basename "$file")"
    fi
done

# 更新博客
cd "$BLOG_DIR" || exit 1
log "开始部署博客..."

sleep 1

# 停止hexo服务
if ! stop_hexo; then
    log "警告：继续部署，但hexo服务可能未完全停止"
fi

# 清理缓存
log "清理 Hexo 缓存..."
hexo clean

# 生成静态文件
log "生成静态文件..."
hexo generate

# 检查生成是否成功
if [ $? -ne 0 ]; then
    log "错误：生成静态文件失败"
    exit 1
fi

# 启动hexo服务
if ! start_hexo; then
    log "错误：部署完成但服务启动失败"
    exit 1
fi

log "部署完成！"
exit 0
```

## GitHub Webhook配置

1. 在GitHub仓库设置中添加Webhook：
   - 进入仓库 Settings -> Webhooks -> Add webhook
   - Payload URL: `https://your-domain.com:8080/webhook`
   - Content type: `application/json`
   - Secret: 与config.yaml中的secret相同
   - SSL verification: Enable SSL verification
   - Events: 选择 `push` 事件
   - Active: ✓ 勾选

2. 确保仓库有适当的访问权限

## 日志查看

1. 查看服务状态：
```bash
sudo systemctl status hexo-autocd
```

2. 查看实时日志：
```bash
sudo journalctl -u hexo-autocd -f
```

3. 查看最近的日志：
```bash
sudo journalctl -u hexo-autocd -n 50 --no-pager
```

4. 查看应用日志文件：
```bash
sudo cat /etc/hexo-autocd/logs/webhooks.log
```

## 卸载服务

如需卸载服务，可以执行：
```bash
sudo make uninstall
```

这将停止并禁用服务，移除所有相关文件。

## 故障排除

1. 服务无法启动：
   - 检查配置文件权限：`ls -l /etc/hexo-autocd/config.yaml`
   - 检查证书文件是否存在：`ls -l /etc/hexo-autocd/cert/`
   - 检查证书权限：`ls -l /etc/hexo-autocd/cert/*.pem`
   - 检查端口是否被占用：`sudo lsof -i :8080`
   - 查看详细错误日志：`sudo journalctl -u hexo-autocd -n 50 --no-pager`

2. Webhook调用失败：
   - 检查GitHub Webhook配置是否正确
   - 确认服务器防火墙是否允许8080端口
   - 验证SSL证书是否有效：`openssl x509 -in /etc/hexo-autocd/cert/fullchain.pem -text -noout`
   - 检查Webhook密钥是否匹配

3. 部署失败：
   - 检查deploy.sh脚本权限：`ls -l /etc/hexo-autocd/scripts/deploy.sh`
   - 确认deploy.sh是否有执行权限：`chmod +x /etc/hexo-autocd/scripts/deploy.sh`
   - 检查Hexo环境是否正确配置
   - 查看部署日志中的具体错误信息

4. 配置文件路径错误：
   - 注意：如果出现 "no such file or directory" 错误，请检查配置文件中的路径是否与实际安装路径一致
   - 确保所有路径使用 `/etc/hexo-autocd/` 而非 `/etc/hexo-webhooks-autocd/`

## 许可证

MIT License 