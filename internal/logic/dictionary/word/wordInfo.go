package word

import (
	"bytes"
	"context"
	"english-study/internal/aiapplication/wordexample"
	"english-study/internal/aiapplication/wordpicture"
	"english-study/internal/dictionary"
	"english-study/internal/model/bean"
	"english-study/internal/svc"
	"english-study/internal/types"
	"fmt"
	"io"
	"time"

	"github.com/go-kratos/kratos/v2/errors"
)

// 用于做单词相关的操作类
type WordInfo struct {
	svc    *svc.ServiceContext
	userId uint
}

func NewWordInfo(svcCtx *svc.ServiceContext, userId uint) *WordInfo {
	return &WordInfo{
		svc:    svcCtx,
		userId: userId,
	}
}

// GetCustomizedWordInfo 获取用户化的单词信息
func (wi *WordInfo) GetCustomizedWordInfo(ctx context.Context, word *types.Word) (*types.Word, error) {
	if word == nil {
		return nil, fmt.Errorf("word is nil")
	}
	// 从字典对象获取单词的基本信息
	mainWord, err := wi.getWordFromDictionary(ctx, word.Word)
	if err != nil {
		return nil, err
	}

	// 把word里为空的, 用mainWord的补上
	if word.USPhonetic == "" {
		word.USPhonetic = mainWord.USPhonetic
	}
	if word.USAudio == "" {
		word.USAudio = mainWord.USAudio
	}
	if word.UKPhonetic == "" {
		word.UKPhonetic = mainWord.UKPhonetic
	}
	if word.UKAudio == "" {
		word.UKAudio = mainWord.UKAudio
	}

	if len(word.Pos) != 0 { // 如果词性有传入, 就用传入的
		return word, nil
	}
	// 没有则从mainWord里获取
	for _, pos := range mainWord.Pos {
		word.Pos = append(word.Pos, &types.WordPos{
			WordID:      mainWord.ID,
			Pos:         pos.Pos,
			Translation: pos.Translation,
			Example:     pos.Example,
			Picture:     pos.Picture,
			Exchange:    pos.Exchange,
		})
	}
	return word, nil
}

func (wi *WordInfo) getWordFromDictionary(ctx context.Context, word string) (*types.Word, error) {
	w, err := wi.svc.Dictionary.GetWord(ctx, word)
	if err == nil { // 取得到, 直接返回
		return w, nil
	}
	if !errors.Is(err, dictionary.ErrWordNotExist) { // 不是单词不存在错误, 则返回错误
		return nil, err
	}
	// 在字典表里新增（带上触发用户ID, 便于异步例句生成完成后回填该用户的表）
	err = wi.svc.Dictionary.AddWord(ctx, word, wi.userId)
	if err == nil {
		return wi.svc.Dictionary.GetWord(ctx, word) // 新增成功, 则返回单词
	}
	if errors.Is(err, dictionary.ErrWordAdding) { // 单词正在添加中, 则等待
		for wi.svc.Dictionary.IsWordAdding(ctx, word) {
			time.Sleep(time.Millisecond * 1000)
		}
		return wi.svc.Dictionary.GetWord(ctx, word)
	}
	if errors.Is(err, dictionary.ErrWordExist) {
		return wi.svc.Dictionary.GetWord(ctx, word) // 单词已存在, 则返回单词
	}
	return nil, err // 其他错误, 则返回错误
}

// IncreaseWord 新增单词
func (wi *WordInfo) IncreaseWord(ctx context.Context, w *types.Word, userId *uint) (err error) {
	// 单词表新增单词
	bw := &bean.Word{
		Word:                       w.Word,
		AmericanPronunciation:      w.USPhonetic,
		AmericanPronunciationAudio: w.USAudio,
		BritishPronunciation:       w.UKPhonetic,
		BritishPronunciationAudio:  w.UKAudio,
		Pos:                        make([]*bean.WordPos, len(w.Pos)),
	}
	// 词性表新增单词的词性
	for i, pos := range w.Pos {
		bw.Pos[i] = &bean.WordPos{
			WordID:      bw.ID,
			Word:        bw.Word,
			Pos:         pos.Pos,
			Translation: pos.Translation,
			Example:     pos.ExampleString(),
			Picture:     pos.Picture,
			Exchange:    pos.ExchangeString(),
		}
	}

	err = wi.svc.Model.InsertWord(ctx, bw, userId)
	if err != nil {
		return err
	}

	if userId == nil {
		return nil
	}

	// 自动变为学习状态
	err = wi.svc.Model.Gen.WordStatus.WithContext(ctx).Create(&bean.WordStatus{
		WordID:   bw.ID,
		WordType: types.WordTypeWord,
		Status:   types.WordStatusStudy,
		UserID:   *userId,
	})

	return nil
}

// GeneratePicture 生成词性的图片
func (wi *WordInfo) GeneratePicture(ctx context.Context, word string, pos int) (link string, err error) {
	// 取配置文件的prompt模板
	p, err := wi.svc.WordPic.Generate(ctx, word,
		wordpicture.WithPos(types.ToPosChinese(pos)),
		wordpicture.WithPromptTemplate(wi.svc.ViperConfig.GetWordPicturePromptTemplate()))
	if err != nil {
		return "", fmt.Errorf("generate picture failed, word: %s, pos: %d, err: %w", word, pos, err)
	}
	path, err := wi.svc.Oss.Upload(ctx,
		types.OssBucket,
		fmt.Sprintf("picture/user_word_%d/%s/%d_%d.png", wi.userId, word, pos, time.Now().Unix()),
		io.NopCloser(bytes.NewReader(p)), int64(len(p)))
	if err != nil {
		return "", fmt.Errorf("upload picture failed, word: %s, pos: %d, err: %w", word, pos, err)
	}
	return path, nil
}

// GenerateExample 生成词性的例句
func (wi *WordInfo) GenerateExample(ctx context.Context, word string, pos int, count int, extraTrans string) (e []*types.Example, err error) {
	return wi.svc.WordExam.Generate(ctx, word, wordexample.WithPos(types.ToPosChinese(pos)), wordexample.WithCount(count), wordexample.WithTranslation(extraTrans))
}

func (wi *WordInfo) IsWord(word string) bool {
	return wi.svc.Dictionary.IsWord(context.Background(), word)
}
