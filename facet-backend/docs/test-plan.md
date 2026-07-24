# Facet 全面测试计划

## 环境信息

| 项目 | 值 |
|------|-----|
| 前端地址 | http://localhost:5174 |
| 后端地址 | http://localhost:8080 |
| 系统管理员 | admin@facet.com / admin123 |
| 租户管理员 | admin@acme.com / NewStr0ng!Pass99 |
| 只读用户 | viewer@acme.com / Viewer@123 |

---

## 1. 认证与授权

### 1.1 登录页面

| # | 行为 | 预期结果 |
|---|------|---------|
| 1.1.1 | 访问 http://localhost:5174/login | 显示 Facet 品牌标题 + 邮箱/密码/验证码输入框 + 登录按钮 |
| 1.1.2 | 验证码图片加载 | 验证码图片正常显示（非空白/断裂图标），点击可刷新 |
| 1.1.3 | 输入正确凭据 + 正确验证码后点击 Sign In | 登录成功，跳转至 /overview |
| 1.1.4 | 输入错误密码 | 显示红色错误提示（如 "invalid email or password"），验证码自动刷新 |
| 1.1.5 | 输入错误验证码 | 显示错误提示，验证码自动刷新 |
| 1.1.6 | 邮箱/密码/验证码任一为空时点击登录 | 显示 "All fields are required" |
| 1.1.7 | 已经登录后访问 /login | 自动重定向到 /overview |
| 1.1.8 | 输入 ` admin@facet.com`（带空格） | 自动去除前后空格，登录成功 |

### 1.2 权限控制

| # | 行为 | 预期结果 |
|---|------|---------|
| 1.2.1 | 未登录直接访问 /overview | 自动重定向到 /login |
| 1.2.2 | 租户管理员（admin@acme.com）登录 | 侧边栏**不显示** "System" 区域（Tenants / System Notifications） |
| 1.2.3 | 系统管理员（admin@facet.com）登录 | 侧边栏**显示** "System" 区域（Tenants / System Notifications） |
| 1.2.4 | 只读用户（viewer@acme.com）登录 | 侧边栏显示功能菜单，但部分页面的创建/编辑/删除按钮可能隐藏或禁用 |

### 1.3 退出登录

| # | 行为 | 预期结果 |
|---|------|---------|
| 1.3.1 | 点击右上角头像 → 点击 "Logout" | 清除 token，跳转至 /login |
| 1.3.2 | 退出后按浏览器「后退」 | 无法回到之前页面，回到 /overview 时重定向到 /login |

---

## 2. 系统管理（系统管理员专有）

### 2.1 租户管理

| # | 行为 | 预期结果 |
|---|------|---------|
| 2.1.1 | 左侧导航 "System" → "Tenants" | 显示租户列表（表格：Name, Level, Status, Created At, Actions） |
| 2.1.2 | 列表为空时显示 | 显示空状态提示 "No tenants found" |
| 2.1.3 | 点击 "Create" 按钮 | 弹出对话框：Name / Admin Email / Admin Password / Admin Role |
| 2.1.4 | 填写必填字段后提交 | 创建成功显示成功提示，列表刷新显示新租户 |
| 2.1.5 | 创建时邮箱格式错误 | 显示验证错误提示，不提交 |
| 2.1.6 | 创建时密码少于8位 | 显示验证错误提示 |
| 2.1.7 | 点击租户行中的编辑图标 | 弹出编辑对话框：Name(禁用) / Level(下拉) / Status(下拉) |
| 2.1.8 | 修改 Level 为 Enterprise 后保存 | 更新成功，列表显示新 Level |
| 2.1.9 | 修改 Status 为 Suspended 后保存 | 该租户下的 SDK 评估应返回 tenant inactive |
| 2.1.10 | 点击重置密码图标 | 弹出对话框显示 Name(禁用) + New Password 输入框 |
| 2.1.11 | 重置密码后，用新密码登录 | 租户管理员可用新密码登录成功 |
| 2.1.12 | 点击归档（垃圾桶）图标 | 弹出确认对话框，确认后删除 |
| 2.1.13 | 搜索栏输入租户名关键词搜索 | 表格过滤显示匹配的租户 |

