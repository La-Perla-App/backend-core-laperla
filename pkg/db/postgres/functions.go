package dbpq

import (
	"context"
	"database/sql"
	"strings"
	"sync"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/La-Perla-App/backend-core-laperla/pkg/config"
	"github.com/La-Perla-App/backend-core-laperla/pkg/db/postgres/driver"
	"github.com/jmoiron/sqlx"
)

var (
	databaseInstances     = map[string]*sql.DB{}
	databaseIntancesMutex = sync.Mutex{}
)

var DefaultConnectionString = config.GetString("database.connectionString")

// ConnectToNewSQLInstance create a handled database instance, returned from a connection pool.
// If connectionString is an empty string, the configmaps default is used instead.
func ConnectToNewSQLInstance(connectionString string) (*sql.DB, error) {
	if connectionString == "" {
		connectionString = strings.TrimSpace(DefaultConnectionString)
	}

	databaseIntancesMutex.Lock()
	if db, ok := databaseInstances[connectionString]; ok && db != nil {
		ctx, cancelfunc := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancelfunc()
		if err := db.PingContext(ctx); err == nil {
			databaseIntancesMutex.Unlock()
			return db, nil
		}
		db.Close()
		delete(databaseInstances, connectionString)
	}
	databaseIntancesMutex.Unlock()

	driver := driver.CoreSQLDriver

	var err error

	// Create connection pool
	db, err := sql.Open(driver, connectionString)
	if err != nil {
		return db, err
	}

	maxIdleConnections := config.GetInt("database.maxIdleConnections")
	maxOpenConnections := config.GetInt("database.maxOpenConnections")
	connMaxIdleTime := config.GetInt("database.connMaxIdleTime")
	connMaxLifetime := config.GetInt("database.connMaxLifetime")

	if maxIdleConnections == 0 {
		maxIdleConnections = ValuesPoolConnection.MaxIdleConnections
	}

	if maxOpenConnections == 0 {
		maxOpenConnections = ValuesPoolConnection.MaxOpenConnections
	}

	if connMaxIdleTime == 0 {
		connMaxIdleTime = ValuesPoolConnection.ConnMaxIdleTime
	}

	if connMaxLifetime == 0 {
		connMaxLifetime = ValuesPoolConnection.ConnMaxLifetime
	}

	// Maximum Idle Connections
	db.SetMaxIdleConns(maxIdleConnections)
	// Maximum Open Connections
	db.SetMaxOpenConns(maxOpenConnections)
	// Idle Connection Timeout
	db.SetConnMaxIdleTime(time.Duration(connMaxIdleTime) * time.Second)
	// Connection Lifetime
	db.SetConnMaxLifetime(time.Duration(connMaxLifetime) * time.Second)

	ctx, cancelfunc := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancelfunc()
	if err := db.PingContext(ctx); err != nil {
		return db, err
	}

	databaseIntancesMutex.Lock()
	databaseInstances[connectionString] = db
	databaseIntancesMutex.Unlock()

	return db, nil
}

func GetSQLXInstance(db *sql.DB) *sqlx.DB {
	return sqlx.NewDb(db, "postgres")
}

func ContextWithNoLogs(ctx context.Context) context.Context {
	return context.WithValue(ctx, "silent", true)
}

func QueryBuilder() sq.StatementBuilderType {
	return sq.StatementBuilder.PlaceholderFormat(sq.Dollar)
}
