-- For testing

-- set up zone
INSERT INTO zone(fqdn, ex_period_min, ex_period_max, dots_max, warning_letter) VALUES('ex.com', 12, 12, 1, true);

INSERT INTO price_list(zone_id, operation_id, valid_from, price, quantity) VALUES(1, (SELECT id FROM enum_operation WHERE operation='CreateDomain'), '2010-01-01', 0, 1);
INSERT INTO price_list(zone_id, operation_id, valid_from, price, quantity) VALUES(1, (SELECT id FROM enum_operation WHERE operation='RenewDomain'), '2010-01-01', 0, 1);
INSERT INTO price_list(zone_id, operation_id, valid_from, price, quantity) VALUES(1, (SELECT id FROM enum_operation WHERE operation='TransferDomain'), '2010-01-01', 0, 1);

-- set up system registrar
INSERT INTO registrar(object_id, handle, name, intpostal, system) VALUES(1,'SYSTEM-REG', 'System Registrar', 'Company l.t.d.', 't');
INSERT INTO registraracl(registrarid, cert, password) VALUES(1, 'A1:DD:46:43:35:51:EB:5F:42:8B:DF:A1:77:19:EA:DD', 'password');
INSERT INTO registrarinvoice(registrarid, zone, fromdate) VALUES(1,1,'2010-01-01');

SELECT create_object(1, 'SYSTEM-REG', (select id from enum_object_type where name='registrar'));
INSERT INTO object(id, clid) VALUES(1, 1);

-- set up test registrar
INSERT INTO registrar(object_id, handle, name, intpostal, system) VALUES(2,'TEST-REG', 'Test Registrar', 'Company l.t.d.', 'f');
INSERT INTO registraracl(registrarid, cert, password) VALUES(2, 'A1:DD:46:43:35:51:EB:5F:42:8B:DF:A1:77:19:EA:DD', 'password');
INSERT INTO registrarinvoice(registrarid, zone, fromdate) VALUES(2,1,'2010-01-01');
INSERT INTO registrar_credit(credit, registrar_id, zone_id) VALUES(10, 2, 1);

SELECT create_object(2, 'TEST-REG', (select id from enum_object_type where name='registrar'));
INSERT INTO object(id, clid) VALUES(2, 2);
