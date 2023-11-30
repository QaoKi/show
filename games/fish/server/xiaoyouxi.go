package fish

import (
	"bygame/common/conf"
	"bygame/common/conf/fish/cfg"
	"bygame/common/log"
	"bygame/common/proto/common_proto"
	"bygame/common/proto/fish_proto"
	"bygame/common/rdb"
	"bygame/common/utils"

	"strconv"
	"time"
)

// ************************ 拉霸小游戏 **************************

func touchLaBa(u *user, roomType int32) (succ bool, data *fish_proto.LabaData) {
	data = &fish_proto.LabaData{}

	strRoomType := strconv.Itoa(int(roomType))
	laBaConfig := conf.Cf.FishTable.XiaoYouXiLaBa.GetDataList()
	// 计算一下底注
	zhenZhuList := u.fishData.ZhenZhuMap[strRoomType]
	for i := 0; i < int(_zhenZhuTouchCount); i++ {
		data.DiZhu += float32(zhenZhuList[i])
	}
	data.DiZhu /= float32(_zhenZhuTouchCount)
	log.Inf("玩家珍珠数量[%d], 底注[%f]", len(u.fishData.ZhenZhuMap[strRoomType]), data.DiZhu)

	// 随机一个滚轴
	randIndex := utils.RandInt(0, len(laBaConfig)-1)

	// 取出所有鱼的倍率，放到map中
	mapRate := make(map[int32]int32)
	for _, fishRate := range laBaConfig[0].FishRate {
		mapRate[fishRate.FishId] = fishRate.Rate
	}

	rate := int32(0)
	for _, fishId := range laBaConfig[randIndex].Roller {
		if v, b := mapRate[fishId]; b {
			rate += v
		}
	}
	// 奖励的金币数量
	data.CfgId = laBaConfig[randIndex].Id
	data.Coin = int64(float64(rate) * float64(data.DiZhu))
	succ = true

	log.Inf("随机数[%d], 配置id[%d], 奖励金币数[%d]", randIndex, data.CfgId, data.Coin)
	return
}

func fishPlayLaBa(mid string, req *fish_proto.ReqFishPlayLaBa, ret *fish_proto.RetFishPlayLaBa) (code int32) {
	code = int32(common_proto.ErrorCode_Fail)

	log.Inf("玩家[%s]玩拉霸", mid)

	u, b := _userM.getUser(mid)
	if !b {
		log.Wrn("失败：没有找到userinfo")
		return
	}

	if b := u.playLaBa(); !b {
		log.Wrn("失败：玩家没有拉霸buffer")
		return
	}

	// 发奖
	if succ := u.userInfo.AddCoin(u.playData.labaData.Coin); !succ {
		log.Wrn("失败：加钱失败")
		return
	}

	code = int32(common_proto.ErrorCode_Succ)

	// 通知其他人
	ev := &fish_proto.EventFishLaBaResult{}
	ev.Mid = u.userInfo.Mid
	ev.Coin = u.playData.labaData.Coin
	u.playData.room.send2all(ev)

	return
}

// ************************ buffer转盘 **************************

func touchBufferZhuanPan(u *user, gunValue int32, num int32) (succ bool, dataList []*fish_proto.BufferZhuanPanData) {
	cfgList := conf.Cf.FishTable.XiaoYouXiBufferZhuanPanConfig.GetDataList()
	if len(cfgList) == 0 {
		log.Wrn("失败：没有找到buffer转盘的配置")
		return
	}

	cellSlc := []utils.RandomItem{}
	for _, cfg := range cfgList {
		pro := float32(0)
		for _, p := range cfg.Prob {
			pro += p
		}
		cellSlc = append(cellSlc, utils.RandomItem{Id: cfg.CellNum, Weight: pro})
	}

	for i := 0; i < int(num); i++ {
		// 随机一个格子数
		cellNum := utils.RandomWeight(cellSlc)
		log.Inf("随机到的格子数：%d", cellNum)

		cfg := conf.Cf.FishTable.XiaoYouXiBufferZhuanPanConfig.Get(cellNum)
		if cfg == nil {
			log.Wrn("失败: 没有找到格子数对应的配置")
			continue
		}

		// 随机一个倍率
		rateSlc := []utils.RandomItem{}
		for index, pro := range cfg.Prob {
			rateSlc = append(rateSlc, utils.RandomItem{Id: int32(index), Weight: pro})
		}

		index := utils.RandomWeight(rateSlc)
		log.Inf("随机到的转盘下标[%d]", index)

		if index < 0 || index >= int32(len(cfg.Rate)) {
			log.Wrn("失败：下标不对，倍率的配置长度为[%d]", len(cfg.Rate))
			continue
		}

		rate := cfg.Rate[index]
		coin := int64(rate * gunValue)
		log.Inf("倍率[%d], 炮值[%d], 奖励的金币数[%d]", rate, gunValue, coin)
		dataList = append(dataList, &fish_proto.BufferZhuanPanData{
			CellNum:  cellNum,
			GunValue: gunValue,
			Index:    index,
			Rate:     rate,
			Coin:     coin,
		})
	}

	succ = true

	return
}

