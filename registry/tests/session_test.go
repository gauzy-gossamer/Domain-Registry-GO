/*
go test -coverpkg=./... -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
*/
package tests

import (
    "testing"
//    "time"
    "registry/server"
)

func TestSessionQueryLimit(t *testing.T) {
    db := prepareDB()
    epp_session := server.EPPSessions{}
    epp_session.InitSessions(db)
    regid := uint(1)
    
    for i := 0; i < 1000; i ++ {
        if epp_session.QueryLimitExceeded(regid) {
            t.Error("exceeded on 0 limit")
        }
    }

    epp_session.MaxQueriesPerMinute = 100

    for i := 0; i < 1000; i ++ {
        if epp_session.QueryLimitExceeded(regid) {
            if i != 100 {
                t.Error("wrong limit ", i)
            }
            break
        }
        if i > 110 {
            t.Error("not exceeded")
            break
        }
    }
    if epp_session.QueryLimitExceeded(2) {
        t.Error("registrar 2 should be fine")
    }
    /*
    time.Sleep(1*time.Minute)
    if epp_session.QueryLimitExceeded(regid) {
        t.Error("limit didn't reset")
    }
    */
}
