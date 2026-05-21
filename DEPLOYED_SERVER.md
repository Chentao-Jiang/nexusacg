# NexusACG 已部署服务器信息

> 服务器：阿里云轻量应用服务器
> 公网 IP: 101.133.169.72
> 内网 IP: 172.24.21.161
> 配置: 2GB 内存，40GB 磁盘，Ubuntu
> SSH: root@101.133.169.72

## 部署路径
- 项目目录: `/opt/nexusacg`
- Docker Compose: `/opt/nexusacg/docker-compose.yml`
- 环境变量: `/opt/nexusacg/.env`
- 支付密钥: `/opt/nexusacg/alipay_app_private_key.pem`, `alipay_public_key.pem`
- 上传文件: `/opt/nexusacg/uploads/`
- 数据库数据: Docker volume `postgres_data`

## 当前运行容器
| 容器 | 镜像 | 端口 | 状态 |
|------|------|------|------|
| nexusacg-api | nexusacg-api:latest | 8080:8080 | healthy |
| nexusacg-db | postgres:16-alpine | 5432:5432 | healthy |
| nexusacg-redis | redis:7-alpine | 6379:6379 | healthy |

## 服务器限制
- 内存仅 2GB，**不可在服务器上执行 go build / docker build**
- 必须本地交叉编译为 Linux 二进制后上传
- **更新方式**: 本地编译 → scp 上传 → `docker cp` 替换容器内二进制 → `docker restart`

## 部署方式（二进制流 — 已验证）
```bash
# 1. 本地交叉编译
cd /home/jct/nexusacg/backend
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o /tmp/server_linux ./cmd/server

# 2. 上传到服务器
scp /tmp/server_linux root@101.133.169.72:/tmp/server_linux

# 3. 替换容器内二进制（关键：不能用 docker compose up --force-recreate，那会从镜像重建）
ssh root@101.133.169.72 "docker stop nexusacg-api && docker cp /tmp/server_linux nexusacg-api:/app/server && docker start nexusacg-api"

# 4. 验证
ssh root@101.133.169.72 "curl -s http://localhost:8080/health && docker logs --tail 5 nexusacg-api"
```

## 环境变量摘要（脱敏）
- ENV=development
- BASE_URL=http://localhost:8080
- ALIPAY_APP_ID=2021006153686187（沙箱）
- ALIPAY_SANDBOX=true
- PLATFORM_FEE_PERCENT=0.05
- AUTO_RELEASE_DAYS=7
- ORDER_TIMEOUT_MINUTES=30
- SMTP: QQ 邮箱配置
- AI 审核: DeepSeek + Qwen API Key

## API 地址
- 服务器端: http://101.133.169.72:8080
- Flutter 端: `http://101.133.169.72:8080/api/v1`
- Flutter 配置: `lib/core/constants/app_constants.dart`

## 已完成部署（2026-05-18 更新）

### 新增模型（AutoMigrate 已执行）
- `Dispute` — 纠纷
- `DisputeMessage` — 纠纷沟通消息
- `RefundApplication` — 退换货申请
- `Order.DisputeStatus` — 订单纠纷状态字段

### 新增 API 端点

**纠纷处理 (Dispute)**
- `POST /api/v1/disputes` — 买家发起纠纷
- `GET /api/v1/disputes/my` — 我的纠纷列表
- `GET /api/v1/disputes/order/:order_no` — 查看订单纠纷
- `POST /api/v1/disputes/:id/respond` — 卖家回应
- `POST /api/v1/disputes/:id/messages` — 发送消息
- `GET /api/v1/disputes/:id/messages` — 查看消息
- `GET /api/v1/admin/disputes` — 管理员查看所有纠纷
- `POST /api/v1/admin/disputes/:id/resolve` — 管理员仲裁

**退换货申请 (RefundApplication)**
- `POST /api/v1/refund-applications` — 买家提交申请（支持部分退款+凭证）
- `GET /api/v1/refund-applications/my` — 买家查看我的申请
- `POST /api/v1/refund-applications/:id/review` — 卖家审核
- `GET /api/v1/seller/refund-applications` — 卖家查看所有申请
- `GET /api/v1/admin/refund-applications` — 管理员查看所有申请
- `POST /api/v1/admin/refund-applications/:id/execute` — 管理员执行退款

