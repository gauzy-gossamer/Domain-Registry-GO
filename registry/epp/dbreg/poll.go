package dbreg

import (
    "registry/server"
    . "registry/epp/eppcom"
    "github.com/jackc/pgtype"
)

func CreatePollMessage(db *server.DBConn, registrar_id uint, msg_type int) (uint, error) {
    row := db.QueryRow("INSERT INTO message(clid, msgtype, crdate, exdate)" +
                       " VALUES($1::bigint,$2::bigint,now(), now() + interval '7 day')" +
                       " RETURNING id", registrar_id, msg_type)

    var poll_msg_id uint
    err := row.Scan(&poll_msg_id)

    return poll_msg_id, err
}

func GetPollMessageCount(db *server.DBConn, regid uint) (uint, error) {
    query := "SELECT count(*) FROM message WHERE seen='f' " +
             "and clid=$1::bigint and exdate > now() and msgtype in(22,1)"
    row := db.QueryRow(query, regid)

    var count uint
    err := row.Scan(&count)
    return count, err
}

func getPollTransferObject(db *server.DBConn, msgid uint) (*PollMessage, error) {
    var poll_msg PollMessage
    row := db.QueryRow("SELECT request_id, status FROM epp_transfer_request_state_change "+
                       "WHERE msgid=$1::integer", msgid)
    var requestid, status uint
    err := row.Scan(&requestid, &status)
    if err != nil {
        return &poll_msg, err
    }
    find_transfer := FindTransferRequest{TrID:requestid}
    transfer_obj, err := find_transfer.Exec(db)
    if err != nil {
        return &poll_msg, err
    }

    poll_msg.Msgid = msgid
    poll_msg.Msg = "changed state"
    poll_msg.Content = transfer_obj

    return &poll_msg, nil
}

func GetFirstUnreadPollMessage(db *server.DBConn, regid uint) (*PollMessage, error) {
    query := "SELECT id, msgtype, crdate, exdate FROM message WHERE seen='f' " +
             "and clid=$1::bigint and exdate > now() and msgtype in(22,1) ORDER BY id LIMIT 1"
    row := db.QueryRow(query, regid)

    var msgtype, msgid uint
    var crdate, exdate pgtype.Timestamp
    err := row.Scan(&msgid, &msgtype, &crdate, &exdate)
    if err != nil {
        return nil, err
    }

    var poll_msg *PollMessage
    switch msgtype {
        case POLL_LOW_CREDIT:
            poll_msg = &PollMessage{Msgid:msgid, Msg:"Credit balance low."}
        case POLL_TRANSFER_REQUEST:
            poll_msg, err = getPollTransferObject(db, msgid)
            if err != nil {
                return poll_msg, err
            }
        default:
            poll_msg = &PollMessage{Msgid:msgid, Msg:"unsupported"}
    }
    poll_msg.MsgType = msgtype
    poll_msg.QDate = crdate

    return poll_msg, nil
}

func MarkMessageRead(db *server.DBConn, regid uint, msgid uint64) error {
    query := "UPDATE message SET seen = 't' WHERE id= $1::integer and clid=$2::bigint"
     _, err := db.Exec(query, msgid, regid)
    return err
}

