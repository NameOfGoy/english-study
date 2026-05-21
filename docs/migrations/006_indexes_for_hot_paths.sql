-- 006: 热路径索引补强 (来自 database-reviewer 多项 P0/P1 发现)
-- 包含:
--   1. pg_trgm GIN 索引: stardict 中文模糊搜索 (LIKE '%关键词%') 不再全表扫 40 万行
--   2. word_statuses 复合索引扩展: 加入 word_type 列, 让 review/strength 卡片查询走索引
--   3. word_statuses(user_id, study_time) 索引: dashboard 今日完成数聚合走索引
--   4. import_tasks(user_id, created_at DESC) 复合索引: 历史按时间倒序检索
--   5. word_pos(word, pos) 复合索引: 迁移 005 类的回填脚本, 主词典与用户表关联
--
-- 所有索引用 CONCURRENTLY 创建, 不锁表, 失败可重试.
-- 注意: CONCURRENTLY 不能在事务中执行, 此脚本需要在 psql 命令行中独立执行,
--      不要 wrap 在 BEGIN/COMMIT 里.

-- 1. pg_trgm 扩展 + stardict.translation GIN 索引
CREATE EXTENSION IF NOT EXISTS pg_trgm;

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_stardict_translation_trgm
    ON stardict USING GIN (translation gin_trgm_ops);

-- 2. word_statuses: review/strength 卡片查询的复合索引 (user_id, status, word_type, next_review_at)
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_word_statuses_srs_type
    ON word_statuses (user_id, status, word_type, next_review_at);

-- 3. word_statuses: dashboard 今日完成聚合 (user_id, study_time)
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_word_statuses_user_studytime
    ON word_statuses (user_id, study_time);

-- 4. import_tasks: 历史按时间倒序检索
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_import_tasks_user_created
    ON import_tasks (user_id, created_at DESC);

-- 5. word_pos: 主词典按 (word, pos) 关联回填用户表
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_word_pos_word_pos
    ON word_pos (word, pos);

-- 验证: 列出所有索引大小, 确认创建成功
-- SELECT indexname, pg_size_pretty(pg_relation_size(indexname::regclass)) AS size
-- FROM pg_indexes
-- WHERE indexname IN (
--     'idx_stardict_translation_trgm',
--     'idx_word_statuses_srs_type',
--     'idx_word_statuses_user_studytime',
--     'idx_import_tasks_user_created',
--     'idx_word_pos_word_pos'
-- )
-- ORDER BY indexname;
