package postgres

import (
    "github.com/jackc/pgx/v5/pgxpool"
)

type DBConfig struct {
    Host     string
    Port     int
    User     string
    Password string
    DBname   string
}

type PostgresStorage struct {
    DBConf DBConfig
    Pool *pgxpool.Pool

    Partitions string
}

func NewPostgresStorage() PostgresStorage {
    storage := PostgresStorage{}
    return storage
}

func (p *PostgresStorage) InitModule(dbconf *DBConfig) error {
    var err error
    p.Pool, err = CreatePool(dbconf)
    return err
}
