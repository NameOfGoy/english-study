package model

import (
	"context"
	"os"
	"testing"

	"english-study/internal/model/bean"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// article 的查询用到 Postgres 专有语法(ILIKE / jsonb / 多表 join), sqlite 无法替身.
// 故本测试为"集成测试", 仅在设置 ARTICLE_TEST_DSN 时运行; 否则跳过.
//   go test ./internal/model -run TestArticle  (需 ARTICLE_TEST_DSN=host=... user=... ...)
func openArticleTestDB(t *testing.T) *Model {
	t.Helper()
	dsn := os.Getenv("ARTICLE_TEST_DSN")
	if dsn == "" {
		t.Skip("未设置 ARTICLE_TEST_DSN, 跳过 article 集成测试")
	}
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("连接测试库失败: %v", err)
	}
	if err := db.AutoMigrate(&bean.Article{}, &bean.ArticleWord{}, &bean.WordTag{}, &bean.Tag{}); err != nil {
		t.Fatalf("AutoMigrate 失败: %v", err)
	}
	return &Model{DB: db}
}

func TestArticleListAndTagUnion(t *testing.T) {
	m := openArticleTestDB(t)
	ctx := context.Background()
	const uid = 990001 // 用一个不太可能撞车的测试用户

	// 清理上次残留
	m.DB.Where("user_id = ?", uid).Delete(&bean.Article{})
	m.DB.Where("user_id = ?", uid).Delete(&bean.ArticleWord{})
	m.DB.Where("user_id = ?", uid).Delete(&bean.WordTag{})
	m.DB.Where("user_id = ?", uid).Delete(&bean.Tag{})

	// 两个标签
	tagA := bean.Tag{Tag: "动物", Style: "#f00", UserID: uid}
	tagB := bean.Tag{Tag: "动作", Style: "#0f0", UserID: uid}
	if err := m.DB.Create(&tagA).Error; err != nil {
		t.Fatal(err)
	}
	if err := m.DB.Create(&tagB).Error; err != nil {
		t.Fatal(err)
	}
	// 词 cat(单词,id=101)->动物, run(单词,id=102)->动作
	m.DB.Create(&bean.WordTag{WordID: 101, WordType: 1, TagID: tagA.ID, UserID: uid})
	m.DB.Create(&bean.WordTag{WordID: 102, WordType: 1, TagID: tagB.ID, UserID: uid})

	// 文章: 含 cat / run
	art := &bean.Article{
		UserID:  uid,
		TitleEn: "The Cat Who Loves To Run",
		TitleZh: "爱奔跑的猫",
		Body:    `{"sentences":[{"en":"The cat likes to run.","zh":"猫喜欢跑。"}],"used_words":[]}`,
	}
	words := []bean.ArticleWord{
		{WordID: 101, WordType: 1, WordText: "cat"},
		{WordID: 102, WordType: 1, WordText: "run"},
	}
	id, err := m.CreateArticleWithWords(ctx, art, words)
	if err != nil || id == 0 {
		t.Fatalf("创建文章失败: id=%d err=%v", id, err)
	}

	// 标题(英)搜索
	if items, total, err := m.ListArticles(ctx, uid, "cat", nil, 0, 10); err != nil || total != 1 || len(items) != 1 {
		t.Fatalf("标题英文搜索期望命中1篇, 实际 total=%d len=%d err=%v", total, len(items), err)
	}
	// 标题(中)搜索
	if _, total, _ := m.ListArticles(ctx, uid, "奔跑", nil, 0, 10); total != 1 {
		t.Fatalf("标题中文搜索期望命中1篇, 实际 %d", total)
	}
	// 含词(英文)搜索
	if _, total, _ := m.ListArticles(ctx, uid, "run", nil, 0, 10); total != 1 {
		t.Fatalf("含词搜索期望命中1篇, 实际 %d", total)
	}
	// 标签搜索(选 动物)
	if _, total, _ := m.ListArticles(ctx, uid, "", []uint{tagA.ID}, 0, 10); total != 1 {
		t.Fatalf("标签搜索期望命中1篇, 实际 %d", total)
	}
	// 不存在的词
	if _, total, _ := m.ListArticles(ctx, uid, "zzzz", nil, 0, 10); total != 0 {
		t.Fatalf("无关键词期望0篇, 实际 %d", total)
	}

	// 标签并集(批量)
	tagMap, err := m.ArticleTagsByArticleIDs(ctx, uid, []uint{id})
	if err != nil || len(tagMap[id]) != 2 {
		t.Fatalf("标签并集期望2个, 实际 %d err=%v", len(tagMap[id]), err)
	}
	// 标签并集(按词对, 未入库路径)
	pairs := []WordPair{{WordID: 101, WordType: 1}, {WordID: 102, WordType: 1}}
	if tb, err := m.TagsForWordPairs(ctx, uid, pairs); err != nil || len(tb) != 2 {
		t.Fatalf("按词对标签并集期望2个, 实际 %d err=%v", len(tb), err)
	}
	// 含词批量
	wm, err := m.ArticleWordsByArticleIDs(ctx, uid, []uint{id})
	if err != nil || len(wm[id]) != 2 {
		t.Fatalf("含词批量期望2个, 实际 %d err=%v", len(wm[id]), err)
	}

	// 删除词 101 的标签关联 -> 并集应只剩1个, 但含词原文仍在(删除兜底)
	m.DB.Where("user_id = ? AND word_id = ?", uid, 101).Delete(&bean.WordTag{})
	if tagMap2, _ := m.ArticleTagsByArticleIDs(ctx, uid, []uint{id}); len(tagMap2[id]) != 1 {
		t.Fatalf("删标签后并集期望1个, 实际 %d", len(tagMap2[id]))
	}
	if wm2, _ := m.ArticleWordsByArticleIDs(ctx, uid, []uint{id}); len(wm2[id]) != 2 {
		t.Fatalf("删标签不应影响含词原文, 期望2个, 实际 %d", len(wm2[id]))
	}
}
