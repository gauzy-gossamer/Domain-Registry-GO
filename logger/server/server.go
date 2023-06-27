package server

import (
    "logger/logrpc"
)

type Server struct {
    RGconf RegConfig
    Storage logrpc.StorageModule
}
