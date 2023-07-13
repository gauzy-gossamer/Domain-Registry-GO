package maintenance

import (
    "registry/server"
    "registry/epp"
    "registry/epp/dbreg"

    "github.com/robfig/cron"
)

func FinishExpiredTransferRequests(serv *server.Server, logger server.Logger, dbconn *server.DBConn) error {
    rows, err:= dbconn.Query("SELECT id, registrar_id, acquirer_id FROM epp_transfer_request WHERE status=$1::integer and acdate < now();", dbreg.TrPending)
    if err != nil {
        return err
    }
    defer rows.Close()

    for rows.Next() {
        var trid, regid, acid uint
        err = rows.Scan(&trid, &regid, &acid)
        if err != nil {
            return err
        }

        dbconn2, err := server.AcquireConn(serv.Pool, logger)
        if err != nil {
            return err
        }
        defer dbconn2.Close()

        err = dbreg.ChangeTransferRequestState(dbconn2, trid, dbreg.TrServerCancelled, regid, regid)
        if err != nil {
            return err
        }
    }

    return nil
}

func CreateLowCreditMessages(serv *server.Server, logger server.Logger, dbconn *server.DBConn) error {
    // for each reagistrar and zone count credit from advance invoices.
    // if credit is lower than limit and last poll message for this
    // registrar and zone is older than last advance invoice,
    // insert new poll message
    rows, err := dbconn.Query("SELECT rc.zone_id, rc.registrar_id, rc.credit, l.credlimit " +
         "FROM (SELECT r1.registrar_id, coalesce(zg.group_id, r1.zone_id), in_zone_group(r1.zone_id::integer, zg.zone_id::integer), min(r1.zone_id) as zone_id, sum(r1.credit) as credit " +
             "FROM registrar_credit r1 " +
                   "LEFT JOIN zone_groups zg ON r1.zone_id=zg.zone_id group by 1,2,3 ) rc " +
             "JOIN poll_credit_zone_limit l ON rc.zone_id=l.zone or in_zone_group(rc.zone_id::integer, l.zone::integer) " +
             "LEFT JOIN (SELECT m.clid, pc.zone, MAX(m.crdate) AS crdate " +
                         "FROM message m, poll_credit pc " +
                         "WHERE m.id=pc.msgid GROUP BY m.clid, pc.zone) AS mt " +
                 "ON (mt.clid=rc.registrar_id AND (mt.zone=rc.zone_id or in_zone_group(mt.zone::integer, rc.zone_id::integer))) " +
             "LEFT JOIN (SELECT i.registrar_id, i.zone_id, MAX(i.crdate) AS crdate " +
                         "FROM invoice i JOIN invoice_prefix ip ON i.invoice_prefix_id = ip.id AND ip.typ=0 " +
                         "GROUP BY i.registrar_id, i.zone_id) AS iii " + 
                 "ON (iii.registrar_id=rc.registrar_id AND (iii.zone_id=rc.zone_id or in_zone_group(iii.zone_id::integer, rc.zone_id::integer))) " +
         "WHERE rc.credit < l.credlimit AND (mt.crdate IS NULL or mt.crdate < iii.crdate );")
    if err != nil {
        return err
    }
    defer rows.Close()

    for rows.Next() {
        var zone_id, regid int
        var credit, credlimit float32

        err = rows.Scan(&zone_id, &regid, &credit, &credlimit)
        if err != nil {
            return err
        }

        dbconn2, err := server.AcquireConn(serv.Pool, logger)
        if err != nil {
            return err
        }
        defer dbconn2.Close()
        err = dbconn2.Begin()
        if err != nil {
            return err
        }
        defer dbconn2.Rollback()

        msgid, err := dbreg.CreatePollMessage(dbconn2, uint(regid), dbreg.POLL_LOW_CREDIT)
        if err != nil {
            return err
        }

        // insert into table poll_credit appropriate part from temp table
        _, err = dbconn2.Exec("INSERT INTO poll_credit(msgid, zone, credlimit, credit) VALUES($1::integer, $2::integer, $3::numeric, $4::numeric)", msgid, zone_id, credit, credlimit)
        if err != nil {
            return err
        }

        // create mail request notification
        _, err = dbreg.NewCreateMailRequest(uint64(regid), dbreg.MAIL_LOW_CREDIT).Exec(dbconn2)
        if err != nil {
            return err
        }

        err = dbconn2.Commit()
        if err != nil {
            return err
        }
    }

    return nil
}

func Maintenance(serv *server.Server) {
    fn := func() {
        logger := server.NewLogger("maintenance")

        logger.Info("run schedule")

        dbconn, err := server.AcquireConn(serv.Pool, logger)
        if err != nil {
            logger.Error(err)
            return
        }
        defer dbconn.Close()

        err = epp.UpdateObjectStates(dbconn, 0)
        if err != nil {
            logger.Error(err)
            return
        }

        if err = FinishExpiredTransferRequests(serv, logger, dbconn); err != nil {
            logger.Error(err)
        }

        if err = CreateLowCreditMessages(serv, logger, dbconn); err != nil {
            logger.Error(err)
        }
    }

    logger := server.NewLogger("maintenance")
    if serv.RGconf.CronSchedule == "" {
        logger.Info("no maintenance schedule")
        return
    }

    c := cron.New()
    err := c.AddFunc(serv.RGconf.CronSchedule, fn)
    if err != nil {
        logger.Fatal(err)
    }

    c.Start()
}
