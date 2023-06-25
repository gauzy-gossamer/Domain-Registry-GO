-- DROP TABLE  OBJECT  CASCADE;

CREATE TABLE enum_object_type
(
  id integer NOT NULL,
  name TEXT NOT NULL,
  CONSTRAINT enum_object_type_pkey PRIMARY KEY (id),
  CONSTRAINT enum_object_type_name_key UNIQUE (name)
);


INSERT INTO enum_object_type (id,name) VALUES ( 1 , 'contact' );
INSERT INTO enum_object_type (id,name) VALUES ( 2 , 'nsset' );
INSERT INTO enum_object_type (id,name) VALUES ( 3 , 'domain' );
INSERT INTO enum_object_type (id,name) VALUES ( 4 , 'keyset' );
INSERT INTO enum_object_type (id,name) VALUES ( 5 , 'registrar' );

--Helps to write query using conditional index like type=1
-- IMPORTANT!!! without "SET search_path=public" pg_dump/restore not working correctly with postgresql > 9.3!!!!
CREATE FUNCTION get_object_type_id(TEXT)
RETURNS integer AS
'SELECT id FROM enum_object_type WHERE name=$1'
LANGUAGE SQL IMMUTABLE RETURNS NULL ON NULL INPUT SET search_path=public;


CREATE TABLE OBJECT_registry (
       ID SERIAL CONSTRAINT object_registry_pkey PRIMARY KEY,
       ROID varchar(255) CONSTRAINT object_registry_roid_key UNIQUE NOT NULL , -- unique roid
       type smallint , -- object type 1 contact 2 nsset 3 domain
       NAME varchar(255)  NOT NULL , -- handle or FQDN
       CrID INTEGER NOT NULL CONSTRAINT object_registry_crid_fkey REFERENCES Registrar,
       CrDate timestamp NOT NULL DEFAULT now(),
       ErDate timestamp DEFAULT NULL, -- erase date 
       CrhistoryID INTEGER CONSTRAINT object_registry_crhistoryid_fkey REFERENCES History, -- link into create history
       historyID integer CONSTRAINT object_registry_historyid_fkey REFERENCES history, -- link to last change in history
       CONSTRAINT name_case_check CHECK(CASE WHEN (type = get_object_type_id('domain') OR type = get_object_type_id('contact')) 
	                                THEN name = LOWER(name) ELSE name = UPPER(name) END), 
       CONSTRAINT object_registry_type_fkey FOREIGN KEY (type)
           REFERENCES enum_object_type (id)
       );


-- index
CREATE INDEX object_registry_upper_name_1_idx 
 ON object_registry (UPPER(name)) WHERE type=1;
CREATE INDEX object_registry_upper_name_2_idx 
 ON object_registry (UPPER(name)) WHERE type=2;
CREATE INDEX object_registry_name_3_idx 
 ON object_registry  (NAME) WHERE type=3;
CREATE INDEX object_registry_historyid_idx ON object_registry (historyid);

comment on column OBJECT_registry.ID is 'unique automatically generated identifier';
comment on column OBJECT_registry.ROID is 'unique roid';
comment on column OBJECT_registry.type is 'object type (1-contact, 2-nsset, 3-domain)';
comment on column OBJECT_registry.name is 'handle of fqdn';
comment on column OBJECT_registry.CrID is 'link to registrar';
comment on column OBJECT_registry.CrDate is 'object creation date and time';
comment on column OBJECT_registry.ErDate is 'object erase date';
comment on column OBJECT_registry.CrhistoryID is 'link into create history';
comment on column OBJECT_registry.historyID is 'link to last change in history';

-- For updating previous history record (or current in case of deletion of object)
CREATE OR REPLACE FUNCTION object_registry_update_history_rec() RETURNS TRIGGER AS $$
BEGIN
    -- when updation object, set valid_to and next of previous history record
    IF OLD.historyid != NEW.historyid THEN
        UPDATE history
            SET valid_to = NOW(), -- NOW() is the same during the transaction, so this will be the same as valid_from of new history record
                next = NEW.historyid
            WHERE id = OLD.historyid;
    END IF; 
    
    -- when deleting object (setting object_registry.erdate), set valid_to of current history record    
    IF OLD.erdate IS NULL and NEW.erdate IS NOT NULL THEN
        UPDATE history
            SET valid_to = NEW.erdate
            WHERE id = OLD.historyid;
    END IF; 
    
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

