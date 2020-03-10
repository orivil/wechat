// Copyright 2018 orivil.com. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found at https://mit-license.org.

package payment

import (
	"crypto/md5"
	"crypto/sha1"
	"fmt"
	"github.com/orivil/utils"
	"github.com/orivil/wechat"
	"time"
)

//type TicketProvider func() (ticket string, err error)

type Config struct {
	AppID string
	// 商户号
	MchID string
	// 商户Key
	ShopKey string
}

// 发起微信支付统一下单
// 如：
//	request := &PayRequest{
//		// 标题
//		Body: "书城充值中心",
//
//		// 商户订单号
//		OutTradeNo: order.Number,
//
//		// 金额
//		TotalFee: p.money,
//
//		// 用户 IP 地址
//		SpbillCreateIp: ctx.GetIP(),
//
//		// 设置支付成功通知地址
//		NotifyUrl: "http://" + ctx.Host + PayNotifyRoute,
//
//		// 交易类型（公众号支付）
//		TradeType: "JSAPI",
//
//		// 公众号支付必须带上 openid 参数
//		Openid: user.Openid,
//	}
//
//	// 大多数情况下可以通过 Referer 头信息获得支付域名, 因为微信规定只有在支付域名下才可以发起支付
//  var req *http.Request
//	payUrl := req.Header.Get("Referer")
//
// Tips: 返回的 UserConfig 可以自行配置 Debug
func InitPayment(request *PayRequest, cfg *Config, ticket, payUrl string) (uc *UserConfig, pay *ChoosePay, err error) {
	// 生成随机字符串
	nonceStr := string(utils.RandomBytes(32))
	request.BaseInfo = BaseInfo{
		AppID: cfg.AppID,

		// 商户号
		MchID: cfg.MchID,

		// 随机字符串
		NonceStr: nonceStr,
	}
	// 生成签名
	sign, err := wechat.SignSchema(request, md5.New(), cfg.ShopKey)
	if err != nil {
		return nil, nil, err
	}
	request.Sign = sign
	payRes := &payResponse{}
	// 发起统一下单请求并获得响应
	err = wechat.PostSchema(wechat.KindXml, "https://api.mch.weixin.qq.com/pay/unifiedorder", request, payRes)
	if err != nil {
		return nil, nil, err
	}
	if payRes.ReturnCode == "FAIL" {
		return nil, nil, fmt.Errorf("微信支付统一下单失败：%v\n", payRes.ReturnMsg)
	} else if payRes.ResultCode == "FAIL" {
		return nil, nil, fmt.Errorf("微信支付统一下单系统内部错误，错误码：%s，描述：%s\n", payRes.ErrCode, payRes.ErrCodeDes)
	} else {

		/**
		生成客户端配置数据
		*/
		now := time.Now().Unix()
		// 生成配置签名
		s := &configSign{
			Noncestr:    nonceStr,
			JsapiTicket: ticket,
			Timestamp:   now,
			Url:         payUrl,
		}
		cSign, err := wechat.SignSchema(s, sha1.New(), "")
		if err != nil {
			return nil, nil, fmt.Errorf("config 签名出错：%s\n", err)
		}
		// 生成配置文件
		uc = &UserConfig{
			AppID:     cfg.AppID,
			Timestamp: now,
			NonceStr:  nonceStr,
			Signature: cSign,
		}

		/**
		生成客户端支付数据
		*/

		// 生成支付签名
		ps := &paySign{
			AppID:     cfg.AppID,
			Timestamp: now,
			NonceStr:  nonceStr,
			Package:   "prepay_id=" + payRes.PrepayID,
			SignType:  "MD5",
		}
		pSign, err := wechat.SignSchema(ps, md5.New(), cfg.ShopKey)
		if err != nil {
			return nil, nil, fmt.Errorf("支付签名出错：%s\n", err)
		}
		// 生成支付文件
		pay = &ChoosePay{
			Timestamp: now,
			NonceStr:  nonceStr,
			Package:   "prepay_id=" + payRes.PrepayID,
			SignType:  "MD5",
			PaySign:   pSign,
		}
		return uc, pay, nil
	}
}

type ChoosePay struct {
	Timestamp int64  `json:"timestamp"`
	NonceStr  string `json:"nonceStr"`
	Package   string `json:"package"`
	SignType  string `json:"signType"`
	PaySign   string `json:"paySign"`
}

type UserConfig struct {
	AppID     string `json:"appId"`
	Timestamp int64  `json:"timestamp"`
	NonceStr  string `json:"nonceStr"`
	Signature string `json:"signature"`
}

type configSign struct {
	Noncestr    string `url:"noncestr"`
	JsapiTicket string `url:"jsapi_ticket"`
	Timestamp   int64  `url:"timestamp"`
	Url         string `url:"url"`
}

type paySign struct {
	AppID     string `url:"appId"`
	Timestamp int64  `url:"timeStamp"`
	NonceStr  string `url:"nonceStr"`
	Package   string `url:"package"`
	SignType  string `url:"signType"`
}

type PayRequest struct {
	BaseInfo
	Body           string `xml:"body" url:"body"`
	Detail         string `xml:"detail" url:"detail"`
	Attach         string `xml:"attach" url:"attach"`
	OutTradeNo     string `xml:"out_trade_no" url:"out_trade_no"`
	FeeType        string `xml:"fee_type" url:"fee_type"`
	TotalFee       int64  `xml:"total_fee" url:"total_fee"`
	SpbillCreateIp string `xml:"spbill_create_ip" url:"spbill_create_ip"`
	TimeStart      string `xml:"time_start" url:"time_start"`
	TimeExpire     string `xml:"time_expire" url:"time_expire"`
	GoodsTag       string `xml:"goods_tag" url:"goods_tag"`
	NotifyUrl      string `xml:"notify_url" url:"notify_url"`
	TradeType      string `xml:"trade_type" url:"trade_type"`
	ProductId      string `xml:"product_id" url:"product_id"`
	LimitPay       string `xml:"limit_pay" url:"limit_pay"`
	Openid         string `xml:"openid" url:"openid"`
	SceneInfo      string `xml:"scene_info" url:"scene_info"`
}

type payResponse struct {
	RetCode
	BaseInfo
	ResCode
	TradeType string `xml:"trade_type" url:"trade_type"`
	PrepayID  string `xml:"prepay_id" url:"prepay_id"`
	CodeUrl   string `xml:"code_url" url:"code_url"`
}

// 返回给用户的签名
//type userSignature struct {
//	NonceStr    string `url:"noncestr"`
//	JsApiTicket string `url:"jsapi_ticket"`
//	Timestamp   string `url:"timestamp"`
//	Url         string `url:"url"`
//}

type BaseInfo struct {
	AppID      string `xml:"appid" url:"appid"`
	MchID      string `xml:"mch_id" url:"mch_id"`
	DeviceInfo string `xml:"device_info" url:"device_info"`
	NonceStr   string `xml:"nonce_str" url:"nonce_str"`
	Sign       string `xml:"sign" url:"sign"`
	SignType   string `xml:"sign_type" url:"sign_type"`
}
