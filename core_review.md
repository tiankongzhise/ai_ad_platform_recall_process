# Core Review 意见回复

## 评审日期
2026年4月28日

---

## 问题1: 注册增加手机验证码

**用户疑问**: 注册时要求用户输入手机号码，暂时mock一个930108为正确的校验码，但需预留接口供后续替换。

**当前实现状态**:
- 当前注册接口 `POST /api/auth/register` 仅支持 username + password
- 用户表 (users) 暂无 phone 字段

**Review意见 - 采纳**:
✅ 支持此需求，建议按以下方案实现：

1. **数据库修改**: 用户表新增 `phone` 字段
```go
type User struct {
    Phone string `gorm:"type:varchar(20);comment:手机号" json:"phone"`
}
```

2. **新增接口**:
   - `POST /api/auth/send-code` - 发送验证码 (mock: 930108)
   - 注册时传入验证码进行校验

3. **Mock设计**:
```go
// service/sms.go
const MockRegisterCode = "930108"  // 可配置化

func (s *SMSService) SendRegisterCode(phone string) error {
    // 当前mock实现
    log.Printf("[SMS Mock] 发送注册验证码 %s 到 %s", MockRegisterCode, phone)
    return nil
}
```

4. **预留真实短信接口**:
```go
// 后续实现时可注入不同provider
type SMSProvider interface {
    SendCode(phone, templateID string) error
}
```

**工作量评估**: 0.5天

---

## 问题2: 忘记密码功能

**用户疑问**: 登录界面增加忘记密码选项，允许通过手机号获取验证码修改密码。mock验证码 931216。

**当前实现状态**:
- 当前无忘记密码功能
- 无验证码发送接口

**Review意见 - 采纳**:
✅ 支持此需求，建议按以下方案实现：

1. **新增接口**:
   - `POST /api/auth/send-reset-code` - 发送重置密码验证码 (mock: 931216)
   - `POST /api/auth/reset-password` - 重置密码（需传入手机号+验证码+新密码）

2. **Mock设计**:
```go
// service/sms.go
const MockResetCode = "931216"  // 可配置化

func (s *SMSService) SendResetCode(phone string) error {
    // 当前mock实现
    log.Printf("[SMS Mock] 发送重置密码验证码 %s 到 %s", MockResetCode, phone)
    return nil
}
```

3. **业务流程**:
```
用户输入手机号 → 发送验证码 → 用户输入验证码+新密码 → 验证通过 → 更新密码
```

4. **前端**:
- 登录页面增加"忘记密码"入口
- 跳转至重置密码页面

**工作量评估**: 0.5天

---

## 问题3: 修改密码功能

**用户疑问**: 登录后的账户管理界面允许修改密码。

**当前实现状态**:
- 当前无修改密码接口
- 前端 index.html 有Token管理，但无修改密码功能

**Review意见 - 采纳**:
✅ 支持此需求，需新增以下内容：

1. **新增接口**:
   - `POST /api/auth/change-password` (需鉴权)
   - Request: `{ "old_password": "xxx", "new_password": "xxx" }`

2. **Service层实现**:
```go
func (s *AuthService) ChangePassword(userID uint64, oldPwd, newPwd string) error {
    user, err := s.userRepo.FindByID(userID)
    if err != nil { return err }
    
    if !utils.CheckPassword(oldPwd, user.Password) {
        return ErrInvalidCredentials
    }
    
    hashedPwd, err := utils.HashPassword(newPwd)
    if err != nil { return err }
    
    return s.userRepo.UpdatePassword(userID, hashedPwd)
}
```

3. **前端**:
- index.html 账户管理区新增"修改密码"表单

**工作量评估**: 0.5天

---

## 问题4: 查询接口只返回最新一条

**用户疑问**: 查询接口应总是返回最新的一条数据，而不是全部记录。

**当前实现状态**:
```go
// repository/recall.go
func (r *RecallRepository) Query(params QueryParams) (*QueryResult, error) {
    // ... 支持分页，返回多条记录
}
```

当前 `/api/query` 接口支持分页查询，可返回多条记录。

