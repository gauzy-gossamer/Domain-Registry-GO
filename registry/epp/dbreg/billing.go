package dbreg

import (
//    . "registry/epp/eppcom"
    "registry/server"
    "github.com/jackc/pgtype"
    "github.com/shopspring/decimal"
)

type RegistrarCredit struct {
    credit_id uint
    balance decimal.Decimal
}

func getRegistrarCredit(db *server.DBConn, regid uint, zoneid int) (*RegistrarCredit, error)  {
    //get_registrar_credit - lock record in registrar_credit table for registrar and zone
    // There could be several credits for each registrar grouped by zones (as in zone_groups table). 
    // In addition, there could be future payments in registrar_promised_payment table.
    // So we have to sum grouped credits and add promised payments to them. This value is then used to determine whethere there's sufficient credit for the current operation
    // 
    row := db.QueryRow("SELECT id, (SELECT sum(credit) FROM registrar_credit c WHERE rc.registrar_id=c.registrar_id " +
                " AND (zone_id = rc.zone_id or in_zone_group(zone_id::integer, rc.zone_id::integer))) " +
                " + coalesce((SELECT sum(amount) FROM registrar_promised_payment WHERE valid_until > now() AT TIME ZONE 'UTC'" +
                " and registrar_id = rc.registrar_id and (zone_id = rc.zone_id or in_zone_group(zone_id::integer, rc.zone_id::integer))), 0)" +
                " FROM registrar_credit rc " +
                " WHERE rc.registrar_id = $1::bigint " +
                    " AND (rc.zone_id = $2::integer or in_zone_group(rc.zone_id::integer, $2::integer)) " +
            " ORDER BY credit desc " +
            " FOR UPDATE LIMIT 1", regid, zoneid)

    var credit RegistrarCredit

    err := row.Scan(&credit.credit_id, &credit.balance)
    if err != nil {
        return nil, err
    }

    return &credit, nil
}

func createCreditTransaction(db *server.DBConn, credit_id uint, price decimal.Decimal) (int, error) {
    row := db.QueryRow("INSERT INTO registrar_credit_transaction " +
                    " (id, balance_change, registrar_credit_id) " +
                    " VALUES (DEFAULT, $1::numeric , $2::bigint) " +
                " RETURNING id ", price.Neg(), credit_id)
    var credit_transaction_id int

    err := row.Scan(&credit_transaction_id)

    return credit_transaction_id, err
}

func getOperationPrice(db *server.DBConn, operation string, zoneid int, op_timestamp pgtype.Timestamp) (decimal.Decimal, int, error) {
    query := "SELECT eo.id, price FROM price_list pl " +
             " JOIN enum_operation eo ON pl.operation_id = eo.id JOIN zone z ON z.id = pl.zone_id "

    query += " WHERE pl.valid_from <= $1::timestamp " +
             " AND (pl.valid_to is NULL OR pl.valid_to > $1::timestamp ) " +
             " AND pl.zone_id = $2::bigint and operation = $3::text"

    row := db.QueryRow(query, op_timestamp, zoneid, operation)

    var price decimal.Decimal
    var opid int
    err := row.Scan(&opid, &price)

    return price, opid, err
}

func createInvoice(db *server.DBConn, object_id uint64, regid uint, zoneid int, opid int, quantity int, registrar_transaction_id int, op_timestamp pgtype.Timestamp) error {
     cols := "INSERT INTO invoice_operation(object_id, registrar_id, operation_id, zone_id" +
             " , crdate, quantity, date_from, registrar_credit_transaction_id "
     vals := ") VALUES ($1::bigint, $2::bigint, $3::bigint, $4::bigint " +
             " , CURRENT_TIMESTAMP::timestamp, $5::integer, $6::date, $7::integer)"

     var params []any
     params = append(params, object_id)
     params = append(params, regid)
     params = append(params, opid)
     params = append(params, zoneid)
     params = append(params, quantity)
     params = append(params, op_timestamp)
     params = append(params, registrar_transaction_id)

     _, err := db.Exec(cols + vals, params...)

    return err
}

func charge_billing_op(db *server.DBConn, operation string, object_id uint64, regid uint, zoneid int, op_timestamp pgtype.Timestamp) error {
    credit, err := getRegistrarCredit(db, regid, zoneid)
    if err != nil {
        return err
    }
    op_price, opid, err := getOperationPrice(db, operation, zoneid, op_timestamp)
    if err != nil {
        return err
    }
    if op_price != decimal.NewFromInt(0) && credit.balance.LessThan(op_price) {
        return &BillingFailure{}
    }
    credit_transaction_id, err := createCreditTransaction(db, credit.credit_id, op_price)
    if err != nil {
        return err
    }
    return createInvoice(db, object_id, regid, zoneid, opid, 1, credit_transaction_id, op_timestamp)
}

func ChargeCreateOp(db *server.DBConn, object_id uint64, regid uint, zoneid int, op_timestamp pgtype.Timestamp) error {
    return charge_billing_op(db, "CreateDomain", object_id, regid, zoneid, op_timestamp)
}

func ChargeRenewOp(db *server.DBConn, object_id uint64, regid uint, zoneid int, op_timestamp pgtype.Timestamp) error {
    return charge_billing_op(db, "RenewDomain", object_id, regid, zoneid, op_timestamp)
}
