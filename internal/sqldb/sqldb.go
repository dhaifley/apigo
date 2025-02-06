// Package sqldb provides functionality for working with SQL databases.
package sqldb

import (
	"context"
	"fmt"
	"net"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/dhaifley/apid/internal/config"
	"github.com/dhaifley/apid/internal/errors"
	"github.com/dhaifley/apid/internal/logger"
	"github.com/dhaifley/apid/internal/metric"
	"github.com/dhaifley/apid/internal/request"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

const (
	// DefaultMaxConnections is the default max allowed open connections used
	// by the database connection pool.
	DefaultMaxConnections = 20

	// MaxConnectionLimit is highest possible  max allowed open connections used
	// by the database connection pool. Attempting to configure a higher value
	// will result in a warning message in the log.
	MaxConnectionsLimit = 40
)

// SQLResult values are used to represent sql result values.
type SQLResult interface {
	RowsAffected() int64
}

// sqlResult values are used to represent sql.Result values with telemetry
// recording functionality.
type sqlResult struct {
	res    pgconn.CommandTag
	finish func(err error)
	tx     *SQLTrans
}

// RowsAffected returns the number of rows affected by the SQL operation.
func (sr *sqlResult) RowsAffected() int64 {
	n := sr.res.RowsAffected()

	ctx := context.Background()

	if sr.tx != nil {
		if err := sr.tx.CloseTx(ctx, nil); err != nil {
			sr.tx.sc.log.Log(context.Background(), logger.LvlError,
				"unable to close database transaction",
				"error", err)
		}
	}

	if sr.finish != nil {
		sr.finish(nil)
	}

	return n
}

// SQLRow types represent single SQL query result rows.
type SQLRow interface {
	Scan(dest ...any) error
}

// sqlRow values are used to represent SQLRow values with telemetry
// recording functionality.
type sqlRow struct {
	row    SQLRow
	finish func(err error)
	err    error
	tx     *SQLTrans
}

// Scan reads the values in a SQL row into variables.
func (sr *sqlRow) Scan(dest ...any) error {
	var err error

	if sr.err != nil {
		err = sr.err
	} else {
		err = sr.row.Scan(dest...)
	}

	ctx := context.Background()

	if err != nil {
		if sr.tx != nil {
			if err := sr.tx.CloseTx(ctx, err); err != nil {
				sr.tx.sc.log.Log(context.Background(), logger.LvlError,
					"unable to rollback database transaction",
					"error", err)
			}
		}

		if sr.finish != nil {
			sr.finish(err)
		}

		if errors.ErrorHas(sr.err, `"app.account_id"`) {
			err = errors.Wrap(err, errors.ErrForbidden,
				"unable to access database: missing account_id")
		}
	} else {
		if sr.tx != nil {
			if err := sr.tx.CloseTx(ctx, nil); err != nil {
				sr.tx.sc.log.Log(context.Background(), logger.LvlError,
					"unable to close database transaction",
					"error", err)
			}
		}

		if sr.finish != nil {
			sr.finish(nil)
		}
	}

	return err
}

// SQLRows types represent cursors iterating over SQL query results.
type SQLRows interface {
	Close()
	Next() bool
	Scan(dest ...any) error
	Err() error
}

// sqlRows values are used to represent SQLRows values with telemetry
// recording functionality.
type sqlRows struct {
	rows   SQLRows
	finish func(err error)
	err    error
	tx     *SQLTrans
}

// Close closes the SQL cursor.
func (sr *sqlRows) Close() {
	sr.rows.Close()

	ctx := context.Background()

	if err := sr.err; err != nil {
		if sr.tx != nil {
			if err := sr.tx.CloseTx(ctx, err); err != nil {
				sr.tx.sc.log.Log(context.Background(), logger.LvlError,
					"unable to rollback database transaction",
					"error", err)
			}
		}

		if sr.finish != nil {
			sr.finish(err)
		}
	} else {
		if sr.tx != nil {
			if err := sr.tx.CloseTx(ctx, nil); err != nil {
				sr.tx.sc.log.Log(context.Background(), logger.LvlError,
					"unable to close database transaction",
					"error", err)
			}
		}

		if sr.finish != nil {
			sr.finish(nil)
		}
	}
}

// Next advances the SQL cursor row.
func (sr *sqlRows) Next() bool {
	return sr.rows.Next()
}

// Scan reads the values at the current SQL cursor row into variables.
func (sr *sqlRows) Scan(dest ...any) error {
	err := sr.rows.Scan(dest...)
	if err != nil {
		sr.err = err
	}

	return err
}

// Err gets the error returned by the last SQL cursor operation.
func (sr *sqlRows) Err() error {
	err := sr.rows.Err()
	if err != nil {
		sr.err = err
	}

	return err
}

// SQLTX values represent individual SQL transactions.
type SQLTX interface {
	Commit(ctx context.Context) error
	Exec(ctx context.Context,
		query string, args ...any) (SQLResult, error)
	Query(ctx context.Context,
		query string, args ...any) (SQLRows, error)
	QueryRow(ctx context.Context,
		query string, args ...any) SQLRow
	Rollback(ctx context.Context) error
	CloseTx(ctx context.Context, err error) error
}

// SQLTrans values implement the SQLTX interface.
type SQLTrans struct {
	tx     pgx.Tx
	sc     *SQLConn
	finish func(err error)
}

// Commit completes a sql transaction.
func (tx *SQLTrans) Commit(ctx context.Context) error {
	if err := tx.tx.Commit(ctx); err != nil {
		return errors.Wrap(err, errors.ErrDatabase,
			"unable to commit transaction")
	}

	return nil
}

// setAccount is the command used to set the account ID in the database.
func setAccount(accountID string) string {
	return "SET app.account_id = '" + accountID + "'"
}

// Exec abstracts the sql database driver exec context function.
func (tx *SQLTrans) Exec(ctx context.Context,
	query string, args ...any,
) (SQLResult, error) {
	ctx, finish := tx.sc.startDBSpan(ctx, "exec", query)

	var opErr *net.OpError

	if accountID, err := request.ContextAccountID(ctx); err == nil {
		if _, err := tx.tx.Exec(ctx, setAccount(accountID)); err != nil {
			if errors.As(err, &opErr) {
				if e := tx.sc.Reconnect(ctx); e != nil {
					finish(err)

					return nil, err
				}

				if _, err := tx.tx.Exec(ctx,
					setAccount(accountID)); err != nil {
					finish(err)

					return nil, errors.Wrap(err, errors.ErrDatabase,
						"unable to set account for statement execute")
				}
			} else {
				return nil, err
			}
		}
	}

	res, err := tx.tx.Exec(ctx, query, args...)
	if err != nil && errors.As(err, &opErr) {
		if e := tx.sc.Reconnect(ctx); e != nil {
			finish(err)

			return nil, err
		}

		res, err = tx.tx.Exec(ctx, query, args...)
	}

	if err != nil {
		finish(err)

		if errors.ErrorHas(err, `"app.account_id"`) {
			err = errors.Wrap(err, errors.ErrForbidden,
				"unable to access database: missing account_id")
		}

		return nil, errors.Wrap(err, errors.ErrDatabase,
			"unable to execute statement")
	}

	return &sqlResult{
		res:    res,
		finish: finish,
	}, nil
}

// Query abstracts the sql database driver query context function.
func (tx *SQLTrans) Query(ctx context.Context,
	query string, args ...any,
) (SQLRows, error) {
	ctx, finish := tx.sc.startDBSpan(ctx, "query", query)

	var opErr *net.OpError

	if accountID, err := request.ContextAccountID(ctx); err == nil {
		if _, err := tx.tx.Exec(ctx, setAccount(accountID)); err != nil {
			if errors.As(err, &opErr) {
				if e := tx.sc.Reconnect(ctx); e != nil {
					finish(err)

					return nil, err
				}

				if _, err := tx.tx.Exec(ctx,
					setAccount(accountID)); err != nil {
					finish(err)

					return nil, errors.Wrap(err, errors.ErrDatabase,
						"unable to set account for query")
				}
			} else {
				return nil, err
			}
		}
	}

	rows, err := tx.tx.Query(ctx, query, args...)
	if err != nil && errors.As(err, &opErr) {
		if e := tx.sc.Reconnect(ctx); e != nil {
			finish(err)

			return nil, err
		}

		rows, err = tx.tx.Query(ctx, query, args...)
	}

	if err != nil {
		finish(err)

		if errors.ErrorHas(err, `"app.account_id"`) {
			err = errors.Wrap(err, errors.ErrForbidden,
				"unable to access database: missing account_id")
		}

		return nil, errors.Wrap(err, errors.ErrDatabase,
			"unable to perform query")
	}

	return &sqlRows{
		rows:   rows,
		finish: finish,
	}, nil
}

// QueryRow abstracts the sql database driver query row context function.
func (tx *SQLTrans) QueryRow(ctx context.Context,
	query string, args ...any,
) SQLRow {
	ctx, finish := tx.sc.startDBSpan(ctx, "query_row", query)

	var opErr *net.OpError

	if accountID, err := request.ContextAccountID(ctx); err == nil {
		if _, err := tx.tx.Exec(ctx, setAccount(accountID)); err != nil {
			if errors.As(err, &opErr) {
				if e := tx.sc.Reconnect(ctx); e != nil {
					return &sqlRow{
						err:    err,
						finish: finish,
					}
				}

				if _, err := tx.tx.Exec(ctx,
					setAccount(accountID)); err != nil {
					return &sqlRow{
						err: errors.Wrap(err, errors.ErrDatabase,
							"unable to set account for query row"),
						finish: finish,
					}
				}
			} else {
				return &sqlRow{
					err: errors.Wrap(err, errors.ErrDatabase,
						"unable to set account for query row"),
					finish: finish,
				}
			}
		}
	}

	return &sqlRow{
		row:    tx.tx.QueryRow(ctx, query, args...),
		finish: finish,
	}
}

// Rollback cancels and reverses a sql transaction.
func (tx *SQLTrans) Rollback(ctx context.Context) error {
	if err := tx.tx.Rollback(ctx); err != nil {
		return errors.Wrap(err, errors.ErrDatabase,
			"unable to rollback transaction")
	}

	return nil
}

// CloseTx commits or rolls back a transaction based on the provided
// error and panic state. It is intended to be called with defer after
// beginning a transaction.
func (tx *SQLTrans) CloseTx(ctx context.Context, err error) error {
	var res error
	if err != nil {
		res = tx.Rollback(ctx)
	} else {
		res = tx.Commit(ctx)
	}

	if tx.finish != nil {
		tx.finish(err)
	}

	return res
}

// SQLDB types represent SQL database connection pools.
type SQLDB interface {
	BeginTx(ctx context.Context, opts pgx.TxOptions) (SQLTX, error)
	Exec(ctx context.Context,
		query string, args ...any) (SQLResult, error)
	Query(ctx context.Context,
		query string, args ...any) (SQLRows, error)
	QueryRow(ctx context.Context,
		query string, args ...any) SQLRow
	Close()
	Ping(ctx context.Context) error
	Stat() *pgxpool.Stat
}

// SQLConn values implement the SQLDB interface.
type SQLConn struct {
	*sync.RWMutex
	cfg    *config.Config
	db     *pgxpool.Pool
	log    logger.Logger
	metric metric.Recorder
	tracer trace.Tracer
	cancel context.CancelFunc
	inst   string
	user   string
	svc    string
	mode   int
}

// NewSQLConn initializes and returns a new sql connection pool.
func NewSQLConn(cfg *config.Config,
	log logger.Logger,
	metric metric.Recorder,
	tracer trace.Tracer,
) *SQLConn {
	if log == nil || (reflect.ValueOf(log).Kind() == reflect.Ptr &&
		reflect.ValueOf(log).IsNil()) {
		log = logger.NullLog
	}

	if metric == nil || (reflect.ValueOf(metric).Kind() == reflect.Ptr &&
		reflect.ValueOf(metric).IsNil()) {
		metric = nil
	}

	if tracer == nil || (reflect.ValueOf(tracer).Kind() == reflect.Ptr &&
		reflect.ValueOf(tracer).IsNil()) {
		tracer = nil
	}

	return &SQLConn{
		RWMutex: &sync.RWMutex{},
		cfg:     cfg,
		svc:     cfg.ServiceName(),
		log:     log,
		metric:  metric,
		tracer:  tracer,
	}
}

// DB is used to access the internal database connection pool.
func (sc *SQLConn) DB() *pgxpool.Pool {
	sc.RLock()
	defer sc.RUnlock()

	return sc.db
}

// Svc is used to access the name of the service using this connection pool.
func (sc *SQLConn) Svc() string {
	sc.RLock()
	defer sc.RUnlock()

	return sc.svc
}

// Log gets the logger for the connection pool.
func (sc *SQLConn) Log() logger.Logger {
	sc.RLock()
	defer sc.RUnlock()

	return sc.log
}

// Metric gets the metric data recorder for the connection pool.
func (sc *SQLConn) Metric() metric.Recorder {
	sc.RLock()
	defer sc.RUnlock()

	return sc.metric
}

// Tracer gets the trace data recorder for the connection pool.
func (sc *SQLConn) Tracer() trace.Tracer {
	sc.RLock()
	defer sc.RUnlock()

	return sc.tracer
}

// Inst gets the name of the instance for the connection pool.
func (sc *SQLConn) Inst() string {
	sc.RLock()
	defer sc.RUnlock()

	return sc.inst
}

// User gets the name of the user for the connection pool.
func (sc *SQLConn) User() string {
	sc.RLock()
	defer sc.RUnlock()

	return sc.user
}

// Mode indicates whether the connection pool is setup to apply database
// migrations or perform database initialization.
func (sc *SQLConn) Mode() int {
	sc.RLock()
	defer sc.RUnlock()

	return sc.mode
}

// SetMode sets whether the connection pool is setup to apply database
// migrations.
func (sc *SQLConn) SetMode(mode int) {
	sc.Lock()
	defer sc.Unlock()

	sc.mode = mode
}

// LogErrorf records annotated error messages to the log.
func (sc *SQLConn) LogErrorf(ctx context.Context, op string, err error,
	format string, args ...any,
) {
	msg := fmt.Sprintf(format, args...)

	if err == nil {
		sc.Log().Log(ctx, logger.LvlError, msg)

		return
	}

	e, ok := err.(*errors.Error)
	if !ok {
		e = errors.Wrap(err, errors.ErrDatabase, err.Error())
	}

	sc.Log().Log(ctx, logger.LvlError, msg,
		"error", e,
		"service", sc.Svc(),
		"operation", op)
}

// LogWarnf records annotated warning messages to the log.
func (sc *SQLConn) LogWarnf(ctx context.Context, op, format string,
	args ...any,
) {
	msg := fmt.Sprintf(format, args...)

	sc.Log().Log(ctx, logger.LvlWarn, msg,
		"service", sc.Svc(),
		"operation", op)
}

// LogInfof records annotated information messages to the log.
func (sc *SQLConn) LogInfof(ctx context.Context, op, format string,
	args ...any,
) {
	msg := fmt.Sprintf(format, args...)

	sc.Log().Log(ctx, logger.LvlInfo, msg,
		"service", sc.Svc(),
		"operation", op)
}

// LogDebugf records annotated debug messages to the log.
func (sc *SQLConn) LogDebugf(ctx context.Context, op, format string,
	args ...any,
) {
	msg := fmt.Sprintf(format, args...)

	sc.Log().Log(ctx, logger.LvlDebug, msg,
		"service", sc.Svc(),
		"operation", op)
}

// Connect establishes the connection between the pool and the SQL database.
func (sc *SQLConn) Connect(ctx context.Context) error {
	sc.Lock()
	defer sc.Unlock()

	// Create the database connection pool.
	var err error

	conn := sc.cfg.DBConn(sc.mode)

	sc.db, err = pgxpool.New(ctx, conn)
	if err != nil {
		return errors.Wrap(err, errors.ErrDatabase,
			"unable to open database",
			"service", sc.svc)
	}

	// Extract the user and instance information from the connection string.
	dsi := strings.Index(conn, "//")
	ai := strings.Index(conn, "@")

	if dsi != -1 && ai != -1 {
		sc.user = conn[dsi+2 : ai]
	}

	sli := strings.LastIndex(conn, "/")
	if sli != -1 && sli != dsi+1 {
		i := strings.Index(conn, "?")
		if i == -1 {
			i = len(conn)
		}

		sc.inst = conn[sli+1 : i]
	}

	return nil
}

// Test checks the connectivity of the database connection.
func (sc *SQLConn) Test() error {
	if sc.DB() == nil {
		return errors.New(errors.ErrDatabase,
			"database connection pool is not started")
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	ech := make(chan error, 1)

	go func() {
		var i int64

		if err := sc.QueryRow(ctx, "SELECT 1").Scan(&i); err != nil {
			ech <- err
		}

		close(ech)
	}()

	select {
	case <-ctx.Done():
		return errors.New(errors.ErrDatabase,
			"timeout connecting to database",
			"service", sc.Svc())
	case err := <-ech:
		if err != nil {
			return errors.Wrap(err, errors.ErrDatabase,
				"unable to connect to database",
				"service", sc.Svc())
		}
	}

	return nil
}

// Reconnect tests the database connection and attempts to reconnect if
// the connection is not functional.
func (sc *SQLConn) Reconnect(ctx context.Context) error {
	db := sc.DB()

	if db != nil {
		if err := sc.Test(); err != nil {
			sc.LogWarnf(ctx, "TestReconnect", "unable to connect to "+
				"database, attempting to reestablish connection")
		}

		db.Close()

		if err := sc.Connect(ctx); err != nil {
			sc.LogErrorf(ctx, "TestReconnect", err,
				"unable to reestablish database connection")

			return err
		}

		if err := sc.Test(); err != nil {
			sc.LogErrorf(ctx, "TestReconnect", err,
				"unable to reestablish database connection")

			return err
		}
	}

	return nil
}

// Monitor periodically checks the status of database connection pools.
func (sc *SQLConn) Monitor(ctx context.Context) {
	sc.Lock()
	ctx, sc.cancel = context.WithCancel(ctx)
	sc.Unlock()

	sc.LogDebugf(ctx, "Monitor", "monitoring database connection pool")

	mon := time.Minute
	if sc.cfg.DBMonitor() != 0 {
		mon = sc.cfg.DBMonitor()
	}

	go func() {
		tick := time.NewTimer(mon)

		for {
			select {
			case <-ctx.Done():
				return
			case <-tick.C:
				db := sc.DB()

				if db != nil {
					if err := sc.Test(); err != nil {
						sc.LogWarnf(ctx, "Monitor", "unable to connect to "+
							"database, attempting to reestablish connection")

						db.Close()

						if err = sc.Connect(ctx); err != nil {
							sc.LogErrorf(ctx, "Monitor", err,
								"unable to reestablish database connection")
						}

						if err := sc.Test(); err != nil {
							sc.LogErrorf(ctx, "Monitor", err,
								"unable to reestablish database connection")
						}
					}
				}

				tick = time.NewTimer(mon)
			}
		}
	}()
}

// BeginTx starts a sql transaction.
func (sc *SQLConn) BeginTx(ctx context.Context,
	opts pgx.TxOptions,
) (SQLTX, error) {
	if sc.DB() == nil {
		return nil, errors.New(errors.ErrDatabase,
			"database connection pool is not started")
	}

	tx, err := sc.DB().BeginTx(ctx, opts)
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrDatabase,
			"unable to begin transaction")
	}

	newTx := &SQLTrans{
		tx: tx,
		sc: sc,
	}

	_, newTx.finish = sc.startDBSpan(ctx, "transaction", "")

	return newTx, nil
}

