package eppcom

type NullableVal struct {
    val interface{}
    null bool
}

func (n *NullableVal) Get() interface{} {
    if n.null {
        return nil
    }
    return n.val
}

func (n *NullableVal) Set(val interface{}) {
    n.val = val
    n.null = val == nil
}

func (n *NullableVal) IsNull() bool {
    return n.null
}

func (n *NullableVal) Scan(src interface{}) error {
    if src == nil {
        n.val = nil
        n.null = true
    } else {
        n.val = src
        n.null = false
    }
    return nil
}

type NullableUint struct {
    NullableVal
}

func (n *NullableUint) Get() uint {
    if n.null {
        return 0
    }
    switch n.val.(type) {
        case int:
            return uint(n.val.(int))
        case uint:
            return n.val.(uint)
        case int64:
            return uint(n.val.(int64))
        case uint64:
            return uint(n.val.(uint64))
        default:
            return 0

    }
}
