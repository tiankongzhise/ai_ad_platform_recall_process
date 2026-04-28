# 广告商业平台Recall回调处理系统 PRD

## 1. 产品概述

### 1.1 产品名称
广告商业平台Recall回调处理系统

### 1.2 产品定位
为广告商业平台提供统一的Recall回调处理服务，实现用户回调信息的接收、存储、查询和实时推送功能。

### 1.3 目标用户
- 广告平台运营人员
- 第三方回调服务提供商
- 系统集成开发者

---

## 2. 功能需求

### 2.1 用户认证模块

#### 2.1.1 用户注册
- 用户名（recall_service_name）：唯一标识，用于回调识别
- 密码：用户密码，需加密存储
- 支持Email/手机号等联系方式（用于通知推送）

#### 2.1.2 用户登录
- 用户名+密码登录
- 登录成功后返回Token
- 支持Token刷新

#### 2.1.3 用户注销
- 清除用户会话
- 使当前Token失效

#### 2.1.4 Token管理
- 登录后生成Token
- 支持重新生成Token（原Token失效）
- Token有效期设置
- 所有敏感接口需要Token鉴权

### 2.2 Recall回调处理模块

#### 2.2.1 回调接收接口 (GET /recall)
- **直接访问返回**：返回200及接口使用说明
- **正式调用格式**：
  ```
  /recall?status={urlcode(recall_service_name=XXXX&platform=XXXX&user_name=XXXX)}&other_params
  ```
- **参数说明**：
  - `recall_service_name`：注册的回调服务用户名
  - `platform`：回调来源平台
  - `user_name`：授权用户名称
  - `other_params`：其他自定义参数

#### 2.2.2 参数校验
- 校验status参数是否为URL编码格式
- 校验status内容是否为key=value格式
- 校验必填参数是否存在
- **错误提示**：明确指出缺失的参数（支持多个参数同时缺失提示）

#### 2.2.3 数据存储
- 存储所有回调参数
- 关联用户和平台信息
- 记录回调时间

### 2.3 查询接口模块

#### 2.3.1 回调记录查询
- **鉴权要求**：需要有效Token
- **查询条件**：
  - recall_service_name
  - platform
  - user_name
- **返回内容**：匹配的回调参数列表

### 2.4 通知推送模块

#### 2.4.1 通知接口设置
- 用户登录后可设置通知回调URL
- 支持HTTPS URL

#### 2.4.2 实时推送
- 当收到某用户的回调时
- 主动POST推送信息到设置的URL
- **推送参数**：recall_service_name, platform, user_name

---

## 3. 非功能需求

### 3.1 性能需求
- 回调响应时间 < 100ms
- 支持并发请求处理

### 3.2 兼容性需求
- 数据库可无感替换（MySQL ↔ PostgreSQL）
- 使用GORM框架实现数据库抽象

### 3.3 安全性需求
- 密码加密存储
- Token鉴权机制
- 防止SQL注入

### 3.4 可维护性
- 代码结构清晰
- 配置与代码分离
- 完整的日志记录

---

## 4. 技术架构

### 4.1 技术栈
- **前端**：HTML + JavaScript (Axios)
- **后端**：Go (Gin框架)
- **数据库**：MySQL（支持PgSQL）

### 4.2 项目结构
```
ai_ad_platform_recall_process/
├── cmd/                 # 入口文件
├── internal/           # 内部包
│   ├── handler/         # 处理器
│   ├── middleware/      # 中间件
│   ├── model/           # 数据模型
│   ├── service/         # 业务逻辑
│   └── repository/      # 数据访问
├── config/              # 配置
├── pkg/                 # 公共包
├── web/                 # 前端静态文件
└── docs/                # 文档
```

---

## 5. API接口设计

### 5.1 认证接口
| 接口 | 方法 | 描述 | 鉴权 |
|------|------|------|------|
| /api/auth/register | POST | 用户注册 | 否 |
| /api/auth/login | POST | 用户登录 | 否 |
| /api/auth/logout | POST | 用户注销 | 是 |
| /api/auth/refresh | POST | 刷新Token | 是 |

### 5.2 Recall接口
| 接口 | 方法 | 描述 | 鉴权 |
|------|------|------|------|
| /recall | GET | 接收回调 | 否 |
| /api/query | GET | 查询回调记录 | 是 |

### 5.3 通知接口
| 接口 | 方法 | 描述 | 鉴权 |
|------|------|------|------|
| /api/notify/set | POST | 设置通知URL | 是 |
| /api/notify/get | GET | 获取通知URL | 是 |

---

## 6. 数据库设计

### 6.1 用户表 (users)
| 字段 | 类型 | 描述 |
|------|------|------|
| id | BIGINT | 主键 |
| recall_service_name | VARCHAR(64) | 用户名（唯一） |
| password | VARCHAR(255) | 加密密码 |
| notify_url | VARCHAR(512) | 通知回调URL |
| created_at | DATETIME | 创建时间 |
| updated_at | DATETIME | 更新时间 |

### 6.2 Token表 (tokens)
| 字段 | 类型 | 描述 |
|------|------|------|
| id | BIGINT | 主键 |
| user_id | BIGINT | 用户ID |
| token | VARCHAR(128) | Token值 |
| expires_at | DATETIME | 过期时间 |
| created_at | DATETIME | 创建时间 |

### 6.3 回调记录表 (recall_records)
| 字段 | 类型 | 描述 |
|------|------|------|
| id | BIGINT | 主键 |
| recall_service_name | VARCHAR(64) | 服务用户名 |
| platform | VARCHAR(64) | 平台 |
| user_name | VARCHAR(128) | 用户名 |
| params | TEXT | 完整参数JSON |
| created_at | DATETIME | 创建时间 |

---

## 7. 里程碑

### M1: 基础框架搭建
- 项目结构搭建
- 数据库连接配置
- 用户注册登录基础功能

### M2: Recall回调核心功能
- /recall接口实现
- 参数校验逻辑
- 数据存储

### M3: 查询和推送功能
- 查询接口实现
- 通知推送功能
- 前端页面完善

### M4: 测试与部署
- 单元测试
- 集成测试
- 部署文档

---

## 8. 验收标准

1. 用户可以完成注册、登录、注销全流程
2. Token生成和刷新功能正常
3. /recall接口正确处理回调请求
4. 参数校验能准确提示缺失参数
5. 查询接口能正确返回匹配的回调记录
6. 通知推送功能正常触发
7. 数据库可从MySQL切换到PgSQL无报错
