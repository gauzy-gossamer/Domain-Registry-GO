package epp

import (
    "registry/server"
    "registry/epp/dbreg"
    "github.com/jackc/pgtype"
    "github.com/jackc/pgx/v5"
)

const (
    serverDeleteProhibited   = 1
    serverRenewProhibited    = 2
    serverTransferProhibited = 3
    serverUpdateProhibited   = 4
    serverBlocked   = 7
    stateExpired    = 9
    stateInactive   = 15
    stateLinked     = 16
    pendingDelete   = 17
    clientDeleteProhibited   = 29
    clientUpdateProhibited   = 30
    clientTransferProhibited = 31
    clientRenewProhibited    = 32
    clientHold               = 33
    pendingTransfer          = 34
    changeProhibited         = 35
    serverHold               = 36
)

type ObjectState struct {
    id int
    name string
    valid_from pgtype.Timestamp
    valid_to pgtype.Timestamp
    external bool
    manual bool
    importance int
}

type ObjectStates struct {
    States []ObjectState
}

func (o *ObjectStates) copyObjectStates() ([]string) {
    var states []string
    for _, v := range o.States {
        if v.external {
            states = append(states, v.name)
        }
    }
    return states
}

func (o *ObjectStates) hasState(stateid int) bool {
    for _, v := range o.States {
        if v.id == stateid {
            return true
        }
    }

    return false
}

func (o *ObjectStates) deleteState(stateid int) {
    delete_i := -1
    for i, v := range o.States {
        if v.id == stateid {
            delete_i = i
            break
        }
    }

    if delete_i != -1 {
        l := len(o.States)
        o.States[delete_i] = o.States[l-1]
        o.States = o.States[:l-1]
    }
}

func getObjectStates(db *server.DBConn, object_id uint64) (*ObjectStates, error) {
    query := "SELECT eos.id, eos.name, os.valid_from, os.valid_to " +
        " , eos.external, eos.manual, coalesce(eos.importance, 0) " +
        " FROM object_state os " +
            " JOIN enum_object_states eos ON eos.id = os.state_id " +
            " WHERE os.object_id = $1::bigint " +
                " AND os.valid_from <= CURRENT_TIMESTAMP " +
                " AND (os.valid_to IS NULL OR os.valid_to > CURRENT_TIMESTAMP) " +
        " ORDER BY eos.importance "
    rows, err := db.Query(query, object_id)
    defer rows.Close()

    if err != nil {
        return nil, err
    }

    var states ObjectStates

    for rows.Next() {
        var state ObjectState
        err := rows.Scan(&state.id, &state.name, &state.valid_from, &state.valid_to,
                &state.external, &state.manual, &state.importance)
        states.States = append(states.States, state)
        if err != nil {
            return nil, err
        }
    }

    return &states, nil
}

func UpdateObjectStates(db *server.DBConn, object_id uint64) error {
    _, err := db.Exec("SELECT update_object_states($1::integer);", object_id)
    return err
}

func updateHostStates(db *server.DBConn, hosts []dbreg.HostObj) error {
    for _, host := range hosts {
        if err := UpdateObjectStates(db, host.Id); err != nil {
            return err
        }
    }
    return nil
}

func getClientObjectStates(db *server.DBConn, states []string, object_type string) (map[string]int, error) {
    state_ids := make(map[string]int)

    for _, state := range states {
        row := db.QueryRow("SELECT id FROM enum_object_states WHERE manual = 't' and client_state = 't' and " +
                           "name = $1::text and types && array[get_object_type_id($2::text)]::int[];", state, object_type)
        var state_id int
        err := row.Scan(&state_id)
        if err != nil {
            if err == pgx.ErrNoRows {
                return state_ids, &dbreg.ParamError{Val:state + " is not available"}
            }
            return state_ids, err
        }
        state_ids[state] = state_id
    }

    return state_ids, nil
}
