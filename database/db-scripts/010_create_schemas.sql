
CREATE TABLE IF NOT EXISTS repositories (
	id SERIAL,
	owner TEXT,
	name TEXT,
	stars INTEGER,
	forks INTEGER,
	subscribers INTEGER,
	watchers INTEGER,
	open_issues INTEGER,
	size INTEGER,
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	PRIMARY KEY (id, owner, name));

CREATE OR REPLACE FUNCTION trigger_set_timestamp()
RETURNS TRIGGER AS $$
BEGIN
	NEW.updated = NOW();
	RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER set_timestamp
BEFORE UPDATE ON repositories
FOR EACH ROW
EXECUTE PROCEDURE trigger_set_timestamp();

