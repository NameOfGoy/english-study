package bean

// Tag 用户标签. 同一用户下 tag 文本必须唯一, 防止并发新增同名.
type Tag struct {
	ID     uint   `gorm:"primaryKey"`
	Tag    string `gorm:"column:tag;uniqueIndex:idx_tag_user_name,priority:2"`
	Style  string `gorm:"column:style"`
	UserID uint   `gorm:"column:user_id;uniqueIndex:idx_tag_user_name,priority:1"`
}
