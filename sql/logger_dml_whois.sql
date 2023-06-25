INSERT INTO service (id, partition_postfix, name) VALUES 
(0, 'whois_', 'Unix whois');

INSERT INTO request_type (service_id, id, name) VALUES 
(0, 1105, 'Info');

INSERT INTO result_code (service_id, result_code, name) VALUES 
(0, 101 , 'NoEntriesFound'),
(0, 107 , 'UsageError'),
(0, 108 , 'InvalidRequest'),
(0, 501 , 'InternalServerError'),
(0, 0   , 'Ok');

