-- classifier of error messages  reason
CREATE TABLE enum_reason (
        id SERIAL CONSTRAINT enum_reason_pkey PRIMARY KEY,
        reason varchar(128) CONSTRAINT enum_reason_reason_key UNIQUE NOT NULL,
        reason_cs varchar(128) CONSTRAINT enum_reason_reason_cs_key UNIQUE  NOT NULL, -- czech translation
        reason_ru varchar(128) CONSTRAINT enum_reason_reason_ru_key UNIQUE  NOT NULL -- russian translation
        );


INSERT INTO enum_reason VALUES(  1 ,  'bad format of contact handle'   , 'neplatný formát ukazatele kontaktu', '1' );
INSERT INTO enum_reason VALUES(  2 ,  'bad format of host handle' ,  'neplatný formát ukazatele nssetu', '2' );
INSERT INTO enum_reason VALUES(  3 ,  'bad format of fqdn domain'  , 'neplatný formát názvu domény', '3' );

INSERT INTO enum_reason VALUES(  4 ,  'Domain name not applicable.'  , 'nepoužitelný název domény', '4' );

-- for check
INSERT INTO enum_reason VALUES(  5 , 'invalid format'  , 'neplatný formát', '5' );
INSERT INTO enum_reason VALUES(  6 ,  'already registered.'   , 'již zaregistrováno', '6' );
INSERT INTO enum_reason VALUES(  7 , 'within protection period.' , 'je v ochranné lhůtě', '7' );


INSERT INTO enum_reason VALUES( 8 , 'Invalid IP address.' , 'neplatná IP adresa', '8'  );
INSERT INTO enum_reason VALUES(  9 ,  'Invalid nameserver hostname.'  ,  'neplatný formát názvu jmenného serveru DNS', '9' );
INSERT INTO enum_reason VALUES(  10 ,  'Duplicate nameserver address.' , 'duplicitní adresa jmenného serveru DNS', '10'  );
INSERT INTO enum_reason VALUES(  11 ,  'Glue IP address not allowed here.'  , 'nepovolená  IP adresa glue záznamu', '11' );
INSERT INTO enum_reason VALUES(  12 ,  'At least two nameservers required.'  , 'jsou zapotřebí alespoň dva DNS servery', '12' );


-- badly entered period in domain renew 
INSERT INTO enum_reason VALUES(  13 , 'invalid date of period' ,  'neplatná hodnota periody', '13' );
INSERT INTO enum_reason VALUES( 14 , 'period exceedes maximal allowed validity time.' , 'perioda je nad maximální dovolenou hodnotou', '14' );
INSERT INTO enum_reason VALUES( 15 , 'period is not aligned with allowed step.' , 'perioda neodpovídá dovolenému intervalu', '15' );


-- country code doesn't exist
INSERT INTO enum_reason VALUES(  16 , 'Unknown country code'  , 'neznámý kód země', '16' );

-- unknown msgID doesn't exist
INSERT INTO enum_reason VALUES(  17 , 'Unknown message ID' ,  'neznámé msgID', '17' );


-- for ENUMval	
INSERT INTO enum_reason VALUES(  18 , 'Validation expiration date can not be used here.' ,  'datum vypršení platnosti se nepoužívá', '18'  );
INSERT INTO enum_reason VALUES(  19 , 'Validation expiration date does not match registry data.' , 'datum vypršení platnosti je neplatné', '19' );
INSERT INTO enum_reason VALUES(  20 , 'Validation expiration date is required.' , 'datum vypršení platnosti je požadováno', '20' );

-- Update NSSET it can not be removed or added  DNS host or replaced tech contact 1 tech is minimum 
INSERT INTO enum_reason VALUES(  21 , 'Can not remove nameserver.'  , 'nelze odstranit jmenný server DNS', '21' );
INSERT INTO enum_reason VALUES(  22 , 'Can not add nameserver'  , 'nelze přidat jmenný server DNS', '22' );

INSERT INTO enum_reason VALUES(  23 ,  'Can not remove technical contact'  , 'nelze vymazat technický kontakt', '23'  );

-- when technical/administrative contact not exist or is already assigned to object (domain/keyset/nsset)
INSERT INTO enum_reason VALUES(  24 , 'Technical contact is already assigned to this object.'  , 'Technický kontakt je již přiřazen k tomuto objektu','24' );
INSERT INTO enum_reason VALUES(  25 , 'Technical contact does not exist' ,  'Technický kontakt neexistuje', '25');

INSERT INTO enum_reason VALUES(  26 , 'Administrative contact is already assigned to this object.'  , 'Administrátorský kontakt je již přiřazen k tomuto objektu', '26' );
INSERT INTO enum_reason VALUES(  27 , 'Administrative contact does not exist' ,  'Administrátorský kontakt neexistuje','27'   );
 
-- for domain when owner or nsset doesn't exist
INSERT INTO enum_reason VALUES( 28 ,  'host handle does not exist.' , 'sada jmenných serverů není vytvořena', '28' );
INSERT INTO enum_reason VALUES( 29 ,  'contact handle of registrant does not exist.' , 'ukazatel kontaktu vlastníka není vytvořen', '29' );

-- if dns host cannot be added or removed
INSERT INTO enum_reason VALUES( 30 , 'Nameserver is already set to this nsset.' , 'jmenný server DNS je již přiřazen sadě jmenných serverů', '30' );
INSERT INTO enum_reason VALUES( 31 , 'Nameserver is not set to this nsset.'  , 'jmenný server DNS není přiřazen sadě jmenných serverů', '31' );

-- for domain renew when entered date of epiration doesn't fit 
INSERT INTO enum_reason VALUES( 32 ,  'Expiration date does not match registry data.' , 'Nesouhlasí datum expirace', '32' );
 
