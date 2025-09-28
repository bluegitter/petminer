#!/bin/bash

# =============================================================================
# PetMiner 部署检查脚本 - 验证部署环境和连接性
# =============================================================================

set -e

# 配置
REMOTE_HOST="154.29.153.211"
REMOTE_PORT="36602"
REMOTE_USER="root"
REMOTE_APP_DIR="/app"

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# 检查结果
CHECKS_PASSED=0
CHECKS_TOTAL=0

# 日志函数
log() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

check_pass() {
    echo -e "${GREEN}[✓]${NC} $1"
    CHECKS_PASSED=$((CHECKS_PASSED + 1))
}

check_fail() {
    echo -e "${RED}[✗]${NC} $1"
}

check_warn() {
    echo -e "${YELLOW}[!]${NC} $1"
}

run_check() {
    CHECKS_TOTAL=$((CHECKS_TOTAL + 1))
    echo -n "检查: $1 ... "
}

# 检查本地环境
check_local_environment() {
    echo ""
    log "=== 本地环境检查 ==="
    
    # 检查必要工具
    for tool in ssh rsync node npm go; do
        run_check "本地工具 $tool"
        if command -v "$tool" >/dev/null 2>&1; then
            check_pass "$tool 已安装"
        else
            check_fail "$tool 未安装"
        fi
    done
    
    # 检查项目文件
    run_check "项目结构"
    if [ -f "frontend/package.json" ] && [ -f "backend/go.mod" ]; then
        check_pass "项目文件完整"
    else
        check_fail "项目文件不完整"
    fi
}

# 检查SSH连接
check_ssh_connection() {
    echo ""
    log "=== SSH连接检查 ==="
    
    run_check "SSH连接"
    if ssh -p "$REMOTE_PORT" -o ConnectTimeout=10 -o BatchMode=yes "$REMOTE_USER@$REMOTE_HOST" exit 2>/dev/null; then
        check_pass "SSH免密连接成功"
    else
        check_fail "SSH连接失败，请检查网络和密钥配置"
        return 1
    fi
    
    run_check "服务器响应时间"
    local start_time=$(date +%s%N)
    ssh -p "$REMOTE_PORT" "$REMOTE_USER@$REMOTE_HOST" "echo test" >/dev/null 2>&1
    local end_time=$(date +%s%N)
    local duration=$(( (end_time - start_time) / 1000000 ))
    
    if [ $duration -lt 1000 ]; then
        check_pass "响应时间: ${duration}ms (优秀)"
    elif [ $duration -lt 3000 ]; then
        check_warn "响应时间: ${duration}ms (一般)"
        CHECKS_PASSED=$((CHECKS_PASSED + 1))
    else
        check_fail "响应时间: ${duration}ms (较慢)"
    fi
}

