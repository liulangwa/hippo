package jaeger

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/liulangwa/hippo/libraries/constants"
	"github.com/opentracing/opentracing-go"
	jc "github.com/uber/jaeger-client-go"
	jcf "github.com/uber/jaeger-client-go/config"
)

func NewJaegerTracer(serviceName string, addr string) (opentracing.Tracer, io.Closer, error) {
	cfg := &jcf.Configuration{
		Sampler: &jcf.SamplerConfig{
			Type:  "const", //固定采样
			Param: 1,       //1=全采样、0=不采样
		},

		Reporter: &jcf.ReporterConfig{
			LogSpans:           true,
			LocalAgentHostPort: addr,
		},

		ServiceName: serviceName,
	}

	tracer, closer, err := cfg.NewTracer()
	if err != nil {
		return nil, nil, err
	}
	opentracing.SetGlobalTracer(tracer)
	return tracer, closer, nil
}

//GetSpanFromContext 从context提取span
func GetSpanFromContext(ctx context.Context) opentracing.Span {
	spanContext, ok := ctx.Value(constants.ETracingSpan).(context.Context)

	if ok {
		return opentracing.SpanFromContext(spanContext)
	}

	return nil
}

//ContextConvert opentracing context 转换为jaeger context
func ContextConvert(spanContext opentracing.SpanContext) jc.SpanContext {
	if sc, ok := spanContext.(jc.SpanContext); ok {
		return sc
	} else {
		return jc.SpanContext{}
	}
}

func GetTraceID(span opentracing.Span) string {
	ctx := ContextConvert(span.Context())
	return ctx.TraceID().String()
}

func GetSpanID(span opentracing.Span) string {
	ctx := ContextConvert(span.Context())
	return ctx.SpanID().String()
}

func InjectHTTP(ctx context.Context, tracer opentracing.Tracer, header http.Header) error {

	span := GetSpanFromContext(ctx)

	if span == nil {
		return errors.New("inject failed,not span find in context")
	}

	return tracer.Inject(span.Context(), opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(header))

}

type Response struct {
	HTTPCode int
	Response string
}

// Send 发送Jaeger请求,不是源端发起
func Send(ctx context.Context, tracer opentracing.Tracer, method, url string, header map[string]string, body io.Reader, timeout time.Duration) (ret Response, err error) {
	var req *http.Request

	client := &http.Client{
		Transport: http.DefaultTransport,
		Timeout:   timeout,
	}

	//构建req
	req, err = http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return
	}

	//设置请求header
	for k, v := range header {
		req.Header.Add(k, v)
	}

	//注入Jaeger
	err = InjectHTTP(ctx, tracer, req.Header)
	if err != nil {
		log.Printf("inject failed %s", err.Error())
	}

	//发送请求
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}

	ret.HTTPCode = resp.StatusCode
	if resp.StatusCode != http.StatusOK {
		err = errors.New(fmt.Sprintf("http code is %d", resp.StatusCode))
		return
	}

	if b != nil {
		ret.Response = string(b)
	}

	return
}
