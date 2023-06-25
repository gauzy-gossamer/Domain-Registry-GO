-- error messages classifier
-- DROP TABLE enum_error  CASCADE;
CREATE TABLE enum_error (
        id SERIAL CONSTRAINT enum_error_pkey PRIMARY KEY,
        status varchar(128) CONSTRAINT enum_error_status_key UNIQUE NOT NULL,
        status_cs varchar(128) CONSTRAINT enum_error_status_cs_key UNIQUE NOT NULL, -- czech translation
        status_ru varchar(128) CONSTRAINT enum_error_status_ru_key UNIQUE NOT NULL -- russian translation
        );
                        
                        

-- error messages EN and CS


INSERT INTO enum_error VALUES(  1000 , 'Command completed successfully',    'Příkaz úspěšně proveden', 'Команда выполнена успешно');
INSERT INTO enum_error VALUES(  1001 , 'Command completed successfully; action pending',  'Příkaz úspěšně proveden; vykonání akce odloženo', 'Команда выполнена успешно; ожидайте выполнения команды');

INSERT INTO enum_error VALUES(  1300 , 'Command completed successfully; no messages',    'Příkaz úspěšně proveden; žádné nové zprávy', 'Команда выполнена успешно; нет сообщений');
INSERT INTO enum_error VALUES(  1301 , 'Command completed successfully; ack to dequeue',    'Příkaz úspěšně proveden; potvrď za účelem vyřazení z fronty', 'Команда выполнена успешно; получено новое сообщение');
INSERT INTO enum_error VALUES(  1500 , 'Command completed successfully; ending session',    'Příkaz úspěšně proveden; konec relace', 'Команда выполнена успешно; сеанс завершен');


INSERT INTO enum_error VALUES(  2000 ,    'Unknown command',    'Neznámý příkaz', 'Не поддерживаемая команда');
INSERT INTO enum_error VALUES(  2001 ,    'Command syntax error',    'Chybná syntaxe příkazu', 'Синтаксическая ошибка команды');
INSERT INTO enum_error VALUES(  2002 ,    'Command use error',     'Chybné použití příkazu', 'Команда не может быть выполнена');
INSERT INTO enum_error VALUES(  2003 ,    'Required parameter missing',    'Požadovaný parametr neuveden', 'Требуемый параметр пропущен');
INSERT INTO enum_error VALUES(  2004 ,    'Parameter value range error',    'Chybný rozsah parametru', 'Требуемый параметр ошибочен');
INSERT INTO enum_error VALUES(  2005 ,    'Parameter value syntax error',    'Chybná syntaxe hodnoty parametru', 'Синтаксическая ошибка в значении параметра');


INSERT INTO enum_error VALUES( 2100 ,   'Unimplemented protocol version',    'Neimplementovaná verze protokolu', 'Неподдерживаемая версия протокола');
INSERT INTO enum_error VALUES( 2101 ,   'Unimplemented command',     'Neimplementovaný příkaz', 'Неподдерживаемая команда');
INSERT INTO enum_error VALUES( 2102 ,   'Unimplemented option',    'Neimplementovaná volba', 'Неподдерживаемая опция');
INSERT INTO enum_error VALUES( 2103 ,   'Unimplemented extension',    'Neimplementované rozšíření', 'Неподдерживаемое расширение');
INSERT INTO enum_error VALUES( 2104 ,   'Billing failure',     'Účetní selhání', 'Операция не возможна из-за отсутствия финансовых средств');
INSERT INTO enum_error VALUES( 2105 ,   'Object is not eligible for renewal',     'Objekt je nezpůsobilý pro obnovení', 'Состояние объекта запрещает выполнение операции renew');
INSERT INTO enum_error VALUES( 2106 ,   'Object is not eligible for transfer',    'Objekt je nezpůsobilý pro transfer', 'Состояние объекта запрещает выполнение операции transfer');


INSERT INTO enum_error VALUES( 2200 ,    'Authentication error',    'Chyba ověření identity', 'Ошибка аутентификации');
INSERT INTO enum_error VALUES( 2201 ,    'Authorization error',     'Chyba oprávnění', 'Ошибка авторизации');
INSERT INTO enum_error VALUES( 2202 ,    'Invalid authorization information',    'Chybná autorizační informace', 'Ошибка в авторизационной информации');

