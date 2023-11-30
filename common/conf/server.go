package conf

import (
	"bygame/common/log"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
)

type serverConf struct {
	Servers       []*Server           `json:"servers"`
	gameServerMap map[string][]string `json:"-"`
}

type Server struct {
	Id      string   `json:"id"`
	Name    string   `json:"name"`
	Addr    string   `json:"addr"`
	Port    string   `json:"port"`
	GameIds []string `json:"gameIds"`
}

func (conf *serverConf) GetServer(id string) (*Server, bool) {
	for _, v := range conf.Servers {
		if v.Id == id {
			return v, true
		}
	}
	return nil, false
}

func (conf *serverConf) GetGameServerAddr(gameId string) ([]string, bool) {
	slc, ok := conf.gameServerMap[gameId]
	return slc, ok
}

func initServer() {
	var conf serverConf
	f, err := os.Open("./common/conf/server.json")
	if err != nil {
		log.Ftl("加载配置文件失败 err: %v", err)
	}
	b, _ := ioutil.ReadAll(f)
	err = json.Unmarshal(b, &conf)
	if err != nil {
		log.Ftl("server.json 反序列化失败 err: %v", err)
	}
	Cf.ServerConf = &conf

	conf.gameServerMap = make(map[string][]string)

	for _, s := range conf.Servers {
		for _, gameId := range s.GameIds {
			conf.gameServerMap[gameId] = append(conf.gameServerMap[gameId], fmt.Sprintf("http://%v:%v", s.Addr, s.Port))
		}
	}
}
