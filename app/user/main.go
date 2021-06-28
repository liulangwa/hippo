package main

import (
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/liulangwa/hippo/controller/middleware"
)

func health(ctx *gin.Context) {
	ctx.String(http.StatusOK, "ok")
}

func user(ctx *gin.Context) {
	time.Sleep(time.Second * 2)
	ctx.String(http.StatusOK, "root")
}

func main() {

	g := gin.Default()

	g.Use(middleware.OpenTracer("com.service.user", "172.26.158.228:6831"))
	g.GET("/healthz", health)
	g.GET("/user", user)

	err := g.Run(":8080")
	if err != nil {
		log.Panicln(err)
		return
	}

}
