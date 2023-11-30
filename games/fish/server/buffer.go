package fish

import (
	"bygame/common/conf"
	"bygame/common/conf/fish/cfg"
	"bygame/common/log"
	"bygame/common/proto/fish_proto"
	"bygame/common/utils"

	"strconv"
	"time"
)

type bufferFunc func(*user, *fish, *room)

var bufferFuncMap map[int32]bufferFunc

func InitBufferFunc() {
	bufferFuncMap = make(map[int32]bufferFunc)
	registerBufferFunc(int32(cfg.FishBuffer_BingDong), bingDong)
	registerBufferFunc(int32(cfg.FishBuffer_RewardZhenZhu), onceZhenZhu)
	registerBufferFunc(int32(cfg.FishBuffer_ZhuanPan), onceZhuanPan)
	registerBufferFunc(int32(cfg.FishBuffer_MoGu), onceMogu)
	registerBufferFunc(int32(cfg.FishBuffer_Athena), athena)
	registerBufferFunc(int32(cfg.FishBuffer_Poseidon), poseidon)
	registerBufferFunc(int32(cfg.FishBuffer_Hardess), hardess)
	registerBufferFunc(int32(cfg.FishBuffer_MeiRenYuBox), meiRenYuBaoXiang)

	registerBufferFunc(int32(cfg.FishBuffer_TwiceRewardZhenZhu), twiceZhenZhu)
	registerBufferFunc(int32(cfg.FishBuffer_TwiceZhuanPan), twiceZhuanPan)
	registerBufferFunc(int32(cfg.FishBuffer_TwiceMoGu), twiceMogu)
}

func registerBufferFunc(bufferId int32, f bufferFunc) {
	bufferFuncMap[bufferId] = f
}

func afterKill(u *user, f *fish, r *room) {
	// 如果触发了特殊buffer
	if f.touchSpecialBuffer {
		if fun, b := bufferFuncMap[f.fc.SpecialBuffer]; b {
			fun(u, f, r)
		}
	} else {
		if fun, b := bufferFuncMap[f.fc.Buffer]; b {
			fun(u, f, r)
		}
	}

}

// ************************ 冰冻效果 **************************
func bingDong(user *user, fish *fish, r *room) {
	log.Inf("玩家[%s]触发冰冻", user.userInfo.Mid)

	r.freezeTime += _bingDongTime
	r.latestFreezeTime = time.Now().Unix()

	// 广播
	event := &fish_proto.EventFishBingDongBuffer{}
	event.StartTime = r.latestFreezeTime
	event.EndTime = r.latestFreezeTime + int64(_bingDongTime)
	r.send2all(event)
}

// ************************ 获得珍珠 **************************
func onceZhenZhu(u *user, f *fish, r *room) {
	rewardZhenZhu(u, f, r, 1)
}

func twiceZhenZhu(u *user, f *fish, r *room) {
	rewardZhenZhu(u, f, r, 2)
}

func rewardZhenZhu(u *user, f *fish, r *room, num int32) {
	log.Inf("玩家[%s]获得珍珠, room type[%d], num[%s]", u.userInfo.Mid, r.roomType, num)

	strRoomType := strconv.Itoa(int(r.roomType))
	if u.fishData.ZhenZhuMap == nil {
		u.fishData.ZhenZhuMap = make(map[string][]int32)
	}

	for i := 0; i < int(num); i++ {
		u.fishData.ZhenZhuMap[strRoomType] = append(u.fishData.ZhenZhuMap[strRoomType], f.touchBufferRate)
	}

	// 广播房间
	event := &fish_proto.EventFishAddZhenZhu{}
	event.Count = num
	event.Rate = f.touchBufferRate
	r.send2all(event)

	laBaConfig := conf.Cf.FishTable.XiaoYouXiLaBa.GetDataList()
	if len(laBaConfig) == 0 {
		log.Wrn("触发拉霸失败：没有找到拉霸配置")
		return
	}

	// 是否触发拉霸
	if len(u.fishData.ZhenZhuMap[strRoomType]) < int(_zhenZhuTouchCount) {
		return
	}

	// 触发
	succ := false
	succ, u.playData.labaData = touchLaBa(u, r.roomType)
	if !succ {
		return
	}

	// 保存状态
	u.addLaBaBuffer()

	u.setUserStatus(fish_proto.FishUserStatus_PlayLaBa)
	// 广播房间
	labaEvent := &fish_proto.EventFishAddLaBaBuffer{}
	labaEvent.Mid = u.userInfo.Mid
	labaEvent.Data = u.playData.labaData
	r.send2all(labaEvent)

	// 减去珍珠
	u.fishData.ZhenZhuMap[strRoomType] = u.fishData.ZhenZhuMap[strRoomType][_zhenZhuTouchCount:]

}

