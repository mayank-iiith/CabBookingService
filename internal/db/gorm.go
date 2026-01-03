package db

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gorm_logger "gorm.io/gorm/logger"
)

const (
	defaultDatabaseQueryTimeout = 1 * time.Minute
)

// PostgresPoolConfig defines connection pool settings
type PostgresPoolConfig struct {
	PoolMaxConnections    int    `json:"poolMaxConnections"`
	PoolMaxConnLifetime   string `json:"poolMaxConnLifetime"`
	PoolMaxConnIdleTime   string `json:"poolMaxConnIdleTime"`
	PoolHealthCheckPeriod string `json:"poolHealthCheckPeriod"`
	ConnectTimeout        string `json:"connectTimeout"`
}

var (
	defaultPostgresPoolConfiguration = PostgresPoolConfig{
		PoolMaxConnections:    20,
		PoolMaxConnLifetime:   "1h",
		PoolMaxConnIdleTime:   "30m",
		PoolHealthCheckPeriod: "1m",
		ConnectTimeout:        "30",
	}
)

func NewGormDbConn(config PostgresConfig) (*gorm.DB, error) {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=UTC",
		config.Host,
		config.User,
		config.Password,
		config.DBName,
		config.Port,
	)

	// 1. Connect to Database
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}
	log.Println("Database connection established.")

	// 2. Configure Connection Pooling
	db, err = applyPoolConfig(db)
	if err != nil {
		return nil, fmt.Errorf("failed to apply pool config: %w", err)
	}
	log.Println("Database connection established with connection pooling.")

	return db, nil
}

func applyPoolConfig(db *gorm.DB) (*gorm.DB, error) {
	pooledConfig := defaultPostgresPoolConfiguration

	// GORM builds on top of the standard database/sql package.
	// We need to get the underlying *sql.DB object to configure the pool.
	sqlDb, err := db.DB()
	if err != nil {
		log.Println("applyPoolConfig: db.DB() failed:", err)
		//logger.WithError(err).Error("db.DB() failed")
		return nil, err
	}

	poolMaxConnLifeTime, err := time.ParseDuration(pooledConfig.PoolMaxConnLifetime)
	if err != nil {
		log.Println("applyPoolConfig: time.ParseDuration PoolMaxConnLifetime failed:", err)
		//logger.WithError(err).Error("time.ParseDuration PoolMaxConnLifetime failed")
		return nil, err
	}

	poolMaxConnIdleTime, err := time.ParseDuration(pooledConfig.PoolMaxConnIdleTime)
	if err != nil {
		log.Println("applyPoolConfig: time.ParseDuration PoolMaxConnIdleTime failed:", err)
		//logger.WithError(err).Error("time.ParseDuration PoolMaxConnIdleTime failed")
		return nil, err
	}

	// SetMaxIdleConns sets the maximum number of connections in the idle connection pool.
	// If you have too many idle connections, they consume memory on the DB server.
	// If you have too few, new requests have to wait for a new connection to be created (slow).
	sqlDb.SetMaxIdleConns(pooledConfig.PoolMaxConnections)

	// SetMaxOpenConns sets the maximum number of open connections to the database.
	// This prevents your app from overwhelming the database with too many concurrent queries.
	sqlDb.SetMaxOpenConns(pooledConfig.PoolMaxConnections * 2)

	// SetConnMaxIdleTime sets the maximum amount of time a connection may be idle.
	// This helps recycle connections that may have been closed by the DB server.
	sqlDb.SetConnMaxIdleTime(poolMaxConnIdleTime)

	// SetConnMaxLifetime sets the maximum amount of time a connection may be reused.
	// This is important to prevent issues with firewalls or load balancers closing idle connections silently.
	sqlDb.SetConnMaxLifetime(poolMaxConnLifeTime)
	return db, nil
}

// NewGormTx creates a new GORM session with context and logger
func NewGormTx(ctx context.Context, db *gorm.DB) *gorm.DB {
	if db == nil {
		return nil
	}

	// if env is development, enable gorm debug logging
	gormLogger := db.Logger.LogMode(gorm_logger.Warn)
	if isDevelopmentEnv() {
		// Log sparingly in high traffic, but useful for dev
		log.Println("Gorm debug logging is enabled")
		gormLogger = db.Logger.LogMode(gorm_logger.Info)
	}

	// if the context does not have a deadline, set a timeout
	ctx = withDatabaseTimeout(ctx, defaultDatabaseQueryTimeout)

	// create a new session with the context and logger
	return db.Session(&gorm.Session{
		Context: ctx,
		Logger:  gormLogger,
	})
}

// withDatabaseTimeout applies a default timeout only if no deadline exists,
// otherwise preserves the existing deadline (whether shorter or longer).
// This provides a safety net for runaway queries while respecting caller intentions.
// Uses WithDeadline to avoid premature cancellation that would affect row iteration.
func withDatabaseTimeout(ctx context.Context, timeout time.Duration) context.Context {
	// Check if context already has a deadline
	if _, hasDeadline := ctx.Deadline(); hasDeadline {
		// Preserve existing deadline regardless of duration
		return ctx
	}

	// No existing deadline - apply default timeout for safety
	timeoutCtx, _ := context.WithTimeout(ctx, timeout)
	// Note: We don't call cancel() immediately to allow the context to remain valid
	// for operations like row iteration that happen after the query completes.
	// The context will automatically cancel when the deadline is reached.
	return timeoutCtx
}

func isDevelopmentEnv() bool {
	return os.Getenv("APP_ENV") == "development"
}
