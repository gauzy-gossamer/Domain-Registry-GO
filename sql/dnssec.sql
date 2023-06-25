
CREATE TABLE zone_dnssec (
	id	SERIAL PRIMARY KEY,
	zoneid INTEGER REFERENCES zone(id) NOT NULL,
	enabled BOOLEAN DEFAULT FALSE,
	nsec3enabled BOOLEAN DEFAULT TRUE,
	nsec3options INTEGER DEFAULT 0,
	valid_from TIMESTAMPTZ DEFAULT now(),
	valid_till TIMESTAMPTZ DEFAULT NULL,
	profiles TEXT DEFAULT ''
);


CREATE TABLE zone_dnssec_notify (
	id	SERIAL PRIMARY KEY,
	dnssecid INTEGER REFERENCES zone_dnssec(id),
	zoneid INTEGER REFERENCES zone(id) NOT NULL,
	valid_from TIMESTAMPTZ DEFAULT now(),
	valid_till TIMESTAMPTZ DEFAULT NULL,
	msg TEXT DEFAULT NULL
);


