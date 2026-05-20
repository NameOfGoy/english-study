package utils

import (
	"english-study/internal/types"
	"github.com/go-kratos/kratos/v2/errors"
)

// WrapResponse 包装响应
func WrapResponse(resp any, err error) (wrap any) {
	if err != nil {
		var e *errors.Error
		if errors.As(err, &e) {
			var reason string
			if ue := e.Unwrap(); ue != nil {
				reason = ue.Error()
			}
			return types.CommonReply{
				Code:   e.GetCode(),
				Msg:    e.GetMessage(),
				Reason: reason, // kratos的错误reason不过是错误码的字符串. 具体错误信息在cause中
			}
		} else {
			return types.CommonReply{
				Code:   500,
				Msg:    "未知错误",
				Reason: err.Error(),
			}
		}
	}
	return resp
}
