# 广告商业平台 Recall 回调处理系统 PRD（实现对齐版）

## 1. 产品目标

为广告平台与集成方提供一个统一、可鉴权、可追踪的 Recall 回调接入服务，支持：

- 回调接收与存储
- 用户维度数据查询
- 通知回推
- 账号与凭证管理

## 2. 已落地功能范围

### 2.1 账号与认证

- 注册：`username + password + phone + code`
- 登录：返回 JWT（`token`）
- 注销：使当前 JWT 失效
- 忘记密码：发送重置码、校验后重置
- 登录后改密
- 账户注销（软删除）
- 用户名查询 UID（含活跃/全部两种）

### 2.2 凭证体系

- JWT（短期）
- RefreshToken（长期刷新）
- ApiToken（长期静态凭证）
- 支持：
  - ApiToken 换取 JWT/RefreshToken
  - RefreshToken 刷新 JWT/RefreshToken
  - ApiToken 直接刷新 JWT/RefreshToken
  - 查询当前 Token 信息
  - 更换 ApiToken（并清理旧 JWT/RefreshToken）

### 2.3 Recall 回调

- 回调入口：`GET /recall`
- `state` 格式：`uid_platform_user_tag`
- 校验规则：
  - `uid`：32 位十六进制
  - `platform`：1~13 位数字
  - `user_tag`：1~13 位数字
- 存储字段：`user_name/uid/platform/user_tag/params/created_at`
- 支持附加 query 参数透传存储

### 2.4 查询能力

- `GET /api/query`：分页查询
- `GET /api/query/latest`：最新一条
- `GET /api/history`：历史列表（最多 1000 条）
- 所有查询在服务端强制限定为当前登录用户

### 2.5 通知能力

- 配置通知地址：`/api/notify/set`
- 查询通知地址：`/api/notify/get`
- 回调成功后异步 POST 通知
- 通知内容：`user_name/platform/user_tag`
- 支持超时与重试策略

## 3. 核心业务约束

- 用户注销后不可作为活跃账户登录/鉴权
- 鉴权统一使用 `Authorization: Bearer <JWT>`
- 所有接口统一返回 `code/message/data`
- 数据库驱动支持 `mysql` 与 `postgres`

## 4. 非功能要求（当前实现）

- 并发处理：由 Gin + Go 协程支持
- 数据安全：密码 bcrypt 存储
- 认证安全：JWT HS256 签名，支持过期控制
- 可维护性：分层结构（handler/service/repository/model）

## 5. 数据模型（产品视角）

- 用户：`user_name`、`uid`、`phone`、`api_token`、`status`、`logout_at`、`notify_url`
- Token：JWT 记录，按 `user_id` 关联
- RefreshToken：刷新令牌记录，按 `user_id` 关联
- Recall 记录：回调核心字段与扩展参数

## 6. API 清单

完整接口参数与返回值详见 `接口文档.md`。

关键对外接口：

- 回调接收：`GET /recall`
- 认证与凭证：`/api/auth/*`、`/api/token/info`
- 账户管理：`/api/account/*`
- 回调查询：`/api/query`、`/api/query/latest`、`/api/history`
- 通知设置：`/api/notify/set`、`/api/notify/get`

## 7. 验收标准（按当前实现）

1. 可完成注册、登录、鉴权访问、注销闭环
2. 支持 JWT/RefreshToken/ApiToken 全链路操作
3. `/recall` 能校验并落库 `state=uid_platform_user_tag`
4. 查询接口仅返回当前登录用户的数据
5. 回调成功可触发异步通知
6. 在 MySQL / PostgreSQL 下可启动并迁移模型
