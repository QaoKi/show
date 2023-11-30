package conf

import (
	"bygame/common/log"
	"encoding/json"
	"io/ioutil"
	"os"
)

type centerConf struct {
	GameList []*GameInfo     `json:"gameList"`
	PlayRoom []*PlayRoomInfo `json:"playRoom"`
}

type GameInfo struct {
	GameId        string `json:"gameId"`        // GameId
	Name          string `json:"name"`          // 游戏名字前端显示使用
	Package       string `json:"package"`       // 包名
	Type          int    `json:"type"`          // 类型 1 普通 2 折叠
	Priority      int    `json:"priority"`      // 优先级 type=2 生效
	Disabel       bool   `json:"disable"`       // 禁用
	RequiredLevel int    `json:"requiredLevel"` // 开放等级
	Classify      string `json:"classify"`      // 折叠名 type=2 生效

	ServerName string `json:"serverName"` // 游戏服对应的名字
}

type PlayRoomInfo struct {
	GameId   string `json:"gameId"`   // gameId
	Ante     int    `json:"ante"`     // 基础倍率
	MinCoin  int    `json:"minCoin"`  // 大厅门槛
	MaxCoin  int    `json:"maxCoin"`  // 金币上限 -1 表示无穷
	RoomType int    `json:"roomType"` // 1 普通房 2 好友房
}

func (c *centerConf) GetGameId(serverName string) string {
	for _, v := range c.GameList {
		if v.ServerName == serverName {
			return v.GameId
		}
	}
	return ""
}

func (c *centerConf) GetServerName(gameId string) string {
	for _, v := range c.GameList {
		if v.GameId == gameId {
			return v.ServerName
		}
	}
	return ""
}

func (c *centerConf) GetPlayRoom(gameid string) (slc []*PlayRoomInfo) {
	for _, v := range c.PlayRoom {
		if v.GameId == gameid && v.RoomType == 1 {
			slc = append(slc, v)
		}
	}
	return
}

func initCenter() {
	var conf centerConf
	f, err := os.Open("./common/conf/center.json")
	if err != nil {
		log.Ftl("加载配置文件失败 err: %v", err)
	}
	b, _ := ioutil.ReadAll(f)
	err = json.Unmarshal(b, &conf)
	if err != nil {
		log.Ftl("center.json 反序列化失败 err: %v", err)
	}
	Cf.CenterConf = &conf
}
