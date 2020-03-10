// Copyright 2019 orivil.com. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found at https://mit-license.org.

package platform

import (
	"github.com/pkg/errors"
	"sync"
	"time"
)

// 数据缓存容易
type DataContainer struct {
	storage DataStorage
	mu      sync.Mutex
}

func NewDataContainer(storage DataStorage) *DataContainer {
	return &DataContainer{
		storage: storage,
	}
}

func (dc *DataContainer) set(key, value string) error {
	dc.mu.Lock()
	defer dc.mu.Unlock()
	return dc.storage.Store(key, value)
}

func SetData(container *DataContainer, key, value string) error {
	return container.set(key, value)
}

func (dc *DataContainer) Get(key string) (value string, err error) {
	dc.mu.Lock()
	defer dc.mu.Unlock()
	return dc.storage.Read(key)
}

// 数据刷新器, refreshToken = (newRefreshToken != "" ? newRefreshToken : value)
type ExpireDataRefresher func(key, refreshToken string) (value, newRefreshToken string, expiresIn int64, err error)

// 过期数据模型
type ExpireData struct {
	// 数据
	Value string

	// 刷新数据时所需要的令牌
	RefreshToken string

	// 过期时间
	ExpireAt time.Time
}

func GetRefreshToken(ed *ExpireData) string {
	if ed == nil {
		return ""
	} else {
		if ed.RefreshToken != "" {
			return ed.RefreshToken
		} else {
			return ed.Value
		}
	}
}

// 有时间限制的数据容器
type ExpireDataContainer struct {
	name      string
	refresher ExpireDataRefresher
	storage   ExpireDataStorage
	//data      map[string]*ExpireData
	refreshBeforeExpire time.Duration
	mu                  sync.Mutex
}

// 设置数据, 设置之后会将数据保存到 storage 中
func (e *ExpireDataContainer) set(key, value, refreshToken string, expiresIn int64) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	data := &ExpireData{
		Value:        value,
		RefreshToken: refreshToken,
		ExpireAt:     time.Now().Add(time.Duration(expiresIn)*time.Second - e.refreshBeforeExpire),
	}
	return e.storage.Store(key, data)
}

func SetExpireData(container *ExpireDataContainer, key, value, refreshToken string, expiresIn int64) error {
	return container.set(key, value, refreshToken, expiresIn)
}

var ticker = NewTimeProvider(1 * time.Minute)

func (e *ExpireDataContainer) Del(key string) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.storage.Del(key)
}

// 强制刷新获得新的数据
func (e *ExpireDataContainer) Refresh(key string) (value string, err error) {
	return e.get(key, true)
}

// Get 用于获得数据, 如果缓存中没有数据则从 storage 中获取数据, 如果未设置数据或者数据过期则调用函数刷新获得新的数据
func (e *ExpireDataContainer) Get(key string) (value string, err error) {
	return e.get(key, false)
}

func (e *ExpireDataContainer) get(key string, refresh bool) (value string, err error) {
	e.mu.Lock()
	defer e.mu.Unlock()
	var data *ExpireData
	data, err = e.storage.Read(key)
	if err != nil {
		return "", err
	}
	now := ticker.Now()
	if refresh || data == nil || now.After(data.ExpireAt) {
		value, refreshToken, expiresIn, err := e.refresher(key, GetRefreshToken(data))
		if err != nil {
			return "", errors.Wrapf(err, "expire data container [%s]", e.name)
		}
		data = &ExpireData{
			Value:        value,
			RefreshToken: refreshToken,
			ExpireAt:     now.Add(time.Duration(expiresIn)*time.Second - e.refreshBeforeExpire),
		}
		if e.storage != nil {
			err = e.storage.Store(key, data)
			if err != nil {
				return "", errors.WithMessage(err, "store expire data failed")
			}
		}
	}
	return data.Value, nil
}

// 新建过期数据, storage 用于保存数据以及读取数据
func NewExpireDataContainer(name string, storage ExpireDataStorage, refreshBeforeExpire time.Duration, refresher ExpireDataRefresher) *ExpireDataContainer {
	return &ExpireDataContainer{
		name:                name,
		refresher:           refresher,
		storage:             storage,
		refreshBeforeExpire: refreshBeforeExpire,
	}
}
