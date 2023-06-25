
CREATE TABLE registrar_notify (
	id SERIAL PRIMARY KEY,
	registrarid INT REFERENCES registrar(id),
	url VARCHAR(4096),
	active INT DEFAULT 1
);

