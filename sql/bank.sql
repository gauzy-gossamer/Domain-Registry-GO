
-- drop tables
-- drop table bank_account;

-- bank classifier 
-- CREATE TABLE enum_bank_code (
-- code char(4) CONSTRAINT enum_bank_code_pkey PRIMARY KEY,
-- name_short CONSTRAINT enum_bank_code_name_short_key varchar(4) UNIQUE NOT NULL , -- shortcut
-- name_full CONSTRAINT enum_bank_code_name_full_key varchar(64) UNIQUE  NOT NULL -- full name
-- );

-- ACCOUNT -- table of our accounts
CREATE TABLE bank_account 
(
id serial NOT NULL CONSTRAINT bank_account_pkey PRIMARY KEY, -- unique primary key
Zone INTEGER CONSTRAINT bank_account_zone_fkey REFERENCES Zone (ID), -- for which zone should be account executed
account_number char(16) NOT NULL , -- account number
account_name  char(20) , -- account name
bank_code char(4) CONSTRAINT bank_account_bank_code_fkey REFERENCES enum_bank_code,   -- bank code
balance  numeric(10,2) default 0.0, -- actual balance 
last_date date, -- date of last statement 
last_num int  -- number of last statement
);

-- coupling variable symbol of registrar is in a table registrar ( it is his ICO for CZ ) a it is valid for all zones

comment on table bank_account is
'This table contains information about registry administrator bank account';
comment on column bank_account.id is 'unique automatically generated identifier';
comment on column bank_account.zone is 'for which zone should be account executed';
comment on column bank_account.balance is 'actual balance';
comment on column bank_account.last_date is 'date of last statement';
comment on column bank_account.last_num is 'number of last statement';

