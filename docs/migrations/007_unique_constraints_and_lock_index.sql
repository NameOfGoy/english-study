-- 007: 唯一约束 + finish-trio FOR UPDATE 谓词索引
-- 来自第二轮 reviewer (DB-P0-1/2, DB-P1-4)
-- - word_statuses(user_id, word_id, word_type) 唯一约束: 防并发 add 双插重复状态行
-- - tags(user_id, tag) 唯一约束: 防同名重复标签
-- - word_statuses(user_id, word_id, word_type) 索引: finish-trio 的 SELECT FOR UPDATE 走点查
--
-- 注意: ADD CONSTRAINT 不支持 IF NOT EXISTS, 用 DO 块按需添加
-- CONCURRENTLY 不能在事务中执行, 索引部分独立运行

-- 1. word_statuses 复合唯一约束
DO $$
BEGIN
    -- 先清理已有的重复数据 (若有), 保留 id 最小的那条
    DELETE FROM word_statuses ws1
    USING word_statuses ws2
    WHERE ws1.id > ws2.id
      AND ws1.user_id = ws2.user_id
      AND ws1.word_id = ws2.word_id
      AND ws1.word_type = ws2.word_type;

    -- 添加约束
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint
        WHERE conname = 'idx_ws_user_word_type'
    ) THEN
        ALTER TABLE word_statuses
        ADD CONSTRAINT idx_ws_user_word_type
        UNIQUE (user_id, word_id, word_type);
        RAISE NOTICE 'Added unique constraint idx_ws_user_word_type';
    END IF;
END$$;

-- 2. tags 复合唯一约束 (注意: 系统默认标签 user_id=0, 之间不能同名)
DO $$
BEGIN
    -- 清理重复 (保留 id 最小的)
    DELETE FROM tags t1
    USING tags t2
    WHERE t1.id > t2.id
      AND t1.user_id = t2.user_id
      AND t1.tag = t2.tag;

    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint
        WHERE conname = 'idx_tag_user_name'
    ) THEN
        ALTER TABLE tags
        ADD CONSTRAINT idx_tag_user_name
        UNIQUE (user_id, tag);
        RAISE NOTICE 'Added unique constraint idx_tag_user_name';
    END IF;
END$$;

-- 3. word_statuses(user_id, word_id, word_type) 索引: 给 FOR UPDATE 用
--    上面的 UNIQUE 约束本身就会自动建索引, 这里其实是同一个; 留作记录
--    (实际不需要再单独建, 上面的 ADD CONSTRAINT 已经隐含 CREATE INDEX)
