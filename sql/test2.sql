-- set up second registrar for regcore tests
INSERT INTO registrar(object_id, handle, name, intpostal, system) VALUES(3,'TEST2-REG', 'Test Registrar', 'Company l.t.d.', 'f');
INSERT INTO registraracl(registrarid, cert, password) VALUES(3, 'A1:DD:46:43:35:51:EB:5F:42:8B:DF:A1:77:19:EA:DD', 'password');
INSERT INTO registrarinvoice(registrarid, zone, fromdate) VALUES(3,1,'2010-01-01');
INSERT INTO registrar_credit(credit, registrar_id, zone_id) VALUES(10, 3, 1);

SELECT create_object(3, 'TEST2-REG', (select id from enum_object_type where name='registrar'));
INSERT INTO object(id, clid) VALUES(3, 3); 

-- for testing low credit messages
INSERT INTO poll_credit_zone_limit VALUES(1, 20);
