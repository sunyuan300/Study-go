package config

import (
	"fmt"
	jaegerConfig "github.com/uber/jaeger-client-go/config"
	"io"
)

const (
	CollectorEndpoint  = "应用层://192.168.56.101:14268/api/traces"
	LocalAgentHostPort = "192.168.56.101:6831"
)

func NewTracer(service string) (opentracing.Tracer, io.Closer) {
	return newTracer(service, "")
}

func newTracer(service, collectorEndpoint string) (opentracing.Tracer, io.Closer) {
	// 参数详解 https://www.jaegertracing.io/docs/1.27/sampling/

	cfg := jaegerConfig.Configuration{
		// 服务名称
		ServiceName: service,

		// 采样配置
		Sampler: &jaegerConfig.SamplerConfig{
			Type:  jaeger.SamplerTypeConst,
			Param: 1,
		},

		Reporter: &jaegerConfig.ReporterConfig{
			LogSpans: true,

			// 将span发往jaeger-collector的服务地址
			CollectorEndpoint: CollectorEndpoint,
			//LocalAgentHostPort:LocalAgentHostPort
		},
	}

	// 不传递 logger 就不会打印日志
	tracer, closer, err := cfg.NewTracer(jaegerConfig.Logger(jaeger.StdLogger))
	if err != nil {
		panic(fmt.Sprintf("ERROR: cannot init Jaeger: %v\n", err))
	}
	opentracing.SetGlobalTracer(tracer)
	return tracer, closer
}
