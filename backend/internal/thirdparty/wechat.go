package thirdparty

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"cs-assistant-backend/config"
)

const code2SessionURL = "https://api.weixin.qq.com/sns/jscode2session"

// Code2SessionResp 微信 code2Session 接口返回
type Code2SessionResp struct {
	OpenID     string `json:"openid"`
	SessionKey string `json:"session_key"`
	UnionID    string `json:"unionid"`
	ErrCode    int    `json:"errcode"`
	ErrMsg     string `json:"errmsg"`
}

// IsOK 返回是否调用成功
func (r Code2SessionResp) IsOK() bool {
	return r.ErrCode == 0 && r.OpenID != ""
}

// ExchangeCode 用临时 code 换取 openid
func ExchangeCode(cfg config.WechatConfig, code string) (*Code2SessionResp, error) {
	u, _ := url.Parse(code2SessionURL)
	q := url.Values{}
	q.Set("appid", cfg.AppID)
	q.Set("secret", cfg.AppSecret)
	q.Set("js_code", code)
	q.Set("grant_type", "authorization_code")
	u.RawQuery = q.Encode()

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(u.String())
	if err != nil {
		return nil, fmt.Errorf("call code2session: %w", err)
	}
	defer resp.Body.Close()

	var result Code2SessionResp
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode code2session resp: %w", err)
	}

	if !result.IsOK() {
		return nil, fmt.Errorf("wechat error [%d]: %s", result.ErrCode, result.ErrMsg)
	}

	return &result, nil
}
