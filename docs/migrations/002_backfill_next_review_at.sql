-- 回填 SRS 上线前遗留的 next_review_at NULL 数据
-- 影响：复习模式 (status=2) 和强化模式 (status=3) 中 next_review_at 为 NULL 的行
-- 行为：把这些词标记为"立刻可复习/可强化"
-- 一次性执行，可重复运行（幂等）

UPDATE word_statuses
SET next_review_at = NOW()
WHERE status IN (2, 3)
  AND next_review_at IS NULL;
