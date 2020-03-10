// Copyright 2018 orivil.com. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found at https://mit-license.org.

package oauth2

import (
	"github.com/google/go-querystring/query"
	"github.com/orivil/wechat"
)

// 公众号获得 access token, access token 包含了用户的 openid
// 有 IP 白名单限制
func GetAccessToken(appid, secret, code string) (token *AccessToken, err error) {
	u, _ := query.Values(&config{
		AppID:     appid,
		Secret:    secret,
		Code:      code,
		GrantType: "authorization_code",
	})
	token = &AccessToken{}
	err = wechat.GetJson("https://api.weixin.qq.com/sns/oauth2/access_token?"+u.Encode(), token)
	if err != nil {
		return nil, err
	} else {
		return token, nil
	}
}

type config struct {
	AppID                string `url:"appid,omitempty"`
	Secret               string `url:"secret,omitempty"`
	Code                 string `url:"code,omitempty"`
	GrantType            string `url:"grant_type,omitempty"`
	ComponentAppid       string `url:"component_appid,omitempty"`
	ComponentAccessToken string `url:"component_access_token,omitempty"`
}

// auth access token 与 公众号 access token 是两个不同的东西, auth access token 包含了用户的 openid
// 或根据 AccessToken 近一步获取更多用户信息
type AccessToken struct {
	AccessToken  string `json:"access_token"`
	ExpiresIn    int64  `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
	Openid       string `json:"openid"` // 一个公众号对应一个用户 openid, openid 是永久的
	Scope        string `json:"scope"`
}

// 第三方平台获得 access token, 有 IP 白名单限制
func GetComponentAccessToken(appid, code, componentAppid, componentAccessToken string) (token *AccessToken, err error) {
	u, _ := query.Values(&config{
		AppID:                appid,
		Code:                 code,
		GrantType:            "authorization_code",
		ComponentAppid:       componentAppid,
		ComponentAccessToken: componentAccessToken,
	})
	token = &AccessToken{}
	err = wechat.GetJson("https://api.weixin.qq.com/sns/oauth2/component/access_token?"+u.Encode(), token)
	if err != nil {
		return nil, err
	} else {
		return token, nil
	}
}

// 刷新access_token（如果需要）
//
// 由于access_token拥有较短的有效期，当access_token超时后，可以使用refresh_token进行刷新，
// refresh_token有效期为30天，当refresh_token失效之后，需要用户重新授权。
func RefreshAccessToken(appid, refreshToken string) (token *AccessToken, err error) {
	u, _ := query.Values(&refreshConfig{
		Appid:        appid,
		GrantType:    "refresh_token",
		RefreshToken: refreshToken,
	})
	token = &AccessToken{}
	err = wechat.GetJson("https://api.weixin.qq.com/sns/oauth2/refresh_token?"+u.Encode(), token)
	if err != nil {
		return nil, err
	} else {
		return token, nil
	}
}

// 三方平台刷新access_token（如果需要）
//
// 由于access_token拥有较短的有效期，当access_token超时后，可以使用refresh_token进行刷新，
// refresh_token拥有较长的有效期（30天），当refresh_token失效的后，需要用户重新授权。
func RefreshComponentAccessToken(appid, refreshToken, componentAppid, componentAccessToken string) (token *AccessToken, err error) {
	u, _ := query.Values(&refreshConfig{
		Appid:                appid,
		GrantType:            "refresh_token",
		RefreshToken:         refreshToken,
		ComponentAppid:       componentAppid,
		ComponentAccessToken: componentAccessToken,
	})
	token = &AccessToken{}
	err = wechat.GetJson("https://api.weixin.qq.com/sns/oauth2/component/refresh_token?"+u.Encode(), token)
	if err != nil {
		return nil, err
	} else {
		return token, nil
	}
}

type refreshConfig struct {
	Appid                string `url:"appid,omitempty"`
	GrantType            string `url:"grant_type,omitempty"`
	RefreshToken         string `url:"refresh_token,omitempty"`
	ComponentAppid       string `url:"component_appid,omitempty"`
	ComponentAccessToken string `url:"component_access_token,omitempty"`
}
