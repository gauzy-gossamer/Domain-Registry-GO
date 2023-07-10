package contact

import (
    "registry/epp/dbreg"
    "registry/server"
)

func CopyContact(db *server.DBConn, contactid uint64, new_handle string, target_regid uint) error {
    err := dbreg.LockObjectById(db, contactid, "contact")
    if err != nil {
        return err
    }

    createObj := dbreg.NewCreateObjectDB("contact")
    create_result, err := createObj.Exec(db, new_handle, target_regid)

    if err != nil {
        return err 
    }   

    query := "INSERT INTO contact(id, contact_type, email, telephone, intpostal, intaddress, locpostal, locaddress, vat, legaladdress, fax, birthday, passport) "
    query += "SELECT $1::bigint, contact_type, email, telephone, intpostal, intaddress, locpostal, locaddress, vat, legaladdress, fax, birthday, passport FROM contact WHERE id = $2::bigint"

    _, err = db.Exec(query, create_result.Id, contactid)
    if err != nil {
        return err 
    }   

    return nil 
}
