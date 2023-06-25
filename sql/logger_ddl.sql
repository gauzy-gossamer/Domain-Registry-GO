--
--  create temporary table and if temporary table already
--  exists truncate it for immediate usage (used for querying)
--
CREATE OR REPLACE FUNCTION create_tmp_table(tname VARCHAR)
RETURNS VOID AS $$
BEGIN
 EXECUTE 'CREATE TEMPORARY TABLE ' || tname || ' (id BIGINT PRIMARY KEY)';
 EXCEPTION
 WHEN DUPLICATE_TABLE THEN EXECUTE 'TRUNCATE TABLE ' || tname;
END;
$$ LANGUAGE plpgsql;


CREATE TABLE session (
    id bigserial CONSTRAINT session_pkey primary key,
    user_name varchar(255) not null,       -- user name for Webadmin or id from registrar table for EPP
    login_date timestamp not null,
    logout_date timestamp,
        user_id integer
);

CREATE TABLE service (
    id SERIAL CONSTRAINT service_pkey PRIMARY KEY,
    partition_postfix varchar(10) CONSTRAINT service_partition_postfix_key UNIQUE NOT NULL,
    name varchar(64) CONSTRAINT service_name_key UNIQUE NOT NULL
);

CREATE TABLE request_type (
        id SERIAL CONSTRAINT request_type_pkey PRIMARY KEY,
        name varchar(64) NOT NULL,
        service_id integer NOT NULL CONSTRAINT request_type_service_id_fkey REFERENCES service(id)
);
ALTER TABLE request_type ADD CONSTRAINT request_type_name_service_id_key UNIQUE(name, service_id);

CREATE TABLE result_code (
    id SERIAL CONSTRAINT result_code_pkey PRIMARY KEY,
    service_id INTEGER CONSTRAINT result_code_service_id_fkey REFERENCES service(id),
    result_code INTEGER NOT NULL,
    name VARCHAR(64) NOT NULL
);

CREATE TABLE request_object_type (
    id SERIAL CONSTRAINT request_object_type_pkey PRIMARY KEY,
    name VARCHAR(64)
);


ALTER TABLE result_code ADD CONSTRAINT result_code_unique_code  UNIQUE (service_id, result_code );
ALTER TABLE result_code ADD CONSTRAINT result_code_unique_name  UNIQUE (service_id, name );

COMMENT ON TABLE result_code IS 'all possible operation result codes';
COMMENT ON COLUMN result_code.id IS 'result_code id';
COMMENT ON COLUMN result_code.service_id IS 'reference to service table. This is needed to distinguish entries with identical result_code values';
COMMENT ON COLUMN result_code.result_code IS 'result code as returned by the specific service, it''s only unique within the service';
COMMENT ON COLUMN result_code.name IS 'short name for error (abbreviation) written in camelcase';

-- for CloseRequest result_code_id updates, exception commented out until request.result_code_id optional
CREATE OR REPLACE FUNCTION get_result_code_id( integer, integer)
RETURNS integer AS $$
DECLARE
    result_code_id INTEGER;
BEGIN

    SELECT id FROM result_code INTO result_code_id
        WHERE service_id=$1 and result_code=$2 ;

    IF result_code_id is null THEN
        RAISE WARNING 'result_code.id not found for service_id=% and result_code=% ', $1, $2;
    END IF;
    RETURN result_code_id;
END;
$$ LANGUAGE plpgsql;

CREATE TABLE request (
    id BIGSERIAL CONSTRAINT request_pkey PRIMARY KEY,
    time_begin timestamp NOT NULL,    -- begin of the transaction
    time_end timestamp,        -- end of transaction, it is set if the information is complete
                    -- e.g. if an error message from backend is successfully logged, it's still set
                    -- NULL in cases like crash of the server
    source_ip INET,
    service_id integer NOT NULL CONSTRAINT request_service_id_fkey REFERENCES service(id),   -- service_id code - enum LogServiceType in IDL
    request_type_id integer CONSTRAINT request_request_type_id_fkey REFERENCES request_type(id) DEFAULT 1000,
    session_id  bigint,            --  REFERENCES session(id),
        user_name varchar(255),         -- name of the user who issued the request (from session table)

    is_monitoring boolean NOT NULL,
    result_code_id INTEGER,
        user_id INTEGER
);