# 检查远程环境
check_remote_environment() {
    echo ""
    log "=== 远程环境检查 ==="
    
    # 获取系统信息
    local system_info=$(ssh -p "$REMOTE_PORT" "$REMOTE_USER@$REMOTE_HOST" "uname -a" 2>/dev/null || echo "获取失败")
    echo "系统信息: $system_info"
    
    # 检查Go环境
    run_check "Go环境"
    local go_version=$(ssh -p "$REMOTE_PORT" "$REMOTE_USER@$REMOTE_HOST" "go version 2>/dev/null || echo 'not found'")
    if [[ $go_version == *"go version"* ]]; then
        check_pass "Go已安装: $go_version"
    else
        check_fail "Go未安装或不在PATH中"
    fi
    
    # 检查Node.js环境
    run_check "Node.js环境"
    local node_version=$(ssh -p "$REMOTE_PORT" "$REMOTE_USER@$REMOTE_HOST" "node --version 2>/dev/null || echo 'not found'")
    if [[ $node_version == v* ]]; then
        check_pass "Node.js已安装: $node_version"
    else
        check_fail "Node.js未安装或不在PATH中"
    fi
    
    # 检查必要工具
    for tool in systemctl nginx curl netstat; do
        run_check "工具 $tool"
        if ssh -p "$REMOTE_PORT" "$REMOTE_USER@$REMOTE_HOST" "command -v $tool >/dev/null 2>&1"; then
            check_pass "$tool 可用"
        else
            check_warn "$tool 不可用"
            CHECKS_PASSED=$((CHECKS_PASSED + 1))
        fi
    done
    
    # 检查端口占用
    run_check "端口8081占用"
    if ssh -p "$REMOTE_PORT" "$REMOTE_USER@$REMOTE_HOST" "netstat -tlnp 2>/dev/null | grep -q ':8081 '"; then
        check_warn "端口8081已被占用"
        CHECKS_PASSED=$((CHECKS_PASSED + 1))
    else
        check_pass "端口8081可用"
    fi
    
    run_check "端口80占用"
    if ssh -p "$REMOTE_PORT" "$REMOTE_USER@$REMOTE_HOST" "netstat -tlnp 2>/dev/null | grep -q ':80 '"; then
        check_warn "端口80已被占用"
        CHECKS_PASSED=$((CHECKS_PASSED + 1))
    else
        check_pass "端口80可用"
    fi
}

# 检查系统资源
check_system_resources() {
    echo ""
    log "=== 系统资源检查 ==="
    
    # 检查内存
    run_check "可用内存"
    local memory_info=$(ssh -p "$REMOTE_PORT" "$REMOTE_USER@$REMOTE_HOST" "free -h | grep Mem" 2>/dev/null || echo "")
    if [ -n "$memory_info" ]; then
        local available=$(echo "$memory_info" | awk '{print $7}' | sed 's/[^0-9.]//g')
        local total=$(echo "$memory_info" | awk '{print $2}' | sed 's/[^0-9.]//g')
        echo "内存信息: $memory_info"
        if [ -n "$available" ] && (( $(echo "$available > 0.5" | bc -l 2>/dev/null || echo 0) )); then
            check_pass "内存充足"
        else
            check_warn "内存可能不足"
            CHECKS_PASSED=$((CHECKS_PASSED + 1))
        fi
    else
        check_warn "无法获取内存信息"
        CHECKS_PASSED=$((CHECKS_PASSED + 1))
    fi
    
    # 检查磁盘空间
    run_check "磁盘空间"
    local disk_info=$(ssh -p "$REMOTE_PORT" "$REMOTE_USER@$REMOTE_HOST" "df -h / | tail -1" 2>/dev/null || echo "")
    if [ -n "$disk_info" ]; then
        echo "磁盘信息: $disk_info"
        local usage=$(echo "$disk_info" | awk '{print $5}' | sed 's/%//')
        if [ -n "$usage" ] && [ "$usage" -lt 80 ]; then
            check_pass "磁盘空间充足"
        else
            check_warn "磁盘使用率较高: ${usage}%"
            CHECKS_PASSED=$((CHECKS_PASSED + 1))
        fi
    else
        check_warn "无法获取磁盘信息"
        CHECKS_PASSED=$((CHECKS_PASSED + 1))
    fi
}