func fishPlayBufferZhuanPan(mid string, req *fish_proto.ReqFishPlayBufferZhuanPan, ret *fish_proto.RetFishPlayBufferZhuanPan) (code int32) {
	code = int32(common_proto.ErrorCode_Fail)

	log.Inf("玩家[%s]玩buffer转盘, index[%d]", mid, req.Index)

	u, b := _userM.getUser(mid)
	if !b {
		log.Wrn("失败：没有找到userinfo")
		return
	}

	if b := u.playBufferZhuanPan(); !b {
		log.Wrn("失败：玩家身上没有转盘buffer")
		return
	}

	if int(req.Index) >= len(u.playData.bufferZhuanPanDataList) {
		log.Wrn("失败：index 超过了奖励的长度")
		return
	}

	// 发奖
	if succ := u.userInfo.AddCoin(u.playData.bufferZhuanPanDataList[req.Index].Coin); !succ {
		log.Wrn("失败：加钱失败")
		return
	}

	code = int32(common_proto.ErrorCode_Succ)

	// 通知其他人
	ev := &fish_proto.EventFishBufferZhuanPanResult{}
	ev.Mid = u.userInfo.Mid
	ev.Coin = u.playData.bufferZhuanPanDataList[req.Index].Coin
	u.playData.room.send2all(ev)

	return
}

// ************************ 提莫的蘑菇庄园 **************************

func touchMoGu(u *user, gunValue, num int32) (succ bool, dataList []*fish_proto.MoguData) {
	cfgList := conf.Cf.FishTable.XiaoYouXiMoGu.GetDataList()
	if len(cfgList) == 0 {
		log.Wrn("失败：没有找到蘑菇庄园配置")
		return
	}

	slc := []utils.RandomItem{}
	for _, cfg := range cfgList {
		slc = append(slc, utils.RandomItem{Id: cfg.Id, Weight: cfg.Prob})
	}

	for i := 0; i < int(num); i++ {
		randomId := utils.RandomWeight(slc)
		log.Inf("随机数[%d]", randomId)

		cfg := conf.Cf.FishTable.XiaoYouXiMoGu.Get(randomId)
		if cfg == nil {
			log.Wrn("失败：没有找到随机数对应的配置")
			continue
		}

		coin := int64(gunValue * cfg.Rate)
		dataList = append(dataList, &fish_proto.MoguData{
			GunValue: gunValue,
			CfgId:    cfg.Id,
			Name:     cfg.Name,
			Rate:     cfg.Rate,
			Coin:     coin,
		})

		log.Inf("倍率[%d], 炮值[%d], 奖励金币数[%d]", cfg.Rate, gunValue, coin)
	}
	succ = true

	return
}

func fishPlayMoGu(mid string, req *fish_proto.ReqFishPlayMoGu, ret *fish_proto.RetFishPlayMoGu) (code int32) {
	code = int32(common_proto.ErrorCode_Fail)

	log.Inf("玩家[%s]玩蘑菇庄园, index[%d]", mid, req.Index)

	u, b := _userM.getUser(mid)
	if !b {
		log.Wrn("失败：没有找到userinfo")
		return
	}

	if b := u.playMoGu(); !b {
		log.Wrn("失败：玩家没有蘑菇庄园buffer")
		return
	}

	if int(req.Index) >= len(u.playData.moguDataList) {
		log.Wrn("失败：index 超过了奖励的长度")
		return
	}

	// 发奖
	if succ := u.userInfo.AddCoin(u.playData.moguDataList[req.Index].Coin); !succ {
		log.Wrn("失败：加钱失败")
		return
	}
	code = int32(common_proto.ErrorCode_Succ)

	// 通知其他人
	ev := &fish_proto.EventFishMoGuResult{}
	ev.Mid = u.userInfo.Mid
	ev.Coin = u.playData.moguDataList[req.Index].Coin
	u.playData.room.send2all(ev)

	return
}

// ************************ activity转盘 **************************

