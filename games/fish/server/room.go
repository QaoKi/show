package fish

import (
	"bygame/common/conf"
	"bygame/common/conf/fish/cfg"
	"bygame/common/def"
	"bygame/common/log"
	"bygame/common/proto/fish_proto"
	"bygame/common/utils"
	"bygame/gate"

	"context"
	"sync"
	"time"

	"google.golang.org/protobuf/reflect/protoreflect"
	"gopkg.in/mgo.v2/bson"
)

type room struct {
	roomType int32  // 0:初级场，1:中级场，2:高级场
	roomId   string // 房间自己的编号

	seatIdUserMap map[int32]string // 座位号对应的user
	userCount     int32            // 房间中玩家数量
	userMu        sync.RWMutex     // 锁

	disConnUserMap map[string]bool // 离线的玩家

	fishMap map[string]*fish
	fishMu  sync.RWMutex // 锁

	genFishMap          map[int32][]*fish // 出鱼时间表
	fishFormatList      []*fishFormat     // 鱼阵
	currFishFormatIndex int32             // 鱼阵出到第几个了

	// 房间停止
	ctxStop  context.Context
	stopFunc context.CancelFunc

	freezeTime       int32 // 剩余冰冻时间
	latestFreezeTime int64 // 最新一次冰冻产生时的时间

	currRoundTime int32 // 当前回合时间
}

func newRoom(roomType int32) *room {
	var r room
	r.roomType = roomType
	r.roomId = bson.NewObjectId().Hex()

	r.seatIdUserMap = make(map[int32]string)
	r.fishMap = make(map[string]*fish)
	r.disConnUserMap = make(map[string]bool)

	r.ctxStop, r.stopFunc = context.WithCancel(context.Background())

	r.initRoom()
	return &r
}

func (r *room) resetRoom() {
	r.userMu.Lock()
	defer r.userMu.Unlock()
	r.seatIdUserMap = make(map[int32]string)
	r.userCount = 0
}

func (r *room) initRoom() {

	// 把鱼和鱼阵都生成出来
	r.genFinsh()
	r.genFishFormat()
}

func (r *room) start() {
	// 出鱼
	//go r.genFishTimerStart()
	go r.testFish()
}

func (r *room) testFish() {
	i := 0
	timer := time.NewTimer(1 * time.Second)
	defer timer.Stop()
	for {
		select {
		case <-timer.C:
			slc := []*fish{}
			for i := 0; i < 3; i++ {
				fishInfo := conf.Cf.FishTable.TableFishInfo.Get(520)
				paths := getRandBezier()
				f := newFish(fishInfo, fishInfo.Speed, paths)
				slc = append(slc, f)
			}

			r.addFish(slc)

			c := len(slc)
			t := slc[c-1].pathList[0].desTime - time.Now().Unix()

			log.Inf("===testlog  生成10鱼 i[%d], t[%d]", i, t)
			i++
			timer.Reset(time.Duration(t) * time.Second)
		}
	}
}

func (r *room) destroyRoom() {
	r.stopFunc()
}

func (r *room) getRoomInfo() *fish_proto.RoomInfo {
	roomInfo := &fish_proto.RoomInfo{}
	roomInfo.LatestFreezeStartTime = r.latestFreezeTime
	roomInfo.LatestFreezeEndTime = r.latestFreezeTime + int64(_bingDongTime)

	roomInfo.UserInfos = r.getAllUserInfo()
	fishlist := r.getAllAliveFish()
	log.Inf("====testlog 一共有[%d]条活着的鱼", len(fishlist))

	p := fishConvertProto(fishlist)
	roomInfo.Fishs = append(roomInfo.Fishs, p...)
	return roomInfo
}