// ************************ 转盘 **************************
func onceZhuanPan(u *user, f *fish, r *room) {
	bufferZhuanPan(u, f, r, 1)
}

func twiceZhuanPan(u *user, f *fish, r *room) {
	bufferZhuanPan(u, f, r, 2)
}

func bufferZhuanPan(u *user, f *fish, r *room, num int32) {
	log.Inf("玩家[%s] 击杀鱼后触发转盘, num[%d]", u.userInfo.Mid, num)

	succ := false
	succ, u.playData.bufferZhuanPanDataList = touchBufferZhuanPan(u, f.touchBufferRate, num)
	if !succ {
		return
	}

	// 保存状态
	u.addBufferZhuanPan(num)

	u.setUserStatus(fish_proto.FishUserStatus_PlayBufferZhuanPan)
	// 广播房间
	event := &fish_proto.EventFishAddBufferZhuanPan{}
	event.Mid = u.userInfo.Mid
	event.Data = u.playData.bufferZhuanPanDataList

	r.send2all(event)
}

// ************************ 提莫的蘑菇庄园 **************************
func onceMogu(u *user, f *fish, r *room) {
	mogu(u, f, r, 1)
}

func twiceMogu(u *user, f *fish, r *room) {
	mogu(u, f, r, 2)
}

func mogu(u *user, f *fish, r *room, num int32) {
	log.Inf("玩家[%s]触发蘑菇庄园, num[%d]", u.userInfo.Mid, num)

	succ := false
	succ, u.playData.moguDataList = touchMoGu(u, f.touchBufferRate, num)
	if !succ {
		return
	}

	// 保存状态
	u.addMoGuBuffer(num)

	u.setUserStatus(fish_proto.FishUserStatus_PlayMoGu)
	// 广播房间
	event := &fish_proto.EventFishAddMoGuBuffer{}
	event.Mid = u.userInfo.Mid
	event.Data = u.playData.moguDataList
	r.send2all(event)
}

// ************************ 金蟾 **************************
func jinChan(u *user, f *fish, r *room) {
	log.Inf("玩家[%s]触发金蟾buffer", u.userInfo.Mid)

	cfgList := conf.Cf.FishTable.XiaoYouXiJinChan.GetDataList()
	if len(cfgList) == 0 {
		log.Wrn("失败：没有发现金蟾配置")
		return
	}

	slc := []utils.RandomItem{}
	for _, cfg := range cfgList {
		slc = append(slc, utils.RandomItem{Id: cfg.Id, Weight: cfg.Prob})
	}

	cfgId := utils.RandomWeight(slc)
	log.Inf("随机到的配置id[%d]", cfgId)

	cfg := conf.Cf.FishTable.XiaoYouXiJinChan.Get(cfgId)
	if cfg == nil {
		log.Wrn("失败：通过配置id没有找到配置")
		return
	}

	coin := int64(f.touchBufferRate * cfg.Rate)
	log.Inf("倍率[%d], 炮值[%d], 奖励的金币数[%d]", cfg.Rate, f.touchBufferRate, coin)

	// 发奖
	if succ := u.userInfo.AddCoin(coin); !succ {
		log.Wrn("失败：加钱失败")
		return
	}

	u.setUserStatus(fish_proto.FishUserStatus_PlayJinChan)
	// 广播房间
	event := &fish_proto.EventFishAddJinChanBuffer{}
	event.CfgId = cfgId
	event.Mid = u.userInfo.Mid
	event.Name = cfg.Name
	event.Rate = cfg.Rate
	event.Coin = coin
	r.send2all(event)
}

