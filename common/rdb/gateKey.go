package rdb

import "fmt"

// 玩家在所在服务器信息  addr(ip+port) gameId 用hash
func KeyGateUsePos(mid string) string {
	return fmt.Sprintf("%v:userpos:%v", preFixGate, mid)
}

// 每个服当前有哪些玩家在线 set
func KeyGateServerUser(sid string, gameId string) string {
	return fmt.Sprintf("%v:serveruser:%v_%v", preFixGate, sid, gameId)
}

// set
func KeyGateActiveServer(gameId string) string {
	return fmt.Sprintf("%v:activeserver:%v", preFixGate, gameId)
}

// set
func KeyGateClosingServer(gameId string) string {
	return fmt.Sprintf("%v:closingserver:%v", preFixGate, gameId)
}