func fishGetActivityZhuanPanInfo(mid string, req *fish_proto.ReqFishGetActivityZhuanPanInfo, ret *fish_proto.RetFishGetActivityZhuanPanInfo) (code int32) {
	code = int32(common_proto.ErrorCode_Fail)

	log.Inf("玩家[%s]获取activity转盘数据", mid)

	u, b := _userM.getUser(mid)
	if !b {
		log.Wrn("失败：没有找到userinfo")
		return
	}

	ret.ZhuanPanScore = u.fishData.ZhuanPanScore
	ret.TotalCoount = u.fishData.ZhuanPanTotalCount
	log.Inf("转盘积分[%d], 累计转的次数[%d]", u.fishData.ZhuanPanScore, u.fishData.ZhuanPanTotalCount)

	code = int32(common_proto.ErrorCode_Succ)
	if u.fishData.ZhuanPanRewardStatus == nil {
		log.Wrn("没有找到次数奖励状态")
		return
	}

	cfg := conf.Cf.FishTable.XiaoYouXiActivityZhuanPan.GetDataList()
	if len(cfg) == 0 {
		log.Wrn("失败：没有找到转盘配置")
		return
	}

	for _, c := range cfg[0].CountReward {
		count := strconv.Itoa(int(c.One))
		ret.RewardStatus = append(ret.RewardStatus, u.fishData.ZhuanPanRewardStatus[count])
	}

	log.Inf("奖励状态[%v]", ret.RewardStatus)

	//todo 记录
	return
}

func fishPlayActivityZhuanPan(mid string, req *fish_proto.ReqFishPlayActivityZhuanPan, ret *fish_proto.RetFishPlayActivityZhuanPan) (code int32) {
	code = int32(common_proto.ErrorCode_Fail)
	u, b := _userM.getUser(mid)
	if !b {
		log.Wrn("失败：没有找到userinfo")
		return
	}

	cfg := conf.Cf.FishTable.XiaoYouXiActivityZhuanPan.GetDataList()
	if len(cfg) == 0 {
		log.Wrn("失败：没有找到配置")
		return
	}

	log.Inf("玩家[%s]玩activity转盘，身上积分[%d]，累计转盘次数[%d]", mid, u.fishData.ZhuanPanScore, u.fishData.ZhuanPanTotalCount)

	// 积分是否够
	if u.fishData.ZhuanPanScore < int64(cfg[0].Jifen) {
		log.Wrn("失败：转盘积分不够，需要[%d]", cfg[0].Jifen)
		return
	}

	slc := []utils.RandomItem{}
	for _, c := range cfg {
		slc = append(slc, utils.RandomItem{Id: c.Id, Weight: c.Prob})
	}

	cfgId := utils.RandomWeight(slc)
	log.Inf("随机到的配置id[%d]", cfgId)

	ret.CfgId = cfgId
	ret.Coin = int64(cfg[cfgId].Coin)
	// 发奖
	if succ := u.userInfo.AddCoin(ret.Coin); !succ {
		log.Wrn("失败：加钱失败")
		return
	}

	code = int32(common_proto.ErrorCode_Succ)

	// 减积分
	u.addActivityZhuanPanScore(int64(cfg[0].Jifen))

	// 加次数
	u.fishData.ZhuanPanTotalCount++
	setActivityZhuanPanRewardStatus(u, cfg[0].CountReward)

	// todo 加记录
	return
}

func setActivityZhuanPanRewardStatus(u *user, countReward []*cfg.TwoIntBean) {
	if u.fishData.ZhuanPanRewardStatus == nil {
		u.fishData.ZhuanPanRewardStatus = make(map[string]common_proto.RewardStatus)
	}

	log.Inf("改变次数奖励状态，改变前[%v]", u.fishData.ZhuanPanRewardStatus)
	for _, c := range countReward {
		strCount := strconv.Itoa(int(c.One))
		if u.fishData.ZhuanPanTotalCount < int32(c.One) {
			break
		}
		if u.fishData.ZhuanPanRewardStatus[strCount] == common_proto.RewardStatus_NotAvailable {
			u.fishData.ZhuanPanRewardStatus[strCount] = common_proto.RewardStatus_Available
		}
	}
	log.Inf("改变后[%v]", u.fishData.ZhuanPanRewardStatus)
}

