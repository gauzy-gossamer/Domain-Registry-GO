
CREATE TABLE registrar_disconnect (
    id SERIAL CONSTRAINT registrar_disconnect_pkey PRIMARY KEY,
    registrarid INTEGER NOT NULL CONSTRAINT registrar_disconnect_registrarid_fkey REFERENCES registrar(id),
    blocked_from TIMESTAMP NOT NULL DEFAULT now(),
    blocked_to TIMESTAMP,
    unblock_request_id BIGINT
);

