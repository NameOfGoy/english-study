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
	"strings"

	"github.com/zeromicro/go-zero/core/logx"
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
	// 转换出路径
	path := utils.ToOssPath(types.OssBucket, req.FilePath)

	// 取得文件
	file, err := l.svcCtx.Oss.Download(l.ctx, types.OssBucket, path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var words []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" {
			words = append(words, line)
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	// 创建导入任务记录
	fileName := req.FileName
	if fileName == "" {
		fileName = req.FilePath
	}
	task := &bean.ImportTask{
		UserID:   l.ui.ID,
		FileName: fileName,
		Status:   0, // 待处理
		Total:    len(words),
	}
	if err := l.svcCtx.Model.DB.Create(task).Error; err != nil {
		return nil, err
	}

	wi := word.NewWordInfo(l.svcCtx, l.ui.ID)
	wpi := word.NewPhraseInfo(l.svcCtx, l.ui.ID)

	go func() {
		ctx := context.Background()
		db := l.svcCtx.Model.DB

		// 更新状态为进行中
		db.Model(task).Update("status", 1)

		logx.Infof("用户[%s] 导入任务[%d] 开始, 总数: %d", l.ui.Username, task.ID, len(words))

		var failWords []string
		var successCount, failCount int

		for k, w := range words {
			// 更新当前进度
			db.Model(task).Updates(map[string]interface{}{
				"current":      k + 1,
				"current_word": w,
			})

			logx.Infof("用户[%s] 任务[%d] 导入第 %d/%d 个: %s", l.ui.Username, task.ID, k+1, len(words), w)

			success := false
			if wi.IsWord(w) {
				mainWord, err := wi.GetCustomizedWordInfo(ctx, &types.Word{Word: w})
				if err != nil {
					logx.Errorf("获取单词 %s 失败: %v", w, err)
				} else {
					// 生成图片
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
					}
				}
			} else {
				wordPhrase, err := wpi.GetCustomizedPhraseInfo(ctx, &types.WordPhrase{Phrase: w})
				if err != nil {
					logx.Errorf("获取短语 %s 失败: %v", w, err)
				} else {
					wordPhrase.Picture, err = wpi.GeneratePicture(ctx, w)
					if err != nil {
						logx.Errorf("生成短语图片失败, phrase: %s, err: %v", w, err)
					}
					if err = wpi.IncreasePhrase(ctx, wordPhrase, &l.ui.ID); err != nil {
						logx.Errorf("新增短语 %s 失败: %v", w, err)
					} else {
						success = true
					}
				}
			}

			if success {
				successCount++
				db.Model(task).Update("success_count", successCount)
			} else {
				failCount++
				failWords = append(failWords, w)
				failWordsJSON, _ := json.Marshal(failWords)
				db.Model(task).Updates(map[string]interface{}{
					"fail_count": failCount,
					"fail_words": string(failWordsJSON),
				})
			}

			logx.Infof("用户[%s] 任务[%d] 第 %d/%d 个 %s %s",
				l.ui.Username, task.ID, k+1, len(words), w, map[bool]string{true: "成功", false: "失败"}[success])
		}

		// 更新为已完成
		db.Model(task).Update("status", 2)
		logx.Infof("用户[%s] 任务[%d] 导入完成, 成功: %d, 失败: %d", l.ui.Username, task.ID, successCount, failCount)
	}()

	return &types.ImportWordResp{
		TaskId: task.ID,
	}, nil
}
