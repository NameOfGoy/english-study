package file

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"english-study/internal/svc"
	"english-study/internal/types"
	"english-study/internal/utils"

	"github.com/zeromicro/go-zero/core/logx"
)

// 必应图片爬取 — 复用连接池 + 8s 超时（之前每次请求都 new http.Client{} 且无超时）
var bingSearchHTTPClient = &http.Client{Timeout: 8 * time.Second}

type SearchImagesLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
	ui     *utils.UserInfo
}

func NewSearchImagesLogic(ctx context.Context, svcCtx *svc.ServiceContext, ui *utils.UserInfo) *SearchImagesLogic {
	return &SearchImagesLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
		ui:     ui,
	}
}

// bingImageResult represents a single image from Bing's response
type bingImageResult struct {
	MUrl   string `json:"murl"`   // media url (full image)
	TUrl   string `json:"turl"`   // thumbnail url
	PUrl   string `json:"purl"`   // page url
	Width  int    `json:"width"`  // image width (from metadata)
	Height int    `json:"height"` // image height (from metadata)
}

func (l *SearchImagesLogic) SearchImages(req *types.SearchImagesReq) (resp *types.SearchImagesResp, err error) {
	if strings.TrimSpace(req.Query) == "" {
		return &types.SearchImagesResp{
			Images: []types.ImageResult{},
		}, nil
	}

	count := req.Count
	if count <= 0 {
		count = 20
	}
	if count > 50 {
		count = 50
	}
	offset := req.Offset
	if offset < 0 {
		offset = 0
	}

	images, err := l.searchBingImages(req.Query, offset, count)
	if err != nil {
		l.Errorf("搜索图片失败: %v", err)
		return &types.SearchImagesResp{
			Images: []types.ImageResult{},
		}, nil
	}

	return &types.SearchImagesResp{
		Images: images,
	}, nil
}

func (l *SearchImagesLogic) searchBingImages(query string, offset, count int) ([]types.ImageResult, error) {
	// Use Bing image search API (undocumented JSON endpoint)
	encodedQuery := url.QueryEscape(query)
	apiURL := fmt.Sprintf(
		"https://www.bing.com/images/async?q=%s&first=%d&count=%d&qft=+filterui:photo-photo&SFX=1",
		encodedQuery, offset, count,
	)

	req, err := http.NewRequestWithContext(l.ctx, http.MethodGet, apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")

	httpResp, err := bingSearchHTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求必应失败: %w", err)
	}
	defer httpResp.Body.Close()

	// Bing 异步 HTML 4MB 已绰绰有余, 防止上游异常返回 GB 级 padding
	body, err := io.ReadAll(io.LimitReader(httpResp.Body, 4<<20))
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	return parseBingHTML(string(body)), nil
}

// parseBingHTML extracts image URLs from Bing's async HTML response
// Bing returns HTML with <a class="iusc" m="{json}"> tags containing image metadata
func parseBingHTML(html string) []types.ImageResult {
	var results []types.ImageResult

	// Find all m="{...}" attributes in iusc elements
	searchStr := `m="{`
	pos := 0
	for {
		idx := strings.Index(html[pos:], searchStr)
		if idx == -1 {
			break
		}
		pos += idx + len(searchStr) - 1 // position at the opening {

		// Find the closing "
		endIdx := strings.Index(html[pos:], `"`)
		if endIdx == -1 {
			break
		}

		jsonStr := html[pos : pos+endIdx]
		// Unescape HTML entities
		jsonStr = strings.ReplaceAll(jsonStr, "&quot;", `"`)
		jsonStr = strings.ReplaceAll(jsonStr, "&amp;", "&")
		jsonStr = strings.ReplaceAll(jsonStr, "&#39;", "'")

		var imgData bingImageResult
		if err := json.Unmarshal([]byte(jsonStr), &imgData); err == nil {
			if imgData.MUrl != "" {
				result := types.ImageResult{
					Url:      imgData.MUrl,
					ThumbUrl: imgData.TUrl,
					Width:    imgData.Width,
					Height:   imgData.Height,
				}
				if result.ThumbUrl == "" {
					result.ThumbUrl = result.Url
				}
				results = append(results, result)
			}
		}

		pos += endIdx + 1
	}

	return results
}
