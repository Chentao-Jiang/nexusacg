#!/usr/bin/env bash
# NexusACG ECS Deployment Script
# Packages the backend and deploys to Alibaba Cloud ECS
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
DEPLOY_DIR="$SCRIPT_DIR"
REMOTE_USER="${REMOTE_USER:-root}"
REMOTE_HOST="${REMOTE_HOST:-}"
REMOTE_DIR="${REMOTE_DIR:-/opt/nexusacg}"

##############################################################################
# Colors
##############################################################################
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

info()  { echo -e "${GREEN}[INFO]${NC} $*"; }
warn()  { echo -e "${YELLOW}[WARN]${NC} $*"; }
error() { echo -e "${RED}[ERROR]${NC} $*"; }

##############################################################################
# Usage
##############################################################################
usage() {
    cat <<EOF
Usage: $0 <command> [options]

Commands:
  prepare       Create deploy directory with all production files
  push          Upload deploy package to ECS server
  deploy        Full pipeline: prepare + push + install
  start         Start services on ECS
  stop          Stop services on ECS
  status        Check service status on ECS
  logs          Tail API logs on ECS

Options:
  -h, --host    Remote ECS IP address (required for push/deploy)
  -u, --user    Remote SSH user (default: root)
  -d, --dir     Remote deploy directory (default: /opt/nexusacg)
  -c, --clean   Clean existing deployment before push

Examples:
  $0 prepare
  $0 push -h 47.xxx.xxx.xxx
  $0 deploy -h 47.xxx.xxx.xxx -u root
EOF
    exit 1
}

