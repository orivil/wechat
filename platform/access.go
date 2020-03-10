// Copyright 2019 orivil.com. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found at https://mit-license.org.

package platform

import (
	"errors"
	"fmt"
	"github.com/orivil/wechat"
	"github.com/orivil/wechat/access"
	"github.com/orivil/wechat/oauth2"
	"github.com/orivil/wechat/open-platform"
	errors2 "github.com/pkg/errors"
	"sync"
	"time"
)

var ErrUserNotAuthorized = errors.New("未获得用户授权")
var ErrComponentAccessIsNotSet = errors.New("未设置第三方平台")
var ErrNeedAuthorizerApp = errors.New("必须为授权平台")

type ErrAppNotAuthorized struct {
	ComponentAppid  string
	AuthorizerAppid string
}

func (na *ErrAppNotAuthorized) Error() string {
	return fmt.Sprintf("第三方平台 %s(appid) 未获得公众号: %s(appid) 授权", na.ComponentAppid, na.AuthorizerAppid)
}

type ComponentAccess struct {
	Appid        string
	VerifyTicket *DataContainer
	PreAuthCode  *ExpireDataContainer
	AccessToken  *ExpireDataContainer
}

// 获取授权方的帐号的详细信息.
// 可通过 ComponentAuthNotify 监听授权方最新的授权权限动态. 由于授权方所授权的权限并不一定就是第三
// 方平台所设置的权限, 授权方可选择部分权限, 因此有必要根据授权事件更新授权方信息.
func (ca *ComponentAccess) GetAuthorizer(appid string) (info *open_platform.Authorizer, err error) {
	accessToken, err := ca.AccessToken.Get(ca.Appid)
	if err != nil {
		return nil, err
	}
	return open_platform.GetAuthorizerInfo(ca.Appid, appid, accessToken)
}

func (ca *ComponentAccess) GetAccessToken() (token string, err error) {
	return ca.AccessToken.Get(ca.Appid)
}

type AppAccess struct {
	Appid           string
	ComponentAccess *ComponentAccess

	// 公众号 access token(操作令牌) 提供器
	AppAccessToken *ExpireDataContainer
	AppTicket      *ExpireDataContainer

	// UserAccessToken 存储的是 oauth2.AccessToken, 主要用于换取用户信息.
	// 如果 access token 过期, 则会自动调用 refresh token 进行刷新, 如果 refresh token 长时间不用, 30天后将会过期, 需要用户从新授权.
	//
	// 如果是关注用户, 也可以调用 GetSubscribedUser() 方法获得用户的信息.
	UserAccessToken *ExpireDataContainer
}

// 获取授权方的帐号的详细信息.
// 可通过 ComponentAuthNotify 监听授权方最新的授权权限动态. 由于授权方所授权的权限并不一定就是第三
// 方平台所设置的权限, 授权方可选择部分权限, 因此有必要根据授权事件更新授权方信息.
func (a *AppAccess) GetAuthorizer() (info *open_platform.Authorizer, err error) {
	if a.ComponentAccess != nil {
		return a.ComponentAccess.GetAuthorizer(a.Appid)
	} else {
		return nil, ErrComponentAccessIsNotSet
	}
}

// 获得用户信息，需要用户授权(scope 必须是 snsapi_userinfo).
// 使用之前需要先保存用户授权令牌
func (a *AppAccess) GetUser(openid string) (info *oauth2.User, err error) {
	token, err := a.UserAccessToken.Get(openid)
	if err != nil {
		return nil, err
	}
	if token == "" {
		return nil, ErrUserNotAuthorized
	}
	return oauth2.GetUserInfo(openid, token)
}

// 获得用户信息，需要用户关注.
// 可通过监听用户关注事件, 然后再调用该方法获得用户信息.
func (a *AppAccess) GetSubscribedUsers(openids []string) (users []*oauth2.User, err error) {
	err = splitStrs(openids, 100, func(subs []string) error {
		token, err := a.AppAccessToken.Get(a.Appid)
		if err != nil {
			return err
		}
		us, err := oauth2.GetSubscribersInfo(subs, token)
		if err != nil {
			return err
		} else {
			users = append(users, us...)
			return nil
		}
	})
	if err != nil {
		return nil, err
	} else {
		return users, nil
	}
}

