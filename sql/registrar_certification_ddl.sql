CREATE DOMAIN classification_type AS integer NOT NULL
    CONSTRAINT classification_type_check CHECK (VALUE IN (0, 1, 2, 3, 4, 5)); 

COMMENT ON DOMAIN classification_type 
    IS 'allowed values of classification for registrar certification';

CREATE TABLE registrar_certification
(
    id serial CONSTRAINT registrar_certification_pkey PRIMARY KEY, -- certification id
    registrar_id integer NOT NULL CONSTRAINT registrar_certification_registrar_id_fkey REFERENCES registrar(id), -- registrar id
    valid_from date NOT NULL, --  registrar certification valid from
    valid_until date NOT NULL, --  registrar certification valid until = valid_from + 1year
    classification classification_type NOT NULL -- registrar certification result checked 0-5
);


-- check whether registrar_certification life is valid
CREATE OR REPLACE FUNCTION registrar_certification_life_check() 
RETURNS "trigger" AS $$
DECLARE
    last_reg_cert RECORD;
BEGIN
    IF NEW.valid_from > NEW.valid_until THEN
        RAISE EXCEPTION 'Invalid registrar certification life: valid_from > valid_until';
    END IF;

    IF TG_OP = 'INSERT' THEN
        SELECT * FROM registrar_certification INTO last_reg_cert
            WHERE registrar_id = NEW.registrar_id AND id < NEW.id
            ORDER BY valid_from DESC, id DESC LIMIT 1;
        IF FOUND THEN
            IF last_reg_cert.valid_until > NEW.valid_from  THEN
                RAISE EXCEPTION 'Invalid registrar certification life: last valid_until > new valid_from';
            END IF;
        END IF;
    ELSEIF TG_OP = 'UPDATE' THEN
        IF NEW.valid_from <> OLD.valid_from THEN
            RAISE EXCEPTION 'Change of valid_from not allowed';
        END IF;
        IF NEW.valid_until > OLD.valid_until THEN
            RAISE EXCEPTION 'Certification prolongation not allowed';
        END IF;
        IF NEW.registrar_id <> OLD.registrar_id THEN
            RAISE EXCEPTION 'Change of registrar not allowed';
        END IF;
    END IF;

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION registrar_certification_life_check()
    IS 'check whether registrar_certification life is valid'; 

CREATE TRIGGER "trigger_registrar_certification"
  AFTER INSERT OR UPDATE ON registrar_certification
  FOR EACH ROW EXECUTE PROCEDURE registrar_certification_life_check();


CREATE INDEX registrar_certification_valid_from_idx ON registrar_certification(valid_from);
CREATE INDEX registrar_certification_valid_until_idx ON registrar_certification(valid_until);

COMMENT ON TABLE registrar_certification IS 'result of registrar certification';
COMMENT ON COLUMN registrar_certification.registrar_id IS 'certified registrar id';
COMMENT ON COLUMN registrar_certification.valid_from IS
    'certification is valid from this date';
COMMENT ON COLUMN registrar_certification.valid_until IS
    'certification is valid until this date, certification should be valid for 1 year';
COMMENT ON COLUMN registrar_certification.classification IS
    'registrar certification result checked 0-5';

CREATE TABLE registrar_group
(
    id serial CONSTRAINT registrar_group_pkey PRIMARY KEY, -- registrar group id
    short_name varchar(255) NOT NULL CONSTRAINT registrar_group_short_name_key UNIQUE, -- short name of the group
    cancelled timestamp -- when the group was cancelled
);


--check whether registrar_group is empty and not cancelled
CREATE OR REPLACE FUNCTION cancel_registrar_group_check() RETURNS "trigger" AS $$
DECLARE
    registrars_in_group INTEGER;
BEGIN
    IF OLD.cancelled IS NOT NULL THEN
        RAISE EXCEPTION 'Registrar group already cancelled';
    END IF;

    IF NEW.cancelled IS NOT NULL AND EXISTS(
        SELECT * 
          FROM registrar_group_map 
         WHERE registrar_group_id = NEW.id
          AND registrar_group_map.member_from <= CURRENT_DATE
          AND (registrar_group_map.member_until IS NULL 
                  OR (registrar_group_map.member_until >= CURRENT_DATE 
                          AND  registrar_group_map.member_from 
                              <> registrar_group_map.member_until))) 
    THEN 
        RAISE EXCEPTION 'Unable to cancel non-empty registrar group';
    END IF;
    
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION cancel_registrar_group_check()
    IS 'check whether registrar_group is empty and not cancelled'; 

