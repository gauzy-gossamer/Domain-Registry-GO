package server

import (
    "errors"
    "sync"
    "fmt"
    "time"
    "github.com/jackc/pgx/v5"
    "github.com/jackc/pgtype"
    "github.com/kpango/glg"
)

type EPPSession struct {
    Sessionid uint64
    Regid uint
    Lang uint
    System bool
    last_access pgtype.Timestamp
    requests_limit int
}

type QueryLimitWindow struct {
    queries uint /* number of queries */
    started time.Time /* when the window started */
}

type EPPSessions struct {
    MaxQueriesPerMinute uint

    MaxRegistrarSessions uint
    /* expire session after timeout */
    SessionTimeoutSec uint

    registrar_session_count map[uint]uint
    registrar_queries_count map[uint]*QueryLimitWindow
    sessions map[uint64]EPPSession

    mu sync.Mutex
}

func (s *EPPSessions) InitSessions(db *DBConn) {
    s.mu.Lock()
    defer s.mu.Unlock()

    s.removeExpiredSessions(db)

    rows, err := db.Query("SELECT clientid, regid, lang, last_access AT TIME ZONE 'UTC', logd_session_id, r.system, r.epp_requests_limit FROM epp_session eps " +
                          "INNER JOIN registrar r ON eps.regid=r.id;")
    if err != nil {
        panic(err)
    }
    defer rows.Close()
    s.registrar_session_count = make(map[uint]uint)
    s.registrar_queries_count = make(map[uint]*QueryLimitWindow)
    s.sessions = map[uint64]EPPSession{}

    for rows.Next() {
        var session EPPSession
        var logdsessionid uint64
        var sessionid int64
        err = rows.Scan(&sessionid, &session.Regid, &session.Lang, &session.last_access, &logdsessionid, &session.System, &session.requests_limit)
        if err != nil {
            glg.Fatal(err)
        }
        session.Sessionid = uint64(sessionid)

        s.sessions[session.Sessionid] = session

        if _, ok := s.registrar_session_count[session.Regid]; ok {
            s.registrar_session_count[session.Regid] += 1
        } else {
            s.registrar_session_count[session.Regid] = 1
        }

    }
}

/* returns true if query limit per minute exceeded */
func (s *EPPSessions) QueryLimitExceeded(regid uint) bool {
    if s.MaxQueriesPerMinute == 0 {
        return false
    }
    s.mu.Lock()
    defer s.mu.Unlock()

    t_now := time.Now()

    queries_data, ok := s.registrar_queries_count[regid]
    /* if more than one minute passed, restart the window */
    if !ok || t_now.Sub(queries_data.started) > time.Minute {
        s.registrar_queries_count[regid] = &QueryLimitWindow{queries:1, started:t_now}
        return false
    }

    queries_data.queries += 1

    return queries_data.queries > s.MaxQueriesPerMinute
}

func (s *EPPSessions) LoginSession(db *DBConn, regid uint, lang uint) (uint64, error) {
    s.mu.Lock()
    defer s.mu.Unlock()

    if _, ok := s.registrar_session_count[regid]; ok {
        /* this will update session counter if there are expired sessions */
        s.removeExpiredSessions(db)
        if s.registrar_session_count[regid] >= s.MaxRegistrarSessions {
            return 0, errors.New("session limit exceeded")
        }
    }
    var sessionid int64
    row := db.QueryRow("INSERT INTO epp_session(login_date, last_access, lang, regid) " +
                       "VALUES(now() AT TIME ZONE 'UTC', now() AT TIME ZONE 'UTC', $1::integer, $2::integer) returning clientid", lang, regid)
    err := row.Scan(&sessionid)
    if err != nil {
        return 0, err
    }

    s.sessions[uint64(sessionid)] = EPPSession{Sessionid:uint64(sessionid), Lang:lang, Regid:regid}

    if _, ok := s.registrar_session_count[regid]; ok {
        s.registrar_session_count[regid] += 1
    } else {
        s.registrar_session_count[regid] = 1
    }
    glg.Trace("registrar sessions", s.registrar_session_count[regid])

    return uint64(sessionid), nil
}

func (s *EPPSessions) CheckSession(db *DBConn, sessionid uint64) *EPPSession {
    row := db.QueryRow("SELECT lang, regid, logd_session_id, last_access AT TIME ZONE 'UTC', r.system, r.epp_requests_limit, now() AT TIME ZONE 'UTC' FROM epp_session eps "+
                       "INNER JOIN registrar r ON eps.regid=r.id WHERE clientid = $1::bigint", int64(sessionid))
    var logsessionid int
    var t_now pgtype.Timestamp
    session := EPPSession{Sessionid:sessionid}
    err := row.Scan(&session.Lang, &session.Regid, &logsessionid, &session.last_access, &session.System, &session.requests_limit, &t_now)
    if err != nil {
        if err != pgx.ErrNoRows {
            glg.Error(err, session.Regid)
        }
        return nil
    }

    /* we don't need to update last_access every query, only update if the difference is more than 1 second */
    if session.last_access.Status == pgtype.Null || (t_now.Status != pgtype.Null && t_now.Time.Sub(session.last_access.Time) > time.Second) {
        s.updateSessionTimer(db, sessionid)
    }

    return &session
}

func (s *EPPSessions) updateSessionTimer(db *DBConn, sessionid uint64) {
    _, err := db.Exec("UPDATE epp_session SET last_access = now() AT TIME ZONE 'UTC' WHERE clientid = $1::bigint", int64(sessionid))
    if err != nil {
        glg.Error(err)
    }
}

func (s *EPPSessions) LogoutSession(db *DBConn, sessionid uint64) error {
    s.mu.Lock()
    defer s.mu.Unlock()
    return s.logoutSessionLockFree(db, sessionid)
}

func (s *EPPSessions) logoutSessionLockFree(db *DBConn, sessionid uint64) error {
    var regid uint

    if session_obj, ok := s.sessions[sessionid]; !ok {
        return nil
    } else {
        regid = session_obj.Regid
    }

    if _, ok := s.registrar_session_count[regid]; ok {
        registrar_sessions := s.registrar_session_count[regid]
        if registrar_sessions != 0 {
            s.registrar_session_count[regid] -= 1
        }
    }

    delete(s.sessions, sessionid)

    _, err := db.Exec("DELETE FROM epp_session WHERE clientid = $1::bigint", int64(sessionid))

    if err != nil {
        glg.Error(err)
    }

    return err
}

func (s *EPPSessions) removeExpiredSessions(db *DBConn) {
    rows, err := db.Query(fmt.Sprintf("SELECT clientid FROM epp_session " +
                    "WHERE last_access < now() AT TIME ZONE 'UTC' - interval '%d seconds'", s.SessionTimeoutSec))
    defer rows.Close()
    if err != nil {
        glg.Error(err)
        return
    }
    var expired_sessions []int64
    for rows.Next() {
        var sessionid int64
        rows.Scan(&sessionid)
        expired_sessions = append(expired_sessions, sessionid)
    }

    for _, sessionid := range expired_sessions {
        glg.Error("remove", sessionid)
        if _, ok := s.sessions[uint64(sessionid)]; ok {
            s.logoutSessionLockFree(db, uint64(sessionid))
        } else {
            _, err = db.Exec("DELETE FROM epp_session WHERE clientid = $1::bigint", sessionid)
            if err != nil {
                glg.Error(err)
            }
        }
    }
}
