# 兴趣圈层 (Groups) 设计文档

## 后端

### 新增模型
- **GroupMember**: group_id, user_id, role(owner/admin/member), joined_at

### API 端点
- GET /groups — 列表(热门/搜索)
- POST /groups — 创建
- GET /groups/:id — 详情
- PUT /groups/:id — 编辑(owner)
- POST /groups/:id/join — 加入
- POST /groups/:id/leave — 退出
- GET /groups/:id/members — 成员列表
- GET /groups/:id/posts — 小组帖子
- GET /groups/my — 我的小组

### Service
- GroupService: Create/List/Get/Update/Delete
- Join: 检查重复，原子增加 member_count
- Leave: 检查 role，原子减少 member_count
- 小组帖子: Post模型已有 group_id 字段支持

## 前端

### GroupListScreen
- 搜索 + 热门(按成员数) + 我的小组入口 + 创建FAB

### GroupDetailScreen  
- SliverAppBar(封面) + 描述 + 加入/退出按钮
- TabBar: 帖子 / 成员
- 组长可管理(编辑/审核)

### GroupCreateScreen
- 名称 + 描述 + 可选封面上传

## 验证
- 创建 → 加入 → 发帖 → 查看成员 → 退出 全链路