INSERT INTO enum_error VALUES( 2300 ,    'Object pending transfer',    'Objekt čeká na transfer', 'Объект ожидает подтверждения ''передачи''');
INSERT INTO enum_error VALUES( 2301 ,    'Object not pending transfer',    'Objekt nečeká na transfer', 'Объект не ожидает подтверждения ''передачи''');
INSERT INTO enum_error VALUES( 2302 ,    'Object exists',    'Objekt existuje', 'Объект существует');
INSERT INTO enum_error VALUES( 2303 ,    'Object does not exist',    'Objekt neexistuje', 'Объект не существует');
INSERT INTO enum_error VALUES( 2304 ,    'Object status prohibits operation',    'Status objektu nedovoluje operaci', 'Текущее состояние объекта запрещает выполнение операции');
INSERT INTO enum_error VALUES( 2305 ,    'Object association prohibits operation',    'Asociace objektu nedovoluje operaci', 'Операция с объектом запрещена');
INSERT INTO enum_error VALUES( 2306 ,    'Parameter value policy error',    'Chyba zásady pro hodnotu parametru', 'Значение параметра противоречит политики реестра');
INSERT INTO enum_error VALUES( 2307 ,    'Unimplemented object service',    'Neimplementovaná služba objektu', 'Операция временно невозможна');
INSERT INTO enum_error VALUES( 2308 ,    'Data management policy violation',    'Porušení zásady pro správu dat', 'Пользователю заблокированы операции модификации данных');

INSERT INTO enum_error VALUES( 2400 ,    'Command failed',    'Příkaz selhal', 'Ошибка');
INSERT INTO enum_error VALUES( 2401 ,    'Internal server error',    'Internal Příkaz selhal', 'Внутренняя ошибка сервера');
INSERT INTO enum_error VALUES( 2500 ,    'Command failed; server closing connection',    'Příkaz selhal; server uzavírá spojení', 'Ошибка; сервер закрывает сеанс');
INSERT INTO enum_error VALUES( 2501 ,    'Authentication error; server closing connection',    'Chyba ověření identity; server uzavírá spojení', 'Ошибка аутентификации; сервер закрывает сеанс.');
INSERT INTO enum_error VALUES( 2502 ,    'Session limit exceeded; server closing connection',    'Limit na počet relací překročen; server uzavírá spojení', 'Превышен лимит одновременных сеансов; сервер закрывает сеанс');
INSERT INTO enum_error VALUES( 2504 ,    'Credit balance low; server closing connection',    'Credit na počet relací překročen; server uzavírá spojení', 'Кредит исчерпан; сервер закрывает сеанс');

select setval('enum_error_id_seq', 2504);

comment on table enum_error is
'Table of error messages
id   - message
1000 - command completed successfully
1001 - command completed successfully, action pending
1300 - command completed successfully, no messages
1301 - command completed successfully, act to dequeue
1500 - command completed successfully, ending session
2000 - unknown command
2001 - command syntax error
2002 - command use error
2003 - required parameter missing
2004 - parameter value range error
2005 - parameter value systax error
2100 - unimplemented protocol version
2101 - unimplemented command
2102 - unimplemented option
2103 - unimplemented extension
2104 - billing failure
2105 - object is not eligible for renewal
2106 - object is not eligible for transfer
2200 - authentication error
2201 - authorization error
2202 - invalid authorization information
2300 - object pending transfer
2301 - object not pending transfer
2302 - object exists
2303 - object does not exists
2304 - object status prohibits operation
2305 - object association prohibits operation
2306 - parameter value policy error
2307 - unimplemented object service
2308 - data management policy violation
2400 - command failed
2500 - command failed, server closing connection
2501 - authentication error, server closing connection
2502 - session limit exceeded, server closing connection';
comment on column enum_error.id is 'id of error';
comment on column enum_error.status is 'error message in english language';
comment on column enum_error.status_cs is 'error message in native language';
