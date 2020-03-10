// Copyright 2018 orivil.com. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found at https://mit-license.org.

package payment

import (
	"bytes"
	"crypto/md5"
	"encoding/xml"
	"fmt"
	"github.com/orivil/wechat"
	"github.com/pkg/errors"
	"net/http"
)

// TODO: 最多只支持 4 个代金券, 否则签名验证不通过
// 接收订单通知, 收到通知后还需要响应微信服务器订单是否成功处理
func ListenNotify(r *http.Request, shopKey string) (notify *PayNotify, err error) {
	notify = &PayNotify{}
	decoder := xml.NewDecoder(r.Body)
	// see: https://stackoverflow.com/questions/35191202/unmarshal-xml-with-unescaped-character-inside?rq=1
	decoder.Strict = false
	err = decoder.Decode(notify)
	if err != nil {
		return nil, err
	}

	// 业务结果
	if !notify.RetCode.IsSuccess() {
		return nil, notify.RetCode
		// 系统错误
	} else if !notify.ResCode.IsSuccess() {
		return nil, notify.ResCode
	} else {
		// 验证签名
		originSign := notify.Sign
		notify.Sign = ""
		sign, err := wechat.SignSchema(notify, md5.New(), shopKey)
		if err != nil {
			return nil, err
		}
		if originSign != sign {
			return nil, errors.New("签名错误!")
		} else {
			return notify, nil
		}
	}
}

// 响应微信服务器订单已成功处理
func ResponseOrderSuccess(writer http.ResponseWriter) {
	buf := bytes.NewBuffer(nil)
	data := &RetCode{
		ReturnCode: "SUCCESS",
		ReturnMsg:  "OK",
	}
	_ = xml.NewEncoder(buf).Encode(data)
	writer.Header().Set("Content-Type", "application/xml;charset=UTF-8")
	_, _ = writer.Write(buf.Bytes())
}

type PayNotify struct {
	RetCode
	ResCode
	AppID              string `xml:"appid" url:"appid"`
	MchID              string `xml:"mch_id" url:"mch_id"`
	SubAppID           string `xml:"sub_appid" url:"sub_appid"`
	SubMchID           string `xml:"sub_mch_id" url:"sub_mch_id"`
	DeviceInfo         string `xml:"device_info" url:"device_info"`
	NonceStr           string `xml:"nonce_str" url:"nonce_str"`
	Sign               string `xml:"sign" url:"sign"`
	Openid             string `xml:"openid" url:"openid"`
	IsSubscribe        string `xml:"is_subscribe" url:"is_subscribe"`
	SubOpenid          string `xml:"sub_openid" url:"sub_openid"`
	SubIsSubscribe     string `xml:"sub_is_subscribe" url:"sub_is_subscribe"`
	TradeType          string `xml:"trade_type" url:"trade_type"`
	BankType           string `xml:"bank_type" url:"bank_type"`
	TotalFee           int    `xml:"total_fee" url:"total_fee"`
	FeeType            string `xml:"fee_type" url:"fee_type"`
	CashFee            int64  `xml:"cash_fee" url:"cash_fee"`
	CashFeeType        string `xml:"cash_fee_type" url:"cash_fee_type"`
	SettlementTotalFee int    `xml:"settlement_total_fee" url:"settlement_total_fee"`
	CouponFee          int    `xml:"coupon_fee" url:"coupon_fee"`
	CouponCount        int    `xml:"coupon_count" url:"coupon_count"`

	// 代金券类型
	CouponType0 string `xml:"coupon_type_0" url:"coupon_type_0"`
	// 代金券ID
	CouponID0 string `xml:"coupon_id_0" url:"coupon_id_0"`
	// 单个代金券支付金额
	CouponFee0 int `xml:"coupon_fee_0" url:"coupon_fee_0"`

	CouponType1 string `xml:"coupon_type_1" url:"coupon_type_1"`
	CouponID1   string `xml:"coupon_id_1" url:"coupon_id_1"`
	CouponFee1  int    `xml:"coupon_fee_1" url:"coupon_fee_1"`

	CouponType2 string `xml:"coupon_type_2" url:"coupon_type_2"`
	CouponID2   string `xml:"coupon_id_2" url:"coupon_id_2"`
	CouponFee2  int    `xml:"coupon_fee_2" url:"coupon_fee_2"`

	CouponType3 string `xml:"coupon_type_3" url:"coupon_type_3"`
	CouponID3   string `xml:"coupon_id_3" url:"coupon_id_3"`
	CouponFee3  int    `xml:"coupon_fee_3" url:"coupon_fee_3"`

	TransactionID string `xml:"transaction_id" url:"transaction_id"`
	OutTradeNo    string `xml:"out_trade_no" url:"out_trade_no"`
	Attach        string `xml:"attach" url:"attach"`
	TimeEnd       string `xml:"time_end" url:"time_end"`
}

// 通知结果
type RetCode struct {
	ReturnCode string `xml:"return_code" url:"return_code"`
	ReturnMsg  string `xml:"return_msg" url:"return_msg"`
}

func (rc RetCode) IsSuccess() bool {
	return rc.ReturnCode == "SUCCESS"
}

func (rc RetCode) Error() string {
	return fmt.Sprintf("return_code: %s return_msg: %s", rc.ReturnCode, rc.ReturnMsg)
}

// 系统错误
type ResCode struct {
	ResultCode string `xml:"result_code" url:"result_code"`
	ErrCode    string `xml:"err_code" url:"err_code"`
	ErrCodeDes string `xml:"err_code_des" url:"err_code_des"`
}

func (rc ResCode) IsSuccess() bool {
	return rc.ResultCode == "SUCCESS"
}

func (rc ResCode) Error() string {
	return fmt.Sprintf("err_code: %s err_code_des: %s", rc.ErrCode, rc.ErrCodeDes)
}
