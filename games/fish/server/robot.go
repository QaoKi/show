package fish

import (
	"bygame/common/conf"
	"bygame/common/conf/fish/cfg"
	"bygame/common/data"
	"bygame/common/log"
	"bygame/common/proto/fish_proto"
	"bygame/common/utils"
	"fmt"
	"sync/atomic"

	"context"
	"time"

	"gopkg.in/mgo.v2/bson"
)

type robot struct {
	u *user // 玩家数据

	gunValue int32 // 炮值

	roomType int32 // 所属的roomType
	gameTime int32 // 游戏时长

	currAction   int32 // 当前行为
	gunTarget    vec2  // 开炮角度
	gunSpeedMill int32 // 开炮速度，单位毫秒

	isSuoDing     atomic.Bool // 是否锁定攻击
	suoDingFishId string      // 被锁定鱼的实例id

	// 机器人停止
	ctxStop  context.Context
	stopFunc context.CancelFunc

	// 切换炮值定时器
	changeGunTimer *time.Timer
	// 退出定时器
	quitTimer *time.Timer
	// 切换行为定时器
	changeActionTimer *time.Timer
	// 切换锁定目标
	changeSuoDingTimer *time.Timer
	// 切换开炮角度
	changeJiaoDuTimer *time.Timer
	// 切换手动开炮速度
	changeSuDuTimer *time.Timer
	// 开炮定时器
	gunTimer *time.Timer
}

func newRobot(roomType int32) (*robot, bool) {
	log.Inf("新建机器人 roomType[%d]", roomType)
	r := &robot{}
	r.u = &user{}
	r.u.userInfo = &data.UserInfo{}
	r.u.fishData = &data.FishData{}
	r.u.playData = &playData{}

	r.roomType = roomType
	r.ctxStop, r.stopFunc = context.WithCancel(context.Background())

	ret := r.initRobotInfo()
	return r, ret
}

func (r *robot) initRobotInfo() bool {
	r.u.playData.isRobot = true

	r.u.userInfo.Mid = bson.NewObjectId().Hex()

	// todo 昵称

	cfg := conf.Cf.FishTable.TableRobotConfig.Get(r.roomType)
	actionCfg := conf.Cf.FishTable.TableRobotAction.Get(1)
	if cfg == nil || actionCfg == nil {
		log.Wrn("失败：没有找到机器人配置")
		return false
	}

	r.u.userInfo.Coin = int64(utils.RandInt32(cfg.Coin.One, cfg.Coin.Two))
	r.gameTime = utils.RandInt32(actionCfg.GameTime.One, actionCfg.GameTime.Two)

	log.Inf("新建机器人 成功，mid[%s], coin[%d], gameTime[%d]", r.u.userInfo.Mid, r.u.userInfo.Coin, r.gameTime)

	return true
}

func (r *robot) startGame() {
	cfg := conf.Cf.FishTable.TableRobotConfig.Get(r.roomType)
	actionCfg := conf.Cf.FishTable.TableRobotAction.Get(1)
	if cfg == nil || actionCfg == nil {
		log.Wrn("机器人开始失败：没有找到机器人配置")
		return
	}

	// 机器人开始运行，先执行一次切炮和切行为
	r.changeGun()
	r.changeAction(actionCfg)
	go r.startActionTimer()
	go r.startGunTimer()
}

// 结束游戏，退出房间
func (r *robot) stopGame() {
	r.stopFunc()
	r.changeGunTimer.Stop()
	r.quitTimer.Stop()
	r.changeActionTimer.Stop()
	r.changeSuoDingTimer.Stop()
	r.changeJiaoDuTimer.Stop()
	r.changeSuDuTimer.Stop()
	r.gunTimer.Stop()

	quitGame(r.u.userInfo.Mid, fish_proto.UserQuitReason_ZhuDong)
}

func (r *robot) setCurrAction(action int32) {
	atomic.StoreInt32(&r.currAction, action)
}

func (r *robot) setGunSpeed(speed int32) {
	atomic.StoreInt32(&r.gunSpeedMill, speed)
}

func (r *robot) setSuoDing(b bool, fishId string) {
	r.isSuoDing.Store(b)
	r.suoDingFishId = fishId
}

func (r *robot) getSuoDing() (bool, string) {
	return r.isSuoDing.Load(), r.suoDingFishId
}