# 检查现有部署
check_existing_deployment() {
    echo ""
    log "=== 现有部署检查 ==="
    
    # 检查应用目录
    run_check "应用目录 $REMOTE_APP_DIR"
    if ssh -p "$REMOTE_PORT" "$REMOTE_USER@$REMOTE_HOST" "[ -d '$REMOTE_APP_DIR' ]"; then
        check_warn "应用目录已存在（将被覆盖）"
        CHECKS_PASSED=$((CHECKS_PASSED + 1))
        
        # 显示现有文件
        local existing_files=$(ssh -p "$REMOTE_PORT" "$REMOTE_USER@$REMOTE_HOST" "ls -la '$REMOTE_APP_DIR' 2>/dev/null || echo '空目录'")
        echo "现有文件:"
        echo "$existing_files" | head -5
    else
        check_pass "应用目录不存在（全新部署）"
    fi
    
    # 检查服务状态
    run_check "PetMiner服务"
    if ssh -p "$REMOTE_PORT" "$REMOTE_USER@$REMOTE_HOST" "systemctl is-active petminer >/dev/null 2>&1"; then
        check_warn "PetMiner服务正在运行（将被重启）"
        CHECKS_PASSED=$((CHECKS_PASSED + 1))
    else
        check_pass "PetMiner服务未运行"
    fi
    
    # 检查nginx配置
    run_check "Nginx配置"
    if ssh -p "$REMOTE_PORT" "$REMOTE_USER@$REMOTE_HOST" "[ -f /etc/nginx/nginx.conf ]"; then
        check_warn "Nginx配置存在（将被备份和替换）"
        CHECKS_PASSED=$((CHECKS_PASSED + 1))
    else
        check_pass "Nginx配置不存在（全新配置）"
    fi
}

# 网络连通性测试
check_network_connectivity() {
    echo ""
    log "=== 网络连通性检查 ==="
    
    # 检查远程服务器网络
    run_check "外网连接"
    if ssh -p "$REMOTE_PORT" "$REMOTE_USER@$REMOTE_HOST" "curl -s --connect-timeout 5 google.com >/dev/null 2>&1"; then
        check_pass "外网连接正常"
    else
        check_warn "外网连接可能有问题"
        CHECKS_PASSED=$((CHECKS_PASSED + 1))
    fi
    
    # 检查DNS解析
    run_check "DNS解析"
    if ssh -p "$REMOTE_PORT" "$REMOTE_USER@$REMOTE_HOST" "nslookup google.com >/dev/null 2>&1"; then
        check_pass "DNS解析正常"
    else
        check_warn "DNS解析可能有问题"
        CHECKS_PASSED=$((CHECKS_PASSED + 1))
    fi
}

# 显示建议
show_recommendations() {
    echo ""
    log "=== 部署建议 ==="
    
    if [ $CHECKS_PASSED -eq $CHECKS_TOTAL ]; then
        echo -e "${GREEN}🎉 所有检查通过！可以安全地进行部署。${NC}"
    elif [ $CHECKS_PASSED -ge $((CHECKS_TOTAL * 80 / 100)) ]; then
        echo -e "${YELLOW}⚠️  大部分检查通过，可以尝试部署，但请注意警告项。${NC}"
    else
        echo -e "${RED}❌ 多项检查失败，建议先解决问题再部署。${NC}"
    fi
    
    echo ""
    echo "检查结果: $CHECKS_PASSED/$CHECKS_TOTAL 通过"
    echo ""
    
    # 给出具体建议
    if command -v ssh >/dev/null 2>&1; then
        echo "💡 部署命令:"
        echo "   ./deploy.sh                    # 完整部署"
        echo "   ./deploy.sh build              # 仅构建"
        echo "   ./manage-remote.sh status      # 检查服务状态"
    fi
    
    echo ""
    echo "💡 故障排除:"
    echo "   - SSH连接问题: 检查网络、IP地址、SSH密钥"
    echo "   - 环境缺失: 在远程服务器安装Go、Node.js"
    echo "   - 端口占用: 停止占用端口的服务"
    echo "   - 权限问题: 确保使用root用户或配置sudo"
}

# 主函数
main() {
    echo "PetMiner 部署环境检查"
    echo "======================="
    echo "远程服务器: $REMOTE_HOST"
    echo "部署目录: $REMOTE_APP_DIR"
    echo ""
    
    # 执行所有检查
    check_local_environment
    
    if check_ssh_connection; then
        check_remote_environment
        check_system_resources
        check_existing_deployment
        check_network_connectivity
    else
        echo ""
        echo -e "${RED}❌ SSH连接失败，跳过远程检查${NC}"
    fi
    
    # 显示结果和建议
    show_recommendations
}

# 执行检查
main "$@"