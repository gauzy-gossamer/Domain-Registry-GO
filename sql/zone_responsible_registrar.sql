
CREATE TABLE zone_responsible_registrar (
	id SERIAL PRIMARY KEY,
	zoneid INT REFERENCES public.zone(id) NOT NULL UNIQUE,
	registrarid INT REFERENCES public.registrar(id) NOT NULL
);