// 玩家加入房间
func (r *room) joinRoom(u *user) (bool, int32) {

	r.userMu.Lock()
	// 获取一个空座位
	seatId := int32(-1)
	for i := int32(0); i < def.FishRoomMaxUserCount; i++ {
		if i == 1 {
			continue
		}
		if r.seatIdUserMap[i] == "" {
			seatId = i
			break
		}
	}

	if seatId == -1 {
		return false, -1
	}

	r.seatIdUserMap[seatId] = u.userInfo.Mid
	r.userCount++

	r.userMu.Unlock()

	// 默认给个房间最小倍率的炮值
	roomConfig := conf.Cf.FishTable.TableRoomConfig.Get(r.roomType)
	if roomConfig != nil {
		u.playData.gunValue = roomConfig.MinRate
	}

	log.Inf("user[%s] 加入房间 room id[%s] 座位号[%d] 炮值[%d]", u.userInfo.Mid, r.roomId, seatId, u.playData.gunValue)

	// 广播
	fishInfo := &fish_proto.FishUserInfo{
		UserInfo: u.userInfo.UserInfo2ProtoUserInfo(),
		SeatId:   seatId,
		GunValue: u.playData.gunValue,
	}

	event := &fish_proto.EventFishUserJoinRoom{
		UserInfo: fishInfo,
	}
	r.send2all(event)

	if r.getUserCount() == 1 {
		// 进人了，直接出鱼
		r.start()
	}

	return true, seatId
}

// 重连
func (r *room) reJoinRoom(u *user) (bool, *fish_proto.RoomInfo) {
	log.Inf("玩家[%s]重新进入房间", u.userInfo.Mid)
	exist := false
	for _, mid := range r.seatIdUserMap {
		if mid == u.userInfo.Mid {
			exist = true
			break
		}
	}

	if !exist {
		log.Wrn("失败：房间中没有找到这个玩家")
		return false, nil
	}

	roomInfo := r.getRoomInfo()
	return true, roomInfo
}

// 退出房间
func (r *room) quitRoom(u *user, reason fish_proto.UserQuitReason) {
	log.Inf("user[%s] 退出房间[%s], reason[%d]", u.userInfo.Mid, r.roomId, reason)
	r.userMu.Lock()

	if u.userInfo.Mid == r.seatIdUserMap[u.playData.seat] {
		r.seatIdUserMap[u.playData.seat] = ""
		r.userCount--
	} else {
		key := int32(-1)
		for k, v := range r.seatIdUserMap {
			if u.userInfo.Mid == v {
				key = k
			}
		}
		if key != -1 {
			r.seatIdUserMap[key] = ""
			r.userCount--
		}
	}

	r.userMu.Unlock()

	// 广播
	event := &fish_proto.EventFishUserQuitRoom{
		Mid:    u.userInfo.Mid,
		Reason: reason,
	}

	r.send2all(event)
}

func (r *room) genFishTimerStart() {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-r.ctxStop.Done():
			log.Inf("停止生成鱼")
			return
		case <-ticker.C:
			// 如果是冰冻时间则不生成鱼
			if r.freezeTime > 0 {
				r.freezeTime -= 1
				continue
			}
			r.currRoundTime += 1
			slc := r.genFishMap[r.currRoundTime]
			for _, v := range slc {
				v.updateTime()
			}
			r.addFish(slc)

			// 一轮鱼出完了，出鱼阵
			if r.currRoundTime == int32(len(r.genFishMap)) {
				r.genFishFormatStart()
				// 重新随机出鱼时间
				r.genFinsh()
			}
		}
	}
}

func (r *room) genFishFormatStart() {
	log.Inf("第[%d]鱼阵开始出鱼", r.currFishFormatIndex)
	fishFormat := r.fishFormatList[r.currFishFormatIndex]
	if len(fishFormat.LotFishList) == 0 {
		log.Inf("失败：鱼阵中没有鱼")
		return
	}

	if fishFormat.LotInterval > 0 {
		lotIndex := 0
		ticker := time.NewTicker(time.Duration(fishFormat.LotInterval) * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-r.ctxStop.Done():
				log.Inf("鱼阵停止出鱼")
				return
			case <-ticker.C:
				fishList := fishFormat.LotFishList[lotIndex]
				for _, v := range fishList {
					v.updateTime()
				}
				r.addFish(fishList)

				lotIndex++
				if lotIndex == len(fishFormat.LotFishList) {
					// 调整下一个要出的鱼阵
					formatCount := int32(len(r.fishFormatList))
					r.currFishFormatIndex = (r.currFishFormatIndex + 1) % formatCount
					return
				}
			}
		}
	} else {
		fishList := []*fish{}
		for _, f := range fishFormat.LotFishList {
			fishList = append(fishList, f...)
		}

		for _, v := range fishList {
			v.updateTime()
		}
		r.addFish(fishList)

		// 调整下一个要出的鱼阵
		formatCount := int32(len(r.fishFormatList))
		r.currFishFormatIndex = (r.currFishFormatIndex + 1) % formatCount
	}
}

