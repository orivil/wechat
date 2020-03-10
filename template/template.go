// Copyright 2018 orivil.com. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found at https://mit-license.org.

// template 包用于模板消息
package template

import (
	"errors"
	"github.com/orivil/wechat"
)

/***
行业代码查询

主行业	副行业	代码
IT科技	互联网/电子商务	1
IT科技	IT软件与服务	2
IT科技	IT硬件与设备	3
IT科技	电子技术	4
IT科技	通信与运营商	5
IT科技	网络游戏	6
金融业	银行	7
金融业	基金理财信托	8
金融业	保险	9
餐饮	    餐饮	10
酒店旅游	酒店	11
酒店旅游	旅游	12
运输与仓储	快递	13
运输与仓储	物流	14
运输与仓储	仓储	15
教育	培训	16
教育	院校	17
政府与公共事业	学术科研	18
政府与公共事业	交警	19
政府与公共事业	博物馆	20
政府与公共事业	公共事业非盈利机构	21
医药护理	医药医疗	22
医药护理	护理美容	23
医药护理	保健与卫生	24
交通工具	汽车相关	25
交通工具	摩托车相关	26
交通工具	火车相关	27
交通工具	飞机相关	28
房地产	建筑	29
房地产	物业	30
消费品	消费品	31
商业服务	法律	32
商业服务	会展	33
商业服务	中介服务	34
商业服务	认证	35
商业服务	审计	36
文体娱乐	传媒	37
文体娱乐	体育	38
文体娱乐	娱乐休闲	39
印刷	印刷	40
其它	其它	41
*/

// 设置行业可在微信公众平台后台完成，每月可修改行业1次
func SetIndustry(token string, industry []int) error {
	ul := "https://api.weixin.qq.com/cgi-bin/template/api_set_industry?access_token=" + token
	var param map[string]int
	ln := len(industry)
	if ln == 1 {
		param = map[string]int{"industry_id1": industry[0]}
	} else if ln == 2 {
		param = map[string]int{
			"industry_id1": industry[0],
			"industry_id2": industry[1],
		}
	} else {
		return errors.New("未设置行业")
	}
	return wechat.PostSchema(wechat.KindJson, ul, param, nil)
}

type Industry struct {
	Primary   *Name `json:"primary_industry"`
	Secondary *Name `json:"secondary_industry"`
}

type Name struct {
	First  string `json:"first_class"`
	Second string `json:"second_class"`
}

// 获取设置的行业信息
func GetIndustry(token string) (industry *Industry, err error) {
	ul := "https://api.weixin.qq.com/cgi-bin/template/get_industry?access_token=" + token
	err = wechat.GetJson(ul, &industry)
	return
}

func GetTemplateID(token, shortID string) (long string, err error) {
	ul := "https://api.weixin.qq.com/cgi-bin/template/api_add_template?access_token=" + token
	var res struct {
		Long string `json:"template_id"`
	}
	err = wechat.PostSchema(wechat.KindJson, ul, map[string]interface{}{
		"template_id_short": shortID,
	}, &res)
	if err != nil {
		return "", err
	} else {
		return res.Long, nil
	}
}

func GetTemplates(token string) (templates []*Template, err error) {
	ul := "https://api.weixin.qq.com/cgi-bin/template/get_all_private_template?access_token=" + token
	var res = struct {
		Templates []*Template `json:"template_list"`
	}{}
	err = wechat.GetJson(ul, &res)
	if err != nil {
		return nil, err
	} else {
		return res.Templates, nil
	}
}

func DelTemplate(token, id string) error {
	ul := "https://api.weixin.qq.com/cgi-bin/template/del_private_template?access_token=" + token
	return wechat.PostSchema(wechat.KindJson, ul, map[string]string{"template_id": id}, nil)
}

type Message struct {
	ToUser      string          `json:"touser"`
	TemplateID  string          `json:"template_id"`
	Url         string          `json:"url,omitempty"`
	MiniProgram *MiniProgram    `json:"miniprogram,omitempty"`
	Data        map[string]Data `json:"data,omitempty"`
}

func (msg *Message) Send(token, openID string) (msgID int64, err error) {
	ul := "https://api.weixin.qq.com/cgi-bin/message/template/send?access_token=" + token
	var res struct {
		MsgID int64 `json:"msgid"`
	}
	if openID != "" {
		msg.ToUser = openID
	}
	err = wechat.PostSchema(wechat.KindJson, ul, msg, &res)
	if err != nil {
		return 0, err
	} else {
		return res.MsgID, nil
	}
}

type Data struct {
	Value string `json:"value"`
	Color string `json:"color"`
}

type MiniProgram struct {
	AppID    string `json:"appid"`
	PagePath string `json:"pagepath"`
}

type Template struct {
	ID      string `json:"template_id"`
	Title   string `json:"title"`
	Primary string `json:"primary_industry"`
	Deputy  string `json:"deputy_industry"`
	Content string `json:"content"`
	Example string `json:"example"`
}
