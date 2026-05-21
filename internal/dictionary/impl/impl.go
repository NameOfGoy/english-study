package impl

import (
	"context"
	"english-study/internal/aiapplication/wordexample"
	"english-study/internal/aiapplication/wordpicture"
	"english-study/internal/aiapplication/wordpronounce"
	"english-study/internal/aiapplication/wordtranslation"
	"english-study/internal/model"
	"english-study/internal/oss"
	"sync"
)

// 自行实现
type DictionaryImpl struct {
	o                        oss.OSS
	m                        *model.Model
	examGenerator            wordexample.WordExample
	wordPicGenerator         wordpicture.Picture
	wordPronounceGenerator   wordpronounce.WordPronounce
	wordTranslationGenerator wordtranslation.WordTranslation

	// 并发控制
	adding map[string]struct{}
	lock   sync.RWMutex

	// 后台异步任务追踪 (例句生成等), 优雅关停时 Wait
	bgWG sync.WaitGroup
}

func NewDictionaryImpl(o oss.OSS, m *model.Model, examGenerator wordexample.WordExample, wordPicGenerator wordpicture.Picture, wordPronounceGenerator wordpronounce.WordPronounce, wordTranslationGenerator wordtranslation.WordTranslation) *DictionaryImpl {
	return &DictionaryImpl{
		o:                        o,
		m:                        m,
		examGenerator:            examGenerator,
		wordPicGenerator:         wordPicGenerator,
		wordPronounceGenerator:   wordPronounceGenerator,
		wordTranslationGenerator: wordTranslationGenerator,
		adding:                   make(map[string]struct{}),
	}
}

// WaitBackground 等所有后台例句生成 goroutine 完成, 用于优雅关停
func (d *DictionaryImpl) WaitBackground() {
	d.bgWG.Wait()
}

func (d *DictionaryImpl) IsWordAdding(ctx context.Context, word string) bool {
	d.lock.RLock()
	defer d.lock.RUnlock()
	_, ok := d.adding[word]
	return ok
}
