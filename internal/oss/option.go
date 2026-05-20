package oss

type UploadOption struct {
	ContentType string
}

type UploadOptions func(*UploadOption)

func WithContentType(contentType string) UploadOptions {
	return func(o *UploadOption) {
		o.ContentType = contentType
	}
}
