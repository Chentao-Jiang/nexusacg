---
name: nexusacg
description: NexusACG（次元链）项目开发专用技能。Go+Gin 后端 + Flutter 客户端，已部署阿里云服务器。包含项目上下文、部署流程、任务追踪和修复记录规范。
---

# NexusACG 项目开发指南

## 项目概述

次元链（NexusACG）—— ACG 商品交易 + 社区平台

**技术栈**: Go 1.25 + Gin + GORM + PostgreSQL 16 + Redis 7（后端），Flutter + Dart + BLoC（客户端）

**项目根目录**: `/home/jct/nexusacg/`
- `backend/` — Go 后端
- `client/` — Flutter 客户端

## 已部署服务器

- **公网 IP**: 101.133.169.72
- **SSH**: root@101.133.169.72
- **API 地址**: `http://101.133.169.72:8080/api/v1`
- **Flutter 配置**: `client/lib/core/constants/app_constants.dart`

**服务器限制**: 2GB 内存，**不可在服务器上构建**，必须本地交叉编译后上传。

## 必读文件

每次任务开始时，先读取这两个文件了解当前状态：

1. **`TASKS.md`** (`/home/jct/nexusacg/TASKS.md`) — 项目任务清单
   - Phase 1 (MVP): 核心功能实现状态
   - Phase 2 (增长): AI 试穿、服务者平台等
   - Phase 3 (成熟): 3D 预览、国际化等
   - "新增任务" 章节包含用户反馈需求

2. **`DEPLOYED_SERVER.md`** (`/home/jct/nexusacg/DEPLOYED_SERVER.md`) — 部署日志 + 修复记录
   - 服务器信息、部署方式、环境变量摘要
   - **修复记录表**: 记录所有已修复的问题，**避免重复试错**
   - 每次修复问题后必须追加记录

## 后端部署流程（已验证）

```bash
# 1. 本地交叉编译
cd /home/jct/nexusacg/backend
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o /tmp/server_linux ./cmd/server

# 2. 上传二进制
scp /tmp/server_linux root@101.133.169.72:/tmp/server_linux

# 3. 替换容器内二进制
ssh root@101.133.169.72 "docker stop nexusacg-api && docker cp /tmp/server_linux nexusacg-api:/app/server && docker start nexusacg-api"

# 4. 验证
ssh root@101.133.169.72 "curl -s http://localhost:8080/health && docker logs --tail 5 nexusacg-api"
```

**关键注意事项**:
- 必须用 `docker stop` + `docker cp` + `docker start`，不能用 `docker compose up --force-recreate`（会从镜像重建，覆盖上传的二进制）
- `.env` 文件在 `/opt/nexusacg/.env`，容器创建时加载。如需修改环境变量且不能重建容器，需重建时重新加载 `.env`

## Flutter APK 构建

```bash
cd /home/jct/nexusacg/client
export PATH="/home/jct/flutter/bin:$PATH"
export JAVA_HOME="/home/jct/jdks/jdk-17.0.12"
export ANDROID_HOME="/home/jct/Android"
flutter build apk --release
cp build/app/outputs/flutter-apk/app-release.apk /home/jct/nexusacg/次元链-v0.1.0-release.apk
```

**已知问题**:
- `AndroidManifest.xml` 必须有 `<uses-permission android:name="android.permission.INTERNET"/>`，否则连接被拒绝
- `compileSdk` 需设为 35（flutter_plugin_android_lifecycle 要求）
- APK 包名: `com.nexusacg.app`

## 认证系统现状

### 邮箱注册
- 发送验证邮件（SMTP: QQ 邮箱）
- 验证链接使用 deep link `nexusacg://verify?token=xxx`（避免 QQ 邮箱拦截 HTTP URL）
- 邮箱需验证后状态为 `active`

### 手机号注册
- 阿里云短信认证 API（SendSmsVerifyCode），已配置 AccessKey
- 注册表单: 昵称 + 手机号 + 验证码 + 密码 + 确认密码
- 注册成功直接登录并返回 token
- 登录时手机号用户也需要密码（密码 hash 在注册时保存）

### 登录
- 支持手机号/邮箱 + 密码登录（自动识别 `@`）
- SMS 注册用户必须有密码才能登录

## 修复记录规范

每次修复问题后，**必须**在 `DEPLOYED_SERVER.md` 的修复记录表中追加一行：

```
| 日期 | 问题描述 | 修复方案（文件 + 改动） |
```

这样下次遇到相同问题时可以先查阅记录，避免重复排查。

## 数据库

- PostgreSQL 16 在 Docker 中
- DB 密码: `nexusacg_dev_pass`
- 用户: `nexusacg`
- 数据库: `nexusacg`

测试数据清理命令：
```bash
ssh root@101.133.169.72 "PGPASSWORD=nexusacg_dev_pass docker exec -e PGPASSWORD=nexusacg_dev_pass nexusacg-db psql -U nexusacg -d nexusacg -c 'TRUNCATE users CASCADE; TRUNCATE refresh_tokens; TRUNCATE email_verification_tokens;'"
```

## 开发规范

1. **每次任务开始**: 先读 `TASKS.md` 和 `DEPLOYED_SERVER.md` 了解当前状态
2. **修复问题后**: 立即更新 `DEPLOYED_SERVER.md` 修复记录
3. **后端部署**: 本地交叉编译 + scp + docker cp，不要在服务器上构建
4. **Flutter 修改**: 修改后重新构建 APK，提醒用户卸载旧版再安装
5. **数据库变更**: 涉及模型变更时需执行 AutoMigrate，必要时手动迁移
6. **环境变量**: 修改后需重启容器，新建容器时需确保 `.env` 包含最新变量