**Review意见 - 部分采纳，建议保留分页**:
⚠️ 需要明确需求：

| 场景 | 建议方案 |
|------|----------|
| 仪表盘展示 | 新增 `GET /api/query/latest` 只返回最新1条 |
| 后台管理 | 保留当前分页查询接口 |

**工作量评估**: 0.25天

---

## 问题5: 新增历史数据接口

**用户疑问**: 新增一个历史数据接口，返回满足条件的全部记录。

**当前实现状态**:
- 当前 `/api/query` 支持按条件查询，但有分页限制
- 无"导出全部"功能

**Review意见 - 采纳**:
✅ 支持此需求，建议新增：

1. **新增接口**:
   - `GET /api/history` - 返回满足条件的全部记录（不分页或限制最大1000条）

2. **参数**:
   - recall_service_name (可选)
   - platform (可选)
   - user_name (可选)

3. **实现**:
```go
func (r *RecallRepository) QueryAll(params QueryParams) ([]model.RecallRecord, error) {
    query := r.db.Model(&model.RecallRecord{})
    
    if params.RecallServiceName != "" {
        query = query.Where("recall_service_name = ?", params.RecallServiceName)
    }
    // ... 其他条件
    
    var records []model.RecallRecord
    return records, query.Order("created_at DESC").Limit(1000).Find(&records).Error
}
```

**注意**: 需设置最大返回条数(如1000)防止数据量过大。

**工作量评估**: 0.25天

---

## 问题6: 多租户Token覆盖问题

**用户疑问**: config里的token是谁的token？多租户情况下token会不会互相覆盖？

**当前实现状态**:
```go
// config.yaml
token:
  secret: "your-secret-key-change-in-production-2026"  # 全局密钥
  expiry_hours: 24
```

**Review意见 - 明确说明**:
📌 当前设计为**多租户隔离**，Token不会互相覆盖：

| 组件 | 说明 |
|------|------|
| `token.secret` | **全局签名密钥** - 所有用户共用此密钥签名JWT |
| `tokens` 表 | **每用户独立存储** - 每个用户有独立的token记录 |
| JWT Payload | 包含 `user_id` 标识用户身份 |

**数据隔离机制**:
```
用户A登录 → 生成JWT(tokenA), 存入tokens表(user_id=A)
用户B登录 → 生成JWT(tokenB), 存入tokens表(user_id=B)
          ↓ (密钥相同，但内容不同)
    JWT签名不同，user_id不同，隔离正确
```

**潜在优化方向** (非必须):
- 如需更高隔离性，可考虑每个租户配置独立 secret
- 当前方案已满足常规多租户安全需求

**无需修改** ✅

---

## 汇总

| 序号 | 问题 | 采纳情况 | 工作量 |
|------|------|----------|--------|
| 1 | 注册手机验证码 | ✅ 采纳 | 0.5天 |
| 2 | 忘记密码功能 | ✅ 采纳 | 0.5天 |
| 3 | 修改密码功能 | ✅ 采纳 | 0.5天 |
| 4 | 查询返回最新1条 | ⚠️ 部分采纳 | 0.25天 |
| 5 | 历史数据接口 | ✅ 采纳 | 0.25天 |
| 6 | 多租户Token | ✅ 明确说明 | 0天 |

**新增工作量总计**: 约2天 (不影响核心Recall功能)

---

## 后续开发建议

### Phase 1 (当前已完成)
- 基础框架 ✅
- 认证模块 ✅
- Recall核心 ✅
- 查询通知 ✅

### Phase 2 (新增需求)
| 优先级 | 任务 | 工作量 |
|--------|------|--------|
| P1 | 短信验证码Mock框架 | 0.5天 |
| P1 | 注册+验证码流程 | 0.5天 |
| P2 | 忘记密码流程 | 0.5天 |
| P2 | 修改密码功能 | 0.5天 |
| P3 | 最新1条查询接口 | 0.25天 |
| P3 | 历史数据接口 | 0.25天 |

**建议**: 核心Recall功能稳定后，再推进Phase 2。
