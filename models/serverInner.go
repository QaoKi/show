package models

import (
	"bygame/common/conf"
	"bygame/common/utils"
	"encoding/json"
	"fmt"

	"google.golang.org/protobuf/reflect/protoreflect"
)

const (
	ApiBroadcastPath = "/serverInner/broadcast"
)

// 广播到某个游戏的所有服
func BroadcastGame(gameId string, msg protoreflect.ProtoMessage) {
	go func() {
		if slc, ok := conf.Cf.ServerConf.GetGameServerAddr(gameId); ok {
			for _, url := range slc {
				utils.HttpPostJson(url, broadData(gameId, msg))
			}
		}
	}()
}

// 广播到平台的所有游戏 有些服可能是没有启动的还是要尝试一下
func BroadcastAll(msg protoreflect.ProtoMessage) {
	go func() {
		for _, s := range conf.Cf.ServerConf.Servers {
			url := fmt.Sprintf("http://%v:%v%v", s.Addr, s.Port, ApiBroadcastPath)
			go utils.HttpPostJson(url, broadData("", msg))
		}
	}()
}

func broadData(gameId string, msg protoreflect.ProtoMessage) []byte {
	jbts, _ := utils.Marshal(true, msg)
	pbts, _ := utils.Marshal(false, msg)
	var req struct {
		GameId   string `json:"gameId"`
		JsonMsg  []byte `json:"jsonMsg"`
		ProtoMsg []byte `json:"protoMsg"`
	}
	req.GameId = gameId
	req.JsonMsg = jbts
	req.ProtoMsg = pbts
	bts, _ := json.Marshal(req)
	return bts
}
