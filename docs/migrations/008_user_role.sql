-- 008: users 加 role 字段, 默认 0=普通; id=0 升超管
-- 来自 tag 系统改造设计 §1.1
-- 幂等

ALTER TABLE users ADD COLUMN IF NOT EXISTS role INT NOT NULL DEFAULT 0;
CREATE INDEX IF NOT EXISTS idx_users_role ON users(role);

-- 把 id=0 的用户升为超管 (设计 §9 确认: tags.user_id=0 已是系统标签占位归属, 同一身份)
UPDATE users SET role = 1 WHERE id = 0;