/*
根据roundId生成鱼
三种类型	1 小组鱼 一次出三条 排队出

	2 普通鱼 在范围内随机出
	3 特殊鱼 范围第一秒出
*/
func (r *room) genFinsh() {
	r.genFishMap = make(map[int32][]*fish)
	// 先构造每秒生成的鱼
	for roundId, data := range conf.Cf.FishTable.TableGenFish.GetDataMap() {
		log.Dbg("roundId: %v,data: %v", roundId, data)
		for _, v := range data.Item {
			timeStart := (roundId-1)*30 + 1
			timeEnd := roundId * 30
			fishInfo := conf.Cf.FishTable.TableFishInfo.Get(v.FishId)
			if fishInfo == nil {
				log.Wrn("出鱼错误：未找到鱼配置,fishId: %v", v.FishId)
				continue
			}
			// 三种鱼
			switch v.Type {
			case cfg.FishType_XiaoZuYu:
				// 按小组出
				for i := 0; i < int(v.Num)/3; i++ {
					tk := utils.RandInt(int(timeStart), int(timeEnd))
					paths := getRandBezier()
					for j := 0; j < 3; j++ {
						f := newFish(fishInfo, fishInfo.Speed, paths)
						// 这里一组鱼设置的出生时间一致,通过出生时间晚1s来达到出一队的效果
						// 1s 的时间可能不够有的鱼太大了，后面考虑乘鱼的模型大小参数
						r.addGenFish(int32(tk+j), f)
					}
				}
			case cfg.FishType_PuTongYu:
				for i := 0; i < int(v.Num); i++ {
					paths := getRandBezier()
					tk := utils.RandInt(int(timeStart), int(timeEnd))
					f := newFish(fishInfo, fishInfo.Speed, paths)
					// 出生时间等真正出生的时候在设置
					r.addGenFish(int32(tk), f)
				}
			case cfg.FishType_TeShuYu:
				// 区间第一秒出
				for i := 0; i < int(v.Num); i++ {
					paths := getRandBezier()
					f := newFish(fishInfo, fishInfo.Speed, paths)
					// 出生时间等真正出生的时候在设置
					r.addGenFish(int32(timeStart), f)
				}
			}
		}
	}
}

// 构造鱼阵中的鱼
func (r *room) genFishFormat() {

	// 第一鱼阵
	if format := r.genSpecialFormat1(); format != nil {
		r.fishFormatList = append(r.fishFormatList, format)
	}

	// 第二鱼阵
	if format := r.genSpecialFormat2(); format != nil {
		r.fishFormatList = append(r.fishFormatList, format)
	}

	// 第三鱼阵
	if format := r.genSpecialFormat3(); format != nil {
		r.fishFormatList = append(r.fishFormatList, format)
	}

	// 第四鱼阵
	if format := r.genSpecialFormat4(); format != nil {
		r.fishFormatList = append(r.fishFormatList, format)
	}
}