---

## 3. 环境管理

### 3.1 环境列表

| # | 行为 | 预期结果 |
|---|------|---------|
| 3.1.1 | 左侧导航 → "Environments" | 显示环境列表（表格：Name, SDK Key, Status, Created At, Actions） |
| 3.1.2 | 搜索 + 状态过滤（All / Active / Inactive） | 表格按条件过滤 |

### 3.2 创建环境

| # | 行为 | 预期结果 |
|---|------|---------|
| 3.2.1 | 点击 "Create" | 弹出对话框：Name / Description / Active 开关 |
| 3.2.2 | 填写 Name + Description，保持 Active | 创建成功，SDK Key 自动生成（格式 sdk-xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx） |
| 3.2.3 | 列表显示 SDK Key（等宽字体） | ✅ |
| 3.2.4 | 创建后 Status Chip 显示 "Active"（绿色） | ✅ |

### 3.3 编辑环境

| # | 行为 | 预期结果 |
|---|------|---------|
| 3.3.1 | 点击行中编辑图标 | 弹出编辑对话框，预填当前值 |
| 3.3.2 | 修改 Name，关闭 Active 开关 | 更新成功，Status Chip 变为 "Inactive"（灰色） |

### 3.4 删除环境

| # | 行为 | 预期结果 |
|---|------|---------|
| 3.4.1 | 点击删除图标 | 弹出确认对话框 |
| 3.4.2 | 确认删除 | 环境从列表消失（软删除） |

---

## 4. 用户管理

### 4.1 用户列表

| # | 行为 | 预期结果 |
|---|------|---------|
| 4.1.1 | 左侧导航 → "Users" | 显示用户列表（表格：Display Name, Email, Role, Created At, Actions） |
| 4.1.2 | 搜索姓名或邮箱 | 表格按条件过滤 |

### 4.2 创建用户

| # | 行为 | 预期结果 |
|---|------|---------|
| 4.2.1 | 点击 "Create" | 弹出对话框：Display Name / Email / Password / Role |
| 4.2.2 | 填写所有字段，Role 选 "Read & Write" | 创建成功，列表显示新用户 |
| 4.2.3 | 用新用户凭据登录 | 登录成功 |

### 4.3 编辑用户

| # | 行为 | 预期结果 |
|---|------|---------|
| 4.3.1 | 点击编辑图标 | 弹出编辑对话框，可修改 Display Name / Email / Role |
| 4.3.2 | 修改 Role 后保存 | 更新成功 |

### 4.4 查看用户详情

| # | 行为 | 预期结果 |
|---|------|---------|
| 4.4.1 | 点击查看图标（眼睛） | 弹出详情对话框，显示 ID / Display Name / Email / Role / 创建/更新时间 |

### 4.5 重置密码

| # | 行为 | 预期结果 |
|---|------|---------|
| 4.5.1 | 点击重置密码图标（锁定） | 弹出确认对话框 → 确认后重置 |
| 4.5.2 | 重置后用新密码登录 | ✅ |

### 4.6 删除用户

| # | 行为 | 预期结果 |
|---|------|---------|
| 4.6.1 | 点击删除图标 | 弹出确认对话框 → 确认后删除 |
| 4.6.2 | 被删除用户尝试登录 | 返回 "user not found" |

---

## 5. 功能开关管理

### 5.1 开关列表

