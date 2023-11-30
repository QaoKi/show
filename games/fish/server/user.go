package fish

import (
	"bygame/common/data"
	"bygame/common/proto/fish_proto"

	"sync/atomic"
)

// 玩家
type user struct {
	fishData *data.FishData
	userInfo *data.UserInfo
	playData *playData // 房间内相关的数据
}

type playData struct {
	seat      int32 // 位置
	room      *room // 房间
	isRobot   bool  // 是否是机器人
	gunValue  int32 // 炮值
	isDisCoon bool  // 是否断线

	userStatus fish_proto.FishUserStatus // 玩家状态，是否在玩一些小游戏

	labaBuffer int32                // 是否能玩拉霸小游戏，0：不能玩，1：能玩
	labaData   *fish_proto.LabaData // 拉霸的一些数据

	bufferZhuanPanCount    int32                            // 能玩buffer转盘的次数
	bufferZhuanPanDataList []*fish_proto.BufferZhuanPanData // 转盘的一些数据

	moguBufferCount int32                  // 能玩蘑菇庄园的次数
	moguDataList    []*fish_proto.MoguData // 蘑菇庄园的一些数据
}

func (p *playData) resetData() {
	p.seat = -1
	p.room = nil
	p.isDisCoon = false
}

func NewUser(userInfo *data.UserInfo, fishData *data.FishData) {
	var u user
	u.fishData = fishData
	u.userInfo = userInfo
	u.playData = &playData{}
	_userM.addUser(&u)
}

func (u *user) setDisConStatus(disConn bool) {
	u.playData.isDisCoon = disConn
}

func (u *user) getDisConnStatus() bool {
	return u.playData.isDisCoon
}

func (u *user) setUserStatus(s fish_proto.FishUserStatus) {
	u.playData.userStatus = s
}

func (u *user) getUserStatus() fish_proto.FishUserStatus {
	return u.playData.userStatus
}

// *********************************************

// 触发拉霸小游戏
func (u *user) addLaBaBuffer() {
	atomic.SwapInt32(&u.playData.labaBuffer, 1)
}

// 玩拉霸小游戏
func (u *user) playLaBa() bool {
	return atomic.CompareAndSwapInt32(&u.playData.labaBuffer, 1, 0)
}

// 触发转盘小游戏
func (u *user) addBufferZhuanPan(num int32) {
	atomic.AddInt32(&u.playData.bufferZhuanPanCount, num)
}

// 玩转盘小游戏
func (u *user) playBufferZhuanPan() bool {
	return atomic.AddInt32(&u.playData.bufferZhuanPanCount, -1) >= 0
}

// 触发蘑菇庄园
func (u *user) addMoGuBuffer(num int32) {
	atomic.AddInt32(&u.playData.moguBufferCount, num)
}

// 玩蘑菇庄园
func (u *user) playMoGu() bool {
	return atomic.AddInt32(&u.playData.moguBufferCount, -1) >= 0
}

// 加减activity转盘积分
func (u *user) addActivityZhuanPanScore(score int64) {
	atomic.AddInt64(&u.fishData.ZhuanPanScore, score)
}

// 加减连击奖励
func (u *user) addComboReward(coin int64) {
	atomic.AddInt64(&u.fishData.ComboReward, coin)
}

// 清空连击奖励
func (u *user) resetComboReward() {
	atomic.StoreInt64(&u.fishData.ComboReward, 0)
}
