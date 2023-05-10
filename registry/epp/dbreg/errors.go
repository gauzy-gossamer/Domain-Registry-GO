package dbreg

type ParamError struct {
    Val string
}

func (e *ParamError) Error() string {
    return "param error"
}

type BillingFailure struct {
}

func (e *BillingFailure) Error() string {
    return "billing failure"
}


