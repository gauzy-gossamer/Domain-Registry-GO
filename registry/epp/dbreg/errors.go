package dbreg

import (
    "fmt"
)

type ParamError struct {
    Val string
}

func (e *ParamError) Error() string {
    return fmt.Sprintf("param error")
}

type BillingFailure struct {
}

func (e *BillingFailure) Error() string {
    return fmt.Sprintf("billing failure")
}


