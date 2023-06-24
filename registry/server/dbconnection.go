package server

import (
    "fmt"
    "context"
    "github.com/jackc/pgx/v5"
    "github.com/jackc/pgx/v5/pgconn"
    "github.com/jackc/pgx/v5/pgxpool"
    "github.com/kpango/glg"
)

func CreatePool(dbconf *DBConfig) (*pgxpool.Pool, error) {
    psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
        "password=%s dbname=%s sslmode=disable",
        dbconf.Host, dbconf.Port, dbconf.User, dbconf.Password, dbconf.DBname)
//   use more specific config ?
//    config, err := pgxpool.ParseConfig(psqlInfo)
//    config.MaxConns = 64
//    db, err := pgxpool.NewWithConfig(ctx, config)
    ctx := context.Background()
    db, err := pgxpool.New(ctx, psqlInfo)
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
    logger Logger
}

/* DBConn object is created per request */
func AcquireConn(pool *pgxpool.Pool, logger Logger) (*DBConn, error) {
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

func (d *DBConn) BeginSerializable() error {
    d.logger.Trace("begin serializable transaction")
    var err error
    d.tx, err = d.Conn.BeginTx(d.ctx, pgx.TxOptions{IsoLevel: pgx.Serializable})
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
        glg.Trace("rollback transaction")
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

func (d *DBConn) RetryTx(fn func() error) {
    tries := 3
    err := fn()
    for {
        if pgErr, ok := err.(*pgconn.PgError); !ok || tries < 0 || pgErr.Code != "40001" {
            break
        }
        glg.Trace("retry transaction")
        err = fn()
        tries -= 1
    }
}
