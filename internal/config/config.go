package config

import "github.com/zeromicro/go-zero/rest"

type Config struct {
	rest.RestConf
	Auth struct {
		AccessSecret string
		AccessExpire int64
	}
	Postgresql struct {
		DSN string
	}
	Minio struct {
		Endpoint  string
		AccessKey string
		SecretKey string
		UseSSL    bool
		Domain    string
		Bucket    string
	}
	WxApp struct {
		AppID     string
		AppSecret string
	}
	AliCloud struct {
		AccessKeyId     string
		AccessKeySecret string
		Region          string
		AppKey          string
	}
	BigModel struct {
		APIKey string
	}
}
