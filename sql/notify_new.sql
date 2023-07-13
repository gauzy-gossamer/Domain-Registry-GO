
CREATE TABLE message_type
(
  id  SERIAL CONSTRAINT message_type_pkey PRIMARY KEY,
  type VARCHAR(64) -- domain_expiration, mojeid_pin2, mojeid_pin3,...
);


CREATE TABLE mail_request (
	  id SERIAL PRIMARY KEY,
	  object_id bigint NOT NULL,
	  domain_id bigint NOT NULL,
	  requested timestamp NOT NULL DEFAULT now(),
	  request_type_id integer NOT NULL,
	  sent_mail boolean NOT NULL DEFAULT 'false',
	  tries integer DEFAULT 0,
	  mail_error text DEFAULT NULL  
);

CREATE TABLE mail_request_type (
	  id SERIAL PRIMARY KEY,
	  request_type varchar(255) NOT NULL,
	  subject varchar(255) NOT NULL,
	  template text NOT NULL,
	  active boolean NOT NULL DEFAULT 'false'
);

INSERT INTO mail_request_type(request_type, template, subject, active) VALUES('domain_deletion', 'Dear Colleague,
	
  This is to notify you that some object(s) in the database
  which you either maintain or are listed as to-be-notified have
  been added or changed.
	
  Object: 
	
    domain: {{name}}
    created: {{crdate}}
    reg-till: {{free_date}}
    registrar:  {{registrar}}
	
  {{state}}
	
  {{request_datetime}}', 'Registry Notification: domain {{name}} changed', 'false');
INSERT INTO mail_request_type(request_type, template, subject, active) VALUES('transfer_new', 'Dear Colleague,
	
  This is to notify you that some object(s) in the database
  which you either maintain or are listed as to-be-notified have
  been added or changed.
	
  Object: 
	
    domain: {{name}}
    created: {{crdate}}
    reg-till: {{free_date}}
    registrar:  {{registrar}}
    acquirer-id: {{acid}}
	
  {{state}}
	
  {{request_datetime}}', 'Registry Notification: transfer for {{name}} requested', 'true');
INSERT INTO mail_request_type(request_type, template, subject, active) VALUES('transfer_state_change', 'Dear Colleague,
	
  This is to notify you that some object(s) in the database
  which you either maintain or are listed as to-be-notified have
  been added or changed.
	
  Object: 
	
    domain: {{name}}
    created: {{crdate}}
    reg-till: {{free_date}}
    registrar:  {{registrar}}
    acquirer-id: {{acid}}
	
  {{state}}
	
  {{request_datetime}}', 'Registry Notification: transfer for {{name}} changed', 'true');
INSERT INTO mail_request_type(request_type, template, subject, active) VALUES('domain_state_change', 'Dear Colleague,
	        
  This is to notify you that some object(s) in the database
  which you either maintain or are listed as to-be-notified have
  been added or changed.
				        
  Object: 
					        
    domain: {{name}}
    created: {{crdate}}
    reg-till: {{free_date}}
    registrar:  {{registrar}}
										        
  {{state}}
											        
  {{request_datetime}}', 'Registry Notification: domain {{name}} changed', 'true');

INSERT INTO mail_request_type(request_type, template, subject, active) VALUES('lowcredit', 'Dear Colleague,

  You may not have enough credit in order to continue using the registry.

  Current credit: {{credit}}

  {{request_datetime}}', 'Registry Notification: low credit', 'true');

