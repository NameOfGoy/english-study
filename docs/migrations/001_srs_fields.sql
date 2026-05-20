-- 001_srs_fields.sql
-- Run AFTER deploying new code (GORM AutoMigrate adds columns with defaults first).
-- Backfills SRS params for existing Review/Strengthen/Finish records.

-- Review + Strengthen: estimate SRS params from existing times
UPDATE word_statuses SET
    ease_factor = 2.5,
    repetitions = times,
    interval = CASE
        WHEN times <= 0 THEN 1
        WHEN times = 1 THEN 1
        WHEN times = 2 THEN 6
        ELSE LEAST(CAST(POWER(2.5, times - 2) * 6 AS INTEGER), 180)
    END,
    next_review_at = CASE
        WHEN study_time IS NOT NULL AND study_time > '2000-01-01' THEN
            study_time + INTERVAL '1 day' * CASE
                WHEN times <= 1 THEN 1
                WHEN times = 2 THEN 6
                ELSE LEAST(CAST(POWER(2.5, times - 2) * 6 AS INTEGER), 180)
            END
        ELSE NOW()
    END
WHERE status IN (2, 3);

-- Finish: set a far-future next_review_at
UPDATE word_statuses SET
    ease_factor = 2.5,
    repetitions = GREATEST(times, 4),
    interval = 90,
    next_review_at = NOW() + INTERVAL '90 days'
WHERE status = 4;

-- Study: set default SRS params (not yet in SRS cycle)
UPDATE word_statuses SET
    ease_factor = 2.5,
    repetitions = 0,
    interval = 0,
    next_review_at = '0001-01-01'
WHERE status = 1;

-- Index for SRS queries
CREATE INDEX IF NOT EXISTS idx_word_statuses_next_review
    ON word_statuses(user_id, status, next_review_at);
