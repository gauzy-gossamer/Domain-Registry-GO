-- system operational parameter
-- parameters are accessed through their id not name, so proper
-- numbering is essential

CREATE TABLE enum_parameters (
  id INTEGER CONSTRAINT enum_parameters_pkey PRIMARY KEY, -- primary identification 
  name VARCHAR(100) NOT NULL CONSTRAINT enum_parameters_name_key UNIQUE, -- descriptive name (informational)
  val VARCHAR(100) NOT NULL -- value of parameter
);

-- parameter 1 is for checking data model version and for applying upgrade
-- scripts
INSERT INTO enum_parameters (id, name, val) 
VALUES (1, 'model_version', '2.26.1');
-- parameter 2 is for updating table enum_tlds by data from url
-- http://data.iana.org/TLD/tlds-alpha-by-domain.txt
INSERT INTO enum_parameters (id, name, val) 
VALUES (2, 'tld_list_version', '2008013001');
-- parameter 3 is used to change state of domain to unguarded and remove
-- this domain from DNS. value is number of dates relative to date 
-- domain.exdate 
INSERT INTO enum_parameters (id, name, val) 
VALUES (3, 'expiration_notify_period', '-60');
-- parameter 4 is used to change state of domain to unguarded and remove
-- this domain from DNS. value is number of dates relative to date 
-- domain.exdate 
INSERT INTO enum_parameters (id, name, val) 
VALUES (4, 'expiration_dns_protection_period', '30');
-- parameter 5 is used to change state of domain to deleteWarning and 
-- generate letter with warning. value number of dates relative to date 
-- domain.exdate 
INSERT INTO enum_parameters (id, name, val) 
VALUES (5, 'expiration_letter_warning_period', '34');
-- parameter 6 is used to change state of domain to deleteCandidate and 
-- unregister domain from system. value is number of dates relative to date 
-- domain.exdate 
INSERT INTO enum_parameters (id, name, val) 
VALUES (6, 'expiration_registration_protection_period', '31');
-- parameter 7 is used to change state of domain to validationWarning1 and 
-- send poll message to registrar. value is number of dates relative to date 
-- domain.exdate 
INSERT INTO enum_parameters (id, name, val) 
VALUES (7, 'validation_notify1_period', '-30');
-- parameter 8 is used to change state of domain to validationWarning2 and 
-- send email to registrant. value is number of dates relative to date 
-- domain.exdate 
INSERT INTO enum_parameters (id, name, val) 
VALUES (8, 'validation_notify2_period', '-15');
-- parameter 9 is used to identify hour when objects are deleted
-- value is number of hours relative to date of operation
INSERT INTO enum_parameters (id, name, val) 
VALUES (9, 'regular_day_procedure_period', '0');
-- parameter 10 is used to identify time zone in which parameter 9 and 14
-- are specified
INSERT INTO enum_parameters (id, name, val) 
VALUES (10, 'regular_day_procedure_zone', 'Europe/Prague');
-- parameter 11 is used to change state of objects other than domain to
-- deleteCandidate. It is specified in granularity of months and means, period
-- during which object wasn't linked to other object and wasn't updated 
INSERT INTO enum_parameters (id, name, val) 
VALUES (11, 'object_registration_protection_period', '6');
-- parameter 12 is used to change protection period of deleted object handle
-- (contact, nsset, keyset). value is in months.
INSERT INTO enum_parameters (id, name, val)
VALUES (12, 'handle_registration_protection_period', '0');
-- parameter 13 is used as a suffix in object_registry roid string
-- this suffix should match pattern \w{1,8}
INSERT INTO enum_parameters (id, name, val)
VALUES (13, 'roid_suffix', 'EPP');
-- parameter 14 is used to identify hour when domains are moving outzone. 
-- value is number of hours relative to date of operation
INSERT INTO enum_parameters (id, name, val) 
VALUES (14, 'regular_day_outzone_procedure_period', '14');
-- parameter 18 is used to generate email with warning
-- value is number of days relative to domain.exdate
INSERT INTO enum_parameters (id, name, val) 
VALUES (18, 'outzone_unguarded_email_warning_period', '25');
--opportunity window in days before current ENUM domain validation expiration
-- for new ENUM domain validation to be appended after current ENUM domain validation
INSERT INTO enum_parameters (id, name, val)
VALUES (19, 'enum_validation_continuation_window', '14');

-- minimum number of hosts for delegation
INSERT INTO enum_parameters (id, name, val)
VALUES (20, 'min_delegation_hosts', '2');

-- number of days for which authinfo is valid
INSERT INTO enum_parameters(id, name, val) 
VALUES(21, 'authinfo_period', '-20');

-- don't delete unlinked contacts when domain is deleted
INSERT INTO enum_parameters(id, name, val) 
VALUES(22, 'keep_unlinked_contacts', '1');

comment on table enum_parameters is
'Table of system operational parameters.
Meanings of parameters:

1 - model version - for checking data model version and for applying upgrade scripts
2 - tld list version - for updating table enum_tlds by data from url
3 - expiration notify period - used to change state of domain to unguarded and remove domain from DNS,
    value is number of days relative to date domain.exdate
4 - expiration dns protection period - same as parameter 3
5 - expiration letter warning period - used to change state of domain to deleteWarning and generate letter
    with warning
6 - expiration registration protection period - used to change state of domain to deleteCandidate and
    unregister domain from system
7 - validation notify 1 period - used to change state of domain to validationWarning1 and send poll
    message to registrar
8 - validation notify 2 period - used to change state of domain to validationWarning2 and send
    email to registrant
9 - regular day procedure period - used to identify hout when objects are deleted and domains
    are moving outzone
10 - regular day procedure zone - used to identify time zone in which parameter 9 is specified';
comment on column enum_parameters.id is 'primary identification';
comment on column enum_parameters.name is 'descriptive name of parameter - for information uses only';
comment on column enum_parameters.val is 'value of parameter';
