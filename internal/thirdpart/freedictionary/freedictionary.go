package freedictionary

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
)

var (
	freeDictionaryApi = "https://api.dictionaryapi.dev/api/v2/entries/en/"

	// 第三方公共字典 API 偶尔会很慢；用 10s 超时 + 复用连接池
	httpClient = &http.Client{Timeout: 10 * time.Second}
)

type License struct {
	Name string `json:"name"` // 许可证名称
	URL  string `json:"url"`  // 链接
}

type Phonetic struct {
	Text      string  `json:"text,omitempty"`      // 音标
	Audio     string  `json:"audio,omitempty"`     // 音频链接
	SourceURL string  `json:"sourceUrl,omitempty"` // 来源链接
	License   License `json:"license,omitempty"`   // 许可证
}

type Definition struct {
	Definition string   `json:"definition"`         // 定义
	Synonyms   []string `json:"synonyms,omitempty"` // 近义词
	Antonyms   []string `json:"antonyms,omitempty"` // 反义词
	Example    string   `json:"example,omitempty"`  // 例句
}

type Meaning struct {
	PartOfSpeech string       `json:"partOfSpeech"`       // 词性
	Definitions  []Definition `json:"definitions"`        // 定义
	Synonyms     []string     `json:"synonyms,omitempty"` // 近义词
	Antonyms     []string     `json:"antonyms,omitempty"` // 反义词
}

type DictionaryEntry struct {
	Word       string     `json:"word"`       // 单词
	Phonetics  []Phonetic `json:"phonetics"`  // 音标
	Meanings   []Meaning  `json:"meanings"`   // 意思
	License    License    `json:"license"`    // 许可证
	SourceURLs []string   `json:"sourceUrls"` // 来源链接
}

func Query(ctx context.Context, word string) (entry *DictionaryEntry, err error) {
	url := freeDictionaryApi + word
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		logx.Errorf("freedictionary build req err: %v", err)
		return nil, err
	}
	resp, err := httpClient.Do(req)
	if err != nil {
		logx.Errorf("freedictionary Query err: %v", err)
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		logx.Errorf("freedictionary Query err: %v", err)
		return nil, err
	}
	var entries []DictionaryEntry
	err = json.Unmarshal(body, &entries)
	if err != nil {
		logx.Errorf("freedictionary Query err: %v", err)
		return nil, err
	}
	if len(entries) == 0 {
		return nil, errors.New("no entry found")
	}
	return &entries[0], nil
}

// 解析取得英音和美音的音标及发音。uk-英音音标 uka-英音发音 us-美音音标 usa-美音发音
func (d *DictionaryEntry) GetPronunciation() (uk, uka, us, usa string) {
	for _, p := range d.Phonetics {
		if strings.Contains(p.Audio, "-us.mp3") {
			us = d.phoneticImprove(p.Text)
			usa = p.Audio
		} else if strings.Contains(p.Audio, "-uk.mp3") {
			uk = d.phoneticImprove(p.Text)
			uka = p.Audio
		}
	}
	return
}

// 音标改进
func (d *DictionaryEntry) phoneticImprove(p string) string {
	// TODO 进行语音音标改为音位音标
	return p
}