// ************************ 雅典娜 **************************
func athena(u *user, f *fish, r *room) {
	log.Inf("玩家[%s]触发雅典娜buffer", u.userInfo.Mid)

	cfgList := conf.Cf.FishTable.XiaoYouXiAthena.GetDataList()
	if len(cfgList) == 0 {
		log.Wrn("失败：没有找到雅典娜配置")
		return
	}
	u.setUserStatus(fish_proto.FishUserStatus_PlayAthena)

	// 先随机一个大区间
	index := int32(utils.RandInt(0, len(cfgList)-1))

	cfg := conf.Cf.FishTable.XiaoYouXiAthena.Get(index)

	// 再随机具体的倍率
	rate := utils.RandInt32(cfg.Rate.One, cfg.Rate.Two)

	// 总的奖励金币数
	sumCoin := int64(f.touchBufferRate) * int64(rate)

	// 随机特效次数
	sumCount := utils.RandInt32(cfg.Count.One, cfg.Count.Two)

	// 随机每次的特效的金币数
	effect := athenaEffect(sumCoin, sumCount)

	// 广播房间
	event := &fish_proto.EventFishAddAthenaBuffer{}
	event.Mid = u.userInfo.Mid
	event.SumCoin = sumCoin
	event.Effect = effect
	r.send2all(event)
}

func athenaEffect(sumCoin int64, sumCount int32) []int64 {
	ret := []int64{}

	aver := sumCoin / int64(sumCount)

	// todo，随机一下
	for i := 0; i < int(sumCount); i++ {
		ret = append(ret, aver)
	}

	return ret
}

// ************************ 波塞冬 **************************
func poseidon(u *user, f *fish, r *room) {
	u.setUserStatus(fish_proto.FishUserStatus_PlayPoseidon)
}

// ************************ 哈迪斯 **************************
func hardess(u *user, f *fish, r *room) {
	log.Inf("玩家[%s]触发哈迪斯buffer", u.userInfo.Mid)

	cfgList := conf.Cf.FishTable.XiaoYouXiHardess.GetDataList()
	if len(cfgList) == 0 {
		log.Wrn("失败：没有找到哈迪斯配置")
		return
	}

	u.setUserStatus(fish_proto.FishUserStatus_PlayHardess)

	// 这个鱼以下的鱼全部击杀
	cond := cfgList[0].FishId

	fishlist := r.getAllAliveFish()

	event := &fish_proto.EventFishAddHardessBuffer{}
	event.Mid = u.userInfo.Mid

	for _, fi := range fishlist {
		// 直接击杀
		if fi.cid < cond && fi.killFish() {
			changeCoin := fi.fc.Rate * int32(f.touchBufferRate)
			//加币成功，才能判定为成功击杀
			if u.userInfo.AddCoin(int64(changeCoin)) {
				event.KillList = append(event.KillList, &fish_proto.KillFishData{FishId: fi.id, AddCoin: changeCoin, Rate: f.fc.Rate})
			}
		}
	}

	log.Inf("杀掉鱼的数量[%d]", len(event.KillList))

	if len(event.KillList) > 0 {
		u.playData.room.send2all(event)
	}
}

// ************************ 美人鱼宝箱 **************************

func meiRenYuBaoXiang(u *user, f *fish, r *room) {
}
