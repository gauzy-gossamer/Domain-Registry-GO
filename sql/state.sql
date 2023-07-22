
-- all supported status types
CREATE TABLE enum_object_states (
  -- id of status
  id INTEGER CONSTRAINT enum_object_states_pkey PRIMARY KEY,
  -- code name for status
  name VARCHAR(50) NOT NULL,
  -- what types of objects can have this status (object_registry.type list)
  types INTEGER[] NOT NULL,
  -- if this status is set manualy
  manual BOOLEAN NOT NULL,
  -- if this status can be used be the client
  client_state BOOLEAN NOT NULL default 'f',
  -- if this status is exported to public
  external BOOLEAN NOT NULL,
  -- status importance (ticket #8289)
  importance INTEGER
  CONSTRAINT name_delimiter_check CHECK (name::text !~~ '%,%'::text)
);

comment on table enum_object_states is 'list of all supported status types';
comment on column enum_object_states.name is 'code name for status';
comment on column enum_object_states.types is 'what types of objects can have this status (object_registry.type list)';
comment on column enum_object_states.manual is 'if this status is set manualy';
comment on column enum_object_states.external is 'if this status is exported to public';

--
-- Watch out! Do not use chars "#" and "&" in the text. They are used as delimiters in get_state_descriptions().
--
INSERT INTO enum_object_states
  VALUES (01,'serverDeleteProhibited','{1,2,3}','t','f','t', 32768);
INSERT INTO enum_object_states
  VALUES (02,'serverRenewProhibited','{3}','t','f','t', 16384);
INSERT INTO enum_object_states
  VALUES (03,'serverTransferProhibited','{1,2,3}','t','f','t', 4096);
INSERT INTO enum_object_states
  VALUES (04,'serverUpdateProhibited','{1,2,3}','t','f','t', 2048);
INSERT INTO enum_object_states
  VALUES (05,'serverOutzoneManual','{3}','t','f','f', 128);
INSERT INTO enum_object_states
  VALUES (06,'serverInzoneManual','{3}','t','f','t', 256);
INSERT INTO enum_object_states
  VALUES (07,'serverBlocked','{1,3}','t','f','t', 65536);
INSERT INTO enum_object_states
  VALUES (08,'expirationWarning','{3}','f','f','f', NULL);
INSERT INTO enum_object_states
  VALUES (09,'expired','{3}','f','f','f', 2);
INSERT INTO enum_object_states
  VALUES (10,'unguarded','{3}','f','f','f', NULL);
INSERT INTO enum_object_states
  VALUES (11,'validationWarning1','{3}','f','f','f', NULL);
INSERT INTO enum_object_states
  VALUES (12,'validationWarning2','{3}','f','f','f', NULL);
INSERT INTO enum_object_states
  VALUES (13,'notValidated','{3}','f','f','t', 512);
INSERT INTO enum_object_states
  VALUES (14,'nssetMissing','{3}','f','f','f', NULL);
INSERT INTO enum_object_states
  VALUES (15,'inactive','{3}','f','f','t', 8);
INSERT INTO enum_object_states
  VALUES (16,'linked','{1,2}','f','f','t', 1024);
INSERT INTO enum_object_states
  VALUES (17,'pendingDelete','{1,2,3}','f','f','f', NULL);
INSERT INTO enum_object_states
  VALUES (18,'serverRegistrantChangeProhibited','{3}','t','f','t', 8192);
INSERT INTO enum_object_states
  VALUES (19,'deleteWarning','{3}','f','f','f', NULL);
INSERT INTO enum_object_states
  VALUES (20,'outzoneUnguarded','{3}','f','f','f', NULL);
INSERT INTO enum_object_states
  VALUES (28,'outzoneUnguardedWarning','{3}','f','f','f', NULL);
INSERT INTO enum_object_states
  VALUES (29,'clientDeleteProhibited','{1,2,3,4}','t','t','t', NULL);
INSERT INTO enum_object_states
  VALUES (30,'clientUpdateProhibited','{1,2,3,4}','t','t','t', NULL);
INSERT INTO enum_object_states
  VALUES (31,'clientTransferProhibited','{1,3}','t','t','t', NULL);
INSERT INTO enum_object_states
  VALUES (32,'clientRenewProhibited','{3}','t','t','t', NULL);
INSERT INTO enum_object_states
  VALUES (33,'clientHold','{3}','t','t','t', NULL);
INSERT INTO enum_object_states
  VALUES (34,'pendingTransfer','{3}','t','f','t', NULL);
INSERT INTO enum_object_states
  VALUES (35,'changeProhibited','{3}','t','f','t', NULL);


-- update for keyset
update enum_object_states set types = types || array[4] where 2 = any (types);

-- descriptions fo states in different languages 
CREATE TABLE enum_object_states_desc (
  -- id of status
  state_id INTEGER NOT NULL CONSTRAINT enum_object_states_desc_state_id_fkey REFERENCES enum_object_states (id),
  -- char code of language
  lang CHAR(2) NOT NULL,
  -- descriptive text
  description VARCHAR(255),
  CONSTRAINT enum_object_states_desc_pkey PRIMARY KEY (state_id,lang)
);

comment on table enum_object_states_desc is 'description for states in different languages';
comment on column enum_object_states_desc.lang is 'code of language';
comment on column enum_object_states_desc.description is 'descriptive text';

--
-- Watch out! Do not use chars "#" and "&" in the text. They are used as delimiters in get_state_descriptions().
--
INSERT INTO enum_object_states_desc 
  VALUES (01,'EN','Deletion forbidden');
INSERT INTO enum_object_states_desc 
  VALUES (02,'EN','Registration renewal forbidden');
INSERT INTO enum_object_states_desc 
  VALUES (03,'EN','Sponsoring registrar change forbidden');
INSERT INTO enum_object_states_desc 
  VALUES (04,'EN','Update forbidden');
INSERT INTO enum_object_states_desc 
  VALUES (05,'EN','The domain is administratively kept out of zone');
INSERT INTO enum_object_states_desc 
  VALUES (06,'EN','The domain is administratively kept in zone');
INSERT INTO enum_object_states_desc 
  VALUES (07,'EN','Domain blocked');
INSERT INTO enum_object_states_desc 
  VALUES (08,'EN','The domain expires in 30 days');
INSERT INTO enum_object_states_desc 
  VALUES (09,'EN','Domain expired');
INSERT INTO enum_object_states_desc 
  VALUES (10,'EN','The domain is 30 days after expiration');
INSERT INTO enum_object_states_desc 
  VALUES (11,'EN','The domain validation expires in 30 days');
INSERT INTO enum_object_states_desc 
  VALUES (12,'EN','The domain validation expires in 15 days');
INSERT INTO enum_object_states_desc 
  VALUES (13,'EN','Domain not validated');
INSERT INTO enum_object_states_desc 
  VALUES (14,'EN','The domain doesn''t have associated nsset');
INSERT INTO enum_object_states_desc 
  VALUES (15,'EN','The domain isn''t generated in the zone');
INSERT INTO enum_object_states_desc 
  VALUES (16,'EN','Has relation to other records in the registry');
INSERT INTO enum_object_states_desc 
  VALUES (17,'EN','To be deleted');
INSERT INTO enum_object_states_desc 
  VALUES (18,'EN','Registrant change forbidden');
INSERT INTO enum_object_states_desc 
  VALUES (19,'EN','The domain will be deleted in 11 days');
INSERT INTO enum_object_states_desc 
  VALUES (20,'EN','The domain is out of zone after 30 days in expiration state');
INSERT INTO enum_object_states_desc
  VALUES (28,'EN','The domain is to be out of zone soon');
INSERT INTO enum_object_states_desc
  VALUES (29,'EN','Client delete prohibited');
INSERT INTO enum_object_states_desc
  VALUES (30,'EN','Client update prohibited');
INSERT INTO enum_object_states_desc
  VALUES (31,'EN','Client transfer prohibited');
INSERT INTO enum_object_states_desc
  VALUES (32,'EN','Client renew prohibited');
INSERT INTO enum_object_states_desc
  VALUES (33,'EN','Client hold');
INSERT INTO enum_object_states_desc
  VALUES (34,'EN','Pending transfer');
INSERT INTO enum_object_states_desc
  VALUES (35,'EN','Change prohibited');


-- main table of object states and their changes
CREATE TABLE object_state (
  -- id of moment when object gain status
  id SERIAL CONSTRAINT object_state_pkey PRIMARY KEY,
  -- id of object that has this new status
  object_id INTEGER NOT NULL CONSTRAINT object_state_object_id_fkey REFERENCES object_registry (id),
  -- id of status
  state_id INTEGER NOT NULL CONSTRAINT object_state_state_id_fkey REFERENCES enum_object_states (id),
  -- timestamp when object entered state
  valid_from TIMESTAMP NOT NULL,
  -- timestamp when object leaved state or null if still has status
  valid_to TIMESTAMP,
  -- history id of object in the moment of entering state  (may be NULL)
  ohid_from INTEGER CONSTRAINT object_state_ohid_from_fkey REFERENCES object_history (historyid),
  -- history id of object in the moment of leaving state or null
  ohid_to INTEGER CONSTRAINT object_state_ohid_to_fkey REFERENCES object_history (historyid)
);

CREATE UNIQUE INDEX object_state_now_idx ON object_state (object_id, state_id)
WHERE valid_to ISNULL;

CREATE INDEX object_state_object_id_idx ON object_state (object_id) WHERE valid_to ISNULL;
CREATE INDEX object_state_object_id_all_idx ON object_state (object_id);
CREATE INDEX object_state_valid_from_idx ON object_state (valid_from);
CREATE INDEX object_state_valid_to_idx ON object_state (valid_to);

comment on table object_state is 'main table of object states and their changes';
comment on column object_state.object_id is 'id of object that has this new status';
comment on column object_state.state_id is 'id of status';
comment on column object_state.valid_from is 'date and time when object entered state';
comment on column object_state.valid_to is 'date and time when object leaved state or null if still has this status';
comment on column object_state.ohid_from is 'history id of object in the moment of entering state (may be null)';
comment on column object_state.ohid_to is 'history id of object in the moment of leaving state or null';

-- aggregate function for accumulation of elements into array
CREATE AGGREGATE array_accum (
  BASETYPE = anycompatible,
  sfunc = array_append,
  stype = anycompatiblearray,
  initcond = '{}'
);

-- simple view for all active states of object
CREATE VIEW object_state_now AS
SELECT object_id, array_accum(state_id) AS states
FROM object_state
WHERE valid_to ISNULL
GROUP BY object_id;

-- request for setting manual state
CREATE TABLE object_state_request (
  -- id of request
  id SERIAL CONSTRAINT object_state_request_pkey PRIMARY KEY,
  -- id of object gaining requested state
  object_id INTEGER NOT NULL CONSTRAINT object_state_request_object_id_fkey REFERENCES object_registry (id),
  -- id of requested state
  state_id INTEGER NOT NULL CONSTRAINT object_state_request_state_id_fkey REFERENCES enum_object_states (id),
  -- when object should enter requested state
  valid_from TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  -- when object should leave requested state
  valid_to TIMESTAMP,
  -- could be pointer to some list of administration actions
  crdate TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP, 
  -- could be pointer to some list of administration actions
  canceled TIMESTAMP 
);

CREATE INDEX object_state_request_object_id_idx ON object_state_request (object_id);

comment on table object_state_request is 'request for setting manual state';
comment on column object_state_request.object_id is 'id of object gaining request state';
comment on column object_state_request.state_id is 'id of requested state';
comment on column object_state_request.valid_from is 'when object should enter requested state';
comment on column object_state_request.valid_to is 'when object should leave requested state';
comment on column object_state_request.crdate is 'could be pointed to some list of administation action';
comment on column object_state_request.canceled is 'could be pointed to some list of administation action';

-- simple view for all active requests for state change
CREATE VIEW object_state_request_now AS
SELECT object_id, array_accum(state_id) AS states
FROM object_state_request
WHERE valid_from<=CURRENT_TIMESTAMP 
AND (valid_to ISNULL OR valid_to>=CURRENT_TIMESTAMP) AND canceled ISNULL
GROUP BY object_id;

-- function to test date moved by offset agains current date
CREATE OR REPLACE FUNCTION date_month_test(date, varchar, varchar, varchar)
RETURNS boolean
AS $$
SELECT $1 + ($2||' month')::interval + ($3||' hours')::interval 
       <= CURRENT_TIMESTAMP AT TIME ZONE $4;
$$ IMMUTABLE LANGUAGE SQL;

-- function to test date moved by offset agains current date
CREATE OR REPLACE FUNCTION date_test(date, varchar)
RETURNS boolean
AS $$
SELECT $1 + ($2||' days')::interval <= CURRENT_DATE ;
$$ IMMUTABLE LANGUAGE SQL;

-- function to test date moved by offset agains current date with respect
-- to current time and time zone
CREATE OR REPLACE FUNCTION date_time_test(date, varchar, varchar, varchar)
RETURNS boolean
AS $$
SELECT $1 + ($2||' days')::interval + ($3||' hours')::interval 
       <= CURRENT_TIMESTAMP AT TIME ZONE $4;
$$ IMMUTABLE LANGUAGE SQL;

-- function to test timestamp
CREATE OR REPLACE FUNCTION timestamp_test(timestamp, varchar, varchar, varchar)
RETURNS boolean
AS $$
SELECT $1 + ($2||' days')::interval + ($3||' hours')::interval
       <= CURRENT_TIMESTAMP AT TIME ZONE $4;
$$ IMMUTABLE LANGUAGE SQL;


-- function to test expired date
CREATE OR REPLACE FUNCTION expired_date_test(date, varchar, varchar)
RETURNS boolean
AS $$
SELECT $1 + ($2||' days')::interval <= (SELECT max(allow_date) FROM domain_allow_removal_dates WHERE allow_date <= CURRENT_DATE AT TIME ZONE $3);
$$ IMMUTABLE LANGUAGE SQL;

CREATE OR REPLACE FUNCTION is_subordinate(host text, domain text) RETURNS boolean as $$
    DECLARE
        l integer;
    BEGIN
        l := length(domain);
        IF length(host) = l THEN
            RETURN host = domain;
        ELSE
            RETURN right(host, l + 1) = '.' || domain;
        END IF;
    END
$$ language plpgsql;

-- ========== 1. type ==============================
-- following views and one setting function is for global setting of all
-- states. there are view pro each type of object to catch states of ALL
-- objects. then function is defined that compare these states with
-- actual state and update states appropriatly. That function is long
-- running so it cannot be used for updating states online. That case
-- is handled in 2. type of functions 

-- view for actual domain states
-- ================= DOMAIN ========================
-- DROP VIEW domain_states;
CREATE VIEW public.domain_states AS
 SELECT d.id AS object_id,
    o.historyid AS object_hid,
    ((((COALESCE(osr.states, '{}'::integer[]) ||
        CASE
            WHEN public.timestamp_test(d.exdate, '0'::character varying, '0'::character varying, 'UTC'::character varying) THEN ARRAY[9, 1, 3] -- expired
            ELSE '{}'::integer[]
        END) ||
        CASE
            WHEN (NOT public.date_time_test((d.exdate)::date, ep_ex_not.val, '0'::character varying, ep_tz.val)) THEN ARRAY[2] -- serverRenewProhibited
            ELSE '{}'::integer[]
        END) ||
        CASE
            WHEN (((   SELECT count(*) AS hostsn
               FROM (public.host h
                 JOIN public.domain_host_map dhm ON (((h.hostid = dhm.hostid) AND
                    CASE
                        WHEN public.is_subordinate(lower((h.fqdn)::text), lower((o.name)::text)) THEN (SELECT count(*) 
                            FROM host_ipaddr_map him WHERE him.hostid=dhm.hostid ) > 0
                        ELSE true 
                    END)))                                                                              
                  WHERE (dhm.domainid = d.id)) < (ep_min_hosts.val)::integer)
                  OR (5 = ANY (COALESCE(osr.states, '{}'::integer[]))) -- serverOutzoneManual
                  OR (33 = ANY (COALESCE(osr.states, '{}'::integer[]))) -- clientHold
                ) THEN ARRAY[15] -- inactive
            ELSE '{}'::integer[]
        END) ||
        CASE
            WHEN (public.expired_date_test((timezone('MSK'::text, timezone('UTC'::text, d.exdate)))::date, ep_ex_reg.val, ep_tz.val) AND (NOT (2 = ANY (COALESCE(osr.states, '{}'::integer[])))) AND (NOT (1 = ANY (COALESCE(osr.states, '{}'::integer[])))) AND (NOT (35 = ANY (COALESCE(osr.states, '{}'::integer[]))))) THEN ARRAY[17, 2, 4]
            ELSE '{}'::integer[]
        END) AS states
   FROM public.object_registry o,
    (((((((((public.domain d
     LEFT JOIN public.object_state_request_now osr ON ((d.id = osr.object_id)))
     JOIN public.enum_parameters ep_ex_not ON ((ep_ex_not.id = 3)))
     JOIN public.enum_parameters ep_ex_dns ON ((ep_ex_dns.id = 4)))
     JOIN public.enum_parameters ep_ex_reg ON ((ep_ex_reg.id = 6)))
     JOIN public.enum_parameters ep_tm ON ((ep_tm.id = 9)))
     JOIN public.enum_parameters ep_tz ON ((ep_tz.id = 10)))
     JOIN public.enum_parameters ep_tm2 ON ((ep_tm2.id = 14)))
     JOIN public.enum_parameters ep_ozu_warn ON ((ep_ozu_warn.id = 18)))
     JOIN public.enum_parameters ep_min_hosts ON ((ep_min_hosts.id = 20)))
  WHERE (d.id = o.id);

-- view for actual nsset states
-- for NOW they are not deleted
-- ================= NSSET ========================
-- DROP VIEW nsset_states;
CREATE VIEW public.nsset_states AS
 SELECT o.id AS object_id,
    o.historyid AS object_hid,
    ((COALESCE(osr.states, '{}'::integer[]) ||
        CASE
            WHEN ((NOT (d.hostid IS NULL))) THEN ARRAY[16]
            ELSE '{}'::integer[]
        END) ||
        CASE
            WHEN ((d.hostid IS NULL) AND public.date_month_test(GREATEST((COALESCE(l.last_linked, o.crdate))::date, (COALESCE(ob.update, o.crdate))::date), ep_mn.val, ep_tm.val, ep_tz.val) AND (NOT (1 = ANY (COALESCE(osr.states, '{}'::integer[]))))) THEN ARRAY[17]
            ELSE '{}'::integer[]
        END) AS states
   FROM (((((((public.object ob
     JOIN public.object_registry o ON (((ob.id = o.id) AND (o.type = 2))))
     JOIN public.enum_parameters ep_tm ON ((ep_tm.id = 9)))
     JOIN public.enum_parameters ep_tz ON ((ep_tz.id = 10)))
     JOIN public.enum_parameters ep_mn ON ((ep_mn.id = 11)))
     LEFT JOIN ( SELECT DISTINCT hostid
           FROM public.domain_host_map) d ON ((d.hostid = o.id)))
     LEFT JOIN ( SELECT object_state.object_id,
            max(object_state.valid_to) AS last_linked
           FROM public.object_state
          WHERE (object_state.state_id = 16)
          GROUP BY object_state.object_id) l ON ((o.id = l.object_id)))
     LEFT JOIN public.object_state_request_now osr ON ((o.id = osr.object_id)));

-- ================= KEYSET ========================
-- DROP VIEW keyset_states;
CREATE VIEW keyset_states AS
SELECT
  o.id AS object_id,
  o.historyid AS object_hid,
  COALESCE(osr.states,'{}') ||
  CASE WHEN NOT(d.keyset ISNULL) THEN ARRAY[16] ELSE '{}' END ||
  CASE WHEN d.keyset ISNULL 
            AND date_month_test(
              GREATEST(
                COALESCE(l.last_linked,o.crdate)::date,
                COALESCE(ob.update,o.crdate)::date
              ),
              ep_mn.val,ep_tm.val,ep_tz.val
            )
            AND NOT (1 = ANY(COALESCE(osr.states,'{}')))
       THEN ARRAY[17] ELSE '{}' END 
  AS states
FROM
  object ob
  JOIN object_registry o ON (ob.id=o.id AND o.type=4)
  JOIN enum_parameters ep_tm ON (ep_tm.id=9)
  JOIN enum_parameters ep_tz ON (ep_tz.id=10)
  JOIN enum_parameters ep_mn ON (ep_mn.id=11)
  LEFT JOIN (
    SELECT DISTINCT keyset FROM domain
  ) AS d ON (d.keyset=o.id)
  LEFT JOIN (
    SELECT object_id, MAX(valid_to) AS last_linked
    FROM object_state
    WHERE state_id=16 GROUP BY object_id
  ) AS l ON (o.id=l.object_id)
  LEFT JOIN object_state_request_now osr ON (o.id=osr.object_id);

-- view for actual contact states
-- ================= CONTACT ========================
-- DROP VIEW contact_states;
CREATE VIEW contact_states AS
SELECT
  o.id AS object_id,
  o.historyid AS object_hid,
  COALESCE(osr.states,'{}') ||
  CASE WHEN NOT(cl.cid ISNULL) THEN ARRAY[16] ELSE '{}' END ||
  CASE WHEN cl.cid ISNULL 
            AND date_month_test(
              GREATEST(
                COALESCE(l.last_linked,o.crdate)::date,
                COALESCE(ob.update,o.crdate)::date
              ),
              ep_mn.val,ep_tm.val,ep_tz.val
            )
            AND NOT (1 = ANY(COALESCE(osr.states,'{}')))
       THEN ARRAY[17] ELSE '{}' END 
  AS states
FROM
  object ob
  JOIN object_registry o ON (ob.id=o.id AND o.type=1)
  JOIN enum_parameters ep_tm ON (ep_tm.id=9)
  JOIN enum_parameters ep_tz ON (ep_tz.id=10)
  JOIN enum_parameters ep_mn ON (ep_mn.id=11)
  LEFT JOIN (
    SELECT registrant AS cid FROM domain
    UNION
    SELECT contactid AS cid FROM domain_contact_map
  ) AS cl ON (o.id=cl.cid)
  LEFT JOIN (
    SELECT object_id, MAX(valid_to) AS last_linked
    FROM object_state
    WHERE state_id=16 GROUP BY object_id
  ) AS l ON (o.id=l.object_id)
  LEFT JOIN object_state_request_now osr ON (o.id=osr.object_id);

-- in next function compared arrays need to be sorted
CREATE OR REPLACE FUNCTION array_sort_dist (ANYARRAY)
RETURNS ANYARRAY LANGUAGE SQL
AS $$
SELECT COALESCE(ARRAY(
    SELECT DISTINCT $1[s.i] AS "sort"
    FROM
        generate_series(array_lower($1,1), array_upper($1,1)) AS s(i)
    ORDER BY sort
),'{}');
$$ IMMUTABLE;

-- ================= UPDATE FUNCTION =================
-- CREATE LANGUAGE plpgsql;
CREATE OR REPLACE FUNCTION update_object_states(int)
RETURNS void
AS $$
BEGIN
  IF NOT EXISTS(
    SELECT relname FROM pg_class
    WHERE relname = 'tmp_object_state_change' AND relkind = 'r' AND
    pg_table_is_visible(oid)
  )
  THEN
    CREATE TEMPORARY TABLE tmp_object_state_change (
      object_id INTEGER,
      object_hid INTEGER,
      new_states INTEGER[],
      old_states INTEGER[]
    );
  ELSE
    TRUNCATE tmp_object_state_change;
  END IF;

  IF $1 = 0
  THEN
    INSERT INTO tmp_object_state_change
    SELECT
      st.object_id, st.object_hid, st.states AS new_states, 
      COALESCE(o.states,'{}') AS old_states
    FROM (
      SELECT * FROM domain_states
      UNION
      SELECT * FROM contact_states
      UNION
      SELECT * FROM nsset_states
      UNION
      SELECT * FROM keyset_states
    ) AS st
    LEFT JOIN object_state_now o ON (st.object_id=o.object_id)
    WHERE array_sort_dist(st.states)!=COALESCE(array_sort_dist(o.states),'{}');
  ELSE
    -- domain
    INSERT INTO tmp_object_state_change
    SELECT
      st.object_id, st.object_hid, st.states AS new_states, 
      COALESCE(o.states,'{}') AS old_states
    FROM domain_states st
    LEFT JOIN object_state_now o ON (st.object_id=o.object_id)
    WHERE array_sort_dist(st.states)!=COALESCE(array_sort_dist(o.states),'{}')
    AND st.object_id=$1;
    -- contact
    INSERT INTO tmp_object_state_change
    SELECT
      st.object_id, st.object_hid, st.states AS new_states, 
      COALESCE(o.states,'{}') AS old_states
    FROM contact_states st
    LEFT JOIN object_state_now o ON (st.object_id=o.object_id)
    WHERE array_sort_dist(st.states)!=COALESCE(array_sort_dist(o.states),'{}')
    AND st.object_id=$1;
    -- nsset
    INSERT INTO tmp_object_state_change
    SELECT
      st.object_id, st.object_hid, st.states AS new_states, 
      COALESCE(o.states,'{}') AS old_states
    FROM nsset_states st
    LEFT JOIN object_state_now o ON (st.object_id=o.object_id)
    WHERE array_sort_dist(st.states)!=COALESCE(array_sort_dist(o.states),'{}')
    AND st.object_id=$1;
    -- keyset
    INSERT INTO tmp_object_state_change
    SELECT
      st.object_id, st.object_hid, st.states AS new_states, 
      COALESCE(o.states,'{}') AS old_states
    FROM keyset_states st
    LEFT JOIN object_state_now o ON (st.object_id=o.object_id)
    WHERE array_sort_dist(st.states)!=COALESCE(array_sort_dist(o.states),'{}')
    AND st.object_id=$1;
  END IF;

  INSERT INTO object_state (object_id,state_id,valid_from,ohid_from)
  SELECT c.object_id,e.id,CURRENT_TIMESTAMP,c.object_hid
  FROM tmp_object_state_change c, enum_object_states e
  WHERE e.id = ANY(c.new_states) AND e.id != ALL(c.old_states);

  UPDATE object_state SET valid_to=CURRENT_TIMESTAMP, ohid_to=c.object_hid
  FROM enum_object_states e, tmp_object_state_change c
  WHERE e.id = ANY(c.old_states) AND e.id != ALL(c.new_states)
  AND e.id=object_state.state_id and c.object_id=object_state.object_id 
  AND object_state.valid_to ISNULL;
END;
$$ LANGUAGE plpgsql;

-- ========== 2. type ==============================
-- following functions and triggers are for automatic state update
-- as a reaction on normal EPP commands.

-- RACE CONTIONS! :
-- there are some places where race conditions can play a role. this is
-- because of sharing of nsset and contacts. their status 'linked' is
-- updated when they appear in EPP commands on others object and because
-- of parallel execution of these commands, linked status is 'shared resource'.
--  1. setting linked state. situation when two transactions want to add 
--     linked state is handled by UNIQUE constraint and EXCEPTION 
--     catching in status_set_state function. second setting is let to fail
--     alternative is to lock all table with states which is inefficient
--  2. clearing linked state. when test is done to clear a state, there
--     can appear transaction that change results of this before clear so
--     test and clear need to be atomic. this must be done by row locking
--     of linked state row in states table. this is done by function
--     status_clear_lock.
CREATE OR REPLACE FUNCTION status_clear_lock(integer, integer)
RETURNS boolean
AS $$
SELECT id IS NOT NULL FROM object_state 
WHERE object_id=$1 AND state_id=$2 AND valid_to IS NULL FOR UPDATE;
$$ LANGUAGE SQL;

-- set state of object if condition is satisfied. use optimistic access
-- so there is no need for locking, could be enhanced by checking of 
-- presence of state (to minimize number of failures in unique constraint)
-- but expectation is that this is faster
-- this function only set state, it won't clean it (see next function) 
CREATE OR REPLACE FUNCTION status_set_state(
  _cond BOOL, _state_id INTEGER, _object_id INTEGER
) RETURNS VOID AS $$
 BEGIN
   IF _cond THEN
     -- optimistic access, don't check if status exists
     -- but may fail on UNIQUE constraint, so catching exception 
     INSERT INTO object_state (object_id, state_id, valid_from)
     VALUES (_object_id, _state_id, CURRENT_TIMESTAMP);
   END IF;
 EXCEPTION
   WHEN UNIQUE_VIOLATION THEN
   -- do nothing
 END;
$$ LANGUAGE plpgsql;

-- clear state of object
CREATE OR REPLACE FUNCTION status_clear_state(
  _cond BOOL, _state_id INTEGER, _object_id INTEGER
) RETURNS VOID AS $$
 BEGIN
   IF NOT _cond THEN
     -- condition (valid_to IS NULL) is essential to avoid closing closed
     -- state
     UPDATE object_state SET valid_to = CURRENT_TIMESTAMP
     WHERE state_id = _state_id AND valid_to IS NULL 
     AND object_id = _object_id;
   END IF;
 END;
$$ LANGUAGE plpgsql;

-- set state of object and clear state of object according to condition
CREATE OR REPLACE FUNCTION status_update_state(
  _cond BOOL, _state_id INTEGER, _object_id INTEGER
) RETURNS VOID AS $$
 DECLARE
   _num INTEGER;
 BEGIN
   -- don't know if it's faster to not test condition twise or call EXECUTE
   -- that immidietely return (removing IF), guess is twice test is faster
   IF _cond THEN
     EXECUTE status_set_state(_cond, _state_id, _object_id);
   ELSE
     EXECUTE status_clear_state(_cond, _state_id, _object_id);
   END IF;
 END;
$$ LANGUAGE plpgsql;

-- trigger function to handle dependant state. some states depends on others
-- so trigger is fired on state table after every change and
-- dependant states are updated aproprietly
CREATE OR REPLACE FUNCTION status_update_object_state() RETURNS TRIGGER AS $$
  DECLARE
    _states INTEGER[];
  BEGIN
    IF NEW.state_id = ANY (ARRAY[5,6,10,13,14]) THEN
      -- activation is only done on states that are relevant for
      -- dependant states to stop RECURSION
      SELECT array_accum(state_id) INTO _states FROM object_state
          WHERE valid_to IS NULL AND object_id = NEW.object_id;
      -- set or clear status 15 (outzone)
      EXECUTE status_update_state(
        (14 = ANY (_states)) OR -- nsset is null
        (5  = ANY (_states)) OR -- serverOutzoneManual
        (NOT (6 = ANY (_states)) AND -- not serverInzoneManual
          ((10 = ANY (_states)) OR -- unguarded
           (13 = ANY (_states)))),  -- not validated
        15, NEW.object_id -- => set outzone
      );
      -- set or clear status 15 (outzoneUnguarded)
      EXECUTE status_update_state(
        NOT (6 = ANY (_states)) AND -- not serverInzoneManual
            (10 = ANY (_states)), -- unguarded
        20, NEW.object_id -- => set outzoneUnguarded
      );
    END IF;
    RETURN NEW;
  END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_object_state AFTER INSERT OR UPDATE
  ON object_state FOR EACH ROW EXECUTE PROCEDURE status_update_object_state();

-- update history id of object at status opening and closing
-- it's useful to catch history id of object on state opening and closing
-- to centralize this setting, it's done by trigger here
CREATE OR REPLACE FUNCTION status_update_hid() RETURNS TRIGGER AS $$
  BEGIN
    IF TG_OP = 'UPDATE' AND NEW.ohid_to ISNULL THEN
      SELECT historyid INTO NEW.ohid_to 
      FROM object_registry WHERE id=NEW.object_id;
    ELSE IF TG_OP = 'INSERT' AND NEW.ohid_from ISNULL THEN
        SELECT historyid INTO NEW.ohid_from 
        FROM object_registry WHERE id=NEW.object_id;
      END IF;
    END IF;
    RETURN NEW;
  END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_object_state_hid BEFORE INSERT OR UPDATE
  ON object_state FOR EACH ROW EXECUTE PROCEDURE status_update_hid();

-- trigger to update state of domain, fired with every change on
-- on domain table
CREATE OR REPLACE FUNCTION status_update_domain() RETURNS TRIGGER AS $$
  DECLARE
    _num INTEGER;
    _registrant_old INTEGER;
    _keyset_old INTEGER;
    _registrant_new INTEGER;
    _keyset_new INTEGER;
    _ex_not VARCHAR;
    _ex_dns VARCHAR;
--    _ex_reg VARCHAR;
    _proc_tm VARCHAR;
    _proc_tz VARCHAR;
    _proc_tm2 VARCHAR;
    _ou_warn VARCHAR;
    _states INTEGER[];
  BEGIN
    -- skip inflicted status changes on functional tests
    IF (current_setting('myapp.user', TRUE) = 'tester') THEN
      RETURN NULL;
    END IF;

    _registrant_old := NULL;
    _keyset_old := NULL;
    _registrant_new := NULL;
    _keyset_new := NULL;
    SELECT val INTO _ex_not FROM enum_parameters WHERE id=3;
    SELECT val INTO _ex_dns FROM enum_parameters WHERE id=4;
--    SELECT val INTO _ex_reg FROM enum_parameters WHERE id=6;
    SELECT val INTO _proc_tm FROM enum_parameters WHERE id=9;
    SELECT val INTO _proc_tz FROM enum_parameters WHERE id=10;
    SELECT val INTO _proc_tm2 FROM enum_parameters WHERE id=14;
    SELECT val INTO _ou_warn FROM enum_parameters WHERE id=18;
    -- is it INSERT operation
    IF TG_OP = 'INSERT' THEN
      _registrant_new := NEW.registrant;
      _keyset_new := NEW.keyset;

    -- is it UPDATE operation
    ELSIF TG_OP = 'UPDATE' THEN
      IF NEW.registrant <> OLD.registrant THEN
        _registrant_old := OLD.registrant;
        _registrant_new := NEW.registrant;
      END IF;
      IF COALESCE(NEW.keyset,0) <> COALESCE(OLD.keyset,0) THEN
        _keyset_old := OLD.keyset;
        _keyset_new := NEW.keyset;
      END IF;
      -- take care of all domain statuses
      IF NEW.exdate <> OLD.exdate THEN
        -- at the first sight it seems that there should be checking
        -- for renewProhibited state before setting all of these states
        -- as it's done in global (1. type) views
        -- but the point is that when renewProhibited is set
        -- there is no way to change exdate so this situation can never happen 

        -- state: expired
        EXECUTE status_update_state(
          date_test(NEW.exdate::date,'0'),
          9, NEW.id
        );
      END IF; -- change in exdate
    -- is it DELETE operation
    ELSIF TG_OP = 'DELETE' THEN
      _registrant_old := OLD.registrant;
      _keyset_old := OLD.keyset; -- may be NULL!
      -- exdate is meaningless when deleting (probably)
    END IF;

    -- add registrant's linked status if there is none
    EXECUTE status_set_state(
      _registrant_new IS NOT NULL, 16, _registrant_new
    );
    -- add keyset's linked status if there is none
    EXECUTE status_set_state(
      _keyset_new IS NOT NULL, 16, _keyset_new
    );
    -- remove registrant's linked status if not bound
    -- locking must be done (see comment above)
    IF _registrant_old IS NOT NULL AND 
       status_clear_lock(_registrant_old, 16) IS NOT NULL 
    THEN
      SELECT count(*) INTO _num FROM domain
          WHERE registrant = OLD.registrant;
      IF _num = 0 THEN
        SELECT count(*) INTO _num FROM domain_contact_map
            WHERE contactid = OLD.registrant;
        IF _num = 0 THEN
            EXECUTE status_clear_state(_num <> 0, 16, OLD.registrant);
        END IF;
      END IF;
    END IF;
    -- remove keyset's linked status if not bound
    -- locking must be done (see comment above)
    IF _keyset_old IS NOT NULL AND
       status_clear_lock(_keyset_old, 16) IS NOT NULL  
    THEN
      SELECT count(*) INTO _num FROM domain WHERE keyset = OLD.keyset;
      EXECUTE status_clear_state(_num <> 0, 16, OLD.keyset);
    END IF;
    RETURN NULL;
  END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_domain AFTER INSERT OR DELETE OR UPDATE
  ON domain FOR EACH ROW EXECUTE PROCEDURE status_update_domain();

CREATE OR REPLACE FUNCTION status_update_enumval() RETURNS TRIGGER AS $$
  DECLARE
    _num INTEGER;
  BEGIN
    -- is it UPDATE operation
    IF TG_OP = 'UPDATE' AND NEW.exdate <> OLD.exdate THEN
      -- state: validation warning 1
      EXECUTE status_update_state(
        NEW.exdate::date - INTERVAL '30 days' <= CURRENT_DATE, 11, NEW.domainid
      );
      -- state: validation warning 2
      EXECUTE status_update_state(
        NEW.exdate::date - INTERVAL '15 days' <= CURRENT_DATE, 12, NEW.domainid
      );
      -- state: not validated
      EXECUTE status_update_state(
        NEW.exdate::date + INTERVAL '14 hours' <= CURRENT_TIMESTAMP, 13, NEW.domainid
      );
    END IF;
    RETURN NULL;
  END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_enumval AFTER UPDATE ON enumval
    FOR EACH ROW EXECUTE PROCEDURE status_update_enumval();

CREATE OR REPLACE FUNCTION status_update_contact_map() RETURNS TRIGGER AS $$
  DECLARE
    _num INTEGER;
    _contact_old INTEGER;
    _contact_new INTEGER;
  BEGIN
    _contact_old := NULL;
    _contact_new := NULL;
    -- is it INSERT operation
    IF TG_OP = 'INSERT' THEN
      _contact_new := NEW.contactid;
    -- is it UPDATE operation
    ELSIF TG_OP = 'UPDATE' THEN
      IF NEW.contactid <> OLD.contactid THEN
        _contact_old := OLD.contactid;
        _contact_new := NEW.contactid;
      END IF;
    -- is it DELETE operation
    ELSIF TG_OP = 'DELETE' THEN
      _contact_old := OLD.contactid;
    END IF;

    -- add contact's linked status if there is none
    EXECUTE status_set_state(
      _contact_new IS NOT NULL, 16, _contact_new
    );
    -- remove contact's linked status if not bound
    -- locking must be done (see comment above)
    IF _contact_old IS NOT NULL AND
       status_clear_lock(_contact_old, 16) IS NOT NULL 
    THEN
      SELECT count(*) INTO _num FROM domain WHERE registrant = OLD.contactid;
      IF _num = 0 THEN
        SELECT count(*) INTO _num FROM domain_contact_map
            WHERE contactid = OLD.contactid;
        IF _num = 0 THEN
            EXECUTE status_clear_state(_num <> 0, 16, OLD.contactid);
        END IF;
      END IF;
    END IF;
    RETURN NULL;
  END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_domain_contact_map AFTER INSERT OR DELETE OR UPDATE
  ON domain_contact_map FOR EACH ROW 
  EXECUTE PROCEDURE status_update_contact_map();

-- object history tables are filled after normal object tables (i.e. domain)
-- and so when new state is generated as result of new row in normal 
-- table, no history table is available to reference in ohid_from
-- this trigger fix this be filling unfilled ohid_from after insert
-- into history
CREATE OR REPLACE FUNCTION object_history_insert() RETURNS TRIGGER AS $$
  BEGIN
    UPDATE object_state SET ohid_from=NEW.historyid 
    WHERE ohid_from ISNULL AND object_id=NEW.id;
    RETURN NEW;
  END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_object_history AFTER INSERT
  ON object_history FOR EACH ROW 
  EXECUTE PROCEDURE object_history_insert();

---
--- Ticket #7122 - lock object_state_request for manual state insert or update by state and object to the end of db transaction
--- Ticket #8135 - lock object_state_request for state insert or update by object to the end of db transaction
---

CREATE TABLE object_state_request_lock
(
    object_id integer PRIMARY KEY --REFERENCES object_registry (id) --temporarily commented out because of a lack of time to test implicit locking of object_registry row
);

CREATE OR REPLACE FUNCTION lock_object_state_request_lock(f_object_id BIGINT)
RETURNS void AS $$
DECLARE
BEGIN
    INSERT INTO object_state_request_lock (object_id) VALUES (f_object_id);
    DELETE FROM object_state_request_lock WHERE object_id = f_object_id;
END;
$$ LANGUAGE plpgsql;


-- lock object_state_request for manual states
CREATE OR REPLACE FUNCTION lock_object_state_request()
RETURNS "trigger" AS $$
DECLARE
BEGIN
  --lock for manual states
  PERFORM * FROM enum_object_states WHERE id = NEW.state_id AND manual = true;
  IF NOT FOUND THEN
    RETURN NEW;
  END IF;
  
  RAISE NOTICE 'lock_object_state_request NEW.id: % NEW.state_id: % NEW.object_id: %'
  , NEW.id, NEW.state_id, NEW.object_id ;
    PERFORM lock_object_state_request_lock(NEW.object_id);

  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION lock_object_state_request()
IS 'lock changes of object state requests by object';

CREATE TRIGGER "trigger_lock_object_state_request"
  AFTER INSERT OR UPDATE ON object_state_request
  FOR EACH ROW EXECUTE PROCEDURE lock_object_state_request();

--
-- Collect states into one string
-- Usage: SELECT get_state_descriptions(53, 'CS');
-- retval: string ca be splited by row delimiter '\n' and column delimiter '\t':
--      [[external, importance, name, description], ...]
-- example: 't\t20\toutzone\tDomain is not generated into zone\nf\t\texpirationWarning\tExpires '
--          'within 30 days\nf\t\tunguarded\tDomain is 30 days after expiration\nf\t\tnssetMissi'
--          'ng\tDomain has not associated nsset'
--
CREATE OR REPLACE FUNCTION get_state_descriptions(object_id BIGINT, lang_code varchar)
RETURNS TEXT
AS $$
SELECT array_to_string(ARRAY((
    SELECT
        array_to_string(ARRAY[eos.external::char,
        COALESCE(eos.importance::varchar, ''),
        eos.name,
        COALESCE(osd.description, '')], E'#')
    FROM object_state os
    JOIN enum_object_states eos ON eos.id = os.state_id
    JOIN enum_object_states_desc osd ON osd.state_id = eos.id AND lang = $2
    WHERE os.object_id = $1
        AND os.valid_from <= CURRENT_TIMESTAMP
        AND (os.valid_to IS NULL OR os.valid_to > CURRENT_TIMESTAMP)
    ORDER BY eos.importance
)), E'&')
$$ LANGUAGE SQL;

-- Reason of state change
DROP TABLE IF EXISTS object_state_request_reason;
CREATE TABLE object_state_request_reason
(
    object_state_request_id INTEGER NOT NULL REFERENCES object_state_request (id),
    -- state created
    reason_creation VARCHAR(300) NULL DEFAULT NULL,
    -- state canceled
    reason_cancellation VARCHAR(300) NULL DEFAULT NULL,
    PRIMARY KEY (object_state_request_id)
);

---
--- Ticket #12846 - constraint contact discloseaddr by conditions in ticket #7495
---
CREATE OR REPLACE FUNCTION contact_discloseaddr_constraint_impl(f_object_id BIGINT)
RETURNS void AS $$
DECLARE
    is_contact_hidden_address_and_nonempty_organization BOOLEAN;
    is_contact_hidden_address_and_not_identified_and_not_validated BOOLEAN;
BEGIN
    -- ids of other object types unconstrained
    RAISE NOTICE 'check contact_discloseaddr_constraint_impl f_object_id: %', f_object_id;

    SELECT (trim(both ' ' from COALESCE(c.organization,'')) <> '') AS have_organization
    , osr.id IS NULL AS not_identified_or_validated_contact
    INTO is_contact_hidden_address_and_nonempty_organization,
    is_contact_hidden_address_and_not_identified_and_not_validated
    FROM contact c
    LEFT JOIN object_state_request osr ON (
        osr.object_id=c.id
        AND osr.state_id IN (SELECT id FROM enum_object_states WHERE name IN ('identifiedContact','validatedContact'))
        -- following condition is from view object_state_request_now used by function update_object_states to set object states
        AND osr.valid_from<=CURRENT_TIMESTAMP AND (osr.valid_to ISNULL OR osr.valid_to>=CURRENT_TIMESTAMP)
        AND osr.canceled IS NULL
    )
    WHERE c.discloseaddress = FALSE
    AND c.id = f_object_id;

    IF is_contact_hidden_address_and_nonempty_organization THEN
        --TODO: RAISE EXCEPTION
        RAISE WARNING 'contact_discloseaddr_constraint_impl failed f_object_id: % can''t have organization', f_object_id;
    END IF;

    IF is_contact_hidden_address_and_not_identified_and_not_validated THEN
        --TODO: RAISE EXCEPTION
        RAISE WARNING 'contact_discloseaddr_constraint_impl failed f_object_id: % have to be identified or validated contact', f_object_id;
    END IF;

END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION contact_discloseaddr_constraint_impl(f_object_id BIGINT)
    IS 'check that contact with hidden addres is identified and have no organization set';

CREATE OR REPLACE FUNCTION object_state_request_discloseaddr_constraint()
RETURNS "trigger" AS $$
BEGIN
    IF TG_OP = 'INSERT' THEN
        PERFORM contact_discloseaddr_constraint_impl(NEW.object_id);
    ELSEIF TG_OP = 'UPDATE' THEN
        PERFORM contact_discloseaddr_constraint_impl(NEW.object_id);
        IF NEW.object_id <> OLD.object_id THEN
            PERFORM contact_discloseaddr_constraint_impl(OLD.object_id);
        END IF;
    ELSEIF TG_OP = 'DELETE' THEN
        PERFORM contact_discloseaddr_constraint_impl(OLD.object_id);
        RETURN OLD;
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE CONSTRAINT TRIGGER "trigger_object_state_request_discloseaddr_constraint"
  AFTER INSERT OR UPDATE OR DELETE ON object_state_request
  DEFERRABLE INITIALLY DEFERRED
  FOR EACH ROW EXECUTE PROCEDURE object_state_request_discloseaddr_constraint();

CREATE OR REPLACE FUNCTION contact_discloseaddr_constraint()
RETURNS "trigger" AS $$
BEGIN
    IF TG_OP = 'INSERT' THEN
        PERFORM contact_discloseaddr_constraint_impl(NEW.id);
    ELSEIF TG_OP = 'UPDATE' THEN
        PERFORM contact_discloseaddr_constraint_impl(NEW.id);
        IF NEW.id <> OLD.id THEN
            PERFORM contact_discloseaddr_constraint_impl(OLD.id);
        END IF;
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE CONSTRAINT TRIGGER "trigger_contact_discloseaddr_constraint"
  AFTER INSERT OR UPDATE ON contact
  DEFERRABLE INITIALLY DEFERRED
  FOR EACH ROW
EXECUTE PROCEDURE contact_discloseaddr_constraint();

