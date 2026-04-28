# 广告商业平台Recall回调处理系统

## 项目简介

这是一个为广告商业平台设计的Recall回调处理系统，提供用户认证、回调接收、数据存储、查询和实时通知推送功能。

## 技术栈

- **后端**: Go 1.21+, Gin Web框架
- **数据库**: MySQL/PostgreSQL (通过GORM支持)
- **前端**: HTML + JavaScript (Axios)

## 功能特性

- ✅ 用户注册、登录、注销
- ✅ Token生成和刷新
- ✅ Recall回调接收和参数校验
- ✅ 回调记录查询（支持分页）
- ✅ 实时通知推送（异步）
- ✅ 数据库无感切换（MySQL ↔ PostgreSQL）

## 项目结构

```
ai_ad_platform_recall_process/
├── cmd/server/           # 程序入口
├── internal/
│   ├── config/           # 配置管理
│   ├── handler/          # HTTP处理器
│   ├── middleware/       # 中间件
│   ├── model/            # 数据模型
│   ├── repository/       # 数据访问层
│   └── service/          # 业务逻辑
├── pkg/
│   ├── database/         # 数据库连接
│   ├── response/         # 统一响应
│   └── utils/            # 工具函数
├── web/                  # 前端静态文件
├── scripts/              # 脚本
├── config.yaml           # 配置文件
└── README.md
```

## 快速开始

### 1. 环境要求

- Go 1.21+
- MySQL 8.0+ 或 PostgreSQL 13+

### 2. 数据库配置

编辑 `config.yaml`:

```yaml
database:
  driver: "mysql"  # 切换为 "postgres" 使用PostgreSQL
  host: "localhost"
  port: 3306
  user: "root"
  password: "password"
  dbname: "recall_platform"
```

### 3. 初始化数据库

```bash
mysql -u root -p < scripts/init.sql
```

### 4. 安装依赖

```bash
go mod download
```

### 5. 运行

```bash
go run cmd/server/main.go
```

服务将在 `http://localhost:8080` 启动。

## API接口

### 认证接口

| 接口 | 方法 | 描述 | 鉴权 |
|------|------|------|------|
| `/api/auth/register` | POST | 用户注册 | 否 |
| `/api/auth/login` | POST | 用户登录 | 否 |
| `/api/auth/logout` | POST | 用户注销 | 是 |
| `/api/auth/refresh` | POST | 刷新Token | 是 |

### Recall接口

| 接口 | 方法 | 描述 | 鉴权 |
|------|------|------|------|
| `/recall` | GET | 接收回调 | 否 |
| `/api/query` | GET | 查询回调记录 | 是 |

### 通知接口

| 接口 | 方法 | 描述 | 鉴权 |
|------|------|------|------|
| `/api/notify/set` | POST | 设置通知URL | 是 |
| `/api/notify/get` | GET | 获取通知URL | 是 |

## Recall接口使用

### 回调格式

```
GET /recall?state={uid}_{platform}_{user_tag}&other_params=value
```

### 示例

```bash
# URL编码后的请求
curl "http://localhost:8080/recall?state=e8b5f1a2c3d4e5f6a7b8c9d0e1f2a3b4_2_10001"
```

### 参数说明

| 参数 | 必填 | 说明 |
|------|------|------|
| uid | 是 | 用户唯一UID |
| platform | 是 | 回调来源平台 |
| user_tag | 是 | 授权用户标签 |
| other_params | 否 | 其他自定义参数 |

## 前端页面

- `/web/index.html` - 控制台主页
- `/web/login.html` - 登录页面
- `/web/register.html` - 注册页面

## 错误码

| 错误码 | 说明 |
|--------|------|
| 0 | 成功 |
| 1001 | 用户名已存在 |
| 1002 | 用户名或密码错误 |
| 1003 | Token无效或已过期 |
| 2001 | 缺少必填参数 |
| 3001 | 通知URL格式错误 |

## 数据库切换

切换数据库只需修改 `config.yaml` 中的 `driver` 字段：

- `mysql` - 使用MySQL
- `postgres` - 使用PostgreSQL

GORM会自动处理SQL方言差异。

## License

MIT
