CREATE TABLE bank_statement 
(
    id serial NOT NULL CONSTRAINT bank_statement_pkey PRIMARY KEY, -- unique primary key
    account_id int CONSTRAINT bank_statement_account_id_fkey REFERENCES bank_account, -- processing for given account link to account tabel
    num int, -- serial number statement
    create_date date , --  create date of a statement
    balance_old_date date , -- date of a last balance
    balance_old numeric(10,2) , -- old balance
    balance_new numeric(10,2) ,  -- new balance
    balance_credit  numeric(10,2) , -- income during statement ( credit balance )
    balance_debet numeric(10,2) -- expenses during statement ( debet balance )
);

comment on column bank_statement.id is 'unique automatically generated identifier';
comment on column bank_statement.account_id is 'link to used bank account';
comment on column bank_statement.num is 'statements number';
comment on column bank_statement.create_date is 'statement creation date';
comment on column bank_statement.balance_old is 'old balance state';
comment on column bank_statement.balance_credit is 'income during statement';
comment on column bank_statement.balance_debet is 'expenses during statement';


CREATE TABLE bank_payment
(
    id serial NOT NULL CONSTRAINT bank_payment_pkey PRIMARY KEY, -- unique primary key
    statement_id int CONSTRAINT bank_payment_statement_id_fkey REFERENCES bank_statement default null, -- link into table heads of bank statements
    account_id int CONSTRAINT bank_payment_account_id_fkey REFERENCES bank_account default null, -- link into table of accounts
    account_number text NOT NULL , -- contra-account number from which came or was sent a payment
    bank_code varchar(35) NOT NULL,   -- bank code
    code int, -- account code 1 debet item 2 credit item 4  cancel debet 5 cancel credit 
    type int NOT NULL default 1, -- transfer type
    status int, -- payment status
    KonstSym varchar(10), -- constant symbol ( it contains bank code too )
    VarSymb varchar(10), -- variable symbol
    SpecSymb varchar(10), -- constant symbol
    price numeric(10,2) NOT NULL,  -- applied amount if a debet is negative amount 
    account_evid varchar(20), -- account evidence 
    account_date date NOT NULL, --  accounting date of credit or sending 
    account_memo  varchar(64), -- note
    account_name  varchar(64), -- account name
    crtime timestamp NOT NULL default now(),
    CONSTRAINT bank_payment_account_id_account_evid_key UNIQUE(account_id, account_evid)
);

comment on column bank_payment.id is 'unique automatically generated identifier';
comment on column bank_payment.statement_id is 'link to statement head';
comment on column bank_payment.account_id is 'link to account table';
comment on column bank_payment.account_number is 'contra-account number from which came or was sent a payment';
comment on column bank_payment.bank_code is 'contra-account bank code';
comment on column bank_payment.code is 'operation code (1-debet item, 2-credit item, 4-cancel debet, 5-cancel credit)';
comment on column bank_payment.type is 'transfer type (1-not decided (not processed), 2-from/to registrar, 3-from/to bank, 4-between our own accounts, 5-related to academia, 6-other transfers';
comment on column bank_payment.status is 'payment status (1-Realized (only this should be further processed), 2-Partially realized, 3-Not realized, 4-Suspended, 5-Ended, 6-Waiting for clearing )';
comment on column bank_payment.KonstSym is 'constant symbol (contains bank code too)';
comment on column bank_payment.VarSymb is 'variable symbol';
comment on column bank_payment.SpecSymb is 'spec symbol';
comment on column bank_payment.price is 'applied positive(credit) or negative(debet) amount';
comment on column bank_payment.account_evid is 'account evidence';
comment on column bank_payment.account_date is 'accounting date';
comment on column bank_payment.account_memo is 'note';
comment on column bank_payment.account_name is 'account name';
comment on column bank_payment.crtime is 'create timestamp';

CREATE TABLE bank_payment_registrar_credit_transaction_map
(
    id BIGSERIAL CONSTRAINT bank_payment_registrar_credit_transaction_map_pkey PRIMARY KEY
    , bank_payment_id bigint NOT NULL REFERENCES bank_payment(id)
    , registrar_credit_transaction_id bigint
    CONSTRAINT bank_payment_registrar_credit_registrar_credit_transaction__key UNIQUE
    NOT NULL
    CONSTRAINT bank_payment_registrar_credit_registrar_credit_transaction_fkey REFERENCES registrar_credit_transaction(id)
);

COMMENT ON TABLE bank_payment_registrar_credit_transaction_map
	IS 'payment assigned to credit items';





