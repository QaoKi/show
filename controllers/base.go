package controllers

import (
	"bygame/common/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

func MiddleInnerServer(ctx *gin.Context) {
	// 只能内网访问的接口

}

func MiddleAuth(ctx *gin.Context) {
	// 鉴权
	tokenString := ctx.GetHeader("token")
	if tokenString == "" {
		ctx.JSON(http.StatusOK, BaseRet{10900, "authentication failure"})
		ctx.Abort()
		return
	}
	if ok, uid := utils.VerifyJwt(tokenString); ok {
		ctx.Set("TokenId", uid)
	} else {
		ctx.JSON(http.StatusOK, BaseRet{10900, "authentication failure"})
		ctx.Abort()
		return
	}
	ctx.Next()
}

type BaseRet struct {
	ErrCode int    `json:"code"` //0 成功 1 错误
	ErrMsg  string `json:"msg"`  // 错误信息
}

type BaseReq struct {
	TokenId string `swaggerignore:"true"`
}
