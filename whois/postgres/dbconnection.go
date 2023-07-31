package postgres

import (
    "fmt"
    "context"
    "github.com/jackc/pgx/v5"
    "github.com/jackc/pgx/v5/pgconn"
    "github.com/jackc/pgx/v5/pgxpool"
    "github.com/kpango/glg"
)

type DBConfig struct {
    Host     string
    Port     int
    User     string
    Password string
    DBname   string
}

func CreatePool(dbconf *DBConfig) (*pgxpool.Pool, error) {
    psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
        "password=%s dbname=%s sslmode=disable",
        dbconf.Host, dbconf.Port, dbconf.User, dbconf.Password, dbconf.DBname)
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
}

func AcquireConn(pool *pgxpool.Pool) (*DBConn, error) {
    ctx := context.Background()
    db, err := pool.Acquire(ctx)
    if err != nil {
        return nil, err
    }

    return &DBConn{Conn:db, Pool:pool, ctx:ctx}, nil
}

func (d *DBConn) QueryRow(query string, params... any) pgx.Row {
    glg.Trace(query, params)
    if d.tx != nil {
        return d.tx.QueryRow(d.ctx, query, params...)
    }
    return d.Conn.QueryRow(d.ctx, query, params...)
}

func (d *DBConn) Query(query string, params... any) (pgx.Rows, error) {
    glg.Trace(query, params)
    if d.tx != nil {
        return d.tx.Query(d.ctx, query, params...)
    }
    return d.Conn.Query(d.ctx, query, params...)
}

func (d *DBConn) Exec(query string, params... any) (pgconn.CommandTag, error) {
    glg.Trace(query, params)
    if d.tx != nil {
        return d.tx.Exec(d.ctx, query, params...)
    }
    return d.Conn.Exec(d.ctx, query, params...)
}

func (d *DBConn) Begin() error {
    glg.Trace("begin transaction")
    var err error
    d.tx, err = d.Conn.Begin(d.ctx)
    return err
}

func (d *DBConn) Commit() error {
    glg.Trace("commit transaction")
    err := d.tx.Commit(d.ctx)
    d.tx = nil
    return err
}

func (d *DBConn) Rollback() {
    if d.tx != nil {
        glg.Trace("rollback transaction")
        d.tx.Rollback(d.ctx)
        d.tx = nil
    }
}

func  (d *DBConn) Close() {
    glg.Trace("released connection")
    d.Conn.Release()
}
