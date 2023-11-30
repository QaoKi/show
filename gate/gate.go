package gate

import (
	"bygame/common/control"
	"bygame/common/data"
	"bygame/common/log"
	"bygame/common/mdb"
	"bygame/common/proto/common_proto"
	"bygame/common/utils"

	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"google.golang.org/protobuf/reflect/protoreflect"
	"gopkg.in/mgo.v2/bson"
)

func Init() {
	gum.um = make(map[string]*gateUser)
	gum.funcm = make(map[string]map[int32]func(mid string, baseCmd *common_proto.BaseCmd))
	gum.offFunc = make(map[string]func(mid string))
	gum.reConnFunc = make(map[string]func(mid string))
	gum.kickOutFunc = make(map[string]func(mid string))
	gum.autoClose = make(map[string]int)
	go saveGameData()
	go saveCoin()
}

var gum gateUserManage

type gateUser struct {
	mid          string          // 玩家Id
	conn         *websocket.Conn // 网络连接
	json         bool            // 是否采用json序列化,默认使用proto
	gameId       string          // 游戏id
	lastTime     int64           // 最后发消息时间
	gameDataItem any             // 游戏内数据
	gameDatakey  string          // 游戏数据key
	gameDateMd5  string          // md5,用来计算游戏数据的变动
	userInfo     *data.UserInfo  // 玩家详情,放在这里也是为了做快速的修改,GM修改数据的时候要直接修改内存数据或者等玩家下线
	userInfoMd5  string          // md5

	sendMessageChan chan []byte
	ctxStop         context.Context
	stopFunc        context.CancelFunc
	closeOnce       sync.Once
}

func NewGateUser(mid string, conn *websocket.Conn, isjson bool, gameId string, d any, userInfo *data.UserInfo) {
	fullName := fmt.Sprintf("%v", reflect.TypeOf(d))
	index := strings.Index(fullName, ".")
	var key string
	if index != -1 {
		key = strings.ToLower(fullName[index+1:])
	} else {
		key = strings.ToLower(fullName)
	}
	u := gateUser{mid: mid, conn: conn, json: isjson, gameId: gameId, gameDataItem: d, gameDatakey: key, userInfo: userInfo}

	u.sendMessageChan = make(chan []byte, 10)
	u.ctxStop, u.stopFunc = context.WithCancel(context.Background())

	add(&u)
}

func IsReconn(mid string, gameId string) (bool, error) {
	if gu, ok := get(mid); ok {
		if gu.gameId != gameId {
			return false, fmt.Errorf("other games going")
		}
		return true, nil
	}
	return false, nil
}

func Reconn(mid string, conn *websocket.Conn) {
	if gu, ok := get(mid); ok {
		gu.conn = conn
		startMessage(gu)
		// 重连回调
		f, ok := gum.reConnFunc[gu.gameId]
		if ok {
			f(mid)
		}
	}

}

func startMessage(gu *gateUser) {
	gu.sendMessageChan = make(chan []byte, 10)
	gu.ctxStop, gu.stopFunc = context.WithCancel(context.Background())
	gu.closeOnce = sync.Once{}
	go gu.readMessage()
	go gu.sendMessage()
}

type gateUserManage struct {
	um          map[string]*gateUser
	mu          sync.RWMutex
	funcm       map[string]map[int32]func(mid string, baseCmd *common_proto.BaseCmd) // 所有的消息监听
	offFunc     map[string]func(mid string)                                          // 断连回调
	reConnFunc  map[string]func(mid string)                                          // 重连回调
	kickOutFunc map[string]func(mid string)                                          // 踢人回调
	autoClose   map[string]int                                                       // 自动断连时间
}

func add(u *gateUser) {
	bts, _ := json.Marshal(u.gameDataItem)
	arr := md5.Sum(bts)
	md5str := hex.EncodeToString(arr[:])
	u.gameDateMd5 = md5str
	gum.mu.Lock()
	gum.um[u.mid] = u
	gum.mu.Unlock()
	control.UserJoin(u.mid, u.gameId)

	startMessage(u)
}

func get(mid string) (*gateUser, bool) {
	gum.mu.RLock()
	defer gum.mu.RUnlock()
	u, ok := gum.um[mid]
	return u, ok
}

