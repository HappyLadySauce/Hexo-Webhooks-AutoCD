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

