package server

import (
    "github.com/jackc/pgx/v5/pgxpool"
)

type Server struct {
    RGconf RegConfig
    Pool *pgxpool.Pool
}
