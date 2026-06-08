package oss

import (
	"context"
	"io"
)

type OSS interface {
	// Upload 上传文件
	Upload(ctx context.Context, bucket string, object string, reader io.ReadCloser, size int64, opts ...UploadOptions) (path string, err error)
	// Download 下载文件
	Download(ctx context.Context, bucket string, object string) (reader io.ReadCloser, err error)
	// EnsureBucket 确保 bucket 存在 (不存在则创建, 默认私有, 不设匿名公开读策略). 幂等.
	EnsureBucket(ctx context.Context, bucket string) error
}