// 生成公众号菜单.
// 开放平台可通过监听授权方最新的授权权限动态, 为相应的公众号生成菜单
func (a *AppAccess) GenerateMenus(menus *wechat.Menus) (err error) {
	token, err := a.AppAccessToken.Get(a.Appid)
	if err != nil {
		return err
	}
	return wechat.GenerateMenus(token, menus)
}

// 获得 js 接口签名. 一个 refererUrl 只需要一次签名
func (a *AppAccess) GetJsApiSignature(nonce, refererUrl string, timestamp int64) (signature string, err error) {
	ticket, err := a.AppTicket.Get(a.Appid)
	if err != nil {
		return "", err
	}
	return wechat.GetJsApiSignature(ticket, nonce, refererUrl, timestamp)
}

func NewComponentAccess(storage *Storage, appid string, secretProvider AppSecretProvider) (access *ComponentAccess) {
	// 在第三方平台创建审核通过后，微信服务器会向其“授权事件接收URL”每隔10分钟定时推送component_verify_ticket
	verifyTicket := NewDataContainer(storage.VerifyTicket)

	// 第三方平台component_access_token是第三方平台的下文中接口的调用凭据，也叫做令牌（component_access_token）。
	// 每个令牌是存在有效期（2小时）的，且令牌的调用不是无限制的，请第三方平台做好令牌的管理，在令牌快过期时（比如1小
	// 时50分）再进行刷新。
	accessToken := NewExpireDataContainer("componentAccessToken", storage.ComponentAccessToken, 20*time.Minute, func(componentAppid, refreshToken string) (value, newRefreshToken string, expiresIn int64, err error) {
		ticket, err := verifyTicket.Get(componentAppid)
		if err != nil {
			return "", "", 0, err
		}
		secret, err := secretProvider(appid)
		if err != nil {
			return "", "", 0, err
		}
		token, err := open_platform.GetComponentAccessToken(componentAppid, secret, ticket)
		if err != nil {
			return "", "", 0, err
		} else {
			return token.Token, "", token.ExpiresIn, nil
		}
	})

	// 预授权码。预授权码用于公众号或小程序授权时的第三方平台方安全验证。
	preAuthCode := NewExpireDataContainer("preAuthCode", storage.PureAuthCode, 20*time.Minute, func(componentAppid, refreshToken string) (value, newRefreshToken string, expiresIn int64, err error) {
		token, err := accessToken.Get(componentAppid)
		if err != nil {
			return "", "", 0, err
		} else {
			code, err := open_platform.GetPureAuthCode(componentAppid, token)
			if err != nil {
				return "", "", 0, err
			} else {
				return code.PreAuthCode, "", code.ExpiresIn, nil
			}
		}
	})
	return &ComponentAccess{
		Appid:        appid,
		VerifyTicket: verifyTicket,
		AccessToken:  accessToken,
		PreAuthCode:  preAuthCode,
	}
}

