// Copyright 2019 orivil.com. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found at https://mit-license.org.

package statistics

import (
	"github.com/orivil/wechat"
	"time"
)

type Summary struct {
	RefDate    string `json:"ref_date"`
	UserSource int    `json:"user_source"`
	NewUser    int    `json:"new_user"`
	CancelUser int    `json:"cancel_user"`
}

type Cumulate struct {
	RefDate      string `json:"ref_date"`
	CumulateUser int    `json:"cumulate_user"`
}

type summaryList struct {
	List []*Summary `json:"list"`
}

type cumulateList struct {
	List []*Cumulate `json:"list"`
}

// 获取用户增减数据
func GetUserSummary(token string, beginDate, endDate time.Time) ([]*Summary, error) {
	URL := "https://api.weixin.qq.com/datacube/getusersummary?access_token=" + token
	data := map[string]string{
		"begin_date": beginDate.Format("2006-01-02"),
		"end_date":   endDate.Format("2006-01-02"),
	}
	res := &summaryList{}
	err := wechat.PostSchema(wechat.KindJson, URL, data, res)
	if err != nil {
		return nil, err
	} else {
		return res.List, nil
	}
}

// 获取累计用户数据
func GetUserCumulate(token string, beginDate, endDate time.Time) ([]*Cumulate, error) {
	URL := "https://api.weixin.qq.com/datacube/getusercumulate?access_token=" + token
	data := map[string]string{
		"begin_date": beginDate.Format("2006-01-02"),
		"end_date":   endDate.Format("2006-01-02"),
	}
	res := &cumulateList{}
	err := wechat.PostSchema(wechat.KindJson, URL, data, res)
	if err != nil {
		return nil, err
	} else {
		return res.List, nil
	}
}
