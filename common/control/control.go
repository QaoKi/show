package control

import (
	"bygame/common/conf"
	"bygame/common/rdb"
	"context"
)

var server *conf.Server

// 会阻塞进程
func Run(s *conf.Server) {
	server = s

	for _, gameId := range server.GameIds {
		rdb.Client().SAdd(context.TODO(), rdb.KeyGateActiveServer(gameId), server.Id)
		rdb.Client().Del(context.TODO(), rdb.KeyGateServerUser(server.Id, gameId))
	}

	run()
}

// 停止
func stop() {
	for _, gameId := range server.GameIds {
		rdb.Client().SRem(context.TODO(), rdb.KeyGateActiveServer(gameId), server.Id)
		rdb.Client().SAdd(context.TODO(), rdb.KeyGateClosingServer(gameId), server.Id)
	}
}

// 真正关闭
func remove() {
	for _, gameId := range server.GameIds {
		rdb.Client().SRem(context.TODO(), rdb.KeyGateActiveServer(gameId), server.Id)
		rdb.Client().SRem(context.TODO(), rdb.KeyGateClosingServer(gameId), server.Id)
		rdb.Client().Del(context.TODO(), rdb.KeyGateServerUser(server.Id, gameId))

	}
}

// 玩家进入游戏
func UserJoin(mid string, gameId string) {
	rdb.Client().SAdd(context.TODO(), rdb.KeyGateServerUser(server.Id, gameId), mid)
	rdb.Client().HSet(context.TODO(), rdb.KeyGateUsePos(mid), map[string]interface{}{"gameId": gameId, "sid": server.Id})
}

// 玩家退出游戏
func UserExit(mid string, gameId string) {
	rdb.Client().SRem(context.TODO(), rdb.KeyGateServerUser(server.Id, gameId), mid)
	rdb.Client().Del(context.TODO(), rdb.KeyGateUsePos(mid))
}

// 获取一个服务器端口，如果有正在进行的就返回老服务器，如果没有就返回新的
func GetGamePort(mid string, gameId string) (addr string, port string, ok bool) {
	m := rdb.Client().HGetAll(context.TODO(), rdb.KeyGateUsePos(mid)).Val()
	existsGameId, ok1 := m["gameId"]
	existsSid, ok2 := m["sid"]

	if ok1 && ok2 {
		if existsGameId == gameId {

			sc, ok := conf.Cf.ServerConf.GetServer(existsSid)
			if ok {
				return sc.Addr, sc.Port, true
			}
		}
	}

	// 随机获取一个
	result := rdb.Client().SRandMember(context.TODO(), rdb.KeyGateActiveServer(gameId)).Val()
	sc, ok := conf.Cf.ServerConf.GetServer(result)
	if ok {
		return sc.Addr, sc.Port, true
	}
	return
}