**商家认证 (Certification)**
- `POST /api/v1/certifications/merchant` — 商家认证申请
- `POST /api/v1/certifications/service-provider` — 服务商认证申请
- `GET /api/v1/certifications/my` — 查看我的认证申请
- `GET /api/v1/admin/certifications` — 管理员查看所有认证
- `POST /api/v1/admin/certifications/:id/review` — 管理员审核

**服务商品 (ServiceProduct)**
- `GET /api/v1/service-products` — 服务商品列表
- `GET /api/v1/service-products/:id` — 服务商品详情
- `GET /api/v1/service-products/:id/schedules` — 排期查询
- `POST /api/v1/service-products` — 创建服务商品
- `PUT /api/v1/service-products/:id` — 更新
- `DELETE /api/v1/service-products/:id` — 删除
- `GET /api/v1/service-products/my` — 我的服务商品

**推广申请 (PromotionApplication)**
- `POST /api/v1/promotions` — 提交推广申请
- `GET /api/v1/promotions/my` — 我的推广
- `GET /api/v1/admin/promotions` — 管理员查看所有推广
- `POST /api/v1/admin/promotions/:id/review` — 管理员审核

### 修复记录
| 日期 | 问题 | 修复 |
|------|------|------|
| 2026-05-18 | APK 注册失败: `_dio` 未初始化 | `main.dart` 添加 `await ApiClient().init()` |
| 2026-05-18 | APK 注册失败: Railway 域名无法解析 | `app_constants.dart` 改为服务器 IP |
| 2026-05-18 | APK 注册失败: 局域网 IP 手机无法访问 | `app_constants.dart` 改为公网 IP 101.133.169.72 |
| 2026-05-18 | 服务器 DB 密码认证失败 | 容器重建后重置密码 `nexusacg_dev_pass` |
| 2026-05-18 | `docker compose up --force-recreate` 替换二进制失败 | 改用 `docker stop` + `docker cp` + `docker start` |
| 2026-05-18 | APK 注册失败: SocketException Operation not permitted | `AndroidManifest.xml` 添加 `INTERNET` 权限 |
| 2026-05-19 | 手机号注册: 无验证码输入、手机号选填 | `register_screen.dart` 添加 SMS 验证码流程 + 必填验证 |
| 2026-05-19 | 邮箱注册 400 报错不友好 | `api_client.dart` 添加 `validateStatus: true` 解析错误消息 |
| 2026-05-19 | 登录页仅支持手机号 | `login_screen.dart` 改为"手机号/邮箱"自动识别 |
| 2026-05-19 | 短信服务未配置 | `.env` 添加阿里云 AccessKey，启用短信认证 API |
| 2026-05-19 | 邮箱验证链接 localhost 被QQ邮箱拦截 | `BASE_URL` 改为公网 IP 101.133.169.72:8080 |
| 2026-05-19 | 手机号注册"获取验证码"按钮无反应 | `register_screen.dart` 跳过验证码表单验证直接发送 |
| 2026-05-19 | `/verify` 页面 404 | 容器二进制为旧版本，重新编译 Go + scp + docker cp 替换 |
| 2026-05-19 | 邮箱验证链接 HTTP URL 被 QQ 邮箱拦截 | `email.go` 验证链接改为 deep link `nexusacg://verify` |
| 2026-05-19 | 手机号注册成功但无 UI 反馈 | `register_screen.dart` 添加 `AuthAuthenticated` 监听，成功后返回主页 |
| 2026-05-19 | 手机号注册缺少密码 | `register_screen.dart` 添加密码+确认密码，后端 `SMSLogin` 新增 password+nickname 并保存密码 hash |
| 2026-05-19 | SMS 注册后无法登录（密码为空） | `service/auth.go` 恢复 `Login()` 密码验证（要求 PasswordHash 非空且匹配） |
| 2026-05-19 | 数据库残留测试数据 | `TRUNCATE users + refresh_tokens + email_verification_tokens` |
| 2026-05-19 | 保存资料类型错误 `String is not a subtype` | `repositories.dart` updateProfile 添加类型校验，`api_client.dart` uploadImage/Video 添加 Map 校验 |
| 2026-05-19 | 头像/图片上传失败 | `api_client.dart` 上传路径从 `/upload/image` 修正为 `/upload` |
| 2026-05-19 | `/auth/me` 不存在 → 重新登录后用户名变"用户" | 后端新增 `Me` + `UpdateProfile` handler + `GetMe` + `UpdateProfile` service |
| 2026-05-19 | `/auth/profile` 不存在 → 无法修改资料 | 同上，新增 `POST /auth/profile` |
| 2026-05-19 | 邮箱验证 deep link 运行时无反应 | `app.dart` 改为 platformDispatcher 冷启动方案 |
| 2026-05-19 | 社区帖子卡片操作按钮无反应 | `community_screen.dart` 点赞接入 API，评论跳转详情，分享显示提示 |
| 2026-05-19 | 个人中心菜单项点击无反应 | `profile_screen.dart` 未实现功能显示"开发中"提示，订单快捷入口传递状态筛选 |
| 2026-05-19 | 首页搜索/通知无反应 + 活动占位 | `home_screen.dart` 按钮显示提示，活动加载真实数据 + 下拉刷新 |
| 2026-05-19 | 帖子详情分享按钮无反应 | `post_detail_screen.dart` 添加"开发中"提示 |
| 2026-05-19 | 编辑资料保存成功但用户名不更新 | `edit_profile_screen.dart` 保存后刷新 AuthBloc 状态，同时添加头像上传进度反馈 |
| 2026-05-19 | 邮箱未验证用户登录无反应（无重发入口） | `login_screen.dart` 检测"邮箱未验证"错误时显示重发按钮，`AuthRepository` 新增 `resendEmailVerification` |
| 2026-05-19 | 无法查看/管理自己的帖子 | 后端新增 `GET /posts/my` + `DELETE /posts/:id`；前端新增 `MyPostsScreen` + 个人中心菜单项"我的帖子" |
| 2026-05-19 | 社区页面为"朋友圈"风格，需改为小红书风格 | `community_screen.dart` 全面重写为瀑布流网格布局（MasonryGridView），含封面图、视频缩略图、作者信息 |
| 2026-05-19 | v0.1.3 全面测试 + 代码清理 | 公网 API 全面测试通过（注册/登录/me/profile/products/posts/my/upload/orders/events），修复 22 个 Dart analyzer 警告，GitHub 同步，后端编译通过 |
| 2026-05-19 | 社区图片/视频上传后无法显示 | `api_client.dart` 的 `uploadImage`/`uploadVideo` 读取响应层级错误（`data['url']` 改为 `data['data']['url']`），修复后上传返回正确 URL |
| 2026-05-19 | 邮箱验证邮件中的 deep link 无法被点击 | `email.go` 验证链接从 `nexusacg://verify` 改为 HTTP URL `%s/verify?token=%s`（BASE_URL + /verify），所有邮箱客户端均可正常打开 |
| 2026-05-19 | 邮箱注册等待页无"重新发送"按钮 | `email_pending_screen.dart` 添加"重新发送验证邮件"按钮，调用 `AuthRepository.resendEmailVerification` |
| 2026-05-19 | 视频上传失败无提示 | `post_create_screen.dart` 添加 `uploadVideo` 返回 null 时的错误提示 |
| 2026-05-19 | "我要入驻"菜单点击无反应（缺少入驻表单） | 新建 `certification_screen.dart`，实现商家入驻（店铺名称+营业执照上传）和服务者入驻（服务类型下拉框+描述+作品图片最多10张）双模式表单，提交至对应认证 API；`profile_screen.dart` 菜单项接入 |
| 2026-05-20 | 社区上传视频后无显示（转圈后消失） | `api_client.dart` 的 `uploadVideo` 添加 `contentType: DioMediaType.parse('video/mp4')`，后端因缺少 MIME 类型拒绝上传 |
| 2026-05-20 | 邮箱登录无反应 | `login_screen.dart` 的 BlocListener 新增 `AuthAuthenticated` 状态处理（显示登录成功提示）；`auth_bloc.dart` 添加 `result` 类型守卫防止非 Map 响应崩溃 |
| 2026-05-20 | APK 构建失败: Gradle Daemon 卡死 | WSL2 swap 耗尽导致 Daemon 挂起；需重启 WSL2 释放内存后再构建 |
| 2026-05-20 | APK 构建失败: jlink 不存在 | OpenJDK 缺少 jlink，改用 Oracle JDK 17 (`JAVA_HOME=/home/jct/jdks/jdk-17.0.12`) 构建 |
| 2026-05-21 | SSH 连接全挂（端口通但认证阶段卡死） | 默认 kex 算法 `sntrup761x25519-sha512` 在服务端 ECDH 响应阶段挂死；SSH 命令添加 `-o KexAlgorithms=curve25519-sha256` 绕过 |
| 2026-05-21 | 后端部署（视频上传 MIME 类型修复 + 邮箱登录修复 + auth_bloc 类型守卫） | 本地交叉编译 → scp（带 kex 参数）上传 → docker stop + docker cp + docker start 替换容器内二进制 → 健康检查通过 |
| 2026-05-21 | 上传视频后显示"此视频无法播放" | `post_detail_screen.dart` 视频播放器添加 `_videoError` 状态 + 错误 UI 显示 + 重试按钮 + `setLooping(true)`，构建 APK v0.1.6 |
| 2026-05-21 | 视频播放 Media error (unknown) — HTTP 明文流量被 Android 禁止 | `AndroidManifest.xml` 添加 `android:usesCleartextTraffic="true"`，原生 ExoPlayer 默认拒绝 http:// 视频 URL，构建 APK v0.1.7 |
| 2026-05-21 | 视频上传 Broken pipe — 服务器 ReadTimeout 15s 太短 | `main.go` ReadTimeout 改为 5 分钟，WriteTimeout 改为 5 分钟，已部署 |
| 2026-05-21 | 视频上传无实时进度 | `post_create_screen.dart` 添加 CircularProgressIndicator + 百分比文字 + LinearProgressIndicator |
| 2026-05-21 | 视频文件过大（67MB > 50MB 上限）| 服务器 `maxVideoSize` + `MaxMultipartMemory` 均改为 200MB |
| 2026-05-21 | 视频上传失败只显示"服务器未返回URL" | `api_client.dart` 新增 `UploadResult` 类替换 `String?` 返回，携带服务器错误消息 |
| 2026-05-21 | 发布帖子超时 DioException receive timeout | `api_client.dart` receiveTimeout 10s -> 30s，connectTimeout 10s -> 15s |
| 2026-05-21 | 帖子详情页"朋友圈"风格 + 视频比例不符 | `post_detail_screen.dart` 全面重写为小红书风格：AppBar 作者+关注、媒体区在最前、BoxFit.cover 保持比例、固定底部操作栏 |
| 2026-05-21 | 缺少关注/粉丝功能 | 后端 Follow 模型 + 5 个 API；前端 FollowRepository + FollowListScreen + 帖子详情关注按钮 |
| 2026-05-21 | 分享/报名/预约/帮助 等板块显示"开发中" | 分享底部弹窗 + 活动报名 API + MyRegistrationsScreen + 帮助与反馈页面 + 搜索 SearchDelegate + 通知页面 + 设置页清除缓存/关于/协议 |
| 2026-05-21 | 缺少收藏(Bookmark)系统 | 后端 Bookmark 模型 + POST/DELETE /posts/:id/bookmark + GET /my-bookmarks；前端 MyBookmarksScreen |
| 2026-05-21 | 缺少收货地址管理 | 后端 Address CRUD API；前端 AddressListScreen + AddressEditScreen（含默认地址） |
| 2026-05-21 | 缺少用户协议/隐私政策页面 | LegalPage 完整阅读页面 |
| 2026-05-21 | 视频帖无缩略图（只有黑色封面+播放图标）| Android MediaMetadataRetriever 平台通道 + 上传视频时自动生成缩略图并作为封面 |
| 2026-05-21 | 缺少站内聊天消息系统 | 后端 Message 模型 + 3 个 API（Send/Conversations/GetMessages）；前端 ConversationsScreen + ChatScreen 聊天气泡 UI |
| 2026-05-21 | 缺少兴趣圈层(Groups) | 后端 GroupMember 模型 + 9 个 API；前端 GroupListScreen + GroupDetailScreen（Tab 帖子/成员）+ GroupCreateScreen |
| 2026-05-21 | 缺少服务者平台前端 + 预约/评价 | ServiceProviderListScreen + ProviderDetailScreen（作品/评价/预约/星级）+ Booking API + Review API |
| 2026-05-21 | 缺少商家信用体系 | SellerRating 模型 + POST /rate/order/:id + GET /ratings/seller/:id + User 模型 seller_rating 字段 |
| 2026-05-21 | Code Review: seller_rating 列名错误 status->order_status + 事务错误未检查 + ID 错误 + Controller 泄漏 | 修正列名 + 事务添加 rollback + 重新查询获取正确ID + 添加 dispose |
| 2026-05-21 | 服务器容器启动失败（路由冲突）| Gin /orders 组路由冲突导致启动 panic；改为 /rate 和 /ratings 独立前缀 |

