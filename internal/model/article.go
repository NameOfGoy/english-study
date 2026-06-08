package model

import (
	"context"
	stderrors "errors"
	"strings"

	"english-study/internal/model/bean"
	"english-study/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

// TagBrief 文章标签(并集计算结果). name 取自 tags.tag 列.
type TagBrief struct {
	TagID uint   `json:"tag_id"`
	Name  string `json:"name"`
	Style string `json:"style"`
}

// WordPair (word_id, word_type) 元组, 用于未入库文章的标签并集计算.
type WordPair struct {
	WordID   uint
	WordType int
}

// WordSense 一个词性下的释义.
type WordSense struct {
	PosLabel string
	Meaning  string
}

// WordBrief 词条简要信息(气泡卡/列表富化用). Found=false 时其余字段留空.
type WordBrief struct {
	PosLabel string
	Meaning  string
	Senses   []WordSense // 全部词性+释义(气泡卡按词性分行)
	Phonetic string
	Found    bool
}

// CreateArticleWithWords 事务写入 article + article_words(每个词一行).
// 保证两者要么都写入要么都不写, 避免出现没有 article_words 的孤儿文章.
func (m *Model) CreateArticleWithWords(ctx context.Context, art *bean.Article, words []bean.ArticleWord) (id uint, err error) {
	tx := m.DB.Begin()
	defer func() {
		if err != nil {
			if txe := tx.Rollback().Error; txe != nil {
				logx.Errorf("CreateArticleWithWords rollback failed: %v", txe)
			}
			return
		}
		if txe := tx.Commit().Error; txe != nil {
			logx.Errorf("CreateArticleWithWords commit failed: %v", txe)
			err = txe
		}
	}()
	if err = tx.WithContext(ctx).Create(art).Error; err != nil {
		return 0, err
	}
	if len(words) > 0 {
		for i := range words {
			words[i].ArticleID = art.ID
			words[i].UserID = art.UserID
		}
		if err = tx.WithContext(ctx).Create(&words).Error; err != nil {
			return 0, err
		}
	}
	return art.ID, nil
}

// GetArticleByID 取单篇文章(按 user 维度隔离).
func (m *Model) GetArticleByID(ctx context.Context, id, userID uint) (*bean.Article, error) {
	var a bean.Article
	if err := m.DB.WithContext(ctx).Where("id = ? AND user_id = ?", id, userID).Take(&a).Error; err != nil {
		return nil, err
	}
	return &a, nil
}

// GetArticleWords 取一篇文章包含的词条行(顺序稳定).
func (m *Model) GetArticleWords(ctx context.Context, articleID, userID uint) ([]bean.ArticleWord, error) {
	var ws []bean.ArticleWord
	err := m.DB.WithContext(ctx).
		Where("article_id = ? AND user_id = ?", articleID, userID).
		Order("id asc").Find(&ws).Error
	return ws, err
}

// ListArticles 分页 + 搜索. keyword 同时命中标题(中英)与含词(英文); tagIDs 命中"含任一选中标签的词".
func (m *Model) ListArticles(ctx context.Context, userID uint, keyword string, tagIDs []uint, offset, limit int) ([]bean.Article, int64, error) {
	q := m.DB.WithContext(ctx).Model(&bean.Article{}).Where("user_id = ?", userID)

	if kw := strings.TrimSpace(keyword); kw != "" {
		like := "%" + kw + "%"
		q = q.Where(
			m.DB.Where("title_en ILIKE ?", like).
				Or("title_zh ILIKE ?", like).
				Or("EXISTS (SELECT 1 FROM article_words aw WHERE aw.article_id = article.id AND aw.user_id = article.user_id AND aw.word_text ILIKE ?)", like),
		)
	}
	if len(tagIDs) > 0 {
		q = q.Where("EXISTS (SELECT 1 FROM article_words aw "+
			"JOIN word_tags wt ON wt.word_id = aw.word_id AND wt.word_type = aw.word_type AND wt.user_id = aw.user_id "+
			"WHERE aw.article_id = article.id AND aw.user_id = article.user_id AND wt.tag_id IN ?)", tagIDs)
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var items []bean.Article
	if err := q.Order("created_at DESC").Offset(offset).Limit(limit).Find(&items).Error; err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

// ArticleWordsByArticleIDs 批量取多篇文章的含词原文, 按 article_id 分组(列表页用, 避免 N+1).
func (m *Model) ArticleWordsByArticleIDs(ctx context.Context, userID uint, ids []uint) (map[uint][]string, error) {
	out := make(map[uint][]string)
	if len(ids) == 0 {
		return out, nil
	}
	var rows []struct {
		ArticleID uint
		WordText  string
	}
	err := m.DB.WithContext(ctx).Table("article_words").
		Select("article_id, word_text").
		Where("article_id IN ? AND user_id = ?", ids, userID).
		Order("article_id asc, id asc").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	for _, r := range rows {
		out[r.ArticleID] = append(out[r.ArticleID], r.WordText)
	}
	return out, nil
}

// ArticleTagsByArticleIDs 批量计算多篇文章的标签并集, 按 article_id 分组(列表页用, 避免 N+1).
func (m *Model) ArticleTagsByArticleIDs(ctx context.Context, userID uint, ids []uint) (map[uint][]TagBrief, error) {
	out := make(map[uint][]TagBrief)
	if len(ids) == 0 {
		return out, nil
	}
	var rows []struct {
		ArticleID uint
		TagID     uint
		Name      string
		Style     string
	}
	err := m.DB.WithContext(ctx).Table("article_words AS aw").
		Select("DISTINCT aw.article_id AS article_id, t.id AS tag_id, t.tag AS name, t.style AS style").
		Joins("JOIN word_tags wt ON wt.word_id = aw.word_id AND wt.word_type = aw.word_type AND wt.user_id = aw.user_id").
		Joins("JOIN tags t ON t.id = wt.tag_id").
		Where("aw.article_id IN ? AND aw.user_id = ?", ids, userID).
		Order("aw.article_id asc, t.id asc").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	for _, r := range rows {
		out[r.ArticleID] = append(out[r.ArticleID], TagBrief{TagID: r.TagID, Name: r.Name, Style: r.Style})
	}
	return out, nil
}

// TagsForWordPairs 计算给定 (word_id, word_type) 集合的标签并集(用于未入库的即时文章).
// 精确按元组匹配, 避免 word/phrase 同 id 互相串标签.
func (m *Model) TagsForWordPairs(ctx context.Context, userID uint, pairs []WordPair) ([]TagBrief, error) {
	out := []TagBrief{}
	if len(pairs) == 0 {
		return out, nil
	}
	conds := make([]string, 0, len(pairs))
	args := make([]interface{}, 0, len(pairs)*2+1)
	args = append(args, userID)
	for _, p := range pairs {
		conds = append(conds, "(wt.word_id = ? AND wt.word_type = ?)")
		args = append(args, p.WordID, p.WordType)
	}
	sql := "SELECT DISTINCT t.id AS tag_id, t.tag AS name, t.style AS style " +
		"FROM word_tags wt JOIN tags t ON t.id = wt.tag_id " +
		"WHERE wt.user_id = ? AND (" + strings.Join(conds, " OR ") + ") ORDER BY t.id"
	err := m.DB.WithContext(ctx).Raw(sql, args...).Scan(&out).Error
	return out, err
}

// GetWordBrief 按词条文本取简要信息(词性/释义/音标). 词条不存在 -> Found=false 留空.
// 词条被删等 RecordNotFound 视为正常缺失; 其它错误记日志但同样返回空(展示场景容错优先).
func (m *Model) GetWordBrief(ctx context.Context, userID uint, word string, wordType int) WordBrief {
	uid := userID
	switch wordType {
	case types.WordTypeWord:
		w, err := m.GetWordWithPosByWord(ctx, word, &uid)
		if err != nil {
			if !stderrors.Is(err, gorm.ErrRecordNotFound) {
				logx.WithContext(ctx).Errorf("GetWordBrief 查询单词 %q 失败: %v", word, err)
			}
			return WordBrief{}
		}
		return wordBriefFromWord(w)
	case types.WordTypePhrase:
		p, err := m.GetWordPhraseByPhrase(ctx, word, &uid)
		if err != nil {
			if !stderrors.Is(err, gorm.ErrRecordNotFound) {
				logx.WithContext(ctx).Errorf("GetWordBrief 查询短语 %q 失败: %v", word, err)
			}
			return WordBrief{}
		}
		return WordBrief{
			PosLabel: "",
			Meaning:  p.Translation,
			Senses:   []WordSense{{PosLabel: "", Meaning: p.Translation}},
			Phonetic: p.Pronunciation,
			Found:    true,
		}
	}
	return WordBrief{}
}

func wordBriefFromWord(w *bean.Word) WordBrief {
	var posLabel string
	var meanings []string
	var senses []WordSense
	for _, p := range w.Pos {
		tr := strings.TrimSpace(p.Translation)
		label := types.ToPosSw(p.Pos)
		if posLabel == "" {
			posLabel = label
		}
		if tr != "" {
			meanings = append(meanings, tr)
			senses = append(senses, WordSense{PosLabel: label, Meaning: tr})
		}
	}
	phonetic := w.BritishPronunciation
	if phonetic == "" {
		phonetic = w.AmericanPronunciation
	}
	return WordBrief{
		PosLabel: posLabel,
		Meaning:  strings.Join(meanings, "；"),
		Senses:   senses,
		Phonetic: phonetic,
		Found:    true,
	}
}
