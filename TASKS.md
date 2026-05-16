# 次元链 NexusACG — 开发任务清单

> 基于《项目计划书：次元链 (NexusACG)》+《OPC 团队项目方案》拆解
> 每个阶段完成后需整体测试，确保该阶段所有任务无 bug
> 被要求添加新功能时，在此文件追加新任务并设为 [ ] 待完成

---

## Phase 1：MVP（目标：4 人可运行的核心闭环）

**目标**：认证 + 商品双区 + 社区基础 + 支付接入 + 担保交易 + 活动列表 + AI 内容审核
**时间**：0-3 个月

### 1.1 认证系统 [PARTIAL]
- [x] JWT 注册/登录 (`internal/service/auth.go`, `internal/handler/auth.go`)
- [x] JWT 鉴权中间件 (`internal/middleware/auth.go`)
- [x] Refresh Token 管理 (`internal/model/model.go` RefreshToken)
- [x] CORS 中间件 (`internal/middleware/auth.go`)
- [x] 手机号 + 短信验证码登录 (`internal/service/sms.go` + `internal/handler/auth.go`) — 阿里云 SMS API + dev 模式 fallback，速率限制 1次/分钟 + 5次/天
- [x] 微信 OAuth 登录 (`internal/service/wechat_oauth.go` + `internal/handler/auth.go`) — 授权码流程 + 自动注册，需 WECHAT_OAUTH_APP_ID/SECRET
- [x] QQ OAuth 登录（Flutter 客户端 WebView 内嵌授权流程）

### 1.2 商品模块 [COMPLETED]
- [x] 商品 CRUD + 列表查询 (`internal/service/product.go`) — 7 个集成测试通过
- [x] 商品模型：双区字段 (zone: cosplay/merch) (`internal/model/model.go` Product)
- [x] 商品来源标注 (source_type: self_made/official/agent)
- [x] 图片列表 (JSONB)
- [x] 库存管理（WHERE stock >= N 防超卖）
- [x] 商品搜索（多关键词 ILIKE name/anime/character/description）
- [x] 分类关联查询 (Category JOIN, category_name 返回)
- [x] 商品标签过滤 (PostgreSQL JSONB `@>` 运算符，json.Marshal 安全编码)
- [x] 价格区间过滤 (min_price/max_price)
- [x] 多排序选项 (price_asc, price_desc, newest)
- [x] Category CRUD API (`internal/service/category.go`)

### 1.3 社区模块 [COMPLETED]
- [x] 帖子 CRUD + 列表 (`internal/service/post.go`) — 集成测试通过
- [x] 帖子模型：图文/视频类型 (`internal/model/model.go` Post)
- [x] 点赞功能 (`internal/model/model.go` Like) — 集成测试通过
- [x] 评论功能 (`internal/model/model.go` Comment) — ParentID 验证 + 审核状态
- [x] 帖子搜索（关键词 ILIKE title/content，多词分词搜索）— 集成测试通过
- [x] 评论嵌套回复 + 分页（ParentID 字段 + ListComments API）— 集成测试通过
- [x] 视频上传 + VOD 集成（后端支持视频上传 mp4/webm/mov，50MB 限制；Flutter 视频播放器集成）

### 1.4 活动模块 [COMPLETED]
- [x] 活动 CRUD + 列表 (`internal/service/event.go`) — 集成测试通过（含时间解析）
- [x] 活动模型：时间/地址/LBS (`internal/model/model.go` Event) — RFC3339 时间解析
- [x] 活动列表 + 详情 Flutter 页面（API 对接，分页加载，详情含状态标签/时间/地址）
- [x] 活动创建功能（Flutter 创建页面 + 表单验证，C 端发起，暂设用户阈值门控）
- [x] 漫展 AI 信息爬取聚合（event_crawler.go，bilibili + acg17 爬取 + 种子数据 fallback）