func (r *robot) startActionTimer() {
	actionCfg := conf.Cf.FishTable.TableRobotAction.Get(1)

	// 切行为
	changeActionT := utils.RandInt32(actionCfg.ChangeActionTime.One, actionCfg.ChangeActionTime.Two)
	r.changeActionTimer = time.NewTimer(time.Duration(changeActionT) * time.Second)
	// 切枪
	t := utils.RandInt32(actionCfg.ChangeGunTime.One, actionCfg.ChangeGunTime.Two)
	r.changeGunTimer = time.NewTimer(time.Duration(t) * time.Second)
	// 退出
	r.quitTimer = time.NewTimer(time.Duration(r.gameTime) * time.Second)
	for {
		select {
		case <-r.ctxStop.Done():
			return
		case <-r.quitTimer.C: // 退出
			log.Inf("机器人[%s]游戏时间结束，退出房间", r.u.userInfo.Mid)
			r.stopGame()
		case <-r.changeGunTimer.C:
			r.changeGun()
			t := utils.RandInt32(actionCfg.ChangeGunTime.One, actionCfg.ChangeGunTime.Two)
			r.changeGunTimer.Reset(time.Duration(t) * time.Second)
		case <-r.changeActionTimer.C:
			r.changeAction(actionCfg)
			t := utils.RandInt32(actionCfg.ChangeActionTime.One, actionCfg.ChangeActionTime.Two)
			r.changeActionTimer.Reset(time.Duration(t) * time.Second)
		case <-r.changeSuoDingTimer.C:
			fmt.Printf("===testlog 切锁定\n")
			if r.currAction != cfg.RobotActionEnums_SuoDing {
				continue
			}
			r.changeSuoDing(actionCfg)
			// 随机一个时间后，变换锁定目标
			t := utils.RandInt32(actionCfg.SuoDingTime.One, actionCfg.SuoDingTime.Two)
			r.changeSuoDingTimer.Reset(time.Duration(t) * time.Second)
			fmt.Printf("===testlog 切锁定时间[%d]\n", t)
		case <-r.changeJiaoDuTimer.C:
			fmt.Printf("===testlog 切角度\n")
			if r.currAction != cfg.RobotActionEnums_ZiDongGun && r.currAction != cfg.RobotActionEnums_ShouDongGun {
				continue
			}
			r.changeJiaoDu()
			// 随机一个时间后，变换角度
			t := utils.RandInt32(actionCfg.ChangeAngleTime.One, actionCfg.ChangeAngleTime.Two)
			r.changeJiaoDuTimer.Reset(time.Duration(t) * time.Second)
			fmt.Printf("===testlog 切角度时间[%d]\n", t)
		case <-r.changeSuDuTimer.C:
			fmt.Printf("===testlog 切速度\n")
			if r.currAction != cfg.RobotActionEnums_ShouDongGun {
				continue
			}
			r.changeSuDu(actionCfg)
			// 随机一个时间后，变换速度
			t := utils.RandInt32(actionCfg.GunSpeedTime.One, actionCfg.GunSpeedTime.Two)
			r.changeSuDuTimer.Reset(time.Duration(t) * time.Second)
			fmt.Printf("===testlog 切速度时间[%d]\n", t)
		}
	}
}

func (r *robot) startGunTimer() {
	r.gunTimer = time.NewTimer(time.Duration(r.gunSpeedMill) * time.Millisecond)
	for {
		select {
		case <-r.ctxStop.Done():
			return
		case <-r.gunTimer.C:
			fmt.Printf("=====testlog 机器人开枪\n")
			// 广播
			if r.u.playData.room != nil {
				// 扣钱
				if r.u.userInfo.Coin < int64(r.gunValue) {
					// 钱不够了，退出
					log.Inf("机器人[%s]没钱了，退出游戏\n", r.u.userInfo.Mid)
					r.stopGame()
					return
				}

				r.u.userInfo.Coin -= int64(r.gunValue)

				// 开枪
				event := &fish_proto.EventFishPlay{
					Mid: r.u.userInfo.Mid,
					Bet: r.gunValue,
				}
				if suoDing, fishId := r.getSuoDing(); suoDing {
					event.FishId = fishId
					event.BulletSkill = int32(fish_proto.FishBulletSkill_SuoDing)
				} else {
					event.Target = &fish_proto.Vec2{
						X: r.gunTarget.X,
						Y: r.gunTarget.Y,
					}
				}
				r.u.playData.room.send2all(event)
			}

			r.gunTimer.Reset(time.Duration(r.gunSpeedMill) * time.Millisecond)
		}
	}
}

