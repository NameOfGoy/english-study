package word

import (
	"bytes"
	"context"
	"english-study/internal/aiapplication/wordexample"
	"english-study/internal/aiapplication/wordpicture"
	"english-study/internal/dictionary"
	"english-study/internal/model/bean"
	"english-study/internal/oss"
	"english-study/internal/svc"
	"english-study/internal/types"
	"errors"
	"fmt"
	"io"
	"time"
)

// 用于做短语相关的操作类
type PhraseInfo struct {
	svc    *svc.ServiceContext
	userId uint
}

func NewPhraseInfo(svc *svc.ServiceContext, userId uint) *PhraseInfo {
	return &PhraseInfo{svc, userId}
}

// GetCustomizedWordInfo 获取用户化的单词信息
func (pi *PhraseInfo) GetCustomizedPhraseInfo(ctx context.Context, phrase *types.WordPhrase) (*types.WordPhrase, error) {
	if phrase == nil {
		return nil, fmt.Errorf("phrase is nil")
	}
	// 从字典对象获取短语的基本信息
	mainPhrase, err := pi.getPhraseFromDictionary(ctx, phrase.Phrase)
	if err != nil {
		return nil, err
	}

	// 把phrase里为空的, 用mainPhrase的补上
	if phrase.Pronunciation == "" {
		phrase.Pronunciation = mainPhrase.Pronunciation
	}
	if phrase.Translation == "" {
		phrase.Translation = mainPhrase.Translation
	}
	if phrase.Example == nil {
		phrase.Example = mainPhrase.Example
	}
	if phrase.Picture == "" {
		phrase.Picture = mainPhrase.Picture
	}

	return phrase, nil
}

func (pi *PhraseInfo) getPhraseFromDictionary(ctx context.Context, phrase string) (*types.WordPhrase, error) {
	w, err := pi.svc.Dictionary.GetPhrase(ctx, phrase)
	if err == nil { // 取得到, 直接返回
		return w, nil
	}
	if !errors.Is(err, dictionary.ErrWordNotExist) { // 不是单词不存在错误, 则返回错误
		return nil, err
	}
	// 在字典表里新增
	err = pi.svc.Dictionary.AddPhrase(ctx, phrase)
	if err == nil {
		return pi.svc.Dictionary.GetPhrase(ctx, phrase) // 新增成功, 则返回短语
	}
	if errors.Is(err, dictionary.ErrWordAdding) { // 短语正在添加中, 则等待
		// 带 ctx 超时和 5s 总等待上限, 防止后台 goroutine panic 导致这里永久 spin
		deadline := time.NewTimer(5 * time.Second)
		defer deadline.Stop()
		tick := time.NewTicker(200 * time.Millisecond)
		defer tick.Stop()
		for pi.svc.Dictionary.IsWordAdding(ctx, phrase) {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-deadline.C:
				return nil, fmt.Errorf("等待添加短语超时: %s", phrase)
			case <-tick.C:
			}
		}
		if w, gerr := pi.svc.Dictionary.GetPhrase(ctx, phrase); gerr == nil {
			return w, nil
		}
		time.Sleep(100 * time.Millisecond)
		return pi.svc.Dictionary.GetPhrase(ctx, phrase)
	}
	if errors.Is(err, dictionary.ErrWordExist) {
		return pi.svc.Dictionary.GetPhrase(ctx, phrase) // 短语已存在, 则返回短语
	}
	return nil, err // 其他错误, 则返回错误
}

// IncreasePhrase 新增短语
func (pi *PhraseInfo) IncreasePhrase(ctx context.Context, w *types.WordPhrase, userId *uint) (err error) {
	// 单词表新增单词
	bw := &bean.WordPhrase{
		Phrase:        w.Phrase,
		Translation:   w.Translation,
		Pronunciation: w.Pronunciation,
		Example:       w.ExampleString(),
		Picture:       w.Picture,
	}
	err = pi.svc.Model.InsertWordPhrase(ctx, bw, userId)
	if err != nil {
		return err
	}
	// 把新分配的 user 表 ID 回写到调用方 w, 给后续(贴标签)使用
	w.ID = bw.ID
	if userId == nil {
		return nil
	}
	// 自动变为学习状态
	err = pi.svc.Model.Gen.WordStatus.WithContext(ctx).Create(&bean.WordStatus{
		WordID:   bw.ID,
		WordType: types.WordTypePhrase,
		Status:   types.WordStatusStudy,
		UserID:   *userId,
	})
	return
}

// GeneratePicture 生成图片
func (pi *PhraseInfo) GeneratePicture(ctx context.Context, phrase string) (link string, err error) {
	// 取配置文件的prompt模板
	p, err := pi.svc.WordPic.Generate(ctx, phrase, wordpicture.WithPromptTemplate(pi.svc.ViperConfig.GetWordPicturePromptTemplate()))
	if err != nil {
		return "", fmt.Errorf("generate picture failed, phrase: %s, err: %w", phrase, err)
	}
	// phrase 含空格/标点, 在 key 中需要 sanitize, 防止 ../ 越权
	path, err := pi.svc.Oss.Upload(ctx,
		types.OssBucket,
		fmt.Sprintf("picture/user_phrase_%d/%s/%d.png", pi.userId, oss.SafeKeyPart(phrase), time.Now().Unix()),
		io.NopCloser(bytes.NewReader(p)), int64(len(p)))
	if err != nil {
		return "", fmt.Errorf("upload picture failed, phrase: %s, err: %w", phrase, err)
	}
	return path, nil
}

// GenerateExample 生成短语的例句
func (pi *PhraseInfo) GenerateExample(ctx context.Context, phrase string, count int, extraTrans string) (e []*types.Example, err error) {
	return pi.svc.WordExam.Generate(ctx, phrase, wordexample.WithCount(count), wordexample.WithTranslation(extraTrans))
}
