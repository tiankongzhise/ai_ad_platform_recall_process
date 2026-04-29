# 广告商业平台 Recall 回调处理系统

## 项目简介

本项目用于接收广告平台回调、落库并按用户维度查询，同时提供完整账号体系（注册/登录/JWT/RefreshToken/ApiToken）与异步通知能力。

## 当前实现能力

- 用户注册、登录、注销、改密、忘记密码重置
- 注册/重置短信验证码（当前为 Mock 实现）
- JWT + RefreshToken + ApiToken 三类凭证管理
- 回调入口 `/recall`（`state=uid_platform_user_tag`）
- 回调数据查询（分页、最新一条、历史列表）
- 每用户通知地址配置与异步 POST 推送
- MySQL/PostgreSQL 双驱动支持（GORM）

## 技术栈

- Go 1.21+
- Gin
- GORM
- MySQL / PostgreSQL
- 原生 HTML + JavaScript 页面（`/web`）

## 目录结构

```text
ai_ad_platform_recall_process/
├── cmd/server/main.go
├── internal/
│   ├── config/
│   ├── handler/
│   ├── middleware/
│   ├── model/
│   ├── repository/
│   └── service/
├── pkg/
│   ├── database/
│   ├── response/
│   └── utils/
├── web/
├── config.yaml
└── 接口文档.md
```

## 快速开始

### 1) 环境准备

- Go 1.21+
- MySQL 8+ 或 PostgreSQL 13+

### 2) 配置

项目先读取 `config.yaml`，再用环境变量覆盖机密字段：

- `DB_HOST`
- `DB_PORT`
- `DB_USER`
- `DB_PASSWORD`
- `DB_NAME`
- `TOKEN_SECRET`

可参考 `.env.example`。

### 3) 安装依赖并运行

```bash
go mod download
go run cmd/server/main.go
```

默认地址：`http://localhost:8080`

## 接口总览

完整参数与返回值请看 `接口文档.md`，这里给出路由清单：

### 公开接口（无需鉴权）

- `GET /recall`
- `POST /api/auth/register`
- `POST /api/auth/login`
- `POST /api/auth/create-jwt-token`
- `POST /api/auth/refresh-jwt-by-refresh-token`
- `POST /api/auth/refresh-jwt-by-api-token`
- `POST /api/auth/send-code`
- `POST /api/auth/send-reset-code`
- `POST /api/auth/reset-password`
- `GET /api/auth/get-uid`
- `GET /api/auth/get-activate-uid`

### 鉴权接口（`Authorization: Bearer <JWT>`）

- `POST /api/auth/logout`
- `POST /api/auth/refresh`
- `GET /api/query`
- `GET /api/query/latest`
- `GET /api/history`
- `POST /api/notify/set`
- `GET /api/notify/get`
- `POST /api/account/change-password`
- `POST /api/account/delete`
- `GET /api/account/get-api-token`
- `POST /api/account/update-api-token`
- `GET /api/account/info`
- `GET /api/token/info`

## 统一响应格式

所有接口返回：

```json
{
  "code": 0,
  "message": "操作成功",
  "data": {}
}
```

- `code=0` 表示成功
- 非 0 表示业务错误（详见 `接口文档.md`）

## 说明

- 账户删除为软删除（`status=0` 且 `logout_at!=-1`）
- `/api/query*` 系列会强制按当前登录用户过滤，无法跨用户查询
- `send-code` 与 `send-reset-code` 当前使用 Mock 验证码：`930108`、`931216`
