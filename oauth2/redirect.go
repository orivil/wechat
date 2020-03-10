// Copyright 2018 orivil.com. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found at https://mit-license.org.

package oauth2

import (
	"github.com/google/go-querystring/query"
	"net/http"
)

// 初始化微信登陆地址, 该地址仅用于微信内登录
// scope 分两种, 一种是 "snsapi_base", 只能根据返回的 code 换取用户的 openid, 另一种是 "snsapi_userinfo", 能够根据返回的
// code 换取用户的所有信息.
// redirectUrl 是用户授权后所跳转的地址, 用户授权后会跳转到该地址并添加上 code 及 state 参数, 如果用户取消授权则只会获得 state 参数
// appid 是授权公众号的 appid
// componentAppid 是开放平台 appid
// state 为用户自定义参数, 通常用于防止跨站伪造请求
func InitAppRedirect(scope, redirectUrl, appid, componentAppid, state string) string {
	u, _ := query.Values(&redirectConfig{
		AppID:          appid,
		RedirectUri:    redirectUrl,
		ResponseType:   "code",
		Scope:          scope,
		State:          state,
		ComponentAppid: componentAppid,
	})
	return "https://open.weixin.qq.com/connect/oauth2/authorize?" + u.Encode() + "#wechat_redirect"
}

// 初始化微信登录地址, 该地址主要用于浏览器扫描登录, 需要在开放平台创建网站应用
func InitBrowserRedirect(redirectUrl, appid, state string) string {
	u, _ := query.Values(&redirectConfig{
		AppID:        appid,
		RedirectUri:  redirectUrl,
		ResponseType: "code",
		Scope:        ScopeApiLogin,
		State:        state,
	})
	return "https://open.weixin.qq.com/connect/qrconnect?" + u.Encode() + "#wechat_redirect"
}

// 用户授权后跳转至指定的 URI 时会带上, code 及 state 参数, 如果用户未同一授权, 则只会获得 state 参数
func GetUriCode(req *http.Request) (code, state string) {
	q := req.URL.Query()
	return q.Get("code"), q.Get("state")
}

type redirectConfig struct {
	AppID          string `url:"appid"`
	RedirectUri    string `url:"redirect_uri"`
	ResponseType   string `url:"response_type"`
	Scope          string `url:"scope"`
	State          string `url:"state"`
	ComponentAppid string `url:"component_appid,omitempty"`
}