--DROP TRIGGER trigger_object_registry_update_history_rec ON object_registry;
CREATE TRIGGER trigger_object_registry_update_history_rec AFTER UPDATE
    ON object_registry FOR EACH ROW
    EXECUTE PROCEDURE object_registry_update_history_rec();


CREATE TABLE OBJECT (
        ID INTEGER CONSTRAINT object_pkey PRIMARY KEY CONSTRAINT object_id_fkey REFERENCES object_registry (id),
        ClID INTEGER NOT NULL CONSTRAINT object_clid_fkey REFERENCES Registrar,
        UpID INTEGER CONSTRAINT object_upid_fkey REFERENCES Registrar,
        TrDate timestamp DEFAULT NULL,
        UpDate timestamp DEFAULT NULL,
        AuthInfoPw varchar(300), -- in XML schemas
	AuthInfoUpdate timestamp DEFAULT NULL
        );

-- index
CREATE INDEX object_upid_idx ON "object" (upid);
CREATE INDEX object_clid_idx ON "object" (clid);


-- DROP TABLE Contact CASCADE;
CREATE TABLE Contact (
        ID INTEGER CONSTRAINT contact_pkey PRIMARY KEY CONSTRAINT contact_id_fkey REFERENCES object (id),
        Name varchar(1024),
        Organization varchar(1024),
        City varchar(1024),
        PostalCode varchar(32),
        Country char(2) CONSTRAINT contact_country_fkey REFERENCES enum_country,
        Telephone jsonb,
        Fax jsonb,
        Email jsonb,
        intpostal varchar(512),
        intaddress jsonb,
        locpostal varchar(512),
        locaddress jsonb,
        legaladdress jsonb,
        passport jsonb,
        birthday date,
        verified boolean NOT NULL DEFAULT False,
        additional text,
        contact_type integer NOT NULL,
        DiscloseName boolean NOT NULL DEFAULT True,
        DiscloseOrganization boolean NOT NULL DEFAULT True,
        DiscloseAddress boolean NOT NULL DEFAULT True,
        DiscloseTelephone boolean NOT NULL DEFAULT False,
        DiscloseFax boolean NOT NULL DEFAULT False,
        DiscloseEmail boolean NOT NULL DEFAULT False,
        DisclosetaxpayerNumbers boolean NOT NULL DEFAULT False,
        Disclosepassport boolean NOT NULL DEFAULT False,
        Discloseintpostal boolean NOT NULL DEFAULT False,
        Discloseintaddress boolean NOT NULL DEFAULT False,
        Discloselocpostal boolean NOT NULL DEFAULT False,
        Discloselocaddress boolean NOT NULL DEFAULT False,
        Discloselegaladdress boolean NOT NULL DEFAULT False,
        Disclosebirthday boolean NOT NULL DEFAULT False,

        NotifyEmail varchar(1024),
        VAT varchar(32),
        SSN varchar(64),
	    SSNtype INTEGER CONSTRAINT contact_ssntype_fkey REFERENCES enum_ssntype,
        DiscloseVAT boolean NOT NULL DEFAULT False,
        DiscloseIdent boolean NOT NULL DEFAULT False,
        DiscloseNotifyEmail boolean NOT NULL DEFAULT False,
        warning_letter boolean DEFAULT NULL
        );