-- error from mod_eppd, if it is missing 'op' attribute in transfer command 
INSERT INTO enum_reason VALUES( 33 ,  'Attribute op in element transfer is missing', 'Chybí atribut op u elementu transfer', '33' );
-- error from mod_eppd, if it is missing a type of ident element
INSERT INTO enum_reason VALUES( 34 ,  'Attribute type in element ident is missing', 'Chybí atribut type u elementu ident', '34' );
-- error from z mod_eppd, if it is missing attribute msgID in element poll
INSERT INTO enum_reason VALUES( 35 ,  'Attribute msgID in element poll is missing', 'Chybí atribut msgID u elementu poll', '35' );

-- blacklist domain
INSERT INTO enum_reason VALUES( 36 ,  'Registration is prohibited'  , 'Registrace je zakázána', '36' );
-- XML validation process failed
INSERT INTO enum_reason VALUES( 37 ,  'Schemas validity error: ' , 'Chyba validace XML schemat: ', '37' );

-- duplicate contact for tech or admin 
INSERT INTO enum_reason VALUES(  38 , 'Duplicity contact' , 'Duplicitní kontakt',  '38' );

---
--- moved from keyset.sql
---
INSERT INTO enum_reason VALUES (39, 'Bad format of keyset handle', 'Neplatný formát ukazatele keysetu', '39');
INSERT INTO enum_reason VALUES (40, 'Keyset handle does not exist', 'Ukazatel keysetu není vytvořen', '40');
INSERT INTO enum_reason VALUES (41, 'DSRecord does not exists', 'DSRecord záznam neexistuje', '41');
INSERT INTO enum_reason VALUES (42, 'Can not remove DSRecord', 'Nelze odstranit DSRecord záznam', '42');
INSERT INTO enum_reason VALUES (43, 'Duplicity DSRecord', 'Duplicitní DSRecord záznam', '43');
INSERT INTO enum_reason VALUES (44, 'DSRecord already exists for this keyset', 'DSRecord již pro tento keyset existuje', '44');
INSERT INTO enum_reason VALUES (45, 'DSRecord is not set for this keyset', 'DSRecord pro tento keyset neexistuje', '45');
INSERT INTO enum_reason VALUES (46, 'Field ``digest type'''' must be 1 (SHA-1)', 'Pole ``digest type'''' musí být 1 (SHA-1)', '46');
INSERT INTO enum_reason VALUES (47, 'Digest must be 40 characters long', 'Digest musí být dlouhý 40 znaků', '47');


INSERT INTO enum_reason VALUES (48, 'Object does not belong to the registrar', 'Objekt nepatří registrátorovi', '48');
INSERT INTO enum_reason VALUES (49, 'Too many technical administrators contacts.', 'Příliš mnoho administrátorských kontaktů', '49');
INSERT INTO enum_reason VALUES (50, 'Too many DS records', 'Příliš mnoho DS záznamů', '50');
INSERT INTO enum_reason VALUES (51, 'Too many DNSKEY records', 'Příliš mnoho DNSKEY záznamů', '51');

INSERT INTO enum_reason VALUES (52, 'Too many nameservers in this nsset', 'Příliš mnoho jmenných serverů DNS je přiřazeno sadě jmenných serverů', '52');
INSERT INTO enum_reason VALUES (53, 'No DNSKey record', 'Žádný DNSKey záznam', '53');
INSERT INTO enum_reason VALUES (54, 'Field ``flags'''' must be 0, 256 or 257', 'Pole ``flags'''' musí být 0, 256 nebo 257', '54');
INSERT INTO enum_reason VALUES (55, 'Field ``protocol'''' must be 3', 'Pole ``protocol'''' musí být 3', '55');
INSERT INTO enum_reason VALUES (56, 'Unsupported value of field "alg", see http://www.iana.org/assignments/dns-sec-alg-numbers', 'Nepodporovaná hodnota pole "alg", viz http://www.iana.org/assignments/dns-sec-alg-numbers', '56');
INSERT INTO enum_reason VALUES (57, 'Field ``key'''' has invalid length', 'Pole ``key'''' má špatnou délku', '57');
INSERT INTO enum_reason VALUES (58, 'Field ``key'''' contains invalid character', 'Pole ``key'''' obsahuje neplatný znak', '58');
INSERT INTO enum_reason VALUES (59, 'DNSKey already exists for this keyset', 'DNSKey již pro tento keyset existuje', '59');
INSERT INTO enum_reason VALUES (60, 'DNSKey does not exist for this keyset', 'DNSKey pro tento keyset neexistuje', '60');
INSERT INTO enum_reason VALUES (61, 'Duplicity DNSKey', 'Duplicitní DNSKey', '61');
INSERT INTO enum_reason VALUES (62, 'Keyset must have DNSKey or DSRecord', 'Keyset musí mít DNSKey nebo DSRecord', '62');
INSERT INTO enum_reason VALUES (63, 'Duplicated nameserver hostname', 'Duplicitní jméno jmenného serveru DNS', '63');
INSERT INTO enum_reason VALUES (64, 'Administrative contact not assigned to this object', 'Administrátorský kontakt není přiřazen k tomuto objektu', '64');
INSERT INTO enum_reason VALUES (65, 'Temporary contacts are obsolete', 'Dočasné kontakty již nejsou podporovány', '65');

INSERT INTO enum_reason VALUES (66, 'Too many addr records', '', '66');

-- TODO remove this after translating reasons
update enum_reason set reason_ru = reason;

SELECT setval('enum_reason_id_seq', 66);

comment on table enum_reason is 'Table of error messages reason';
comment on column enum_reason.reason is 'reason in english language';
comment on column enum_reason.reason_cs is 'reason in native language';