func (r *room) genSpecialFormat1() *fishFormat {
	format := &fishFormat{}
	fishInfoMap := conf.Cf.FishTable.TableFishInfo.GetDataMap()

	// 第一鱼阵
	specialFormat := conf.Cf.FishTable.TableSpecialFormat1.GetDataList()
	if len(specialFormat) == 0 {
		log.Dbg("第一鱼阵无数据")
		return nil
	}
	pathNum := specialFormat[0].FishNum
	if pathNum <= 0 {
		log.Dbg("第一鱼阵 fish num is 0")
		return nil
	}
	// 路线
	paths := genSpecialFormat1Bezier(pathNum)

	// 创建每个轮次的鱼
	for _, data := range specialFormat {
		fishList := []*fish{}
		// 轮次的间隔
		format.LotInterval = data.LotInterval
		// 创建每条鱼
		for i := 0; i < int(data.FishNum); i++ {
			if info, b := fishInfoMap[data.FishId]; b {
				p := [][4]vec2{paths[i]}
				fish := newFish(info, data.Speed, p)
				fishList = append(fishList, fish)
			}
		}

		format.LotFishList = append(format.LotFishList, fishList)
	}
	return format
}

func (r *room) genSpecialFormat2() *fishFormat {
	format := &fishFormat{}
	fishInfoMap := conf.Cf.FishTable.TableFishInfo.GetDataMap()

	// 第二鱼阵
	specialFormat := conf.Cf.FishTable.TableSpecialFormat2.GetDataList()
	if len(specialFormat) == 0 {
		log.Dbg("第二鱼阵无数据")
		return nil
	}

	for _, data := range specialFormat {
		fishList := []*fish{}
		// 创建第一路线上的鱼
		if info, b := fishInfoMap[data.FirstPathFishId]; b {
			p1 := genSpecialFormat2Bezier(data.Lot, 1)
			p := [][4]vec2{p1}
			fish := newFish(info, data.Speed, p)
			fishList = append(fishList, fish)
		}

		// 创建第二路线上的鱼
		if info, b := fishInfoMap[data.SecondPathFishId]; b {
			p1 := genSpecialFormat2Bezier(data.Lot, 2)
			p := [][4]vec2{p1}
			fish := newFish(info, data.Speed, p)
			fishList = append(fishList, fish)
		}

		// 创建第三路线上的鱼
		if info, b := fishInfoMap[data.ThirdPathFishId]; b {
			p1 := genSpecialFormat2Bezier(data.Lot, 3)
			p := [][4]vec2{p1}
			fish := newFish(info, data.Speed, p)
			fishList = append(fishList, fish)
		}

		format.LotFishList = append(format.LotFishList, fishList)
	}

	return format
}

func (r *room) genSpecialFormat3() *fishFormat {
	format := &fishFormat{}
	fishInfoMap := conf.Cf.FishTable.TableFishInfo.GetDataMap()

	// 第三鱼阵
	specialFormat := conf.Cf.FishTable.TableSpecialFormat3.GetDataList()
	if len(specialFormat) == 0 {
		log.Dbg("第三鱼阵无数据")
		return nil
	}

	fishList := []*fish{}
	// 屏幕的左边和右边
	for _, data := range specialFormat {
		bLeft := true
		if data.ScreenDirection == cfg.ScreenDirection_ScreenRight {
			bLeft = false
		}

		outerPaths, innerPaths, midPath := genSpecialFormat3Bezier(data.OuterFishNum, data.InnerFishNum, bLeft)
		// 外圈
		for i := 0; i < len(outerPaths); i++ {
			if info, b := fishInfoMap[data.OuterFishId]; b {
				p := [][4]vec2{outerPaths[i]}
				fish := newFish(info, data.Speed, p)
				fishList = append(fishList, fish)
			}
		}

		// 内圈
		for i := 0; i < len(innerPaths); i++ {
			if info, b := fishInfoMap[data.InnerFishId]; b {
				p := [][4]vec2{innerPaths[i]}
				fish := newFish(info, data.Speed, p)
				fishList = append(fishList, fish)
			}
		}

		// 中间
		if info, b := fishInfoMap[data.MidFishId]; b {
			p := [][4]vec2{midPath}
			fish := newFish(info, data.Speed, p)
			fishList = append(fishList, fish)
		}
	}
	format.LotFishList = append(format.LotFishList, fishList)

	return format
}