comment on table Contact is 'List of contacts which act in registry as domain owners and administrative contacts for nameservers group';
comment on column Contact.ID is 'references into object table';
comment on column Contact.Name is 'name of contact person';
comment on column Contact.Organization is 'full trade name of organization';
comment on column Contact.PostalCode is 'part of address - postal code';
comment on column Contact.Country is 'two character country code (e.g. cz) from enum_country table';
comment on column Contact.Telephone is 'telephone number';
comment on column Contact.Fax is 'fax number';
comment on column Contact.Email is 'email address';
comment on column Contact.DiscloseName is 'whether reveal contact name';
comment on column contact.DiscloseOrganization is 'whether reveal organization';
comment on column Contact.DiscloseAddress is 'whether reveal address';
comment on column Contact.DiscloseTelephone is 'whether reveal phone number';
comment on column Contact.DiscloseFax is 'whether reveal fax number';
comment on column Contact.DiscloseEmail is 'whether reveal email address';
comment on column Contact.NotifyEmail is 'to this email address will be send message in case of any change in domain or nsset affecting contact';
comment on column Contact.VAT is 'tax number';
comment on column Contact.SSN is 'unambiguous identification number (e.g. Social Security number, identity card number, date of birth)';
comment on column Contact.SSNtype is 'type of identification number from enum_ssntype table';
comment on column Contact.DiscloseVAT is 'whether reveal VAT number';
comment on column Contact.DiscloseIdent is 'whether reveal SSN number';
comment on column Contact.DiscloseNotifyEmail is 'whether reveal notify email';
COMMENT ON COLUMN contact.warning_letter IS 'user preference whether to send domain expiration letters (NULL - no user preference; TRUE - send domain expiration letters; FALSE - don''t send domain expiration letters';


---
---  Ticket #11106 merge contact fn index
---

CREATE INDEX contact_name_coalesce_trim_idx ON contact (trim(both ' ' from COALESCE(name,'')));

-- DROP TABLE NSSet CASCADE;
CREATE TABLE NSSet (
        ID INTEGER CONSTRAINT nsset_pkey PRIMARY KEY CONSTRAINT nsset_id_fkey REFERENCES object (id),
        checklevel smallint default 0,
        client_host boolean default false
        );

CREATE TABLE public.domain_host_map (
    domainid integer NOT NULL, 
    hostid integer NOT NULL
);

CREATE INDEX domain_host_map_idx ON public.domain_host_map USING btree (domainid);
CREATE INDEX domain_host_map_idx2 ON public.domain_host_map USING btree (hostid);


CREATE TABLE public.host (
    hostid integer NOT NULL,
    fqdn character varying(255) NOT NULL,
    delegatable boolean default true
);

CREATE UNIQUE INDEX host_idx ON public.host USING btree (hostid);

-- DROP TABLE Host CASCADE;
--CREATE TABLE Host (
--        ID SERIAL CONSTRAINT host_pkey PRIMARY KEY,
--        NSSetID INTEGER CONSTRAINT host_nssetid_fkey REFERENCES NSSet ON UPDATE CASCADE,
---        FQDN VARCHAR(255)   NOT NULL,  -- it cannot be UNIQUE for two different NSSET same dns host 
--        CONSTRAINT host_nssetid_fqdn_key UNIQUE (NSSetID, FQDN ) -- unique key
--        );


--CREATE INDEX host_nsset_idx ON Host (NSSetID);
--CREATE INDEX host_fqdn_idx ON Host (FQDN);

--comment on table Host is 'Records of relationship between nameserver and ip address';
--comment on column Host.id is 'unique automatically generatet identifier';
--comment on column Host.NSSetID is 'in which nameserver group belong this record';
--comment on column Host.FQDN is 'fully qualified domain name that is in zone file as NS';

-- DROP TABLE  host_ipaddr_map  CASCADE;


CREATE TABLE host_ipaddr_map (
           ID SERIAL CONSTRAINT host_ipaddr_map_pkey PRIMARY KEY,
           HostID  INTEGER NOT NULL CONSTRAINT host_ipaddr_map_hostid_fkey REFERENCES HOST(hostid) ON UPDATE CASCADE ON DELETE CASCADE,
           IpAddr INET NOT NULL -- IP address
         );

CREATE INDEX host_ipaddr_map_hostid_idx ON host_ipaddr_map (hostid);


-- DROP TABLE Domain CASCADE;
CREATE TABLE Domain (
        ID INTEGER CONSTRAINT domain_pkey PRIMARY KEY CONSTRAINT domain_id_fkey REFERENCES object (ID),
        Zone INTEGER NOT NULL CONSTRAINT domain_zone_fkey REFERENCES Zone (ID),
        Registrant INTEGER NOT NULL CONSTRAINT domain_registrant_fkey REFERENCES Contact,
        NSSet INTEGER NULL CONSTRAINT domain_nsset_fkey REFERENCES NSSet, -- link to nsset can be also NULL, it can register domain without nsset
        Exdate timestamp NOT NULL,
        Description jsonb
        );
