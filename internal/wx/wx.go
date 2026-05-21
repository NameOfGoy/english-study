package wx

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const (
	WxHost = "https://api.weixin.qq.com"
)

// httpClient 复用连接池 + 10 秒超时，避免默认 http.Get 永不超时把 goroutine 挂死。
var httpClient = &http.Client{Timeout: 10 * time.Second}

type Client struct {
	appID       string
	appSecret   string
	accessToken string
	expiresIn   int64
}

func NewWxClient(appID, appSecret string) *Client {
	return &Client{
		appID:     appID,
		appSecret: appSecret,
	}
}

func (c *Client) GetAccessToken() error {
	path := fmt.Sprintf("/cgi-bin/token?grant_type=client_credential&appid=%s&secret=%s", c.appID, c.appSecret)
	resp := &struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int64  `json:"expires_in"`
	}{}
	rsp, err := httpClient.Get(WxHost + path)
	if err != nil {
		return err
	}
	defer rsp.Body.Close()
	// 微信接口响应通常 < 1KB, 设 64KB 上限是充足且偏紧的兜底
	body, err := io.ReadAll(io.LimitReader(rsp.Body, 64<<10))
	if err != nil {
		return err
	}
	err = json.Unmarshal(body, resp)
	if err != nil {
		return err
	}
	c.accessToken = resp.AccessToken
	c.expiresIn = time.Now().Unix() + resp.ExpiresIn
	return nil
}

func (c *Client) Code2Session(code string) (openid string, sessionKey string, err error) {
	path := fmt.Sprintf("/sns/jscode2session?appid=%s&secret=%s&js_code=%s&grant_type=authorization_code", c.appID, c.appSecret, code)
	resp := &struct {
		OpenID     string `json:"openid"`
		SessionKey string `json:"session_key"`
	}{}
	rsp, err := httpClient.Get(WxHost + path)
	if err != nil {
		return "", "", err
	}
	defer rsp.Body.Close()
	// 微信接口响应通常 < 1KB, 设 64KB 上限是充足且偏紧的兜底
	body, err := io.ReadAll(io.LimitReader(rsp.Body, 64<<10))
	if err != nil {
		return "", "", err
	}
	err = json.Unmarshal(body, resp)
	if err != nil {
		return "", "", err
	}
	return resp.OpenID, resp.SessionKey, nil
}
