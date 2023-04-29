package eppcom

const (
        EPP_UNKNOWN_CMD = iota
        EPP_DUMMY
        EPP_HELLO
        EPP_LOGIN
        EPP_LOGOUT

        /* query commands */
        EPP_CHECK_CONTACT
        EPP_CHECK_DOMAIN
        EPP_CHECK_HOST
        EPP_INFO_CONTACT
        EPP_INFO_DOMAIN
        EPP_INFO_HOST
        EPP_INFO_REGISTRAR
        EPP_POLL_REQ
        EPP_POLL_ACK

        /* transform commands */
        EPP_CREATE_CONTACT
        EPP_CREATE_DOMAIN
        EPP_CREATE_HOST
        EPP_DELETE_CONTACT
        EPP_DELETE_DOMAIN
        EPP_DELETE_HOST
        EPP_UPDATE_CONTACT
        EPP_UPDATE_DOMAIN
        EPP_UPDATE_HOST
        EPP_UPDATE_REGISTRAR
        EPP_TRANSFER_DOMAIN
        EPP_RENEW_DOMAIN
)