CREATE TABLE request_object_ref (
    id BIGSERIAL CONSTRAINT request_object_ref_pkey PRIMARY KEY,
    request_time_begin TIMESTAMP NOT NULL,
    request_service_id INTEGER  NOT NULL,
    request_monitoring BOOLEAN NOT NULL,
    request_id BIGINT NOT NULL CONSTRAINT request_object_ref_request_id_fkey REFERENCES request(id),

    object_type_id INTEGER  NOT NULL CONSTRAINT request_object_ref_object_type_id_fkey REFERENCES request_object_type(id),
    object_id INTEGER NOT NULL
);


ALTER TABLE request ADD CONSTRAINT request_result_code_id_fkey FOREIGN KEY (result_code_id) REFERENCES result_code(id);

COMMENT ON COLUMN request.result_code_id IS 'result code as returned by the specific service, it''s only unique within the service';

CREATE TABLE request_data (
        id BIGSERIAL CONSTRAINT request_data_pkey PRIMARY KEY,
    request_time_begin timestamp NOT NULL, -- TEMP: for partitioning
    request_service_id integer NOT NULL, -- TEMP: for partitioning
    request_monitoring boolean NOT NULL, -- TEMP: for partitioning

    request_id bigint NOT NULL CONSTRAINT request_data_request_id_fkey REFERENCES request(id),
    content text NOT NULL,
    is_response boolean DEFAULT False -- true if the content is response, false if it's request
);

CREATE TABLE request_property_name (
    id SERIAL CONSTRAINT request_property_name_pkey PRIMARY KEY,
    name varchar(256) CONSTRAINT request_property_name_name_key UNIQUE NOT NULL
);

CREATE TABLE request_property_value (
    request_time_begin timestamp NOT NULL, -- TEMP: for partitioning
    request_service_id integer NOT NULL, -- TEMP: for partitioning
    request_monitoring boolean NOT NULL, -- TEMP: for partitioning

    id BIGSERIAL CONSTRAINT request_property_value_pkey PRIMARY KEY,
    request_id bigint NOT NULL CONSTRAINT request_property_value_request_id_fkey REFERENCES request(id),
    property_name_id integer NOT NULL CONSTRAINT request_property_value_property_name_id_fkey REFERENCES request_property_name(id),
    value text NOT NULL,        -- property value
    output boolean DEFAULT False,        -- whether it's output (response) property; if False it's input (request)

    parent_id bigint CONSTRAINT request_property_value_parent_id_fkey REFERENCES request_property_value(id)
                        -- in case of child property, the id of the parent, NULL otherwise
);

CREATE INDEX request_time_begin_idx ON request(time_begin);
CREATE INDEX request_time_end_idx ON request(time_end);
CREATE INDEX request_source_ip_idx ON request(source_ip);
CREATE INDEX request_service_idx ON request(service_id);
CREATE INDEX request_action_type_idx ON request(request_type_id);
CREATE INDEX request_monitoring_idx ON request(is_monitoring);
CREATE INDEX request_user_name_idx ON request(user_name);
CREATE INDEX request_user_id_idx ON request(user_id);

CREATE INDEX request_data_entry_time_begin_idx ON request_data(request_time_begin);
CREATE INDEX request_data_entry_id_idx ON request_data(request_id);
CREATE INDEX request_data_is_response_idx ON request_data(is_response);

CREATE INDEX request_property_name_idx ON request_property_name(name);

