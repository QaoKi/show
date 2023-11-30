package data

import (
	"bygame/common/proto/common_proto"
	"bygame/common/proto/fish_proto"
)

type FishData struct {
	ZhenZhuMap map[string][]int32 // 珍珠，需要落库

	ZhuanPanScore        int64                                // 转盘积分
	ZhuanPanTotalCount   int32                                // 累计转盘次数
	ZhuanPanRewardStatus map[string]common_proto.RewardStatus // 转盘次数奖励的状态，key为下标，从 0 开始

	GodBlessRewardMap map[string][]*fish_proto.GodBlessReward // 女神赐福，3种房间类型，又有3种不同的奖励

	ComboReward int64 // 连击奖励
}

// activity 转盘抽奖记录
type FishActivityZhuanPanPlayRecord struct {
	Mid       string `json:"mid"` // 玩家mid
	Time      int64  `json:"time"`
	TotalCoin int64  `json:"totalCoin"` // 总奖励
}

// activity 转盘积分记录
type FishActivityZhuanPanScoreRecord struct {
	Mid       string `json:"mid"` // 玩家mid
	Time      int64  `json:"time"`
	TotalCoin int64  `json:"totalCoin"` // 总奖励
}

// 女神赐福触发记录
type FishGodBlessTouchRecord struct {
	Mid       string `json:"mid"` // 玩家mid
	Time      int64  `json:"time"`
	TotalCoin int64  `json:"totalCoin"` // 总奖励
}

// 女神赐福领奖记录
type FishGodBlessGetReardRecord struct {
	Mid        string `json:"mid"` // 玩家mid
	Time       int64  `json:"time"`
	RewardType int32  `json:"rewardType"` // 第几阶段的奖励
	Coin       int64  `json:"coin"`       // 奖励
}
