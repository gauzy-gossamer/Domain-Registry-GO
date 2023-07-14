package dbreg

import (
    "registry/server"
)

const (
    MAIL_NEW_TRANSFER = "transfer_new"
    MAIL_LOW_CREDIT = "lowcredit"
)

/* object_id is usually registrar_id
   domain_id is optional and only used for transfer notifications */
type CreateMailRequest struct {
    object_id uint64
    domain_id uint64
    mail_type string
}

func NewCreateMailRequest(object_id uint64, mail_type string) *CreateMailRequest {
    return &CreateMailRequest{
        object_id:object_id,
        mail_type:mail_type,
    }
}

func (c *CreateMailRequest) SetDomainID(domain_id uint64) *CreateMailRequest {
    c.domain_id = domain_id
    return c
}

func (c *CreateMailRequest) Exec(db *server.DBConn) (uint, error) {
    row := db.QueryRow("SELECT id FROM mail_request_type WHERE request_type = $1::text", c.mail_type)

    var request_type_id int
    err := row.Scan(&request_type_id)
    if err != nil {
        return 0, err
    }

    row = db.QueryRow("INSERT INTO mail_request(object_id, domain_id, request_type_id)" +
                      " VALUES($1::bigint, $2::bigint, $3::integer) RETURNING id", 
                      c.object_id, c.domain_id, request_type_id)

    var msg_id uint
    err = row.Scan(&msg_id)

    return msg_id, err
}
