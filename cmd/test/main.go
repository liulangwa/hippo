package main

import (
	"log"
	"net/http"
	"time"

	"github.com/liulangwa/hippo/libraries/jaeger"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
)

func main() {

	tracer, _, err := jaeger.NewJaegerTracer("com.service.client", "172.26.158.228:6831")

	if err != nil {
		log.Panic(err)
		return
	}

	clientSpan := tracer.StartSpan("client")
	defer clientSpan.Finish()

	url := "http://localhost:8081/product"
	req, _ := http.NewRequest("GET", url, nil)

	// Set some tags on the clientSpan to annotate that it's the client span. The additional HTTP tags are useful for debugging purposes.
	ext.SpanKindRPCClient.Set(clientSpan)
	ext.HTTPUrl.Set(clientSpan, url)
	ext.HTTPMethod.Set(clientSpan, "GET")

	// Inject the client span context into the headers
	err = tracer.Inject(clientSpan.Context(), opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(req.Header))
	if err != nil {
		log.Printf("inject failed:%s", err.Error())
	}
	resp, _ := http.DefaultClient.Do(req)

	log.Printf("resp:%v", resp)

	clientSpan.Finish()

	time.Sleep(time.Second * 5)

}
