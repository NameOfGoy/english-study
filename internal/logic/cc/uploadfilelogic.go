package cc

import (
	"context"
	"fmt"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"english-study/internal/errors"
	"english-study/internal/svc"
	"english-study/internal/types"
	"english-study/internal/utils"

	"github.com/zeromicro/go-zero/core/logx"
)

const (
	// AI 桥上传单文件上限 20MB (跟 file-service 一致)
	ccMaxUploadBytes = 20 << 20
	// 下载 token 有效期: 24h. 够 CC 处理(通常秒级); 又不像公开读那样永久, 满足"临时链接"语义.
	ccFileTokenTTLSec = 24 * 3600
)

// ccAllowedMIME AI 桥上传 MIME 白名单 (基于内容嗅探, 非 client claim).
// 注意: http.DetectContentType 对 .xlsx(zip 容器) 返回 application/zip, 单独靠扩展名兜底放行(见 isAllowedCCUpload).
var ccAllowedMIME = map[string]bool{
	"text/plain; charset=utf-8": true, // txt / csv (csv 实测也 detect 成 text/plain)
	"text/csv; charset=utf-8":   true,
	"application/pdf":           true,
	"image/jpeg":                true,
	"image/png":                 true,
	"image/gif":                 true,
	"image/webp":                true,
	"audio/mpeg":                true,
	"audio/wav":                 true,
	"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet": true,
}

// isAllowedCCUpload: 内容嗅探类型在白名单内, 或 (zip 容器 且 文件名 .xlsx) — 后者专门放行 Excel.
func isAllowedCCUpload(detected, filename string) bool {
	if ccAllowedMIME[detected] {
		return true
	}
	if detected == "application/zip" && strings.HasSuffix(strings.ToLower(filename), ".xlsx") {
		return true // xlsx 是 zip 容器, DetectContentType 看不出是 xlsx, 用扩展名兜底
	}
	return false
}

type UploadFileLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
	ui     *utils.UserInfo
}

func NewUploadFileLogic(ctx context.Context, svcCtx *svc.ServiceContext, ui *utils.UserInfo) *UploadFileLogic {
	return &UploadFileLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
		ui:     ui,
	}
}

// UploadFileResult 上传结果 (手写 handler 直接 JSON 回, 不走 goctl types).
// 必须内嵌 CommonReply: 前端 request.js 拦截器靠 data.code===0 判成功, 不带 code 会被当失败 reject.
type UploadFileResult struct {
	types.CommonReply        // 成功时 Code 零值=0 → 前端认成功
	FileURL           string `json:"file_url"`   // 相对下载地址 /api/v1/cc/download?t=<token>, 前端拼 origin 后放进 envelope.metadata
	FileName          string `json:"file_name"`  // 原始文件名 (UTF-8, 供前端/CC 显示)
	FileMime          string `json:"file_mime"`  // 内容嗅探出的真实 MIME
	FileSize          int64  `json:"file_size"`
	ExpiresAt         int64  `json:"expires_at"` // 下载 token 过期 unix sec
}

// UploadFile: 仅 admin; 上传到私有桶 cc-uploads; 返回签名下载地址.
func (l *UploadFileLogic) UploadFile(file multipart.File, fh *multipart.FileHeader) (*UploadFileResult, error) {
	if err := utils.RequireAdmin(l.ui); err != nil {
		return nil, errors.ErrorPermissionError("仅管理员可上传 AI 辅助文件")
	}
	if fh.Size > ccMaxUploadBytes {
		return nil, errors.ErrorRequestParamError("文件大小超过 20MB 限制")
	}

	// MIME 内容嗅探 (不信任 client Content-Type)
	sniff := make([]byte, 512)
	n, _ := file.Read(sniff)
	detected := http.DetectContentType(sniff[:n])
	if !isAllowedCCUpload(detected, fh.Filename) {
		return nil, errors.ErrorRequestParamError("不支持的文件类型: " + detected)
	}
	// 重置 reader 到开头, 否则上传会少前 512 字节
	if seeker, ok := file.(interface {
		Seek(offset int64, whence int) (int64, error)
	}); ok {
		if _, sErr := seeker.Seek(0, 0); sErr != nil {
			return nil, errors.ErrorDataUploadError("重置上传文件指针失败").WithCause(sErr)
		}
	}

	// 确保私有桶存在 (幂等; 默认私有不公开读)
	if err := l.svcCtx.Oss.EnsureBucket(l.ctx, types.CCUploadBucket); err != nil {
		return nil, errors.ErrorDataUploadError("初始化存储桶失败").WithCause(err)
	}

	// object key: cc/<uid>/<uuid>/<safe-name>; uuid 让 key 不可枚举, safe-name 仅 ASCII 安全字符
	uuid, err := newJTI()
	if err != nil {
		return nil, errors.ErrorDataUploadError("生成对象 id 失败").WithCause(err)
	}
	safeName := sanitizeObjectName(fh.Filename)
	object := fmt.Sprintf("cc/%d/%s/%s", l.ui.ID, uuid, safeName)

	if _, err := l.svcCtx.Oss.Upload(l.ctx, types.CCUploadBucket, object, file, fh.Size); err != nil {
		return nil, errors.ErrorDataUploadError("文件上传失败").WithCause(err)
	}

	// 签发下载 token + 拼相对下载地址
	iat := time.Now().Unix()
	tok, err := utils.GenerateFileToken(l.svcCtx.Config.Auth.AccessSecret, types.CCUploadBucket, object, iat, ccFileTokenTTLSec)
	if err != nil {
		return nil, errors.ErrorTokenGenerateError("签发下载 token 失败").WithCause(err)
	}

	return &UploadFileResult{
		FileURL:   "/api/v1/cc/download?t=" + tok,
		FileName:  fh.Filename,
		FileMime:  detected,
		FileSize:  fh.Size,
		ExpiresAt: iat + ccFileTokenTTLSec,
	}, nil
}

// sanitizeObjectName 把原始文件名压成 object key 安全片段: 只留 [a-zA-Z0-9._-], 其余→_; 取 basename 防穿越.
// 原始名仍通过 FileName 字段保留, 这里只为存储 key 安全 (CC 落盘时会再次 sanitize).
func sanitizeObjectName(name string) string {
	name = filepath.Base(name)
	var b strings.Builder
	for _, r := range name {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9', r == '.', r == '_', r == '-':
			b.WriteRune(r)
		default:
			b.WriteByte('_')
		}
	}
	s := strings.Trim(b.String(), ".")
	if s == "" || s == "_" {
		s = "file"
	}
	return s
}
