package tests

import (
    "testing"

    "registry/tests/epptests"
    "registry/epp/dbreg"
    "registry/server"
    "registry/maintenance"
)

func TestMaintenance(t *testing.T) {
    serv := epptests.PrepareServer("../server.conf")
    logger := server.NewLogger("")
    dbconn, err := server.AcquireConn(serv.Pool, logger)
    if err != nil {
        panic(err)
    }

    /* test registrar = 3 test zone = 1 */
    regid := 3
    err = maintenance.CreateLowCreditMessages(serv, logger, dbconn)
    if err != nil {
        t.Error(err)
    }

    row := dbconn.QueryRow("SELECT count(*) FROM mail_request WHERE object_id = $1::integer and request_type_id = " +
                           "(select id from mail_request_type where request_type=$2::text)", regid, dbreg.MAIL_LOW_CREDIT)
    var count int
    err = row.Scan(&count)
    if err != nil {
        t.Error(err)
    }
    if count == 0 {
        t.Error("expected mail_request")
    }
}
