package middleware

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
)

func Jaeger() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var rootSpan opentracing.Span
		// 直接从 ctx.Request.Header 中提取span,如果没有就新建一个
		spCtx,err := opentracing.GlobalTracer().Extract(opentracing.HTTPHeaders,opentracing.HTTPHeadersCarrier(ctx.Request.Header))
		if err != nil {
			rootSpan = opentracing.GlobalTracer().StartSpan(ctx.Request.URL.Path)
			defer rootSpan.Finish()
		} else {
			rootSpan = opentracing.StartSpan(
				ctx.Request.URL.Path,
				opentracing.ChildOf(spCtx),
				opentracing.Tag{Key: string(ext.Component),Value: "HTTP"},
				ext.SpanKindRPCServer,
				)
			defer rootSpan.Finish()
		}

		ctx.Set("tracer",opentracing.GlobalTracer())
		ctx.Set("ctx",opentracing.ContextWithSpan(context.Background(),rootSpan))
		ctx.Next()
	}
}