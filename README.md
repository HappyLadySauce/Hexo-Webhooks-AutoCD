# Hexo Webhooks AutoCD

<div align="center">

![GitHub](https://img.shields.io/github/license/ssddffaa/Hexo-Webhooks-AutoCD)
![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/ssddffaa/Hexo-Webhooks-AutoCD)
![GitHub last commit](https://img.shields.io/github/last-commit/ssddffaa/Hexo-Webhooks-AutoCD)

</div>

这是一个用于自动化部署 Hexo 博客的 Webhook 服务器。当 GitHub 仓库收到 push 事件时，会自动触发部署脚本，实现博客的自动更新。

部署脚本完全自定义，你可以设置你想要的脚本文件（目前仅支持`shell`脚本），**Hexo Webhooks AutoCD** 仅仅只是接受 **Github WebHooks** 并触发你的脚本，为用户提供了高度的自由。

虽然本项目的初衷是为了实现 Hexo 的持续部署（`CD`），但是也可以用于实现其他基于 **Github WebHooks** 的`CD`工作。

## 📝 目录

- [前言](#前言)
- [功能特点](#功能特点)
- [快速开始](#快速开始)
  - [安装](#安装)
  - [配置](#配置)
  - [运行](#运行)
- [部署脚本示例](#部署脚本示例)
- [GitHub Webhook 配置](#github-webhook-配置)
- [最佳实践](#最佳实践)
- [常见问题](#常见问题)
- [安全建议](#安全建议)
- [贡献指南](#贡献指南)
- [许可证](#许可证)

## 🎯 前言

本人 Hexo 部署在VPS服务器上并不是 github 的静态页面。

每次 Hexo 提交文章都需要：
1. `hexo new 新文章` 
2. 把自己 md 文件中的内容拷贝过去
3. 增加分类、添加标签和封面图片
4. `hexo clean && hexo generate && hexo server`

这个流程实在太繁琐了！

### 💡 解决方案

我设计了一套简单的规范来自动化这个过程：

1. 创建专门的 Github 仓库存放 md 文件
2. 使用**目录结构**自动设置文章分类
3. 通过文件名规则设置标签：`文章标题&标签1&标签2&标签3.md`
4. 自动获取随机图片作为文章封面
5. 使用 WebHooks 触发自动部署

## ✨ 功能特点

- 🔒 支持 GitHub Webhooks 安全验证
- 🔐 内置 HTTPS 支持
- 📜 完全自定义的部署脚本
- 🚦 智能的并发控制和超时处理
- 🔍 详细的日志记录
- 🔄 自动重试机制

## 🚀 快速开始

### 安装

1. 克隆仓库：
```bash
git clone https://github.com/ssddffaa/Hexo-Webhooks-AutoCD.git
cd Hexo-Webhooks-AutoCD
```

2. 安装依赖：
```bash
go mod tidy
```

### 配置

1. 复制配置文件示例：
```bash
cp config_example.yaml config.yaml
```

2. 编辑 `config.yaml`：
```yaml
webhook:
    port: 8080                    # 服务器端口
    path: /webhook                # Webhook 路径
    secret: your_secret_key       # GitHub Webhook 密钥

scripts:
    scripts_path: .               # 脚本目录
    push: deploy.sh               # push 事件触发的脚本

ssl:
    enabled: true                 # 是否启用 HTTPS
    cert_file: cert/fullchain.pem # SSL 证书文件路径
    key_file: cert/privkey.pem    # SSL 私钥文件路径
```

## 📜 部署脚本示例

这里提供一个完整的部署脚本示例 `deploy.sh`：

```bash
#!/bin/bash

# 设置错误时退出
set -e

# 日志函数
log() {
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] $1"
}

# 工作目录
BLOG_DIR="/path/to/your/hexo/blog"
POSTS_DIR="/path/to/your/posts/repo"

# 更新文章仓库
cd "$POSTS_DIR"
log "拉取最新文章..."
git pull origin main

# 更新博客
cd "$BLOG_DIR"
log "开始部署博客..."

# 同步文章
log "同步文章文件..."
rsync -av --delete "$POSTS_DIR/source/_posts/" "$BLOG_DIR/source/_posts/"

# 生成静态文件
log "清理缓存..."
hexo clean

log "生成静态文件..."
hexo generate

log "部署完成！"
```

确保脚本具有执行权限：
```bash
chmod +x deploy.sh
```

## 🔧 GitHub Webhook 配置

1. 在 GitHub 仓库中进入：`Settings` → `Webhooks` → `Add webhook`
2. 配置 Webhook：
   ```
   Payload URL: https://your-domain:8080/webhook
   Content type: application/json
   Secret: your_secret_key（与配置文件中的secret一致）
   SSL verification: Enable SSL verification
   Which events would you like to trigger this webhook?: Just the push event
   Active: ✓
   ```

## 💡 最佳实践

1. 文章仓库组织
   ```
   posts/
   ├── 技术/
   │   ├── Go开发&golang&后端&服务器.md
   │   └── Python学习&python&编程.md
   ├── 生活/
   │   └── 旅行日记&旅行&摄影.md
   └── 阅读/
       └── 读书笔记&阅读&笔记.md
   ```

2. 使用 Supervisor 管理服务
   ```ini
   [program:hexo-webhooks]
   command=/path/to/Hexo-Webhooks-AutoCD/hexo-webhooks
   directory=/path/to/Hexo-Webhooks-AutoCD
   autostart=true
   autorestart=true
   stderr_logfile=/var/log/hexo-webhooks.err.log
   stdout_logfile=/var/log/hexo-webhooks.out.log
   ```

## ❓ 常见问题

1. **Q: 如何处理部署失败？**
   
   A: 检查日志文件，确保脚本权限正确，网络连接正常。

2. **Q: 支持哪些 Webhook 事件？**
   
   A: 目前仅支持 push 事件，后续会添加更多事件支持。

3. **Q: 如何自定义部署流程？**
   
   A: 修改 deploy.sh 脚本，添加你需要的任何自定义操作。

## 🔐 安全建议

1. 使用强密码作为 Webhook secret
2. 始终启用 HTTPS
3. 定期更新依赖包
4. 限制部署脚本的权限范围
5. 使用环境变量存储敏感信息
6. 配置防火墙只允许 GitHub 的 IP 范围访问

## 🤝 贡献指南

1. Fork 本仓库
2. 创建特性分支 (`git checkout -b feature/AmazingFeature`)
3. 提交改动 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 提交 Pull Request

## 📄 许可证

本项目采用 MIT 许可证 - 查看 [LICENSE](LICENSE) 文件了解详情 