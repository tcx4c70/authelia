package middlewares

import (
	"fmt"
	"net"
	"strings"

	"github.com/valyala/fasthttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	semconv "go.opentelemetry.io/otel/semconv/v1.40.0"
	"go.opentelemetry.io/otel/trace"

	"github.com/authelia/authelia/v4/internal/telemetry"
)

var propagator = otel.GetTextMapPropagator()

func NewTracesRequest(provider *telemetry.Provider) (middleware Basic) {
	return func(next fasthttp.RequestHandler) (handler fasthttp.RequestHandler) {
		return func(ctx *fasthttp.RequestCtx) {
			if provider.Tracer == nil {
				next(ctx)
				return
			}

			carrier := fasthttpCarrier{ctx: ctx}
			propagatedCtx := propagator.Extract(ctx, carrier)
			options := []trace.SpanStartOption{
				trace.WithSpanKind(trace.SpanKindServer),
				trace.WithTimestamp(ctx.Time()),
				trace.WithAttributes(getSpanAttribute(ctx)...),
			}

			spanCtx, span := provider.Tracer.Start(propagatedCtx, getSpanName(ctx), options...)
			defer span.End()

			propagator.Inject(spanCtx, carrier)
			_ = telemetry.WithSpanContext(ctx, spanCtx)

			next(ctx)

			statusCode := ctx.Response.StatusCode()
			span.SetStatus(status(statusCode))
			span.SetAttributes(semconv.HTTPResponseStatusCode(statusCode))
		}
	}
}

func getSpanName(ctx *fasthttp.RequestCtx) string {
	// Use the request method and path as the span name.
	method := string(ctx.Method())
	target := string(ctx.Path())

	if method == "" {
		method = "HTTP"
	}

	if target == "" {
		return method
	}

	return method + " " + target
}

func getScheme(ctx *fasthttp.RequestCtx) string {
	if scheme := ctx.Request.Header.Peek(fasthttp.HeaderXForwardedProto); len(scheme) > 0 {
		return string(scheme)
	}
	return string(ctx.URI().Scheme())
}

func getDomain(ctx *fasthttp.RequestCtx) string {
	if host := ctx.Request.Header.Peek(fasthttp.HeaderXForwardedHost); len(host) > 0 {
		return string(host)
	}
	if host := ctx.Request.Header.Peek(fasthttp.HeaderHost); len(host) > 0 {
		return string(host)
	}
	return string(ctx.URI().Host())
}

func getClientIP(ctx *fasthttp.RequestCtx) string {
	if header := ctx.Request.Header.Peek(fasthttp.HeaderXForwardedFor); len(header) > 0 {
		ips := strings.Split(string(header), ",")
		if len(ips) > 0 {
			return strings.TrimSpace(ips[0])
		}
	}
	if remoteAddr, ok := ctx.RemoteAddr().(*net.TCPAddr); ok {
		return remoteAddr.IP.String()
	}
	return ""
}

func getProtocol(proto []byte) (name string, version string) {
	name, version, _ = strings.Cut(string(proto), "/")
	name = strings.ToLower(name)
	return
}

func getSpanAttribute(ctx *fasthttp.RequestCtx) []attribute.KeyValue {
	attrs := []attribute.KeyValue{
		semconv.HTTPRequestMethodKey.String(string(ctx.Method())),
		semconv.URLSchemeKey.String(getScheme(ctx)),
		semconv.URLDomainKey.String(getDomain(ctx)),
		semconv.URLPathKey.String(string(ctx.Path())),
		semconv.HTTPRouteKey.String(string(ctx.Path())),
		semconv.UserAgentNameKey.String(string(ctx.UserAgent())),
		semconv.NetworkLocalAddressKey.String(ctx.LocalIP().String()),
		semconv.NetworkTransportKey.String(ctx.LocalAddr().Network()),
	}

	if remoteAddr, ok := ctx.RemoteAddr().(*net.TCPAddr); ok {
		attrs = append(attrs, semconv.NetworkPeerAddressKey.String(remoteAddr.IP.String()))
		attrs = append(attrs, semconv.NetworkPeerPortKey.Int(remoteAddr.Port))
	}

	if clientIP := getClientIP(ctx); clientIP != "" {
		attrs = append(attrs, semconv.ClientAddressKey.String(clientIP))
	}

	name, version := getProtocol(ctx.Request.Header.Protocol())
	if name != "" {
		attrs = append(attrs, semconv.NetworkProtocolNameKey.String(name))
	}
	if version != "" {
		attrs = append(attrs, semconv.NetworkProtocolVersionKey.String(version))
	}

	return attrs
}

func status(httpStatusCode int) (codes.Code, string) {
	if httpStatusCode < 100 || httpStatusCode >= 600 {
		return codes.Error, fmt.Sprintf("Invalid HTTP status code %d", httpStatusCode)
	}

	if httpStatusCode >= 500 {
		return codes.Error, ""
	}

	return codes.Unset, ""
}

// fasthttpCarrier is a type that adapts fasthttp request to TextMapCarrier.
type fasthttpCarrier struct {
	ctx *fasthttp.RequestCtx
}

// Get returns the value associated with the passed key.
func (c fasthttpCarrier) Get(key string) string {
	return string(c.ctx.Request.Header.Peek(key))
}

// Set stores the key-value pair.
func (c fasthttpCarrier) Set(key string, value string) {
	c.ctx.Request.Header.Set(key, value)
}

// Keys lists the keys stored in this carrier.
func (c fasthttpCarrier) Keys() []string {
	keys := make([]string, 0, c.ctx.Request.Header.Len())
	for key := range c.ctx.Request.Header.All() {
		keys = append(keys, string(key))
	}

	return keys
}
