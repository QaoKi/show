package controllers

/*
	因为现在domino和slots是ts实现的所有有些调用只能通过接口
*/

import (
	"bygame/common/data"
	"bygame/common/log"
	"bygame/common/mdb"
	"bygame/models"

	"gopkg.in/mgo.v2/bson"
)

type TestController struct{}

type reqGetUserByUid struct {
	Uid string `json:"uid"`
}

type retGetUserByUid struct {
	BaseRet
	UserInfo data.UserInfo `json:"userInfo"`
}

// 暂时根据uid获取信息 目前只会做测试使用
func (TestController) GetUserByUid(req *reqGetUserByUid, ret *retGetUserByUid) {
	// 可能是uid 也可能是zyuid
	db := mdb.GetMdb()
	var u data.User
	db.C(mdb.DB_USER).Find(bson.M{"userinfo.uid": req.Uid}).One(&u)
	if u.Mid == "" {
		ret.ErrCode = 1
		ret.ErrMsg = "user not found"
		return
	}
	if u.UserInfo.Account.Platform == models.Zy {
		u2, err := models.ZyLogin(u.UserInfo.Account.Sign)
		if err != nil {
			ret.ErrCode = 1
			ret.ErrMsg = "zyLogin err " + err.Error()
			return
		}
		ret.UserInfo = u2.UserInfo
	} else {
		ret.UserInfo = u.UserInfo
	}
}

type reqAddCoin struct {
	BaseReq
	Uid       string `json:"uid"`
	Increment int    `json:"increment"`
}

type retAddCoin struct {
	BaseRet
}

// 目前只会做测试使用 加币
func (TestController) AddCoin(req *reqAddCoin, ret *retAddCoin) {
	db := mdb.GetMdb()
	var account data.Account
	db.C(mdb.DB_ACCOUNT).Find(bson.M{"code": req.Uid}).One(&account)
	if account.Code == req.Uid {
		err := models.AddCoinZy(account.Sign, req.Increment)
		log.Dbg("之艺玩家加金币 uid: %v,coin: %v, err: %v", req.Uid, req.Increment, err)
		if err != nil {
			ret.ErrCode = 1
			ret.ErrMsg = err.Error()
			return
		}
		return
	}
	selecter := bson.M{"userinfo.uid": req.Uid}
	if req.Increment < 0 {
		selecter["userinfo.coin"] = bson.M{"$gte": req.Increment}
	}
	err := db.C(mdb.DB_USER).Update(selecter, bson.M{"$inc": bson.M{"userinfo.coin": req.Increment}})
	if err != nil {
		ret.ErrCode = 1
	}
}

type reqTestAddPopCoin struct {
	Uid  string `json:"uid"`
	Coin int    `json:"coin"`
}

type retTestAddPopCoin struct {
	BaseRet
}

func (TestController) TestAddPopCoin(req *reqTestAddPopCoin, ret *retTestAddPopCoin) {
	db := mdb.GetMdb()
	var user data.User
	db.C(mdb.DB_USER).Find(bson.M{"userinfo.uid": req.Uid}).One(&user)
	if user.UserInfo.Uid != req.Uid {
		ret.ErrCode = 1
		ret.ErrMsg = "未找到用户"
		return
	}
	if user.UserInfo.Account.Platform != models.Pop {
		ret.ErrCode = 1
		ret.ErrMsg = "非pop用户"
		return
	}

	err := data.PopAddCoin(int64(req.Coin), user.UserInfo.Account.AccessToken)
	if err != nil {
		ret.ErrCode = 1
		ret.ErrMsg = err.Error()
		return
	}
}
