package schema

import (
	"net/url"
	"time"
)

// Telemetry represents the telemetry config.
type Telemetry struct {
	Metrics TelemetryMetrics `koanf:"metrics" yaml:"metrics,omitempty" toml:"metrics,omitempty" json:"metrics,omitempty" jsonschema:"title=Metrics" jsonschema_description:"The telemetry metrics server configuration."`
	Traces  TelemetryTraces  `koanf:"traces" yaml:"traces,omitempty" toml:"traces,omitempty" json:"traces,omitempty" jsonschema:"title=Traces" jsonschema_description:"The telemetry traces server configuration."`
	Logs    TelemetryLogs    `koanf:"logs" yaml:"logs,omitempty" toml:"logs,omitempty" json:"logs,omitempty" jsonschema:"title=Logs" jsonschema_description:"The telemetry logs server configuration."`
}

// TelemetryMetrics represents the telemetry metrics config.
type TelemetryMetrics struct {
	Enabled bool        `koanf:"enabled" yaml:"enabled" toml:"enabled" json:"enabled" jsonschema:"default=false,title=Enabled" jsonschema_description:"Enables the metrics server."`
	Address *AddressTCP `koanf:"address" yaml:"address,omitempty" toml:"address,omitempty" json:"address,omitempty" jsonschema:"default=tcp://:9959/,title=Address" jsonschema_description:"The address for the metrics server to listen on."`

	Buffers  ServerBuffers  `koanf:"buffers" yaml:"buffers,omitempty" toml:"buffers,omitempty" json:"buffers,omitempty" jsonschema:"title=Buffers" jsonschema_description:"The server buffers configuration for the metrics server."`
	Timeouts ServerTimeouts `koanf:"timeouts" yaml:"timeouts,omitempty" toml:"timeouts,omitempty" json:"timeouts,omitempty" jsonschema:"title=Timeouts" jsonschema_description:"The server timeouts configuration for the metrics server."`
}

// TelemetryTraces represents the telemetry traces config.
type TelemetryTraces struct {
	Enabled  bool              `koanf:"enabled" yaml:"enabled" toml:"enabled" json:"enabled" jsonschema:"default=false,title=Enabled" jsonschema_description:"Enables the traces server."`
	Exporter TelemetryExporter `koanf:"exporter" yaml:"exporter,omitempty" toml:"exporter,omitempty" json:"exporter,omitempty" jsonschema:"title=Exporter Configuration" jsonschema_description:"The exporter configuration for the traces server."`
}

type TelemetryLogs struct {
	Enabled  bool              `koanf:"enabled" yaml:"enabled" toml:"enabled" json:"enabled" jsonschema:"default=false,title=Enabled" jsonschema_description:"Enables the traces server."`
	Exporter TelemetryExporter `koanf:"exporter" yaml:"exporter,omitempty" toml:"exporter,omitempty" json:"exporter,omitempty" jsonschema:"title=Exporter Configuration" jsonschema_description:"The exporter configuration for the traces server."`
}

type TelemetryExporter struct {
	Grpc *TelemetryExporterGrpc `koanf:"grpc" yaml:"grpc,omitempty" toml:"grpc,omitempty" json:"grpc,omitempty" jsonschema:"title=gRPC Exporter Configuration" jsonschema_description:"The gRPC exporter configuration for the traces server."`
	Http *TelemetryExporterHttp `koanf:"http" yaml:"http,omitempty" toml:"http,omitempty" json:"http,omitempty" jsonschema:"title=HTTP Exporter Configuration" jsonschema_description:"The HTTP exporter configuration for the traces server."`
}

type TelemetryExporterOtel struct {
	Endpoint *url.URL `koanf:"endpoint" yaml:"emdpoint,omitempty" toml:"endpoint,omitempty" json:"endpoint,omitempty" jsonschema:"default=localhost:4317,title=Endpoint" jsonschema_description:"The endpoint for the traces server to send data to."`
	Insecure bool     `koanf:"insecure" yaml:"insecure" toml:"insecure" json:"insecure" jsonschema:"default=false,title=Insecure" jsonschema_description:"If true, the traces server will not use TLS."`
}

type TelemetryExporterGrpc struct {
	TelemetryExporterOtel `koanf:",squash"`
}

type TelemetryExporterHttp struct {
	TelemetryExporterOtel `koanf:",squash"`
	Path     string   `koanf:"path" yaml:"path,omitempty" toml:"path,omitempty" json:"path,omitempty" jsonschema:"default=/v1/traces,title=Path" jsonschema_description:"The path for the HTTP traces exporter."`
}

var DefaultTelemetryExporterGrpc = TelemetryExporterGrpc{
	TelemetryExporterOtel: TelemetryExporterOtel{
		Endpoint: &url.URL{Host: "localhost:4317"},
		Insecure: false,
	},
}

var DefaultTelemetryExporterHttp = TelemetryExporterHttp{
	TelemetryExporterOtel: TelemetryExporterOtel{
		Endpoint: &url.URL{Host: "localhost:4318"},
		Insecure: false,
	},
	Path: "/",
}

var DefaultTelemetryExporter = TelemetryExporter{
	Grpc: nil,
	Http: nil,
}

// DefaultTelemetryConfig is the default telemetry configuration.
var DefaultTelemetryConfig = Telemetry{
	Metrics: TelemetryMetrics{
		Address: &AddressTCP{Address{true, false, -1, 9959, nil, &url.URL{Scheme: AddressSchemeTCP, Host: ":9959", Path: "/metrics"}}},
		Buffers: ServerBuffers{
			Read:  4096,
			Write: 4096,
		},
		Timeouts: ServerTimeouts{
			Read:  time.Second * 6,
			Write: time.Second * 6,
			Idle:  time.Second * 30,
		},
	},
	Traces: TelemetryTraces{
		Exporter: DefaultTelemetryExporter,
	},
	Logs: TelemetryLogs{
		Exporter: DefaultTelemetryExporter,
	},
}
