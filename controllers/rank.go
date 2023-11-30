package controllers

import (
	"bygame/common/data"
	"bygame/common/def"
	"bygame/common/mdb"
	"bygame/common/rdb"
	"context"
	"fmt"

	"gopkg.in/mgo.v2/bson"
)

type RankController struct{}

type reqSlotRank struct {
	BaseReq
	Today bool `json:"today"`
	Page  int
}

type retSlotRank struct {
	BaseRet
	List []slotRankItem `json:"list"`
	Next bool           `json:"next"`
}

type slotRankItem struct {
	UserInfo data.UserInfo `json:"userInfo"`
	Score    float64       `json:"score"`
	Rank     int           `json:"rank"`
}

func (RankController) SlotRank(req *reqSlotRank, ret *retSlotRank) {
	if req.Page <= 0 {
		req.Page = 1
	}

	if req.Page < 4 {
		ret.Next = false
	}

	start := (req.Page - 1) * 50
	end := req.Page*50 - 1

	var rankId string
	if req.Today {
		rdb.TodayRankId(def.GameIdBaliTreasure)
	} else {
		rdb.YesterdayRankId(def.GameIdBaliTreasure)
	}

	zc := rdb.Client().ZRevRangeWithScores(context.TODO(), rdb.KeyCenterRank(rankId), int64(start), int64(end)).Val()
	m := make(map[string]float64)
	var mids []bson.ObjectId
	for _, v := range zc {
		mid := fmt.Sprint(v.Member)
		m[mid] = v.Score
		mids = append(mids, bson.ObjectIdHex(mid))
	}
	db := mdb.GetMdb()
	var slc []data.User
	db.C(mdb.DB_USER).Find(bson.M{"_id": bson.M{"$in": mids}}).All(&slc)
	um := make(map[string]data.UserInfo)
	for _, v := range slc {
		um[v.UserInfo.Mid] = v.UserInfo
	}
	for index, v := range mids {
		ret.List = append(ret.List, slotRankItem{UserInfo: um[v.Hex()], Score: m[v.Hex()], Rank: start + index + 1})
	}

	if req.Page < 4 && len(ret.List) == 50 {
		ret.Next = true
	}
}
