package tag

import (
	"testing"

	"english-study/internal/utils"
)

// 纯函数权限矩阵测试; 不碰 DB / 不起 server, 只验证 canMutateTag 的真值表
func TestCanMutateTag(t *testing.T) {
	admin := &utils.UserInfo{ID: 0, Username: "admin", Role: utils.RoleAdmin}
	normal := &utils.UserInfo{ID: 1, Username: "alice", Role: utils.RoleNormal}
	other := &utils.UserInfo{ID: 2, Username: "bob", Role: utils.RoleNormal}

	// 关键回归: sssadmin 是 user_id=0 的真实超管, 她的私有标签(isSystem=false, ownerID=0)
	// 只能本人改; 其他超管/普通用户都不能(否则就是当年 user_id=0 撞系统哨兵的老 bug).
	adminUser0 := &utils.UserInfo{ID: 0, Username: "sssadmin", Role: utils.RoleAdmin}
	otherAdmin := &utils.UserInfo{ID: 5, Username: "boss", Role: utils.RoleAdmin}

	tests := []struct {
		name     string
		caller   *utils.UserInfo
		isSystem bool
		ownerID  uint
		want     bool
	}{
		// 系统标签 (isSystem=true)
		{"admin 可改系统标签", admin, true, 0, true},
		{"普通用户不可改系统标签", normal, true, 0, false},
		{"其他普通用户不可改系统标签", other, true, 0, false},

		// 用户标签 (alice 拥有, ownerID=1, 非系统)
		{"admin 不可改 alice 的标签 (非自己, 且不是系统)", admin, false, 1, false},
		{"alice 可改自己的标签", normal, false, 1, true},
		{"bob 不可改 alice 的标签", other, false, 1, false},

		// user_id=0 的真实用户私有标签 (isSystem=false, ownerID=0)
		{"user0 本人可改自己的私有标签", adminUser0, false, 0, true},
		{"其它超管不可改 user0 的私有标签", otherAdmin, false, 0, false},
		{"普通用户不可改 user0 的私有标签", normal, false, 0, false},

		// nil 守卫
		{"nil caller 一律拒(系统)", nil, true, 0, false},
		{"nil caller 用户标签也拒", nil, false, 1, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := canMutateTag(tt.caller, tt.isSystem, tt.ownerID)
			if got != tt.want {
				t.Errorf("canMutateTag(role=%v, isSystem=%v, ownerID=%d) got %v want %v",
					describeRole(tt.caller), tt.isSystem, tt.ownerID, got, tt.want)
			}
		})
	}
}

func describeRole(ui *utils.UserInfo) string {
	if ui == nil {
		return "nil"
	}
	if ui.IsAdmin() {
		return "admin"
	}
	return "normal"
}
