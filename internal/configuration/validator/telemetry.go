package validator

import (
	"fmt"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
)

// ValidateTelemetry validates the telemetry configuration.
func ValidateTelemetry(config *schema.Configuration, validator *schema.StructValidator) {
	validateMetrics(&config.Telemetry.Metrics, validator)
	validateTraces(&config.Telemetry.Traces, validator)
	validateLogs(&config.Telemetry.Logs, validator)
}

func validateMetrics(config *schema.TelemetryMetrics, validator *schema.StructValidator) {
	if config.Address == nil {
		config.Address = schema.DefaultTelemetryConfig.Metrics.Address
	}

	if err := config.Address.ValidateHTTP(); err != nil {
		validator.Push(fmt.Errorf(errFmtTelemetryMetricsAddress, config.Address.String(), err))
	}

	if config.Address.Port() == 0 {
		config.Address.SetPort(schema.DefaultTelemetryConfig.Metrics.Address.Port())
	}

	if config.Address.RouterPath() == "" {
		config.Address.SetPath(schema.DefaultTelemetryConfig.Metrics.Address.Path())
	}

	if config.Buffers.Read <= 0 {
		config.Buffers.Read = schema.DefaultTelemetryConfig.Metrics.Buffers.Read
	}

	if config.Buffers.Write <= 0 {
		config.Buffers.Write = schema.DefaultTelemetryConfig.Metrics.Buffers.Write
	}

	if config.Timeouts.Read <= 0 {
		config.Timeouts.Read = schema.DefaultTelemetryConfig.Metrics.Timeouts.Read
	}

	if config.Timeouts.Write <= 0 {
		config.Timeouts.Write = schema.DefaultTelemetryConfig.Metrics.Timeouts.Write
	}

	if config.Timeouts.Idle <= 0 {
		config.Timeouts.Idle = schema.DefaultTelemetryConfig.Metrics.Timeouts.Idle
	}
}

func validateTraces(config *schema.TelemetryTraces, validator *schema.StructValidator) {
	if config.Exporter.Grpc == nil && config.Exporter.Http == nil {
		validator.Push(fmt.Errorf(errStrTelemetryExporterRequired, "traces"))
		return
	}
	if config.Exporter.Grpc != nil && config.Exporter.Http != nil {
		validator.Push(fmt.Errorf(errStrTelemetryExporterMutuallyExclusive, "traces"))
		return
	}

	if config.Exporter.Grpc != nil {
		validateTraceExporterGrpc(config.Exporter.Grpc, validator)
	}
	if config.Exporter.Http != nil {
		validateTraceExporterHttp(config.Exporter.Http, validator)
	}
}

func validateTraceExporterGrpc(config *schema.TelemetryExporterGrpc, validator *schema.StructValidator) {
	if config.Endpoint == nil {
		config.Endpoint = schema.DefaultTelemetryExporterGrpc.Endpoint
	}
}

func validateTraceExporterHttp(config *schema.TelemetryExporterHttp, validator *schema.StructValidator) {
	if config.Endpoint == nil {
		config.Endpoint = schema.DefaultTelemetryExporterHttp.Endpoint
	}

	if config.Path == "" {
		config.Path = schema.DefaultTelemetryExporterHttp.Path
	}
}

func validateLogs(config *schema.TelemetryLogs, validator *schema.StructValidator) {
	if config.Exporter.Grpc == nil && config.Exporter.Http == nil {
		validator.Push(fmt.Errorf(errStrTelemetryExporterRequired, "logs"))
		return
	}
	if (config.Exporter.Grpc != nil && config.Exporter.Http != nil) {
		validator.Push(fmt.Errorf(errStrTelemetryExporterMutuallyExclusive, "logs"))
		return
	}

	if config.Exporter.Grpc != nil {
		validateLogExporterGrpc(config.Exporter.Grpc, validator)
	}
	if config.Exporter.Http != nil {
		validateLogExporterHttp(config.Exporter.Http, validator)
	}
}

func validateLogExporterGrpc(config *schema.TelemetryExporterGrpc, validator *schema.StructValidator) {
	if config.Endpoint == nil {
		config.Endpoint = schema.DefaultTelemetryExporterGrpc.Endpoint
	}
}

func validateLogExporterHttp(config *schema.TelemetryExporterHttp, validator *schema.StructValidator) {
	if config.Endpoint == nil {
		config.Endpoint = schema.DefaultTelemetryExporterHttp.Endpoint
	}

	if config.Path == "" {
		config.Path = schema.DefaultTelemetryExporterHttp.Path
	}
}