CREATE INDEX request_property_value_entry_time_begin_idx ON request_property_value(request_time_begin);
CREATE INDEX request_property_value_entry_id_idx ON request_property_value(request_id);
CREATE INDEX request_property_value_name_id_idx ON request_property_value(property_name_id);
CREATE INDEX request_property_value_value_idx ON request_property_value(value);
CREATE INDEX request_property_value_output_idx ON request_property_value(output);
CREATE INDEX request_property_value_parent_id_idx ON request_property_value(parent_id);

CREATE INDEX request_object_ref_id_idx ON request_object_ref(request_id);
CREATE INDEX request_object_ref_time_begin_idx ON request_object_ref(request_time_begin);
CREATE INDEX request_object_ref_service_id_idx ON request_object_ref(request_service_id);
CREATE INDEX request_object_ref_object_type_id_idx ON request_object_ref(object_type_id);
CREATE INDEX request_object_ref_object_id_idx ON request_object_ref(object_id);

CREATE INDEX session_user_name_idx ON session(user_name);
CREATE INDEX session_login_date_idx ON session(login_date);
CREATE INDEX session_user_id_idx ON session(user_id);



COMMENT ON TABLE request_type IS
'List of requests which can be used by clients

id  - status
100 - ClientLogin
101 - ClientLogout
105 - ClientGreeting
120 - PollAcknowledgement
121 - PollResponse
200 - ContactCheck
201 - ContactInfo
202 - ContactDelete
203 - ContactUpdate
204 - ContactCreate
205 - ContactTransfer
400 - NSsetCheck
401 - NSsetInfo
402 - NSsetDelete
403 - NSsetUpdate
404 - NSsetCreate
405 - NSsetTransfer
500 - DomainCheck
501 - DomainInfo
502 - DomainDelete
503 - DomainUpdate
504 - DomainCreate
505 - DomainTransfer
506 - DomainRenew
507 - DomainTrade
600 - KeysetCheck
601 - KeysetInfo
602 - KeysetDelete
603 - KeysetUpdate
604 - KeysetCreate
605 - KeysetTransfer
1000 - UnknownAction
1002 - ListContact
1004 - ListNSset
1005 - ListDomain
1006 - ListKeySet
1010 - ClientCredit
1012 - nssetTest
1101 - ContactSendAuthInfo
1102 - NSSetSendAuthInfo
1103 - DomainSendAuthInfo
1104 - Info
1106 - KeySetSendAuthInfo
1200 - InfoListContacts
1201 - InfoListDomains
1202 - InfoListNssets
1203 - InfoListKeysets
1204 - InfoDomainsByNsset
1205 - InfoDomainsByKeyset
1206 - InfoDomainsByContact
1207 - InfoNssetsByContact
1208 - InfoNssetsByNs
1209 - InfoKeysetsByContact
1210 - InfoGetResults

1300 - Login
1301 - Logout
1302 - DomainFilter
1303 - ContactFilter
1304 - NSSetFilter
1305 - KeySetFilter
1306 - RegistrarFilter
1307 - InvoiceFilter
1308 - EmailsFilter
1309 - FileFilter
1310 - ActionsFilter
1311 - PublicRequestFilter

1312 - DomainDetail
1313 - ContactDetail
1314 - NSSetDetail
1315 - KeySetDetail
1316 - RegistrarDetail
1317 - InvoiceDetail
1318 - EmailsDetail
1319 - FileDetail
1320 - ActionsDetail
1321 - PublicRequestDetail

1322 - RegistrarCreate
1323 - RegistrarUpdate

1324 - PublicRequestAccept
1325 - PublicRequestInvalidate

1326 - DomainDig
1327 - FilterCreate

1328 - RequestDetail
1329 - RequestFilter

1330 - BankStatementDetail
1331 - BankStatementFilter

1400 -  Login
1401 -  Logout

1402 -  DisplaySummary
1403 -  InvoiceList
1404 -  DomainList
1405 -  FileDetail';