func (u *gateUser) readMessage() {

	defer func() {
		err := recover()
		log.Inf("玩家[%s]接收数据协程退出 err:%v", u.mid, err)
		u.close(true)
	}()

	for {
		select {
		case <-u.ctxStop.Done():
			log.Inf("玩家[%s]接收数据协程，收到退出信号", u.mid)
			return
		default:
			_, bts, err := u.conn.ReadMessage()
			if err != nil {
				log.Err("接收玩家[%s] 数据出错, err[%v]", u.mid, err)
				return
			}

			baseCmd := common_proto.BaseCmd{}
			if err := utils.Unmarshal(u.json, bts, &baseCmd); err != nil {
				log.Wrn("反序列化玩家[%s] baseCmd 数据出错, err[%v]", u.mid, err)
				continue
			}

			if baseCmd.CmdId != int32(common_proto.CmdId_UserKeepAlive) {
				log.Inf("接收到玩家[%s]数据, baseCmdId[%d]", u.mid, baseCmd.CmdId)
			}

			// 统一的心跳处理
			if baseCmd.CmdId == int32(common_proto.CmdId_UserKeepAlive) {
				var ret common_proto.RetUserKeepAlive
				ret.CurrTime = time.Now().UnixMilli()
				baseCmd.Data, _ = utils.Marshal(u.json, &ret)
				u.send2u(&baseCmd)
				continue
			}

			if f, ok := getFunc(u.gameId, baseCmd.CmdId); ok {
				u.lastTime = time.Now().Unix()
				f(u.mid, &baseCmd)
			} else {
				baseCmd.Code = 1
				baseCmd.Data = nil
				u.send2u(&baseCmd)
			}
		}
	}
}

func (u *gateUser) sendMessage() {
	defer func() {
		err := recover()
		log.Inf("玩家[%s]发送数据协程退出 err:%v", u.mid, err)
		u.close(true)
	}()

	for {
		select {
		case <-u.ctxStop.Done():
			log.Inf("玩家[%s]发送数据协程，收到退出信号", u.mid)
			return
		case bts, ok := <-u.sendMessageChan:
			if !ok {
				return
			}

			baseCmd := common_proto.BaseCmd{}
			utils.Unmarshal(u.json, bts, &baseCmd)

			err := u.writeMessage(bts)
			if err != nil {
				log.Err("发送消息到客户端失败,发送失败 mid: %v, msg: %+v, err: %v", u.mid, baseCmd.CmdId, err)
				return
			}
		}
	}
}

func (u *gateUser) close(cb bool) {
	u.closeOnce.Do(func() {
		u.stopFunc()

		if u.conn != nil {
			u.conn = nil
		}

		if cb {
			f, ok := gum.offFunc[u.gameId]
			if ok {
				f(u.mid)
			}
		}
	})
}

func (u *gateUser) kickOut() {
	f, ok := gum.kickOutFunc[u.gameId]
	if ok {
		f(u.mid)
	}
}

func (u *gateUser) send2u(msg protoreflect.ProtoMessage) error {
	bts, err := utils.Marshal(u.json, msg)
	if err != nil {
		log.Err("发送消息到客户端失败,消息序列化失败 mid: %v, msg: %+v, err: %v", u.mid, msg, err)
		return err
	}

	{
		// 测试代码
		basCmd := common_proto.BaseCmd{}
		err = utils.Unmarshal(u.json, bts, &basCmd)
		log.Inf("======testlog 发给客户端的消息，err[%v]，消息号[%d]", err, basCmd.CmdId)
	}

	select {
	case u.sendMessageChan <- bts:
	case <-u.ctxStop.Done():
		return nil
	}
	return nil
}

// 这个函数不再单独调用，统一通过 sendMessage() 调用
func (u *gateUser) writeMessage(bts []byte) (err error) {
	if u.json {
		err = u.conn.WriteMessage(websocket.TextMessage, bts)
	} else {
		err = u.conn.WriteMessage(websocket.BinaryMessage, bts)
	}
	return
}

