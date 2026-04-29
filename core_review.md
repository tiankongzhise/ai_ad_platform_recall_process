# Core Review 落地状态

## 评审项与当前代码状态

| 评审项 | 当前状态 | 对应接口/能力 |
|---|---|---|
| 注册增加手机验证码 | 已完成 | `POST /api/auth/send-code`、`POST /api/auth/register` |
| 忘记密码流程 | 已完成 | `POST /api/auth/send-reset-code`、`POST /api/auth/reset-password` |
| 登录后修改密码 | 已完成 | `POST /api/account/change-password` |
| 查询最新一条 | 已完成 | `GET /api/query/latest` |
| 历史数据查询 | 已完成 | `GET /api/history`（最大 1000 条） |
| 多租户 Token 隔离说明 | 已明确 | JWT 含 `user_id`，并按用户校验与过滤 |

## 当前接口命名对齐说明

- 账户相关接口已统一在 `/api/account/*` 路径下
- Query 条件以 `platform`、`user_tag` 为主，`user_name` 仅作兼容参数
- Recall `state` 格式为 `uid_platform_user_tag`，不再使用 `recall_service_name` 风格

## 验证码实现说明

- 注册验证码 mock：`930108`
- 重置验证码 mock：`931216`
- 验证码有效期：5 分钟（内存保存）

## 结论

本文件对应的 review 项已全部在代码中落地。完整参数类型与返回结构见 `接口文档.md`。
