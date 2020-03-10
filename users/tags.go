// Copyright 2019 orivil.com. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found at https://mit-license.org.

package users

import (
	"github.com/orivil/wechat"
)

// 用户标签接口, see: https://mp.weixin.qq.com/wiki?t=resource/res_main&id=mp1421140837

// 创建标签
func CreateTag(name, token string) (tag *Tag, err error) {
	uri := "https://api.weixin.qq.com/cgi-bin/tags/create?access_token=" + token
	data := map[string]map[string]string{
		"tag": {
			"name": name,
		},
	}
	tag = &Tag{}
	err = wechat.PostSchema(wechat.KindJson, uri, data, tag)
	if err != nil {
		return nil, err
	} else {
		return tag, nil
	}
}

type Tag struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Count int    `json:"count"`
}

// 获取标签
func GetTags(token string) (tags []*Tag, err error) {
	uri := "https://api.weixin.qq.com/cgi-bin/tags/get?access_token=" + token
	res := &struct {
		Tags []*Tag `json:"tags"`
	}{}
	err = wechat.GetJson(uri, res)
	if err != nil {
		return nil, err
	} else {
		return res.Tags, nil
	}
}

// 修改标签
func UpdateTag(tag *Tag, token string) error {
	uri := "https://api.weixin.qq.com/cgi-bin/tags/update?access_token=" + token
	return wechat.PostSchema(wechat.KindJson, uri, map[string]*Tag{"tag": tag}, nil)
}

// 删除标签
func DeleteTag(tag *Tag, token string) error {
	uri := "https://api.weixin.qq.com/cgi-bin/tags/delete?access_token=" + token
	return wechat.PostSchema(wechat.KindJson, uri, map[string]*Tag{"tag": tag}, nil)
}

// 获取标签下粉丝列表
// nextOpenid 为第一个拉取的OPENID，不填默认从头开始拉取
func GetTagUsers(tagID int, nextOpenid, token string) (res *Users, err error) {
	uri := "https://api.weixin.qq.com/cgi-bin/user/tag/get?access_token=" + token
	data := map[string]interface{}{"tagid": tagID}
	if nextOpenid != "" {
		data["next_openid"] = nextOpenid
	}
	res = &Users{}
	err = wechat.PostSchema(wechat.KindJson, uri, data, res)
	if err != nil {
		return nil, err
	} else {
		return res, nil
	}
}

// 批量设置用户标签
func TagUsers(tagID int, openids []string, token string) error {
	uri := "https://api.weixin.qq.com/cgi-bin/tags/members/batchtagging?access_token=" + token
	params := map[string]interface{}{
		"tagid":       tagID,
		"openid_list": openids,
	}
	return wechat.PostSchema(wechat.KindJson, uri, params, nil)
}

// 取消用户标签
func UntagUsers(tagID int, openids []string, token string) error {
	uri := "https://api.weixin.qq.com/cgi-bin/tags/members/batchuntagging?access_token=" + token
	params := map[string]interface{}{
		"tagid":       tagID,
		"openid_list": openids,
	}
	return wechat.PostSchema(wechat.KindJson, uri, params, nil)
}

// 获取用户所有标签
func GetUserTags(openid, token string) (tagIDs []int, err error) {
	uri := "https://api.weixin.qq.com/cgi-bin/tags/getidlist?access_token=" + token
	res := &struct {
		TagIDList []int `json:"tagid_list"`
	}{}
	err = wechat.PostSchema(wechat.KindJson, uri, map[string]string{"openid": openid}, res)
	if err != nil {
		return nil, err
	} else {
		return res.TagIDList, nil
	}
}
