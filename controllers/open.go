package controllers

import (
	"bygame/common/data"
	"bygame/common/log"
	"bygame/common/utils"
	"bygame/models"
	"strings"
	"time"
)

type OpenController struct{}

type reqPing struct {
	BaseReq
}

type retPing struct {
	BaseRet
	Time int    `json:"time"`
	Ip   string `json:"ip"`
}

// ping
func (OpenController) Ping(req *reqPing, ret *retPing) {
	ret.Time = int(time.Now().Unix())
}

type reqGuestLogin struct {
	BaseReq
	DeviceId string `json:"deviceId" binding:"required"`
}

type retGuestLogin struct {
	BaseRet
	Token    string        `json:"token"`
	UserInfo data.UserInfo `json:"userInfo"`
}

// 游客登录 第三方登录是不是也可以这样统一处理
func (OpenController) GuestLogin(req *reqGuestLogin, ret *retGuestLogin) {
	u, err := models.GuestLogin(req.DeviceId)
	if err != nil {
		ret.ErrCode = 1
		ret.ErrMsg = err.Error()
		return
	}
	token, err := utils.CreateJwt(u.Mid)
	if err != nil {
		ret.ErrCode = 1
		ret.ErrMsg = err.Error()
		return
	}
	ret.UserInfo = u.UserInfo
	ret.Token = token
}

type reqZyLogin struct {
	BaseReq
	Sign string `json:"sign"`
}

type retZyLogin struct {
	BaseRet
	Token    string        `json:"token"`
	UserInfo data.UserInfo `json:"userInfo"`
}

func (OpenController) ZyLogin(req *reqZyLogin, ret *retZyLogin) {
	u, err := models.ZyLogin(req.Sign)
	if err != nil {
		ret.ErrCode = 1
		ret.ErrMsg = err.Error()
		return
	}
	token, err := utils.CreateJwtZy(u.Mid, u.UserInfo.Uid, u.UserInfo.Uid)
	if err != nil {
		ret.ErrCode = 1
		ret.ErrMsg = err.Error()
		return
	}
	ret.UserInfo = u.UserInfo
	ret.Token = token
}

type reqPopLogin struct {
	IdToken string `json:"idToken" binding:"required"`
}

type retPopLogin struct {
	BaseRet
	Token    string        `json:"token"`
	UserInfo data.UserInfo `json:"userInfo"`
}

func (OpenController) PopLogin(req *reqPopLogin, ret *retPopLogin) {
	log.Inf("pop登录 idToken: %v", req.IdToken)
	req.IdToken = strings.ReplaceAll(req.IdToken, "?p", "")
	u, err := models.PopLogin(req.IdToken)
	if err != nil {
		ret.ErrCode = 1
		ret.ErrMsg = err.Error()
		return
	}
	ret.UserInfo = u.UserInfo
	token, err := utils.CreateJwt(u.Mid)
	if err != nil {
		ret.ErrCode = 1
		ret.ErrMsg = err.Error()
		return
	}
	ret.Token = token
}
