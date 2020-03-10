// Copyright 2018 orivil.com. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found at https://mit-license.org.

package account

import (
	"github.com/orivil/wechat"
)

type response struct {
	ShortUrl string `json:"short_url"`
}

// 获得短连接
func GetShortUrl(token, longUrl string) (shortUrl string, err error) {
	data := map[string]interface{}{
		"action":   "long2short",
		"long_url": longUrl,
	}
	url := "https://api.weixin.qq.com/cgi-bin/shorturl?access_token=" + token
	var res *response
	err = wechat.PostSchema(wechat.KindJson, url, data, &res)
	if err != nil {
		return "", err
	} else {
		return res.ShortUrl, nil
	}
}
