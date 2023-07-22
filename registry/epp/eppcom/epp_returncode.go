package eppcom

const (
    EPP_OK           = 1000

    EPP_POLL_NO_MSG  = 1300 // when poll returns no messages
    EPP_POLL_ACK_MSG = 1301 // if there are messages in the poll queue

    EPP_CLOSING_LOGOUT = 1500 // on logout

    EPP_UNKNOWN_ERR = 2000
    EPP_SYNTAX_ERR  = 2001
    EPP_MISSING_PARAM = 2003 // missing parameter 
    EPP_PARAM_RANGE_ERR = 2004 // parameter is out of boundaries
    EPP_PARAM_ERR = 2005 // parameter value syntax error

    EPP_EXT_UNIMPLEMENTED = 2103 // unimplemented extension
    EPP_BILLING_FAILURE = 2104  // domain billing failure
    EPP_NOT_ELIGIBLE_FOR_RENEW =  2105    // Object is not eligible for renewal
    EPP_NOT_ELIGIBLE_FOR_TRANSFER = 2106 // Object is not eligible for transfer  

    EPP_AUTHENTICATION_ERR = 2200 // session/login information is incorrect
    EPP_AUTHORIZATION_ERR  = 2201 // client isn't authorized to access this object/perform operation

    EPP_OBJECT_EXISTS     = 2302
    EPP_OBJECT_NOT_EXISTS = 2303

    EPP_STATUS_PROHIBITS_OPERATION = 2304 // current status flag doesn't correspond
    EPP_LINKED_PROHIBITS_OPERATION = 2305
    EPP_PARAM_VALUE_POLICY         = 2306 // bad value e.g. status flag server from client

    EPP_FAILED = 2400
    EPP_INTERNAL_ERR = 2401
    EPP_AUTH_CLOSING_ERR   = 2501

    EPP_SESSION_LIMIT   = 2502
)

