package bean

import "time"

type WordTag struct {
	ID        uint      `gorm:"primaryKey"`
	WordID    uint      `gorm:"column:word_id;index;not null;comment:单词ID"` // 单词ID
	WordType  int       `gorm:"column:word_type;comment:单词类型"`              // 单词类型, 1-单词 2-短语
	TagID     uint      `gorm:"column:tag_id;index;not null;comment:标签ID"`  // 标签ID
	UserID    uint      `gorm:"column:user_id;index;not null;comment:用户ID"` // 用户ID
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt time.Time `gorm:"column:updated_at;autoUpdateTime"`
}
