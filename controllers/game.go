package controllers

import (
	"bygame/common/conf"
	"bygame/common/data"
	"bygame/common/mdb"
	"bygame/common/rdb"
	"bygame/models"
	"context"
	"encoding/json"
	"fmt"

	"gopkg.in/mgo.v2/bson"
)

type GameController struct{}

type reqGameAddr struct {
	BaseReq
	GameId  string `json:"gameId"`
	Version int    `json:"version"`
}

type retGameAddr struct {
	BaseRet
	Addr   string `json:"addr"`   // 服务地址
	GameId string `json:"gameId"` // 游戏id
}

// 游戏内所有的数据流转都应该使用我们的mid现在先做个妥协

// 获取游戏服务器,通过这个接口做会话保持和负载
func (GameController) GameAddr(req *reqGameAddr, ret *retGameAddr) {
	db := mdb.GetMdb()
	var account data.Account
	db.C(mdb.DB_ACCOUNT).Find(bson.M{"mid": req.TokenId}).One(&account)
	var user data.User
	db.C(mdb.DB_USER).Find(bson.M{"_id": bson.ObjectIdHex(req.TokenId)}).One(&user)
	var uid string
	if account.Mid == req.TokenId && account.Platform == models.Zy {
		uid = account.Code
	} else {
		uid = user.UserInfo.Uid
	}
	serverStr := rdb.Get("game:user:server:" + uid).Val()
	var data struct {
		GameName string `json:"gameName"`
		ServerId string `json:"serverId"`
	}
	json.Unmarshal([]byte(serverStr), &data)
	if serverStr != "" && data.GameName != "" && data.ServerId != "" {
		// 有正在进行的游戏
		ret.GameId = conf.Cf.CenterConf.GetGameId(data.GameName)
		ret.Addr = data.ServerId
	} else {
		serverName := conf.Cf.CenterConf.GetServerName(req.GameId)
		if serverName == "" {
			ret.ErrCode = 1
			ret.ErrMsg = "Game id not found."
			return
		}
		sc := rdb.Client().SRandMember(context.TODO(), fmt.Sprintf(`game:%v:active`, serverName)).Val()
		if sc == "" {
			ret.ErrCode = 1
			ret.ErrMsg = "Game server not found."
			return
		}
		ret.Addr = sc
		ret.GameId = req.GameId
	}
}

type reqGameList struct {
	BaseReq
}

type retGameList struct {
	BaseRet
	List []*conf.GameInfo `json:"list"`
}

// GameList 获取游戏列表
func (GameController) GameList(req *reqGameList, ret *retGameList) {
	ret.List = conf.Cf.CenterConf.GameList
}

type reqPlayRoom struct {
	BaseReq
	GameId string `json:"gameId"`
}

type retPlayRoom struct {
	BaseRet
	PlayRoom []*conf.PlayRoomInfo `json:"playRoom"` // 子大厅列表
}

// PlayRoom 获取子大厅列表
func (GameController) PlayRoom(req *reqPlayRoom, ret *retPlayRoom) {
	ret.PlayRoom = conf.Cf.CenterConf.GetPlayRoom(req.GameId)
	if len(ret.PlayRoom) == 0 {
		ret.ErrCode = 1
		ret.ErrMsg = "未找到配置"
	}
}

type reqVerifyToken struct {
	BaseReq
}

type retVerifyToken struct {
	BaseRet
	UserInfo data.UserInfo `json:"userInfo"` // 用户信息
}

// verifyToken 验证token
func (GameController) VerifyToken(req *reqVerifyToken, ret *retVerifyToken) {
	db := mdb.GetMdb()
	var user data.User
	db.C(mdb.DB_USER).Find(bson.M{"_id": bson.ObjectIdHex(req.TokenId)}).One(&user)
	if user.Mid != req.TokenId {
		ret.ErrCode = 10900
		ret.ErrMsg = "authentication failure"
		return
	}
	ret.UserInfo = user.UserInfo
}
