CREATE TABLE History (
        ID SERIAL CONSTRAINT history_pkey PRIMARY KEY,
        valid_from TIMESTAMP NOT NULL DEFAULT NOW(),
        valid_to TIMESTAMP,
        next INTEGER,
        request_id BIGINT
);

COMMENT ON TABLE history IS
'Main evidence table with modified data, it join historic tables modified during same operation
create - in case of any change';
COMMENT ON COLUMN history.id IS 'unique automatically generated identifier';
COMMENT ON COLUMN history.valid_from IS 'date from which was this history created';
COMMENT ON COLUMN history.valid_to IS 'date to which was history actual (NULL if it still is)';
COMMENT ON COLUMN history.next IS 'next history id';

CREATE INDEX history_action_valid_from_idx ON history (valid_from);
CREATE UNIQUE INDEX history_next_idx ON history (next);
CREATE INDEX history_request_id_idx ON history (request_id) WHERE request_id IS NOT NULL;