| 2026-05-21 | 视频比例不符 + 图片无法左右滑动 + 被遮挡 + 无全屏查看 | `post_detail_screen.dart` 重写：图片PageView左右滑动+页码指示器、视频BoxFit.contain+黑色背景保持原比例、点击图片InteractiveViewer全屏缩放(0.5x-4x)、Container白色背景自适应边框 |
| 2026-05-21 | "我要入驻"文案错误: "摊主" → "商家" | `profile_screen.dart` 文案修正 |
| 2026-05-21 | 兴趣圈层空态显示"暂无商品" | group_list_screen 改为"暂无小组" |
| 2026-05-21 | 首页Cosplay/周边/圈层入口大小不一致 + 圈层无独立入口 | 首页重构: Cosplay+周边一行(Expanded等宽)、圈层单独全宽行(独立入口)、导航至GroupListScreen |
| 2026-05-21 | 分享功能全部无效(复制/保存/微信/更多/长按) | Clipboard.setData复制链接、HttpClient下载保存图片、微信复制提示、更多菜单(举报/不感兴趣)、长按保存 |
| 2026-05-21 | 我的帖子/收藏无法跳转详情 + 收藏无封面预览 | my_posts_screen + my_bookmarks_screen 添加 onTap → PostDetailScreen + CachedNetworkImage封面 |
| 2026-05-21 | 关注数/粉丝数不增加 + 可关注自己 | User模型新增 follower_count/following_count 字段、Follow服务事务更新计数、前端隐藏自己帖子的关注按钮 + auth_bloc保存user_id到prefs |
| 2026-05-21 | APK错误打包为Flutter Demo Home Page | /home/jct 存在独立Flutter Demo项目；必须从 client/ 目录执行 flutter build |

## APK 信息
- Release APK: `/home/jct/nexusacg/client/build/app/outputs/flutter-apk/app-release.apk`
- 大小: 6.8MB (arm64-only)
- API 地址: `http://101.133.169.72:8080/api/v1`
- 版本: 0.2.1
- 构建时间: 2026-05-21 16:24
- 构建命令: `JAVA_HOME=/home/jct/jdks/jdk-17.0.12 flutter build apk --release --target-platform=android-arm64`
- **注意**: 安装前需先卸载旧版本

## DB 密码重置方法（容器重建后）
```bash
# 1. 修改 pg_hba.conf 为 trust
docker exec nexusacg-db sed -i 's/scram-sha-256/trust/g' /var/lib/postgresql/data/pg_hba.conf
docker restart nexusacg-db
sleep 3

# 2. 重置密码
docker exec nexusacg-db psql -U nexusacg -d nexusacg -c "ALTER USER nexusacg WITH PASSWORD 'nexusacg_dev_pass';"

# 3. 恢复认证
docker exec nexusacg-db sed -i 's/trust/scram-sha-256/g' /var/lib/postgresql/data/pg_hba.conf
docker restart nexusacg-db
```