### 1.5 订单模块 [COMPLETED]
- [x] 创建订单 (`internal/service/order.go`) — 集成测试通过（事务 + 原子库存扣减）
- [x] 订单状态机 (pending → paid → shipped → completed)
- [x] 订单项 (OrderItem)
- [x] 等幂键 (IdempotencyKey)
- [x] 订单列表查询（分页 + 状态过滤，Preload Items）
- [x] 订单详情查询（按 order_no 查询，含 Items）
- [x] 取消订单（事务内恢复库存）— 集成测试通过
- [x] 退款流程（事务内恢复库存 + 状态更新）— 集成测试通过

### 1.6 支付模块 [COMPLETED]
- [x] 支付宝统一下单 TradeAppPay (`internal/service/payment/callback.go`) — 集成测试通过
- [x] 支付宝回调签名验证 RSA2 (`internal/service/payment/signature.go`)
- [x] 微信支付 v3 统一下单 + 回调 (`internal/service/payment/wechat.go`)
- [x] 支付幂等处理 (transaction_id 唯一 + FOR UPDATE 行锁)
- [x] 支付日志审计 (`internal/model/model.go` PaymentLog)
- [x] 订单超时取消逻辑 (`CancelTimeoutOrders`，事务内恢复库存)
- [x] 集成测试 9/9 通过 (`alipay_test.go` + `callback_test.go`)
- [x] 支付宝沙箱完整流程（SDK 签名验证 + 订单字符串生成 + 回调解析，sandbox AppID `2021006153686187`）
- [ ] 微信商户号注册 + 真实沙箱测试（需企业注册）
- [x] 担保交易/分账系统（确认收货后分账，PLATFORM_FEE_PERCENT 默认 5%，AUTO_RELEASE_DAYS 默认 7 天）
- [x] 订单超时自动取消 cron 部署（每 5 分钟 ticker，ORDER_TIMEOUT_MINUTES 环境变量配置）

### 1.7 基础设施 [COMPLETED]
- [x] Go + Gin 后端 (`cmd/server/main.go`)
- [x] PostgreSQL 数据库连接 (`internal/database/database.go`)
- [x] 配置管理 (`internal/config/config.go`)
- [x] 安全中间件（限流、安全头）(`internal/middleware/ratelimit.go`)
- [x] 数据库模型完整定义 (`internal/model/model.go`)
- [x] 缓存层（in-memory fallback，Redis 因网络 blocked 未安装）
- [x] 图片上传 OSS/MinIO（本地文件系统存储 + 接口化设计，后续可切换 MinIO/OSS，`internal/storage/storage.go`）— 3 个集成测试通过
- [x] Swagger API 文档（`swag init` 生成 45 个端点，`GET /swagger/index.html` 在线浏览，8 个 Tag 分组）
- [x] Docker Compose 部署（`docker-compose.yml` — Go 多阶段构建 + PostgreSQL 16 + Redis 7，health check 依赖编排）
- [x] GitHub Actions CI/CD（`.github/workflows/ci.yml` — test/lint/build 三 job，PostgreSQL 服务容器，go.sum 缓存）

### 1.8 AI 内容审核 [COMPLETED]
- [x] 阿里云内容安全 API 接入（`internal/service/moderation.go`，API 未配置时 fallback 到本地关键词过滤 + HMAC-SHA256 签名实现）
- [x] 帖子发布前自动审核（PostHandler Create 中集成 AutoModeratePost）
- [x] 人工复审后台 API（AdminHandler PendingPosts/ApprovePost/RejectPost）

### 1.9 管理后台 [COMPLETED]
- [x] 管理员角色 + 权限（复用 JWT 鉴权中间件，后续可扩展 RBAC）
- [x] 商品审核上架（PendingProducts / ApproveProduct / RejectProduct）
- [x] 帖子审核（PendingPosts / ApprovePost / RejectPost）
- [x] 活动管理（复用 Event CRUD，管理端可操作）
- [x] 订单审核/退款操作（ListOrders / ProcessRefund，事务内恢复库存）
- [x] 数据统计看板（GetDashboardStats：用户/商品/订单/帖子/收入统计）

