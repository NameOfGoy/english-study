package bean

// Tag 用户标签. 同一用户下 tag 文本必须唯一, 防止并发新增同名.
// 系统标签由 IsSystem=true 标识(对所有用户可见), 不再用 user_id=0 作为哨兵——
// 因为真实用户(如 sssadmin)的 user_id 也可能是 0, 会与"系统标签"撞车导致私有标签外泄.
type Tag struct {
	ID       uint   `gorm:"primaryKey"`
	Tag      string `gorm:"column:tag;uniqueIndex:idx_tag_user_name,priority:2"`
	Style    string `gorm:"column:style"`
	UserID   uint   `gorm:"column:user_id;uniqueIndex:idx_tag_user_name,priority:1"`
	IsSystem bool   `gorm:"column:is_system;not null;default:false;index;comment:是否系统标签(对所有用户可见)"`
}
