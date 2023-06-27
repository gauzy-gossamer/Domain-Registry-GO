package postgres

import (
    "fmt"
    "context"
    "logger/logging"
    "github.com/jackc/pgx/v5"
    "github.com/jackc/pgx/v5/pgconn"
    "github.com/jackc/pgx/v5/pgxpool"
    "github.com/kpango/glg"
)

func CreatePool(dbconf *DBConfig) (*pgxpool.Pool, error) {
    psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
        "password=%s dbname=%s sslmode=disable",
        dbconf.Host, dbconf.Port, dbconf.User, dbconf.Password, dbconf.DBname)
    config, err := pgxpool.ParseConfig(psqlInfo)
    if  err != nil {
        return nil, err
    }
    config.MaxConns = 24
    ctx := context.Background()
    db, err := pgxpool.NewWithConfig(ctx, config)
    if  err != nil {
        return nil, err
    }
    glg.Trace("created pool")

    return db, nil
}

type DBConn struct {
    Conn *pgxpool.Conn
    /* pool in case we need another connection */
    Pool *pgxpool.Pool
    tx pgx.Tx
    ctx context.Context
    logger logging.Logger
}

/* DBConn object is created per request */
func AcquireConn(pool *pgxpool.Pool, logger logging.Logger) (*DBConn, error) {
    ctx := context.Background()
    db, err := pool.Acquire(ctx)
    if err != nil {
        return nil, err
    }

    return &DBConn{Conn:db, Pool:pool, ctx:ctx, logger:logger}, nil
}

func (d *DBConn) QueryRow(query string, params... any) pgx.Row {
    d.logger.Trace(query, params)
    if d.tx != nil {
        return d.tx.QueryRow(d.ctx, query, params...)
    }
    return d.Conn.QueryRow(d.ctx, query, params...)
}

func (d *DBConn) Query(query string, params... any) (pgx.Rows, error) {
    d.logger.Trace(query, params)
    if d.tx != nil {
        return d.tx.Query(d.ctx, query, params...)
    }
    return d.Conn.Query(d.ctx, query, params...)
}

func (d *DBConn) Exec(query string, params... any) (pgconn.CommandTag, error) {
    d.logger.Trace(query, params)
    if d.tx != nil {
        return d.tx.Exec(d.ctx, query, params...)
    }
    return d.Conn.Exec(d.ctx, query, params...)
}

func (d *DBConn) Begin() error {
    d.logger.Trace("begin transaction")
    var err error
    d.tx, err = d.Conn.Begin(d.ctx)
    return err
}

func (d *DBConn) Commit() error {
    d.logger.Trace("commit transaction")
    err := d.tx.Commit(d.ctx)
    d.tx = nil
    return err
}

func (d *DBConn) Rollback() {
    if d.tx != nil {
        d.logger.Trace("rollback transaction")
        err := d.tx.Rollback(d.ctx)
        if err != nil {
            glg.Error(err)
        }
        d.tx = nil
    }
}

func (d *DBConn) Close() {
    d.logger.Trace("released connection")
    d.Conn.Release()
}
