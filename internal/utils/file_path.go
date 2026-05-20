package utils

import (
	"fmt"
	"strings"
)

func ToOssUri(bucket string, object string) string {
	if object == "" {
		return ""
	}
	special := "http"
	l := len(special)
	if len(object) >= l && object[:l] == special { // 已经是oss路径, 直接返回
		return object
	}
	return fmt.Sprintf("/api/v1/file/%s/%s", bucket, object)
}

func ToOssPath(bucket string, uri string) string {
	if uri == "" {
		return ""
	}
	special := "http"
	l := len(special)
	if len(uri) >= l && uri[:l] == special { // 已经是成品http的路径, 直接返回
		return uri
	}
	parts := strings.SplitN(uri, "/api/v1/file/"+bucket+"/", 2)
	if len(parts) != 2 {
		return ""
	}
	return parts[1]
}
