package dbreg

import (
    "strings"
    "strconv"
    "registry/server"
    . "registry/epp/eppcom"
)

const (
    TrPending = iota
    TrClientCancelled
    TrClientRejected
    TrClientApproved
    TrServerCancelled
)

func GetTransferMsg(status uint) string {
    switch status {
        case TrPending:
            return "Transfer requested."
        case TrClientCancelled:
            return "Transfer cancelled."
        case TrClientRejected:
            return "Transfer rejected."
        case TrClientApproved:
            return "Transfer approved."
        case TrServerCancelled:
            return "Transfer cancelled."
        default:
            return "unknown"
    }
}

type FindTransferRequest struct {
    Ownerid uint
    Acquirerid uint
    Domainid uint64
    TrID uint
    ActiveOnly bool
    lock bool
}

func (f *FindTransferRequest) SetLock(lock bool) *FindTransferRequest {
    f.lock = lock
    return f
}

func (f *FindTransferRequest) Exec(db *server.DBConn) (*TransferRequestObject, error) {
    var query strings.Builder
    query.WriteString("SELECT et.id, status, st.name, created, acdate, registrar_id, r1.handle as registrar_handle, acquirer_id,r3.handle as acquirer_handle, ")
    query.WriteString(" upid, r2.handle as upid_handle, acdate < now() AT TIME ZONE 'UTC'")
    query.WriteString(" FROM epp_transfer_request et JOIN enum_transfer_states st ON et.status=st.id")
    query.WriteString(" INNER JOIN registrar r1 on et.registrar_id=r1.id INNER JOIN registrar r2 on et.upid=r2.id ")
    query.WriteString(" INNER JOIN registrar r3 on et.acquirer_id=r3.id ")
    query.WriteString(" WHERE 1=1 ")

    var params []any
    if f.Domainid > 0 {
        params = append(params, f.Domainid)
        query.WriteString("and domain_id = $1::bigint ")
    }
    if f.ActiveOnly {
        query.WriteString("and status = 0 and acdate > now() AT TIME ZONE 'UTC' ")
    }
    if f.Acquirerid > 0 {
        params = append(params, f.Acquirerid)
        query.WriteString("and acquirer_id = $" + strconv.Itoa(len(params)) + "::bigint ")
    }
    if f.TrID > 0 {
        params = append(params, f.TrID)
        query.WriteString("and et.id = $" + strconv.Itoa(len(params)) + "::bigint ")
    }
    if f.Ownerid > 0 {
        params = append(params, f.Ownerid)
        query.WriteString("and registrar_id = $" + strconv.Itoa(len(params)) + "::bigint ")
    }
    query.WriteString("ORDER BY created DESC ")
    if f.lock {
        query.WriteString("FOR UPDATE of et ")
    } else {
        query.WriteString("FOR SHARE of et ")
    }
    query.WriteString("LIMIT 1")

    row := db.QueryRow(query.String(), params...)
    var tr_request TransferRequestObject
    var active bool
    err := row.Scan(&tr_request.Id, &tr_request.StatusId, &tr_request.Status, &tr_request.ReDate, &tr_request.AcDate, &tr_request.ReID.Id, &tr_request.ReID.Handle,
                    &tr_request.AcID.Id, &tr_request.AcID.Handle, &tr_request.UpID.Id, &tr_request.UpID.Handle,  &active)

    return &tr_request, err
}

/* check whether there's a pending transfer for this registrar on provided domain */
func CheckExistingTransferByDomain(db *server.DBConn, domainid uint64, acquirerid uint) (bool, error) {
    query := "SELECT count(*) FROM epp_transfer_request WHERE domain_id = $1::bigint and status = 0 and acdate > now() AT TIME ZONE 'UTC' and acquirer_id = $2::integer"

    row := db.QueryRow(query, domainid, acquirerid)
    var count int
    err := row.Scan(&count)
    if err != nil {
        return false, err
    }

    return count > 0, nil
}

/* check whether there's a pending transfer for this registrar on any domains associated with this contact */
func CheckExistingTransferByContact(db *server.DBConn, contactid uint64, acquirerid uint) (bool, error) {
    query := "SELECT count(*) FROM epp_transfer_request et JOIN domain d on et.domain_id = d.id JOIN contact c ON d.registrant = c.id " +
             "WHERE c.id = $1::bigint and status = 0 and acdate > now() AT TIME ZONE 'UTC' and acquirer_id = $2::integer"

    row := db.QueryRow(query, contactid, acquirerid)
    var count int
    err := row.Scan(&count)
    if err != nil {
        return false, err
    }

    return count > 0, nil
}

type CreateTransferRequest struct {
    ownerid uint
    acquirerid uint
    domainid uint64
}

func (c *CreateTransferRequest) SetParams(ownerid uint, acquirerid uint, domainid uint64) *CreateTransferRequest {
    c.ownerid = ownerid
    c.acquirerid = acquirerid
    c.domainid = domainid
    return c
}

func (c *CreateTransferRequest) createNotifications(db *server.DBConn, notify_registrar uint) (uint, error) {
    _, err := NewCreateMailRequest(uint64(notify_registrar), MAIL_NEW_TRANSFER).SetDomainID(c.domainid).Exec(db)
    if err != nil {
        return 0, err
    }

    return CreatePollMessage(db, notify_registrar, POLL_TRANSFER_REQUEST)
}

func (c *CreateTransferRequest) Exec(db *server.DBConn, notify_registrar uint) (uint, error) {
    poll_msg_id, err := c.createNotifications(db, notify_registrar)
    if err != nil {
        return 0, err
    }

    var params []any
    params = append(params, c.domainid)
    params = append(params, c.ownerid)
    params = append(params, c.acquirerid)
    params = append(params, c.acquirerid)

    row := db.QueryRow("INSERT INTO epp_transfer_request(domain_id, status, registrar_id, acquirer_id, upid) " +
                       "VALUES($1::bigint, 0, $2::bigint, $3::bigint, $4::bigint) returning id", params...)

    var tr_request_id uint
    err = row.Scan(&tr_request_id)
    if err != nil {
        return 0, err
    }

    _, err = db.Exec("INSERT INTO epp_transfer_request_state_change(request_id, msgid, status) " +
                     "VALUES($1::bigint, $2::bigint, 0)", tr_request_id, poll_msg_id)

    return tr_request_id, err
}

func ChangeTransferRequestState(db *server.DBConn, tr_id uint, status int, updated_registrar uint, notify_registrar uint) error {
    poll_msg_id, err := CreatePollMessage(db, notify_registrar, POLL_TRANSFER_REQUEST)
    if err != nil {
        return err
    }

    row := db.QueryRow("UPDATE epp_transfer_request SET status = $1::integer, " +
                       " acdate = now() AT TIME ZONE 'UTC', upid = $2::bigint " +
                       " WHERE id = $3::integer returning id", status, updated_registrar, tr_id)
    var changed_transfer int
    err = row.Scan(&changed_transfer)
    if err != nil {
        return err
    }

    _, err = db.Exec("INSERT INTO epp_transfer_request_state_change(request_id, msgid, status) " +
                     " VALUES($1::bigint, $2::bigint, $3::integer)", tr_id, poll_msg_id, status)

    return err
}
