package share

import "testing"

// 这些常量是 API 契约 (与前端约定 0/1/2). 不能随手改值, 否则前端发的 mode=1 会被后端误读.
func TestTagImportModeConstants_StableValues(t *testing.T) {
	cases := []struct {
		name string
		got  int
		want int
	}{
		{"TagImportNone", TagImportNone, 0},
		{"TagImportSystemOnly", TagImportSystemOnly, 1},
		{"TagImportAll", TagImportAll, 2},
	}
	for _, c := range cases {
		if c.got != c.want {
			t.Errorf("%s = %d, want %d (API 契约改动需要同步前端 ShareImportModal)", c.name, c.got, c.want)
		}
	}
}
