package file

import (
	"context"
	"english-study/internal/errors"
	"english-study/internal/utils"
	"mime/multipart"

	"english-study/internal/svc"
	"english-study/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

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
	// 上传文件
	path, err := l.svcCtx.Oss.Upload(l.ctx, req.Bucket, req.Object, file, fileHeader.Size)
	if err != nil {
		return nil, errors.ErrorDataUploadError("文件上传失败").WithCause(err)
	}

	// 返回文件路径
	return &types.FileUploadResp{
		Path: utils.ToOssUri(req.Bucket, path),
	}, nil
}
