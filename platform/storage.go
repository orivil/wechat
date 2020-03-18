// Copyright 2019 orivil.com. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found at https://mit-license.org.

package platform

type Storage struct {
	// 第三方验证码存储器
	VerifyTicket DataStorage

	PureAuthCode ExpireDataStorage

	// 第三方授权令牌存储器
	AppAccessToken ExpireDataStorage

	// 第三方授权 ticket 存储器
	AppTicket ExpireDataStorage

	// 用户授权令牌存储器
	UserAccessToken ExpireDataStorage

	// 第三方操作令牌存储器
	ComponentAccessToken ExpireDataStorage
}

func NewStorage(dt DataStorage, edt ExpireDataStorage) *Storage {
	return &Storage{
		VerifyTicket:dt,
		PureAuthCode:edt,
		AppAccessToken:edt,
		AppTicket:edt,
		UserAccessToken:edt,
		ComponentAccessToken:edt,
	}
}

// 永久数据读写接口
type DataStorage interface {
	Store(key, value string) error
	Read(key string) (value string, err error)
	// 删除数据
	Del(key string) (err error)
}

// 过期数据读写接口
type ExpireDataStorage interface {

	// 存储会过期的数据
	Store(key string, data *ExpireData) error

	// 读取所有会过期的数据
	Read(key string) (data *ExpireData, err error)

	// 删除数据
	Del(key string) (err error)
}
