package main

import (
	"bygame/common/conf"
	"bygame/common/control"
	"bygame/common/log"
	"bygame/common/mdb"
	"bygame/common/rdb"
	"bygame/common/utils"
	fish "bygame/games/fish/server"
	"bygame/gate"
	"bygame/routers"
	"net/http"
	"net/http/pprof"

	"github.com/gin-gonic/gin"
)

// @title			bygame
// @version		v1.0
// @contact.name	lizenghui
// @contact.email	lizenghui0827@163.com
// @securitydefinitions.apikey ApiKeyAuth
// @in header
// @name token
func main() {
	serverId := utils.GetArgs("serverId")
	log.Init(`{"LogDir":"logs","FuncCallDepth":3,"Console":true}`)
	conf.Init()
	server, ok := conf.Cf.ServerConf.GetServer(serverId)
	if !ok {
		log.Ftl("未找到当前服务id对应的配置")
		return
	}
	conf.Cf.SetSelfServerConf(server)
	startPProf()
	mdb.Init()
	rdb.Init("center")
	gate.Init()
	fish.Init()
	ginInit(server.Port)
	control.Run(server)
}

func startPProf() {
	pprofAddr := ":7890"
	pprofHandler := http.NewServeMux()
	pprofHandler.Handle("/debug/pprof/", http.HandlerFunc(pprof.Index))
	server := &http.Server{Addr: pprofAddr, Handler: pprofHandler}
	go server.ListenAndServe()
}

func ginInit(p string) {
	go func() {
		gin.SetMode(gin.ReleaseMode)
		r := gin.Default()
		routers.Init(r)
		r.Run(":" + p)
	}()
}
