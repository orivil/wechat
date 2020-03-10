// Copyright 2018 orivil.com. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found at https://mit-license.org.

package oauth2

// 微信 APP 内授权 scope
const (
	// 该权限只能获得 access token, 其中包括了 openid
	ScopeBase = "snsapi_base"

	// 获得所有权限, 需要用户授权
	ScopeUserInfo = "snsapi_userinfo"
)

// 浏览器授权 scope
const (
	ScopeApiLogin = "snsapi_login"
)
