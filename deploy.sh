#!/bin/bash

# =============================================================================
# PetMiner 一键部署脚本 - 部署到远程服务器
# =============================================================================

set -e

# 项目配置
PROJECT_NAME="petminer"
VERSION="1.0.0"
BUILD_TIME=$(date -u '+%Y-%m-%dT%H:%M:%SZ')
GIT_COMMIT=$(git rev-parse --short HEAD 2>/dev/null || echo 'unknown')

# 本地配置
PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
LOCAL_BUILD_DIR="$PROJECT_ROOT/dist/deploy-build"
FRONTEND_DIR="$PROJECT_ROOT/frontend"
BACKEND_DIR="$PROJECT_ROOT/backend"

# 远程服务器配置
REMOTE_HOST="154.29.153.211"
REMOTE_PORT="36602"
REMOTE_USER="root"
REMOTE_APP_DIR="/app"
REMOTE_STATIC_DIR="/app/static"
REMOTE_BIN_DIR="/app/bin"
REMOTE_SERVICE_NAME="petminer"

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
NC='\033[0m'

# 日志函数
log() {
    echo -e "${BLUE}[$(date '+%H:%M:%S')]${NC} $1"
}

error() {
    echo -e "${RED}[ERROR]${NC} $1" >&2
}

success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

progress() {
    echo -e "${PURPLE}[PROGRESS]${NC} $1"
}

