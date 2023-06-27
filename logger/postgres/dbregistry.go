package postgres

import (
    "logger/logrpc"
    "logger/logging"
)

var SERVICE_ID = 3

func (st *PostgresStorage) StartRequest(ctx *logrpc.RequestContext, logreq *logrpc.LogRequest) (uint64, error) {
    dbconn, err := AcquireConn(st.Pool, ctx.Logger)
    if err != nil {
        return 0, err
    }
    defer dbconn.Close()

    _, err = dbconn.Exec("INSERT INTO request(time_begin, source_ip, service_id, request_type_id, session_id, user_id, is_monitoring) " +
                         "VALUES(now(), $1::inet,  $2::integer, $3::integer, $4::bigint, $5::integer, $6::boolean); ",
                         logreq.SourceIP, SERVICE_ID, logreq.RequestType, logreq.SessionID, logreq.UserID, logreq.IsMonitoring)
    if err != nil {
        return 0, err
    }
  
    row := dbconn.QueryRow("SELECT currval(pg_get_serial_sequence('request', 'id'));")

    var request_id uint64
    err = row.Scan(&request_id)
    if err != nil {
        return 0, err
    }

    return request_id, nil
}

func (st *PostgresStorage) EndRequest(ctx *logrpc.RequestContext, request_id uint64, result_code_id uint32) error {
    dbconn, err := AcquireConn(st.Pool, logging.NewLogger(""))
    if err != nil {
        return err
    }
    defer dbconn.Close()

    _, err = dbconn.Exec("UPDATE request SET time_end = now(), result_code_id = $1::integer " +
                          "WHERE id = $2::bigint", result_code_id, request_id)

    if err != nil {
        return err
    }

    return nil
}
