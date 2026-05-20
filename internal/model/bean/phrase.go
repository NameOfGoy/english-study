package bean

import "fmt"

type WordPhrase struct {
	ID            uint   `gorm:"primaryKey"`                               // id
	Phrase        string `gorm:"column:phrase;unique;not null;comment:短语"` // 短语
	Translation   string `gorm:"column:translation;comment:中文翻译"`          // 中文翻译
	Pronunciation string `gorm:"column:pronunciation;comment:发音"`          // 发音
	Example       string `gorm:"column:example;comment:例句"`                // 例句, 格式为["",""], 元素用\n分割例句和中文翻译
	Picture       string `gorm:"column:picture;comment:图片"`                // 图片
}

func (WordPhrase) TableName() string {
	return "word_phrase"
}

func (WordPhrase) UserTableName(userId *uint) string {
	if userId == nil {
		return WordPhrase{}.TableName()
	}
	return fmt.Sprintf("word_phrase_user_%d", *userId)
}