func NewAppAccess(storage *Storage, appid string, secretProvider AppSecretProvider, componentAccess *ComponentAccess) *AppAccess {

	isAuthorizedPlatform := componentAccess != nil
	// 公众号授权方操作令牌
	var appAccessToken *ExpireDataContainer
	var componentAppid string
	if isAuthorizedPlatform {
		componentAppid = componentAccess.Appid
		// 开放平台
		appAccessToken = NewExpireDataContainer("openPlatformAppAccessToken", storage.AppAccessToken, 20*time.Minute, func(appid, refreshToken string) (value, newRefreshToken string, expiresIn int64, err error) {
			cAccessToken, err := componentAccess.AccessToken.Get(componentAppid)
			if err != nil {
				return "", "", 0, err
			}
			if refreshToken == "" {
				at, err := open_platform.GetAuthorizerInfo(componentAppid, appid, cAccessToken)
				if err != nil {
					return "", "", 0, errors2.Wrap(err, componentAppid+":"+appid+" get refresh token")
				} else {
					refreshToken = at.AuthorizationInfo.AuthorizerRefreshToken
				}
			}
			at, err := open_platform.RefreshAuthorizerToken(componentAppid, appid, refreshToken, cAccessToken)
			if err != nil {
				return "", "", 0, err
			} else {
				return at.AuthorizerAccessToken, at.AuthorizerRefreshToken, at.ExpiresIn, nil
			}
		})
	} else {
		// 公众平台
		appAccessToken = NewExpireDataContainer("publicPlatformAppAccessToken", storage.AppAccessToken, 20*time.Minute, func(key, refreshToken string) (value, newRefreshToken string, expiresIn int64, err error) {
			secret, err := secretProvider(appid)
			if err != nil {
				return "", "", 0, err
			}
			token, err := access.GetAccessToken(appid, secret)
			if err != nil {
				return "", "", 0, err
			} else {
				return token.Value, "", token.ExpiresIn, nil
			}
		})
	}

	// app ticket
	appTicket := NewExpireDataContainer("appTicket", storage.AppTicket, 20*time.Minute, func(appid, refreshToken string) (value, newRefreshToken string, expiresIn int64, err error) {
		token, err := appAccessToken.Get(appid)
		if err != nil {
			return "", "", 0, err
		} else {
			ticket, err := access.GetTicket(token)
			if err != nil {
				return "", "", 0, err
			} else {
				return ticket.Value, "", ticket.ExpiresIn, nil
			}
		}
	})

	// 用户授权操作令牌
	var userAccessToken *ExpireDataContainer
	if isAuthorizedPlatform {
		// 开放平台
		userAccessToken = NewExpireDataContainer("openPlatformUserAccessToken", storage.UserAccessToken, 20*time.Minute, func(key, refreshToken string) (value, newRefreshToken string, expiresIn int64, err error) {
			if refreshToken == "" {
				return "", "", 0, ErrUserNotAuthorized
			}
			cAccessToken, err := componentAccess.AccessToken.Get(componentAppid)
			if err != nil {
				return "", "", 0, err
			}
			token, err := oauth2.RefreshComponentAccessToken(appid, refreshToken, componentAppid, cAccessToken)
			if err != nil {
				return "", "", 0, err
			} else {
				return token.AccessToken, token.RefreshToken, token.ExpiresIn, nil
			}
		})
	} else {
		// 公众平台
		userAccessToken = NewExpireDataContainer("publicPlatformUserAccessToken", storage.UserAccessToken, 20*time.Minute, func(key, refreshToken string) (value, newRefreshToken string, expiresIn int64, err error) {
			if refreshToken == "" {
				return "", "", 0, ErrUserNotAuthorized
			}
			token, err := oauth2.RefreshAccessToken(appid, refreshToken)
			if err != nil {
				return "", "", 0, err
			} else {
				return token.AccessToken, token.RefreshToken, token.ExpiresIn, nil
			}
		})
	}
	return &AppAccess{
		Appid:           appid,
		ComponentAccess: componentAccess,
		AppAccessToken:  appAccessToken,
		AppTicket:       appTicket,
		UserAccessToken: userAccessToken,
	}
}

// 提供 app 配置
type AppSecretProvider func(appid string) (secret string, err error)
type AppAesKeyProvider func(appid string) (token, aesKey string, err error)
type ComponentAppidProvider func(appid string) (componentAppid string, err error)

type AccessContainer struct {
	storage                *Storage
	AppSecretProvider      AppSecretProvider
	AppAesKeyProvider      AppAesKeyProvider
	ComponentAppidProvider ComponentAppidProvider
	appAccess              map[string]*AppAccess
	componentAccess        map[string]*ComponentAccess
	decrypter              map[string]*wechat.WXBizMsgCrypt
	mu                     sync.RWMutex
}

func (ac *AccessContainer) GetComponentAccess(appid string) (a *ComponentAccess, err error) {
	ac.mu.RLock()
	a, ok := ac.componentAccess[appid]
	ac.mu.RUnlock()
	if !ok {
		ac.mu.Lock()
		defer ac.mu.Unlock()
		a = NewComponentAccess(ac.storage, appid, ac.AppSecretProvider)
		ac.componentAccess[appid] = a
	}
	return a, nil
}

