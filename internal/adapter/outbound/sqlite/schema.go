package sqlite

const schema = `
CREATE TABLE IF NOT EXISTS packages (
	id				INTEGER PRIMARY KEY AUTOINCREMENT,
	name 			TEXT NOT NULL,
	version			TEXT NOT NULL,
	last_updated_at DATETIME NOT NULL,
	UNIQUE (name)
);

CREATE TABLE IF NOT EXISTS dependency_nodes (
	id			INTEGER PRIMARY KEY AUTOINCREMENT,
	package_id	INTEGER NOT NULL REFERENCES packages(id) ON DELETE CASCADE,
	name		TEXT NOT NULL,
	version		TEXT NOT NULL,
	relation	TEXT NOT NULL,
	score		REAL
);
`
