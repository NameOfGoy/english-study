package bean

import "time"

type User struct {
	ID        uint      `gorm:"primaryKey;autoIncrement;comment:主键ID"`              // 主键ID
	Account   string    `gorm:"type:varchar(100);unique;not null;comment:账号"`       // 账号
	WxOpenID  string    `gorm:"type:varchar(100);unique;not null;comment:微信openid"` // 微信openid
	Username  string    `gorm:"type:varchar(100);not null;comment:用户名"`             // 用户名
	Password  string    `gorm:"type:varchar(255);not null;comment:密码"`              // 密码
	Phone     string    `gorm:"type:varchar(20);comment:手机"`                        // 手机
	Email     string    `gorm:"type:varchar(100);comment:邮箱"`                       // 邮箱
	Avatar    string    `gorm:"type:varchar(255);comment:头像"`                       // 头像
	LastLogin time.Time `gorm:"comment:最后登录时间"`                                     // 最后登录时间
	CreatedAt time.Time `gorm:"autoCreateTime;comment:创建时间"`                        // 创建时间
	UpdatedAt time.Time `gorm:"autoUpdateTime;comment:更新时间"`                        // 更新时间
}
