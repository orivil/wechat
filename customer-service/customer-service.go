// Copyright 2018 orivil.com. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found at https://mit-license.org.

// 客服服务
package customer_service

import (
	"encoding/json"
	"github.com/orivil/wechat"
	"io/ioutil"
	"net/http"
)

// 微信客服
type CustomerService struct {
	KfAccount    string `json:"kf_account"`
	KfNick       string `json:"kf_nick"`
	KfHeadImgUrl string `json:"kf_headimgurl"`
}

// 创建/修改/删除微信客服数据模型
type PostCS struct {
	KfAccount string `json:"kf_account"`
	Nickname  string `json:"nickname"`
	Password  string `json:"password"`
}

// 创建客服
func CreateCS(token string, cs *PostCS) error {
	return wechat.PostSchema(wechat.KindJson, "https://api.weixin.qq.com/customservice/kfaccount/add?access_token="+token, cs, nil)
}

// 修改客服
func UpdateCS(token string, cs *PostCS) error {
	return wechat.PostSchema(wechat.KindJson, "https://api.weixin.qq.com/customservice/kfaccount/update?access_token="+token, cs, nil)
}

// 删除客服
func DeleteCS(token string, cs *PostCS) error {
	return wechat.PostSchema(wechat.KindJson, "https://api.weixin.qq.com/customservice/kfaccount/del?access_token="+token, cs, nil)
}

// 上传头像
func UploadAvatar(r http.Request, token string, kfAccount string) error {
	url := "http://api.weixin.qq.com/customservice/kfaccount/uploadheadimg"
	url += "?access_token=" + token + "&kf_account=" + kfAccount
	req, err := http.NewRequest("POST", url, r.Body)
	if err != nil {
		return err
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	resErr := &wechat.Error{}
	err = json.NewDecoder(res.Body).Decode(resErr)
	if err != nil {
		return err
	} else {
		if resErr.ErrCode != 0 {
			return resErr
		}
	}
	return nil
}

type customerServiceResponse struct {
	wechat.Error
	KfList []*CustomerService `json:"kf_list"`
}

// 获得所有客服列表
func GetAllCS(token string) (css []*CustomerService, err error) {
	url := "https://api.weixin.qq.com/cgi-bin/customservice/getkflist?access_token=" + token
	res, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	data, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	csResponse := &customerServiceResponse{}
	err = json.Unmarshal(data, &csResponse)
	if err != nil {
		return nil, err
	} else {
		if csResponse.ErrCode != 0 {
			return nil, &csResponse.Error
		} else {
			return csResponse.KfList, nil
		}
	}
}