CREATE INDEX domain_zone_idx ON Domain (Zone);
CREATE INDEX domain_registrant_idx ON Domain (registrant);
CREATE INDEX domain_nsset_idx ON Domain (nsset);
CREATE INDEX domain_exdate_idx ON domain (exdate);

comment on table Domain is 'Evidence of domains';
comment on column Domain.ID is 'point to object table';
comment on column Domain.Zone is 'zone in which domain belong';
comment on column Domain.Registrant is 'domain owner';
comment on column Domain.NSSet is 'link to nameserver set, can be NULL (when is domain registered withou nsset)';
comment on column Domain.Exdate is 'domain expiry date';

-- delete domains only on dates present in this table
CREATE TABLE domain_allow_removal_dates (
    allow_date date
    );

CREATE UNIQUE INDEX allow_dates_idx ON domain_allow_removal_dates(allow_date);

-- DROP TABLE domain_contact_map CASCADE;
CREATE TABLE domain_contact_map (
        DomainID INTEGER CONSTRAINT domain_contact_map_domainid_fkey REFERENCES Domain ON UPDATE CASCADE ON DELETE CASCADE NOT NULL,
        ContactID INTEGER CONSTRAINT domain_contact_map_contactid_fkey REFERENCES Contact ON UPDATE CASCADE NOT NULL,
        Role INTEGER NOT NULL DEFAULT 1,
        CONSTRAINT domain_contact_map_pkey PRIMARY KEY (DomainID, ContactID)
        );
CREATE INDEX domain_contact_map_domainid_idx ON domain_contact_map (DomainID);
CREATE INDEX domain_contact_map_contactid_idx ON domain_contact_map (ContactID);


-- DROP TABLE ENUMVal CASCADE;
CREATE TABLE ENUMVal (
        DomainID INTEGER NOT NULL
        CONSTRAINT enumval_pkey PRIMARY KEY
        CONSTRAINT enumval_domainid_fkey REFERENCES Domain ON UPDATE CASCADE ON DELETE CASCADE,
        ExDate date NOT NULL,
        publish BOOLEAN NOT NULL DEFAULT false
        );
        
---
--- Ticket #7873
---

-- enumval domainid unique constraint
ALTER TABLE enumval ADD CONSTRAINT enumval_domainid_key UNIQUE (domainid);

---
--- Ticket #16022 - nsset dnshost prohibited IP address config
---

--DROP TABLE nsset_dnshost_prohibited_ipaddr;
CREATE TABLE nsset_dnshost_prohibited_ipaddr (
        network INET NOT NULL,
        netmask INET NOT NULL,
        UNIQUE (network, netmask),
        CHECK (family(network) = family(netmask)),
        CHECK (host(network)::inet = network),
        CHECK (host(netmask)::inet = netmask)
        );

COMMENT ON TABLE nsset_dnshost_prohibited_ipaddr IS 'nsset dnshost prohibited IP address config, IP address $1 is prohibited if  SELECT bool_or(($1::inet & netmask) = network)  FROM nsset_dnshost_prohibited_ipaddr WHERE family($1::inet) = family(network); returns true.';


ALTER TABLE ONLY public.host_ipaddr_map
    ADD CONSTRAINT host_hostid_fkey FOREIGN KEY (hostid) REFERENCES public.host(hostid) ON UPDATE CASCADE ON DELETE CASCADE;

ALTER TABLE ONLY public.domain_host_map
    ADD CONSTRAINT domain_hostid_fkey FOREIGN KEY (hostid) REFERENCES public.host(hostid) ON UPDATE CASCADE ON DELETE CASCADE;

ALTER TABLE ONLY public.domain_host_map
    ADD CONSTRAINT domain_host_domainid_fkey FOREIGN KEY (domainid) REFERENCES public.domain(id) ON UPDATE CASCADE ON DELETE CASCADE;

ALTER TABLE ONLY public.host
    ADD CONSTRAINT host_hostid_fkey FOREIGN KEY (hostid) REFERENCES public.object_registry(id) ON UPDATE CASCADE ON DELETE CASCADE;

