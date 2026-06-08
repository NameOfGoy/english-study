package article

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"english-study/internal/aiapplication/articlegen"
	"english-study/internal/logic/wordselect"
	"english-study/internal/model"
	"english-study/internal/model/bean"
	"english-study/internal/model/dto"
	"english-study/internal/types"
)

// storedBody 持久化到 article.body 的 JSON 结构.
// 只存正文 + 高亮锚点(surfaces); 词性/释义/音标读时由词库重算(天然支持"词被删则留空").
type storedBody struct {
	Sentences []storedSentence `json:"sentences"`
	UsedWords []storedUsedWord `json:"used_words"`
}

type storedSentence struct {
	En string `json:"en"`
	Zh string `json:"zh"`
}

type storedUsedWord struct {
	Word     string   `json:"word"`
	Type     int      `json:"type"`
	Surfaces []string `json:"surfaces"`
}

// selectedWord 选中词条的解析结果(供 prompt 输入 + 即时响应补齐).
type selectedWord struct {
	WordID   uint
	WordType int
	Word     string
	PosLabel string
	Meaning  string
	Senses   []types.ArticleSense
	Phonetic string
	Forms    string
}

func pairKey(id uint, t int) string {
	return strconv.FormatUint(uint64(id), 10) + "|" + strconv.Itoa(t)
}

func pairKeyText(word string, t int) string {
	return strings.ToLower(strings.TrimSpace(word)) + "|" + strconv.Itoa(t)
}

// buildWordStatusChain 按 状态/类别/标签 构造 word_statuses 查询链.
// 注意: 文章选词只读, 绝不做 SRS 时间门控(不加 next_review_at <= now).
func buildWordStatusChain(ctx context.Context, m *model.Model, userID uint, status, category int, tagIDs []uint) (dto.IWordStatusDo, error) {
	wsg := m.Gen.WordStatus
	find := wsg.WithContext(ctx).Where(wsg.UserID.Eq(userID))
	if status != 0 { // 0=全部: 不加状态谓词
		find = find.Where(wsg.Status.Eq(status))
	}
	switch category {
	case 1: // 标签(OR 语义); tagIDs 为空时 ApplyTagFilter 不筛选
		var err error
		find, err = wordselect.ApplyTagFilter(ctx, m, find, userID, tagIDs)
		if err != nil {
			return nil, err
		}
	case 2: // 单词
		find = find.Where(wsg.WordType.Eq(types.WordTypeWord))
	case 3: // 词语
		find = find.Where(wsg.WordType.Eq(types.WordTypePhrase))
	default: // 0/4 全部: 不加类别谓词
	}
	return find, nil
}

func dedupeStatuses(in []*bean.WordStatus) []*bean.WordStatus {
	seen := make(map[string]struct{}, len(in))
	out := make([]*bean.WordStatus, 0, len(in))
	for _, ws := range in {
		k := pairKey(ws.WordID, ws.WordType)
		if _, ok := seen[k]; ok {
			continue
		}
		seen[k] = struct{}{}
		out = append(out, ws)
	}
	return out
}

// resolveSelfSelect 校验自选词条归属当前用户并取回对应 word_status 行.
func resolveSelfSelect(ctx context.Context, m *model.Model, userID uint, words []types.SelfSelectWord) ([]*bean.WordStatus, error) {
	if len(words) == 0 {
		return nil, nil
	}
	ids := make([]uint, 0, len(words))
	for _, w := range words {
		ids = append(ids, w.WordID)
	}
	wsg := m.Gen.WordStatus
	rows, err := wsg.WithContext(ctx).Where(wsg.UserID.Eq(userID), wsg.WordID.In(ids...)).Find()
	if err != nil {
		return nil, err
	}
	idx := make(map[string]*bean.WordStatus, len(rows))
	for _, r := range rows {
		idx[pairKey(r.WordID, r.WordType)] = r
	}
	out := make([]*bean.WordStatus, 0, len(words))
	for _, w := range words {
		if r, ok := idx[pairKey(w.WordID, w.WordType)]; ok {
			out = append(out, r)
		}
	}
	return out, nil
}

