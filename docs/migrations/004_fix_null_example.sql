-- 修复 word_pos_user_* / word_phrase_user_* / word_pos / word_phrase 中
-- example 字段被错误存为字符串 'null' 的问题
-- 根因：types.WordPos.ExampleString() 调用 json.Marshal(nil) 输出 "null" 字符串
-- 已在代码层修复（v0.0.39+），此 SQL 回填历史脏数据
-- 幂等

DO $$
DECLARE
    r RECORD;
    updated_total INTEGER := 0;
    updated_count INTEGER;
BEGIN
    FOR r IN
        SELECT tablename FROM pg_tables
        WHERE schemaname = 'public'
          AND (
              tablename LIKE 'word_pos_user_%' OR tablename = 'word_pos' OR
              tablename LIKE 'word_phrase_user_%' OR tablename = 'word_phrase'
          )
    LOOP
        EXECUTE format('UPDATE %I SET example = '''' WHERE example = ''null''', r.tablename);
        GET DIAGNOSTICS updated_count = ROW_COUNT;
        IF updated_count > 0 THEN
            RAISE NOTICE '% : updated % rows', r.tablename, updated_count;
            updated_total := updated_total + updated_count;
        END IF;
    END LOOP;
    RAISE NOTICE 'TOTAL UPDATED: %', updated_total;
END$$;
