package minio

import (
	"context"
	"english-study/internal/oss"
	"english-study/internal/utils"
	"fmt"
	"io"
	"mime"
	"path/filepath"
	"strings"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type Minio struct {
	client *minio.Client
}

func NewMinio(endpoint, accessKey, secretKey string, useSSL bool) (*Minio, error) {
	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		return nil, err
	}
	return &Minio{client: client}, nil
}

func (m *Minio) Upload(ctx context.Context, bucket string, object string, data io.ReadCloser, size int64, opts ...oss.UploadOptions) (string, error) {
	opt := utils.GetOptionFromOptions[oss.UploadOption, oss.UploadOptions](opts, func() oss.UploadOption {
		return oss.UploadOption{
			ContentType: "application/octet-stream",
		}
	})
	// 检查bucket是否存在 (BucketExists 返回 (bool,err); 之前丢了 bool 导致桶不存在时不报错绕一圈)
	exists, err := m.client.BucketExists(ctx, bucket)
	if err != nil {
		return "", err
	}
	if !exists {
		return "", fmt.Errorf("bucket %q does not exist", bucket)
	}

	var contentType string
	if opt.ContentType != "" {
		contentType = opt.ContentType
	} else {
		// 解析文件后缀获取Content-Type
		ext := strings.ToLower(filepath.Ext(object)) // 获取扩展名并转为小写
		contentType = mime.TypeByExtension(ext)

		// 如果无法识别类型，使用默认二进制流类型:cite[3]:cite[6]
		if contentType == "" {
			contentType = "application/octet-stream"
		}
	}

	_, err = m.client.PutObject(ctx, bucket, object, data, size, minio.PutObjectOptions{ContentType: contentType})
	if err != nil {
		return "", err
	}
	return object, nil
}

// EnsureBucket 确保 bucket 存在; 不存在则创建. 新建 bucket 默认私有 (MinIO 默认无匿名访问策略),
// 不调用 SetBucketPolicy, 所以匿名 GET 会被拒 —— 这正是 AI 桥私有桶想要的.
func (m *Minio) EnsureBucket(ctx context.Context, bucket string) error {
	exists, err := m.client.BucketExists(ctx, bucket)
	if err != nil {
		return err
	}
	if exists {
		return nil
	}
	if err := m.client.MakeBucket(ctx, bucket, minio.MakeBucketOptions{}); err != nil {
		// 并发首次创建: 另一个请求已建好同名桶, 忽略这种竞态错误
		code := minio.ToErrorResponse(err).Code
		if code == "BucketAlreadyOwnedByYou" || code == "BucketAlreadyExists" {
			return nil
		}
		return err
	}
	return nil
}

func (m *Minio) Download(ctx context.Context, bucket string, object string) (io.ReadCloser, error) {
	// 检查bucket是否存在
	_, err := m.client.BucketExists(ctx, bucket)
	if err != nil {
		return nil, err
	}

	// 检查object是否存在
	_, err = m.client.StatObject(ctx, bucket, object, minio.StatObjectOptions{})
	if err != nil {
		return nil, err
	}

	// 下载文件
	reader, err := m.client.GetObject(ctx, bucket, object, minio.GetObjectOptions{})
	if err != nil {
		return nil, err
	}
	return reader, nil
}
