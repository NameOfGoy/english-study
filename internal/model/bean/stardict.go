package bean

import (
	"strings"
	"time"

	"gorm.io/gorm"
)

// Word 词典单词结构体
type StarDict struct {
	ID          uint           `gorm:"primaryKey;autoIncrement;not null;unique;comment:主键ID" json:"id"`
	Word        string         `gorm:"column:word;type:varchar(255);not null;unique;comment:单词" json:"word"`
	Sw          string         `gorm:"column:sw;type:varchar(255);not null;comment:处理后的单词用于搜索" json:"sw"`
	Phonetic    string         `gorm:"column:phonetic;type:varchar(255);comment:音标" json:"phonetic"`
	Definition  string         `gorm:"column:definition;type:text;comment:英文定义" json:"definition"`
	Translation string         `gorm:"column:translation;type:text;comment:中文翻译" json:"translation"`
	Pos         string         `gorm:"column:pos;type:varchar(64);comment:词性" json:"pos"`
	Collins     int            `gorm:"column:collins;type:int;default:0;comment:柯林斯词典等级" json:"collins"`
	Oxford      int            `gorm:"column:oxford;type:int;default:0;comment:牛津词典等级" json:"oxford"`
	Tag         string         `gorm:"column:tag;type:varchar(255);comment:标签" json:"tag"`
	Bnc         int            `gorm:"column:bnc;type:int;comment:BNC频率" json:"bnc"`
	Frq         int            `gorm:"column:frq;type:int;comment:词频" json:"frq"`
	Exchange    string         `gorm:"column:exchange;type:text;comment:词形变化" json:"exchange"`
	Detail      string         `gorm:"column:detail;type:text;comment:详细信息JSON" json:"detail"`
	Audio       string         `gorm:"column:audio;type:text;comment:音频文件路径" json:"audio"`
	CreatedAt   time.Time      `gorm:"autoCreateTime;comment:创建时间" json:"created_at"`
	UpdatedAt   time.Time      `gorm:"autoUpdateTime;comment:更新时间" json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index;comment:软删除时间" json:"deleted_at"`
}

// TableName 指定表名
func (StarDict) TableName() string {
	return "stardict"
}

// GetPos 获取词性, 1维是词性简写, 2维是释义
func (s StarDict) GetPos() [][2]string {
	ss := strings.Split(s.Translation, `\n`)
	pos := make([][2]string, 0)
	for _, v := range ss {
		// 取得s中第一个.的索引
		idx := strings.Index(v, `.`)
		if idx == -1 {
			continue
		}
		// +1是指向下一个空格位, 再+1则指向释义的首位
		idx += 1 + 1
		// 词性简写
		psw := v[:idx-1] // 不要空格位
		// 赋值
		pos = append(pos, [2]string{psw, v[idx:]})
	}
	return pos
}

// GetExchange 获取词形变化
func (s StarDict) GetExchange() map[string]string {
	if s.Exchange == "" {
		return nil
	}
	exchanges := make(map[string]string)
	for _, v := range strings.Split(s.Exchange, `/`) {
		idx := strings.Index(v, `:`)
		if idx == -1 {
			continue
		}
		exchanges[v[:idx]] = v[idx+1:]
	}
	return exchanges
}

func (s StarDict) IsValid() bool {
	return s.Word != "" && s.Phonetic != "" && s.Definition != "" && s.Translation != ""
}