func getFunc(gameId string, cmdId int32) (func(string, *common_proto.BaseCmd), bool) {
	fs, ok := gum.funcm[gameId]
	if !ok {
		return nil, ok
	}
	if f, ok := fs[cmdId]; ok {
		return f, ok
	}
	return nil, false
}

/*
注册消息
*/
type reqFunc[TReq protoreflect.ProtoMessage, TRet protoreflect.ProtoMessage] func(mid string, req TReq, ret TRet) int32

func AddLintener[TReq protoreflect.ProtoMessage, TRet protoreflect.ProtoMessage](gameId string, fn reqFunc[TReq, TRet]) {
	var req1 TReq
	// *xxx.ReqPing
	fullName := fmt.Sprintf("%v", reflect.TypeOf(req1))

	// *xxx.ReqPing --> *xxx.Ping
	fullName = strings.Replace(fullName, "Req", "", 1)

	// *xxx.Ping --> Ping
	index := strings.Index(fullName, ".")
	if index == -1 {
		log.Err("消息监听注册失败,检查消息结构体是否正确 %v", fullName)
		return
	}
	cmdName := fullName[index+1:]

	// cmdName := strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(fmt.Sprintf("%v", reflect.TypeOf(req1)), "cmd.", ""), "Req", ""), "*", "")
	cmdId := common_proto.CmdId_value[cmdName]
	fs, ok := gum.funcm[gameId]
	if !ok {
		fs = make(map[int32]func(mid string, baseCmd *common_proto.BaseCmd))
		gum.funcm[gameId] = fs
	}
	fs[cmdId] = func(mid string, basCmd *common_proto.BaseCmd) {
		u, ok := get(mid)
		if !ok {
			log.Err("onmsg 网关未找到玩家 mid: %v, cmd: %+v", mid, basCmd)
			return
		}
		var req TReq
		var ret TRet

		// 实例化对象 `{}`
		json.Unmarshal([]byte{123, 125}, &ret)
		json.Unmarshal([]byte{123, 125}, &req)

		// 选择序列化器
		utils.Unmarshal(u.json, basCmd.Data, req)
		code := fn(mid, req, ret)
		// 构造响应对象
		if code != 0 {
			basCmd.Code = int32(code)
			basCmd.Data = nil
		} else {
			b, err := utils.Marshal(u.json, ret)
			if err != nil {
				basCmd.Code = 1
			}
			basCmd.Data = b
		}
		u.send2u(basCmd)
	}
}

func (u *gateUser) isChangeGd() bool {
	bts, _ := json.Marshal(u.gameDataItem)
	arr := md5.Sum(bts)
	md5str := hex.EncodeToString(arr[:])
	if md5str != u.gameDateMd5 {
		u.gameDateMd5 = md5str
		return true
	}
	return false
}

func (u *gateUser) isChangeUi() bool {
	bts, _ := json.Marshal(u.userInfo)
	arr := md5.Sum(bts)
	md5str := hex.EncodeToString(arr[:])
	if md5str != u.userInfoMd5 {
		u.userInfoMd5 = md5str
		return true
	}
	return false
}

// 发送消息到客户端 msg 小包 框架会自动包外层大包
func SendEventToUser(mid string, msg protoreflect.ProtoMessage) error {
	u, ok := get(mid)
	if !ok {
		log.Err("发送消息到客户端失败,网关未找到玩家 mid: %v, msg: %+v", mid, msg)
		return fmt.Errorf("user not found")
	}
	// *common_proto.EventPong
	fullName := fmt.Sprintf("%v", reflect.TypeOf(msg))

	// *common_proto.EventPong --> EventPong
	index := strings.Index(fullName, ".")
	if index == -1 {
		return fmt.Errorf("event type err")
	}
	fullName = fullName[index+1:]

	// EventPong --> Pong
	// eventName := strings.Replace(fullName, "Event", "", 1)
	eventName := fullName

	// eventName := strings.Replace(strings.Replace(fmt.Sprintf("%v", reflect.TypeOf(msg)), "*cmd.", "", 1), "Event", "", 1)
	if cmdId, ok := common_proto.CmdId_value[eventName]; ok {
		var baseCmd common_proto.BaseCmd
		bts, err := utils.Marshal(u.json, msg)
		if err != nil {
			return err
		}
		baseCmd.CmdId = cmdId
		baseCmd.Data = bts
		u.send2u(&baseCmd)
	} else {
		return fmt.Errorf("cmdId not found")
	}
	return nil
}

