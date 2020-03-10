// Copyright 2018 orivil.com. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found at https://mit-license.org.

package oauth2

import (
	"errors"
	"github.com/orivil/wechat"
)

var ErrGetUsersMoreThan100 = errors.New("批量获取用户信息最多一次只能拉取100个")

type Response struct {
	User
	wechat.Error
}

type UserSex int

const (
	Unknown UserSex = iota
	Male
	Female
)

// see: https://mp.weixin.qq.com/wiki?t=resource/res_main&id=mp1421140839
type User struct {
	Subscribe     int     `gorm:"index" json:"subscribe" desc:"是否关注, 1-关注 0-未关注"`
	SubscribeTime int64   `gorm:"index;not null" json:"subscribe_time" desc:"关注时间"`
	Openid        string  `gorm:"unique_index" json:"openid"`
	Nickname      string  `json:"nickname"`
	Sex           UserSex `gorm:"index" json:"sex" desc:"用户性别: 0-未知 1-男性 2-女性"`
	Province      string  `json:"province"`
	City          string  `json:"city"`
	Country       string  `json:"country"`
	HeadImgUrl    string  `json:"headimgurl"`
	Language      string  `json:"language"`
	UnionID       string  `gorm:"index" json:"unionid"`

	// 粉丝备注
	Remark string `json:"remark"`

	// 用户所在分组ID
	GroupID int `json:"groupid"`

	// 用户被打上的标签ID列表
	TagIDList []int `json:"tagid_list" gorm:"-"`

	// 用户关注的渠道来源，ADD_SCENE_SEARCH 公众号搜索，ADD_SCENE_ACCOUNT_MIGRATION 公众号迁移，
	// ADD_SCENE_PROFILE_CARD 名片分享，ADD_SCENE_QR_CODE 扫描二维码，
	// ADD_SCENEPROFILE LINK 图文页内名称点击，ADD_SCENE_PROFILE_ITEM 图文页右上角菜单，
	// ADD_SCENE_PAID 支付后关注，ADD_SCENE_OTHERS 其他
	SubscribeScene string `json:"subscribe_scene" desc:"用户关注的渠道来源"`

	// 二维码扫码场景（开发者自定义）
	QRScene int `json:"qr_scene"`

	// 二维码扫码场景描述（开发者自定义）
	QRSceneStr string `json:"qr_scene_str"`
}

// accessToken 必须是在 scope 为 "snsapi_userinfo" 下获得的(即用户点击确认授权后)才有效
func GetUserInfo(openid, accessToken string) (user *User, err error) {
	uri := "https://api.weixin.qq.com/sns/userinfo?lang=zh_CN&access_token=" + accessToken + "&openid=" + openid
	user = &User{}
	err = wechat.GetJson(uri, user)
	if err != nil {
		return nil, err
	} else {
		return user, nil
	}
}

// accessToken 为用户所关注公众号的 access token, 非用户授权获得的 token, 只能在用户关注公众号之后才能使用
func GetSubscribersInfo(openids []string, accessToken string) (users []*User, err error) {
	if ln := len(openids); ln > 100 {
		return nil, ErrGetUsersMoreThan100
	} else if ln > 0 {
		uri := "https://api.weixin.qq.com/cgi-bin/user/info/batchget?access_token=" + accessToken
		list := make([]*openid, len(openids))
		for key, id := range openids {
			list[key] = &openid{Openid: id}
		}
		res := &userInfoList{}
		err = wechat.PostSchema(wechat.KindJson, uri, &getUsers{UserList: list}, res)
		if err != nil {
			return nil, err
		} else {
			return res.UserInfoList, nil
		}
	} else {
		return nil, nil
	}
}

type getUsers struct {
	UserList []*openid `json:"user_list"`
}

type openid struct {
	Openid string `json:"openid"`
}

type userInfoList struct {
	UserInfoList []*User `json:"user_info_list"`
}
