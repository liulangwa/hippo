package middleware

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/liulangwa/hippo/libraries/constants"
	"github.com/liulangwa/hippo/libraries/jaeger"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	opentracinglog "github.com/opentracing/opentracing-go/log"
)

//OpenTracer tracer中间件
func OpenTracer(serviceName string, addr string) gin.HandlerFunc {

	tracer, _, err := jaeger.NewJaegerTracer(serviceName, addr)

	if err != nil {
		panic(err)
	}

	opentracing.SetGlobalTracer(tracer)

	return func(ctx *gin.Context) {

		//从 ctx.Request.Header 中提取 span,如果没有就新建一个
		parentSpanCtx, err := opentracing.GlobalTracer().Extract(opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(ctx.Request.Header))
		var span opentracing.Span

		//root span
		if err != nil {
			span = tracer.StartSpan(ctx.Request.URL.String())
			log.Printf("root span:%+v", span)
		} else {
			span = tracer.StartSpan(ctx.Request.URL.String(), ext.RPCServerOption(parentSpanCtx))
			log.Printf("child span:%+v", span)
		}

		defer span.Finish()

		ext.HTTPMethod.Set(span, ctx.Request.Method)
		ext.Component.Set(span, ctx.Request.URL.Scheme)

		ctx.Set(constants.ETracingSpan, opentracing.ContextWithSpan(context.Background(), span))

		span.SetTag(constants.ETracingTraceId, jaeger.GetTraceID(span))
		span.SetTag(constants.ETracingSpanId, jaeger.GetSpanID(span))

		//ctx.Request = ctx.Request.WithContext(
		//	opentracing.ContextWithSpan(ctx.Request.Context(), span))

		ctx.Next()

		ext.HTTPStatusCode.Set(span, uint16(ctx.Writer.Status()))

		if ctx.Writer.Status() >= 500 {

			span.LogFields(
				opentracinglog.String("event", "error"),
				opentracinglog.String("message", errors.New("test error").Error()),
				opentracinglog.Int64("error time ", time.Now().Unix()),
			)
			//设置span为错误
			span.SetTag("error", true)
		}

	}
}
