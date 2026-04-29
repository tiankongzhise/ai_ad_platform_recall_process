1. [已修复] 用户注销以后，任然可以继续登录，存在严重的安全风险
   - 修复：Login 方法添加用户状态检查，status=0 的用户无法登录
   - 修复：FindByApiToken 添加 logout_at = -1 条件

2. [已修复] 用户注销后，新用户无法注册相同用户名
   - 修复方案：
     - 用户表新增 logout_at 字段（bigint，秒级时间戳）
     - 活跃用户 logout_at = -1
     - 注销用户 logout_at = 注销时的 Unix 时间戳
     - 建立 (user_name, logout_at) 联合唯一索引
     - 注册时查找活跃用户检查是否存在同名
     - 注销用户记录保留，新注册用户可使用相同用户名

**数据库迁移**（如已有数据库需要执行）：
```sql
ALTER TABLE users ADD COLUMN logout_at BIGINT DEFAULT -1 COMMENT '注销时间戳(秒)，-1表示活跃用户';
CREATE UNIQUE INDEX idx_user_name_logout ON users(user_name, logout_at);
UPDATE users SET logout_at = -1 WHERE status = 1;
UPDATE users SET logout_at = UNIX_TIMESTAMP() WHERE status = 0 AND logout_at = -1;
```
