package storage

import (
	"crypto/x509"
	"fmt"
	"strings"
	"time"

	"github.com/go-sql-driver/mysql"
	"github.com/XSAM/otelsql"
	semconv "go.opentelemetry.io/otel/semconv/v1.40.0"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/utils"
)

// MySQLProvider is a MySQL provider.
type MySQLProvider struct {
	SQLProvider
}

// NewMySQLProvider a MySQL provider.
func NewMySQLProvider(config *schema.Configuration, caCertPool *x509.CertPool) (provider *MySQLProvider, err error) {
	var sqlProvider SQLProvider

	if sqlProvider, err = NewSQLProvider(config, providerMySQL, providerMySQL, dsnMySQL(config.Storage.MySQL, caCertPool), otelOptionsMySQL(config.Storage.MySQL)...); err != nil {
		return nil, err
	}

	provider = &MySQLProvider{
		SQLProvider: sqlProvider,
	}

	// All providers have differing SELECT existing table statements.
	provider.sqlSelectExistingTables = queryMySQLSelectExistingTables

	provider.sqlFmtRenameTable = queryFmtMySQLRenameTable

	return provider, nil
}

func otelOptionsMySQL(config *schema.StorageMySQL) ([]otelsql.Option) {
	opts := []otelsql.Option{
		otelsql.WithAttributes(
			semconv.DBSystemNameMySQL,
			semconv.DBNamespace(config.Database),
		),
	}
	if (config.Address.IsUnixDomainSocket()) {
		opts = append(
			opts,
			otelsql.WithAttributes(semconv.ServerAddress(config.Address.NetworkAddress())),
		)
	} else {
		opts = append(
			opts,
			otelsql.WithAttributes(semconv.ServerAddress(config.Address.SocketHostname()), semconv.ServerPort(int(config.Address.Port()))),
		)
	}
	return opts
}

func dsnMySQL(config *schema.StorageMySQL, caCertPool *x509.CertPool) (dataSourceName string) {
	dsnConfig := mysql.NewConfig()

	dsnConfig.Net = config.Address.Network()
	dsnConfig.Addr = config.Address.NetworkAddress()

	if config.TLS != nil {
		dsnConfig.TLSConfig = fmt.Sprintf("authelia-%s-storage", utils.Version())

		_ = mysql.RegisterTLSConfig(dsnConfig.TLSConfig, utils.NewTLSConfig(config.TLS, caCertPool))
	}

	dsnConfig.DBName = config.Database
	dsnConfig.User = config.Username
	dsnConfig.Passwd = config.Password
	dsnConfig.Timeout = config.Timeout
	dsnConfig.MultiStatements = true
	dsnConfig.ParseTime = true
	dsnConfig.RejectReadOnly = true
	dsnConfig.Loc = time.Local
	dsnConfig.Collation = "utf8mb4_unicode_520_ci"
	dsnConfig.ConnectionAttributes = fmt.Sprintf("program_name:authelia,program_version:%s", strings.ReplaceAll(utils.Version(), ",", ""))

	return dsnConfig.FormatDSN()
}