// Exec executes the provided SQL query returning a result value.
func (sc *SQLConn) ExecNoTx(ctx context.Context,
	query string, args ...any,
) (SQLResult, error) {
	db := sc.DB()

	if db == nil {
		return nil, errors.New(errors.ErrDatabase,
			"database connection pool is not started")
	}

	r, err := db.Exec(ctx, query, args...)
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrDatabase,
			"unable to execute statement")
	}

	return r, nil
}

// Exec executes the provided SQL query returning a result value.
func (sc *SQLConn) Exec(ctx context.Context,
	query string, args ...any,
) (SQLResult, error) {
	if sc.DB() == nil {
		return nil, errors.New(errors.ErrDatabase,
			"database connection pool is not started")
	}

	tx, err := sc.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		sc.log.Log(ctx, logger.LvlError,
			"unable to begin database transaction",
			"error", err,
			"query", query,
			"args", args)

		return nil, err
	}

	r, err := tx.Exec(ctx, query, args...)
	if err != nil {
		if err := tx.CloseTx(ctx, err); err != nil {
			sc.log.Log(ctx, logger.LvlError,
				"unable to rollback database transaction",
				"error", err,
				"query", query,
				"args", args)
		}

		return nil, err
	}

	if rv, ok := r.(*sqlResult); ok {
		if txv, ok := tx.(*SQLTrans); ok {
			rv.tx = txv

			return rv, nil
		}
	}

	return r, nil
}