### 1.10 Flutter 客户端 [NOT STARTED]
- [x] 项目初始化 + 路由
- [x] 登录/注册页面
- [x] 商品列表 + 详情页面
- [x] 商品双区切换
- [x] 帖子列表 + 详情 + 发布
- [x] 点赞/评论交互
- [x] 活动列表页面
- [x] 订单列表 + 详情
- [x] 支付流程（微信/支付宝 SDK 集成）
- [x] 用户个人中心

---

## Phase 2：增长（目标：差异化功能 + 用户增长）

**目标**：AI 虚拟试穿 + 角色推荐 + 服务者平台 + LBS 地图 + 兴趣圈层 + 视频
**时间**：3-6 个月

### 2.1 AI 虚拟试穿
- [ ] Stable Diffusion XL 部署（火山引擎/阿里云 PAI）
- [ ] IDM-VTON / OOTDiffusion 虚拟试穿模型
- [ ] ControlNet + OpenPose 姿态控制
- [ ] IP-Adapter + LoRA 角色适配
- [ ] 上传照片 → 生成效果图 API
- [ ] 效果图标注推荐商品

### 2.2 角色推荐引擎
- [ ] Neo4j 知识图谱搭建
- [ ] 角色 → 商品映射关系
- [ ] 推荐清单 API
- [ ] 用户行为数据收集

### 2.3 服务者平台
- [ ] ServiceProvider 模型扩展 (`internal/model/model.go`)
- [ ] 妆娘/摄影师入驻申请
- [ ] 服务者个人主页
- [ ] 排期管理
- [ ] 预约下单
- [ ] 评价系统
- [ ] 服务者认证 (加V)

### 2.4 LBS 活动地图
- [ ] 高德地图 Flutter 集成
- [ ] 附近漫展/活动查询
- [ ] 活动导航
- [ ] 服务者排期与活动关联

### 2.5 兴趣圈层
- [ ] Group 模型 + CRUD (`internal/model/model.go` Group)
- [ ] 用户加入/退出小组
- [ ] 小组内帖子
- [ ] 小组搜索
- [ ] 热门小组推荐

### 2.6 视频内容
- [ ] 阿里云/腾讯云 VOD 集成
- [ ] 视频上传 API
- [ ] 视频播放页面
- [ ] 视频审核

### 2.7 AI 内容审核增强
- [ ] 图片自动审核
- [ ] 文本敏感词过滤
- [ ] 人工复审后台
- [ ] 审核日志

### 2.8 Phase 2 整体测试
- [ ] 全量 API 集成测试
- [ ] Flutter 端到端测试
- [ ] AI 试穿效果验证
- [ ] 性能压测

---

## Phase 3：成熟（目标：平台化 + 商业化）

**目标**：3D 预览 + 数据分析 + 广告 + 票务 + 国际化
**时间**：6-12 个月

### 3.1 3D 商品预览
- [ ] NeRF/3DGS 轻量方案调研
- [ ] 3D 模型生成 API
- [ ] Flutter 3D 展示组件
- [ ] 卖家 3D 上传工具

### 3.2 数据分析平台
- [ ] 用户画像收集
- [ ] 消费趋势分析
- [ ] 推荐算法优化
- [ ] 数据看板（Grafana 或自建）
- [ ] 数据报告导出

### 3.3 广告系统
- [ ] 信息流广告
- [ ] 开屏广告
- [ ] 活动置顶广告
- [ ] 广告投放管理后台
- [ ] 广告效果统计

### 3.4 票务系统
- [ ] 漫展线上购票
- [ ] 电子票生成
- [ ] 验票 API
- [ ] 退票流程

### 3.5 信息爬取聚合
- [ ] 漫展信息爬虫
- [ ] 商品比价聚合
- [ ] AI 去重 + 信息融合
- [ ] 用户申诉通道

### 3.6 国际化
- [ ] 多语言支持 (i18n)
- [ ] 海外支付 (PayPal/Stripe)
- [ ] 跨境物流 API
- [ ] 多时区支持

### 3.7 合规与安全加固
- [ ] ICP/EDI 许可证办理
- [ ] 等保二级备案
- [ ] 数据出境合规
- [ ] 隐私政策完善
- [ ] 安全渗透测试
- [ ] 定期安全审计

