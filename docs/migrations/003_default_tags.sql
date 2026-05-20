-- 插入系统默认标签 (user_id = 0)
-- 用户共享，不可编辑/删除
-- 幂等：先清掉所有 user_id=0 的旧默认标签再插入，避免重复

DELETE FROM tags WHERE user_id = 0;

INSERT INTO tags (tag, style, user_id) VALUES
  ('重点词汇', '#ff6b6b', 0),
  ('高频词',   '#4ecdc4', 0),
  ('易错词',   '#45b7d1', 0),
  ('考试词汇', '#96ceb4', 0),
  ('日常用词', '#feca57', 0),
  ('专业词汇', '#ff9ff3', 0),
  ('口语词汇', '#54a0ff', 0),
  ('写作词汇', '#5f27cd', 0);