CREATE TRIGGER "trigger_cancel_registrar_group"
  AFTER UPDATE  ON registrar_group
  FOR EACH ROW EXECUTE PROCEDURE cancel_registrar_group_check();


CREATE INDEX registrar_group_short_name_idx ON registrar_group(short_name);

COMMENT ON TABLE registrar_group IS 'available groups of registars';
COMMENT ON COLUMN registrar_group.id IS 'group id';
COMMENT ON COLUMN registrar_group.short_name IS 'group short name';
COMMENT ON COLUMN registrar_group.cancelled IS 'time when the group was cancelled';

CREATE TABLE registrar_group_map
(
    id serial CONSTRAINT registrar_group_map_pkey PRIMARY KEY, -- membership of registrar in group id
    registrar_id integer NOT NULL CONSTRAINT registrar_group_map_registrar_id_fkey REFERENCES registrar(id), -- registrar id
    registrar_group_id integer NOT NULL CONSTRAINT registrar_group_map_registrar_group_id_fkey REFERENCES registrar_group(id), -- registrar group id
    member_from date NOT NULL, --  registrar membership in the group from this date
    member_until date --  registrar membership in the group until this date or unspecified
);

CREATE OR REPLACE FUNCTION registrar_group_map_check() RETURNS "trigger" AS $$
DECLARE
    last_reg_map RECORD;
BEGIN
    IF NEW.member_until IS NOT NULL AND NEW.member_from > NEW.member_until THEN
        RAISE EXCEPTION 'Invalid registrar membership life: member_from > member_until';
    END IF;

    IF TG_OP = 'INSERT' THEN
        SELECT * INTO last_reg_map
           FROM registrar_group_map 
          WHERE registrar_id = NEW.registrar_id
            AND registrar_group_id = NEW.registrar_group_id
            AND id < NEW.id
          ORDER BY member_from DESC, id DESC 
          LIMIT 1;
        IF FOUND THEN
            IF last_reg_map.member_until IS NULL THEN
                UPDATE registrar_group_map 
                   SET member_until = NEW.member_from
                  WHERE id = last_reg_map.id;
                last_reg_map.member_until := NEW.member_from;
            END IF;
            IF last_reg_map.member_until > NEW.member_from  THEN
                RAISE EXCEPTION 'Invalid registrar membership life: last member_until > new member_from';
            END IF;
        END IF;

    ELSEIF TG_OP = 'UPDATE' THEN
        IF NEW.member_from <> OLD.member_from THEN
            RAISE EXCEPTION 'Change of member_from not allowed';
        END IF;
        
        IF NEW.member_until IS NULL AND OLD.member_until IS NOT NULL THEN
            RAISE EXCEPTION 'Change of member_until not allowed';
        END IF;
        
        IF NEW.member_until IS NOT NULL AND OLD.member_until IS NOT NULL 
            AND NEW.member_until <> OLD.member_until THEN
            RAISE EXCEPTION 'Change of member_until not allowed';
        END IF;
        
        IF NEW.registrar_group_id <> OLD.registrar_group_id THEN
            RAISE EXCEPTION 'Change of registrar_group not allowed';
        END IF;
        
        IF NEW.registrar_id <> OLD.registrar_id THEN
            RAISE EXCEPTION 'Change of registrar not allowed';
        END IF;
    END IF;

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION registrar_group_map_check()
    IS 'check whether registrar membership change is valid'; 

CREATE TRIGGER "trigger_registrar_group_map"
  AFTER INSERT OR UPDATE ON registrar_group_map
  FOR EACH ROW EXECUTE PROCEDURE registrar_group_map_check();


CREATE INDEX registrar_group_map_member_from_idx ON registrar_group_map(member_from);
CREATE INDEX registrar_group_map_member_until_idx ON registrar_group_map(member_until);

COMMENT ON TABLE registrar_group_map IS 'membership of registar in group';
COMMENT ON COLUMN registrar_group_map.id IS 'registrar group membership id';
COMMENT ON COLUMN registrar_group_map.registrar_id IS 'registrar id';
COMMENT ON COLUMN registrar_group_map.registrar_group_id IS 'group id';
COMMENT ON COLUMN registrar_group_map.member_from 
    IS 'registrar membership in the group from this date';
COMMENT ON COLUMN registrar_group_map.member_until 
    IS 'registrar membership in the group until this date or unspecified';
