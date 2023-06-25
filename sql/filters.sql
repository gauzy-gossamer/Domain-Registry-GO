-- login table for saved filters
-- DROP TABLE Filters CASCADE;
CREATE TABLE Filters (
	ID SERIAL CONSTRAINT filters_pkey PRIMARY KEY, 
	Type SMALLINT NOT NULL, 
	Name VARCHAR(255) NOT NULL, 
	UserID INTEGER NOT NULL, 
	GroupID INTEGER,
	Data TEXT NOT NULL
	);

COMMENT ON TABLE Filters is
'Table for saved object filters';

COMMENT ON COLUMN Filters.ID is 'unique automatically generated identifier';
COMMENT ON COLUMN Filters.Type is 'filter object type -- 0 = filter on filter, 1 = filter on registrar, 2 = filter on object, 3 = filter on contact, 4 = filter on nsset, 5 = filter on domain, 6 = filter on action, 7 = filter on invoice, 8 = filter on authinfo, 9 = filter on mail';
COMMENT ON COLUMN Filters.Name is 'human readable filter name';
COMMENT ON COLUMN Filters.UserID is 'filter creator';
COMMENT ON COLUMN Filters.GroupID is 'filter accessibility for group';
COMMENT ON COLUMN Filters.Data is 'filter definition';

