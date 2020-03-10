// Copyright 2018 orivil.com. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found at https://mit-license.org.

// migrate 包用于迁移已订阅用户
package users

import (
	"github.com/orivil/wechat"
)

// 获得已订阅用户列表
func GetSubscribers(token string, walk func(openids []string) error) error {
	var nextOpenid string
	for {
		users, err := GetNextSubscribers(token, nextOpenid)
		if err != nil {
			return err
		} else {
			err = walk(users.Data.Openid)
			if err != nil {
				return err
			} else {
				if users.NextOpenid == "" {
					return nil
				} else {
					nextOpenid = users.NextOpenid
				}
			}
		}
	}
}

func GetNextSubscribers(token, nextOpenID string) (users *Users, err error) {
	uri := "https://api.weixin.qq.com/cgi-bin/user/get?access_token=" + token
	if nextOpenID != "" {
		uri += "&next_openid=" + nextOpenID
	}
	users = &Users{}
	err = wechat.GetJson(uri, users)
	if err != nil {
		return nil, err
	} else {
		return users, nil
	}
}

type Users struct {
	Total      int    `json:"total"`
	Count      int    `json:"count"`
	Data       Data   `json:"data"`
	NextOpenid string `json:"next_openid"`
}

type Data struct {
	Openid []string `json:"openid"`
}
