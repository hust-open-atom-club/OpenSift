package controller

import (
	"github.com/HUSTSecLab/OpenSift/cmd/apiserver/internal/controller/admin"
	"github.com/gin-gonic/gin"
)

func Regist(e gin.IRouter) {
	registResult(e)
	admin.Regist(e)
}
