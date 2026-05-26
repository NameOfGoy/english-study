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

	tests := []struct {
		name    string
		caller  *utils.UserInfo
		ownerID uint
		want    bool
	}{
		// 系统标签
		{"admin 可改系统标签", admin, 0, true},
		{"普通用户不可改系统标签", normal, 0, false},
		{"其他普通用户不可改系统标签", other, 0, false},

		// 用户标签 (alice 拥有, ownerID=1)
		{"admin 不可改 alice 的标签 (非自己, 且不是系统)", admin, 1, false},
		{"alice 可改自己的标签", normal, 1, true},
		{"bob 不可改 alice 的标签", other, 1, false},

		// nil 守卫
		{"nil caller 一律拒", nil, 0, false},
		{"nil caller 用户标签也拒", nil, 1, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := canMutateTag(tt.caller, tt.ownerID)
			if got != tt.want {
				t.Errorf("canMutateTag(role=%v, ownerID=%d) got %v want %v",
					describeRole(tt.caller), tt.ownerID, got, tt.want)
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