| # | 行为 | 预期结果 |
|---|------|---------|
| 5.1.1 | 左侧导航 → "Feature Flags" | 显示开关列表前检查**是否已选择环境** |
| 5.1.2 | **未选择环境**时 | 显示蓝色提示：请在顶部选择环境 |
| 5.1.3 | 选择环境后 | 显示该环境的开关列表（表格：Name, Type, Status, Rollout, Expires, Updated At, Actions） |
| 5.1.4 | 搜索 + Show archived 切换 | 按名称搜索，显示/隐藏已归档开关 |
| 5.1.5 | Status Chip 显示 | Enabled 为绿色，Disabled 为灰色 |
| 5.1.6 | 已归档行变淡（opacity 0.6）并显示 Archived 标签 | ✅ |

### 5.2 创建开关

| # | 行为 | 预期结果 |
|---|------|---------|
| 5.2.1 | 选择环境后点击 + 按钮 | 弹出创建对话框：Name / Description / Type / Enabled / Rollout / 过期时间/ 白名单规则 / 用户段 |
| 5.2.2 | 填写 Name（test-flag-1），选择 Type=RELEASE，保持 Disabled | 创建成功，列表显示 Enabled=Disabled 的状态 |
| 5.2.3 | Type 下拉可选值 | RELEASE / EXPERIMENT / OPERATIONAL / PERMISSION |
| 5.2.4 | 启用 Rollout 百分比滑块 | 滑块 0-100，步长 5 |
| 5.2.5 | 添加白名单规则：Account ID + User IDs（逗号分隔） | 保存后规则生效 |
| 5.2.6 | 添加用户段关联 | 搜索已有用户段，支持多选 |
| 5.2.7 | 设置过期时间，关闭「Never expires」开关 | 显示 datetime-local 选择器 |

### 5.3 编辑开关

| # | 行为 | 预期结果 |
|---|------|---------|
| 5.3.1 | 点击编辑图标 | 弹出编辑对话框，Name 禁用，其余可修改 |
| 5.3.2 | 将 Enabled 打开，Rollout=50% | 保存后列表显示 Enabled=绿色，Rollout=50% |
| 5.3.3 | 添加/删除白名单规则 | 保存后规则更新 |

### 5.4 查看开关详情

| # | 行为 | 预期结果 |
|---|------|---------|
| 5.4.1 | 点击查看图标（眼睛） | 弹出详情对话框：Name, Description, Type, Enabled, Rollout, Expires, 创建/更新时间，用户段列表，白名单 |

### 5.5 开关变更历史

| # | 行为 | 预期结果 |
|---|------|---------|
| 5.5.1 | 点击时钟图标 | 弹出历史对话框，按时间倒序显示变更记录（change type, name, enabled status, expires） |
| 5.5.2 | 创建/编辑开关后查看历史 | 历史记录包含创建和每次编辑的记录 |

### 5.6 归档 / 取消归档

| # | 行为 | 预期结果 |
|---|------|---------|
| 5.6.1 | 点击归档图标（盒子） | 成功归档，行变淡，显示 Archived 标签 |
| 5.6.2 | 显示归档开关时，点取消归档图标 | 恢复为未归档状态 |

### 5.7 删除开关

| # | 行为 | 预期结果 |
|---|------|---------|
| 5.7.1 | 点击删除图标 | 确认后从列表中删除 |

---

## 6. 用户段管理

### 6.1 用户段列表

| # | 行为 | 预期结果 |
|---|------|---------|
| 6.1.1 | 左侧导航 → "Segments" | 显示用户段列表（表格：Name, Match Type, Conditions, Updated At, Actions） |
| 6.1.2 | Match Type 显示 | ALL = 蓝色标签，ANY = 默认标签 |

### 6.2 创建用户段

| # | 行为 | 预期结果 |
|---|------|---------|
| 6.2.1 | 点击 "Create" | 弹出对话框：Name / Description / Match Type / Conditions |
| 6.2.2 | Match Type 选 ALL，添加条件：Attribute Key=plan, Operator=equals, Value=beta | 创建成功 |
| 6.2.3 | 点击 "+Condition" 添加多个条件 | 可添加多条，每条可单独删除 |
| 6.2.4 | 12种操作符可用 | equals, not_equals, contains, not_contains, starts_with, ends_with, in, not_in, gt, gte, lt, lte |