##############################################################################
# Prepare: Create deploy directory
##############################################################################
prepare() {
    info "Preparing deployment package..."

    local deploy_tmp="$SCRIPT_DIR/../deploy-package"
    rm -rf "$deploy_tmp"
    mkdir -p "$deploy_tmp"/{backend,keys,nginx/{certs,html},uploads}

    # Copy backend source
    cp -r "$PROJECT_ROOT/backend"/* "$deploy_tmp/backend/"

    # Copy deploy configs
    cp "$DEPLOY_DIR/docker-compose.prod.yml" "$deploy_tmp/docker-compose.yml"
    cp "$DEPLOY_DIR/nginx/nginx.conf" "$deploy_tmp/nginx/"
    cp -r "$DEPLOY_DIR/nginx/certs" "$deploy_tmp/nginx/"
    cp -r "$DEPLOY_DIR/nginx/html" "$deploy_tmp/nginx/"

    # Copy .env template
    cp "$DEPLOY_DIR/.env.production.template" "$deploy_tmp/.env.production"

    # Copy payment keys if they exist
    if [ -f "$PROJECT_ROOT/alipay_app_private_key.pem" ]; then
        cp "$PROJECT_ROOT/alipay_app_private_key.pem" "$deploy_tmp/keys/"
    fi
    if [ -f "$PROJECT_ROOT/alipay_public_key.pem" ]; then
        cp "$PROJECT_ROOT/alipay_public_key.pem" "$deploy_tmp/keys/"
    fi

    # Copy ECS deployment guide
    cat > "$deploy_tmp/DEPLOY.md" <<'GUIDE'
# NexusACG ECS 部署指南

## 前置要求

1. 阿里云 ECS 实例（建议：2核4G，Ubuntu 22.04）
2. 安全组开放端口：80, 443, 22
3. Docker + Docker Compose 已安装
4. 域名已解析到 ECS IP（用于 HTTPS）

## 快速部署

```bash
# 1. 配置环境变量
cp .env.production.template .env.production
vim .env.production  # 填写所有必填值

# 2. 上传密钥文件到 keys/ 目录
# - alipay_app_private_key.pem
# - alipay_public_key.pem

# 3. 启动服务
docker compose up -d --build

# 4. 检查日志
docker compose logs -f api
```

## 获取 HTTPS 证书（可选）

```bash
# 使用 certbot
apt-get install certbot -y
certbot certonly --standalone -d your-domain.com
cp /etc/letsencrypt/live/your-domain.com/fullchain.pem nginx/certs/
cp /etc/letsencrypt/live/your-domain.com/privkey.pem nginx/certs/

# 编辑 nginx/nginx.conf 取消注释 HTTPS server 块
# 重启 nginx
docker compose up -d nginx
```

## 更新 APK 的 API 地址

修改 Flutter 项目：
`lib/core/constants/app_constants.dart`
```dart
static const String apiBaseUrl = 'https://your-domain.com/api/v1';
```
然后重新构建 APK：
```bash
cd client
flutter build apk --release
```
GUIDE

    info "Deployment package ready at: $deploy_tmp"
    info "Next steps:"
    echo "  1. Copy .env.production.template to .env.production and fill in values"
    echo "  2. Upload payment keys to keys/ directory"
    echo "  3. Run: $0 push -h <ECS_IP>"
}

##############################################################################
# Push: Upload to ECS
##############################################################################
push() {
    if [ -z "$REMOTE_HOST" ]; then
        error "Remote host (-h) is required for push"
        exit 1
    fi

    local deploy_tmp="$SCRIPT_DIR/../deploy-package"
    if [ ! -d "$deploy_tmp" ]; then
        error "Deploy package not found. Run '$0 prepare' first."
        exit 1
    fi

    info "Uploading to $REMOTE_USER@$REMOTE_HOST:$REMOTE_DIR ..."

    # Create remote directory
    ssh "$REMOTE_USER@$REMOTE_HOST" "mkdir -p $REMOTE_DIR"

    # Sync files
    rsync -avz --delete \
        "$deploy_tmp/" \
        "$REMOTE_USER@$REMOTE_HOST:$REMOTE_DIR/"

    info "Upload complete!"
}

##############################################################################
# Remote commands
##############################################################################
remote_exec() {
    ssh "$REMOTE_USER@$REMOTE_HOST" "cd $REMOTE_DIR && $*"
}

start_services() {
    if [ -z "$REMOTE_HOST" ]; then error "Remote host (-h) required"; exit 1; fi
    info "Starting services on $REMOTE_HOST..."
    remote_exec "docker compose up -d --build"
    info "Services started. Checking health..."
    sleep 5
    remote_exec "docker compose ps"
}

stop_services() {
    if [ -z "$REMOTE_HOST" ]; then error "Remote host (-h) required"; exit 1; fi
    info "Stopping services on $REMOTE_HOST..."
    remote_exec "docker compose down"
}

check_status() {
    if [ -z "$REMOTE_HOST" ]; then error "Remote host (-h) required"; exit 1; fi
    info "Service status on $REMOTE_HOST..."
    remote_exec "docker compose ps"
    info "API health check..."
    remote_exec "curl -s http://localhost:8080/health || echo 'API not responding'"
}

tail_logs() {
    if [ -z "$REMOTE_HOST" ]; then error "Remote host (-h) required"; exit 1; fi
    info "Tailing API logs on $REMOTE_HOST..."
    ssh -t "$REMOTE_USER@$REMOTE_HOST" "cd $REMOTE_DIR && docker compose logs -f api"
}

##############################################################################
# Main
##############################################################################
case "${1:-}" in
    prepare)
        prepare
        ;;
    push)
        shift
        while [[ $# -gt 0 ]]; do
            case "$1" in
                -h|--host) REMOTE_HOST="$2"; shift 2 ;;
                -u|--user) REMOTE_USER="$2"; shift 2 ;;
                -d|--dir)  REMOTE_DIR="$2"; shift 2 ;;
                *) shift ;;
            esac
        done
        push
        ;;
    deploy)
        shift
        while [[ $# -gt 0 ]]; do
            case "$1" in
                -h|--host) REMOTE_HOST="$2"; shift 2 ;;
                -u|--user) REMOTE_USER="$2"; shift 2 ;;
                -d|--dir)  REMOTE_DIR="$2"; shift 2 ;;
                *) shift ;;
            esac
        done
        if [ -z "$REMOTE_HOST" ]; then error "Remote host (-h) required"; exit 1; fi
        prepare
        push
        start_services
        ;;
    start)
        shift
        while [[ $# -gt 0 ]]; do
            case "$1" in
                -h|--host) REMOTE_HOST="$2"; shift 2 ;;
                -u|--user) REMOTE_USER="$2"; shift 2 ;;
                *) shift ;;
            esac
        done
        start_services
        ;;
    stop)
        shift
        while [[ $# -gt 0 ]]; do
            case "$1" in
                -h|--host) REMOTE_HOST="$2"; shift 2 ;;
                -u|--user) REMOTE_USER="$2"; shift 2 ;;
                *) shift ;;
            esac
        done
        stop_services
        ;;
    status)
        shift
        while [[ $# -gt 0 ]]; do
            case "$1" in
                -h|--host) REMOTE_HOST="$2"; shift 2 ;;
                -u|--user) REMOTE_USER="$2"; shift 2 ;;
                *) shift ;;
            esac
        done
        check_status
        ;;
    logs)
        shift
        while [[ $# -gt 0 ]]; do
            case "$1" in
                -h|--host) REMOTE_HOST="$2"; shift 2 ;;
                -u|--user) REMOTE_USER="$2"; shift 2 ;;
                *) shift ;;
            esac
        done
        tail_logs
        ;;
    *)
        usage
        ;;
esac
