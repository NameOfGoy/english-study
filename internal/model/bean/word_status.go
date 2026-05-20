package bean

import "time"

// 单词标签
type WordStatus struct {
	ID        uint      `gorm:"primaryKey"`
	WordID    uint      `gorm:"column:word_id;index;not null;comment:单词ID"` // 单词ID
	WordType  int       `gorm:"column:word_type;comment:单词类型"`              // 单词类型, 1-单词 2-短语
	Status    int       `gorm:"column:status"`                              // 状态
	Times     int       `gorm:"column:times"`                               // 学习次数
	Weight    float64   `gorm:"column:weight"`                              // 权重
	StudyTime    time.Time `gorm:"column:study_time;comment:学习时间"`                      // 学习时间
	EaseFactor   float64   `gorm:"column:ease_factor;default:2.5;comment:难度因子(SM-2)"`   // SM-2难度因子
	Interval     int       `gorm:"column:interval;default:0;comment:当前复习间隔(天)"`        // 复习间隔天数
	NextReviewAt time.Time `gorm:"column:next_review_at;comment:下次复习时间"`               // 下次复习时间
	Repetitions  int       `gorm:"column:repetitions;default:0;comment:连续正确次数"`        // 连续正确次数
	UserID       uint      `gorm:"column:user_id;index;not null;comment:用户ID"`
	SourceUserID uint      `gorm:"column:source_user_id;default:0;comment:来源用户ID(跨用户导入时记录), 0表示自建"`
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt time.Time `gorm:"column:updated_at;autoUpdateTime"`
}
