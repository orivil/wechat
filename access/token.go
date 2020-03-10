// Copyright 2018 orivil.com. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found at https://mit-license.org.

package access

import (
	"github.com/orivil/wechat"
)

type Token struct {
	Value     string `json:"access_token"`
	ExpiresIn int64  `json:"expires_in"`
}

type Ticket struct {
	Value     string `json:"ticket"`
	ExpiresIn int64  `json:"expires_in"`
}

func GetAccessToken(appid, appSecret string) (token *Token, err error) {
	uri := "https://api.weixin.qq.com/cgi-bin/token?grant_type=client_credential&appid=" + appid + "&secret=" + appSecret
	token = &Token{}
	err = wechat.GetJson(uri, token)
	if err != nil {
		return nil, err
	} else {
		return token, nil
	}
}

func GetTicket(accessToken string) (ticket *Ticket, err error) {
	uri := "https://api.weixin.qq.com/cgi-bin/ticket/getticket?type=jsapi&access_token=" + accessToken
	ticket = &Ticket{}
	err = wechat.GetJson(uri, ticket)
	if err != nil {
		return nil, err
	} else {
		return ticket, nil
	}
}
