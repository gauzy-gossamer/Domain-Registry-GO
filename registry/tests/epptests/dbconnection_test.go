package epptests

import (
    "time"
    "testing"
    "registry/server"
    "log"
)

func TestTxRetry(t *testing.T) {
    serv := prepareServer()

    dbconn, err := server.AcquireConn(serv.Pool, server.NewLogger(""))
    if err != nil {
        t.Fatal(err)
    }

    done := make(chan struct{})

    _, err = dbconn.Exec("create table if not exists testTx(s text);")
    if err != nil {
        t.Fatal(err)
    }

    defer func() {
        _, err := dbconn.Exec("drop table testTx;")
        if err != nil {
            t.Fatal(err)
        }
    }()

    go dbconn.RetryTx(func() error {
        dbconn2, err := server.AcquireConn(serv.Pool, server.NewLogger(""))
        if err != nil {
            t.Fatal(err)
        }
        dbconn2.BeginSerializable()
        defer dbconn2.Rollback()
        row := dbconn2.QueryRow("SELECT count(*) FROM testTx")
        count := 0
        err = row.Scan(&count)
        if err != nil {
            t.Fatal(err)
        }

        time.Sleep(time.Millisecond*200)

        _, err = dbconn2.Exec("INSERT INTO testTx(s) VALUES('test1');")
        if err != nil {
            log.Println(err)
            return err
        }

        err = dbconn2.Commit()
        if err != nil {
            return err
        }
        done <- struct{}{}
        return nil
    })

    dbconn.RetryTx(func() error {
        dbconn.BeginSerializable()
        defer dbconn.Rollback()
        row := dbconn.QueryRow("SELECT count(*) FROM testTx")
        count := 0
        err := row.Scan(&count)
        if err != nil {
            t.Fatal(err)
        }

        _, err = dbconn.Exec("INSERT INTO testTx(s) VALUES('test2');")
        if err != nil {
            return err
        }

        return dbconn.Commit()
    })

    select {
        case _ = <-done:
            break
        case <-time.After(time.Second*2):
            t.Error("timeout")
    }
}
