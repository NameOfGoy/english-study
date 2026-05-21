package impl

import (
	"bytes"
	"context"
	"english-study/internal/aiapplication/wordexample"
	"english-study/internal/aiapplication/wordpicture"
	"english-study/internal/aiapplication/wordpronounce"
	"english-study/internal/dictionary"
	"english-study/internal/model/bean"
	"english-study/internal/oss"
	"english-study/internal/types"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

func (d *DictionaryImpl) AddWord(ctx context.Context, word string, triggerUserID uint) error {
	word = strings.ToLower(word)
	// 看看是否已存在
	_, err := d.m.GetWordWithPosByWord(ctx, word, nil)
	if err == nil {
		return dictionary.ErrWordExist
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	// 再看看是否正在添加中
	d.lock.Lock()
	_, ok := d.adding[word]
	if ok {
		d.lock.Unlock()
		return dictionary.ErrWordAdding
	}
	// 库没有, 又不正在添加中, 就加入添加中
	d.adding[word] = struct{}{}
	d.lock.Unlock()

	// 无论是否成功, 都要从adding中删除
	defer func() {
		d.lock.Lock()
		delete(d.adding, word)
		d.lock.Unlock()
	}()

	// 新增
	// 从std库查
	source := "stardict"
	basicWord, err := d.m.Gen.StarDict.Where(d.m.Gen.StarDict.Word.Eq(word)).WithContext(ctx).First()
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		// stardict 没有，调 AI 生成兜底
		logx.Infof("stardict 缺失单词 [%s], 调用 AI 兜底", word)
		basicWord, err = d.aiFallbackToStarDict(ctx, word)
		if err != nil {
			return err
		}
		source = "ai"
	}
	// 看看单词是否相对合法
	if !basicWord.IsValid() {
		return fmt.Errorf("word %s is not valid", word)
	}
	// 构造
	britishPhonetic := basicWord.Phonetic
	if !strings.HasPrefix(britishPhonetic, "/") {
		britishPhonetic = fmt.Sprintf("/%s/", britishPhonetic)
	}
	w := &bean.Word{
		Word:                       basicWord.Word,
		BritishPronunciation:       britishPhonetic,
		AmericanPronunciation:      d.generatePronouncePhonetic(ctx, basicWord.Word, "us"),
		BritishPronunciationAudio:  d.generatePronounceLink(ctx, basicWord.Word, "uk"),
		AmericanPronunciationAudio: d.generatePronounceLink(ctx, basicWord.Word, "us"),
		Source:                     source,
	}

	// 形式变换
	exchanges := basicWord.GetExchange()

	// 词性处理
	poses := make([]*bean.WordPos, 0)
	for _, v := range basicWord.GetPos() {
		// 词性
		p := types.WordPosSwToEnum(v[0])
		// 变化形式
		var exs map[string]string
		if exchanges != nil {
			exs = make(map[string]string)
			for _, e := range types.PosExchange(p) {
				exs[e] = exchanges[e]
			}
		}
		tps := &bean.WordPos{
			Word:        basicWord.Word,
			Pos:         p,
			Translation: v[1],
			Example:     "", // 例句异步生成, 避免阻塞主流程
			Exchange:    (&types.WordPos{Exchange: exs}).ExchangeString(),
			//Picture:     d.generatePictureLink(ctx, word, p), 不默认生成图片, 等用户需要时再生成
		}
		poses = append(poses, tps)
	}
	w.Pos = poses
	if err := d.m.InsertWord(ctx, w, nil); err != nil {
		return err
	}
	// 异步生成例句, 完成后 UPDATE 回主词典 word_pos 表 + 触发用户表回填
	// WaitGroup 让 graceful shutdown 能等待这些后台任务完成, 避免半写状态
	d.bgWG.Add(1)
	go func() {
		defer d.bgWG.Done()
		d.generateExamplesAsync(w, triggerUserID)
	}()
	return nil
}

// generateExamplesAsync 后台为每个词性生成例句, 更新主词典 + 触发者的用户表
// 用 context.Background() 避免请求 ctx 被取消导致 LLM 调用中断
// triggerUserID = 0 时只更新主词典(不回填任何用户表)
func (d *DictionaryImpl) generateExamplesAsync(w *bean.Word, triggerUserID uint) {
	bgCtx := context.Background()
	for _, pos := range w.Pos {
		example := d.generateWordExample(bgCtx, w.Word, pos.Pos)
		if example == "" {
			continue
		}
		// 1. 更新主词典
		if err := d.m.DB.WithContext(bgCtx).
			Table((&bean.WordPos{}).TableName()).
			Where("id = ?", pos.ID).
			Update("example", example).Error; err != nil {
			logx.Errorf("update example for word_pos id=%d failed: %v", pos.ID, err)
		}
		// 2. 回填触发用户的表里 example 为空的同名词条
		if triggerUserID > 0 {
			if err := d.backfillUserExample(bgCtx, triggerUserID, w.Word, pos.Pos, example); err != nil {
				logx.Errorf("backfill user[%d] example for word=%s pos=%d failed: %v", triggerUserID, w.Word, pos.Pos, err)
			}
		}
	}
}

// backfillUserExample 把新生成的例句回填到指定用户的 word_pos_user_{uid} 表
// 只覆盖 example 为空的行，避免覆盖用户已编辑的内容
func (d *DictionaryImpl) backfillUserExample(ctx context.Context, userID uint, word string, pos int, example string) error {
	table := fmt.Sprintf("word_pos_user_%d", userID)
	return d.m.DB.WithContext(ctx).Exec(
		fmt.Sprintf(
			`UPDATE %s SET example = ? WHERE word = ? AND pos = ? AND (example = '' OR example IS NULL OR example = 'null')`,
			table,
		),
		example, word, pos,
	).Error
}

// GetWord 获取std + free dictionary接口组成单词主表的信息
func (d *DictionaryImpl) GetWord(ctx context.Context, word string) (*types.Word, error) {
	word = strings.ToLower(word)
	// 从数据库查
	w, err := d.m.GetWordWithPosByWord(ctx, word, nil)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, dictionary.ErrWordNotExist
		}
		return nil, err
	}
	// 结构转换
	tw := &types.Word{
		ID:         w.ID,
		Word:       w.Word,
		UKPhonetic: w.BritishPronunciation,
		UKAudio:    w.BritishPronunciationAudio,
		USPhonetic: w.AmericanPronunciation,
		USAudio:    w.AmericanPronunciationAudio,
		Pos:        make([]*types.WordPos, 0, len(w.Pos)),
	}
	for _, pos := range w.Pos {
		p := &types.WordPos{
			ID:          pos.ID,
			WordID:      pos.WordID,
			Word:        pos.Word,
			Pos:         pos.Pos,
			Translation: pos.Translation,
			Picture:     pos.Picture,
		}
		_ = p.ExampleObject(pos.Example)
		p.ExchangeObject(pos.Exchange)
		tw.Pos = append(tw.Pos, p)
	}
	return tw, nil
}

