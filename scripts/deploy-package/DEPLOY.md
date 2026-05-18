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