// 设置每个游戏的连接断开回调
func SetOfflineFunc(gameId string, f func(mid string)) {
	gum.offFunc[gameId] = f
}

// 重连回调
func SetReConnFunc(gameId string, f func(mid string)) {
	gum.reConnFunc[gameId] = f
}

// 踢人回调
func SetKickOutFunc(gameId string, f func(mid string)) {
	gum.kickOutFunc[gameId] = f
}

// 销毁网关用户,游戏user销毁时必须调用
func DestroyGateUser(mid string) {
	if u, ok := get(mid); ok {
		log.Inf("销毁网关用户 mid: %v", mid)
		gum.mu.Lock()
		delete(gum.um, mid)
		gum.mu.Unlock()
		control.UserExit(mid, u.gameId)

		u.close(false)

		// 下面是数据
		if u.isChangeGd() {
			save(u.gameDatakey, u.mid, u.gameDataItem)
		}

		if u.isChangeUi() {
			// 保存用户共享数据
			db := mdb.GetMdb()
			db.C(mdb.DB_USER).Update(bson.M{"_id": bson.ObjectIdHex(mid)}, bson.M{"$set": bson.M{"userinfo": u.userInfo}})
		}
	}
}

// 多少秒没有业务操作自动断开,并不准确要 + [0,30]
func SetAutoCloseTime(gameId string, d int) {
	gum.autoClose[gameId] = d
}

func saveCoin() {
	f := func() {
		keys := make([]string, 0, len(gum.um))
		gum.mu.RLock()
		for k := range gum.um {
			keys = append(keys, k)
		}
		gum.mu.RUnlock()
		for _, mid := range keys {
			if u, ok := get(mid); ok {

				if u.isChangeUi() {
					// 这里主要更新一下金币
					db := mdb.GetMdb()
					db.C(mdb.DB_USER).Update(bson.M{"_id": bson.ObjectIdHex(mid)}, bson.M{"$set": bson.M{"userinfo": u.userInfo}})
				}
			}
		}
	}

	for range time.NewTicker(30 * time.Second).C {
		f()
	}
}

func saveGameData() {
	f := func() {
		keys := make([]string, 0, len(gum.um))
		gum.mu.RLock()
		for k := range gum.um {
			keys = append(keys, k)
		}
		gum.mu.RUnlock()

		now := time.Now().Unix()
		for _, mid := range keys {
			if u, ok := get(mid); ok {
				d := gum.autoClose[u.gameId]
				// 保存游戏数据
				if u.isChangeGd() {
					save(u.gameDatakey, u.mid, u.gameDataItem)
				}

				if u.isChangeUi() {
					// 保存用户共享数据
					db := mdb.GetMdb()
					db.C(mdb.DB_USER).Update(bson.M{"_id": bson.ObjectIdHex(mid)}, bson.M{"$set": bson.M{"userinfo": u.userInfo}})
				}

				// 踢人
				if d > 0 && u.lastTime+int64(d) < now {
					u.kickOut()
				}

			}
		}
	}

	for range time.NewTicker(2 * time.Second).C {
		f()
	}
}

func Broadcast2AllClient(gameId string, jsonMsg []byte, protMsg []byte) {
	// 转发消息
	gum.mu.RLock()
	keys := make([]string, 0, len(gum.um))
	for k := range gum.um {
		keys = append(keys, k)
	}
	gum.mu.RUnlock()
	for _, mid := range keys {
		if u, ok := get(mid); ok && (gameId == "" || gameId == u.gameId) {
			basCmd := common_proto.BaseCmd{}
			if u.json {
				utils.Unmarshal(u.json, jsonMsg, &basCmd)
			} else {
				utils.Unmarshal(u.json, protMsg, &basCmd)
			}
			u.send2u(&basCmd)
		}
	}
}

func save(key, mid string, data any) {
	db := mdb.GetMdb()
	db.C(mdb.DB_GAMEDATA).Upsert(bson.M{"_id": bson.ObjectIdHex(mid)}, bson.M{"$set": bson.M{key: data}})
}
