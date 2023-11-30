package fish

import (
	"bygame/common/conf"
	"bygame/common/def"
	"bygame/common/log"
	"bygame/common/proto/common_proto"
	"bygame/common/proto/fish_proto"
	"bygame/gate"
)

func addListener() {
	gate.AddLintener(def.GameIdNewFishingEra, userJoinRoom)
	gate.AddLintener(def.GameIdNewFishingEra, userQuitRoom)
	gate.AddLintener(def.GameIdNewFishingEra, changeGun)
	gate.AddLintener(def.GameIdNewFishingEra, fishPlay)
	gate.AddLintener(def.GameIdNewFishingEra, fishHit)

	// 小游戏
	gate.AddLintener(def.GameIdNewFishingEra, fishPlayLaBa)
	gate.AddLintener(def.GameIdNewFishingEra, fishPlayBufferZhuanPan)
	gate.AddLintener(def.GameIdNewFishingEra, fishPlayMoGu)
	gate.AddLintener(def.GameIdNewFishingEra, fishGetActivityZhuanPanInfo)
	gate.AddLintener(def.GameIdNewFishingEra, fishPlayActivityZhuanPan)
	gate.AddLintener(def.GameIdNewFishingEra, fishGetActivityZhuanPanReward)
	gate.AddLintener(def.GameIdNewFishingEra, fishGetGodBlessInfo)
	gate.AddLintener(def.GameIdNewFishingEra, fishGetGodBlessReward)
	gate.AddLintener(def.GameIdNewFishingEra, fishComboRewardInfo)
	gate.AddLintener(def.GameIdNewFishingEra, fishGetComboReward)
}

func userJoinRoom(mid string, req *fish_proto.ReqUserJoinRoom, ret *fish_proto.RetUserJoinRoom) (code int32) {
	log.Inf("mid[%s] 加入房间 room type[%d]", mid, req.RoomType)
	code = int32(common_proto.ErrorCode_Fail)
	roomConfig := conf.Cf.FishTable.TableRoomConfig.Get(req.RoomType)
	if roomConfig == nil {
		log.Wrn("失败：没有找到房间配置")
		return
	}

	user, bool := _userM.getUser(mid)
	if !bool {
		log.Wrn("失败：没有找到 userinfo")
		return
	}

	// 如果已经在房间，重连
	if user.playData.room != nil && user.getDisConnStatus() {
		succ, roomInfo := user.playData.room.reJoinRoom(user)
		if !succ {
			// 把数据清了，再重新走加入房间
			user.playData.resetData()
			return
		}
		user.setDisConStatus(false)
		ret.ReConn = 1
		ret.RoomInfo = roomInfo
		code = int32(common_proto.ErrorCode_Succ)
		return
	}

	// 判断准入
	if user.userInfo.Coin < int64(roomConfig.MinCoin) {
		log.Wrn("失败：user 金币数量不够 coin[%d], 房间最小准入 coin[%d]", user.userInfo.Coin, roomConfig.MinCoin)
		return
	}

	// 取一个房间
	room, bool := _roomM.getRoom(req.RoomType)
	if !bool {
		return
	}

	succ, seatId := room.joinRoom(user)
	if !succ {
		log.Wrn("加入房间失败")
		return
	}

	user.playData.room = room
	user.playData.seat = seatId

	// 房间信息
	ret.RoomInfo = room.getRoomInfo()
	code = int32(common_proto.ErrorCode_Succ)
	return
}

func userQuitRoom(mid string, req *fish_proto.ReqUserQuitRoom, ret *fish_proto.RetUserQuitRoom) (code int32) {
	log.Inf("user[%s]主动退出房间 room type[%d]", mid, req.RoomType)
	code = int32(common_proto.ErrorCode_Fail)

	if quitGame(mid, fish_proto.UserQuitReason_ZhuDong) {
		code = int32(common_proto.ErrorCode_Succ)
	}
	return
}

func quitGame(mid string, reason fish_proto.UserQuitReason) bool {
	u, bool := _userM.getUser(mid)
	if !bool {
		log.Wrn("失败：没有找到 userinfo")
		return false
	}

	r := u.playData.room
	if r == nil {
		log.Wrn("失败：没有找到 user 所在房间")
		return false
	}

	r.quitRoom(u, reason)

	_roomM.returnRoom(r)
	u.playData.resetData()

	// 清掉网关
	gate.DestroyGateUser(u.userInfo.Mid)
	return true
}

// 切换炮台
func changeGun(mid string, req *fish_proto.ReqFishChangeGun, ret *fish_proto.RetFishChangeGun) (code int32) {
	code = int32(common_proto.ErrorCode_Fail)
	log.Inf("user[%s]切抢", mid)
	u, ok := _userM.getUser(mid)
	if !ok {
		log.Wrn("失败：没有找到玩家")
		return
	}

	if u.playData.room == nil {
		log.Wrn("失败：玩家房间为空")
		return
	}

	u.playData.gunValue = req.Bet

	// 广播
	var event fish_proto.EventFishChangeGun
	event.Bet = req.Bet
	event.Mid = mid
	u.playData.room.send2all(&event)

	code = int32(common_proto.ErrorCode_Succ)
	return
}

