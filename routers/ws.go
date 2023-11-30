package routers

import (
	"bygame/common/data"
	"bygame/common/def"
	"bygame/common/log"
	"bygame/common/mdb"
	"bygame/common/utils"
	fish "bygame/games/fish/server"
	"bygame/gate"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"gopkg.in/mgo.v2/bson"
)

func wss(ctx *gin.Context) {
	var upGrader = websocket.Upgrader{CheckOrigin: func(r *http.Request) bool {
		return true
	}}
	conn, err := upGrader.Upgrade(ctx.Writer, ctx.Request, nil)
	if err != nil {
		conn.Close()
		ctx.JSON(500, struct{}{})
		return
	}
	// 验证token
	token := ctx.Request.URL.Query().Get("token")
	gameId := ctx.Request.URL.Query().Get("gameId")
	protoc := ctx.Request.URL.Query().Get("protoc")
	log.Dbg("建立websocket 连接 token: %v,gameId: %v", token, gameId)

	ok, mid := utils.VerifyJwt(token)
	if !ok {
		conn.Close()
		ctx.JSON(http.StatusBadRequest, "user not found")
		return
	}

	log.Inf("连接websocket mid: %v", mid)

	db := mdb.GetMdb()
	var user data.User

	db.C(mdb.DB_USER).Find(bson.M{"_id": bson.ObjectIdHex(mid)}).One(&user)
	if user.Mid != mid {
		log.Wrn("没有找到玩家数据 mid: %v", mid)
		conn.Close()
		ctx.JSON(http.StatusBadRequest, "user not found")
		return
	}

	log.Dbg("建立websocket连接,验证通过 mid: %v,gameId: %v", mid, gameId)

	// 判断是不是重连
	reconn, err := gate.IsReconn(mid, gameId)
	if err != nil {
		conn.Close()
		ctx.JSON(http.StatusBadRequest, err.Error())
		return
	}

	if reconn {
		gate.Reconn(mid, conn)
		return
	}

	log.Dbg("玩家登录游戏 %+v", user)

	// 加载游戏数据
	var gameData data.GameData
	db.C(mdb.DB_GAMEDATA).Find(bson.M{"_id": bson.ObjectIdHex(mid)}).One(&gameData)

	isJson := protoc == "json"

	switch gameId {
	case def.GameIdDomino, def.GameIdDominoBet, def.GameIdDominoWild:
		log.Inf("未实现的游戏")
	case def.GameIdNewFishingEra:
		gate.NewGateUser(mid, conn, isJson, gameId, &gameData.FishData, &user.UserInfo)
		fish.NewUser(&user.UserInfo, &gameData.FishData)
	default:
		ctx.JSON(http.StatusBadRequest, "gameId not found")
	}
}
