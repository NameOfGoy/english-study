package file

import (
	"context"
	"english-study/internal/errors"
	"english-study/internal/utils"
	"mime/multipart"
	"net/http"

	"english-study/internal/svc"
	"english-study/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

const (
	// 单文件上限 20MB. 词典图片 / 头像 / 例句音频, 这个量级足够
	maxUploadBytes = 20 << 20
)

// allowedUploadMIMETypes 允许上传的 MIME 类型白名单 (基于内容嗅探, 非 client claim)
var allowedUploadMIMETypes = map[string]bool{
	"image/jpeg":          true,
	"image/png":           true,
	"image/webp":          true,
	"image/gif":           true,
	"audio/mpeg":          true,
	"audio/mp3":           true, // 同 mpeg, 历史别名
	"audio/wav":           true,
	"application/pdf":     true,
	"text/csv; charset=utf-8": true, // 词单导入
	"text/plain; charset=utf-8": true,
	"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet": true, // .xlsx
}

type UploadLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 文件模块/文件上传
func NewUploadLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UploadLogic {
	return &UploadLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UploadLogic) Upload(req *types.FileUploadReq, file multipart.File, fileHeader *multipart.FileHeader) (resp *types.FileUploadResp, err error) {
	if req.Bucket == "" {
		req.Bucket = types.OssBucket
	}
	if fileHeader.Size > maxUploadBytes {
		return nil, errors.ErrorRequestParamError("文件大小超过 20MB 限制")
	}
	// MIME 嗅探: 不信任 client 的 Content-Type, 只看前 512 字节真实内容
	sniff := make([]byte, 512)
	n, _ := file.Read(sniff)
	detected := http.DetectContentType(sniff[:n])
	if !allowedUploadMIMETypes[detected] {
		return nil, errors.ErrorRequestParamError("不支持的文件类型: " + detected)
	}
	// 关键: 重置 reader 到开头, 否则后面 Upload 写入会少前 512 字节
	if seeker, ok := file.(interface {
		Seek(offset int64, whence int) (int64, error)
	}); ok {
		if _, sErr := seeker.Seek(0, 0); sErr != nil {
			return nil, errors.ErrorDataUploadError("重置上传文件指针失败").WithCause(sErr)
		}
	}

	path, err := l.svcCtx.Oss.Upload(l.ctx, req.Bucket, req.Object, file, fileHeader.Size)
	if err != nil {
		return nil, errors.ErrorDataUploadError("文件上传失败").WithCause(err)
	}

	return &types.FileUploadResp{
		Path: utils.ToOssUri(req.Bucket, path),
	}, nil
}
