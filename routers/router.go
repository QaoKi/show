package routers

import (
	"bygame/common/log"
	"bygame/controllers"

	"net/http"
	"reflect"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"github.com/gin-gonic/gin"
)

/**
 * 自动反序列化封装
 */
type reqFunc[TReq any, TRet any] func(req *TReq, ret *TRet)
type baseRet struct {
	ErrCode int
	ErrMsg  string
}

func Post[TReq any, TRet any](e gin.IRoutes, path string, fn reqFunc[TReq, TRet]) {
	e.POST(path, func(ctx *gin.Context) {
		var req TReq
		if err := ctx.ShouldBindJSON(&req); err != nil {
			ctx.JSON(http.StatusOK, baseRet{1, err.Error()})
			return
		}
		uid := ctx.GetString("TokenId")
		if uid != "" {
			setTokenId(&req, uid)
		}
		var ret TRet
		fn(&req, &ret)
		log.Dbg("path: %v,req: %+v, ret: %+v", path, req, ret)
		ctx.JSON(http.StatusOK, ret)
	})
}

func Get[TReq any, TRet any](e gin.IRoutes, path string, fn reqFunc[TReq, TRet]) {
	e.GET(path, func(ctx *gin.Context) {
		var req TReq
		err := ctx.ShouldBind(&req)
		if err != nil {
			ctx.JSON(http.StatusOK, baseRet{1, err.Error()})
			return
		}
		uid := ctx.GetString("TokenId")
		if uid != "" {
			setTokenId(&req, uid)
		}
		var ret TRet
		fn(&req, &ret)
		ctx.JSON(http.StatusOK, ret)
	})
}

func setTokenId(obj interface{}, uid string) {
	val := reflect.ValueOf(obj).Elem()
	uidField := val.FieldByName("TokenId")
	if uidField.IsValid() && uidField.CanSet() {
		uidField.SetString(uid)
	}
}

/**
 * 开始业务部分
 */

var e *gin.Engine

func Init(en *gin.Engine) {
	e = en
	e.Use(CORSMiddleware())
	defaultInit()
	openRouter()
	gameRouter()
	testRouter()
	serverInner()
	slotRouter()
	rankRouter()
}

// CORSMiddleware 是用于处理跨域请求的中间件
func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, token")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	}
}

func defaultInit() {
	e.GET("/autoApi", func(ctx *gin.Context) { ctx.Redirect(http.StatusFound, "/swagger/index.html") })
	e.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	e.GET("/wss", wss)
}

func openRouter() {
	// 注册路由组
	open := e.Group("open")
	contr := controllers.OpenController{}

	Post(open, "/ping", contr.Ping)
	Post(open, "/guestLogin", contr.GuestLogin)
	Post(open, "/zyLogin", contr.ZyLogin)
	Post(open, "/popLogin", contr.PopLogin)
}

func testRouter() {
	// 注册路由组
	test := e.Group("test")
	contr := controllers.TestController{}
	// todo 内网判断中间件
	// 注册路由
	// 注册和重置密码
	Post(test, "/getUserByUid", contr.GetUserByUid)
	Post(test, "/addCoin", contr.AddCoin)
	Post(test, "/testAddPopCoin", contr.TestAddPopCoin)
}

func gameRouter() {
	// 注册路由组
	game := e.Group("game")
	contr := controllers.GameController{}
	game.Use(controllers.MiddleAuth)

	Post(game, "/gameAddr", contr.GameAddr)
	Post(game, "/gameList", contr.GameList)
	Post(game, "/playRoom", contr.PlayRoom)
	Post(game, "/verifyToken", contr.VerifyToken)
}

func serverInner() {
	serverInner := e.Group("serverInner")
	contr := controllers.ServerInnerController{}
	serverInner.Use(controllers.MiddleInnerServer)

	Post(serverInner, "/broadcast", contr.Broadcast)
}

func slotRouter() {
	slot := e.Group("slot")
	contr := controllers.SlotController{}
	slot.Use(controllers.MiddleAuth)

	Post(slot, "/slotRecord", contr.SlotRecord)
}

func rankRouter() {
	rank := e.Group("rank")
	contr := controllers.RankController{}
	rank.Use(controllers.MiddleAuth)

	Post(rank, "/slotRank", contr.SlotRank)
}
