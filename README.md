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
sudo mkdir -p /etc/hexo-webhooks-autocd/cert

# 复制SSL证书（替换为你的证书路径）
sudo cp /path/to/fullchain.pem /etc/hexo-webhooks-autocd/cert/
sudo cp /path/to/privkey.pem /etc/hexo-webhooks-autocd/cert/

# 设置证书权限
sudo chmod 600 /etc/hexo-webhooks-autocd/cert/*.pem
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

scripts:
    path: /etc/hexo-webhooks-autocd  # 脚本目录
    push: deploy.sh                  # 部署脚本

ssl:
    enabled: true
    cert_file: /etc/hexo-webhooks-autocd/cert/fullchain.pem
    key_file: /etc/hexo-webhooks-autocd/cert/privkey.pem
```

4. 编译安装：
```bash
# 编译
make build

# 安装（需要root权限）
sudo make install
```

5. 配置systemd服务：
```bash
# 复制服务文件
sudo cp hexo-webhooks-autocd.service /etc/systemd/system/

# 重载systemd
sudo systemctl daemon-reload

# 启动服务
sudo systemctl start hexo-webhooks-autocd

# 检查服务状态
sudo systemctl status hexo-webhooks-autocd

# 设置开机自启
sudo systemctl enable hexo-webhooks-autocd
```

## 目录结构

安装完成后的目录结构：
```
/etc/hexo-webhooks-autocd/
├── config.yaml          # 配置文件
├── deploy.sh           # 部署脚本
└── cert/
    ├── fullchain.pem   # SSL证书
    └── privkey.pem     # SSL私钥

/usr/local/bin/
└── hexo-webhooks-autocd  # 可执行文件
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
sudo systemctl status hexo-webhooks-autocd
```

2. 查看实时日志：
```bash
sudo journalctl -u hexo-webhooks-autocd -f
```

3. 查看最近的日志：
```bash
sudo journalctl -u hexo-webhooks-autocd -n 50 --no-pager
```

## 故障排除

1. 服务无法启动：
   - 检查配置文件权限：`ls -l /etc/hexo-webhooks-autocd/config.yaml`
   - 检查证书文件是否存在：`ls -l /etc/hexo-webhooks-autocd/cert/`
   - 检查证书权限：`ls -l /etc/hexo-webhooks-autocd/cert/*.pem`
   - 检查端口是否被占用：`sudo lsof -i :8080`
   - 查看详细错误日志：`sudo journalctl -u hexo-webhooks-autocd -n 50 --no-pager`

2. Webhook调用失败：
   - 检查GitHub Webhook配置是否正确
   - 确认服务器防火墙是否允许8080端口
   - 验证SSL证书是否有效：`openssl x509 -in /etc/hexo-webhooks-autocd/cert/fullchain.pem -text -noout`
   - 检查Webhook密钥是否匹配

3. 部署失败：
   - 检查deploy.sh脚本权限：`ls -l /etc/hexo-webhooks-autocd/deploy.sh`
   - 确认deploy.sh是否有执行权限：`chmod +x /etc/hexo-webhooks-autocd/deploy.sh`
   - 检查Hexo环境是否正确配置
   - 查看部署日志中的具体错误信息

## 许可证

MIT License 