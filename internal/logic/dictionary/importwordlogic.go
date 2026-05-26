package dictionary

import (
	"bufio"
	"context"
	"encoding/json"
	"english-study/internal/logic/dictionary/word"
	"english-study/internal/model/bean"
	"english-study/internal/svc"
	"english-study/internal/types"
	"english-study/internal/utils"

	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm/clause"
)

type ImportWordLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
	ui     *utils.UserInfo
}

// 字典模块/批量导入单词
func NewImportWordLogic(ctx context.Context, svcCtx *svc.ServiceContext, ui *utils.UserInfo) *ImportWordLogic {
	return &ImportWordLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
		ui:     ui,
	}
}

func (l *ImportWordLogic) ImportWord(req *types.ImportWordReq) (resp *types.ImportWordResp, err error) {
	path := utils.ToOssPath(types.OssBucket, req.FilePath)

	file, err := l.svcCtx.Oss.Download(l.ctx, types.OssBucket, path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	// 解析 [tag] / [---] 标记 → 词条 + 当前标签
	items := ParseImportLines(lines)

	// 创建导入任务记录
	fileName := req.FileName
	if fileName == "" {
		fileName = req.FilePath
	}
	task := &bean.ImportTask{
		UserID:   l.ui.ID,
		FileName: fileName,
		Status:   0, // 待处理
		Total:    len(items),
	}
	if err := l.svcCtx.Model.DB.Create(task).Error; err != nil {
		return nil, err
	}

	wi := word.NewWordInfo(l.svcCtx, l.ui.ID)
	wpi := word.NewPhraseInfo(l.svcCtx, l.ui.ID)

	go func() {
		ctx := context.Background()
		db := l.svcCtx.Model.DB

		db.Model(task).Update("status", 1)
		logx.Infof("用户[%s] 导入任务[%d] 开始, 总数: %d", l.ui.Username, task.ID, len(items))

		// 同一个文件内多次出现的标签名只查/建一次
		tagIDCache := map[string]uint{}

		var failWords []string
		var successCount, failCount int

		for k, it := range items {
			db.Model(task).Updates(map[string]interface{}{
				"current":      k + 1,
				"current_word": it.Word,
			})

			logx.Infof("用户[%s] 任务[%d] 导入第 %d/%d 个: %s tags=%v",
				l.ui.Username, task.ID, k+1, len(items), it.Word, it.TagNames)

			success := false
			var newID uint
			var wordType int

			if wi.IsWord(it.Word) {
				wordType = types.WordTypeWord
				mainWord, err := wi.GetCustomizedWordInfo(ctx, &types.Word{Word: it.Word})
				if err != nil {
					logx.Errorf("获取单词 %s 失败: %v", it.Word, err)
				} else {
					for _, wp := range mainWord.Pos {
						if link, err := wi.GeneratePicture(ctx, mainWord.Word, wp.Pos); err != nil {
							logx.Errorf("生成图片失败, word: %s, pos: %d, err: %v", mainWord.Word, wp.Pos, err)
						} else {
							wp.Picture = link
						}
					}
					if err = wi.IncreaseWord(ctx, mainWord, &l.ui.ID); err != nil {
						logx.Errorf("新增单词 %s 失败: %v", mainWord.Word, err)
					} else {
						success = true
						newID = mainWord.ID
					}
				}
			} else {
				wordType = types.WordTypePhrase
				wordPhrase, err := wpi.GetCustomizedPhraseInfo(ctx, &types.WordPhrase{Phrase: it.Word})
				if err != nil {
					logx.Errorf("获取短语 %s 失败: %v", it.Word, err)
				} else {
					wordPhrase.Picture, err = wpi.GeneratePicture(ctx, it.Word)
					if err != nil {
						logx.Errorf("生成短语图片失败, phrase: %s, err: %v", it.Word, err)
					}
					if err = wpi.IncreasePhrase(ctx, wordPhrase, &l.ui.ID); err != nil {
						logx.Errorf("新增短语 %s 失败: %v", it.Word, err)
					} else {
						success = true
						newID = wordPhrase.ID
					}
				}
			}

			// 贴标签: 同名优先复用 (系统标签优先 → 用户已有 → 自动新建用户级)
			if success && newID > 0 && len(it.TagNames) > 0 {
				for _, tagName := range it.TagNames {
					tagID, ok := tagIDCache[tagName]
					if !ok {
						var err error
						tagID, err = l.findOrCreateUserTag(ctx, tagName)
						if err != nil {
							logx.Errorf("找/建标签 %q 失败: %v", tagName, err)
						}
						tagIDCache[tagName] = tagID
					}
					if tagID == 0 {
						continue
					}
					if err := db.WithContext(ctx).
						Clauses(clause.OnConflict{DoNothing: true}).
						Create(&bean.WordTag{
							WordID:   newID,
							WordType: wordType,
							TagID:    tagID,
							UserID:   l.ui.ID,
						}).Error; err != nil {
						logx.Errorf("挂标签失败 word=%s tag=%s err=%v", it.Word, tagName, err)
					}
				}
			}

			if success {
				successCount++
				db.Model(task).Update("success_count", successCount)
			} else {
				failCount++
				failWords = append(failWords, it.Word)
				failWordsJSON, _ := json.Marshal(failWords)
				db.Model(task).Updates(map[string]interface{}{
					"fail_count": failCount,
					"fail_words": string(failWordsJSON),
				})
			}

			logx.Infof("用户[%s] 任务[%d] 第 %d/%d 个 %s %s",
				l.ui.Username, task.ID, k+1, len(items), it.Word,
				map[bool]string{true: "成功", false: "失败"}[success])
		}

		db.Model(task).Update("status", 2)
		logx.Infof("用户[%s] 任务[%d] 导入完成, 成功: %d, 失败: %d",
			l.ui.Username, task.ID, successCount, failCount)
	}()

	return &types.ImportWordResp{
		TaskId: task.ID,
	}, nil
}

// findOrCreateUserTag 找当前用户名下同名标签; 没有就建一个; 系统标签 (user_id=0) 同名也算"已存在"复用
// findOrCreateTagForImport 按名字找标签 ID, 找不到则在当前用户名下新建.
// 优先级 (字符串全等匹配, 同名时):
//  1. 系统标签 (user_id=0) — 共享, 不污染私人空间
//  2. 当前用户自己的标签
// 找不到任何同名时, 创建为当前用户的用户级标签 (不会自动创建系统标签).
func (l *ImportWordLogic) findOrCreateUserTag(ctx context.Context, name string) (uint, error) {
	tg := l.svcCtx.Model.Gen.Tag
	got, err := tg.WithContext(ctx).
		Where(tg.Tag.Eq(name)).
		Where(tg.WithContext(ctx).Where(tg.UserID.Eq(l.ui.ID)).Or(tg.UserID.Eq(0))).
		Order(tg.UserID.Asc()). // user_id=0 (系统) 排最前, 同名时优先用系统标签
		First()
	if err == nil {
		return got.ID, nil
	}
	// 没有同名: 创建用户级标签
	newTag := &bean.Tag{Tag: name, UserID: l.ui.ID}
	if err := tg.WithContext(ctx).
		Clauses(clause.OnConflict{DoNothing: true}).
		Create(newTag); err != nil {
		return 0, err
	}
	// ON CONFLICT 时 newTag.ID 可能仍是 0, 兜底再查一次
	if newTag.ID == 0 {
		got, err := tg.WithContext(ctx).
			Where(tg.Tag.Eq(name), tg.UserID.Eq(l.ui.ID)).
			First()
		if err != nil {
			return 0, err
		}
		return got.ID, nil
	}
	return newTag.ID, nil
}
