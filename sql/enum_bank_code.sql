-- classifier of error messages reason
-- DROP TABLE enum_bank_code  CASCADE;
CREATE TABLE enum_bank_code (
      code char(4) CONSTRAINT enum_bank_code_pkey PRIMARY KEY,
      name_short varchar(4) CONSTRAINT enum_bank_code_name_short_key UNIQUE NOT NULL , -- short cut 
      name_full varchar(64) CONSTRAINT enum_bank_code_name_full_key UNIQUE  NOT NULL -- full name
);

comment on table enum_bank_code is 'list of bank codes';
comment on column enum_bank_code.code is 'bank code';
comment on column enum_bank_code.name_short is 'bank name abbrevation';
comment on column enum_bank_code.name_full is 'full bank name';
