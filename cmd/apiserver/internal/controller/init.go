package controller

import "github.com/gin-gonic/gin"

func Regist(e gin.IRouter, rpcAddr string) {
	registResult(e)
	registAdmin(e, rpcAddr)
}