// Query executes the provided SQL query returning a set of rows.
func (sc *SQLConn) Query(ctx context.Context,
	query string, args ...any,
) (SQLRows, error) {
	if sc.DB() == nil {
		return nil, errors.New(errors.ErrDatabase,
			"database connection pool is not started")
	}

	tx, err := sc.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		sc.log.Log(ctx, logger.LvlError,
			"unable to begin database transaction",
			"error", err,
			"query", query,
			"args", args)

		return nil, err
	}

	r, err := tx.Query(ctx, query, args...)
	if err != nil {
		if err := tx.CloseTx(ctx, err); err != nil {
			sc.log.Log(ctx, logger.LvlError,
				"unable to rollback database transaction",
				"error", err,
				"query", query,
				"args", args)
		}

		return nil, err
	}

	if rv, ok := r.(*sqlRows); ok {
		if txv, ok := tx.(*SQLTrans); ok {
			rv.tx = txv

			return rv, nil
		}
	}

	return r, nil
}

// QueryRow executes the provided SQL query returning a single row.
func (sc *SQLConn) QueryRow(ctx context.Context,
	query string, args ...any,
) SQLRow {
	if sc.DB() == nil {
		return &sqlRow{
			err: errors.New(errors.ErrDatabase,
				"database connection pool is not started"),
		}
	}

	tx, err := sc.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		sc.log.Log(ctx, logger.LvlError,
			"unable to begin database transaction",
			"error", err,
			"query", query,
			"args", args)

		return &sqlRow{
			err: err,
		}
	}

	r := tx.QueryRow(ctx, query, args...)

	if rv, ok := r.(*sqlRow); ok {
		if txv, ok := tx.(*SQLTrans); ok {
			rv.tx = txv

			return rv
		}
	}

	return r
}

