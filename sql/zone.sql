-- DROP TABLE zone CASCADE;
CREATE TABLE zone (
        id SERIAL CONSTRAINT zone_pkey PRIMARY KEY,
        fqdn VARCHAR(255) CONSTRAINT zone_fqdn_key UNIQUE NOT NULL,  --zone fully qualified name
        ex_period_min int NOT NULL,  --minimal prolongation of the period of domains validity in months
        ex_period_max int NOT NULL,  --maximal prolongation of the period of domains validity in months
        dots_max  int NOT NULL DEFAULT 1,  --maximal number of dots in zone name
        warning_letter BOOLEAN NOT NULL DEFAULT TRUE
        );

COMMENT ON TABLE zone IS
'This table contains zone parameters';
COMMENT ON COLUMN zone.id IS 'unique automatically generated identifier';
COMMENT ON COLUMN zone.fqdn IS 'zone fully qualified name';
COMMENT ON COLUMN zone.ex_period_min IS 'minimal prolongation of the period of domains validity in months';
COMMENT ON COLUMN zone.ex_period_max IS 'maximal prolongation of the period of domains validity in months';
COMMENT ON COLUMN zone.dots_max IS 'maximal number of dots in zone name';

CREATE TABLE zone_groups
(
    zone_id integer CONSTRAINT zone_group_key REFERENCES zone(id),
    group_id integer
);

CREATE FUNCTION in_zone_group(integer, integer) RETURNS boolean as $$
    BEGIN
        RETURN exists(SELECT z2.zone_id FROM zone_groups z1 INNER JOIN zone_groups z2 ON z1.group_id=z2.group_id WHERE z1.zone_id = $2 and z2.zone_id = $1);
    END
$$ language plpgsql;

---
--- #9085 domain name validation configuration by zone
---

CREATE TABLE enum_domain_name_validation_checker (
        id SERIAL CONSTRAINT enum_domain_name_validation_checker_pkey PRIMARY KEY,
        name TEXT CONSTRAINT enum_domain_name_validation_checker_name_key UNIQUE NOT NULL,
        description TEXT NOT NULL
        );

COMMENT ON TABLE enum_domain_name_validation_checker IS
'This table contains names of domain name checkers used in server DomainNameValidator';
COMMENT ON COLUMN enum_domain_name_validation_checker.id IS 'unique automatically generated identifier';
COMMENT ON COLUMN enum_domain_name_validation_checker.name IS 'name of the checker';
COMMENT ON COLUMN enum_domain_name_validation_checker.description IS 'description of the checker';

CREATE TABLE zone_domain_name_validation_checker_map (
  id BIGSERIAL CONSTRAINT zone_domain_name_validation_checker_map_pkey PRIMARY KEY,
  checker_id INTEGER NOT NULL CONSTRAINT zone_domain_name_validation_checker_map_checker_id_fkey
    REFERENCES enum_domain_name_validation_checker (id),
  zone_id INTEGER NOT NULL CONSTRAINT zone_domain_name_validation_checker_map_zone_id_fkey
    REFERENCES zone (id),
  CONSTRAINT zone_domain_name_validation_checker_map_key UNIQUE (checker_id, zone_id)
);

COMMENT ON TABLE zone_domain_name_validation_checker_map IS
'This table domain name checkers applied to domain names by zone';
COMMENT ON COLUMN zone_domain_name_validation_checker_map.id IS 'unique automatically generated identifier';
COMMENT ON COLUMN zone_domain_name_validation_checker_map.checker_id IS 'checker';
COMMENT ON COLUMN zone_domain_name_validation_checker_map.zone_id IS 'zone';