### 6.3 编辑用户段

| # | 行为 | 预期结果 |
|---|------|---------|
| 6.3.1 | 点击编辑图标 | 弹出编辑对话框，预填当前值 |
| 6.3.2 | 修改 Match Type 或增删条件 | 保存后更新 |

### 6.4 删除用户段

| # | 行为 | 预期结果 |
|---|------|---------|
| 6.4.1 | 点击删除图标 | 弹出警告对话框（提示可能已关联开关）→ 确认后删除 |

---

## 7. 通知系统

### 7.1 系统通知（系统管理员）

| # | 行为 | 预期结果 |
|---|------|---------|
| 7.1.1 | 系统管理员登录 → "System" → "System Notifications" | 显示通知列表（Title, Level, Type, Created At, Actions） |
| 7.1.2 | 点击 "Create" | 弹出对话框：Title / Content / Level(normal/important/urgent) / Type(announcement/product_update/security_alert) |
| 7.1.3 | Level 和 Type 的 Chip 颜色 | urgent=error(红), important=warning(黄), normal=info(蓝)；Type=outlined 变体 |
| 7.1.4 | 创建新通知 | 成功创建 |
| 7.1.5 | 查看详情 | 弹窗显示完整 Content（pre-wrap） |
| 7.1.6 | 编辑通知 | 可修改全部字段 |
| 7.1.7 | 删除通知 | 确认后删除 |

### 7.2 用户通知

| # | 行为 | 预期结果 |
|---|------|---------|
| 7.2.1 | 顶部导航铃铛图标 | 显示未读数量角标（红点） |
| 7.2.2 | 点击铃铛 → "Notifications" | 显示通知列表（每条：Level Chip, Title(粗体=未读), Content, 时间, Mark Read 按钮） |
| 7.2.3 | 点击 "Mark Read"（单条） | 该条变灰，标记已读 |
| 7.2.4 | 点击 "Mark all as read" | 所有通知标记已读，角标消失 |
| 7.2.5 | 系统管理员发布新通知后 | 所有租户用户的通知列表中出现新通知，角标更新 |

### 7.3 通知角标自动刷新

| # | 行为 | 预期结果 |
|---|------|---------|
| 7.3.1 | 等待 60 秒或发布新通知 | 角标自动更新未读计数 |

---

## 8. SDK 评估（API 测试）

### 8.1 基础评估

| # | 行为 | 预期结果 |
|---|------|---------|
| 8.1.1 | `GET /api/v1/evaluate?sdk_key=xxx&feature_name=test-flag-1&account_id=user001&app_name=myapp` | 返回 `{"code":200,"data":{"matched":false}}` |
| 8.1.2 | `matched=true` 当开关 Enabled 且命中白名单/百分比 | ✅ |
| 8.1.3 | 缺少 `sdk_key` | 400 `missing sdk key` |
| 8.1.4 | 缺少 `feature_name` | 400 `missing feature name` |
| 8.1.5 | 缺少 `app_name` | 400 `missing app name` |
| 8.1.6 | 无效 SDK Key | 404 `invalid sdk key` |
| 8.1.7 | 环境 Inactive | 400 `environment is inactive` |
| 8.1.8 | 租户 Suspended | 400 `tenant is inactive` |

### 8.2 百分比发布

| # | 行为 | 预期结果 |
|---|------|---------|
| 8.2.1 | 开关 Rollout=50%，同一 user_id 多次评估 | 每次结果一致（一致性哈希） |
| 8.2.2 | Rollout=0% | 所有 `matched=false` |
| 8.2.3 | Rollout=100% | 所有 `matched=true` |

### 8.3 白名单规则

