// Copyright 2018 orivil.com. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found at https://mit-license.org.

package open_platform

import (
	"github.com/google/go-querystring/query"
	"net/http"
	"strconv"
)

type AuthOption struct {
	// 必填, 第三方平台方appid
	ComponentAppid string `url:"component_appid"`

	// 必填, 预授权码
	PreAuthCode string `url:"pre_auth_code"`

	// 必填, 回调URI
	RedirectUri string `url:"redirect_uri"`

	// 要授权的帐号类型， 1则商户扫码后，手机端仅展示公众号、2表示仅展示小程序，
	// 3表示公众号和小程序都展示。如果为未制定，则默认小程序和公众号都展示。
	// 第三方平台开发者可以使用本字段来控制授权的帐号类型。
	//
	// 注意: AuthType 与 BizAppid 互斥
	AuthType string `url:"auth_type"`

	// 指定授权唯一的小程序或公众号
	//
	// 注意: AuthType 与 BizAppid 互斥
	BizAppid string `url:"biz_appid"`
}

// 生成公众号授权地址
func NewAuthRedirectUrl(ao *AuthOption) string {
	vs, _ := query.Values(ao)
	return "https://mp.weixin.qq.com/cgi-bin/componentloginpage?" + vs.Encode()
}

// 授权成功之后会跳转到回调 Uri, 且带上 auth code 信息.
// 授权成功之后还会发送事件信息到服务器, 事件信息中也会带上 code 信息.
// code 有过期时间, 需要及时用于换取授权权限信息以及 accesss token.
func GetAuthCode(redirectRequest *http.Request) (code string, expireIn int64) {
	q := redirectRequest.URL.Query()
	code = q.Get("auth_code")
	expire := q.Get("expires_in")
	expireIn, _ = strconv.ParseInt(expire, 10, 64)
	return
}
