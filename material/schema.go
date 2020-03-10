// Copyright 2018 orivil.com. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found at https://mit-license.org.

package material

type MediaType string

const (
	IMAGE MediaType = "image"
	VOICE MediaType = "voice"
	VIDEO MediaType = "video"
	THUMB MediaType = "thumb"
	NEWS  MediaType = "news"
)
