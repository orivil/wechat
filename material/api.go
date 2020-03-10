// Copyright 2018 orivil.com. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found at https://mit-license.org.

package material

import (
	"bytes"
	"encoding/json"
	"github.com/orivil/wechat"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

type UploadedMedia struct {
	MediaID string `json:"media_id" desc:"新增的永久素材的media_id"`
	Url     string `json:"url" desc:"新增的图片素材的图片URL（仅新增图片素材时会返回该字段）"`
}

type VideoDescription struct {
	Title        string `json:"title"`
	Introduction string `json:"introduction"`
}

// 上传一个或多个永久素材
func UploadMaterials(req *http.Request, kind MediaType, videoDesc *VideoDescription, token string) (results []*UploadedMedia, err error) {
	err = req.ParseMultipartForm(0)
	if err != nil {
		return nil, err
	}
	for _, headers := range req.MultipartForm.File {
		for _, header := range headers {
			f, err := header.Open()
			if err != nil {
				return nil, err
			}
			data, err := ioutil.ReadAll(f)
			if err != nil {
				return nil, err
			} else {
				result, err := UploadMaterial(kind, data, header.Filename, token, videoDesc)
				if err != nil {
					return nil, err
				} else {
					results = append(results, result)
				}
			}
		}
	}
	return results, nil
}

func UploadMaterial(mediaType MediaType, data []byte, filename, token string, videoDesc *VideoDescription) (res *UploadedMedia, err error) {
	uri := "https://api.weixin.qq.com/cgi-bin/material/add_material?access_token=" + token + "&type=" + string(mediaType)
	var vs url.Values
	if videoDesc != nil {
		desc, err := json.Marshal(videoDesc)
		if err != nil {
			return nil, err
		} else {
			vs = url.Values{"description": []string{string(desc)}}
		}
	}
	res = &UploadedMedia{}
	err = wechat.UploadFile(uri, data, "media", filename, vs, res)
	if err != nil {
		return nil, err
	} else {
		return res, nil
	}
}

func DelMaterial(mediaID, token string) error {
	ul := "https://api.weixin.qq.com/cgi-bin/material/del_material?access_token=" + token
	return wechat.PostSchema(wechat.KindJson, ul, map[string]string{"media_id": mediaID}, nil)
}

type Count struct {
	Voice int `json:"voice_count"`
	Video int `json:"video_count"`
	Image int `json:"image_count"`
	News  int `json:"news_count"`
}

// 获得素材统计
func CountMaterials(token string) (count *Count, err error) {
	count = &Count{}
	err = wechat.GetJson("https://api.weixin.qq.com/cgi-bin/material/get_materialcount?access_token="+token, count)
	if err != nil {
		return nil, err
	} else {
		return count, nil
	}
}

type MediaList struct {
	TotalCount int         `json:"total_count"`
	ItemCount  int         `json:"item_count"`
	Item       []MediaItem `json:"item"`
}

type MediaItem struct {
	MediaID    string `json:"media_id"`
	Name       string `json:"name"`
	UpdateTime uint64 `json:"update_time"`
	Url        string `json:"url"`
}

// 获得素材列表
func GetMedias(kind MediaType, token string, limit, offset int) (res *MediaList, err error) {
	ul := "https://api.weixin.qq.com/cgi-bin/material/batchget_material?access_token=" + token
	err = wechat.PostSchema(wechat.KindJson, ul, map[string]interface{}{
		"type":   kind,
		"offset": offset,
		"count":  limit,
	}, &res)
	return
}

// 获得素材内容, image 及 voice 类型直接返回二进制文件, video 及 news 返回 json 文件
func GetMedia(token, mediaID string) (data []byte, err error) {
	uri := "https://api.weixin.qq.com/cgi-bin/material/get_material?access_token=" + token
	res, err := http.Post(uri, "application/json;charset=utf-8", strings.NewReader(`{"media_id": "`+mediaID+`"}`))
	if err != nil {
		return nil, err
	} else {
		defer res.Body.Close()
		data, err = ioutil.ReadAll(res.Body)
		if err != nil {
			return nil, err
		}
		if bytes.Contains(data, []byte("errcode")) {
			e := &wechat.Error{}
			_ = json.Unmarshal(data, e)
			if e.ErrCode != 0 {
				return nil, e
			}
		}
		return data, nil
	}
}