### 3.8 Phase 3 整体测试
- [ ] 全量功能回归测试
- [ ] 多语言/多地区测试
- [ ] 高并发压测
- [ ] 安全扫描
- [ ] 合规审查

---

## 新增任务（按需添加）

### [用户反馈] 物流追踪 + 签收后自动确认收货
- [ ] 接入快递查询 API（快递100 / 菜鸟裹裹），根据 tracking_number + carrier 查询物流轨迹
- [ ] 订单物流状态实时展示（揽收 → 运输 → 派送 → 签收）
- [ ] 物流显示"已签收"后启动签收倒计时（默认 7 天，到期自动确认收货）
- [ ] `AutoReleaseOrders` 改造：优先使用签收时间计算超时，无签收数据时回退到发货时间计算
- [ ] 物流查询 cron（每 2 小时轮询有物流单号但未签收的订单）

### [用户反馈] 退换货申请流程
- [ ] RefundApplication 模型（order_id, user_id, refund_type: refund_only/return_refund, reason, evidence_urls, amount, status: pending/seller_review/approved/rejected/completed, created_at, updated_at）
- [ ] 买家提交退换货申请 API（POST /orders/:order_no/refund-application）— 必填理由、金额，可选上传凭证图片
- [ ] 商家审核退换货 API（POST /admin/refund-applications/:id/approve、POST /admin/refund-applications/:id/reject）
- [ ] 退换货申请状态机（pending → seller_review → approved/rejected → completed）
- [ ] 审核通过后触发退款执行（调用现有 Refund 逻辑 + 库存恢复）
- [ ] 支持部分退款（按金额而非整单）
- [ ] 退换货申请列表查询（买家端、商家端、管理端）

### [用户反馈] 纠纷申诉系统
- [ ] Dispute 模型（order_id, claimant_id, respondent_id, claim_type: quality/counterfeit/not_received/misdescription, description, evidence_urls, status: open/mediating/resolved/closed, resolution: refund/return/partial_refund/none, created_at, updated_at）
- [ ] 买家发起纠纷申诉 API（POST /disputes）— 必填申诉类型、描述，可选上传凭证
- [ ] 商家响应纠纷 API（POST /disputes/:id/respond）— 商家回复/提供反证
- [ ] 平台仲裁 API（POST /admin/disputes/:id/resolve）— 管理员裁决并执行结果
- [ ] 纠纷状态流转逻辑 + 超时自动升级（商家 N 天不响应 → 平台自动介入）
- [ ] 纠纷列表查询（买家端、商家端、管理端）
- [ ] 纠纷记录关联到商家信用（影响信用评分）

### [用户反馈] 商家信用体系 + 推荐优先级
- [ ] SellerRating 模型（seller_id, buyer_id, order_id, rating: 1-5, comment, created_at）— 买家对订单卖家评分
- [ ] 订单完成后可评价卖家 API（POST /orders/:order_no/rate）
- [ ] 商家信用分聚合（平均评分、评价数量、纠纷率）— 实时或定时计算
- [ ] User/Product 模型关联信用字段（seller_rating, seller_review_count）
- [ ] 商品搜索/列表排序新增信用算法（popular 模式融入信用加权）
- [ ] 商家信用展示（商品详情页展示卖家信用、店铺入口）
- [ ] 低信用商家限制（评分低于阈值时降权展示或禁止上架新商品）

<!-- 新功能需求在此追加，格式：
### [需求来源] 功能名称
- [ ] 子任务 1
- [ ] 子任务 2
-->

---

## 阶段验收标准

| 阶段 | 验收条件 |
|------|----------|
| Phase 1 | 用户可注册 → 浏览商品 → 下单 → 支付 → 订单状态更新 → 发帖互动，全链路打通 |
| Phase 2 | Phase 1 + AI 试穿可用 + 服务者预约 + 地图导航 + 小组活跃 |
| Phase 3 | Phase 2 + 3D 预览 + 数据分析 + 广告变现 + 多语言可用 |