# 检查依赖
check_dependencies() {
    log "检查本地构建依赖..."
    
    # 检查SSH连接
    if ! ssh -p "$REMOTE_PORT" -o ConnectTimeout=10 -o BatchMode=yes "$REMOTE_USER@$REMOTE_HOST" exit 2>/dev/null; then
        error "无法连接到服务器 $REMOTE_HOST"
        error "请确保已配置SSH免密登录"
        exit 1
    fi
    
    # 检查本地工具
    local missing_tools=()
    for tool in node npm go rsync; do
        if ! command -v "$tool" &> /dev/null; then
            missing_tools+=("$tool")
        fi
    done
    
    if [ ${#missing_tools[@]} -gt 0 ]; then
        error "缺少必要工具: ${missing_tools[*]}"
        exit 1
    fi
    
    success "依赖检查完成"
}

# 检查远程服务器环境
check_remote_environment() {
    log "检查远程服务器环境..."
    
    # 检查远程工具
    ssh -p "$REMOTE_PORT" "$REMOTE_USER@$REMOTE_HOST" bash << 'EOF'
        echo "系统信息: $(uname -a)"
        echo "Go版本: $(go version 2>/dev/null || echo '未安装')"
        echo "Node版本: $(node --version 2>/dev/null || echo '未安装')"
        echo "可用内存: $(free -h | grep Mem | awk '{print $7}' 2>/dev/null || echo '未知')"
        echo "磁盘空间: $(df -h / | tail -1 | awk '{print $4}' 2>/dev/null || echo '未知')"
        
        # 检查端口占用
        if netstat -tlnp 2>/dev/null | grep -q ":8081 "; then
            echo "警告: 端口8081已被占用"
        fi
EOF
    
    success "远程环境检查完成"
}

# 本地构建
build_local() {
    log "开始本地构建..."
    
    # 清理构建目录
    rm -rf "$LOCAL_BUILD_DIR"
    mkdir -p "$LOCAL_BUILD_DIR"
    
    # 构建前端
    progress "构建前端应用..."
    cd "$FRONTEND_DIR"
    
    # 安装依赖
    npm ci --silent
    
    # 设置生产环境变量
    export NODE_ENV=production
    export REACT_APP_API_URL="/api/v1"
    export REACT_APP_WS_URL="/ws"
    export GENERATE_SOURCEMAP=false
    
    # 构建前端
    npm run build
    
    # 复制前端文件到构建目录
    if command -v rsync &> /dev/null; then
        rsync -av build/ "$LOCAL_BUILD_DIR/static/"
    else
        cp -r build/. "$LOCAL_BUILD_DIR/static/"
    fi
    
    # 验证关键前端文件
    local frontend_files=("index.html" "favicon.ico" "logo.png")
    for file in "${frontend_files[@]}"; do
        if [ ! -f "$LOCAL_BUILD_DIR/static/$file" ]; then
            error "前端文件缺失: $file"
            exit 1
        fi
    done
    
    success "前端构建完成"
    
    # 构建后端
    progress "构建后端应用..."
    cd "$BACKEND_DIR"
    
    # 下载依赖
    go mod download
    go mod verify
    
    # 构建Linux二进制文件
    mkdir -p "$LOCAL_BUILD_DIR/bin"
    env GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build \
        -ldflags "-w -s -X main.Version=$VERSION -X main.BuildTime=$BUILD_TIME -X main.GitCommit=$GIT_COMMIT" \
        -o "$LOCAL_BUILD_DIR/bin/$PROJECT_NAME-server" \
        ./cmd/server/main.go
    
    if [ ! -f "$LOCAL_BUILD_DIR/bin/$PROJECT_NAME-server" ]; then
        error "后端构建失败"
        exit 1
    fi
    
    success "后端构建完成"
    
    cd "$PROJECT_ROOT"
}

# 创建部署文件
create_deploy_files() {
    log "创建部署配置文件..."
    
    # 创建systemd服务文件
    cat > "$LOCAL_BUILD_DIR/petminer.service" << EOF
[Unit]
Description=PetMiner Virtual Pet Mining Game
After=network.target
Wants=network.target

[Service]
Type=simple
User=root
WorkingDirectory=$REMOTE_APP_DIR
ExecStart=$REMOTE_BIN_DIR/$PROJECT_NAME-server
Environment=GIN_MODE=release
Environment=PORT=8081
Environment=GOOS=linux
Restart=always
RestartSec=5
StandardOutput=journal
StandardError=journal
SyslogIdentifier=$PROJECT_NAME

# 安全配置
NoNewPrivileges=true
PrivateTmp=true
ProtectSystem=strict
ProtectHome=true
ReadWritePaths=$REMOTE_APP_DIR

# 资源限制
LimitNOFILE=65536
MemoryMax=512M

[Install]
WantedBy=multi-user.target
EOF

    # 创建nginx配置
    cat > "$LOCAL_BUILD_DIR/nginx.conf" << EOF
events {
    worker_connections 1024;
}

http {
    include       /etc/nginx/mime.types;
    default_type  application/octet-stream;
    
    # 日志配置
    access_log /var/log/nginx/petminer-access.log;
    error_log /var/log/nginx/petminer-error.log;
    
    # 基本配置
    sendfile on;
    tcp_nopush on;
    tcp_nodelay on;
    keepalive_timeout 65;
    client_max_body_size 10M;
    
    # gzip压缩
    gzip on;
    gzip_vary on;
    gzip_min_length 1024;
    gzip_types
        text/plain
        text/css
        text/xml
        text/javascript
        application/javascript
        application/xml+rss
        application/json;

    server {
        listen 80;
        server_name _;
        root $REMOTE_STATIC_DIR;
        index index.html;
        
        # 安全头部
        add_header X-Frame-Options "SAMEORIGIN" always;
        add_header X-XSS-Protection "1; mode=block" always;
        add_header X-Content-Type-Options "nosniff" always;
        add_header Referrer-Policy "no-referrer-when-downgrade" always;
        
        # API请求代理到后端
        location /api/ {
            proxy_pass http://127.0.0.1:8081;
            proxy_set_header Host \$host;
            proxy_set_header X-Real-IP \$remote_addr;
            proxy_set_header X-Forwarded-For \$proxy_add_x_forwarded_for;
            proxy_set_header X-Forwarded-Proto \$scheme;
            
            # 超时配置
            proxy_connect_timeout 30s;
            proxy_send_timeout 30s;
            proxy_read_timeout 30s;
        }
        
        # WebSocket代理到后端
        location /ws {
            proxy_pass http://127.0.0.1:8081;
            proxy_http_version 1.1;
            proxy_set_header Upgrade \$http_upgrade;
            proxy_set_header Connection "upgrade";
            proxy_set_header Host \$host;
            proxy_set_header X-Real-IP \$remote_addr;
            proxy_set_header X-Forwarded-For \$proxy_add_x_forwarded_for;
            proxy_set_header X-Forwarded-Proto \$scheme;
            
            # WebSocket特定超时配置
            proxy_connect_timeout 30s;
            proxy_send_timeout 30s;
            proxy_read_timeout 300s;
        }
        
        # 静态文件服务
        location / {
            try_files \$uri \$uri/ /index.html;
            
            # 缓存静态资源
            location ~* \.(js|css|png|jpg|jpeg|gif|ico|svg)$ {
                expires 1y;
                add_header Cache-Control "public, immutable";
            }
            
            # HTML文件不缓存
            location ~* \.html$ {
                expires -1;
                add_header Cache-Control "no-cache, no-store, must-revalidate";
            }
        }
        
        # 版本信息端点
        location /version {
            proxy_pass http://127.0.0.1:8081/version;
            proxy_set_header Host \$host;
        }
        
        # 健康检查端点
        location /health {
            access_log off;
            return 200 "healthy\\n";
            add_header Content-Type text/plain;
        }
        
        # 错误页面
        error_page 404 /index.html;
        error_page 500 502 503 504 /50x.html;
        
        location = /50x.html {
            root $REMOTE_STATIC_DIR;
        }
    }
}
EOF

    # 创建部署脚本
    cat > "$LOCAL_BUILD_DIR/setup-remote.sh" << 'EOF'
#!/bin/bash

# 远程服务器设置脚本

set -e

echo "开始远程服务器配置..."

# 创建应用目录
mkdir -p /app/bin /app/static /app/logs /app/data

# 备份现有数据库（如果存在）
if [ -f /app/data/petminer.db ]; then
    echo "备份现有数据库..."
    cp /app/data/petminer.db /app/data/petminer.db.backup.$(date +%Y%m%d_%H%M%S)
    echo "数据库备份完成"
fi

# 设置权限
chmod +x /app/bin/petminer-server

# 停止旧服务（如果存在）
systemctl stop petminer || true
systemctl disable petminer || true

# 安装新服务
cp /app/petminer.service /etc/systemd/system/
systemctl daemon-reload
systemctl enable petminer

# 备份现有nginx配置
if [ -f /etc/nginx/nginx.conf ]; then
    cp /etc/nginx/nginx.conf /etc/nginx/nginx.conf.backup.$(date +%Y%m%d_%H%M%S)
fi

# 安装新nginx配置
cp /app/nginx.conf /etc/nginx/

# 测试nginx配置
nginx -t

# 重启服务
systemctl restart nginx
systemctl start petminer

# 启用开机自启
systemctl enable nginx
systemctl enable petminer

echo "服务配置完成！"
echo ""
echo "服务状态:"
systemctl status petminer --no-pager || true
echo ""
echo "nginx状态:"
systemctl status nginx --no-pager || true
echo ""
echo "服务端口检查:"
netstat -tlnp | grep -E ":(80|8081) " || echo "端口检查失败"
echo ""
echo "访问地址: http://$(hostname -I | awk '{print $1}')"
EOF

    chmod +x "$LOCAL_BUILD_DIR/setup-remote.sh"
    
    # 创建部署信息文件
    cat > "$LOCAL_BUILD_DIR/deploy-info.json" << EOF
{
  "version": "$VERSION",
  "buildTime": "$BUILD_TIME",
  "gitCommit": "$GIT_COMMIT",
  "deployTime": "$(date -u '+%Y-%m-%dT%H:%M:%SZ')",
  "remoteHost": "$REMOTE_HOST",
  "deployPath": "$REMOTE_APP_DIR"
}
EOF
    
    success "部署文件创建完成"
}

# 上传文件到服务器
upload_files() {
    log "上传文件到服务器..."
    
    # 使用rsync同步文件，排除数据目录以保留数据库
    progress "同步应用文件..."
    rsync -avz -e "ssh -p $REMOTE_PORT" --delete --exclude='data/' --exclude='*.db' "$LOCAL_BUILD_DIR/" "$REMOTE_USER@$REMOTE_HOST:$REMOTE_APP_DIR/"
    
    # 验证上传
    ssh -p "$REMOTE_PORT" "$REMOTE_USER@$REMOTE_HOST" bash << EOF
        echo "验证上传文件..."
        
        # 检查关键文件
        if [ ! -f "$REMOTE_APP_DIR/bin/$PROJECT_NAME-server" ]; then
            echo "错误: 后端二进制文件未找到"
            exit 1
        fi
        
        if [ ! -f "$REMOTE_APP_DIR/static/index.html" ]; then
            echo "错误: 前端入口文件未找到"
            exit 1
        fi
        
        if [ ! -f "$REMOTE_APP_DIR/static/favicon.ico" ]; then
            echo "错误: 网站图标文件未找到"
            exit 1
        fi
        
        if [ ! -f "$REMOTE_APP_DIR/static/logo.png" ]; then
            echo "错误: Logo文件未找到"
            exit 1
        fi
        
        echo "文件验证通过"
        echo "静态文件列表:"
        find "$REMOTE_APP_DIR/static" -name "*.png" -o -name "*.ico" -o -name "*.html" | sort
EOF
    
    success "文件上传完成"
}

# 配置并启动远程服务
setup_remote_services() {
    log "配置远程服务..."
    
    # 执行远程配置脚本
    ssh -p "$REMOTE_PORT" "$REMOTE_USER@$REMOTE_HOST" "cd $REMOTE_APP_DIR && bash setup-remote.sh"
    
    success "远程服务配置完成"
}

# 健康检查
health_check() {
    log "执行健康检查..."
    
    local max_attempts=30
    local attempt=1
    
    while [ $attempt -le $max_attempts ]; do
        progress "健康检查 ($attempt/$max_attempts)..."
        # 检查外部访问
        local health_response=$(curl -f -s --connect-timeout 5 "http://$REMOTE_HOST/health" 2>/dev/null || echo "failed")
        
        if [ "$health_response" = "healthy" ]; then
            success "健康检查通过！"
            return 0
        fi
        
        if [ $attempt -eq $max_attempts ]; then
            error "健康检查失败，查看服务日志:"
            ssh -p "$REMOTE_PORT" "$REMOTE_USER@$REMOTE_HOST" "journalctl -u petminer --no-pager --lines=20"
            return 1
        fi
        
        sleep 2
        attempt=$((attempt + 1))
    done
}

# 显示部署结果
show_deployment_summary() {
    log "获取部署信息..."
    
    local remote_ip=$(ssh -p "$REMOTE_PORT" "$REMOTE_USER@$REMOTE_HOST" "hostname -I | awk '{print \$1}'" 2>/dev/null || echo "$REMOTE_HOST")
    
    echo ""
    success "🎉 部署完成！"
    echo ""
    echo "📋 部署信息："
    echo "  服务器: $REMOTE_HOST"
    echo "  部署路径: $REMOTE_APP_DIR"
    echo "  版本: $VERSION"
    echo "  构建时间: $BUILD_TIME"
    echo ""
    echo "🌐 访问地址："
    echo "  主页: http://$remote_ip"
    echo "  API: http://$remote_ip/api/v1/pets"
    echo "  版本信息: http://$remote_ip/version"
    echo "  健康检查: http://$remote_ip/health"
    echo ""
    echo "🔧 服务管理命令："
    echo "  查看状态: ssh $REMOTE_USER@$REMOTE_HOST systemctl status petminer"
    echo "  查看日志: ssh $REMOTE_USER@$REMOTE_HOST journalctl -u petminer -f"
    echo "  重启服务: ssh $REMOTE_USER@$REMOTE_HOST systemctl restart petminer"
    echo ""
    echo "📁 远程文件："
    echo "  应用目录: $REMOTE_APP_DIR"
    echo "  静态文件: $REMOTE_STATIC_DIR"
    echo "  二进制文件: $REMOTE_BIN_DIR/$PROJECT_NAME-server"
    echo "  数据库目录: $REMOTE_APP_DIR/data (持久化保留)"
    echo "  服务配置: /etc/systemd/system/petminer.service"
    echo "  Nginx配置: /etc/nginx/nginx.conf"
    echo ""
}

# 清理本地临时文件
cleanup() {
    if [ -d "$LOCAL_BUILD_DIR" ]; then
        log "清理本地构建文件..."
        rm -rf "$LOCAL_BUILD_DIR"
        success "清理完成"
    fi
}

# 显示帮助信息
show_help() {
    cat << EOF
PetMiner 一键部署脚本

用法: $0 [选项]

选项:
    deploy        执行完整部署 (默认)
    build         仅本地构建
    upload        仅上传文件 (需要先构建)
    setup         仅配置服务 (需要先上传)
    health        仅健康检查
    status        查看远程服务状态
    logs          查看远程服务日志
    clean         清理本地构建文件
    help          显示此帮助信息

示例:
    $0            # 完整部署
    $0 deploy     # 完整部署
    $0 build      # 仅构建
    $0 status     # 查看服务状态
    $0 logs       # 查看服务日志

远程服务器: $REMOTE_HOST
部署目录: $REMOTE_APP_DIR

注意事项:
- 确保已配置SSH免密登录
- 确保远程服务器已安装Go和Node.js环境
- 部署会自动配置systemd服务和nginx
- 数据库文件(/app/data/)在部署时会被保留

EOF
}

# 主函数
main() {
    local command=${1:-"deploy"}
    
    case $command in
        "deploy"|"")
            log "🚀 开始 PetMiner 一键部署"
            check_dependencies
            check_remote_environment
            build_local
            create_deploy_files
            upload_files
            setup_remote_services
            health_check
            show_deployment_summary
            cleanup
            ;;
        "build")
            check_dependencies
            build_local
            create_deploy_files
            success "本地构建完成"
            ;;
        "upload")
            if [ ! -d "$LOCAL_BUILD_DIR" ]; then
                error "本地构建文件不存在，请先执行构建"
                exit 1
            fi
            upload_files
            ;;
        "setup")
            setup_remote_services
            ;;
        "health")
            health_check
            ;;
        "status")
            ssh -p "$REMOTE_PORT" "$REMOTE_USER@$REMOTE_HOST" "systemctl status petminer nginx --no-pager"
            ;;
        "logs")
            ssh -p "$REMOTE_PORT" "$REMOTE_USER@$REMOTE_HOST" "journalctl -u petminer -f"
            ;;
        "clean")
            cleanup
            ;;
        "help"|*)
            show_help
            ;;
    esac
}

# 捕获中断信号，确保清理
trap cleanup EXIT

# 执行主函数
main "$@"