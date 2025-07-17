package telemetry

import (
	"context"

	"go.opentelemetry.io/contrib/bridges/otellogrus"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploggrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	logsdk "go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/resource"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.40.0"
	"go.opentelemetry.io/otel/log/global"
	"go.opentelemetry.io/otel/trace"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/utils"
	"github.com/sirupsen/logrus"
)

const instrumName = "github.com/authelia/authelia"

type Provider struct {
	config         *schema.Telemetry
	tracerProvider *tracesdk.TracerProvider
	logProvider    *logsdk.LoggerProvider
	Tracer         trace.Tracer
}

func NewProvider(config *schema.Telemetry) *Provider {
	return &Provider{
		config: config,
	}
}

func (p *Provider) Start(ctx context.Context) (err error) {
	err = p.initTrace(ctx)
	if err != nil {
		return
	}

	err = p.initLog(ctx)
	if err != nil {
		return
	}

	return
}

func (p *Provider) Stop(ctx context.Context) {
	if p.tracerProvider != nil {
		_ = p.tracerProvider.Shutdown(ctx)
	}
	p.tracerProvider = nil

	if p.logProvider != nil {
		_ = p.logProvider.Shutdown(ctx)
	}
	p.logProvider = nil
}

func (p *Provider) initTrace(ctx context.Context) (err error) {
	var traceExporter tracesdk.SpanExporter

	if p.config.Traces.Exporter.Grpc != nil {
		traceExporter, err = newTraceGrpcExporter(ctx, p.config.Traces.Exporter.Grpc)
	} else {
		traceExporter, err = newTraceHTTPExporter(ctx, p.config.Traces.Exporter.Http)
	}

	if err != nil {
		return
	}

	res, err := resource.New(
		ctx,
		resource.WithTelemetrySDK(),
		resource.WithAttributes(semconv.ServiceName("authelia"), semconv.ServiceVersion(utils.Version())),
	)
	if err != nil {
		return err
	}

	bsp := tracesdk.NewBatchSpanProcessor(traceExporter)
	p.tracerProvider = tracesdk.NewTracerProvider(
		tracesdk.WithSampler(tracesdk.AlwaysSample()),
		tracesdk.WithResource(res),
		tracesdk.WithSpanProcessor(bsp),
	)
	otel.SetTracerProvider(p.tracerProvider)

	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	p.Tracer = p.tracerProvider.Tracer(instrumName)

	return nil
}

func newTraceGrpcExporter(ctx context.Context, config *schema.TelemetryExporterGrpc) (tracesdk.SpanExporter, error) {
	options := []otlptracegrpc.Option{
		otlptracegrpc.WithEndpoint(config.Endpoint.String()),
	}
	if config.Insecure {
		options = append(options, otlptracegrpc.WithInsecure())
	}

	return otlptracegrpc.New(ctx, options...)
}

func newTraceHTTPExporter(ctx context.Context, config *schema.TelemetryExporterHttp) (tracesdk.SpanExporter, error) {
	options := []otlptracehttp.Option{
		otlptracehttp.WithEndpoint(config.Endpoint.String()),
	}
	if config.Insecure {
		options = append(options, otlptracehttp.WithInsecure())
	}
	if config.Path != "" {
		options = append(options, otlptracehttp.WithURLPath(config.Path))
	}

	return otlptracehttp.New(ctx, options...)
}

func (p *Provider) initLog(ctx context.Context) (err error) {
	var logExporter logsdk.Exporter
	if p.config.Logs.Exporter.Grpc != nil {
		logExporter, err = newLogGrpcExporter(ctx, p.config.Logs.Exporter.Grpc)
	} else {
		logExporter, err = newLogHTTPExporter(ctx, p.config.Logs.Exporter.Http)
	}

	if err != nil {
		return
	}

	res, err := resource.New(
		ctx,
		resource.WithTelemetrySDK(),
		resource.WithAttributes(semconv.ServiceName("authelia"), semconv.ServiceVersion(utils.Version())),
	)
	if err != nil {
		return err
	}

	bsp := logsdk.NewBatchProcessor(logExporter)
	p.logProvider = logsdk.NewLoggerProvider(
		logsdk.WithResource(res),
		logsdk.WithProcessor(bsp),
	)
	global.SetLoggerProvider(p.logProvider)

	hook := otellogrus.NewHook("authelia", otellogrus.WithLoggerProvider(p.logProvider))
	logrus.AddHook(hook)

	return nil
}

func newLogGrpcExporter(ctx context.Context, config *schema.TelemetryExporterGrpc) (logsdk.Exporter, error) {
	options := []otlploggrpc.Option{
		otlploggrpc.WithEndpoint(config.Endpoint.String()),
	}
	if config.Insecure {
		options = append(options, otlploggrpc.WithInsecure())
	}
	return otlploggrpc.New(ctx, options...)
}

func newLogHTTPExporter(ctx context.Context, config *schema.TelemetryExporterHttp) (logsdk.Exporter, error) {
	options := []otlploghttp.Option{
		otlploghttp.WithEndpoint(config.Endpoint.String()),
	}
	if config.Insecure {
		options = append(options, otlploghttp.WithInsecure())
	}
	if config.Path != "" {
		options = append(options, otlploghttp.WithURLPath(config.Path))
	}
	return otlploghttp.New(ctx, options...)
}
