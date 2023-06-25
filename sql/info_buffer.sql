CREATE TABLE epp_info_buffer_content (
    id INTEGER NOT NULL,
    registrar_id INTEGER NOT NULL CONSTRAINT epp_info_buffer_content_registrar_id_fkey REFERENCES registrar (id),
    object_id INTEGER NOT NULL,
    CONSTRAINT epp_info_buffer_content_pkey PRIMARY KEY (id,registrar_id)
);

CREATE TABLE epp_info_buffer (
    registrar_id INTEGER NOT NULL CONSTRAINT epp_info_buffer_registrar_id_fkey REFERENCES registrar (id),
    current INTEGER,
    CONSTRAINT epp_info_buffer_registrar_id_fkey1 FOREIGN KEY (registrar_id, current) REFERENCES epp_info_buffer_content (registrar_id, id),
    CONSTRAINT epp_info_buffer_pkey PRIMARY KEY (registrar_id)
);

CREATE INDEX epp_info_buffer_content_registrar_id_idx ON epp_info_buffer_content (registrar_id);