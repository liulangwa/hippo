package main

import (
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/liulangwa/hippo/controller/middleware"
	"github.com/liulangwa/hippo/libraries/jaeger"
	"github.com/opentracing/opentracing-go"
)

func health(ctx *gin.Context) {
	ctx.String(http.StatusOK, "ok")
}

func product(ctx *gin.Context) {

	jaeger.Send(ctx, opentracing.GlobalTracer(), http.MethodGet, "http://localhost:8080/user", nil, nil, time.Second*10)
	resp, err := jaeger.Send(ctx, opentracing.GlobalTracer(), http.MethodGet, "http://localhost:8080/user", nil, nil, time.Second*10)

	if err != nil {
		log.Printf("%v", err)
	}
	log.Printf("resp:%v", resp)

	ctx.String(http.StatusOK, "ok")
}

func main() {

	g := gin.Default()

	g.Use(middleware.OpenTracer("com.service.product", "172.26.158.228:6831"))
	g.GET("/healthz", health)
	g.GET("/product", product)

	err := g.Run(":8081")
	if err != nil {
		log.Panicln(err)
		return
	}

}
