-- 009: 给练习页全局标签筛选热路径加索引
-- 查询: SELECT DISTINCT word_id, word_type FROM word_tags WHERE user_id=? AND tag_id IN (?)
-- 已有索引: user_id (single col) / tag_id (single col) 都不能高效覆盖这个组合.
-- 加一个 (user_id, tag_id) 复合索引, INCLUDE word_id/word_type 以索引覆盖, 避免回表.
-- 幂等

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_word_tags_filter
    ON word_tags (user_id, tag_id)
    INCLUDE (word_id, word_type);
