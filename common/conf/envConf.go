package conf

import (
	"bygame/common/log"

	"gopkg.in/gcfg.v1"
)

type envConf struct {
	Base struct {
		JwtSecret string `gcfg:"JWTSECRET"`
		JwtExpire int    `gcfg:"JWTEXPIRE"`
	} `gcfg:"BASE"`

	Mongo struct {
		Host     string `gcfg:"HOST"`
		Port     string `gcfg:"PORT"`
		Db       string `gcfg:"DB"`
		User     string `gcfg:"USER"`
		Password string `gcfg:"PASSWORD"`
	} `gcfg:"MONGO"`

	Redis struct {
		Host     string `gcfg:"HOST"`
		Port     string `gcfg:"PORT"`
		Db       int    `gcfg:"DB"`
		User     string `gcfg:"USER"`
		Password string `gcfg:"PASSWORD"`
	} `gcfg:"REDIS"`

	Pop struct {
		Host      string `gcfg:"HOST"`
		Port      string `gcfg:"PORT"`
		AppKey    string `gcfg:"APPKEY"`
		AppSecret string `gcfg:"APPSECRET"`
	} `gcfg:"POP"`
}

func initEnv() {
	var conf envConf
	err := gcfg.ReadFileInto(&conf, ".env")
	if err != nil {
		log.Ftl(err.Error())
	}
	Cf.EnvConf = &conf
}
