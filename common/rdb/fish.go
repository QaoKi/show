package rdb

import (
	"bygame/common/log"
	"context"
	"fmt"
	"strconv"
)

func KeyFishGodBlessPool(roomType int32) string {
	return fmt.Sprintf("%v:godbless:%v", preFixFish, roomType)
}

const (
	keyGodBlessPoolCoin    = "poolCoin"
	keyGodBlessPoolAddTime = "addTime" // 服务器自己更新的时间，客户端触发的不改
	keyGodBlessPoolSubTime = "subTime" // 服务器自己更新的时间，客户端触发的不改
)

func GetFishGodBlessPoolInfo(roomType int32) (ok bool, count, addTime, subTime int64) {
	r := Client().HGetAll(context.Background(), KeyFishGodBlessPool(roomType))
	if r.Err() != nil {
		log.Err("get 女神赐福奖池失败，roomtype[%d], err[%v]", roomType, r.Err())
		return
	}

	m := r.Val()
	count, err := strconv.ParseInt(m[keyGodBlessPoolCoin], 10, 64)
	if err != nil {
		log.Err("get 女神赐福奖池失败：roomtype[%d], key[%s] Atoi, err[%v]", roomType, keyGodBlessPoolCoin, err)
		return
	}

	addTime, err = strconv.ParseInt(m[keyGodBlessPoolAddTime], 10, 64)
	if err != nil {
		log.Err("get 女神赐福奖池失败：roomtype[%d], key[%s] Atoi, err[%v]", roomType, keyGodBlessPoolAddTime, err)
		return
	}

	subTime, err = strconv.ParseInt(m[keyGodBlessPoolSubTime], 10, 64)
	if err != nil {
		log.Err("get 女神赐福奖池失败：roomtype[%d], key[%s] Atoi, err[%v]", roomType, keyGodBlessPoolSubTime, err)
		return
	}

	ok = true
	return
}

func SetFishGodBlessPoolInfo(roomType int32, count, addTime, subTime int64) bool {
	m := map[string]interface{}{}
	m[keyGodBlessPoolCoin] = count
	m[keyGodBlessPoolAddTime] = addTime
	m[keyGodBlessPoolSubTime] = subTime

	r := Client().HSet(context.Background(), KeyFishGodBlessPool(roomType), m)
	if r.Err() != nil {
		log.Err("set 女神赐福奖池失败，roomtype[%d], err[%v]", roomType, r.Err())
		return false
	}
	return true
}

func AddGodBlessPool(roomType int32, count int64) bool {
	ok, c, addT, subT := GetFishGodBlessPoolInfo(roomType)
	if !ok {
		return false
	}

	return SetFishGodBlessPoolInfo(roomType, c+count, addT, subT)
}
