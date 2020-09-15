
CREATE TABLE IF NOT EXISTS repository (
	id SERIAL PRIMARY KEY,
	owner TEXT,
	name TEXT,
	stars INTEGER,
	forks INTEGER,
	subscribers INTEGER,
	watchers INTEGER,
	open_issues INTEGER,
	size INTEGER);

