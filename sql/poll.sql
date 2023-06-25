CREATE TABLE MessageType (
        ID INTEGER CONSTRAINT messagetype_pkey PRIMARY KEY,
	name VARCHAR(30) NOT NULL
);
-- do not change the number codes - current code depends on it!
INSERT INTO MessageType VALUES (01, 'credit');
INSERT INTO MessageType VALUES (02, 'techcheck');
INSERT INTO MessageType VALUES (03, 'transfer_contact');
INSERT INTO MessageType VALUES (04, 'transfer_nsset');
INSERT INTO MessageType VALUES (05, 'transfer_domain');
INSERT INTO MessageType VALUES (06, 'idle_delete_contact');
INSERT INTO MessageType VALUES (07, 'idle_delete_nsset');
INSERT INTO MessageType VALUES (08, 'idle_delete_domain');
INSERT INTO MessageType VALUES (09, 'imp_expiration');
INSERT INTO MessageType VALUES (10, 'expiration');
INSERT INTO MessageType VALUES (11, 'imp_validation');
INSERT INTO MessageType VALUES (12, 'validation');
INSERT INTO MessageType VALUES (13, 'outzone');
INSERT INTO MessageType VALUES (14, 'transfer_keyset');
INSERT INTO MessageType VALUES (15, 'idle_delete_keyset');
INSERT INTO MessageType VALUES (17, 'update_domain');
INSERT INTO MessageType VALUES (18, 'update_nsset');
INSERT INTO MessageType VALUES (19, 'update_keyset');
INSERT INTO MessageType VALUES (20, 'delete_contact');
INSERT INTO MessageType VALUES (21, 'delete_domain');
INSERT INTO MessageType VALUES (22, 'transfer_domain_request');

comment on table MessageType is
'table with message number codes and its names

id - name
01 - credit
02 - techcheck
03 - transfer_contact
04 - transfer_nsset
05 - transfer_domain
06 - delete_contact
07 - delete_nsset
08 - delete_domain
09 - imp_expiration
10 - expiration
11 - imp_validation
12 - validation
13 - outzone';

CREATE TABLE Message (
        ID SERIAL CONSTRAINT message_pkey PRIMARY KEY,
        ClID INTEGER NOT NULL CONSTRAINT message_clid_fkey REFERENCES Registrar ON UPDATE CASCADE,
        CrDate timestamp NOT NULL DEFAULT now(),
        ExDate TIMESTAMP,
        Seen BOOLEAN NOT NULL DEFAULT false,
	MsgType INTEGER CONSTRAINT message_msgtype_fkey REFERENCES messagetype (id)
);
CREATE INDEX message_clid_idx ON message (clid);
CREATE INDEX message_seen_idx ON message (clid,seen,crdate,exdate);

comment on table Message is 'Evidence of messages for registrars, which can be picked up by epp poll funcion';

CREATE TABLE poll_credit (
  msgid INTEGER CONSTRAINT poll_credit_pkey PRIMARY KEY
  CONSTRAINT poll_credit_msgid_fkey REFERENCES message (id),
  zone INTEGER CONSTRAINT poll_credit_zone_fkey REFERENCES zone (id),
  credlimit numeric(10,2) NOT NULL,
  credit numeric(10,2) NOT NULL
);

CREATE TABLE poll_credit_zone_limit (
  zone INTEGER CONSTRAINT poll_credit_zone_limit_pkey PRIMARY KEY
  CONSTRAINT poll_credit_zone_limit_zone_fkey REFERENCES zone(id),
  credlimit numeric(10,2) NOT NULL
);

CREATE TABLE poll_eppaction (
  msgid INTEGER CONSTRAINT poll_eppaction_pkey PRIMARY KEY
  CONSTRAINT poll_eppaction_msgid_fkey REFERENCES message (id),
  objid INTEGER CONSTRAINT poll_eppaction_objid_fkey REFERENCES object_history (historyid)
);

CREATE TABLE poll_techcheck (
  msgid INTEGER CONSTRAINT poll_techcheck_pkey PRIMARY KEY
  CONSTRAINT poll_techcheck_msgid_fkey REFERENCES message (id)
);

CREATE TABLE poll_stateChange (
  msgid INTEGER CONSTRAINT poll_statechange_pkey PRIMARY KEY
  CONSTRAINT poll_statechange_msgid_fkey REFERENCES message (id),
  stateid INTEGER CONSTRAINT poll_statechange_stateid_fkey REFERENCES object_state (id)
);

CREATE INDEX poll_statechange_stateid_idx ON poll_statechange (stateid);

CREATE TABLE epp_transfer_request(
  id serial primary key,
  domain_id bigint CONSTRAINT epp_transfer_domainid_fkey REFERENCES object_registry(id) ON DELETE CASCADE,
  registrar_id integer CONSTRAINT epp_transfer_regid_fkey REFERENCES Registrar,
  acquirer_id integer CONSTRAINT epp_transfer_acid_fkey REFERENCES Registrar,
  upid integer CONSTRAINT epp_transfer_upid_fkey REFERENCES Registrar, /* id of the registrar that changed the state of the request */
  status integer,
  created timestamp NOT NULL DEFAULT now(),
  acdate timestamp NOT NULL DEFAULT now() + '90 days'::interval
);

CREATE INDEX epp_transfer_request_idx ON epp_transfer_request (domain_id, created);

CREATE TABLE epp_transfer_request_state_change(
  request_id bigint CONSTRAINT epp_transfer_state_change_req_fkey REFERENCES epp_transfer_request (id) ON DELETE CASCADE,
  msgid bigint CONSTRAINT epp_transfer_state_change_msgid_fkey REFERENCES message (id),
  status integer
);

CREATE INDEX epp_transfer_request_sc_idx ON epp_transfer_request_state_change (msgid);

CREATE TABLE enum_transfer_states (
  -- id of status
  id INTEGER CONSTRAINT enum_transfer_states_pkey PRIMARY KEY,
  -- code name for status
  name VARCHAR(50) NOT NULL
);

INSERT INTO enum_transfer_states VALUES(0, 'pending');
INSERT INTO enum_transfer_states VALUES(1, 'clientCancelled');
INSERT INTO enum_transfer_states VALUES(2, 'clientRejected');
INSERT INTO enum_transfer_states VALUES(3, 'clientApproved');
INSERT INTO enum_transfer_states VALUES(5, 'serverCancelled');

-- delete messages on delete from epp_transfer_request
CREATE OR REPLACE FUNCTION before_delete_transfer_request() RETURNS TRIGGER AS
$BODY$
BEGIN
    UPDATE message SET seen = 't' WHERE id in (SELECT msgid FROM epp_transfer_request_state_change WHERE request_id = OLD.id);
    RETURN OLD;
END;
$BODY$
language plpgsql;

CREATE TRIGGER epp_transfer_request_delete
     BEFORE DELETE ON epp_transfer_request
     FOR EACH ROW
     EXECUTE PROCEDURE before_delete_transfer_request();

