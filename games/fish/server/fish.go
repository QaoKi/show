package fish

import (
	"bygame/common/conf/fish/cfg"
	"bygame/common/proto/fish_proto"

	"sync/atomic"
	"time"

	"gopkg.in/mgo.v2/bson"
)

/*
 */
type vec2 struct {
	X float64
	Y float64
}

type bezierPath struct {
	path    [4]vec2
	genTime int64 // 创建时间
	desTime int64 // 销毁时间
}

type fish struct {
	id       string        // 鱼的实例id
	cid      int32         // 鱼的配置id
	pathList []*bezierPath // 鱼的路径
	state    int32         // 鱼的状态 0 活着 1 死亡
	speed    int32         // 速度
	fc       *cfg.FishInfo // 鱼的配置

	touchBufferRate    int32 // 鱼触发buffer时，炮的倍率
	touchSpecialBuffer bool  // 是否触发了特殊buffer
}

// 判断鱼是否存活
func (f *fish) isLive() bool {

	if time.Now().Unix() > f.pathList[len(f.pathList)-1].desTime {
		return false
	}

	return atomic.LoadInt32(&f.state) == 0
}

// 击杀鱼
func (f *fish) killFish() bool {
	return atomic.CompareAndSwapInt32(&f.state, 0, 1)
}

// 鱼阵
type fishFormat struct {
	LotFishList [][]*fish // 每个轮次的鱼
	LotInterval int32     // 每个轮次的间隔
}

func newFish(fc *cfg.FishInfo, speed int32, paths [][4]vec2) *fish {
	var f fish
	f.fc = fc
	f.id = bson.NewObjectId().Hex()
	f.cid = int32(fc.FishId)
	f.speed = speed

	timeStart := time.Now().Unix()
	for i := 0; i < len(paths); i++ {
		bp := &bezierPath{
			path: paths[i],
		}

		// 现在只是创建鱼的时间，等鱼要出去的时候再更新时间
		pathLen := bezierLength(paths[i])
		pathTime := pathLen / float64(f.speed)
		bp.genTime = timeStart
		bp.desTime = bp.genTime + int64(pathTime)
		timeStart = bp.desTime

		f.pathList = append(f.pathList, bp)
	}
	return &f
}

func (f *fish) updateTime() {
	// 要出鱼了，更新一下时间
	timeNow := time.Now().Unix()
	for _, bp := range f.pathList {
		t := bp.desTime - bp.genTime
		bp.genTime = timeNow
		bp.desTime = bp.genTime + t
	}
}

// 是否是不死鱼
func (f *fish) isImmortalFish() bool {
	return f.fc.Immortal == 1
}

func fishConvertProto(fishList []*fish) []*fish_proto.FishInfo {
	p := []*fish_proto.FishInfo{}
	for _, f := range fishList {
		fish := &fish_proto.FishInfo{}
		fish.Cid = f.cid
		fish.FishId = f.id
		fish.Speed = f.speed

		for _, path := range f.pathList {
			bp := &fish_proto.BezierPath{}
			bp.GenTime = path.genTime
			bp.DesTime = path.desTime
			for _, v := range path.path {
				v1 := &fish_proto.Vec2{
					X: v.X,
					Y: v.Y,
				}
				bp.Path = append(bp.Path, v1)
			}
			fish.PathList = append(fish.PathList, bp)
		}
		p = append(p, fish)
	}

	return p
}
