package dictionary

import (
	"reflect"
	"testing"
)

func TestParseImportLines(t *testing.T) {
	tests := []struct {
		name string
		in   []string
		want []ImportItem
	}{
		{
			name: "纯单词无标签",
			in:   []string{"apple", "banana"},
			want: []ImportItem{
				{Word: "apple"},
				{Word: "banana"},
			},
		},
		{
			name: "段标签 + 词条 + 终止符 + 词条",
			in:   []string{"[重点词汇]", "cat", "dog", "[---]", "frog"},
			want: []ImportItem{
				{Word: "cat", TagNames: []string{"重点词汇"}},
				{Word: "dog", TagNames: []string{"重点词汇"}},
				{Word: "frog"},
			},
		},
		{
			name: "标签切换 (隐式结束前一个段)",
			in:   []string{"[A]", "x", "[B]", "y"},
			want: []ImportItem{
				{Word: "x", TagNames: []string{"A"}},
				{Word: "y", TagNames: []string{"B"}},
			},
		},
		{
			name: "空行被跳过",
			in:   []string{"", "apple", "  ", "banana"},
			want: []ImportItem{
				{Word: "apple"},
				{Word: "banana"},
			},
		},
		{
			name: "行首尾空格被 trim, 词条本身不变",
			in:   []string{"  apple  ", "  [tag1]  ", "  cat  "},
			want: []ImportItem{
				{Word: "apple"},
				{Word: "cat", TagNames: []string{"tag1"}},
			},
		},
		{
			name: "短语 (内含空格) 当作普通词条",
			in:   []string{"[考试词汇]", "look up", "make it through"},
			want: []ImportItem{
				{Word: "look up", TagNames: []string{"考试词汇"}},
				{Word: "make it through", TagNames: []string{"考试词汇"}},
			},
		},
		{
			name: "[---] 之外的破折号变体不算终止符, 当标签名",
			in:   []string{"[---]", "a", "[--]", "b", "[----]", "c"},
			want: []ImportItem{
				{Word: "a"},
				{Word: "b", TagNames: []string{"--"}},
				{Word: "c", TagNames: []string{"----"}},
			},
		},
		{
			name: "[]空标签忽略 (不算 marker)",
			in:   []string{"[]", "apple"},
			want: []ImportItem{
				{Word: "[]"},
				{Word: "apple"},
			},
		},
		{
			name: "未闭合的 [ 不算 marker",
			in:   []string{"[unfinished", "apple"},
			want: []ImportItem{
				{Word: "[unfinished"},
				{Word: "apple"},
			},
		},
		{
			name: "连续两个 marker (中间无词条) 也合法; 后者覆盖前者",
			in:   []string{"[A]", "[B]", "x"},
			want: []ImportItem{
				{Word: "x", TagNames: []string{"B"}},
			},
		},
		{
			name: "空输入",
			in:   []string{},
			want: []ImportItem{},
		},
		{
			name: "全是 marker 没词条",
			in:   []string{"[a]", "[b]", "[---]"},
			want: []ImportItem{},
		},

		// 新: 内联多标签
		{
			name: "单词带 1 个内联标签",
			in:   []string{"apple [t1]"},
			want: []ImportItem{
				{Word: "apple", TagNames: []string{"t1"}},
			},
		},
		{
			name: "单词带多个内联标签 (带空格)",
			in:   []string{"apple [t1] [t2] [t3]"},
			want: []ImportItem{
				{Word: "apple", TagNames: []string{"t1", "t2", "t3"}},
			},
		},
		{
			name: "单词带多个内联标签 (无空格紧贴)",
			in:   []string{"apple[t1][t2]"},
			want: []ImportItem{
				{Word: "apple", TagNames: []string{"t1", "t2"}},
			},
		},
		{
			name: "短语带内联标签",
			in:   []string{"look up [重点] [考试]"},
			want: []ImportItem{
				{Word: "look up", TagNames: []string{"重点", "考试"}},
			},
		},
		{
			name: "段标签 + 内联标签合并, 段标签在前, 去重",
			in: []string{
				"[段tag]",
				"apple [内联1] [内联2]",
				"banana [段tag] [独有]", // 内联里和段重名, 去重保留段位置
			},
			want: []ImportItem{
				{Word: "apple", TagNames: []string{"段tag", "内联1", "内联2"}},
				{Word: "banana", TagNames: []string{"段tag", "独有"}},
			},
		},
		{
			name: "段标签清空后, 词条还能带内联标签",
			in: []string{
				"[重点]",
				"a",
				"[---]",
				"b [仅内联]",
			},
			want: []ImportItem{
				{Word: "a", TagNames: []string{"重点"}},
				{Word: "b", TagNames: []string{"仅内联"}},
			},
		},
		{
			name: "内联里同一标签写两次, 去重",
			in:   []string{"apple [t] [t]"},
			want: []ImportItem{
				{Word: "apple", TagNames: []string{"t"}},
			},
		},
		{
			name: "未闭合的 [bad 不剥, 留在词条里",
			in:   []string{"apple [bad"},
			want: []ImportItem{
				{Word: "apple [bad"},
			},
		},
		{
			name: "内嵌 ] 字符不算 marker (例: 'apple ]extra]' 不剥)",
			in:   []string{"apple ]extra]"},
			want: []ImportItem{
				{Word: "apple ]extra]"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParseImportLines(tt.in)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseImportLines mismatch\n got:  %+v\n want: %+v", got, tt.want)
			}
		})
	}
}

func TestPeelInlineTags(t *testing.T) {
	tests := []struct {
		in       string
		wantWord string
		wantTags []string
	}{
		{"apple", "apple", nil},
		{"apple [t1]", "apple", []string{"t1"}},
		{"apple [t1] [t2]", "apple", []string{"t1", "t2"}},
		{"apple[t1][t2]", "apple", []string{"t1", "t2"}},
		{"look up [重点]", "look up", []string{"重点"}},
		{"[only]", "", []string{"only"}},
		{"[a][b]", "", []string{"a", "b"}},
		{"apple [bad", "apple [bad", nil},
		{"apple ]extra]", "apple ]extra]", nil},
		{"apple []", "apple []", nil},
	}
	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			w, ts := peelInlineTags(tt.in)
			if w != tt.wantWord {
				t.Errorf("word got %q want %q", w, tt.wantWord)
			}
			if !reflect.DeepEqual(ts, tt.wantTags) {
				t.Errorf("tags got %v want %v", ts, tt.wantTags)
			}
		})
	}
}

func TestParseMarker(t *testing.T) {
	tests := []struct {
		in       string
		wantName string
		wantOK   bool
	}{
		{"[abc]", "abc", true},
		{"[a]", "a", true},
		{"[]", "", false},
		{"[abc", "", false},
		{"abc]", "", false},
		{"a", "", false},
		{"", "", false},
		{"[ ]", " ", true},
		{"[---]", "---", true},
		{"[a]b]", "", false}, // 内部含 ] 不算 marker
	}
	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			name, ok := parseMarker(tt.in)
			if ok != tt.wantOK || name != tt.wantName {
				t.Errorf("parseMarker(%q) = (%q, %v); want (%q, %v)", tt.in, name, ok, tt.wantName, tt.wantOK)
			}
		})
	}
}
