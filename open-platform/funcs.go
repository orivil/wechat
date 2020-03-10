// Copyright 2018 orivil.com. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found at https://mit-license.org.

package open_platform

import (
	"encoding/xml"
	"github.com/orivil/wechat"
	"net/http"
)

const (
	EvtComponentVerifyTicket ComponentAuthorizationEvent = "component_verify_ticket"
	EvtUnauthorized          ComponentAuthorizationEvent = "unauthorized"
	EvtUpdateauthorized      ComponentAuthorizationEvent = "updateauthorized"
	EvtAuthorized            ComponentAuthorizationEvent = "authorized"
)

type ComponentAuthorizationEvent string

// 在第三方平台创建审核通过后，微信服务器会向其“授权事件接收URL”每隔10分钟定时推送component_verify_ticket。
// 第三方平台方在收到ticket推送后也需进行解密, 接收到后必须直接返回字符串success。
//
// 当公众号对第三方平台进行授权、取消授权、更新授权后，微信服务器会向第三方平台方的授权事件
// 接收URL（创建第三方平台时填写）推送相关通知。
type AuthorizationNotify struct {
	XMLName               xml.Name                    `xml:"xml"`
	AppId                 string                      // 第三方平台appid
	CreateTime            int64                       // 时间戳
	InfoType              ComponentAuthorizationEvent // component_verify_ticket, unauthorized 是取消授权，updateauthorized 是更新授权，authorized 是授权成功通知
	ComponentVerifyTicket string

	AuthorizerAppid              string // 公众号或小程序
	AuthorizationCode            string // 授权码，可用于换取公众号的接口调用凭据
	AuthorizationCodeExpiredTime int64  // 授权码过期时间
	PreAuthCode                  string // 预授权码
}

// 监听第三方授权通知
func ListenComponentAuthorizationNotify(r *http.Request, decrypter *wechat.WXBizMsgCrypt) (notify *AuthorizationNotify, err error) {
	data, err := decrypter.DecryptRequest(r)
	if err != nil {
		return nil, err
	}
	notify = &AuthorizationNotify{}
	err = xml.Unmarshal(data, notify)
	if err != nil {
		return nil, err
	} else {
		return notify, nil
	}
}

type ComponentAccessToken struct {
	Token     string `json:"component_access_token"`
	ExpiresIn int64  `json:"expires_in"`
}

// 第三方平台component_access_token是第三方平台的下文中接口的调用凭据，也叫做令牌（component_access_token）。
// 每个令牌是存在有效期（2小时）的，且令牌的调用不是无限制的，请第三方平台做好令牌的管理，在令牌快过期时（比如1小时50分）再进行刷新。
func GetComponentAccessToken(componentAppid, appSecret, verifyTicket string) (token *ComponentAccessToken, err error) {
	data := map[string]string{
		"component_appid":         componentAppid,
		"component_appsecret":     appSecret,
		"component_verify_ticket": verifyTicket,
	}
	token = &ComponentAccessToken{}
	err = wechat.PostSchema(wechat.KindJson, "https://api.weixin.qq.com/cgi-bin/component/api_component_token", data, token)
	if err != nil {
		return nil, err
	} else {
		return token, nil
	}
}

type PreAuthCode struct {
	PreAuthCode string `json:"pre_auth_code"`
	ExpiresIn   int64  `json:"expires_in"`
}

// 获取预授权码pre_auth_code
// 该API用于获取预授权码。预授权码用于公众号或小程序授权时的第三方平台方安全验证。
func GetPureAuthCode(componentAppid, componentAccessToken string) (code *PreAuthCode, err error) {
	data := map[string]string{"component_appid": componentAppid}
	code = &PreAuthCode{}
	err = wechat.PostSchema(wechat.KindJson, "https://api.weixin.qq.com/cgi-bin/component/api_create_preauthcode?component_access_token="+componentAccessToken, data, code)
	if err != nil {
		return nil, err
	} else {
		return code, nil
	}
}

type FuncScope struct {
	FuncscopeCategory Info `json:"funcscope_category"`
}

type Info struct {
	ID int `json:"id"`
}