func (r *room) genSpecialFormat4() *fishFormat {
	// 第四鱼阵
	format := &fishFormat{}
	fishInfoMap := conf.Cf.FishTable.TableFishInfo.GetDataMap()

	specialFormat := conf.Cf.FishTable.TableSpecialFormat4.GetDataList()
	if len(specialFormat) == 0 {
		log.Dbg("第四鱼阵无数据")
		return nil
	}

	fishList := []*fish{}
	// 屏幕的左边和右边
	for _, data := range specialFormat {
		bLeft := true
		if data.ScreenDirection == cfg.ScreenDirection_ScreenRight {
			bLeft = false
		}

		for _, tmp := range data.FishList {
			if info, b := fishInfoMap[tmp.FishId]; b {
				for i := 0; i < int(tmp.Num); i++ {
					p1 := genSpecialFormat4Bezier(int32(i), float64(tmp.YLine), bLeft)
					p := [][4]vec2{p1}
					fish := newFish(info, data.Speed, p)
					fishList = append(fishList, fish)
				}
			}
		}
	}
	format.LotFishList = append(format.LotFishList, fishList)

	return format

}

func (r *room) addGenFish(time int32, f *fish) {
	slc := r.genFishMap[time]
	r.genFishMap[time] = append(slc, f)
}

func (r *room) addFish(slc []*fish) {

	if len(slc) == 0 {
		return
	}
	// 鱼放到房间数据中
	r.fishMu.Lock()
	for _, f := range slc {
		log.Dbg("鱼出生 %v", f.fc.Name)
		r.fishMap[f.id] = f
	}
	r.fishMu.Unlock()

	// 广播
	fishList := fishConvertProto(slc)

	event := &fish_proto.EventFishGenFish{}
	event.FishList = fishList

	r.send2all(event)
}

func (r *room) getFish(fishId string) (*fish, bool) {
	r.fishMu.RLock()
	defer r.fishMu.RUnlock()
	f, ok := r.fishMap[fishId]
	return f, ok
}

// 获取当前房间人数
func (r *room) getUserCount() int32 {
	r.userMu.RLock()
	defer r.userMu.RUnlock()
	return r.userCount
}

// 获取当前房间玩家的信息
func (r *room) getAllUserInfo() []*fish_proto.FishUserInfo {
	r.userMu.RLock()
	defer r.userMu.RUnlock()
	userInfos := []*fish_proto.FishUserInfo{}

	for sid, mid := range r.seatIdUserMap {
		if mid == "" {
			continue
		}
		info := &fish_proto.FishUserInfo{}
		if u, b := _userM.getUser(mid); b {
			info.UserInfo = u.userInfo.UserInfo2ProtoUserInfo()
			info.SeatId = sid
			info.GunValue = u.playData.gunValue
			info.IsDisConn = u.getDisConnStatus()
			info.UserStatus = u.getUserStatus()
		}
		userInfos = append(userInfos, info)
	}

	return userInfos
}

func (r *room) getAllAliveFish() []*fish {
	r.fishMu.RLock()
	defer r.fishMu.RUnlock()
	fishlist := []*fish{}

	for _, f := range r.fishMap {
		if f.isLive() {
			fishlist = append(fishlist, f)
		}
	}
	return fishlist
}

func (r *room) getAliveFish(fishId int32) []*fish {
	r.fishMu.RLock()
	defer r.fishMu.RUnlock()
	fishlist := []*fish{}

	for _, f := range r.fishMap {
		if f.isLive() && f.cid >= fishId {
			fishlist = append(fishlist, f)
		}
	}
	return fishlist
}

func (r *room) send2all(event protoreflect.ProtoMessage) {
	r.userMu.RLock()

	for _, mid := range r.seatIdUserMap {
		if u, ok := _userM.getUser(mid); ok && !u.playData.isRobot {
			gate.SendEventToUser(u.userInfo.Mid, event)
		}
	}

	r.userMu.RUnlock()
}

// 广播桌子，除了 seatId
func (r *room) send2AllExcept(event protoreflect.ProtoMessage, seatId int32) {
	r.userMu.RLock()

	for s, mid := range r.seatIdUserMap {
		if s == seatId {
			continue
		}
		if u, ok := _userM.getUser(mid); ok {
			gate.SendEventToUser(u.userInfo.Mid, event)
		}
	}

	r.userMu.RUnlock()
}
