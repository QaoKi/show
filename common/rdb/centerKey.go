package rdb

import (
	"fmt"
	"time"
)

/*
	slot 两个排行榜 昨日排行榜，今日排行榜

	rankId slot_(日期) 每日排行榜 过期时间设置
*/

func TodayRankId(gameId string) string {
	return fmt.Sprintf("%v_%v", gameId, time.Now().Format("20060102"))
}

func YesterdayRankId(gameId string) string {
	return fmt.Sprintf("%v_%v", gameId, time.Now().AddDate(0, 0, -1).Format("20060102"))
}

func KeyCenterRank(rankId string) string {
	return fmt.Sprintf("%v:rank:%v", preFixCenter, rankId)
}
