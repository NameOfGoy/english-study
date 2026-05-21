package bean

import "time"

// 单词学习状态. 复合唯一约束 (user_id, word_id, word_type) 防止并发 add 双插.
type WordStatus struct {
	ID       uint `gorm:"primaryKey"`
	WordID   uint `gorm:"column:word_id;uniqueIndex:idx_ws_user_word_type,priority:2;not null;comment:单词ID"`
	WordType int  `gorm:"column:word_type;uniqueIndex:idx_ws_user_word_type,priority:3;comment:单词类型, 1-单词 2-短语"`
	Status   int  `gorm:"column:status"`
	Times    int  `gorm:"column:times"`
	Weight   float64 `gorm:"column:weight"`
	StudyTime    time.Time `gorm:"column:study_time;comment:学习时间"`
	EaseFactor   float64   `gorm:"column:ease_factor;default:2.5;comment:难度因子(SM-2)"`
	Interval     int       `gorm:"column:interval;default:0;comment:当前复习间隔(天)"`
	NextReviewAt time.Time `gorm:"column:next_review_at;comment:下次复习时间"`
	Repetitions  int       `gorm:"column:repetitions;default:0;comment:连续正确次数"`
	UserID       uint      `gorm:"column:user_id;uniqueIndex:idx_ws_user_word_type,priority:1;not null;comment:用户ID"`
	SourceUserID uint      `gorm:"column:source_user_id;default:0;comment:来源用户ID(跨用户导入时记录), 0表示自建"`
	CreatedAt    time.Time `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt    time.Time `gorm:"column:updated_at;autoUpdateTime"`
}