// 使用授权码换取公众号或小程序的接口调用凭据和授权信息
// 该API用于使用授权码换取授权公众号或小程序的授权信息，并换取authorizer_access_token和authorizer_refresh_token。
// 授权码的获取，需要在用户在第三方平台授权页中完成授权流程后，在回调URI中通过URL参数提供给第三方平台方。请注意，由于
// 现在公众号或小程序可以自定义选择部分权限授权给第三方平台，因此第三方平台开发者需要通过该接口来获取公众号或小程序具
// 体授权了哪些权限，而不是简单地认为自己声明的权限就是公众号或小程序授权的权限。
//
// 参数	                         说明
//
// authorizer_appid	             授权方appid
//
// authorizer_access_token	     授权方接口调用凭据（在授权的公众号或小程序具备API权限时，才有此返回值），也简称为令牌
//
// expires_in	                 有效期（在授权的公众号或小程序具备API权限时，才有此返回值）
//
// authorizer_refresh_token	     接口调用凭据刷新令牌（在授权的公众号具备API权限时，才有此返回值），刷新令牌主要用于第
//                               三方平台获取和刷新已授权用户的access_token，只会在授权时刻提供，请妥善保存。 一旦丢失，
//                               只能让用户重新授权，才能再次拿到新的刷新令牌
//
// func_info	                 授权给开发者的权限集列表，ID为1到26分别代表： 1、消息管理权限 2、用户管理权限 3、帐号服务权限
//                               4、网页服务权限 5、微信小店权限 6、微信多客服权限 7、群发与通知权限 8、微信卡券权限 9、微信扫
//                               一扫权限 10、微信连WIFI权限 11、素材管理权限 12、微信摇周边权限 13、微信门店权限 15、自定义菜
//                               单权限 16、获取认证状态及信息 17、帐号管理权限（小程序） 18、开发管理与数据分析权限（小程序）
//                               19、客服消息管理权限（小程序） 20、微信登录权限（小程序） 21、数据分析权限（小程序） 22、城市
//                               服务接口权限 23、广告管理权限 24、开放平台帐号管理权限 25、 开放平台帐号管理权限（小程序）
//                               26、微信电子发票权限 41、搜索widget的权限 请注意： 1）该字段的返回不会考虑公众号是否具备该
//                               权限集的权限（因为可能部分具备），请根据公众号的帐号类型和认证情况，来判断公众号的接口权限。
type AuthorizationInfo struct {
	AuthorizerAppid string       `json:"authorizer_appid"`
	FuncInfo        []*FuncScope `json:"func_info"`
	*AuthorizerToken
}

type info struct {
	AuthorizationInfo *AuthorizationInfo `json:"authorization_info"`
}

// 根据 authorizationCode 换取授权信息.
// authorizationCode 是在授权成功时返回的数据，详见第三方平台授权流程说明
func GetAuthorizationInfo(componentAppid, authorizationCode, componentAccessToken string) (ai *AuthorizationInfo, err error) {
	data := map[string]string{
		"component_appid":    componentAppid,
		"authorization_code": authorizationCode,
	}
	res := &info{}
	err = wechat.PostSchema(wechat.KindJson, "https://api.weixin.qq.com/cgi-bin/component/api_query_auth?component_access_token="+componentAccessToken, data, res)
	if err != nil {
		return nil, err
	} else {
		// 过期时间提前 20 分钟
		ai = res.AuthorizationInfo
		return ai, nil
	}
}

type AuthorizerToken struct {
	AuthorizerAccessToken  string `json:"authorizer_access_token"`
	ExpiresIn              int64  `json:"expires_in"`
	AuthorizerRefreshToken string `json:"authorizer_refresh_token"`
}

// 获取（刷新）授权公众号或小程序的接口调用凭据（令牌）
// 该API用于在授权方令牌（authorizer_access_token）失效时，可用刷新令牌（authorizer_refresh_token）获取新的令
// 牌。 请注意，此处token是2小时刷新一次，开发者需要自行进行token的缓存，避免token的获取次数达到每日的限定额度。
func RefreshAuthorizerToken(componentAppid, authorizerAppid, authorizerRefreshToken, componentAccessToken string) (at *AuthorizerToken, err error) {
	data := map[string]string{
		"component_appid":          componentAppid,
		"authorizer_appid":         authorizerAppid,
		"authorizer_refresh_token": authorizerRefreshToken,
	}
	at = &AuthorizerToken{}
	err = wechat.PostSchema(wechat.KindJson, "https://api.weixin.qq.com/cgi-bin/component/api_authorizer_token?component_access_token="+componentAccessToken, data, at)
	if err != nil {
		return nil, err
	} else {
		return at, nil
	}
}

type Business struct {
	OpenStore int `json:"open_store"`
	OpenScan  int `json:"open_scan"`
	OpenPay   int `json:"open_pay"`
	OpenCard  int `json:"open_card"`
	OpenShake int `json:"open_shake"`
}

