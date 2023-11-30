package controllers

import (
	"bygame/common/log"
	"bygame/gate"
)

// 服务器间的消息,需要挂内网判断中间件
type ServerInnerController struct {
}

type reqBroadcast struct {
	GameId   string `json:"gameId"`
	JsonMsg  []byte `json:"jsonMsg"`
	ProtoMsg []byte `json:"protoMsg"`
}

type retBroadcast struct {
	BaseRet
}

// 收到这个消息直接转发到所有服
func (ServerInnerController) Broadcast(req *reqBroadcast, ret *retBroadcast) {
	log.Inf("发送广播 gameId: %v msg: %v", req.GameId, string(req.JsonMsg))
	gate.Broadcast2AllClient(req.GameId, req.JsonMsg, req.ProtoMsg)
}
