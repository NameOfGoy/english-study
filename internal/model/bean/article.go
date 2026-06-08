package bean

import "time"

// Article 收录的短篇文章. 全局表 + user_id (与 word_statuses/import_tasks 一致, 非 per-user 动态表).
// body 用 jsonb 整体存双语正文 + 高亮信息, 不按内部字段查询; 标签不入库, 查询时由 article_words 实时算.
type Article struct {
	ID      uint   `gorm:"primaryKey"`
	UserID  uint   `gorm:"column:user_id;index:idx_article_user_created,priority:1;not null;comment:用户ID"`
	TitleEn string `gorm:"column:title_en;type:varchar(512);comment:英文标题"`
	TitleZh string `gorm:"column:title_zh;type:varchar(512);comment:中文标题"`
	// Body 双语正文 JSON: {"sentences":[{"en":"","zh":""}],"used_words":[{"word":"","type":1,"surfaces":["",...]}]}
	Body      string    `gorm:"column:body;type:jsonb;comment:双语正文+高亮信息JSON"`
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime;index:idx_article_user_created,priority:2"`
	UpdatedAt time.Time `gorm:"column:updated_at;autoUpdateTime"`
}

func (Article) TableName() string { return "article" }

// ArticleWord 文章包含的词条 (规范化, 支撑: 含词搜索 / 标签并集 / 删除兜底显示).
// 复合唯一 (article_id, word_id, word_type) 去重; word_id 用于 join word_tags (词可能已删),
// word_text 为原文快照, 删词后仍可展示并支撑英文含词搜索.
type ArticleWord struct {
	ID        uint   `gorm:"primaryKey"`
	ArticleID uint   `gorm:"column:article_id;uniqueIndex:idx_aw_article_word,priority:1;index:idx_aw_article;not null;comment:文章ID"`
	UserID    uint   `gorm:"column:user_id;index:idx_aw_user;not null;comment:用户ID(冗余,加速跨文章过滤)"`
	WordID    uint   `gorm:"column:word_id;uniqueIndex:idx_aw_article_word,priority:2;comment:词条ID(可能已删,仅join word_tags)"`
	WordType  int    `gorm:"column:word_type;uniqueIndex:idx_aw_article_word,priority:3;comment:1-单词 2-短语"`
	WordText  string `gorm:"column:word_text;type:varchar(256);index:idx_aw_word_text;not null;comment:词条原文(删除兜底+英文含词搜索)"`
}

func (ArticleWord) TableName() string { return "article_words" }
