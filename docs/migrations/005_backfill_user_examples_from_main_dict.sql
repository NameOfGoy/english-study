-- 从主词典 word_pos 表回填所有用户 word_pos_user_* 表中 example 为空的行
-- 根因：异步例句生成只 UPDATE 主词典，不会同步到用户表
-- 已在代码层修复（v0.0.40+），此 SQL 修复历史脏数据
-- 注意：只覆盖空的 example，不会覆盖用户已编辑的内容
-- 幂等

DO $$
DECLARE
    r RECORD;
    updated_total INTEGER := 0;
    updated_count INTEGER;
BEGIN
    FOR r IN
        SELECT tablename FROM pg_tables
        WHERE schemaname = 'public' AND tablename LIKE 'word_pos_user_%'
    LOOP
        EXECUTE format(
            'UPDATE %I u SET example = wp.example FROM word_pos wp ' ||
            'WHERE u.word = wp.word AND u.pos = wp.pos ' ||
            '  AND (u.example = '''' OR u.example IS NULL OR u.example = ''null'') ' ||
            '  AND wp.example IS NOT NULL AND wp.example != '''' AND wp.example != ''null''',
            r.tablename
        );
        GET DIAGNOSTICS updated_count = ROW_COUNT;
        IF updated_count > 0 THEN
            RAISE NOTICE '% : backfilled % rows', r.tablename, updated_count;
            updated_total := updated_total + updated_count;
        END IF;
    END LOOP;
    RAISE NOTICE 'TOTAL BACKFILLED: %', updated_total;
END$$;
