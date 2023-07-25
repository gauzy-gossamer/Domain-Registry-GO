CREATE TABLE domain_blacklist
(
  id serial NOT NULL, -- primary key
  regexp varchar(255) NOT NULL, -- regular expression which is blocked
  valid_from timestamp NOT NULL, -- from when bloc is valid
  valid_to timestamp, -- till when bloc is valid, if it is NULL, it isn't restricted
  reason varchar(255) NOT NULL, -- reason why is domain blocked
  CONSTRAINT domain_blacklist_pkey PRIMARY KEY (id)
);

comment on column domain_blacklist.regexp is 'regular expression which is blocked';
comment on column domain_blacklist.valid_from is 'from when is block valid';
comment on column domain_blacklist.valid_to is 'till when is block valid, if it is NULL, it is not restricted';
comment on column domain_blacklist.reason is 'reason why is domain blocked';

create table admin(
  id serial,
  username varchar(50) not null,
  password varchar(200) not null,
  last_login timestamptz not null,
  email varchar(200) not null,
  created_at timestamptz not null default current_timestamp,
  intro varchar(255),
  constraint admin_pk PRIMARY KEY(id)
);

CREATE TABLE adminsession(
  id serial,
  admin_id integer not null,
  token varchar(255) not null,
  expire timestamptz not null,
  constraint adminsession_pk PRIMARY KEY(id),
  CONSTRAINT admin_fkey FOREIGN KEY (admin_id)
           REFERENCES admin (id)
);

CREATE INDEX admin_name_idx ON admin (username);
CREATE INDEX adminsession_token_idx ON adminsession(token);