func fishGetActivityZhuanPanReward(mid string, req *fish_proto.ReqFishGetActivityZhuanPanReward, ret *fish_proto.RetFishGetActivityZhuanPanReward) (code int32) {
	code = int32(common_proto.ErrorCode_Fail)

	log.Inf("玩家[%s]领转盘次数奖励，count[%d]", mid, req.Count)

	u, b := _userM.getUser(mid)
	if !b {
		log.Wrn("失败：没有找到userinfo")
		return
	}

	cfg := conf.Cf.FishTable.XiaoYouXiActivityZhuanPan.GetDataList()
	if len(cfg) == 0 {
		log.Wrn("失败：没有找到配置")
		return
	}

	if u.fishData.ZhuanPanRewardStatus == nil {
		return
	}
	// 验证状态
	strCount := strconv.Itoa(int(req.Count))
	status, _ := u.fishData.ZhuanPanRewardStatus[strCount]
	if status != common_proto.RewardStatus_Available {
		log.Wrn("失败：奖励状态不对，status[%d]", status)
		return
	}

	coin := int64(0)
	for _, c := range cfg[0].CountReward {
		if c.One == req.Count {
			coin = int64(c.Two)
			break
		}
	}

	log.Inf("奖励[%d]", coin)
	ret.Coin = coin
	// 发奖
	if succ := u.userInfo.AddCoin(coin); !succ {
		log.Wrn("失败：加钱失败")
		return
	}
	code = int32(common_proto.ErrorCode_Succ)

	// 改状态
	u.fishData.ZhuanPanRewardStatus[strCount] = common_proto.RewardStatus_Got
	return
}

// ************************ 女神赐福 **************************

func touchGodBless(u *user, r *room, gunValue int32) {
	strRoomType := strconv.Itoa(int(r.roomType))
	// 如果已经有了，不再触发
	rewardList, ok := u.fishData.GodBlessRewardMap[strRoomType]
	if ok && len(rewardList) > 0 {
		return
	}

	// 判断是否触发女神赐福
	config := conf.Cf.FishTable.XiaoYouXiGoddessBless.Get(r.roomType)
	if config == nil {
		return
	}

	ok, poolCoin, _, _ := rdb.GetFishGodBlessPoolInfo(r.roomType)
	if !ok {
		return
	}

	if poolCoin < int64(config.Cond) {
		return
	}

	prob := float32(gunValue) / float32(config.Prob)

	//log.Inf("=====testlog 女神赐福，概率[%.2f]", prob)

	succ := utils.Rate1w(int(prob * float32(10000)))
	if !succ {
		return
	}

	slc := []utils.RandomItem{}
	for index, rewardConfig := range config.Reward {
		slc = append(slc, utils.RandomItem{Id: int32(index), Weight: rewardConfig.Prob})
	}

	randIndex := utils.RandomWeight(slc)
	rewardConfig := config.Reward[randIndex]

	allCoin := utils.RandInt64(int64(rewardConfig.Reward.One), int64(rewardConfig.Reward.Two))
	log.Inf("触发女神赐福，概率[%.2f]，总奖励[%d]", prob, allCoin)

	// 分三次给，青铜，白银，黄金
	qingTongCoin := (allCoin / 15) * 4
	baiyinCoin := (allCoin / 15) * 5
	huangjinCoin := allCoin - qingTongCoin - baiyinCoin
	coinList := []int64{qingTongCoin, baiyinCoin, huangjinCoin}

	gbr := []*fish_proto.GodBlessReward{}

	for index, coin := range coinList {
		reward := &fish_proto.GodBlessReward{}
		reward.RewardType = fish_proto.GodBlessRewardType(index)
		reward.Coin = coin
		if index == 0 {
			reward.RewardStatus = common_proto.RewardStatus_Available
		}
		gbr = append(gbr, reward)
	}

	// 双判
	rewardList, ok = u.fishData.GodBlessRewardMap[strRoomType]
	if ok && len(rewardList) > 0 {
		log.Inf("双判已经触发女神赐福")
		return
	}

	// 奖池扣除
	rdb.AddGodBlessPool(r.roomType, allCoin)

	u.fishData.GodBlessRewardMap[strRoomType] = gbr

	// 广播桌子
	event := &fish_proto.EventFishTouchGodBlessToUser{
		RoomType:   r.roomType,
		TotalCoin:  allCoin,
		BronzeCoin: qingTongCoin,
		SilverCoin: baiyinCoin,
		GoldCoin:   huangjinCoin,
	}
	r.send2all(event)

	// todo广播全服
	// event := &fish_proto.EventFishUserTouchGodBless{}
}

