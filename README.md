# 次元链 (NexusACG)

> ACG 线下产业生态服务平台 — AI + 电商 + 社交 + LBS

## 项目概述

次元链是一个面向二次元/Cosplay 爱好者的一站式平台，整合了：
- **智能电商** — Cosplay 服饰 + 周边商品双区展示、搜索、下单
- **AI 虚拟试穿** — 上传照片即可预览 Cos 效果（Phase 2）
- **社区互动** — 图文/视频发布、点赞评论、兴趣圈层
- **漫展活动** — LBS 定位、活动地图、妆娘/摄影师预约

采用 OPC（One Person Company）一人公司模式，借助 AI 工具链快速开发迭代。

## 技术栈

| 层 | 技术 |
|---|---|
| 客户端 | Flutter 3.x + Dart |
| 后端 API | Go 1.21+ + Gin + GORM |
| 数据库 | PostgreSQL 15+ + Redis 7 |
| 管理后台 | 原生 HTML/JS (MVP) |
| 部署 | Docker + Docker Compose |

## 快速开始

### 环境要求
- Docker & Docker Compose
- Go 1.21+（本地开发）
- Flutter 3.2+（客户端开发）

### 启动后端

```bash
# 1. 复制环境变量
cp .env.example .env

# 2. 启动数据库 + Redis + 后端
docker compose up -d

# 3. 检查后端是否运行
curl http://localhost:8080/health
# 返回: {"status":"ok","service":"nexusacg"}
```

### 本地开发（非 Docker）

```bash
# 后端
cd backend
go mod download
go run cmd/server/main.go

# 客户端
cd client
flutter pub get
flutter run
```

### 管理后台

直接在浏览器打开 `admin/public/index.html`，确保后端 API 运行在 `http://localhost:8080`。

## 项目结构

```
nexusacg/
├── backend/                    # Go 后端 API
│   ├── cmd/server/             # 入口文件
│   ├── internal/
│   │   ├── config/             # 配置管理
│   │   ├── database/           # 数据库连接
│   │   ├── handler/            # HTTP 处理器
│   │   ├── middleware/         # 中间件 (JWT, CORS)
│   │   ├── model/              # 数据模型
│   │   └── service/            # 业务逻辑
│   └── Dockerfile
├── client/                     # Flutter 客户端
│   ├── lib/
│   │   ├── core/               # 通用层 (models, network, theme)
│   │   ├── presentation/       # UI 层 (screens, blocs, widgets)
│   │   └── main.dart
│   └── pubspec.yaml
├── admin/public/               # 管理后台 (HTML)
├── docker/
│   └── postgres/init.sql       # 数据库初始化脚本
├── docker-compose.yml
└── .env.example
```

## API 端点

### 认证
- `POST /api/v1/auth/register` — 用户注册
- `POST /api/v1/auth/login` — 用户登录
- `POST /api/v1/auth/refresh` — 刷新 Token

### 商品
- `GET /api/v1/products` — 商品列表（支持 zone/keyword 筛选）
- `GET /api/v1/products/:id` — 商品详情
- `POST /api/v1/products` — 创建商品（需登录）

### 社区
- `GET /api/v1/posts` — 帖子列表
- `GET /api/v1/posts/:id` — 帖子详情
- `POST /api/v1/posts` — 发布帖子（需登录）
- `POST /api/v1/posts/:id/like` — 点赞
- `DELETE /api/v1/posts/:id/like` — 取消点赞

### 活动
- `GET /api/v1/events` — 活动列表
- `GET /api/v1/events/:id` — 活动详情
- `POST /api/v1/events` — 创建活动（需登录）

### 订单
- `POST /api/v1/orders` — 创建订单（需登录，支持幂等）
- `GET /api/v1/orders` — 我的订单
- `POST /api/v1/orders/:id/pay` — 支付订单

## MVP 开发路线图

### Phase 1: MVP (当前)
- [x] 用户注册/登录（手机号/邮箱 + 密码）
- [x] 商品双区展示与搜索
- [x] 基础社区功能（发布、点赞）
- [x] 漫展列表
- [x] 管理后台
- [ ] 微信支付/支付宝接入
- [ ] 担保交易/分账系统

### Phase 2: 增长
- [ ] AI 虚拟试穿基础版
- [ ] 角色推荐清单
- [ ] 服务者入驻与预约
- [ ] LBS 活动地图
- [ ] 兴趣圈层

### Phase 3: 成熟
- [ ] 3D 商品预览
- [ ] 数据分析平台
- [ ] 广告系统
- [ ] 票务系统
- [ ] 国际化

## OPC 技能栈

本项目已安装以下 Claude Code 自定义技能：

| 技能 | 用途 |
|---|---|
| `flutter-expert` | Flutter/Dart 开发 |
| `golang-pro` | Go 后端开发 |
| `python-expert` | Python/AI 脚本 |
| `frontend-design` | 前端设计 |
| `feature-dev` | 功能开发流程 |
| `pr-review-toolkit` | 代码审查 |
| `brainstorming` | 需求分析 |
| `code-reviewer` | 代码质量检查 |
| `debugging` | Bug 排查 |
| `owasp-security` | Web 安全基线 |
| `static-analysis` | 静态分析 |
| `supply-chain-risk-auditor` | 依赖安全审计 |
| `insecure-defaults` | 不安全配置检查 |
| `llm-security` | AI 功能安全 |
| `agentic-ai-security` | Agent 系统安全 |

## 合规提醒

- 涉及平台代收代付，必须办理 **ICP 许可证**、**EDI 许可证**
- 支付环节严格走微信/支付宝**官方分账**，不可自建资金池
- 用户数据本地化存储，不传用户数据出境（《数据安全法》）
- 社区内容需要 AI 审核 + 人工复审双重保障

## 许可证

Copyright 2026 永生计划（Plan:Forever）. All rights reserved.
