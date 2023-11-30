package conf

import (
	fishcfg "bygame/common/conf/fish/cfg"
	slotcfg "bygame/common/conf/slot/cfg"
	"bygame/common/log"
	"encoding/json"
	"io/ioutil"
)

var Cf Conf

type Conf struct {
	CenterConf *centerConf
	ServerConf *serverConf
	EnvConf    *envConf

	FishTable *fishcfg.Tables
	SlotTable *slotcfg.Tables

	SelfServerConf *Server // 自身服务信息
}

func Init() {
	initCenter()
	initServer()
	initEnv()

	// luban
	initTableCfg()
}

func (cf *Conf) SetSelfServerConf(s *Server) {
	cf.SelfServerConf = s
}

func initTableCfg() {
	fishTable, err := fishcfg.NewTables(fishLoader)
	if err != nil {
		log.Wrn("配置表加载失败 game: fish, err: %v", err)
	}
	Cf.FishTable = fishTable

	slotTable, err := slotcfg.NewTables(slotLoader)
	if err != nil {
		log.Wrn("配置表加载失败 game: slot, err: %v", err)
	}
	Cf.SlotTable = slotTable
}

func fishLoader(file string) ([]map[string]interface{}, error) {
	if bytes, err := ioutil.ReadFile("/Users/zhaofei/work/bygame/common/conf/fish/json/" + file + ".json"); err != nil {
		return nil, err
	} else {
		jsonData := make([]map[string]interface{}, 0)
		if err = json.Unmarshal(bytes, &jsonData); err != nil {
			return nil, err
		}
		return jsonData, nil
	}
}

func slotLoader(file string) ([]map[string]interface{}, error) {
	// path := "/Users/3000li/go/src/bygame/common/conf/slot/json/"
	path := "./common/conf/slot/json/"
	if bytes, err := ioutil.ReadFile(path + file + ".json"); err != nil {
		return nil, err
	} else {
		jsonData := make([]map[string]interface{}, 0)
		if err = json.Unmarshal(bytes, &jsonData); err != nil {
			return nil, err
		}
		return jsonData, nil
	}
}