// 开炮 击中才扣钱
func fishPlay(mid string, req *fish_proto.ReqFishPlay, ret *fish_proto.RetFishPlay) (code int32) {
	code = int32(common_proto.ErrorCode_Fail)
	log.Inf("user[%s]开炮", mid)

	u, ok := _userM.getUser(mid)
	if !ok {
		log.Wrn("失败：没有找到玩家")
		return
	}

	if u.playData.room == nil {
		log.Wrn("失败：玩家房间为空")
		return
	}

	// todo 有状态不能开炮
	// s := u.getUserStatus()
	// if u.getUserStatus() != fish_proto.FishUserStatus_Normal {
	// 	log.Wrn("失败：玩家状态为[%d]", s)
	// 	return
	// }

	// 广播
	var event fish_proto.EventFishPlay
	event.Bet = req.Bet
	event.Mid = mid
	event.Target = req.Target
	event.FishId = req.FishId
	event.BulletId = req.BulletId
	event.BulletSkill = req.BulletSkill
	event.BulletGenTime = req.BulletGenTime
	u.playData.room.send2all(&event)

	code = int32(common_proto.ErrorCode_Succ)
	return
}

// 击中
func fishHit(mid string, req *fish_proto.ReqFishHit, ret *fish_proto.RetFishHit) (code int32) {
	code = int32(common_proto.ErrorCode_Fail)
	log.Inf("user[%s]击中鱼", mid)

	u, ok := _userM.getUser(mid)
	if !ok {
		log.Wrn("失败：没有找到玩家")
		return
	}

	if u.playData.room == nil {
		log.Wrn("失败：玩家房间为空")
		return
	}

	// todo 有状态不能击杀
	// s := u.getUserStatus()
	// if u.getUserStatus() != fish_proto.FishUserStatus_Normal {
	// 	log.Wrn("失败：玩家状态为[%d]", s)
	// 	return
	// }

	event := &fish_proto.EventFishKill{
		Mid: mid,
	}

	// 过滤同一个子弹打中多条鱼的情况
	hitInfoNew := []*fish_proto.FishHitInfo{}
	bulletMap := make(map[string]string)
	for _, hitInfo := range req.HitInfo {
		if bulletId, exist := bulletMap[hitInfo.BulletId]; exist {
			log.Inf("===== testlog 出现同一个子弹打中多条鱼的情况，子弹id[%s]，上一条鱼的id[%s]，当前鱼的id[%s]", bulletId, bulletMap[hitInfo.BulletId], hitInfo.FishId)
			continue
		}

		bulletMap[hitInfo.BulletId] = hitInfo.FishId
		hitInfoNew = append(hitInfoNew, hitInfo)
	}

	touchBufferFishList := []*fish{}
	// 击中和击杀扣钱 击杀
	for _, hitInfo := range hitInfoNew {
		// 先判断钱够不够攻击
		// todo
		// if u.userInfo.Coin < int64(req.Bet) {
		// 	continue
		// }

		f, ok := u.playData.room.getFish(hitInfo.FishId)

		if !ok {
			continue
		}

		// 如果鱼活着攻击成功就要扣钱
		if !f.isLive() {
			continue
		}

		// 尝试扣钱
		// todo
		// if !u.userInfo.AddCoin(-req.Bet) {
		// 	break
		// }

		// 判断被击杀
		isKilled := fishKilled(f)
		if !isKilled {
			continue
		}
		// 触发了buffer
		touchBufferFishList = append(touchBufferFishList, f)
		f.touchBufferRate = int32(hitInfo.Bet)

		// 这条鱼不是不死鱼，并且修改状态成功
		if !f.isImmortalFish() && f.killFish() {
			// todo
			//changeCoin := f.fc.Rate * int32(req.Bet)
			changeCoin := int32(10)
			//加币成功，才能判定为成功击杀
			// todo
			//if u.userInfo.AddCoin(int64(changeCoin)) {
			event.KillList = append(event.KillList, &fish_proto.KillFishData{FishId: hitInfo.FishId, AddCoin: changeCoin, Rate: f.fc.Rate})
			//}
		}

	}

	if len(event.KillList) > 0 {
		u.playData.room.send2all(event)
	}

	// todo
	// for _, f := range touchBufferFishList {
	// 	afterKill(u, f, u.playData.room)
	// }
	return
}

func disConnect(mid string) {
	log.Inf("用户[%s]断开连接", mid)
	u, bool := _userM.getUser(mid)
	if !bool {
		log.Wrn("失败：没有找到 user info by mid")
		return
	}

	r := u.playData.room
	if r == nil {
		log.Wrn("失败：没有找到 user 所在房间")
		return
	}

	u.setDisConStatus(true)
	// 广播
	event := &fish_proto.EventFishUserDisConnect{
		Mid: mid,
	}
	r.send2all(event)
}

func reConnect(mid string) {
	log.Inf("用户[%s] 重新连接", mid)
}

func kickOut(mid string) {
	log.Inf("用户[%s] 被踢出房间", mid)
	quitGame(mid, fish_proto.UserQuitReason_KickOut)
}