type AuthorizerInfo struct {

	// 授权方昵称
	NickName string `json:"nick_name"`

	// 授权方头像
	HeadImg string `json:"head_img"`

	// 授权方公众号类型，0代表订阅号，1代表由历史老帐号升级后的订阅号，2代表服务号
	ServiceTypeInfo *Info `json:"service_type_info"`

	// 授权方认证类型，-1代表未认证，0代表微信认证，1代表新浪微博认证，2代表腾讯微博认证，
	// 3代表已资质认证通过但还未通过名称认证，4代表已资质认证通过、还未通过名称认证，但通
	// 过了新浪微博认证，5代表已资质认证通过、还未通过名称认证，但通过了腾讯微博认证
	VerifyTypeInfo *Info `json:"verify_type_info"`

	// 授权方公众号的原始ID
	UserName string `json:"user_name"`

	// 公众号的主体名称
	PrincipalName string `json:"principal_name"`

	// 授权方公众号所设置的微信号，可能为空
	Alias string `json:"alias"`

	// 用以了解以下功能的开通状况（0代表未开通，1代表已开通）： open_store:是否开通微信门店
	// 功能 open_scan:是否开通微信扫商品功能 open_pay:是否开通微信支付功能 open_card:是否
	// 开通微信卡券功能 open_shake:是否开通微信摇一摇功能
	BusinessInfo *Business `json:"business_info"`

	// 二维码图片的URL，开发者最好自行也进行保存
	QrcodeUrl string `json:"qrcode_url"`

	Idc int `json:"idc"` // 文档无说明

	// APP 简介
	Signature string `json:"signature"`
}

type Authorizer struct {
	AuthorizerInfo    AuthorizerInfo    `json:"authorizer_info"`
	AuthorizationInfo AuthorizationInfo `json:"authorization_info"`
}

// 获取授权方的帐号基本信息
// 该API用于获取授权方的基本信息，包括头像、昵称、帐号类型、认证类型、微信号、原始ID和二维码图片URL。
//
// 需要特别记录授权方的帐号类型，在消息及事件推送时，对于不具备客服接口的公众号，需要在5秒内立即响应；
// 而若有客服接口，则可以选择暂时不响应，而选择后续通过客服接口来发送消息触达粉丝。
func GetAuthorizerInfo(componentAppid, authorizerAppid, componentAccessToken string) (a *Authorizer, err error) {
	data := map[string]string{
		"component_appid":  componentAppid,
		"authorizer_appid": authorizerAppid,
	}
	res := &Authorizer{}
	err = wechat.PostSchema(wechat.KindJson, "https://api.weixin.qq.com/cgi-bin/component/api_get_authorizer_info?component_access_token="+componentAccessToken, data, res)
	if err != nil {
		return nil, err
	} else {
		return res, nil
	}
}

type Option struct {
	AuthorizerAppid string `json:"authorizer_appid"`
	OptionName      string `json:"option_name"`
	OptionValue     string `json:"option_value"`
}

// 设置授权方的选项信息
// 该API用于设置授权方的公众号或小程序的选项信息，如：地理位置上报，语音识别开关，多客服开关。
// 注意，设置各项选项设置信息，需要有授权方的授权，详见权限集说明。
//
// option_name	                         option_value	  选项值说明
//
// location_report(地理位置上报选项)	     0	              无上报
//	                                     1	              进入会话时上报
//	                                     2	              每5s上报
//
// voice_recognize（语音识别开关选项）	     0	              关闭语音识别
//	                                     1	              开启语音识别
//
// customer_service（多客服开关选项）	     0	              关闭多客服
//	                                     1	              开启多客服
func SetAuthorizerOption(componentAppid, accessToken string, option Option) error {
	data := map[string]string{
		"component_appid":  componentAppid,
		"authorizer_appid": option.AuthorizerAppid,
		"option_name":      option.OptionName,
		"option_value":     option.OptionValue,
	}
	return wechat.PostSchema(wechat.KindJson, "https://api.weixin.qq.com/cgi-bin/component/ api_set_authorizer_option?component_access_token="+accessToken, data, nil)
}

// 获取授权方的选项设置信息
// 该API用于获取授权方的公众号或小程序的选项设置信息，如：地理位置上报，语音识别开关，
// 多客服开关。注意，获取各项选项设置信息，需要有授权方的授权，详见权限集说明。
func GetAuthorizerOption(componentAppid, authorizerAppid, optionName, componentAccessToken string) (option *Option, err error) {
	data := map[string]string{
		"component_appid":  componentAppid,
		"authorizer_appid": authorizerAppid,
		"option_name":      optionName,
	}
	option = &Option{}
	err = wechat.PostSchema(wechat.KindJson, "https://api.weixin.qq.com/cgi-bin/component/ api_get_authorizer_option?component_access_token="+componentAccessToken, data, option)
	if err != nil {
		return nil, err
	} else {
		return option, nil
	}
}
