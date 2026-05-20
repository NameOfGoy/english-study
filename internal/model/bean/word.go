package bean

import (
	"fmt"
	"time"
)

// 单词
type Word struct {
	ID                         uint       `gorm:"primaryKey"`
	Word                       string     `gorm:"column:word;unique;not null;comment:单词全拼"`
	AmericanPronunciation      string     `gorm:"column:american_pronunciation;not null;comment:美式发音"`
	AmericanPronunciationAudio string     `gorm:"column:american_pronunciation_audio;not null;comment:美式发音音频"`
	BritishPronunciation       string     `gorm:"column:british_pronunciation;not null;comment:英式发音"`
	BritishPronunciationAudio  string     `gorm:"column:british_pronunciation_audio;not null;comment:英式发音音频"`
	Source                     string     `gorm:"column:source;default:'stardict';comment:来源 stardict|ai"`
	Pos                        []*WordPos `gorm:"-"`
	CreatedAt                  time.Time  `gorm:"column:created_at;autoCreateTime"`
}

func (w *Word) TableName() string {
	return "word"
}

func (w *Word) UserTableName(userId *uint) string {
	if userId == nil {
		return w.TableName()
	}
	return fmt.Sprintf("word_user_%d", *userId)
}

// 单词词性
type WordPos struct {
	ID          uint      `gorm:"primaryKey"` // 自定义的词性, id 要从100W开始
	WordID      uint      `gorm:"column:word_id;index;not null;comment:单词ID"`
	Word        string    `gorm:"column:word;index;not null;comment:单词"`
	Pos         int       `gorm:"column:pos;not null;comment:词性"`           // 动词、名词等
	Translation string    `gorm:"column:translation;not null;comment:中文翻译"` // 中文翻译
	Example     string    `gorm:"column:example;not null;comment:例句"`       // 例句, 格式为["",""], 元素用\n分割例句和中文翻译
	Picture     string    `gorm:"column:picture;not null;comment:图片"`       // 图片
	Exchange    string    `gorm:"column:exchange;not null;comment:变化形式"`    // 格式为["",""], 元素用:分割变化类型和变化后的单词
	UpdatedAt   time.Time `gorm:"column:updated_at;autoUpdateTime"`
}

func (w *WordPos) TableName() string {
	return "word_pos"
}

func (w *WordPos) UserTableName(userId *uint) string {
	if userId == nil {
		return w.TableName()
	}
	return fmt.Sprintf("word_pos_user_%d", *userId)
}
