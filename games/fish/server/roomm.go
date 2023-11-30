package fish

import (
	"bygame/common/conf"
	"bygame/common/def"
	"bygame/common/log"
	"sync"

	"google.golang.org/protobuf/proto"
)

//var rmm *roomm

//func initRm() {
// rmm = &roomm{Rs: make(map[string]*room)}
// mat = &match{matchGroup: make(map[string][]*user)}
// go func() {
// 	for {
// 		mathcRobot()
// 		time.Sleep(time.Second * 4)
// 	}
// }()
//}

// 使用playRoomInfo 生成key
// func pri2key(info *cmd.PlayRoomInfo) string {
// 	key := fmt.Sprintf("%v-%v-%v-%v", info.GameId, info.Ante, info.MinCoin, info.MaxCoin)
// 	return key
// }

func mathcRobot() {
	// keys := make([]string, 0, len(rmm.Rs))
	// rmm.mu.RLock()
	// for k := range rmm.Rs {
	// 	keys = append(keys, k)
	// }
	// rmm.mu.RUnlock()

	// for _, rid := range keys {
	// 	// if r, ok := rmm.GetRoom(rid); ok && !r.firendRoom && len(r.players)+len(r.prePlayers) < 4 {
	// 	// 	// 根据真人玩家数和空位置
	// 	// 	// idleNum := 4 - len(r.players) - len(r.prePlayers)
	// 	// 	// if c, ok2 := conf.Cf.DominoConf.MatchRobotMap[idleNum]; ok2 && utils.Rate(c.RobotRate) {
	// 	// 	// 	robot := NewRobot(r.roomInfo)
	// 	// 	// 	log.Inf("房间自动进入一个机器人 %v", r.roomId)
	// 	// 	// 	r.JoinRoom(robot)
	// 	// 	// }
	// 	// }
	// }
}

// var mat *match

// type match struct {
// 	mu         sync.RWMutex
// 	matchGroup map[string][]*user
// }

var _roomM *roomm

type roomm struct {
	allRoomMapList []map[int32][]*room // 每种类型的房间都有一个map
	muList         []sync.RWMutex      // 每种类型的房间有自己的锁
}

func initRoomM() {
	_roomM = &roomm{}

	// 初始化桌子
	roomConfig := conf.Cf.FishTable.TableRoomConfig.GetDataList()
	if len(roomConfig) == 0 {
		log.Err("初始化 room manage 失败: 没有找到房间配置")
		return
	}

	for _, cfg := range roomConfig {
		roomMap := make(map[int32][]*room)
		roomMap[0] = make([]*room, cfg.InitTableCount)
		for i := 0; i < int(cfg.InitTableCount); i++ {
			r := newRoom(cfg.RoomType)
			roomMap[0][i] = r
		}
		_roomM.allRoomMapList = append(_roomM.allRoomMapList, roomMap)
		_roomM.muList = append(_roomM.muList, sync.RWMutex{})
	}

}

func (rm *roomm) getRoom(roomType int32) (*room, bool) {
	log.Inf("通过 roomType[%d] 获取房间", roomType)

	if int(roomType) >= len(rm.allRoomMapList) || int(roomType) >= len(rm.muList) {
		log.Wrn("获取房间失败: room type 和配置不对应")
		return nil, false
	}

	rm.muList[roomType].Lock()

	roomMap := rm.allRoomMapList[roomType]
	// 从只差一个人的房间开始找
	var r *room
	for i := def.FishRoomMaxUserCount - 1; i >= 0; i-- {
		for {
			if len(roomMap[i]) == 0 {
				break
			}

			r = roomMap[i][0]
			// 从队列中删除
			roomMap[i] = roomMap[i][1:]
			if r != nil {
				// 当前房间人数
				currUser := r.getUserCount()
				// 人满了
				if currUser == def.FishRoomMaxUserCount {
					roomMap[def.FishRoomMaxUserCount] = append(roomMap[def.FishRoomMaxUserCount], r)
					// 重新找房间
					r = nil
					continue
				}

				roomMap[currUser+1] = append(roomMap[currUser+1], r)
				break
			}
		}

		if r != nil {
			break
		}
	}

	if r == nil {
		// 没找到
		r = newRoom(roomType)
		roomMap[1] = append(roomMap[1], r)
	}

	log.Inf("获取到一个房间, room id[%s], 当前房间中玩家数量[%d]", r.roomId, r.getUserCount())

	rm.muList[roomType].Unlock()
	return r, true
}

// 有一个玩家退出，调整房间队列
func (rm *roomm) returnRoom(r *room) {
	// 进入这个函数的时候，房间人数 已经被减过 1 了
	if r == nil {
		return
	}

	roomConfig := conf.Cf.FishTable.TableRoomConfig.Get(r.roomType)
	if roomConfig == nil {
		log.Wrn("调整房间队列失败：没有找到房间配置")
		return
	}

	rm.muList[r.roomType].Lock()

	oldCount := r.getUserCount() + 1
	log.Inf("调整房间队列, room type[%d], room id[%s] 玩家没退出之前，房间中玩家数量[%d]", r.roomType, r.roomId, oldCount)

	if oldCount > def.FishRoomMaxUserCount || oldCount < 1 {
		log.Wrn("失败：没退出之前，房间中玩家数量不对, 最大数量[%d]", def.FishRoomMaxUserCount)
		return
	}

	find := false
	roomMap := rm.allRoomMapList[r.roomType]
	for i := 0; i < len(roomMap[oldCount]); i++ {
		if roomMap[oldCount][i] == r {
			// 从队列中删除
			roomMap[oldCount] = append(roomMap[oldCount][:i], roomMap[oldCount][i+1:]...)
			find = true
			break
		}
	}

	if !find {
		// 去别的队列中找
		for i := int32(0); i <= def.FishRoomMaxUserCount; i++ {
			if i == oldCount {
				continue
			}

			for j := 0; j < len(roomMap[i]); j++ {
				if roomMap[i][j] == r {
					// 从队列中删除
					roomMap[i] = append(roomMap[i][:j], roomMap[i][j+1:]...)
					find = true
					break
				}
			}

			if find {
				break
			}
		}
	}

	roomMap[oldCount-1] = append(roomMap[oldCount-1], r)
	if oldCount == 1 {
		r.resetRoom()
	}

	allRoomCount := int32(0)
	for _, m := range roomMap {
		allRoomCount += int32(len(m))
	}

	var desRoom *room = nil
	if allRoomCount >= 2*roomConfig.InitTableCount && int32(len(roomMap[0])) > roomConfig.InitTableCount/2 {
		// 销毁一个空房间
		desRoom = roomMap[0][0]
		roomMap[0] = roomMap[0][1:]
	}

	rm.muList[r.roomType].Unlock()

	if desRoom != nil {
		desRoom.destroyRoom()
		desRoom = nil
	}

}

func (rm *roomm) send2AllRoom(event proto.Message) {
	for i, roomMap := range rm.allRoomMapList {
		rm.muList[i].RLock()
		for k, v := range roomMap {
			if k == 0 {
				continue
			}

			for _, r := range v {
				r.send2all(event)
			}

		}
		rm.muList[i].RUnlock()
	}
}
