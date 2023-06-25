---
--- sequence for epp session identifiers
---
CREATE SEQUENCE epp_login_id_seq;

CREATE TABLE epp_session (
    clientid bigint DEFAULT ('x' || substring(replace(uuid_in(md5(random()::text || now()::text)::cstring)::text, '-','') from 16 ))::bit(64)::bigint,
    lang integer NOT NULL,
    regid integer NOT NULL,
    logd_session_id bigint DEFAULT 0,
    login_date timestamp NOT NULL,
    last_access timestamp,
    logout_date timestamp
);

CREATE INDEX epp_session_idx ON epp_session(clientid);
CREATE INDEX epp_session_logout_idx on epp_session(logout_date);
