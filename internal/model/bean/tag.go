package bean

type Tag struct {
	ID     uint   `gorm:"primaryKey"`     // 标签ID
	Tag    string `gorm:"column:tag"`     // 标签
	Style  string `gorm:"column:style"`   // 显示风格
	UserID uint   `gorm:"column:user_id"` // 用户ID
}
