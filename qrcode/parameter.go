// Copyright 2018 orivil.com. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found at https://mit-license.org.

// qrcode 包用于生成带参数二维码.
// 为了满足用户渠道推广分析和用户帐号绑定等场景的需要，公众平台提供了生成带参数二维码的接口。
// 使用该接口可以获得多个带不同场景值的二维码，用户扫描后，公众号可以接收到事件推送。
package qrcode

import (
	"github.com/orivil/wechat"
	"net/url"
)

type action string

/**
	目前有2种类型的二维码：

	1、临时二维码，是有过期时间的，最长可以设置为在二维码生成后的30天（即2592000秒）后过期，但能够生成较多数量。临时二维码主要用于帐号绑定等不要求二维码永久保存的业务场景
	2、永久二维码，是无过期时间的，但数量较少（目前为最多10万个）。永久二维码主要用于适用于帐号绑定、用户来源统计等场景。

	用户扫描带场景值二维码时，可能推送以下两种事件：

	如果用户还未关注公众号，则用户可以关注公众号，关注后微信会将带场景值关注事件推送给开发者。

	如果用户已经关注公众号，在用户扫描后会自动进入会话，微信也会将带场景值扫描事件推送给开发者。

	获取带参数的二维码的过程包括两步，首先创建二维码ticket，然后凭借ticket到指定URL换取二维码。
**/
const (
	QR_SCENE     action = "QR_SCENE"     // 临时的整型参数值
	QR_STR_SCENE        = "QR_STR_SCENE" // 临时的字符串参数值

	// 永久二维码，是无过期时间的，但数量较少（目前为最多10万个）。
	QR_LIMIT_SCENE     = "QR_LIMIT_SCENE"     // 永久的整型参数值
	QR_LIMIT_STR_SCENE = "QR_LIMIT_STR_SCENE" // 永久的字符串参数值
)

type ActionInfo struct {
	Scene Scene `json:"scene"`
}

type Scene struct {
	SceneID  int    `json:"scene_id"`  // 场景值ID，临时二维码时为32位非0整型，永久二维码时最大值为100000（目前参数只支持1--100000）
	SceneStr string `json:"scene_str"` // 场景值ID（字符串形式的ID），字符串类型，长度限制为1到64
}

type PostSchema struct {
	ExpireSeconds int64      `json:"expire_seconds"` // 临时二维码有效时间，以秒为单位。 最大不超过2592000（即30天），此字段如果不填，则默认有效期为30秒。
	ActionName    action     `json:"action_name"`
	ActionInfo    ActionInfo `json:"action_info"`
}

type Result struct {
	Ticket        string `json:"ticket"`         // 获取二维码ticket后，开发者可用ticket换取二维码图片。请注意，本接口无须登录态即可调用。
	ExpireSeconds int64  `json:"expire_seconds"` // 该二维码有效时间，以秒为单位。 最大不超过2592000（即30天）。
	Url           string `json:"url"`            // 二维码图片解析后的地址，开发者可根据该地址自行生成需要的二维码图片
}

func Generate(token string, schema *PostSchema) (res *Result, err error) {
	err = wechat.PostSchema(wechat.KindJson, "https://api.weixin.qq.com/cgi-bin/qrcode/create?access_token="+token, schema, &res)
	return
}

// 根据 ticket 获得二维码图片地址
func QRCodeImage(ticket string) string {
	return "https://mp.weixin.qq.com/cgi-bin/showqrcode?ticket=" + url.QueryEscape(ticket)
}
