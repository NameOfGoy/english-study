package bean

import "time"

// ImportTask 导入任务
type ImportTask struct {
	ID           uint      `gorm:"primaryKey"`
	UserID       uint      `gorm:"column:user_id;index;not null;comment:用户ID"`
	FileName     string    `gorm:"column:file_name;comment:原始文件名"`
	Status       int       `gorm:"column:status;default:0;comment:状态 0-待处理 1-进行中 2-已完成 3-失败"`
	Total        int       `gorm:"column:total;comment:总词数"`
	Current      int       `gorm:"column:current;default:0;comment:当前处理到第几个"`
	CurrentWord  string    `gorm:"column:current_word;comment:当前正在处理的单词"`
	SuccessCount int       `gorm:"column:success_count;default:0;comment:成功数"`
	FailCount    int       `gorm:"column:fail_count;default:0;comment:失败数"`
	FailWords    string    `gorm:"column:fail_words;type:text;comment:失败的词列表JSON"`
	CreatedAt    time.Time `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt    time.Time `gorm:"column:updated_at;autoUpdateTime"`
}