// resolveSelected 把一条 word_status 解析为 selectedWord(含释义/音标/变形, 供 prompt 与响应补齐).
func resolveSelected(ctx context.Context, m *model.Model, ws *bean.WordStatus) (selectedWord, error) {
	uid := ws.UserID
	switch ws.WordType {
	case types.WordTypeWord:
		w, err := m.GetWordWithPosById(ctx, ws.WordID, &uid)
		if err != nil {
			return selectedWord{}, err
		}
		sw := selectedWord{WordID: ws.WordID, WordType: ws.WordType, Word: w.Word, Forms: extractForms(w)}
		var meanings []string
		for _, p := range w.Pos {
			label := types.ToPosSw(p.Pos)
			tr := strings.TrimSpace(p.Translation)
			if sw.PosLabel == "" {
				sw.PosLabel = label
			}
			if tr != "" {
				meanings = append(meanings, tr)
				sw.Senses = append(sw.Senses, types.ArticleSense{PosLabel: label, Meaning: tr})
			}
		}
		sw.Meaning = strings.Join(meanings, "；")
		sw.Phonetic = w.BritishPronunciation
		if sw.Phonetic == "" {
			sw.Phonetic = w.AmericanPronunciation
		}
		return sw, nil
	case types.WordTypePhrase:
		p, err := m.GetWordPhraseById(ctx, ws.WordID, &uid)
		if err != nil {
			return selectedWord{}, err
		}
		return selectedWord{
			WordID: ws.WordID, WordType: ws.WordType, Word: p.Phrase,
			Meaning:  p.Translation,
			Senses:   []types.ArticleSense{{PosLabel: "", Meaning: p.Translation}},
			Phonetic: p.Pronunciation,
		}, nil
	}
	return selectedWord{}, fmt.Errorf("未知词条类型 %d", ws.WordType)
}

// extractForms 从单词各词性 Exchange(JSON 数组 "type:form") 提取去重变形, 逗号分隔, 作为 prompt 变形提示.
func extractForms(w *bean.Word) string {
	seen := make(map[string]struct{})
	var forms []string
	for _, p := range w.Pos {
		ex := strings.TrimSpace(p.Exchange)
		if ex == "" {
			continue
		}
		var arr []string
		if err := json.Unmarshal([]byte(ex), &arr); err != nil {
			continue
		}
		for _, it := range arr {
			parts := strings.SplitN(it, ":", 2)
			if len(parts) != 2 {
				continue
			}
			f := strings.TrimSpace(parts[1])
			if f == "" {
				continue
			}
			if _, ok := seen[f]; ok {
				continue
			}
			seen[f] = struct{}{}
			forms = append(forms, f)
		}
	}
	return strings.Join(forms, ", ")
}

func toArticleTags(tags []model.TagBrief) []types.ArticleTag {
	out := make([]types.ArticleTag, 0, len(tags))
	for _, t := range tags {
		out = append(out, types.ArticleTag{TagID: t.TagID, Name: t.Name, Style: t.Style})
	}
	return out
}

// wordLookup 给定原型+类型 -> (word_id, 简要信息).
type wordLookup func(word string, t int) (uint, model.WordBrief)

func buildUsedWordsFromAI(used []articlegen.UsedWord, lookup wordLookup) []types.ArticleUsedWord {
	out := make([]types.ArticleUsedWord, 0, len(used))
	for _, u := range used {
		id, b := lookup(u.Word, u.Type)
		out = append(out, types.ArticleUsedWord{
			WordID: id, WordType: u.Type, Word: u.Word, Surfaces: u.Surfaces,
			PosLabel: b.PosLabel, Meaning: b.Meaning, Phonetic: b.Phonetic, Found: b.Found,
		})
	}
	return out
}

func buildUsedWordsFromStored(used []storedUsedWord, lookup wordLookup) []types.ArticleUsedWord {
	out := make([]types.ArticleUsedWord, 0, len(used))
	for _, u := range used {
		id, b := lookup(u.Word, u.Type)
		out = append(out, types.ArticleUsedWord{
			WordID: id, WordType: u.Type, Word: u.Word, Surfaces: u.Surfaces,
			PosLabel: b.PosLabel, Meaning: b.Meaning, Senses: toArticleSenses(b.Senses),
			Phonetic: b.Phonetic, Found: b.Found,
		})
	}
	return out
}

// toArticleSenses model.WordSense -> types.ArticleSense.
func toArticleSenses(ss []model.WordSense) []types.ArticleSense {
	out := make([]types.ArticleSense, 0, len(ss))
	for _, s := range ss {
		out = append(out, types.ArticleSense{PosLabel: s.PosLabel, Meaning: s.Meaning})
	}
	return out
}

func aiSentencesToTypes(ss []articlegen.Sentence) []types.ArticleSentence {
	out := make([]types.ArticleSentence, 0, len(ss))
	for _, s := range ss {
		out = append(out, types.ArticleSentence{En: s.En, Zh: s.Zh})
	}
	return out
}

func storedSentencesToTypes(ss []storedSentence) []types.ArticleSentence {
	out := make([]types.ArticleSentence, 0, len(ss))
	for _, s := range ss {
		out = append(out, types.ArticleSentence{En: s.En, Zh: s.Zh})
	}
	return out
}

func firstPosLabel(card *types.WordCard) string {
	if len(card.TranslationItems) > 0 {
		return card.TranslationItems[0].PosLabel
	}
	return ""
}

func pickPhonetic(card *types.WordCard) string {
	if strings.TrimSpace(card.UKPhonetic) != "" {
		return card.UKPhonetic
	}
	return card.USPhonetic
}
