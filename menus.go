// Copyright 2018 orivil.com. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found at https://mit-license.org.

package wechat

type Menus struct {
	Buttons []*MenuButton `json:"button"`
}

// 生成公众号菜单
func GenerateMenus(accessToken string, ms *Menus) error {
	// 提交到公众号平台
	u := "https://api.weixin.qq.com/cgi-bin/menu/create?access_token=" + accessToken
	return PostSchema(KindJson, u, ms, nil)
}

type MenuButton struct {
	// 菜单的响应动作类型，view表示网页类型，click表示点击类型，miniprogram表示小程序类型
	Type string `json:"type"`

	// 	菜单标题，不超过16个字节，子菜单不超过60个字节
	Name string `json:"name"`

	// click等点击类型必须	菜单KEY值，用于消息接口推送，不超过128字节
	Key string `json:"key"`

	// 二级菜单数组，个数应为1~5个
	SubButtons []*MenuButton `json:"sub_button"`

	// view、miniprogram类型必须	网页 链接，用户点击菜单可打开链接，不超过1024字节。
	// type为miniprogram时，不支持小程序的老版本客户端将打开本url。
	Url string `json:"url"`

	// media_id类型和view_limited类型必须, 调用新增永久素材接口返回的合法media_id
	MediaID string `json:"media_id"`

	// 小程序的appid（仅认证公众号可配置）
	AppID string `json:"appid"`

	// 小程序的页面路径
	PagePath string `json:"pagepath"`
}
