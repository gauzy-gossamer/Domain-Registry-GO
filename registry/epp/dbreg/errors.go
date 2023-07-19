package dbreg

import (
    "errors"
)

type ParamError struct {
    Val string
}

func (e *ParamError) Error() string {
    return e.Val
}

type BillingFailure struct {
}

func (e *BillingFailure) Error() string {
    return "billing failure"
}


var ObjectNotFound = errors.New("object not found")

