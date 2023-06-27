package logrpc

import (
    "logger/logging"
)

type RequestContext struct {
    Logger logging.Logger
}

type StorageModule interface {
    StartRequest  (*RequestContext, *LogRequest) (uint64, error)
    EndRequest (*RequestContext, uint64, uint32) error
}