| # | 行为 | 预期结果 |
|---|------|---------|
| 8.3.1 | 创建 Account 规则：Account ID=team-a，无 User IDs | 所有 `account_id=team-a` 的请求 `matched=true` |
| 8.3.2 | 创建 Account+User 规则：Account ID=team-a，User IDs=user1,user2 | 仅 user1 和 user2 的请求 `matched=true`（team-a 下） |
| 8.3.3 | 创建仅 User 规则（空 Account ID）：User IDs=user1 | 不论 account_id，user1 都命中 |

### 8.4 用户段匹配

| # | 行为 | 预期结果 |
|---|------|---------|
| 8.4.1 | 开关关联用户段（plan=beta + matchType=ALL），`context={"plan":"beta"}` | `matched=true` |
| 8.4.2 | `context={"plan":"production"}` | `matched=false` |
| 8.4.3 | 不传 context | 段条件不评估 |

---

## 9. UI/UX 通用

### 9.1 侧边栏导航

| # | 行为 | 预期结果 |
|---|------|---------|
| 9.1.1 | 每个导航项点击 | 正确跳转，选中项高亮 |
| 9.1.2 | 页面刷新后保持当前导航高亮 | ✅ |

### 9.2 环境选择器

| # | 行为 | 预期结果 |
|---|------|---------|
| 9.2.1 | 顶部栏环境下拉 | 显示当前租户所有活跃环境 |
| 9.2.2 | 切换环境 | Feature Flags 页面数据刷新为所选环境的数据 |
| 9.2.3 | URL 参数同步 | 切换环境时 URL 内 envId 同步更新 |

### 9.3 帮助页面

| # | 行为 | 预期结果 |
|---|------|---------|
| 9.3.1 | 点击 "?" 图标 → Help Center | 显示功能卡片和核心概念说明 |
| 9.3.2 | 页面内容 | Multi-Tenant, Feature Flags, Getting Started, Best Practices 卡片 - Core Concepts 区域 - 使用提示 |

### 9.4 Profile 弹窗

| # | 行为 | 预期结果 |
|---|------|---------|
| 9.4.1 | 点击头像 → Profile | 弹窗显示 User ID, Display Name, Email, Role, Created At, Last Login |
| 9.4.2 | 关闭弹窗 | 回到原页面 |

### 9.5 首页 / Overview

| # | 行为 | 预期结果 |
|---|------|---------|
| 9.5.1 | 登录后自动跳转 /overview | 显示欢迎语 + 4 张功能卡片（Feature Flags / A/B Testing / Multi-Tenant / Real-time Evaluation） |

---

## 10. 深色模式

| # | 行为 | 预期结果 |
|---|------|---------|
| 10.1 | 点击顶部栏太阳/月亮图标 | 切换为深色模式，背景变深色，文字变浅色 |
| 10.2 | 再次点击 | 切换回浅色模式 |
| 10.3 | 刷新页面 | 保持之前选择的模式（localStorage 持久化） |

---

## 11. 错误边界

| # | 行为 | 预期结果 |
|---|------|---------|
| 11.1 | 模拟组件崩溃 | 显示降级 UI（非白屏），提示错误信息 |
| 11.2 | API 返回 401 | 自动清除 token，跳转至 /login |
| 11.3 | 网络断开时操作 | 显示 "Network error" 提示 |

---

## 12. 边界与异常场景

| # | 行为 | 预期结果 |
|---|------|---------|
| 12.1 | 创建同名 Feature Flag（同一环境） | 后端返回唯一约束错误 |
| 12.2 | 分页：列表超过 1 页时切换页码 | 正确显示该页数据 |
| 12.3 | 分页：每页行数切换（10/20/50） | 按选择的数量刷新列表 |
| 12.4 | 搜索：输入超长文本 | 无异常，返回空结果或正常搜索 |
| 12.5 | 连续快速点击创建/保存按钮 | 防止重复提交（按钮 disabled） |
| 12.6 | 浏览器窗口缩放 | 页面响应式布局正常，无重叠/溢出 |
| 12.7 | 空列表时创建第一条数据 | 创建后列表正常显示 |