func fishGetGodBlessInfo(mid string, req *fish_proto.ReqFishGetGodBlessInfo, ret *fish_proto.RetFishGetGodBlessInfo) (code int32) {
	code = int32(common_proto.ErrorCode_Fail)

	u, b := _userM.getUser(mid)
	if !b {
		log.Wrn("玩家[%s]获取女神赐福数据失败：没有找到userinfo", mid)
		return
	}

	// 获取奖池大小
	ok, poolCoin, _, _ := rdb.GetFishGodBlessPoolInfo(req.RoomType)
	if !ok {
		log.Wrn("玩家[%s]获取女神赐福数据失败：没有找到奖池", mid)
		return
	}

	ret.PoolCoin = poolCoin

	strRoomType := strconv.Itoa(int(req.RoomType))
	rewardList, ok := u.fishData.GodBlessRewardMap[strRoomType]
	if !ok || len(rewardList) == 0 {
		code = int32(common_proto.ErrorCode_Succ)
		return
	}

	// 更新奖励状态
	for i := 0; i < len(rewardList)-1; i++ {
		reward := rewardList[i]
		nextReward := rewardList[i+1]
		if reward.RewardStatus == common_proto.RewardStatus_Got &&
			nextReward.RewardStatus == common_proto.RewardStatus_NotAvailable {
			log.Inf("第[%d]阶段的奖励已经领过了", i)
			t := time.Now().Unix() - reward.GotTime
			if t >= 60*60*24 {
				log.Inf("第[%d]阶段的奖励能领了", i+1)
				// 下一个奖励能领了
				nextReward.RewardStatus = common_proto.RewardStatus_Available
				nextReward.LeftGetTime = 0
				break
			} else {
				// 更新时间
				nextReward.LeftGetTime = reward.GotTime + 60*60*24 - time.Now().Unix()
				log.Inf("第[%d]阶段的奖励还有[%d]秒就能领了", i+1, nextReward.LeftGetTime)
			}
		}
	}

	ret.IsReward = true
	ret.RewardStatus = rewardList

	code = int32(common_proto.ErrorCode_Succ)
	return
}

func fishGetGodBlessReward(mid string, req *fish_proto.ReqFishGetGodBlessReward, ret *fish_proto.RetFishGetGodBlessReward) (code int32) {
	code = int32(common_proto.ErrorCode_Fail)

	log.Inf("玩家[%s]获取女神赐福奖励，roomType[%d]", mid, req.RoomType)

	u, b := _userM.getUser(mid)
	if !b {
		log.Wrn("失败：没有找到userinfo")
		return
	}

	strRoomType := strconv.Itoa(int(req.RoomType))
	rewardList, ok := u.fishData.GodBlessRewardMap[strRoomType]
	if !ok || len(rewardList) == 0 {
		log.Inf("玩家没有触发女神赐福")
		return
	}

	index := int(req.RewardType)
	if rewardList[index].RewardStatus != common_proto.RewardStatus_Available {
		// 是否有能领奖的
		index = -1
		for i := 0; i < len(rewardList); i++ {
			if rewardList[i].RewardStatus == common_proto.RewardStatus_Available {
				index = i
				break
			}
		}
	}

	if index == -1 {
		log.Wrn("没有找到奖励")
		return
	}

	reward := rewardList[index]

	// 发奖
	if !u.userInfo.AddCoin(int64(reward.Coin)) {
		log.Wrn("发奖失败")
		return
	}

	// 如果是最后一个奖，直接重置
	if index == len(rewardList)-1 {
		rewardList = []*fish_proto.GodBlessReward{}
	} else {
		reward.GotTime = time.Now().Unix()
		reward.RewardStatus = common_proto.RewardStatus_Got
	}

	code = int32(common_proto.ErrorCode_Succ)
	return
}

func fishComboRewardInfo(mid string, req *fish_proto.ReqFishComboRewardInfo, ret *fish_proto.RetFishComboRewardInfo) (code int32) {
	code = int32(common_proto.ErrorCode_Fail)

	u, b := _userM.getUser(mid)
	if !b {
		log.Wrn("失败：没有找到userinfo")
		return
	}

	ret.Coin = u.fishData.ComboReward
	code = int32(common_proto.ErrorCode_Succ)
	return
}

func fishGetComboReward(mid string, req *fish_proto.ReqFishGetComboReward, ret *fish_proto.RetFishGetComboReward) (code int32) {
	code = int32(common_proto.ErrorCode_Fail)

	u, b := _userM.getUser(mid)
	if !b {
		log.Wrn("失败：没有找到userinfo")
		return
	}

	ret.Coin = u.fishData.ComboReward
	// 发奖
	ok := u.userInfo.AddCoin(u.fishData.ComboReward)
	if !ok {
		log.Wrn("玩家[%s]领取连击奖励失败：发奖失败")
		return
	}

	u.resetComboReward()
	code = int32(common_proto.ErrorCode_Succ)
	return
}
