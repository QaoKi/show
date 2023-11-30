package controllers

import (
	"bygame/common/mdb"

	"gopkg.in/mgo.v2/bson"
)

type SlotController struct{}

// 游戏记录

type reqSlotRecord struct {
	BaseReq
	Size int `json:"size"`
	Page int `json:"page"`
}

type retSlotRecord struct {
	BaseRet
	Next bool         `json:"next"`
	List []recordItem `json:"list"`
}

type recordItem struct {
	Id    string `json:"id" bson:"vid"`
	Ante  int    `json:"ante" bson:"ante"`
	Bonus int    `json:"bonus" bson:"bonus"`
}

func (SlotController) SlotRecord(req *reqSlotRecord, ret *retSlotRecord) {
	if req.Size == 0 {
		req.Size = 50
	}
	if req.Page < 1 {
		req.Page = 1
	}
	db := mdb.GetMdb()
	skip := (req.Page - 1) * req.Size
	limit := req.Size
	var slc []recordItem
	db.C(mdb.DB_SLOTRECORD).Find(bson.M{"mid": req.TokenId}).Sort("-_id").Skip(skip).Limit(limit).All(&slc)
	ret.List = slc
	ret.Next = len(slc) >= req.Size
}
