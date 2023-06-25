---
--- remove duplicate elements from array
---
CREATE OR REPLACE FUNCTION array_uniq(anyarray)
RETURNS anyarray as $$
SELECT array(SELECT DISTINCT $1[i] FROM
    generate_series(array_lower($1,1), array_upper($1,1)) g(i));
$$ LANGUAGE SQL STRICT IMMUTABLE;


---
--- remove null elements from array
---
CREATE OR REPLACE FUNCTION array_filter_null(anyarray)
RETURNS anyarray as $$
SELECT array(SELECT $1[i] FROM
    generate_series(array_lower($1,1), array_upper($1,1)) g(i) WHERE $1[i] IS NOT NULL) ;
$$ LANGUAGE SQL STRICT IMMUTABLE;



---
--- create unnest array function if missing
---
CREATE OR REPLACE FUNCTION create_unnest_if_missing()
  RETURNS void AS $$
  DECLARE
  BEGIN
    PERFORM * FROM pg_proc WHERE proname = 'unnest';
    IF NOT FOUND THEN
        CREATE OR REPLACE FUNCTION unnest(anyarray)
          RETURNS SETOF anyelement AS
        $BODY$
        SELECT $1[i] FROM
            generate_series(array_lower($1, 1),
                            array_upper($1, 1)) i;
        $BODY$
        LANGUAGE 'sql' IMMUTABLE;
    END IF;
    DROP FUNCTION create_unnest_if_missing();
  END;
$$ LANGUAGE plpgsql;

SELECT  create_unnest_if_missing();