func (r *robot) changeGun() {
	log.Inf("机器人[%s]切炮值，coin[%d]", r.u.userInfo.Mid, r.u.userInfo.Coin)
	cfg := conf.Cf.FishTable.TableRobotConfig.Get(r.roomType)

	min, max := int32(0), int32(0)
	for _, changeGunConfig := range cfg.ChangeGunCoin {
		if r.u.userInfo.Coin >= int64(changeGunConfig.CoinRange.One) &&
			r.u.userInfo.Coin <= int64(changeGunConfig.CoinRange.Two) {
			min = changeGunConfig.GunRange.One
			max = changeGunConfig.GunRange.Two
			break
		}
	}

	if min == 0 || max == 0 {
		log.Wrn("失败：配置数据不对")
		return
	}
	gun := (utils.RandInt32(min, max) / 100) * 100

	log.Inf("原炮值[%d]，新炮值[%d]", r.gunValue, gun)

	r.gunValue = gun

	// 广播桌子
	if r.u.playData.room != nil {
		var event fish_proto.EventFishChangeGun
		event.Bet = gun
		event.Mid = r.u.userInfo.Mid
		r.u.playData.room.send2all(&event)
	}
}

func (r *robot) changeAction(actionCfg *cfg.RobotAction) int32 {

	// 随机一个行为
	randIndex := utils.RandInt(0, len(actionCfg.RobotActionList)-1)
	actionId := actionCfg.RobotActionList[randIndex]
	r.setCurrAction(actionId)
	log.Inf("机器人切行为 new action[%d]", actionId)
	// 先设自动速度，如果是手动开炮，再随机速度
	r.setGunSpeed(_ZiDongGunSpeedMill)

	if actionId == cfg.RobotActionEnums_SuoDing {
		r.changeSuoDing(actionCfg)

		// 随机一个时间后，变换锁定目标
		t := utils.RandInt32(actionCfg.SuoDingTime.One, actionCfg.SuoDingTime.Two)
		r.changeSuoDingTimer = time.NewTimer(time.Duration(t) * time.Second)
	} else {
		r.changeJiaoDu()

		// 随机一个时间后，变换角度
		t := utils.RandInt32(actionCfg.ChangeAngleTime.One, actionCfg.ChangeAngleTime.Two)
		r.changeJiaoDuTimer = time.NewTimer(time.Duration(t) * time.Second)

		if actionId == cfg.RobotActionEnums_ShouDongGun {
			r.changeSuDu(actionCfg)

			// 随机一个时间后，变换速度
			t := utils.RandInt32(actionCfg.GunSpeedTime.One, actionCfg.GunSpeedTime.Two)
			r.changeSuDuTimer = time.NewTimer(time.Duration(t) * time.Second)
		}
	}

	return actionId
}

func (r *robot) changeSuoDing(cfg *cfg.RobotAction) {
	if r.u.playData.room != nil {
		fishList := r.u.playData.room.getAliveFish(cfg.SuoDingFishId)
		// 随机锁定一个鱼
		randIndex := utils.RandInt(0, len(fishList)-1)
		f := fishList[randIndex]
		r.setSuoDing(true, f.id)
	} else {
		// todo
		fishId := "dsf04343"
		fmt.Printf("=====testlog 机器人切锁定\n")

		r.setSuoDing(true, fishId)
	}

}

func (r *robot) changeJiaoDu() {
	x := utils.RandInt32(int32(-_MaxX), int32(_MaxX))
	y := utils.RandInt32(int32(-_MaxY), int32(_MaxY))
	r.gunTarget = vec2{
		X: float64(x),
		Y: float64(y),
	}
	fmt.Printf("=====testlog 机器人切角度 x[%d], y[%d]\n", x, y)
}

func (r *robot) changeSuDu(cfg *cfg.RobotAction) {
	// 一秒内打 gunCount 炮
	gunCount := utils.RandInt32(cfg.GunSpeed.One, cfg.GunSpeed.Two)
	speed := 1000 / gunCount
	r.setGunSpeed(speed)
	fmt.Printf("=====testlog 机器人切速度 speed[%d]\n", speed)
}