// Close shuts down the database connection.
func (sc *SQLConn) Close() {
	sc.Lock()
	defer sc.Unlock()

	if sc.cancel != nil {
		sc.cancel()
	}

	if sc.db == nil {
		return
	}

	sc.db.Close()

	sc.db = nil
}

// Ping verifies that the database connection is functional.
func (sc *SQLConn) Ping(ctx context.Context) error {
	db := sc.DB()

	if db == nil {
		return errors.New(errors.ErrDatabase,
			"database connection pool is not started")
	}

	return db.Ping(ctx)
}

// Stat retrieves information about the database connection pool.
func (sc *SQLConn) Stat() *pgxpool.Stat {
	_, finish := sc.startDBSpan(context.Background(), "Stat", "")

	defer finish(nil)

	db := sc.DB()

	if db == nil {
		return &pgxpool.Stat{}
	}

	return db.Stat()
}

// startDBSpan starts a database tracing span. It returns an updated context,
// and a span closing function.
func (sc *SQLConn) startDBSpan(ctx context.Context,
	name, query string,
) (context.Context, func(err error)) {
	sc.RLock()

	inst := sc.inst
	user := sc.user
	tracer := sc.tracer
	mr := sc.metric

	sc.RUnlock()

	start := time.Now()

	const operationTag = "operation:"

	if tracer == nil {
		return ctx, func(err error) {
			if mr != nil {
				if err != nil {
					mr.Increment(ctx, "db_errors", operationTag+name)
				}

				mr.RecordDuration(ctx, "db_latency", time.Since(start),
					operationTag+name)
			}
		}
	}

	ctx, span := tracer.Start(ctx, name,
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(
			attribute.String("db.system", "postgresql"),
			attribute.String("db.type", "sql"),
			attribute.String("db.instance", inst),
			attribute.String("db.user", user),
		),
	)

	if query != "" {
		span.SetAttributes(attribute.String("db.statement", query))
	}

	return ctx, func(err error) {
		if err != nil && reflect.ValueOf(err).Kind() == reflect.Ptr &&
			reflect.ValueOf(err).IsNil() {
			err = nil
		}

		if span != nil {
			if err != nil {
				span.SetStatus(codes.Error, name+" failed")
				span.RecordError(err)
			}

			span.End()
		}

		if mr != nil {
			if err != nil {
				mr.Increment(ctx, "db_errors", operationTag+name)
			}

			mr.RecordDuration(ctx, "db_latency", time.Since(start),
				operationTag+name)
		}
	}
}
