package article

import (
	"context"
	"strings"

	"english-study/internal/aiapplication/articlegen"
	"english-study/internal/errors"
	"english-study/internal/logic/wordselect"
	"english-study/internal/model"
	"english-study/internal/model/bean"
	"english-study/internal/svc"
	"english-study/internal/types"
	"english-study/internal/utils"

	"github.com/zeromicro/go-zero/core/logx"
)

type GenerateArticleLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
	ui     *utils.UserInfo
}

func NewGenerateArticleLogic(ctx context.Context, svcCtx *svc.ServiceContext, ui *utils.UserInfo) *GenerateArticleLogic {
	return &GenerateArticleLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
		ui:     ui,
	}
}

// GenerateArticle 即时生成: 选词 -> 调 AI -> 后端补齐词性/释义/音标 + 计算标签并集. 只读, 不改 word_statuses.
func (l *GenerateArticleLogic) GenerateArticle(req *types.GenerateArticleReq) (resp *types.GenerateArticleResp, err error) {
	m := l.svcCtx.Model
	uid := l.ui.ID

	// 1. 取得选中的 word_status 列表
	var statuses []*bean.WordStatus
	if req.Method == 2 { // 自选
		if len(req.Words) < 3 || len(req.Words) > 8 {
			return nil, errors.ErrorRequestParamError("自选词语数量需在 3~8 个之间")
		}
		statuses, err = resolveSelfSelect(l.ctx, m, uid, req.Words)
		if err != nil {
			return nil, errors.ErrorDatabaseQueryError("查询自选词条失败").WithCause(err)
		}
	} else { // 随机
		count := req.Count
		if count < 3 {
			count = 3
		}
		if count > 8 {
			count = 8
		}
		find, ferr := buildWordStatusChain(l.ctx, m, uid, req.Status, req.Category, req.TagIDs)
		if ferr != nil {
			return nil, errors.ErrorDatabaseQueryError("筛选词条失败").WithCause(ferr)
		}
		statuses, err = wordselect.GetRandomWordStatus(find, count)
		if err != nil {
			return nil, err
		}
	}

	statuses = dedupeStatuses(statuses)
	if len(statuses) < 3 {
		return nil, errors.ErrorRequestParamError("可用词条不足, 至少需要 3 个")
	}
	if len(statuses) > 8 {
		statuses = statuses[:8]
	}

	// 2. 解析为 selectedWord + prompt 输入
	selected := make([]selectedWord, 0, len(statuses))
	inputs := make([]articlegen.InputWord, 0, len(statuses))
	for _, ws := range statuses {
		sw, rerr := resolveSelected(l.ctx, m, ws)
		if rerr != nil {
			return nil, errors.ErrorDatabaseQueryError("查询词条详情失败").WithCause(rerr)
		}
		selected = append(selected, sw)
		inputs = append(inputs, articlegen.InputWord{
			Word: sw.Word, Type: sw.WordType, PosLabel: sw.PosLabel, Meaning: sw.Meaning, Forms: sw.Forms,
		})
	}

	// 3. 调 AI(同步; 返回的已是 typed error)
	art, err := l.svcCtx.ArticleGen.Generate(l.ctx, inputs)
	if err != nil {
		return nil, err
	}

	// 4. 用选中词的本地词典信息补齐 used_words(高亮锚点 + 气泡卡), 计算标签并集
	// used_words 严格以"选中的词"为准: 取 AI 返回的 surfaces, 但丢弃 AI 自行新增的词.
	// 例如选了 peeler(释义含"警察"), AI 可能多吐一个 police officer —— 它不在选词里, 不能高亮.
	aiSurfaces := make(map[string][]string, len(art.UsedWords))
	for _, u := range art.UsedWords {
		aiSurfaces[pairKeyText(u.Word, u.Type)] = u.Surfaces
	}
	// 选中词文本集合(小写), 用于剔除"被 AI 错配成别的目标词"的 surface.
	// 例: AI 把 grid 的 surface 误标成 "seasoning"(另一个选中词) -> 必须丢掉, 否则点 seasoning 会弹出 grid 的释义.
	selectedSet := make(map[string]struct{}, len(selected))
	for _, s := range selected {
		selectedSet[strings.ToLower(strings.TrimSpace(s.Word))] = struct{}{}
	}
	usedWords := make([]types.ArticleUsedWord, 0, len(selected))
	for _, s := range selected {
		selfLower := strings.ToLower(strings.TrimSpace(s.Word))
		surfaces := make([]string, 0, 2)
		for _, sf := range aiSurfaces[pairKeyText(s.Word, s.WordType)] {
			sl := strings.ToLower(strings.TrimSpace(sf))
			if sl == "" {
				continue
			}
			if sl != selfLower {
				if _, isOther := selectedSet[sl]; isOther {
					continue // 这个 surface 其实是另一个目标词, 丢弃(AI 错配)
				}
			}
			surfaces = append(surfaces, sf)
		}
		if len(surfaces) == 0 {
			surfaces = []string{s.Word} // 兜底用词本身; 正文里没有就自然不高亮
		}
		usedWords = append(usedWords, types.ArticleUsedWord{
			WordID: s.WordID, WordType: s.WordType, Word: s.Word, Surfaces: surfaces,
			PosLabel: s.PosLabel, Meaning: s.Meaning, Senses: s.Senses, Phonetic: s.Phonetic, Found: true,
		})
	}

	pairs := make([]model.WordPair, 0, len(selected))
	for _, s := range selected {
		pairs = append(pairs, model.WordPair{WordID: s.WordID, WordType: s.WordType})
	}
	tags, terr := m.TagsForWordPairs(l.ctx, uid, pairs)
	if terr != nil {
		return nil, errors.ErrorDatabaseQueryError("计算标签失败").WithCause(terr)
	}

	return &types.GenerateArticleResp{
		Data: types.ArticleView{
			ID:        0,
			TitleEn:   art.TitleEn,
			TitleZh:   art.TitleZh,
			Tags:      toArticleTags(tags),
			Sentences: aiSentencesToTypes(art.Sentences),
			UsedWords: usedWords,
		},
	}, nil
}