// 生成发音链接
func (d *DictionaryImpl) generatePronounceLink(ctx context.Context, word string, accent string) string {
	// 生成音频数据
	audio, err := d.wordPronounceGenerator.GeneratePronounce(ctx, word, wordpronounce.WithAccent(accent))
	if err != nil {
		logx.Errorf("generate word [%s] pronounce of accent [%s] link failed, err: %v", word, accent, err)
		return ""
	}
	// 上传音频数据; word 走过 AI 兜底可能含非法字符, sanitize 防路径越权
	obj := fmt.Sprintf("pronounce/%s/%s.mp3", oss.SafeKeyPart(word), oss.SafeKeyPart(accent))
	link, err := d.o.Upload(ctx, types.OssBucket, obj, io.NopCloser(bytes.NewReader(audio)), int64(len(audio)), oss.WithContentType("audio/mpeg"))
	if err != nil {
		logx.Errorf("upload word [%s] pronounce of accent [%s] link failed, err: %v", word, accent, err)
		return ""
	}
	return link
}

// 生成音标
func (d *DictionaryImpl) generatePronouncePhonetic(ctx context.Context, word string, accent string) string {
	// 生成音标
	phonetic, err := d.wordPronounceGenerator.GeneratePronouncePhonetic(ctx, word, wordpronounce.WithAccent(accent))
	if err != nil {
		logx.Errorf("generate word [%s] pronounce of accent [%s] failed, err: %v", word, accent, err)
		return ""
	}
	return phonetic
}

// 生成例句
func (d *DictionaryImpl) generateWordExample(ctx context.Context, word string, pos int) string {
	examples, err := d.examGenerator.Generate(ctx, word, wordexample.WithPos(types.ToPosChinese(pos)))
	if err != nil {
		logx.Errorf("generate exam for word [%s] pos [%d] exam failed, err: %v", word, pos, err)
		return ""
	}
	return (&types.WordPos{Example: examples}).ExampleString()
}

func (d *DictionaryImpl) generateWordPictureLink(ctx context.Context, word string, pos int) string {
	pic, err := d.wordPicGenerator.Generate(ctx, word, wordpicture.WithPos(types.ToPosChinese(pos)), wordpicture.WithWidth(1024), wordpicture.WithHeight(1024))
	if err != nil {
		logx.Errorf("generate pic for word [%s] pos [%d] exam failed, err: %v", word, pos, err)
		return ""
	}
	// 上传图片数据; word 走过 AI 兜底可能含非法字符, sanitize 防路径越权
	obj := fmt.Sprintf("picture/mainword/%s/%d.png", oss.SafeKeyPart(word), pos) // 主表图片
	link, err := d.o.Upload(ctx, types.OssBucket, obj, io.NopCloser(bytes.NewReader(pic)), int64(len(pic)))
	if err != nil {
		logx.Errorf("upload word [%s] pos [%d] picture link failed, err: %v", word, pos, err)
		return ""
	}
	return link
}

func (d *DictionaryImpl) IsWord(ctx context.Context, word string) bool {
	word = strings.ToLower(word)
	for _, c := range word {
		if c < 'a' || c > 'z' {
			return false
		}
	}
	return true
}
