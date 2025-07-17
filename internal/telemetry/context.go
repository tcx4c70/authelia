package telemetry

import (
	"context"

	"github.com/valyala/fasthttp"
)

type ctxKey int

const ctxKeyCurrentSpanContext ctxKey = iota

func SpanContextFromContext(ctx context.Context) context.Context {
	if ctx == nil {
		return nil
	}
	if spanContext, ok := ctx.Value(ctxKeyCurrentSpanContext).(context.Context); ok {
		return spanContext
	}
	return ctx
}

func WithSpanContext(ctx context.Context, spanCtx context.Context) context.Context {
	if ctx == nil {
		return nil
	}
	if spanCtx == nil {
		return ctx
	}

	switch c := ctx.(type) {
	case *fasthttp.RequestCtx:
		c.SetUserValue(ctxKeyCurrentSpanContext, spanCtx)
		return ctx
	default:
		return context.WithValue(ctx, ctxKeyCurrentSpanContext, spanCtx)
	}
}