// 如果未提供 componentAppid, 则会尝试从存储器中获得该值, 如果最终获得 componentAppid, 则当作开放平台的授权应用处理.
// 如果没有获得该值, 则从存储器中获取 secret, 当作公众平台处理
func (ac *AccessContainer) GetAppAccess(componentAppid, appid string) (a *AppAccess, err error) {
	ac.mu.RLock()
	a, ok := ac.appAccess[appid]
	ac.mu.RUnlock()
	if !ok {
		var componentAccess *ComponentAccess
		if componentAppid == "" {
			componentAppid, _ = ac.ComponentAppidProvider(appid)
		}
		if componentAppid != "" {
			componentAccess, err = ac.GetComponentAccess(componentAppid)
			if err != nil {
				return nil, err
			}
		}
		ac.mu.Lock()
		defer ac.mu.Unlock()
		a = NewAppAccess(ac.storage, appid, ac.AppSecretProvider, componentAccess)
		ac.appAccess[appid] = a
	}
	return a, nil
}

// 获得三方平台 appid, 如果提供的 appid 不是授权 APP 的 appid, 则会返回 ErrNeedAuthorizerApp 错误
func (ac *AccessContainer) GetComponentAppid(appid string) (componentAppid string, err error) {
	appAccess, err := ac.GetAppAccess("", appid)
	if err != nil {
		return "", err
	}
	if appAccess.ComponentAccess != nil {
		return appAccess.ComponentAccess.Appid, nil
	}
	return "", ErrNeedAuthorizerApp
}

func (ac *AccessContainer) GetAppAccessToken(appid string) (token string, err error) {
	appAccess, err := ac.GetAppAccess("", appid)
	if err != nil {
		return "", err
	} else {
		return appAccess.AppAccessToken.Get(appid)
	}
}

func (ac *AccessContainer) GetAppTicket(appid string) (ticket string, err error) {
	appAccess, err := ac.GetAppAccess("", appid)
	if err != nil {
		return "", err
	} else {
		return appAccess.AppTicket.Get(appid)
	}
}

// decrypter 有可能为空, 表示消息传输格式为明文传输.
// appid 类型可以为第三方平台, 授权方平台, 或公众平台
func (ac *AccessContainer) GetDecrypter(appid string) (decrypter *wechat.WXBizMsgCrypt, err error) {
	ac.mu.RLock()
	decrypter, ok := ac.decrypter[appid]
	ac.mu.RUnlock()
	if !ok {
		// 当作授权方平台处理
		var aesAppid string
		componendAppid, err := ac.GetComponentAppid(appid)
		if err != nil {
			// 当前 appid 为三方平台或公众平台
			if err == ErrNeedAuthorizerApp {
				aesAppid = appid
			} else {
				return nil, err
			}
		} else {
			// 当前 appid 为授权平台
			aesAppid = componendAppid
		}
		token, aesKey, err := ac.AppAesKeyProvider(aesAppid)
		if err != nil {
			return nil, err
		}
		if aesKey != "" {
			decrypter, err = wechat.NewWXBizMsgCrypt(token, aesKey, aesAppid)
			if err != nil {
				return nil, err
			} else {
				ac.mu.Lock()
				defer ac.mu.Unlock()
				ac.decrypter[appid] = decrypter
			}
		}
	}
	return decrypter, nil
}

// 清除缓存
func (ac *AccessContainer) Flash(appid string) {
	ac.mu.Lock()
	defer ac.mu.Unlock()
	delete(ac.decrypter, appid)
	delete(ac.appAccess, appid)
	delete(ac.componentAccess, appid)
}

func NewAccessContainer(storage *Storage, appSecret AppSecretProvider, aesKey AppAesKeyProvider, components ComponentAppidProvider) *AccessContainer {
	return &AccessContainer{
		appAccess:              make(map[string]*AppAccess, 5),
		componentAccess:        make(map[string]*ComponentAccess, 5),
		decrypter:              make(map[string]*wechat.WXBizMsgCrypt, 5),
		AppSecretProvider:      appSecret,
		AppAesKeyProvider:      aesKey,
		ComponentAppidProvider: components,
		storage:                storage,
	}
}

func splitStrs(strs []string, limit int, walk func(subs []string) error) error {
	total := len(strs)
	for offset := 0; offset < total; offset += limit {
		end := offset + limit
		if end > total {
			end = total
		}
		subs := strs[offset:end]
		err := walk(subs)
		if err != nil {
			return err
		}
	}
	return nil
}
