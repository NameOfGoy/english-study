package impl

import (
	"context"
	"english-study/internal/dictionary"
	"english-study/internal/model/bean"
	"english-study/internal/types"
	"errors"
	"fmt"
	"strings"

	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

func (d *DictionaryImpl) AddPhrase(ctx context.Context, phrase string) error {
	if !isPhrase(phrase) {
		return fmt.Errorf("短语拼写不合法")
	}
	// 看看是否已存在
	_, err := d.m.Gen.WordPhrase.WithContext(ctx).Where(d.m.Gen.WordPhrase.Phrase.Eq(phrase)).Take()
	if err == nil {
		return dictionary.ErrWordExist
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	// 先查翻译
	translation, err := d.wordTranslationGenerator.Generate(ctx, phrase)
	if err != nil {
		return err
	}
	if strings.TrimSpace(translation) == "" {
		return fmt.Errorf("AI 未能为短语 [%s] 生成有效翻译", phrase)
	}

	// 再生成发音
	pronunciation := d.generatePronounceLink(ctx, phrase, "us")

	// 再生成例句
	example, err := d.generatePhraseExample(ctx, phrase)
	if err != nil {
		return err
	}

	bp := &bean.WordPhrase{
		Phrase:        phrase,
		Translation:   translation,
		Pronunciation: pronunciation,
		Example:       example,
	}

	return d.m.InsertWordPhrase(ctx, bp, nil)
}

func (d *DictionaryImpl) GetPhrase(ctx context.Context, phrase string) (*types.WordPhrase, error) {
	// 从主库查询
	word, err := d.m.Gen.WordPhrase.WithContext(ctx).Where(d.m.Gen.WordPhrase.Phrase.Eq(phrase)).Take()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, dictionary.ErrWordNotExist
		}
		return nil, err
	}
	// 转换为types.Word
	w := &types.WordPhrase{
		ID:            word.ID,
		Phrase:        word.Phrase,
		Translation:   word.Translation,
		Pronunciation: word.Pronunciation,
		Picture:       word.Picture,
	}
	err = w.ExampleObject(word.Example)
	if err != nil {
		return nil, err
	}
	return w, nil
}

// 生成例句
func (d *DictionaryImpl) generatePhraseExample(ctx context.Context, phrase string) (string, error) {
	examples, err := d.examGenerator.Generate(ctx, phrase)
	if err != nil {
		logx.Errorf("generate exam for phrase [%s] exam failed, err: %v", phrase, err)
		return "", err
	}
	return (&types.WordPos{Example: examples}).ExampleString(), nil
}

func isPhrase(phrase string) bool {
	return strings.Contains(strings.TrimSpace(phrase), " ")
}
