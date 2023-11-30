package fish

import (
	"bygame/common/conf"
	"bygame/common/def"
	"bygame/common/log"
	"bygame/common/rdb"
	"bygame/common/utils"
	"bygame/gate"
	"time"
)

var _gameM *gamem

type gamem struct {
}

func initGameM() {
	_gameM = &gamem{}
	_gameM.initGodBlessData()
}

func (gm *gamem) initGodBlessData() {

	// 查看redis中是否有数据

	config := conf.Cf.FishTable.XiaoYouXiGoddessBless.GetDataList()
	if len(config) == 0 {
		log.Wrn("女神赐福初始化失败：没有找到配置")
		return
	}
	ok := false
	t := time.Now().Unix()
	for _, cfg := range config {
		ok, _, _, _ = rdb.GetFishGodBlessPoolInfo(cfg.RoomType)
		if !ok {
			rdb.SetFishGodBlessPoolInfo(cfg.RoomType, int64(cfg.InitPool), t, t)
		}
	}

	go gm.godBlessPoolChangeTimer()
}

func (gm *gamem) godBlessPoolChangeTimer() {

	config := conf.Cf.FishTable.XiaoYouXiGoddessBless.GetDataList()

	addTicker := time.NewTicker(1 * time.Second)
	subRandTime := utils.RandInt32(config[0].SubPoolTime.One, config[0].SubPoolTime.Two)
	//log.Inf("=====testlog, sub rand T[%d]\n", subRandTime)
	subTimer := time.NewTimer(time.Duration(subRandTime) * time.Second)
	for {
		select {
		case <-addTicker.C:
			tNow := time.Now().Unix()
			for _, cfg := range config {
				// 先取，再 set
				ok, count, ut, subT := rdb.GetFishGodBlessPoolInfo(cfg.RoomType)
				//fmt.Printf("=====testlog add 获取 roomtype[%d], ok[%v], c[%d], addT[%d], subT[%d], now-t[%d]\n", cfg.RoomType, ok, count, ut, subT, tNow-ut)
				if !ok {
					continue
				}
				addRand := utils.RandInt64(int64(cfg.AddPool.One), int64(cfg.AddPool.Two))
				if tNow-ut >= 1 {
					rdb.SetFishGodBlessPoolInfo(cfg.RoomType, count+addRand, tNow, subT)
				}
			}
		case <-subTimer.C:
			tNow := time.Now().Unix()
			for _, cfg := range config {
				// 先取，再 set
				ok, count, addT, ut := rdb.GetFishGodBlessPoolInfo(cfg.RoomType)
				if !ok {
					continue
				}
				//log.Inf("=====testlog sub 获取 roomtype[%d], ok[%v], c[%d], addT[%d], subT[%d], now-t[%d]\n", cfg.RoomType, ok, count, addT, ut, tNow-ut)

				subRand := utils.RandInt64(int64(cfg.SubPoolCount.One), int64(cfg.SubPoolCount.Two))
				if tNow-ut >= int64(config[0].SubPoolTime.One) {
					rdb.SetFishGodBlessPoolInfo(cfg.RoomType, count-subRand, addT, tNow)
				}
			}

			subRandTime := utils.RandInt32(config[0].SubPoolTime.One, config[0].SubPoolTime.Two)
			subTimer.Reset(time.Duration(subRandTime) * time.Second)
		}
	}
}

// 整个模块的初始化
func Init() {
	//initGameM()
	initRoomM()
	initUserM()
	addListener()
	gate.SetOfflineFunc(def.GameIdNewFishingEra, disConnect)
	//gate.SetReConnFunc(def.GameIdNewFishingEra, reConnect)
	gate.SetKickOutFunc(def.GameIdNewFishingEra, kickOut)

	//gate.SetAutoCloseTime(def.GameIdNewFishingEra, 30)

}
