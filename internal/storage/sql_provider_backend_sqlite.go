package storage

import (
	"database/sql"
	"encoding/base64"
	"fmt"

	"github.com/mattn/go-sqlite3"
	"github.com/XSAM/otelsql"
	semconv "go.opentelemetry.io/otel/semconv/v1.40.0"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
)

// SQLiteProvider is a SQLite3 provider.
type SQLiteProvider struct {
	SQLProvider
}

// NewSQLiteProvider constructs a SQLite provider.
func NewSQLiteProvider(config *schema.Configuration) (provider *SQLiteProvider, err error) {
	var p SQLProvider

	if p, err = NewSQLProvider(config, providerSQLite, "sqlite3e", fmt.Sprintf(dsnFmtSQLite, config.Storage.Local.Path), otelOptionsSQLite(config.Storage.Local)...); err != nil {
		return nil, err
	}

	provider = &SQLiteProvider{
		SQLProvider: p,
	}

	// All providers have differing SELECT existing table statements.
	provider.sqlSelectExistingTables = querySQLiteSelectExistingTables

	return provider, nil
}

func otelOptionsSQLite(config *schema.StorageLocal) ([]otelsql.Option) {
	return []otelsql.Option{
		otelsql.WithAttributes(
			semconv.DBSystemNameSQLite,
			semconv.DBNamespace(config.Path),
		),
		// Per the implementation of mattn/go-sqlite3, the queries occurs during Rows.Next calls instead of Query or QueryContext calls.
		otelsql.WithSpanOptions(otelsql.SpanOptions{RowsNext: true}),
	}
}

func sqlite3BLOBToTEXTBase64(data []byte) (b64 string) {
	return base64.StdEncoding.EncodeToString(data)
}

func sqlite3TEXTBase64ToBLOB(b64 string) (data []byte, err error) {
	return base64.StdEncoding.DecodeString(b64)
}

func init() {
	sql.Register("sqlite3e", &sqlite3.SQLiteDriver{
		ConnectHook: func(conn *sqlite3.SQLiteConn) (err error) {
			if err = conn.RegisterFunc("BIN2B64", sqlite3BLOBToTEXTBase64, true); err != nil {
				return err
			}

			if err = conn.RegisterFunc("B642BIN", sqlite3TEXTBase64ToBLOB, true); err != nil {
				return err
			}

			return nil
		},
	})
}
